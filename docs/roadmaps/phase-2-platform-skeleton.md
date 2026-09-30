# Workout App - Phase 2 focused roadmap: platform skeleton

This roadmap splits Phase 2 of `docs/roadmaps/high-level-roadmap.md` into pull requests. `docs/roadmaps/README.md` gives the rules. The high-level roadmap keeps the objective, the order, and the exit evidence of the phase.

The date of this version is 2026-09-30.

## 1. Scope and start state

Phase 2 builds the smallest running system on the Decktome stack (D-74), with the React web client of D-84. At the end of the phase, the owner installs the app on the iPhone, signs in, and sees an empty home screen. The API runs on Cloud Run in the development project (D-76). A merge to `main` deploys both parts (D-14, D-18).

The exit of Phase 1 holds:

- The Luna plan spike gave a go (D-101). The report is `docs/research/luna-plan-spike.md`.
- The recognition spike gave a no-go (D-107). The owner then moved photo recognition out of the phases (D-110, D-111).
- The iPhone web platform spike gave a go (D-118, D-121). The report is `docs/research/iphone-platform-spike.md`.

The repository holds no product code yet. The project `gym-route-dev` exists with Firebase Hosting and Firebase Authentication only, and it has no billing account (D-99, D-116). Its Hosting site serves the probe of work area 1.3. Work area 2.3 moved the app to the project `nk-workout-app-prod`, and the app name to "Workout App" (D-136, D-137).

## 2. Owner answers for this phase

| Question | Answer | Decision |
|---|---|---|
| Q-91, kilogram markings | Out of scope. The app supports machines with pound markings only. | D-122 |
| Q-93, the end of the D-4 period | The period ended on 2026-09-29. PR-7 turns on the `review-override` label. | D-125 |
| Q-99, Firestore backups | Point-in-time recovery and a daily backup with a retention of 10 days, from work area 2.3. | D-124 |
| Q-100, cardio log fields | Duration and an effort rating from 1 to 10. Distance, level, pain, and note are optional. | D-123 |
| Q-140, the product checks | Their own `make` targets and CI jobs. `make verify` stays Python only (D-113). | D-126 |
| Q-141, the change of the D-4 flag | PR-7. | D-125 |
| Q-143, required checks | Each CI job of the product code is a required check of `main`. | D-127 |
| Q-144, the split | PR-8 to PR-11, in section 4. | D-128 |

The session of PR-10 read the billing state, and then the owner answered Q-142 (D-139). No question of this phase stays open.

Phase 3 needs D-122 and D-123 for the domain model. Phase 2 records them because the owner answered them in this session.

## 3. Rules for each pull request of this phase

- The layout of the product code follows Decktome (D-74): `proto/` for the contract, `go/` for the API, and `web/` for the web client. `web/` is one npm package, not the pnpm workspace of Decktome (D-135).
- Each product check gets its own `make` target and its own CI job (D-126). `make verify` stays Python only (D-113).
- Each new CI job of the product code becomes a required check of `main` in the pull request that adds it (D-127). The session applies the ruleset only after the owner approves the step at run time. Then it runs `make ruleset-check`.
- Tests and CI make no call to the real project. They use the Firebase emulators and fakes, as the probe does (D-115).
- No account email and no uid goes into the repository. The allowlist of D-75 lives in the live project alone.
- Logs, metrics, and error reports hold ids only (D-80).
- Phase 2 has no Luna call and no role layer. Phase 3 adds them (D-24).
- A change of the live project needs the approval of the owner at run time. Each step that costs money has a **Paid** mark.
- Codex reviews each pull request with code (D-15). A pull request of documents alone can use the `review-override` label (D-125).
- A recommendation of `docs/research/platform-cloud-and-ai.md`, such as REC-1, becomes a decision only when the owner answers it in the session of the pull request.

## 4. Pull requests

The PR-<n> order is the order of work. Each pull request needs the one before it on `main`.

| Id | Work area | Title | Paid step |
|---|---|---|---|
| PR-7 | all | `docs: the Phase 2 focused roadmap (PR-7)` | none |
| PR-8 | 2.1 | `feat: the contract and the API skeleton (PR-8)` | none |
| PR-9 | 2.2 | `feat: the installable web shell (PR-9)` | none |
| PR-10 | 2.3 | `feat: the development project and the deploy (PR-10)` | yes, billing and backups |
| PR-11 | 2.2, 2.3 | `docs: the Phase 2 device and deploy check (PR-11)` | none |

### PR-7 - The Phase 2 focused roadmap

Branch: `docs/pr-7-phase-2-roadmap`.

Concerns:

- this file,
- the owner answers of section 2 (D-122 to D-128),
- the end of the D-4 period: `OVERRIDE_ENABLED` in `docs/tools/review_gate.py`, its tests, and the rule text (D-125).

Acceptance story: the focused roadmap gives each Phase 2 pull request its id, branch, title, concerns, acceptance story, and checks, and `make verify` passes.

Checks: `make verify` and `make pr-check`, free. Codex reviews PR-7, because it changes the review gate (D-125).

### PR-8 - The contract and the API skeleton

Branch: `feat/pr-8-api-skeleton`. Work area 2.1.

Concerns:

- a buf v2 contract in `proto/` with a Workout App package name, and the generated Go code in Git,
- the lint of the contract, and a breaking-change check against `main`,
- one call `GetMe` that returns the uid of the signed-in owner, so the web shell can prove the whole path,
- a Go API on Connect-RPC with the Firebase ID token check and the invite allowlist (D-75),
- CORS for the one origin of the web app (D-82),
- a version endpoint that names the build commit, for the deploy check of PR-11,
- the Firebase emulators for Auth and Firestore, with ports that do not collide with Decktome or the probe,
- the `make` targets and the CI jobs of the contract, the Go tests, and the emulator tests (D-126, D-127).

Decktome gives the patterns: `decktome:go/internal/auth/auth.go` for the token check and the allowlist, and `decktome:go/cmd/api/main.go` for the routes. Decktome reads the build commit on its route `/readyz`. The route `/healthz` does not answer on a `run.app` URL (`decktome:cloudbuild/api.yaml`), so the version endpoint uses another path.

Acceptance story: the new `make` targets run the contract, Go, and emulator tests for free. A request with no token, a bad token, or a uid outside the allowlist fails. A request of an allowed uid gets its uid back.

Checks:

- The new `make` targets and CI jobs, free. `make verify` stays green.
- `make ruleset-check` after the owner approves the change of the live ruleset.

Questions for the session: the startup guard that refuses an emulator variable on Cloud Run (REC-17), and the Go and buf version pins.

### PR-9 - The installable web shell

Branch: `feat/pr-9-web-shell`. Work area 2.2. It needs PR-8 on `main`.

Concerns:

- a web client on the stack of D-84, with the phone layout only (D-20),
- a shell that fills the whole screen of the Home Screen app, with no gap at the bottom edge (D-120),
- a viewport meta that blocks the pinch zoom, as Decktome does (D-120),
- sign-in with email and password on Firebase Authentication (D-75), and no form that makes an account (D-117),
- an empty home screen that calls `GetMe` through the API,
- the offline store and the outbox skeleton (D-62, D-77): a local database, an outbox table, and one write for a change and its outbox entry,
- the generated TypeScript code of the contract, and the `make` target and the CI job of the web tests (D-126, D-127).

The sync call of the outbox comes in Phase 6. PR-9 proves the local write alone.

Acceptance story: the browser tests pass in the WebKit and Chromium engines with phone emulation. The tests sign in against the Auth emulator, and they reach the home screen with the uid from the API. They prove that the page blocks the pinch zoom. They write a change and its outbox entry.

Checks:

- The new `make` target and CI job, free. `make verify` stays green.
- `make ruleset-check` after the owner approves the change of the live ruleset.

Questions for the session:

- the local store and the outbox form (REC-1),
- the update strategy of the service worker (REC-3),
- the persistent storage request after the first sign-in (REC-5).

The owner answered the three questions on 2026-09-29 (D-132, D-133, D-134), and chose one npm package for `web/` (D-135).

### PR-10 - The development project and the deploy

Branch: `feat/pr-10-deploy`. Work area 2.3. It needs PR-9 on `main`.

Before the work, read the billing state of the owner account. Then ask Q-142. `gcloud` shows the billing account and its link. The owner reads the credits of the account in the console.

Concerns:

- the new name of the app and the new project `nk-workout-app-prod`, with Hosting, Auth, the owner account, and the key limits of D-117 (D-136 to D-138). The owner added this concern in the session,
- **Paid:** the billing link of `nk-workout-app-prod` and an alerts-only budget, after the answer to Q-142 (D-99 names this step),
- Firestore in `us-central1` (D-76), with rules that refuse each client read and write (D-77), deployed from the repository,
- **Paid:** point-in-time recovery and a daily backup schedule with a retention of 10 days, read back after the change (D-124),
- the Cloud Run service of the API, with its own service account, and Secret Manager with no secret value yet,
- the Cloud Build deploy from `main` for the API and the web app, with path filters (D-14, D-18),
- the allowlist entry of the owner in the live project, with no uid in the repository (D-75),
- the setup and rollback documents, as `decktome:docs/setup-gcp.md` and `decktome:docs/deploy-and-rollback.md` give them.

The Hosting site of `nk-workout-app-prod` serves the app. The key limits of D-117 name the sites of the new project (D-137). The probe stays on `gym-route-dev` until the owner shuts that project down.

CAUTION: the Firestore location is permanent. Check `us-central1` before the create step.

Acceptance story: the project holds Firestore with point-in-time recovery and the backup schedule, the Cloud Run service, and the Cloud Build triggers on `main`. A merge of PR-10 starts the deploy of both parts, and PR-11 reads the result.

Checks:

- `make verify` and the product checks, free.
- Each change of the live project runs after the owner approves it at run time.
- **Paid:** the backups and point-in-time recovery cost a few cents a month for one user (assumption, `docs/research/platform-cloud-and-ai.md` section 8.2).

Questions for the session: Q-142, the Firestore edition and mode (REC-12), and the Cloud Run cost settings (REC-15). The owner answered them and Q-152 to Q-159 (D-136 to D-142).

### PR-11 - The Phase 2 device and deploy check

Branch: `docs/pr-11-phase-2-check`. Work areas 2.2 and 2.3. It needs the deploy of PR-10 from `main`.

A pull request can not hold a result that comes after its own merge. So this pull request records the exit evidence of the phase.

Concerns:

- the deploy check: the live version endpoint names the merge commit of PR-10,
- the device check on the iPhone of the owner, in the Home Screen app from Chrome (D-29, D-119),
- a short report in `docs/research/`,
- the shutdown of the old project `gym-route-dev`, after the owner signs in to the app on the new project (D-137),
- the package rules `PACKAGE_NO_DELETE` and `PACKAGE_SERVICE_NO_DELETE` in `buf.yaml` (D-138). This concern changes code, so it needs the review of the other provider (D-15, D-144),
- the repair of the build `deploy-rules`, which needed the viewer role of Service Usage (D-146),
- the space above the title that keeps it out of the iOS blur band (D-145). The owner added this concern after the device check.

The device check holds these items:

1. Add the app to the Home Screen from Chrome.
2. Sign in with email and password in the Home Screen app.
3. See the empty home screen with the answer of the API.
4. Look for a gap at the bottom edge of the screen.
5. Pinch the page, and look for a zoom.

Acceptance story: the report gives a result for each item and the commit that the live version endpoint names. Items 1 to 3 pass, and items 4 and 5 show no gap and no zoom.

Checks: `make verify`, `make contract`, and `make web`, free. The owner does the device steps. The report holds no screenshot with personal data. The pull request gets the review of the other provider, with no `review-override` label (D-144).

State: the report is `docs/research/phase-2-check.md`. Each item of the device check passes, and the live endpoints name `df79c16`. The session shut down `gym-route-dev` on 2026-09-30.

## 5. Exit of the phase

Phase 2 ends when PR-11 merges with a pass for each item of its device check. The live version endpoint must name the merge commit of PR-10. The exit evidence of work area 2.1 is the set of product `make` targets of PR-8 (D-126).
