"""Tests of the session binding hook (D-6, D-12).

Run: python3 -m unittest discover -s docs/tools -p 'test_*.py'
"""
import importlib.util
import json
import os
import subprocess
import tempfile
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
HOOK = os.path.join(HERE, "..", "..", ".claude", "hooks", "session_bind.py")
spec = importlib.util.spec_from_file_location("session_bind", HOOK)
sb = importlib.util.module_from_spec(spec)
spec.loader.exec_module(sb)


def branch_of(directory):
    return {"/repo": "pr60-example", "/wt": "pr61-other"}.get(directory)


def branch_of_pr(number, directory):
    return {"182": "feat/load-policy", "190": "pr60-example"}.get(number)


def targets(command, cwd="/repo"):
    return sb.targets(command, cwd, branch_of, branch_of_pr)


class Targets(unittest.TestCase):
    def test_branch_creation_commands(self):
        self.assertEqual(targets("git switch -c pr60-example"), ["pr60-example"])
        self.assertEqual(targets("git checkout -b pr60-example origin/main"), ["pr60-example"])
        self.assertEqual(targets("git worktree add -b pr60-example /tmp/wt origin/main"), ["pr60-example"])

    def test_push_reads_the_refspec_or_the_current_branch(self):
        self.assertEqual(targets("git push -u origin pr60-example"), ["pr60-example"])
        self.assertEqual(targets("git push"), ["pr60-example"])
        self.assertEqual(targets("git push origin HEAD:refs/heads/pr61-other"), ["pr61-other"])
        self.assertEqual(targets("git push origin --delete old-branch"), [])
        self.assertEqual(targets("git push origin :old-branch"), [])
        self.assertEqual(targets("git push origin +:old-branch"), [])

    def test_cd_and_dash_c_move_the_directory(self):
        self.assertEqual(targets("cd /wt && git push"), ["pr61-other"])
        self.assertEqual(targets("git -C /wt push"), ["pr61-other"])
        self.assertEqual(targets("WT=/wt; cd $WT && gh pr create --fill"), ["pr61-other"])

    def test_an_unknown_directory_fails_open(self):
        self.assertEqual(targets("cd $UNSET_DIR_FOR_TEST && git push"), [])

    def test_gh_commands(self):
        self.assertEqual(targets("gh pr create --title x --body-file b.md"), ["pr60-example"])
        self.assertEqual(targets("gh pr create --head pr61-other"), ["pr61-other"])
        self.assertEqual(targets("gh pr checkout 182"), ["feat/load-policy"])
        self.assertEqual(targets('gh pr comment 190 --body "A reply"'), ["pr60-example"])
        self.assertEqual(targets("gh pr view 182 --json state"), [])

    def test_read_only_and_main_commands_bind_nothing(self):
        self.assertEqual(targets("git status && git log --oneline -3"), [])
        self.assertEqual(targets("git switch main"), [])
        self.assertEqual(targets("make where"), [])


class Decide(unittest.TestCase):
    def setUp(self):
        self.store = tempfile.mkdtemp()

    def test_a_clean_session_binds_its_first_pull_request(self):
        allowed, _ = sb.decide(self.store, "s1", ["pr60-example"], "git switch -c pr60-example")
        self.assertTrue(allowed)
        allowed, _ = sb.decide(self.store, "s1", ["pr60-example"], "gh pr create --fill")
        self.assertTrue(allowed)

    def test_the_same_session_can_not_start_a_second_pull_request(self):
        sb.decide(self.store, "s1", ["pr60-example"], "git switch -c pr60-example")
        allowed, message = sb.decide(self.store, "s1", ["pr61-other"], "git switch -c pr61-other")
        self.assertFalse(allowed)
        self.assertIn(sb.BLOCKED, message)

    def test_a_merge_record_branch_after_the_merge_is_blocked(self):
        sb.decide(self.store, "s1", ["feat/load-policy"], "gh pr create --fill")
        allowed, message = sb.decide(self.store, "s1", ["docs/record-merge-12"], "git switch -c docs/record-merge-12")
        self.assertFalse(allowed)
        self.assertIn("bound to branch feat/load-policy", message)
        self.assertIn("D-12", message)

    def test_a_reviewer_session_on_the_same_pull_request_is_allowed(self):
        sb.decide(self.store, "r1", ["pr60-example"], "gh pr checkout 190")
        allowed, _ = sb.decide(self.store, "r1", ["pr60-example"], 'gh pr comment 190 --body "A reply"')
        self.assertTrue(allowed)

    def test_another_session_binds_on_its_own(self):
        sb.decide(self.store, "s1", ["pr60-example"], "git switch -c pr60-example")
        allowed, _ = sb.decide(self.store, "s2", ["pr61-other"], "git switch -c pr61-other")
        self.assertTrue(allowed)


class Hook(unittest.TestCase):
    """The hook as Claude Code runs it: JSON on stdin, exit 2 blocks."""

    def run_hook(self, repo, session_id, command):
        event = {"session_id": session_id, "cwd": repo, "hook_event_name": "PreToolUse", "tool_name": "Bash", "tool_input": {"command": command}}
        return subprocess.run(["python3", HOOK], input=json.dumps(event), capture_output=True, text=True)

    def test_the_store_names_no_repository(self):
        self.assertNotIn("decktome", sb.STORE)
        with open(HOOK, encoding="utf-8") as handle:
            code = handle.read().split('"""', 2)[2]
        self.assertNotIn("decktome", code.lower())

    def test_exit_codes_on_a_real_repository(self):
        repo = tempfile.mkdtemp()
        subprocess.run(["git", "init", "-q", "-b", "main", repo], check=True)
        first = self.run_hook(repo, "abc-1", "git switch -c pr60-example")
        self.assertEqual(first.returncode, 0, first.stderr)
        second = self.run_hook(repo, "abc-1", "git checkout -b pr61-other")
        self.assertEqual(second.returncode, 2)
        self.assertIn(sb.BLOCKED, second.stderr)
        bound = os.path.join(repo, ".git", sb.STORE, "abc-1.json")
        self.assertTrue(os.path.exists(bound))
        other = self.run_hook(repo, "abc-2", "git checkout -b pr61-other")
        self.assertEqual(other.returncode, 0, other.stderr)

    def test_bad_input_fails_open(self):
        out = subprocess.run(["python3", HOOK], input="not json", capture_output=True, text=True)
        self.assertEqual(out.returncode, 0)
        repo = tempfile.mkdtemp()
        self.assertEqual(self.run_hook(repo, "../escape", "git switch -c x").returncode, 0)


if __name__ == "__main__":
    unittest.main()
