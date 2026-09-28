The milestone of this pull request, in two to four sentences.

## Session

- Role: author
- Branch: `branch-name`
- Base: `0000000`

## Milestone

A pull request holds one cohesive milestone with one acceptance story. The milestone can hold two, three, or more concerns when the scope needs them (D-10, D-12).

- Concerns: <each concern of the milestone, separated by commas>
- Acceptance story: <one story that proves the whole milestone>
- Owner approval: <where and when the owner approved this work before it started>

## Documentation impact

Each entry starts with `Changed:`, `Reviewed; no change needed:`, or `Not applicable:`. Name each path in backticks, and give a specific reason. `make pr-check` reads this table.

| Category | Entry |
|---|---|
| Session hand-off | Changed: `docs/session-handoff.md` <reason> |
| Agent rules | <status> `AGENTS.md` `CLAUDE.md` <reason> |
| README | <status> `README.md` <reason> |
| Design | <status> `docs/design.md` <reason> |
| Decisions | <status> `docs/decisions.md` <reason> |
| Questions | <status> `docs/questions.md` <reason> |
| Research | <status> `docs/research/` <reason> |
| Roadmaps | <status> `docs/roadmaps/` <reason> |
| Skills and hooks | <status> `.claude/` <reason> |
| Tools and CI | <status> `docs/tools/` `.github/` `.githooks/` `Makefile` <reason> |

## Checks

The commands that you ran, and their results. A check that did not run gets its reason.

## Review

The state of the Codex review record at `docs/reviews/pr-<n>.md`, and the answer to each finding (D-4, D-8).
