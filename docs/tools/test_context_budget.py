"""Tests of the context budget check (D-6).

The tests read fixtures alone, and never the real rule files or hand-off.

Run: python3 -m unittest discover -s docs/tools -p 'test_*.py'
"""
import importlib.util
import os
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
spec = importlib.util.spec_from_file_location("context_budget", os.path.join(HERE, "context_budget.py"))
cb = importlib.util.module_from_spec(spec)
spec.loader.exec_module(cb)

PAID_LINE = "The one paid target is `make codex-review PR=<n>`. It spends the Codex plan of the owner."
AGENTS = f"# Agent rules\n\n## Cost\n\n- {PAID_LINE}\n- The free `make verify` runs every check.\n"
CLAUDE = "Read `AGENTS.md` first.\n"


def record(n):
    return f"### 2026-09-2{n}: session {n}\n\nAuthor provider: Claude Code\n\nText.\n\n"


HANDOFF = "# Session hand-off\n\n## Resume here\n\nThe state.\n\n## Session records\n\n" + record(1) + record(2)
MAKEFILE = ("verify: ## Run every check, free\n\t@true\n"
            "codex-review: ## Start the review. CAUTION: it spends the Codex plan of the owner\n\t@true\n")
SKILL = ".claude/skills/one-pr-one-session/SKILL.md"


def files(**override):
    base = {cb.AGENTS: AGENTS, cb.CLAUDE: CLAUDE, cb.HANDOFF: HANDOFF, cb.MAKEFILE: MAKEFILE}
    base.update(override)
    return base


def run(texts, skills=()):
    return cb.check(lambda path: texts.get(path), skills)


def has(errors, text):
    return any(text in e for e in errors)


class Files(unittest.TestCase):
    def test_small_files_pass(self):
        report, errors = run(files())
        self.assertEqual(errors, [])
        self.assertIn("paid targets: 1 in the Makefile, 1 in AGENTS.md", report)
        self.assertIn(f"{cb.HANDOFF} session records: 2 of 3", report)

    def test_each_limit_fails_one_byte_over(self):
        for path in (cb.AGENTS, cb.CLAUDE, cb.HANDOFF):
            with self.subTest(path=path):
                texts = files()
                texts[path] = texts[path] + "x" * (cb.FILE_LIMITS[path] - len(texts[path].encode()) + 1)
                _, errors = run(texts)
                self.assertTrue(has(errors, f"{path} holds"), errors)

    def test_the_limits_of_the_owner(self):
        self.assertEqual(cb.FILE_LIMITS, {"AGENTS.md": 11000, "CLAUDE.md": 1000, "docs/session-handoff.md": 24000})
        self.assertEqual((cb.RESUME_LIMIT, cb.SESSIONS_LIMIT, cb.SKILL_LIMIT), (6000, 3, 36864))

    def test_an_absent_file_fails(self):
        _, errors = run(files(**{cb.AGENTS: None, cb.HANDOFF: None}))
        self.assertIn("AGENTS.md does not exist", errors)
        self.assertIn("docs/session-handoff.md does not exist", errors)


class Handoff(unittest.TestCase):
    def test_a_resume_section_over_its_limit_fails(self):
        _, errors = run(files(**{cb.HANDOFF: HANDOFF.replace("The state.", "y" * (cb.RESUME_LIMIT + 1))}))
        self.assertTrue(has(errors, "the resume section holds"), errors)

    def test_the_resume_limit_reads_its_own_section_alone(self):
        _, errors = run(files(**{cb.HANDOFF: HANDOFF + "z" * (cb.RESUME_LIMIT + 1)}))
        self.assertFalse(has(errors, "the resume section"), errors)

    def test_no_resume_section_fails(self):
        _, errors = run(files(**{cb.HANDOFF: "# Session hand-off\n\nNo sections.\n"}))
        self.assertTrue(has(errors, "no '## Resume here' section"), errors)

    def test_the_heading_can_carry_a_date(self):
        _, errors = run(files(**{cb.HANDOFF: HANDOFF.replace("## Resume here", "## Resume here (2026-09-28)")}))
        self.assertEqual(errors, [])

    def test_four_session_records_fail(self):
        _, errors = run(files(**{cb.HANDOFF: HANDOFF + record(3) + record(4)}))
        self.assertTrue(has(errors, "4 session records"), errors)

    def test_three_session_records_pass(self):
        _, errors = run(files(**{cb.HANDOFF: HANDOFF + record(3)}))
        self.assertEqual(errors, [])


class PaidTargets(unittest.TestCase):
    def test_a_caution_target_absent_from_agents_fails(self):
        makefile = MAKEFILE + "load-test: ## Run the load test. CAUTION: it spends money\n\t@true\n"
        _, errors = run(files(**{cb.MAKEFILE: makefile}))
        self.assertTrue(has(errors, "`make load-test` says CAUTION"), errors)

    def test_agents_with_no_paid_sentence_fails(self):
        _, errors = run(files(**{cb.AGENTS: "# Agent rules\n\nNo cost here.\n"}))
        self.assertTrue(has(errors, "holds no sentence with the words 'paid target'"), errors)

    def test_a_named_target_that_does_not_exist_fails(self):
        agents = AGENTS.replace("`make codex-review PR=<n>`", "`make codex-review` and `make deploy`")
        _, errors = run(files(**{cb.AGENTS: agents}))
        self.assertTrue(has(errors, "`make deploy` of AGENTS.md is no Makefile target"), errors)

    def test_a_named_free_target_fails(self):
        agents = AGENTS.replace("`make codex-review PR=<n>`", "`make codex-review` and `make verify`")
        _, errors = run(files(**{cb.AGENTS: agents}))
        self.assertTrue(has(errors, "says no CAUTION"), errors)

    def test_the_names_stop_at_the_end_of_the_sentence(self):
        self.assertEqual(cb.paid_names("Paid targets: `make a-b`. The free `make c-d` stays out.\n"), {"a-b"})

    def test_a_wrapped_sentence_reads_whole(self):
        self.assertEqual(cb.paid_names("The one paid\ntarget is `make codex-review`.\n"), {"codex-review"})

    def test_the_caution_list_reads_the_help_text(self):
        self.assertEqual(cb.caution_targets(MAKEFILE), {"codex-review"})


class Skills(unittest.TestCase):
    def test_a_small_skill_passes(self):
        texts = files(**{SKILL: "x"})
        report, errors = run(texts, skills=[SKILL])
        self.assertEqual(errors, [])
        self.assertIn(f"skill files: 1 read, 0 over {cb.SKILL_LIMIT} bytes", report)

    def test_a_skill_over_the_limit_fails(self):
        _, errors = run(files(**{SKILL: "x" * (cb.SKILL_LIMIT + 1)}), skills=[SKILL])
        self.assertTrue(has(errors, SKILL), errors)


if __name__ == "__main__":
    unittest.main()
