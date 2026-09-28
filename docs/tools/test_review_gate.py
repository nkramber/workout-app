"""Tests of review_gate.py (D-4, D-8, D-15)."""
import importlib.util
import json
import os
import subprocess
import sys
import tempfile
import unittest
import unittest.mock

HERE = os.path.dirname(os.path.abspath(__file__))
spec = importlib.util.spec_from_file_location("review_gate", os.path.join(HERE, "review_gate.py"))
rg = importlib.util.module_from_spec(spec)
spec.loader.exec_module(rg)

N = 212
HEAD = "a1b2c3d4e5f60718293a4b5c6d7e8f9012345678"
CODE = [(HEAD, ["go/internal/policy/load.go"])]
SESSION = "session@example.com"
REPO = "owner/gym-route"


def record(head=HEAD[:7], verdict=f"**{rg.APPROVED}.** This verdict applies to head `x`.", identity=True):
    lines = ["# Pull request 212 review", "", "Date: 2026-09-23", ""]
    if identity:
        lines += ["## Identity", "", "- PR: 212", f"- Head: `{head}`", ""]
    lines += ["## Findings", "", "No finding.", "", "## Verdict", "", verdict, ""]
    return "\n".join(lines)


def run(labels=(), files=("go/internal/policy/load.go",), commits=CODE, authors=(SESSION,), text=None, author="owner",
        fork=""):
    reader = lambda path: text if path == rg.record_path(N) else None
    results = rg.evaluate(N, author, set(labels), list(files), commits, set(authors), reader, fork=fork)
    return [(rule, state) for rule, state, _ in results], results


def faults(results):
    return [r for r in results if r[1] == "FAULT"]


class ReviewRecord(unittest.TestCase):
    def test_an_approved_record_of_the_effective_head_passes(self):
        states, results = run(text=record())
        self.assertEqual(faults(states), [], results)

    def test_no_record_fails(self):
        states, _ = run()
        self.assertIn(("RG 3", "FAULT"), states)

    def test_changes_required_fails(self):
        states, _ = run(text=record(verdict="**Changes required.** Two findings stay open."))
        self.assertIn(("RG 4", "FAULT"), states)

    def test_a_bold_negation_fails(self):
        states, _ = run(text=record(verdict=f"**Not {rg.APPROVED}.**"))
        self.assertIn(("RG 4", "FAULT"), states)

    def test_two_verdicts_fail(self):
        states, _ = run(text=record(verdict=f"**Changes required.** then **{rg.APPROVED}.**"))
        self.assertIn(("RG 4", "FAULT"), states)

    def test_an_earlier_verdict_in_its_own_section_passes(self):
        text = record().replace("## Verdict", "## Earlier verdicts\n\n**Changes required.** Fixed in abc1234.\n\n## Verdict")
        states, results = run(text=text)
        self.assertEqual(faults(states), [], results)

    def test_a_verdict_heading_with_more_words_is_not_the_section(self):
        text = record().replace("## Verdict", "## Verdict history")
        states, _ = run(text=text)
        self.assertIn(("RG 4", "FAULT"), states)

    def test_a_stale_head_fails(self):
        states, _ = run(text=record(head="0000000"))
        self.assertIn(("RG 5", "FAULT"), states)

    def test_a_short_head_fails(self):
        states, _ = run(text=record(head=HEAD[:6]))
        self.assertIn(("RG 5", "FAULT"), states)

    def test_a_head_outside_the_identity_list_fails(self):
        text = record(identity=False) + f"\n## Notes\n\n- Head: `{HEAD}`\n"
        states, _ = run(text=text)
        self.assertIn(("RG 5", "FAULT"), states)

    def test_a_metadata_commit_keeps_the_effective_head(self):
        commits = CODE + [("f" * 40, [rg.record_path(N), "docs/session-handoff.md"])]
        states, results = run(commits=commits, text=record())
        self.assertEqual(faults(states), [], results)

    def test_a_commit_of_another_record_moves_the_effective_head(self):
        commits = CODE + [("f" * 40, ["docs/reviews/pr-211.md"])]
        self.assertEqual(rg.effective_head(commits, N), "f" * 40)
        # D-4: during the roadmap period, the change needs a current review.
        states, _ = run(commits=commits, text=record())
        self.assertIn(("RG 5", "FAULT"), states)

    def test_metadata_commits_alone_have_no_effective_head(self):
        states, _ = run(commits=[(HEAD, ["docs/session-handoff.md"])], text=record())
        self.assertIn(("RG 5", "FAULT"), states)


class RecordCommit(unittest.TestCase):
    """D-87: the gate names the commit that last changed the record, and the owner reads it."""

    def test_an_approved_record_names_its_last_commit(self):
        results = rg.evaluate(N, "owner", set(), ["go/a.go"], CODE, {SESSION}, lambda p: record(),
                              record_commit="abc1234 docs(review): add the review record of #7")
        rule = [r for r in results if r[0] == "RG 6"]
        self.assertEqual(len(rule), 1)
        self.assertEqual(rule[0][1], "INFO")
        self.assertIn("abc1234 docs(review): add the review record of #7", rule[0][2])
        self.assertIn("D-87", rule[0][2])

    def test_a_record_that_the_author_writes_passes_and_names_its_commit(self):
        # The finding P1-1 of PR #1: the gate can not tell a Codex record from an author record.
        # D-87 accepts the risk, so the gate passes, and RG 6 shows the commit to the owner.
        results = rg.evaluate(N, "owner", set(), ["go/a.go"], CODE, {SESSION}, lambda p: record(),
                              record_commit="def5678 fix: write the record by hand")
        self.assertEqual(faults(results), [])
        self.assertIn("def5678 fix: write the record by hand", dict((r[0], r[2]) for r in results)["RG 6"])

    def test_no_commit_name_asks_for_a_check_by_hand(self):
        results = rg.evaluate(N, "owner", set(), ["go/a.go"], CODE, {SESSION}, lambda p: record())
        self.assertIn("by hand", dict((r[0], r[2]) for r in results)["RG 6"])


class Fork(unittest.TestCase):
    def test_a_record_from_a_fork_fails(self):
        rules, results = run(text=record(), fork="someone/gym-route")
        self.assertIn(("RG 3", "FAULT"), rules)
        self.assertNotIn(("RG 5", "PASS"), rules)

    def test_a_record_of_the_same_repository_passes(self):
        rules, results = run(text=record())
        self.assertEqual(faults(rules), [])


class DocumentsDuringTheRoadmapPeriod(unittest.TestCase):
    """D-4, finding P1-4 of PR #1: each document change needs a current review during the period."""

    def test_a_commit_of_each_kind_of_document_needs_a_new_review(self):
        self.assertFalse(rg.OVERRIDE_ENABLED)
        for path in ["docs/roadmaps/high-level-roadmap.md", "docs/decisions.md", ".claude/skills/pr-review/SKILL.md",
                     "AGENTS.md", "README.md"]:
            with self.subTest(path=path):
                states, _ = run(commits=CODE + [("d" * 40, [path])], text=record())
                self.assertIn(("RG 5", "FAULT"), states)

    def test_the_metadata_set_keeps_the_approval(self):
        commits = CODE + [("d" * 40, [rg.record_path(N), f"docs/reviews/pr-{N}-response.md", "docs/session-handoff.md"])]
        states, results = run(commits=commits, text=record())
        self.assertEqual(faults(states), [], results)


class DocumentsAfterTheApproval(unittest.TestCase):
    """After the D-4 period, a commit of documents alone keeps a green gate green."""

    def setUp(self):
        patch = unittest.mock.patch.object(rg, "OVERRIDE_ENABLED", True)
        patch.start()
        self.addCleanup(patch.stop)

    def test_a_commit_of_each_kind_of_document_keeps_the_gate(self):
        for path in ["docs/roadmaps/high-level-roadmap.md", "docs/decisions.md", ".claude/skills/pr-review/SKILL.md",
                     "CLAUDE.md", "AGENTS.md", "README.md", ".github/pull_request_template.md",
                     "docs/research/data.json"]:
            with self.subTest(path=path):
                states, results = run(commits=CODE + [("d" * 40, [path])], text=record())
                self.assertEqual(faults(states), [], results)
                self.assertIn("each later commit changes documents alone", dict((r[0], r[2]) for r in results)["RG 5"])

    def test_many_commits_of_documents_keep_the_gate(self):
        commits = CODE + [("d" * 40, ["docs/decisions.md"]), ("e" * 40, ["docs/session-handoff.md"]),
                          ("f" * 40, ["docs/roadmaps/high-level-roadmap.md", ".claude/skills/ste-writing/SKILL.md"])]
        states, results = run(commits=commits, text=record())
        self.assertEqual(faults(states), [], results)

    def test_a_code_commit_after_a_commit_of_documents_fails(self):
        commits = CODE + [("d" * 40, ["docs/decisions.md"]), ("e" * 40, ["go/internal/policy/round.go"])]
        states, _ = run(commits=commits, text=record())
        self.assertIn(("RG 5", "FAULT"), states)

    def test_a_commit_that_mixes_a_document_and_code_fails(self):
        commits = CODE + [("d" * 40, ["docs/decisions.md", "Makefile"])]
        states, _ = run(commits=commits, text=record())
        self.assertIn(("RG 5", "FAULT"), states)

    def test_a_refused_path_after_the_approval_fails(self):
        for path in ["docs/tools/review_gate.py", ".claude/hooks/session_bind.py", ".claude/settings.json",
                     ".claude/settings.local.json", ".github/workflows/review-gate.yml", "docsx/a.md"]:
            with self.subTest(path=path):
                states, _ = run(commits=CODE + [("d" * 40, [path])], text=record())
                self.assertIn(("RG 5", "FAULT"), states)

    def test_a_merge_commit_after_the_approval_fails(self):
        states, _ = run(commits=CODE + [("d" * 40, [rg.MERGE])], text=record())
        self.assertIn(("RG 5", "FAULT"), states)

    def test_a_head_outside_the_branch_fails_even_before_documents(self):
        commits = CODE + [("d" * 40, ["docs/decisions.md"])]
        states, _ = run(commits=commits, text=record(head="0000000"))
        self.assertIn(("RG 5", "FAULT"), states)

    def test_a_documentation_pull_request_keeps_its_approval(self):
        first = "c" * 40
        commits = [(first, ["docs/decisions.md"]), ("d" * 40, ["docs/roadmaps/high-level-roadmap.md"])]
        states, results = run(files=["docs/decisions.md", "docs/roadmaps/high-level-roadmap.md"], commits=commits,
                              text=record(head=first[:7]))
        self.assertEqual(faults(states), [], results)

    def test_the_check_without_commits_needs_the_effective_head_itself(self):
        state, _ = rg.check_head(rg.record_path(N), record(), "d" * 40)
        self.assertEqual(state, "FAULT")


class OverrideLabel(unittest.TestCase):
    """D-4 keeps OVERRIDE_ENABLED False. The label then satisfies no rule."""

    DOCS = ["docs/decisions.md", "docs/session-handoff.md", ".claude/skills/pr-review/SKILL.md", "CLAUDE.md"]

    def test_the_constant_is_false_for_the_roadmap_period(self):
        self.assertIs(rg.OVERRIDE_ENABLED, False)

    def test_while_disabled_the_label_does_not_pass_documents(self):
        states, results = run(labels=[rg.LABEL], files=self.DOCS)
        self.assertIn(("RG 1", "SKIP"), states)
        self.assertIn(("RG 3", "FAULT"), states)
        self.assertIn("D-4", results[0][2])

    def test_while_disabled_a_documentation_pull_request_needs_the_record(self):
        states, results = run(labels=[rg.LABEL], files=self.DOCS, commits=[(HEAD, self.DOCS)], text=record())
        self.assertEqual(faults(states), [], results)
        self.assertIn(("RG 5", "PASS"), states)

    def test_no_label_needs_the_record_even_for_documents(self):
        states, _ = run(files=["docs/decisions.md"])
        self.assertIn(("RG 3", "FAULT"), states)


@unittest.mock.patch.object(rg, "OVERRIDE_ENABLED", True)
class OverrideLabelEnabled(unittest.TestCase):
    """After the owner ends the D-4 period, the label passes documents alone (D-15)."""

    def test_a_documentation_pull_request_with_the_label_passes(self):
        files = ["docs/decisions.md", "docs/session-handoff.md", ".claude/skills/pr-review/SKILL.md", "CLAUDE.md"]
        states, results = run(labels=[rg.LABEL], files=files)
        self.assertEqual(states, [("RG 1", "PASS")], results)

    def test_the_label_does_not_pass_code(self):
        states, _ = run(labels=[rg.LABEL])
        self.assertIn(("RG 1", "FAULT"), states)
        self.assertIn(("RG 3", "FAULT"), states)

    def test_the_label_does_not_pass_a_refused_path(self):
        for path in ["docs/tools/review_gate.py", ".claude/hooks/session_bind.py", ".claude/settings.json",
                     ".claude/settings.local.json", ".github/workflows/review-gate.yml", ".githooks/commit-msg",
                     "Makefile", "docsx/a.md"]:
            with self.subTest(path=path):
                states, _ = run(labels=[rg.LABEL], files=["docs/decisions.md", path])
                self.assertIn(("RG 1", "FAULT"), states)


class NoExemptAuthor(unittest.TestCase):
    """D-90, finding P1-5 of PR #1: no author skips the review, Dependabot included."""

    def test_a_dependabot_pull_request_needs_the_record(self):
        for files in (["go.mod"], ["docs/decisions.md"]):
            with self.subTest(files=files):
                states, _ = run(files=files, author="dependabot[bot]", authors=["49699333+dependabot[bot]@users.noreply.github.com"])
                self.assertIn(("RG 3", "FAULT"), states)

    def test_the_gate_names_no_exempt_author(self):
        self.assertFalse(hasattr(rg, "DEPENDABOT"))


class GitFacts(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.repo = self.tmp.name
        self.git("init", "-q", "-b", "main")
        self.commit({"README.md": "base\n"}, "base")
        self.base = self.git("rev-parse", "HEAD")
        self.git("checkout", "-q", "-b", "work")

    def tearDown(self):
        self.tmp.cleanup()

    def git(self, *args, email=SESSION, committer=None):
        env = {**os.environ, "GIT_AUTHOR_NAME": "a", "GIT_AUTHOR_EMAIL": email,
               "GIT_COMMITTER_NAME": "a", "GIT_COMMITTER_EMAIL": committer or email}
        out = subprocess.run(["git", *args], cwd=self.repo, capture_output=True, text=True, env=env, check=True)
        return out.stdout.strip()

    def commit(self, files, message, email=SESSION, committer=None):
        for path, text in files.items():
            full = os.path.join(self.repo, path)
            os.makedirs(os.path.dirname(full), exist_ok=True)
            with open(full, "w", encoding="utf-8") as handle:
                handle.write(text)
            self.git("add", path)
        self.git("commit", "-q", "-m", message, email=email, committer=committer)
        return self.git("rev-parse", "HEAD")

    def gate(self, labels=(), author="owner", head_repo=REPO, enabled=False):
        """Run the command. With enabled, run a copy with OVERRIDE_ENABLED True (the state after D-4)."""
        script = os.path.join(HERE, "review_gate.py")
        if enabled:
            with open(script, encoding="utf-8") as handle:
                text = handle.read()
            self.assertIn("OVERRIDE_ENABLED = False  # D-4", text)
            script = os.path.join(self.tmp.name + "-tool.py")
            with open(script, "w", encoding="utf-8") as handle:
                handle.write(text.replace("OVERRIDE_ENABLED = False  # D-4", "OVERRIDE_ENABLED = True"))
            self.addCleanup(os.remove, script)
        event = os.path.join(self.repo, "..", f"event-{os.path.basename(self.repo)}.json")
        head = {"sha": "x", "repo": None if head_repo is None else {"full_name": head_repo}}
        with open(event, "w", encoding="utf-8") as handle:
            json.dump({"pull_request": {"number": N, "user": {"login": author},
                                        "base": {"sha": self.base, "repo": {"full_name": REPO}}, "head": head,
                                        "labels": [{"name": name} for name in labels]}}, handle)
        out = subprocess.run([sys.executable, script, "--event", event,
                              "--head", "HEAD", "--repo", self.repo], capture_output=True, text=True)
        os.remove(event)
        return out.returncode, out.stdout

    def test_the_command_passes_an_approved_head_and_ignores_the_review_commit(self):
        code = self.commit({"go/a.go": "package a\n"}, "code")
        self.commit({rg.record_path(N): record(head=code[:10])}, "review")
        status, out = self.gate()
        self.assertEqual(status, 0, out)
        self.assertIn(f"effective head `{code}`", out)
        self.assertIn("RG 6: INFO", out)
        self.assertIn(" review` last changed", out)

    def test_the_command_fails_an_approved_record_from_a_fork(self):
        code = self.commit({"go/a.go": "package a\n"}, "code")
        self.commit({rg.record_path(N): record(head=code)}, "review")
        status, out = self.gate(head_repo="someone/gym-route")
        self.assertEqual(status, 1, out)
        self.assertIn("RG 3: FAULT", out)
        self.assertIn("`someone/gym-route`", out)

    def test_the_command_fails_a_record_whose_head_repository_is_gone(self):
        code = self.commit({"go/a.go": "package a\n"}, "code")
        self.commit({rg.record_path(N): record(head=code)}, "review")
        status, out = self.gate(head_repo=None)
        self.assertEqual(status, 1, out)
        self.assertIn("RG 3: FAULT", out)

    def test_the_label_passes_documents_from_a_fork_after_the_period_alone(self):
        self.commit({"docs/decisions.md": "| D-1 |\n"}, "docs")
        status, out = self.gate(labels=[rg.LABEL], head_repo="someone/gym-route")
        self.assertEqual(status, 1, out)
        status, out = self.gate(labels=[rg.LABEL], head_repo="someone/gym-route", enabled=True)
        self.assertEqual(status, 0, out)

    def test_the_command_fails_a_code_push_after_the_review(self):
        code = self.commit({"go/a.go": "package a\n"}, "code")
        self.commit({rg.record_path(N): record(head=code)}, "review")
        self.commit({"go/b.go": "package a\n"}, "more code")
        status, out = self.gate()
        self.assertEqual(status, 1, out)
        self.assertIn("RG 5: FAULT", out)

    def test_the_command_fails_a_commit_of_documents_after_the_review_in_the_period(self):
        code = self.commit({"go/a.go": "package a\n"}, "code")
        self.commit({rg.record_path(N): record(head=code)}, "review")
        self.commit({"docs/roadmaps/high-level-roadmap.md": "mark\n"}, "documents")
        status, out = self.gate()
        self.assertEqual(status, 1, out)
        self.assertIn("RG 5: FAULT", out)

    def test_the_command_passes_a_commit_of_documents_after_the_review(self):
        code = self.commit({"go/a.go": "package a\n"}, "code")
        self.commit({rg.record_path(N): record(head=code)}, "review")
        self.commit({"docs/roadmaps/high-level-roadmap.md": "mark\n", "docs/decisions.md": "row\n"}, "documents")
        status, out = self.gate(enabled=True)
        self.assertEqual(status, 0, out)
        self.assertIn("RG 5: PASS", out)
        self.assertIn("each later commit changes documents alone", out)

    def test_the_command_fails_code_after_a_commit_of_documents(self):
        code = self.commit({"go/a.go": "package a\n"}, "code")
        self.commit({rg.record_path(N): record(head=code)}, "review")
        self.commit({"docs/decisions.md": "row\n"}, "documents")
        self.commit({"docs/tools/new.py": "print(1)\n"}, "a tool")
        status, out = self.gate()
        self.assertEqual(status, 1, out)
        self.assertIn("RG 5: FAULT", out)

    def test_a_rename_into_docs_counts_as_a_change_of_code(self):
        self.commit({"go/a.go": "package a\n"}, "code")
        self.base = self.git("rev-parse", "HEAD")
        os.makedirs(os.path.join(self.repo, "docs"))
        self.git("mv", "go/a.go", "docs/a.go")
        self.git("commit", "-q", "-m", "move")
        status, out = self.gate(labels=[rg.LABEL], enabled=True)
        self.assertEqual(status, 1, out)
        self.assertIn("`go/a.go`", out)

    def test_a_merge_of_main_after_the_review_moves_the_effective_head(self):
        code = self.commit({"go/a.go": "package a\n"}, "code")
        self.commit({rg.record_path(N): record(head=code)}, "review")
        self.git("checkout", "-q", "main")
        self.commit({"go/c.go": "package c\n"}, "main moves")
        self.git("checkout", "-q", "work")
        self.git("merge", "-q", "--no-edit", "main")
        merge = self.git("rev-parse", "HEAD")
        status, out = self.gate()
        self.assertEqual(status, 1, out)
        self.assertIn(f"effective head is `{merge}`", out)

    def test_the_effective_head_mode_prints_the_rule_of_the_check(self):
        code = self.commit({"go/a.go": "package a\n"}, "code")
        self.commit({"docs/session-handoff.md": "state\n"}, "hand-off")
        out = subprocess.run([sys.executable, os.path.join(HERE, "review_gate.py"), "--effective-head", str(N),
                              "--base", self.base, "--repo", self.repo], capture_output=True, text=True)
        self.assertEqual((out.returncode, out.stdout.strip()), (0, code), out.stderr)

    def test_the_label_reads_every_commit_and_not_the_last_one(self):
        self.commit({"go/a.go": "package a\n"}, "code")
        self.commit({"docs/a.md": "notes\n"}, "docs")
        status, out = self.gate(labels=[rg.LABEL], enabled=True)
        self.assertEqual(status, 1, out)
        self.assertIn("`go/a.go`", out)
        self.assertIn("2 changed path(s)", out)

    def test_the_label_passes_documents_after_the_period_alone(self):
        self.commit({"docs/decisions.md": "| D-1 |\n"}, "docs")
        status, out = self.gate(labels=[rg.LABEL])
        self.assertEqual(status, 1, out)
        self.assertIn("RG 1: SKIP", out)
        self.assertIn("RG 3: FAULT", out)
        status, out = self.gate(labels=[rg.LABEL], enabled=True)
        self.assertEqual(status, 0, out)

    def test_the_command_fails_a_dependabot_pull_request_with_no_record(self):
        self.commit({"go/go.mod": "module a\n"}, "bump", email="49699333+dependabot[bot]@users.noreply.github.com", committer=rg.GITHUB_COMMITTER)
        status, out = self.gate(author="dependabot[bot]")
        self.assertEqual(status, 1, out)
        self.assertIn("RG 3: FAULT", out)


class RecordTemplate(unittest.TestCase):
    def test_a_filled_skeleton_passes_the_reference_check(self):
        refs = importlib.util.spec_from_file_location("ref_check", os.path.join(HERE, "ref_check.py"))
        rc = importlib.util.module_from_spec(refs)
        refs.loader.exec_module(rc)
        doc = ".claude/skills/pr-review/references/review-record.md"
        with open(os.path.join(rg.ROOT, doc), encoding="utf-8") as handle:
            text = handle.read()
        skeleton = text.split("## The skeleton", 1)[1].split("```markdown\n", 1)[1].split("\n```", 1)[0]
        record_text = skeleton.replace("<number>", "212")
        known = rc.registers(rc.read(rc.DECISIONS), rc.read(rc.QUESTIONS))
        findings = [f for f in rc.check("docs/reviews/pr-212.md", record_text, known, set(), set(), set()) if f[1] == "REF 1"]
        self.assertEqual(findings, [])


class Workflow(unittest.TestCase):
    def test_the_workflow_runs_this_file_from_the_base_and_never_the_head(self):
        with open(os.path.join(rg.ROOT, ".github/workflows/review-gate.yml"), encoding="utf-8") as handle:
            text = handle.read()
        self.assertIn("pull_request_target:", text)
        self.assertIn("python3 docs/tools/review_gate.py", text)
        self.assertIn("name: review-gate", text)
        self.assertNotIn("ref: ${{ github.event.pull_request.head", text)
        self.assertNotIn("write", text.split("permissions:", 1)[1].split("\n\n", 1)[0])

    def test_a_new_base_runs_the_gate_again(self):
        with open(os.path.join(rg.ROOT, ".github/workflows/review-gate.yml"), encoding="utf-8") as handle:
            text = handle.read()
        types = text.split("types: [", 1)[1].split("]", 1)[0]
        self.assertIn("edited", [t.strip() for t in types.split(",")])
        # A skipped job reads as a pass of a required check, so the job
        # takes no condition.
        self.assertNotIn("    if:", text)


if __name__ == "__main__":
    unittest.main()
