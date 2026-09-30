"""Tests of the reference check (D-6).

Run: python3 -m unittest discover -s docs/tools -p 'test_*.py'
"""
import importlib.util
import os
import subprocess
import sys
import tempfile
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
spec = importlib.util.spec_from_file_location("ref_check", os.path.join(HERE, "ref_check.py"))
rc = importlib.util.module_from_spec(spec)
spec.loader.exec_module(rc)

DECISIONS = "| # | Date |\n|---|---|\n| D-1 | a |\n| D-2 (amended by D-3) | b |\n| D-3 | c |\n"
QUESTIONS = "| # | Question |\n|---|---|\n| Q-1 | a |\n| Q-7 | b |\n"

PATHS = {"docs/decisions.md", "docs/questions.md", "docs/tools/ref_check.py",
         ".claude/skills/one-pr-one-session/SKILL.md"}
FOLDERS = {"docs", "docs/tools", ".claude", ".claude/skills", ".claude/skills/one-pr-one-session"}
TOP = {"docs", ".claude"}


def run(text, doc="docs/note.md"):
    known = rc.registers(DECISIONS, QUESTIONS)
    return rc.check(doc, text, known, TOP, PATHS, FOLDERS)


class Ids(unittest.TestCase):
    def test_a_defined_id_passes(self):
        self.assertEqual(run("The rule of D-2 and Q-7."), [])

    def test_an_amended_row_defines_its_id(self):
        self.assertEqual(run("D-2 holds, and D-3 amends it."), [])

    def test_an_undefined_decision_fails(self):
        self.assertEqual(run("The rule of D-9."), [(1, "REF 1", "no register defines D-9")])

    def test_an_undefined_question_fails(self):
        self.assertEqual(run("The answer of Q-2."), [(1, "REF 1", "no register defines Q-2")])

    def test_a_decision_row_does_not_define_a_question(self):
        known = rc.registers("| Q-5 | a |\n", "| D-5 | b |\n")
        self.assertEqual(known, set())

    def test_an_id_of_the_other_repository_takes_no_rule(self):
        self.assertEqual(run("The port of decktome:D-811 and decktome:D-837, in (decktome:D-9)."), [])

    def test_both_forms_of_the_other_repository_take_no_rule(self):
        text = "Decktome holds `decktome:AGENTS.md` and decktome:D-936, and `decktome:D-811` too."
        self.assertEqual(run(text), [])
        # The prefix covers the one token after it alone.
        self.assertEqual(run("See decktome:D-936 and D-9."), [(1, "REF 1", "no register defines D-9")])

    def test_the_other_role_model_repositories_are_skipped(self):
        self.assertEqual(run("It follows what-you-carry:D-198 and the-thing-below:D-8."), [])
        self.assertEqual(run("It follows what-you-carry D-198."), [(1, "REF 1", "no register defines D-198")])

    def test_roadmap_labels_are_no_register_ids(self):
        self.assertEqual(run("Work area 2.1 of Phase 3 holds Scenario A and step M-4, F-2, PR-7."), [])

    def test_a_placeholder_id_takes_no_rule(self):
        self.assertEqual(run("Cite the row as D-<n>."), [])

    def test_the_line_number_reads_the_source(self):
        self.assertEqual(run("A line.\n\nThe rule of D-9.\n")[0][0], 3)

    def test_a_fenced_block_takes_no_rule(self):
        self.assertEqual(run("Run this:\n\n```\nD-9 `docs/absent.md`\n```\n"), [])


class Paths(unittest.TestCase):
    def test_a_live_path_passes(self):
        self.assertEqual(run("Read `docs/tools/ref_check.py` and `docs/tools`."), [])

    def test_a_dead_path_fails(self):
        findings = run("Read `docs/tools/absent.py`.")
        self.assertEqual([f[1] for f in findings], ["REF 2"])

    def test_a_dotted_path_reads_the_rule(self):
        self.assertEqual([f[1] for f in run("Load `.claude/skills/absent/SKILL.md`.")], ["REF 2"])

    def test_a_live_dotted_path_passes(self):
        self.assertEqual(run("Load `.claude/skills/one-pr-one-session/SKILL.md`."), [])

    def test_a_path_resolves_from_the_folder_of_the_document(self):
        self.assertEqual(run("Read `tools/ref_check.py`.", doc="docs/note.md"), [])

    def test_a_placeholder_path_takes_no_rule(self):
        self.assertEqual(run("The record is `docs/reviews/pr-<n>.md`, and the build lands at `docs/X`."), [])

    def test_a_path_of_the_other_repository_takes_no_rule(self):
        self.assertEqual(run("Compare `decktome:docs/design-roadmap.md` and decktome:docs/absent.md."), [])

    def test_a_branch_name_takes_no_rule(self):
        self.assertEqual(run("Work on the branch `docs/foundation-roadmap`.\nBranch: `docs/foundation-roadmap`. Role: author."), [])

    def test_a_dead_path_after_other_words_still_fails(self):
        findings = run("The branch holds `docs/absent-note.md` and `docs/foundation-roadmap`.")
        self.assertEqual([f[1] for f in findings], ["REF 2", "REF 2"])

    def test_a_bare_name_takes_no_rule(self):
        self.assertEqual(run("The file `absent.json` holds the rows."), [])

    def test_a_path_outside_the_repository_takes_no_rule(self):
        self.assertEqual(run("Read `backend/app/models.py`."), [])

    def test_a_scratch_path_takes_no_rule(self):
        self.assertEqual(run("The run wrote `docs/.local/codex-review/pr-1.jsonl`."), [])

    def test_a_review_record_reads_no_path_rule(self):
        text = "The record read `docs/moved.md` under D-9."
        self.assertEqual(run(text, doc="docs/reviews/pr-9.md"), [(1, "REF 1", "no register defines D-9")])
        self.assertEqual(len(run(text, doc="docs/review-notes.md")), 2)

    def test_a_dated_record_is_exempt(self):
        self.assertTrue(rc.DATED.search("docs/research/note-2026-09-17.md"))
        self.assertTrue(rc.DATED.search("docs/session-handoff-archive.md"))
        self.assertIsNone(rc.DATED.search("docs/decisions.md"))


class RepoPaths(unittest.TestCase):
    def test_the_git_file_of_a_worktree_is_no_top_level_entry(self):
        with tempfile.TemporaryDirectory() as root:
            with open(os.path.join(root, ".git"), "w", encoding="utf-8") as handle:
                handle.write("gitdir: /elsewhere/.git/worktrees/x\n")
            os.makedirs(os.path.join(root, "docs"))
            with open(os.path.join(root, "docs", "a.md"), "w", encoding="utf-8") as handle:
                handle.write("a\n")
            saved, rc.ROOT = rc.ROOT, root
            try:
                paths, folders = rc.repo_paths()
            finally:
                rc.ROOT = saved
        self.assertEqual(paths, {"docs/a.md"})
        self.assertEqual(folders, {"docs"})


class Command(unittest.TestCase):
    """The command over a fixture checkout: one passing document and one broken document."""

    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        root = self.tmp.name
        os.makedirs(os.path.join(root, "docs", "tools"))
        files = {
            "docs/decisions.md": DECISIONS,
            "docs/questions.md": QUESTIONS,
            "docs/good.md": "D-1 and Q-1 hold. Read `docs/decisions.md`.\n",
            "docs/bad.md": "D-40 holds. Read `docs/absent.md`.\n",
        }
        for path, text in files.items():
            with open(os.path.join(root, path), "w", encoding="utf-8") as handle:
                handle.write(text)
        with open(os.path.join(HERE, "ref_check.py"), encoding="utf-8") as handle:
            tool = handle.read()
        with open(os.path.join(root, "docs", "tools", "ref_check.py"), "w", encoding="utf-8") as handle:
            handle.write(tool)

    def tearDown(self):
        self.tmp.cleanup()

    def run_tool(self, *docs):
        tool = os.path.join(self.tmp.name, "docs", "tools", "ref_check.py")
        return subprocess.run([sys.executable, tool, *docs], cwd=self.tmp.name, capture_output=True, text=True)

    def test_a_good_document_passes(self):
        out = self.run_tool("docs/good.md")
        self.assertEqual(out.returncode, 0, out.stdout)

    def test_a_broken_document_fails_with_both_rules(self):
        out = self.run_tool("docs/good.md", "docs/bad.md")
        self.assertEqual(out.returncode, 1, out.stdout)
        self.assertIn("docs/bad.md:1: rule REF 1: no register defines D-40", out.stdout)
        self.assertIn("docs/bad.md:1: rule REF 2: no file and no folder holds `docs/absent.md`", out.stdout)


if __name__ == "__main__":
    unittest.main()
