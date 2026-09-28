# Decktome patterns for Gym Route

## 1. Purpose, date, and method

This document compares the Decktome repository with the needs of Gym Route. It gives one disposition for each Decktome pattern: Adopt, Adapt, or Decline. It also lists the Decktome rules that must not transfer, the gaps that Decktome does not cover, and the lessons from its drift.

Date of the research: 2026-09-28.

Method:

1. Read a local copy of the Decktome repository as read-only, at commit 220818b with 263 commits.
2. Do not read the Decktome secret files and local scratch files.
3. Read an earlier role-model analysis. That analysis assumed native mobile apps. The owner rejected native apps (D-17), so this document replaces its dispositions.
4. Read every owner decision in `docs/decisions.md` (D-1 to D-83).
5. Confirm with `ls` or `grep` that each cited Decktome path exists.

### Citation rule

This document cites a Decktome path in the form `decktome:<path>`, for example `decktome:AGENTS.md` or `decktome:docs/decisions.md`. It cites a Decktome decision id in the form decktome:D-811. The ref-check tool of Gym Route skips each token with the `decktome:` prefix. A bare D- id in this document always names a row of the Gym Route register. A bare Decktome path or a bare Decktome decision id must not occur in this document.

## 2. What Decktome is, and its stack and process

Decktome is an agentic deck builder for Magic: The Gathering. A user uploads a card collection and talks to an LLM agent. The agent asks questions and returns a deck. Deterministic Go code owns card data and legality, and every card that the model names goes through the rules engine first (`decktome:README.md`, `decktome:docs/design-roadmap.md`). Decktome runs as an invite-only, installable web app (PWA) for a small group, with one owner. AI agents do almost all of the work.

Gym Route has the same shape: one owner, AI agents, an installable phone-first web app, and an LLM over a strict engine (D-17, D-22, D-23). The table gives the observed Decktome pattern and the disposition for Gym Route.

| Pattern | Evidence | Disposition | Reason and governing D- id |
|---|---|---|---|
| Go backend in one module, with `cmd/api`, `cmd/worker`, and `internal/*` packages | `decktome:go/go.mod`, `decktome:go/cmd/api`, `decktome:go/cmd/worker` | Adopt | The Decktome stack is the default (D-74, D-17). |
| One Protobuf contract, buf v2, committed generated code, breaking-change check | `decktome:buf.yaml`, `decktome:buf.gen.yaml`, `decktome:proto/mtg/v1` | Adopt | Default stack (D-74). Package name changes to a Gym Route name. |
| Connect-RPC between the web client and the API | `decktome:go/go.mod`, `decktome:web/apps/web/package.json` | Adopt | Default stack (D-74). |
| Web client: React 19, Vite 7, TypeScript, Tailwind 4, `vite-plugin-pwa` | `decktome:web/apps/web/package.json` | Adapt | The owner chose the Decktome React stack after a comparison, with Vite 8 in place of Vite 7 (D-84). |
| Firebase Hosting for the web app | `decktome:firebase.json` | Adapt | Default Firebase Hosting URL, no custom domain (D-17, D-82). |
| Firebase Auth, email and password, ID token checked by a Connect interceptor | `decktome:go/internal/auth/auth.go` | Adopt | D-75. |
| Invite allowlist document, checked for each request | `decktome:go/internal/auth/auth.go`, `decktome:docs/setup-gcp.md` | Adopt | D-75. The list holds one account (D-67). |
| CORS limited to allowed origins | `decktome:go/internal/auth/cors.go` | Adapt | One origin: the default Hosting URL of the web app (D-82). |
| Firestore with default-deny client rules, the API is the only door | `decktome:firestore.rules` | Adopt | Firestore through the API is the source of record after a sync (D-77). |
| Cloud Run service and jobs that scale to zero, one service account for each part | `decktome:docs/setup-gcp.md` | Adopt | Follow Decktome for GCP (D-18). |
| API on the default Cloud Run URL | `decktome:docs/setup-gcp.md` | Adopt | D-82. |
| Cloud Build deploy from `main`, with path filters | `decktome:cloudbuild/api.yaml`, `decktome:cloudbuild/web.yaml` | Adopt | D-18, D-14. |
| Commit-order deploy guard | `decktome:docs/tools/deploy_order.py` | Decline | Not necessary until two triggers race (D-74 needs evidence for each addition). |
| Two GCP projects planned, one production project in practice | decktome:D-24, `decktome:.firebaserc` | Adapt | One development project only, in `us-central1` (D-76). |
| Region `us-central1` | `decktome:docs/setup-gcp.md` | Adopt | D-76. |
| Secret Manager with one accessor binding for each secret and service account | `decktome:docs/setup-gcp.md` | Adopt | D-18. |
| LLM role layer: no model id at a call site, roles and prices in files | `decktome:go/internal/llm/roles.json`, `decktome:go/internal/llm/prices.json` | Adapt | Same layer, new contents: `gpt-6-luna` at medium effort (D-22, D-24). |
| Fake LLM provider with fixtures, fakes for every cloud dependency | `decktome:go/internal/llm/fake.go`, `decktome:go/internal/dispatch` | Adopt | D-24 requires a fake provider for tests. |
| Thin agent over a strict engine | `decktome:docs/design-roadmap.md`, `decktome:AGENTS.md` | Adapt | Luna proposes, a deterministic policy checks each set and load (D-23). |
| Spend cap for each user, cloud budget alert | decktome:D-421, `decktome:docs/setup-gcp.md` | Adapt | Monthly cap for each user and for the project, owner approves paid runs (D-25). |
| Rate limit for each IP address | `decktome:go/internal/ratelimit/ratelimit.go` | Adapt | Limit for each uid first. Gym networks and carriers share addresses (D-67). |
| Firebase emulators and a fake storage server for local work | `decktome:firebase.json`, `decktome:compose.yaml` | Adopt | D-74. Use a port map that does not collide with Decktome. |
| Daily Firestore backup, 10-day retention | `decktome:docs/deploy-and-rollback.md` | Adapt | Turn it on at the first deploy, not late (see section 6). |
| Rollback procedure for each part | `decktome:docs/deploy-and-rollback.md` | Adopt | D-18. |
| Makefile as the single entry point, `make help`, `make doctor`, pinned tools | `decktome:Makefile`, `decktome:scripts/doctor.sh`, `decktome:.nvmrc` | Adopt | D-6, D-74. |
| Paid targets marked as paid, with a question before each run | `decktome:docs/reference/paid-targets.md` | Adopt | D-25. |
| `verify` workflow and `make verify` as the local copy of CI | `decktome:.github/workflows/verify.yml` | Adopt | D-6 names `make verify`. |
| Docs-only skip of the heavy CI jobs | `decktome:docs/tools/ci_skip.py` | Decline | No code exists yet, so nothing heavy runs. |
| Dependabot, monthly and grouped | `decktome:.github/dependabot.yml` | Adopt | D-74. |
| Documentation gates: ste-check, ref-check, context-budget | `decktome:docs/tools/ste-check.py`, `decktome:docs/tools/ref_check.py`, `decktome:docs/tools/context_budget.py` | Adopt | D-6, D-83. |
| Pull request template, pr-check, pr-contract workflow | `decktome:.github/pull_request_template.md`, `decktome:docs/tools/pr_check.py` | Adapt | One milestone with two, three, or more concerns (D-10, D-12). |
| Review gate from `main`, head read as data only | `decktome:.github/workflows/review-gate.yml`, `decktome:docs/tools/review_gate.py` | Adopt | D-6. |
| `review-override` label for docs-only pull requests | `decktome:docs/tools/review_gate.py` | Adapt | The label exists, but the Codex review still applies to docs-only pull requests until the roadmaps finish (D-4). |
| Codex review by `make codex-review`, record in the reviews folder, verdict bound to one head | `decktome:docs/tools/codex_review.py`, `decktome:docs/reviews` | Adapt | D-3, D-8. The author session runs it after CI is green. |
| Gitar third-party review before Codex | `decktome:.claude/skills/gitar-review/SKILL.md` | Decline | Out until the owner approves it (D-3). |
| Branch ruleset in the repository, ruleset-check | `decktome:.github/rulesets/review-gate.json`, `decktome:docs/tools/ruleset_check.py` | Adopt | D-6. |
| Pre-commit hook: no commit on `main`, STE check of staged docs | `decktome:.githooks/pre-commit` | Adapt | D-6. Add a commit-msg hook for commit format and attribution (D-14). |
| `make where`, `make hooks` | `decktome:Makefile` | Adopt | D-6. |
| Session hooks: session_bind and context_checkpoint | `decktome:.claude/hooks/session_bind.py`, `decktome:.claude/hooks/context_checkpoint.py` | Adapt | D-6. The hooks hold `decktome-*` names, so the port renames them. |
| Skills: ste-writing, pr-review, one-pr-one-session | `decktome:.claude/skills/ste-writing/SKILL.md`, `decktome:.claude/skills/pr-review/SKILL.md`, `decktome:.claude/skills/one-pr-one-session/SKILL.md` | Adapt | D-6: remove the Gitar steps. |
| CLAUDE.md holds the hard rules, AGENTS.md holds the rest | `decktome:CLAUDE.md`, `decktome:AGENTS.md` | Adapt | AGENTS.md holds all rules, CLAUDE.md only points to it (D-7). |
| One concern for each pull request | `decktome:AGENTS.md` | Adapt | One cohesive milestone with two, three, or more concerns (D-10, D-12). |
| One pull request for each clean session | `decktome:.claude/skills/one-pr-one-session/SKILL.md` | Adopt | D-12. |
| Owner confirms each merge, then auto-merge with a four-part summary | `decktome:AGENTS.md` | Adopt | D-13. |
| Commit titles with roadmap ids, no Conventional Commits (0 of 263 titles) | `git log` of Decktome | Decline | Conventional Commits (D-14). |
| No AI attribution, a written rule only | `decktome:CLAUDE.md` | Adapt | D-14. Add a machine check (see section 6). |
| Session hand-off with a resume section and dated facts | `decktome:docs/SESSION-HANDOFF.md` | Adopt | D-14 requires a current hand-off. |
| Append-only decision register with back-marks | `decktome:docs/decisions.md`, `decktome:docs/tools/test_decision_marks.py` | Adopt | `docs/decisions.md` already uses the format. |
| Two question registers | `decktome:docs/owner-questions.md`, `decktome:docs/open-questions.md` | Adapt | One questions file (see the header of `docs/decisions.md`). |
| One monolithic roadmap with a long correction log | `decktome:docs/design-roadmap.md` | Adapt | One high-level roadmap and focused roadmaps, with work areas and no permanent pull request numbers (D-1, D-4, D-9). |
| "Log ids, never PII or raw prompts" | `decktome:AGENTS.md` | Adapt | Stronger: no workout text, photos, prompts, or health details (D-80). |
| Soft-delete account closure | decktome:D-941 | Decline | No user data controls (D-78). |
| Custom domain | decktome:D-556 | Decline | D-82, D-81. |
| Owner push notification for each feedback verdict | `decktome:go/internal/notify` | Decline | No notifications (D-61). |
| Web tests: vitest, Testing Library, jest-axe, Playwright smoke | `decktome:web/apps/web/package.json` | Adapt | The React stack carries these tools (D-84). D-72 defers accessibility work, so jest-axe is not a gate. |

## 3. Dispositions in detail

### 3.1 Adopt unchanged

These patterns transfer with a new project name and no change of rule:

- The Go module layout, the Protobuf contract, Connect-RPC, and the committed generated code.
- Firebase Auth with the ID-token interceptor and the invite allowlist (D-75).
- Firestore default-deny rules. The API is the only door to the data (D-77).
- Cloud Run, Cloud Build from `main`, Secret Manager, and one service account for each part (D-18).
- The fake provider pattern: each cloud dependency arrives with its local fake in the same pull request.
- The Makefile entry points, pinned tools, `make doctor`, `make verify`, `make where`, and `make hooks` (D-6).
- The review gate that runs from `main` and reads the head as data only.
- The branch ruleset and its checker, squash merge only, and no direct push to `main` (D-14).
- One clean session for each pull request, and the owner confirmation of each merge (D-12, D-13).
- The session hand-off, the append-only decision register, and the paid-target discipline (D-25).

### 3.2 Adapt for a phone-first health and fitness web app

| Pattern | Change for Gym Route | Governing D- id |
|---|---|---|
| PWA from `vite-plugin-pwa` | Offline logging is necessary. The app writes each set to device storage first and syncs later. Each write carries a client id, so a retry is safe. | D-62, D-77 |
| Logging rule | Logs, errors, and metrics carry ids only. No workout text, no photos, no prompts, no health details. | D-80 |
| Thin agent over a strict engine | The engine checks numbers, not card legality. It checks each set, load, rep target, and change before the user sees it. A rules fallback runs when Luna fails. | D-23, D-37, D-65 |
| LLM role layer | One role table for planning, revision, and equipment photos, all on `gpt-6-luna` at medium effort. The fake provider gives fixed plans for tests. | D-22, D-24 |
| GCP projects | One development project in `us-central1`. It holds the live app of the owner. No production project exists. | D-76, D-67 |
| Backups | Turn on the daily Firestore backup at the first deploy. Workout history is the one asset that the project cannot rebuild. | D-18, D-76 |
| Rate limit | Limit each uid first, and each address second. | D-67 |
| Spend cap | A monthly cap for each user and a cap for the project. The owner approves each paid development run. | D-25 |
| CORS | Allow one origin: the `web.app` URL of the Hosting site. | D-82 |
| STE checker | Add fitness terms to the allow list of the checker, for example "loaded", "selectorized", "seated", and "warm-up". Add only after a real false finding. | D-83 |
| Codex review rules | The Decktome review reads its rules from `origin/main`. The first Gym Route pull request adds those rules, so `main` holds none yet. The first round reads the rules from the head, and the record says so. | D-3, D-8 |
| Codex review of docs-only pull requests | The `review-override` label exists, but the review still runs on docs-only pull requests until the owner ends the roadmap period. | D-4, D-15 |
| Pull request scope | The template and pr-check ask for the concerns of one milestone and one combined acceptance story. | D-10, D-12 |
| Commit format | Conventional Commits. A commit-msg hook checks the title and refuses an AI attribution trailer. | D-14 |
| AGENTS.md and CLAUDE.md | AGENTS.md holds every rule. CLAUDE.md only tells Claude Code to read AGENTS.md. | D-7 |
| Skills | Port ste-writing, pr-review, and one-pr-one-session, and remove the Gitar steps. | D-6, D-3 |
| Question registers | One questions file with the owner queue in it. The header of `docs/decisions.md` names that file. | None yet |
| Roadmaps | One high-level roadmap with broad work areas and focused roadmaps. No correction log in the header, because git holds the history. | D-1, D-4, D-9 |

### 3.3 Port list for the first pull request

The first pull request ports the process items of D-3 and D-6. The table maps each item to its Decktome source and to the change that Gym Route needs.

| Item | Decktome source | Change for Gym Route | Governing D- id |
|---|---|---|---|
| Codex review command | `decktome:docs/tools/codex_review.py`, `decktome:Makefile` | Remove the Gitar flag and its steps. Read the rules from the head in the first round only. | D-3, D-8 |
| AGENTS.md and CLAUDE.md link | `decktome:AGENTS.md`, `decktome:CLAUDE.md` | Move the hard rules into AGENTS.md. CLAUDE.md holds one pointer. | D-7 |
| ste-check | `decktome:docs/tools/ste-check.py` | Same rules. Fitness terms go in the allow list only after a false finding. | D-83 |
| ref-check | `decktome:docs/tools/ref_check.py` | Skip each token with the `decktome:` prefix. | D-6 |
| pr-check and the template | `decktome:docs/tools/pr_check.py`, `decktome:.github/pull_request_template.md` | "The one concern" becomes "The concerns of the milestone". | D-10, D-12 |
| review-gate workflow | `decktome:.github/workflows/review-gate.yml`, `decktome:docs/tools/review_gate.py` | Keep the label path for docs-only pull requests. The session still runs the Codex review during the roadmap period. | D-4 |
| Pre-commit hook | `decktome:.githooks/pre-commit` | Keep the `main` guard and the STE check. Add a commit-msg hook. | D-6, D-14 |
| `make where`, `make hooks`, `make verify` | `decktome:Makefile` | Keep the free targets only. Add the paid targets later with a CAUTION line. | D-6, D-25 |
| Ruleset and ruleset-check | `decktome:.github/rulesets/review-gate.json`, `decktome:docs/tools/ruleset_check.py` | Required checks match the Gym Route CI jobs. | D-6 |
| Session hooks | `decktome:.claude/hooks/session_bind.py`, `decktome:.claude/hooks/context_checkpoint.py`, `decktome:.claude/settings.json` | Rename the `decktome-*` state paths. | D-6 |
| context-budget | `decktome:docs/tools/context_budget.py` | Limits for AGENTS.md, the hand-off, and the skills. | D-6 |
| Skills | `decktome:.claude/skills/ste-writing/SKILL.md`, `decktome:.claude/skills/pr-review/SKILL.md`, `decktome:.claude/skills/one-pr-one-session/SKILL.md` | Remove the Gitar steps and the card game examples. | D-6 |
| Tests of the tools | `decktome:docs/tools` | Port the unit test of each ported tool. | D-6 |

### 3.4 Decline

| Pattern | Reason | Governing D- id |
|---|---|---|
| Gitar review and the gitar-review skill | The owner did not approve it. Remove the Gitar steps from each ported skill. | D-3, D-6 |
| Commit-order deploy guard | Two triggers do not race yet. Add it when evidence shows a race. | D-74 |
| Paid evaluation gates, autotune, and feedback loops | They measure deck quality for Magic: The Gathering. Gym Route needs its own policy tests, which cost nothing. | D-23, D-25 |
| Custom domain and its DNS | The app lives on the default Hosting URL, and the API on the default Cloud Run URL. | D-17, D-81, D-82 |
| Soft-delete account closure | No user data controls apply to a single-user app. The pattern also keeps data after closure, so do not copy it. | D-67, D-78 |
| Docs-only CI skip | No code exists yet, so CI has no heavy job to skip. | D-74 |
| Production GCP project | One development project only. | D-76 |
| Owner push notifications | No notifications. | D-61 |

## 4. Decktome product rules that must not transfer

| Decktome item | Evidence | Why it does not transfer |
|---|---|---|
| Guardrails of the card game: rules engine for each card, no ban list in code, exact card names, pool mode, Scryfall image credit | `decktome:docs/design-roadmap.md`, `decktome:AGENTS.md` | Card rules. Gym Route checks sets and loads (D-23). |
| Hard rules about card names, rules text, and the sources of the card game | `decktome:CLAUDE.md` | Gym Route uses current guidance and systematic evidence, with a date (D-38). |
| The mtg-corpus skill | `decktome:.claude/skills/mtg-corpus/SKILL.md` | Domain knowledge of the card game. |
| The live-test skill | `decktome:.claude/skills/live-test/SKILL.md` | It walks the deck screens of the deployed Decktome app. |
| The dogfood agent | `decktome:.claude/agents/deck-builder-dogfood.md` | It builds decks. |
| The design-doc-style skill | `decktome:.claude/skills/design-doc-style/SKILL.md` | It shapes the Decktome monolithic roadmap. |
| Data providers: Scryfall, Commander Spellbook, EDHREC, MTGJSON, Moxfield, ManaBox | `decktome:docs/setup-gcp.md`, `decktome:Makefile` | Card data sources. |
| "A verdict keeps the object it names" and the user record with its counters | decktome:D-635, decktome:D-638 | Product choices of the deck builder. |
| The contents of the role and price files | `decktome:go/internal/llm/roles.json`, `decktome:go/internal/llm/prices.json` | The layer transfers. The roles and models do not (D-24). |
| The domain `decktome.com` | decktome:D-556 | No custom domain (D-82). |
| The mobile and store proposal, with a TWA for Android and a native shell for iOS | `decktome:docs/reference/mobile-and-engagement-2026-09-05.md`, decktome:D-548 | Gym Route never goes to an app store (D-17). |
| Pushover notification of the owner | `decktome:go/internal/notify` | No notifications (D-61). |
| The Decktome port map | `decktome:compose.yaml` | Gym Route needs its own ports on the same Mac. |

## 5. Gaps that Decktome does not address

Decktome stores no health data, uses no camera, and needs no offline writes. So Gym Route must design these items fresh.

| Gap | Why it matters for Gym Route | First direction | D- id |
|---|---|---|---|
| Offline workout logging in a PWA on iOS | The gym has weak signal. iOS keeps the storage of Safari apart from the storage of the Home Screen app, and it can evict web data. | Install to the Home Screen before the first workout. Sync after each set when the network is available. Show the sync state. Test eviction. | D-21, D-29, D-62, D-77 |
| Camera capture in an iOS web app | The equipment flow starts with one photo. | Use the file input with camera capture. Keep the manual path complete. | D-50, D-51, D-55 |
| Exercise-safety claims and the wellness boundary | Gym Route gives fitness guidance only. | State the boundary in the app. The policy blocks unsafe loads. Symptom reports warn first. | D-35, D-36, D-40 |
| Sensitive fitness data in LLM prompts and logs | Injuries, weight, and pain go to Luna. | Send the minimum fields. Keep prompts out of logs and error reports. | D-24, D-80 |
| iOS web limits | No vibration, no background sync, and no reliable timer in the background. | Visual cues only. The rest timer runs on the screen. The app syncs when it opens. | D-21, D-58, D-59, D-61 |
| LLM validation of numeric prescriptions | Decktome checks card legality. Gym Route checks numbers: load steps, rep ranges, reps in reserve, and progression size. | A versioned policy with table tests and a rules fallback. | D-23, D-37, D-64, D-65 |
| Photo privacy | A gym photo can hold location data and other people. | Strip EXIF data on the device before upload. Delete source photos after confirmation. Keep no user photo in an evaluation set. | D-52, D-53, D-56 |
| One development project without production | Decktome never created its planned development project. Gym Route has the reverse case. | Name the project for development, and treat its data as live data of the owner. | D-67, D-76 |

Distribution gaps are NOT APPLICABLE. Gym Route never goes to an app store (D-17). So app store review, signing keys, store privacy forms, and store deletion rules do not apply. User data controls and public policy pages are also out of scope (D-78, D-79, D-81).

## 6. Lessons

Decktome wrote some rules and did not add a machine check. Those rules drifted. For Gym Route, add the small check in the same pull request as the rule.

| Rule | Drift in Decktome | Evidence | Small check for Gym Route |
|---|---|---|---|
| No AI attribution | 10 squash commits on `main` carry a Claude co-author trailer. | `decktome:CLAUDE.md`, `git log` of Decktome | A commit-msg hook and a CI check refuse the trailer (D-14). |
| One concern for each pull request | Many pull requests hold a deploy read and a fix, or a batch of findings. | `decktome:AGENTS.md`, `git log` of Decktome | pr-check asks for the concerns and one milestone story (D-10, D-12). |
| Two GCP projects | Only the production project exists. | decktome:D-24, `decktome:.firebaserc` | The setup document and `make doctor` name the one project (D-76). |
| Daily backups | The backup started on 2026-09-23, after a repository review. | decktome:D-936, `decktome:docs/deploy-and-rollback.md` | The first deploy checklist turns on the backup and reads it back. |
| Small start-read files | The hand-off grew fast, and a byte budget came later. | `decktome:docs/tools/context_budget.py` | Port context-budget in the first pull request (D-6). |

Each check is small. A rule without a check depends on memory, and memory of an agent session ends with the session.
