# Gym Route

Gym Route is a personal workout app for one person, its owner. The owner describes the muscles to train, the schedule, and the experience. The owner photographs the machines of the gym. The app identifies each machine with the owner's confirmation, builds a workout plan, guides each workout, records each set, and adapts the next targets.

OpenAI `gpt-6-luna` proposes each plan and each revision. A deterministic, versioned policy checks every set and load before the owner sees it. Gym Route is an installable, phone-first web app on Google Cloud. It never goes to an app store.

## Status

**No app, service, or cloud resource exists yet.** The repository holds the foundation documents, the research, the roadmaps, and the process tooling. Phase 0 is complete, and Phase 1 (risk spikes) is in progress.

## Repository map

| Path | Content |
|---|---|
| `AGENTS.md` | The rules for every agent and session. Read it first. |
| `CLAUDE.md` | A pointer to `AGENTS.md` for Claude Code |
| `docs/session-handoff.md` | The resume point of the next session |
| `docs/design.md` | Product thesis, experience, system context, safety, privacy |
| `docs/decisions.md` | Every owner decision, with a date |
| `docs/questions.md` | Every question, its answer, and the open questions |
| `docs/roadmaps/` | The high-level roadmap and the roadmap rules |
| `docs/research/` | Research with sources and dates |
| `docs/reviews/` | Codex review records |
| `docs/tools/` | Document and pull request checks, with tests |
| `.claude/skills/` | Skills for writing, pull request work, and review |
| `.github/` | CI workflows, the pull request template, and the ruleset of `main` |
| `.githooks/` | Git hooks that refuse a commit on `main` and AI attribution |

## Setup

Only the documentation checks exist. They need Python 3, Git, and GitHub CLI.

```bash
make hooks    # install the Git hooks once in each checkout
make verify   # run every free check that CI runs
make where    # print the branch, the tree, and the pull request state
```

`make codex-review PR=<n>` starts a Codex review of a pull request that Claude Code writes. `make claude-review PR=<n>` starts a Claude Code review of a pull request that Codex writes. Each spends a plan of the owner.

## License

MIT. See `LICENSE`.
