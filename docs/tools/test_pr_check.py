"""Tests of the one-pr-one-session contract check (D-10, D-12, D-14).

Run: python3 -m unittest discover -s docs/tools -p 'test_*.py'
"""
import importlib.util
import json
import os
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
spec = importlib.util.spec_from_file_location("pr_check", os.path.join(HERE, "pr_check.py"))
pc = importlib.util.module_from_spec(spec)
spec.loader.exec_module(pc)

TITLE = "feat(plan): the policy checks each planned load (PR-4)"
BRANCH = "feat/pr-4-policy-load-check"

SESSION = f"""## Session

- Role: author
- Branch: `{BRANCH}`
- Base: `372d912`
"""

MILESTONE = """## Milestone

- Concerns: the load policy, its tests, and the design section
- Acceptance story: a planned set over the load limit never reaches the user, and the test proves it
- Owner approval: the owner approved the milestone in the session of 2026-09-28
"""

ROWS = {
    "Session hand-off": "Changed: `docs/session-handoff.md` records the finished state and the next step",
    "Agent rules": "Reviewed; no change needed: `AGENTS.md` names no command or rule that this change moves",
    "README": "Reviewed; no change needed: `README.md` names no feature that the policy changes",
    "Design": "Changed: `docs/design.md` describes the load policy that the code adds",
    "Decisions": "Reviewed; no change needed: `docs/decisions.md` holds every owner answer this change needs",
    "Questions": "Reviewed; no change needed: `docs/questions.md` holds no question that this change answers",
    "Research": "Not applicable: the change cites no research and adds no research document",
    "Roadmaps": "Reviewed; no change needed: `docs/roadmaps/` names the same work area after this change",
    "Skills and hooks": "Reviewed; no change needed: `.claude/` holds no rule that the policy code changes",
    "Tools and CI": "Changed: `Makefile` adds the policy target that the verify job runs",
}

CHANGED = [
    "go/internal/policy/load.go",
    "go/internal/policy/load_test.go",
    "docs/session-handoff.md",
    "docs/design.md",
    "Makefile",
]
HANDOFF = "## Resume here\n\nState.\n\n### 2026-09-28: the load policy\n\nAuthor provider: Claude Code\n"


def body(rows=None, extra="", milestone=MILESTONE):
    rows = dict(ROWS, **(rows or {}))
    table = "\n".join(f"| {k} | {v} |" for k, v in rows.items() if v is not None)
    return f"Summary.\n\n{SESSION}\n{milestone}\n## Documentation impact\n\n| Category | Entry |\n|---|---|\n{table}\n\n{extra}"


def exists(path):
    return True


def check(text, changed=CHANGED, title=TITLE, author="owner", handoff_added="", handoff_text=HANDOFF):
    return pc.check_pr(title, text, author, BRANCH, changed, exists, lambda sha: True, handoff_added, handoff_text)


def has(errors, text):
    return any(text in e for e in errors)


class PullRequestContract(unittest.TestCase):
    def test_a_complete_pull_request_passes(self):
        self.assertEqual(check(body()), [])

    def test_the_template_categories_match_the_tool(self):
        self.assertEqual([pc.clean_name(k) for k in ROWS], [name for name, _ in pc.CATEGORIES])

    def test_a_missing_handoff_change_fails(self):
        errors = check(body(), changed=[p for p in CHANGED if p != pc.HANDOFF])
        self.assertTrue(has(errors, "does not change docs/session-handoff.md"), errors)

    def test_a_handoff_marked_unchanged_fails(self):
        errors = check(body({"Session hand-off": "Reviewed; no change needed: `docs/session-handoff.md` reads the right next step"}))
        self.assertTrue(has(errors, "every pull request changes"), errors)

    def test_a_provider_line_can_name_the_company(self):
        self.assertEqual(check(body(), handoff_text=HANDOFF.replace("Claude Code", "Claude Code (Anthropic)")), [])
        self.assertEqual(check(body(), handoff_text=HANDOFF.replace("Claude Code", "Codex (OpenAI)")), [])

    def test_a_provider_line_with_another_name_fails(self):
        errors = check(body(), handoff_text=HANDOFF.replace("Claude Code", "an assistant"))
        self.assertTrue(has(errors, "Author provider"), errors)

    def test_a_handoff_with_no_provider_line_fails(self):
        errors = check(body(), handoff_text="## Resume here\n\nState.\n")
        self.assertTrue(has(errors, "Author provider"), errors)

    def test_a_missing_category_fails(self):
        errors = check(body({"Research": None}))
        self.assertTrue(has(errors, "no row for the category 'research'"), errors)

    def test_a_generic_reason_fails(self):
        errors = check(body({"Questions": "Reviewed; no change needed: `docs/questions.md` no documentation impact"}))
        self.assertTrue(has(errors, "generic"), errors)

    def test_a_short_reason_fails(self):
        self.assertTrue(has(check(body({"Research": "Not applicable: none"})), "generic"))
        self.assertTrue(has(check(body({"Research": "Not applicable: no research here"})), "words"))

    def test_a_placeholder_fails(self):
        errors = check(body({"Questions": "Reviewed; no change needed: `docs/questions.md` <reason> TODO later today"}))
        self.assertTrue(has(errors, "placeholder"), errors)

    def test_a_pending_status_fails(self):
        errors = check(body({"Roadmaps": "Pending: a clean author session completes this row"}))
        self.assertTrue(has(errors, "starts with none of"), errors)

    def test_an_unchanged_claim_on_a_changed_document_fails(self):
        errors = check(body({"Design": "Reviewed; no change needed: `docs/design.md` holds every contract of the policy"}))
        self.assertTrue(has(errors, "the diff changes docs/design.md"), errors)

    def test_a_changed_claim_with_no_change_fails(self):
        errors = check(body({"Questions": "Changed: `docs/questions.md` closes the question that this change answers"}))
        self.assertTrue(has(errors, "changes no file of this category"), errors)

    def test_a_makefile_row_names_a_path(self):
        errors = check(body({"Tools and CI": "Changed: the Makefile adds the policy target of the verify job"}))
        self.assertTrue(has(errors, "names no path in backticks"), errors)

    def test_deferred_documentation_in_the_body_fails(self):
        errors = check(body(extra="The roadmap will update after merge in a follow-up PR."))
        self.assertTrue(has(errors, "defers documentation"), errors)

    def test_a_negated_follow_up_passes(self):
        self.assertEqual(check(body(extra="No follow-up pull request records the documents.")), [])

    def test_deferred_documentation_in_the_handoff_fails(self):
        errors = check(body(), handoff_added="A separate PR updates the roadmap after the merge.")
        self.assertTrue(has(errors, "hand-off defers"), errors)

    def test_a_merge_record_pull_request_fails(self):
        changed = ["docs/session-handoff.md", "docs/roadmaps/high-level-roadmap.md"]
        rows = {
            "Design": "Reviewed; no change needed: `docs/design.md` holds no fact of the earlier merge",
            "Roadmaps": "Changed: `docs/roadmaps/high-level-roadmap.md` reads the merge of the earlier pull request",
            "Tools and CI": "Reviewed; no change needed: `Makefile` holds no target that the record changes",
        }
        errors = check(body(rows), changed=changed, title="docs: record the merge of #12")
        self.assertTrue(has(errors, "records an earlier merge"), errors)

    def test_no_roadmap_merge_mark_is_needed(self):
        # D-9: the roadmap holds no permanent pull request numbers.
        changed = CHANGED + ["docs/roadmaps/high-level-roadmap.md"]
        rows = {"Roadmaps": "Changed: `docs/roadmaps/high-level-roadmap.md` moves the work area to its next state"}
        self.assertEqual(check(body(rows), changed=changed), [])


class Title(unittest.TestCase):
    def test_conventional_titles_pass(self):
        for title in ("docs: add the roadmap", "feat(api)!: drop the old route", "ci(verify): pin the actions", "fix: round the load"):
            with self.subTest(title=title):
                self.assertEqual(pc.check_title(title), [])

    def test_other_titles_fail(self):
        for title in ("Add the roadmap", "docs add the roadmap", "Docs: add the roadmap", "feature: a thing", "docs:add"):
            with self.subTest(title=title):
                self.assertTrue(pc.check_title(title))

    def test_a_bad_title_fails_the_pull_request(self):
        self.assertTrue(has(check(body(), title="The load policy"), "Conventional Commits"))


class Milestone(unittest.TestCase):
    def test_no_milestone_section_fails(self):
        self.assertTrue(has(check(body(milestone="")), "no '## Milestone' section"))

    def test_each_field_is_required(self):
        for field in pc.MILESTONE_FIELDS:
            with self.subTest(field=field):
                text = "\n".join(line for line in MILESTONE.splitlines() if not line.startswith(f"- {field}:"))
                self.assertTrue(has(check(body(milestone=text + "\n")), f"no '{field}:' line"))

    def test_a_template_placeholder_fails(self):
        text = MILESTONE.replace("the load policy, its tests, and the design section", "<each concern of the milestone>")
        self.assertTrue(has(check(body(milestone=text)), "placeholder"))

    def test_a_short_story_fails(self):
        text = MILESTONE.replace("a planned set over the load limit never reaches the user, and the test proves it", "it works")
        self.assertTrue(has(check(body(milestone=text)), "acceptance story holds 2 words"))

    def test_two_stories_fail(self):
        text = MILESTONE + "- Acceptance story: a second story that the pull request also tells\n"
        self.assertTrue(has(check(body(milestone=text)), "one milestone"))


class Attribution(unittest.TestCase):
    def test_a_co_author_trailer_of_an_ai_fails(self):
        for line in ("Co-Authored-By: Claude Opus <noreply@anthropic.com>", "co-authored-by: Codex <codex@openai.com>",
                     "Co-authored-by: ChatGPT <x@y>"):
            with self.subTest(line=line):
                self.assertTrue(has(check(body(extra=line)), "AI attribution"))

    def test_a_generated_with_line_fails(self):
        for line in ("🤖 Generated with [Claude Code](https://claude.com/claude-code)", "Generated by Codex"):
            with self.subTest(line=line):
                self.assertTrue(has(check(body(extra=line)), "AI attribution"))

    def test_a_human_co_author_and_plain_text_pass(self):
        extra = "Co-Authored-By: A Person <person@example.com>\n\nThe code generated with buf stays committed."
        self.assertEqual(check(body(extra=extra)), [])


class Session(unittest.TestCase):
    def test_the_session_block_is_required(self):
        self.assertTrue(has(check(body().replace(SESSION, "")), "no '## Session' section"))

    def test_the_branch_must_match_the_head(self):
        errors = check(body().replace(f"`{BRANCH}`", "`feat/other`"))
        self.assertTrue(has(errors, f"pull request head is {BRANCH}"), errors)

    def test_the_base_must_be_an_ancestor(self):
        errors = pc.check_pr(TITLE, body(), "owner", BRANCH, CHANGED, exists, lambda sha: False, "", HANDOFF)
        self.assertTrue(has(errors, "not an ancestor"), errors)

    def test_the_reviewer_role_passes(self):
        self.assertEqual(check(body().replace("Role: author", "Role: reviewer")), [])

    def test_dependabot_is_exempt(self):
        self.assertEqual(check("", changed=["web/package.json"], title="Bump x", author="dependabot[bot]"), [])


def wiring(**override):
    """A fixture checkout that passes the skill check. A value of None removes the file."""
    skill = f"---\nname: one-pr-one-session\ndescription: one session\n---\n{pc.BLOCKED}\n`{pc.COMPLETE}`\n"
    rows = "\n".join(f"| {name} | x |" for name, _ in pc.CATEGORIES)
    files = {
        pc.SKILL: skill,
        ".claude/skills/ste-writing/SKILL.md": "---\nname: ste-writing\ndescription: STE\n---\nbody\n",
        pc.GITAR_SKILL: "---\nname: gitar-review\ndescription: Gitar\n---\nAsk Gitar for a pass.\n",
        pc.AGENTS: f"Load `{pc.SKILL}` first. Load `gitar-review` after each push.\n",
        pc.CLAUDE: "Read `AGENTS.md`.\n",
        pc.TEMPLATE: f"## Session\n\n## Milestone\n\n## Documentation impact\n\n| Category | Entry |\n|---|---|\n{rows}\n",
        ".claude/settings.json": json.dumps({"hooks": {"PreToolUse": [{"hooks": [{"command": "session_bind.py"}]}],
                                                       "PostToolUse": [{"hooks": [{"command": "context_checkpoint.py"}]}]}}),
        ".githooks/pre-commit": "#!/bin/sh\n",
        ".githooks/commit-msg": "#!/bin/sh\n",
    }
    files.update(override)
    files = {k: v for k, v in files.items() if v is not None}

    def listdir(path):
        prefix = path.rstrip("/") + "/"
        return sorted({p[len(prefix):].split("/", 1)[0] for p in files if p.startswith(prefix) and "/" in p[len(prefix):]})

    return files.get, listdir


class SkillWiring(unittest.TestCase):
    def test_the_fixture_passes(self):
        self.assertEqual(pc.check_skills(*wiring()), [])

    def test_a_skill_without_the_block_text_fails(self):
        errors = pc.check_skills(*wiring(**{pc.SKILL: "---\nname: one-pr-one-session\ndescription: x\n---\nbody"}))
        self.assertTrue(has(errors, "exact text"), errors)

    def test_agents_must_require_the_skill(self):
        self.assertTrue(has(pc.check_skills(*wiring(**{pc.AGENTS: "Rules.\n"})), "AGENTS.md does not require"))

    def test_claude_must_point_to_agents(self):
        self.assertTrue(has(pc.check_skills(*wiring(**{pc.CLAUDE: "Rules live here.\n"})), "does not point to AGENTS.md"))

    def test_absent_rule_files_fail(self):
        errors = pc.check_skills(*wiring(**{pc.AGENTS: None, pc.CLAUDE: None}))
        self.assertTrue(has(errors, "AGENTS.md does not require"), errors)
        self.assertTrue(has(errors, "CLAUDE.md does not point"), errors)

    def test_a_skill_that_names_gitar_passes(self):
        text = "---\nname: ste-writing\ndescription: STE\n---\nAsk Gitar for a pass.\n"
        self.assertEqual(pc.check_skills(*wiring(**{".claude/skills/ste-writing/SKILL.md": text})), [])

    def test_an_absent_gitar_skill_fails(self):
        self.assertTrue(has(pc.check_skills(*wiring(**{pc.GITAR_SKILL: None})), "gitar-review/SKILL.md does not exist"))

    def test_agents_must_name_the_gitar_skill(self):
        errors = pc.check_skills(*wiring(**{pc.AGENTS: f"Load `{pc.SKILL}` first.\n"}))
        self.assertTrue(has(errors, "does not name the `gitar-review` skill"), errors)

    def test_a_bad_frontmatter_fails(self):
        text = "---\nname: other\ndescription: <x>\ncolor: red\n---\nbody\n"
        errors = pc.check_skills(*wiring(**{".claude/skills/ste-writing/SKILL.md": text}))
        self.assertTrue(has(errors, "is not the directory name"), errors)
        self.assertTrue(has(errors, "angle brackets"), errors)
        self.assertTrue(has(errors, "unknown frontmatter key 'color'"), errors)

    def test_the_template_needs_each_section_and_row(self):
        errors = pc.check_skills(*wiring(**{pc.TEMPLATE: "## Session\n"}))
        self.assertTrue(has(errors, "no '## Milestone' section"), errors)
        self.assertTrue(has(errors, "no matrix row for 'tools and ci'"), errors)

    def test_the_hooks_must_be_wired(self):
        errors = pc.check_skills(*wiring(**{".claude/settings.json": "{}", ".githooks/commit-msg": None}))
        self.assertTrue(has(errors, "session_bind.py hook on PreToolUse"), errors)
        self.assertTrue(has(errors, "context_checkpoint.py hook on PostToolUse"), errors)
        self.assertTrue(has(errors, ".githooks/commit-msg does not exist"), errors)


class RepositoryFiles(unittest.TestCase):
    """The files of this checkout that this tool owns. The rule files come from other writers."""

    def read(self, path):
        with open(os.path.join(pc.ROOT, path), encoding="utf-8") as handle:
            return handle.read()

    def test_the_template_holds_each_section_and_row(self):
        template = self.read(pc.TEMPLATE)
        rows = [pc.clean_name(c[0]) for c in pc.table_rows(pc.section(template, "Documentation impact"))]
        self.assertEqual(rows, [name for name, _ in pc.CATEGORIES])
        for heading in ("## Session", "## Milestone", "## Documentation impact"):
            self.assertIn(heading, template)

    def test_the_skill_holds_the_exact_texts(self):
        skill = self.read(pc.SKILL)
        self.assertIn(pc.BLOCKED, skill)
        self.assertIn(pc.COMPLETE, skill)

    @unittest.skipUnless(os.path.exists(os.path.join(pc.ROOT, pc.AGENTS)) and os.path.exists(os.path.join(pc.ROOT, pc.CLAUDE)),
                         "AGENTS.md or CLAUDE.md does not exist yet")
    def test_the_repository_passes_the_skill_check(self):
        def read(path):
            try:
                return self.read(path)
            except OSError:
                return None

        def listdir(path):
            full = os.path.join(pc.ROOT, path)
            return [d for d in os.listdir(full) if os.path.isdir(os.path.join(full, d))]

        self.assertEqual(pc.check_skills(read, listdir), [])


if __name__ == "__main__":
    unittest.main()


class Naming(unittest.TestCase):
    """D-86: the form of the-thing-below for every pull request after #1."""

    def test_a_matching_title_and_branch_pass(self):
        self.assertEqual(pc.check_naming("feat: the rest timer (PR-7)", "feat/pr-7-rest-timer"), [])
        self.assertEqual(pc.check_naming("docs(roadmap): the phase 1 roadmap (PR-1)", "docs/pr-1-phase-one"), [])

    def test_the_branch_of_pr_1_is_exempt(self):
        self.assertEqual(pc.check_naming("docs: establish the foundation", pc.FIRST_BRANCH), [])

    def test_a_title_without_the_id_fails(self):
        errors = pc.check_naming("feat: the rest timer", "feat/pr-7-rest-timer")
        self.assertEqual(len(errors), 1)
        self.assertIn("D-86", errors[0])

    def test_a_branch_without_the_id_fails(self):
        for branch in ("feat/rest-timer", "feat/pr-7", "feat/pr-07-rest-timer", "feat/pr-7-Rest-Timer", "pr-7-rest"):
            with self.subTest(branch=branch):
                self.assertTrue(pc.check_naming("feat: the rest timer (PR-7)", branch))

    def test_a_different_id_or_type_fails(self):
        self.assertTrue(pc.check_naming("feat: the rest timer (PR-8)", "feat/pr-7-rest-timer"))
        self.assertTrue(pc.check_naming("fix: the rest timer (PR-7)", "feat/pr-7-rest-timer"))

    def test_the_check_runs_inside_the_contract(self):
        errors = pc.check_pr("feat: the rest timer", "", "owner", "feat/rest-timer", [], lambda p: True)
        self.assertTrue(any("D-86" in e for e in errors))

