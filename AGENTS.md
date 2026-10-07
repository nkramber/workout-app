# Workout App - AGENTS

This file holds the rules for every agent and every session in this repository. `CLAUDE.md` points here (D-7). Read this file first. Then read `docs/session-handoff.md`. It tells you where the last session stopped.

## Project

Workout App is a personal workout app for one user, the owner (D-67). It is an installable, phone-first web app on Google Cloud, and it never goes to an app store (D-17). OpenAI `gpt-6-luna` proposes plans and revisions. A deterministic, versioned policy checks every set and load before the owner sees it (D-22, D-23).

Stage: Phase 8 of `docs/roadmaps/high-level-roadmap.md`, after the iPhone check of Phase 7 on 2026-10-05 (D-310). `docs/roadmaps/phase-8-personal-use-operations.md` gives the pull requests of Phase 8, and the owner dropped the four weeks of its exit (D-327). Luna runs at xhigh effort (D-253). Policy version 9 holds the limit of the other sets and the rotation of the muscle groups (D-306, D-328). The contract is in `proto/`, and the shell, the screens, and the outbox sync of the web client are in `web/`. The Go API with its calls and the revision, the domain model, the policy, the Luna role layer, and the stores are in `go/`.

The project `nk-workout-app-prod` holds Firestore, the Cloud Run service `api`, Hosting, and Auth, and Cloud Build deploys each merge to `main` (D-137). `docs/setup-gcp.md` and `docs/deploy-and-rollback.md` describe it. `docs/research/phase-2-check.md` holds the exit evidence of Phase 2, `docs/research/phase-3-check.md` holds the Luna evaluation of Phase 3, and `docs/research/luna-effort-check.md` compares two efforts. `docs/research/restore-drill.md` holds the restore drill, and `docs/research/policy-replay.md` holds the replay report. `docs/operations.md` holds the incident runbook, the policy replay, and the model change. `docs/design.md` holds the design.

**The repository is public.** Write no email address, no personal data, no photo, no workout log, and no secret into a file, an issue, or a pull request. The one exception is the author credit that the license of a test image requires (D-106).

## Hard rules from the owner

1. **Write scope.** This repository permits writes. Every other repository is read-only, Decktome included. Work follows the phase order of the high-level roadmap. Do not start a phase before the exit evidence of each phase before it holds.
2. **Write in ASD-STE100** (D-83). Load the `ste-writing` skill before you write a `.md` file. Run `make ste-check` before you commit one.
3. **Ask questions when you think of them.** Use `AskUserQuestion` in batches of three or fewer, with the recommended option first. Record each answer in `docs/decisions.md` and `docs/questions.md`. Never record a recommendation as a decision.
4. **Do the research.** Verify facts against primary sources, and record the date of each fact. Mark each claim that you can not verify as unverified.
5. **Keep the hand-off current.** Update `docs/session-handoff.md` inside each pull request before you call it ready (D-14).
6. **No AI attribution** in a pull request, branch name, commit message, or comment (D-14). The `commit-msg` hook refuses an attribution trailer. A review record and the `Author provider` line of the hand-off can name a provider.
7. **Every change starts on a branch** (D-14). Never commit to `main`, and never push to it. Run `make where` before each commit and push. Run `make hooks` once in each checkout. Use Conventional Commits for commit titles. Name a pull request `<type>: <summary> (PR-<n>)` on the branch `<type>/pr-<n>-<slug>` (D-86).
8. **Deploy from `main` alone** (D-14). Check the branch and the commit before every deploy.
9. **Review before merge.**
   - Gitar reviews every pull request, documents alone included (D-335). After each push, load the `gitar-review` skill, and run `make gitar-wait PR=<n>` in the background (D-338). Answer each Gitar finding, and resolve each thread.
   - After the Gitar pass and green CI, the author session runs the review of the other provider without a separate approval. A Claude Code author runs `make codex-review PR=<n>` (D-8). A Codex author runs `make claude-review PR=<n>` (D-88).
   - Codex reviews Claude Code work, and Claude Code reviews Codex work (D-15). A review by the same provider never counts.
   - The owner ended the D-4 period on 2026-09-29 (D-125). A pull request of documents alone can pass the gate with the `review-override` label (D-15).
   - The review targets refuse to start before the Gitar pass is complete (D-336). `SKIP_GITAR=1` skips that check only in a pause of Gitar that the owner states (D-337). The reviewer never replies to Gitar.
   - The third open round of one finding stops the loop, and the owner decides.
   - The gate can not prove the provider of a record or of an author. Before each merge, the owner reads the commit that last changed the record and the author provider (D-87, D-89).
   - A change of code after an approval needs a new review. No author is exempt, Dependabot included (D-90).
10. **The owner confirms every merge** (D-13). After the Codex approval, ask the owner with a summary in four sections: What, How, CI, and Codex review. Turn on the auto-merge only after the confirmation. Push no commit after the auto-merge turns on, because GitHub can merge it before Gitar reads it.
11. **Push back.** When two owner statements conflict, quote both and ask. When a request rests on a wrong premise, say so with the evidence.
12. **One pull request, one clean session** (D-12). The owner approves the work before it starts. A pull request holds one milestone with one acceptance story, and it can hold two, three, or more concerns (D-10). The session continues until the pull request merges, and a context checkpoint never ends it (D-85). Load `.claude/skills/one-pr-one-session/SKILL.md` for all work on a pull request.
13. **Keep command output small.** Count or list the matches first. Then read a bounded range. Show the full output of a failed test, build, or gate.
14. **Cite what exists.** Cite a D- or Q- id that a register defines, and a path that exists. `make ref-check` fails on either error. Cite a Decktome path or id with the `decktome:` prefix.

## Product guardrails

`docs/design.md` section 5 holds the full list. The four rules that code must never break:

- Every set, load, and change passes the policy before the owner sees it (D-23).
- No model id appears at a call site. Models come from the role layer (D-24).
- Logs, metrics, and error reports hold ids only (D-80).
- A paid AI run in development needs owner approval (D-25).

## Commands

Use the Makefile.

```bash
make help            # list every target
make doctor          # check the local tools
make verify          # every check that CI runs, free
make lint            # ste-check, ref-check, lifecycle-check, context-budget
make test            # the unit tests of docs/tools and of each spike of tools/spikes
make ste-check       # the STE check alone, free
make ref-check       # every cited id and path resolves, free
make pr-check        # the body and the diff of the pull request, free
make gitar-wait PR=n # wait for a current Gitar review after a push, free (D-338)
make where           # the branch, the tree, and the pull request state
make hooks           # install the Git hooks once in each checkout
make ruleset-check   # the live ruleset of main against .github/rulesets
make probe           # the iPhone probe build and its browser tests, needs Node 22 (D-113)
make proto           # buf lint, and write the generated Go code of go/gen
make contract        # buf lint, the generated code in Git, and buf breaking against main
make go-test         # gofmt, go mod tidy, go vet, and the Go unit tests
make emulator-test   # the Go tests over the Auth and Firestore emulators, needs Java 21
make web             # the web client: types, unit tests, build, browser tests, needs Node 22
```

The product targets need Go and have their own CI jobs. `make verify` stays Python only (D-113, D-126). After a change in `proto/`, run `make proto` and commit `go/gen` and `web/src/gen`.

Paid targets: `make codex-review` and `make claude-review`. They spend the owner's Codex plan and Claude plan, never the API (D-8, D-88).

## Skills

| Skill | Use when |
|---|---|
| `ste-writing` | Before you write or edit any `.md` file. |
| `one-pr-one-session` | Before any work on a pull request: start, revision, review, merge question, or hand-off. |
| `gitar-review` | After each push to a pull request, documents alone included (D-335). |
| `pr-review` | For a cross-provider review, `make codex-review`, `make claude-review`, and each answer to a review finding. |

## File map

- `docs/session-handoff.md` - the resume point.
- `docs/design.md` - thesis, experience, system context, safety, privacy.
- `docs/operations.md` - the incidents, the policy replay, and the model change.
- `docs/decisions.md` - every owner decision, with a date.
- `docs/questions.md` - every question and its answer, and the open questions.
- `docs/roadmaps/high-level-roadmap.md` - the phases and their exit evidence.
- `docs/roadmaps/README.md` - the rules for focused roadmaps.
- `docs/research/` - Decktome patterns, exercise safety, platform, cloud, and AI research, and the spike reports.
- `docs/reviews/` - Codex review records.
- `docs/tools/` - the checks and their tests.
- `tools/spikes/` - the Phase 1 spike harnesses, outside the product code, with their tests.
- `proto/` - the Connect-RPC contract. `buf.yaml` and `buf.gen.yaml` configure it.
- `go/` - the Go API: `go/cmd/api` and `go/internal`, with the generated code in `go/gen`. `go/cmd/lunaeval` is the Luna evaluation, and `go/cmd/replay` is the policy replay.
- `web/` - the web client, with the generated code in `web/src/gen`. `web/README.md` describes it.
- `firebase.json` and `emulators/` - the local emulators and their pinned `firebase-tools`.
