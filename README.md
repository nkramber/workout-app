# Gym Route

Gym Route is a personal workout app for one person, its owner. The owner describes the muscles to train, the schedule, and the experience. The owner selects the machines of the gym from a catalog, or enters them as text, and confirms each one. The app builds a workout plan, guides each workout, records each set, and adapts the next targets.

OpenAI `gpt-6-luna` proposes each plan and each revision. A deterministic, versioned policy checks every set and load before the owner sees it. Gym Route is an installable, phone-first web app on Google Cloud. It never goes to an app store.

## Status

**The workout app does not exist yet.** The contract in `proto/` and the Go API skeleton in `go/` exist, and they run only on the local emulators. The development project `gym-route-dev` exists, with Firebase Hosting and Firebase Authentication only. The repository holds the foundation documents, the research, the roadmaps, and the process tooling. Phase 0 is complete. Phase 1 (risk spikes) gave its three reports. The Luna plan and the iPhone web platform got a go, and photo recognition got a no-go. Phase 2 (platform skeleton) has its focused roadmap, `docs/roadmaps/phase-2-platform-skeleton.md`.

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
| `tools/spikes/` | The Phase 1 spike harnesses and the recognition test set, with tests |
| `proto/` | The Connect-RPC contract |
| `go/` | The Go API, with the generated code in `go/gen`. `go/README.md` describes it. |
| `firebase.json`, `emulators/` | The local emulators and their pinned `firebase-tools` |
| `.claude/skills/` | Skills for writing, pull request work, and review |
| `.github/` | CI workflows, the pull request template, and the ruleset of `main` |
| `.githooks/` | Git hooks that refuse a commit on `main` and AI attribution |

## Setup

The documentation checks and the Phase 1 spike harnesses need Python 3, Git, and GitHub CLI. The iPhone probe of `tools/spikes/iphone_probe/` also needs Node 22. The product checks need Go 1.27.1, and the emulator tests also need Node 20 or later and Java 21 (D-130).

```bash
make hooks    # install the Git hooks once in each checkout
make verify   # run the free checks of the verify:lint and verify:test jobs
make where    # print the branch, the tree, and the pull request state
make probe    # build the iPhone probe and run its browser tests, needs Node 22
make contract       # buf lint, the generated code in Git, and buf breaking against main
make go-test        # gofmt, go mod tidy, go vet, and the Go unit tests
make emulator-test  # the Go tests over the Auth and Firestore emulators
```

`make codex-review PR=<n>` starts a Codex review of a pull request that Claude Code writes. `make claude-review PR=<n>` starts a Claude Code review of a pull request that Codex writes. Each spends a plan of the owner.

## License

MIT. See `LICENSE`.
