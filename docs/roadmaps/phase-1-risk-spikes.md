# Gym Route - Phase 1 focused roadmap: risk spikes

This roadmap splits Phase 1 of `docs/roadmaps/high-level-roadmap.md` into pull requests. `docs/roadmaps/README.md` gives the rules. The high-level roadmap keeps the objective, the order, and the exit evidence of the phase.

The date of this version is 2026-09-28.

## 1. Scope and start state

Phase 1 measures three risks before the project commits to an architecture or to content:

- the quality of Luna plans against the first draft of the policy rules (work area 1.1),
- the identification of gym machines from photos (work area 1.2),
- the iPhone web platform in Chrome (work area 1.3).

Each spike gives a report with numbers and a go or no-go. A no-go changes the high-level roadmap before Phase 2 starts. That change needs an owner decision.

The exit evidence of Phase 0 holds. `make ruleset-check` passed against the live ruleset on 2026-09-28 (D-91). The owner confirmed the phase order (D-92).

## 2. Owner answers for this phase

Every question of Phase 1 has an answer. No open question of this phase stays.

| Question | Answer | Decision |
|---|---|---|
| Q-95, usage policies | The dated copy of the policy page. Luna text stays inside the fitness boundary of D-36. | D-93 |
| Q-96, cost per photo | PR-4 measures it with one paid run. The cap is 2 USD. PR-4 measured 0.00037 USD per photo. | D-94 |
| Q-97, image licenses | CC0, public domain, CC BY, CC BY-SA, and CC BY-NC. | D-95 |
| Q-114, PR-<n> ids | One sequence across all focused roadmaps. | D-96 |
| Q-115, image store | A manifest in Git, and the images in a local cache. | D-97 |
| Q-116, plan spike cap | 2 USD for the paid run of PR-2. | D-98 |
| Q-117, probe origin | The development project of D-76, early, with Firebase Hosting and Firebase Authentication only. | D-99 |
| Q-118, spike language | Python. | D-100 |
| Q-130, probe camera page | No camera page, because the app of the roadmap has no photo upload after the no-go of PR-4 (D-110). | D-112 |

## 3. Rules for each spike pull request

- Spike code lives in `tools/spikes/<name>/`. It never enters the product code. `make test` runs the tests of each spike folder.
- The Luna harnesses use Python (D-100). Each spike pull request adds its unit tests to `make test`.
- Each paid call has a fake provider. Tests and CI use the fake provider only, so CI makes no paid call.
- A model id appears in the role configuration of the spike only, and never at a call site (D-24).
- A paid run starts only after the owner approves it at run time (D-25). The pull request records the approval, the cap, and the real cost.
- The OpenAI API key comes from the environment of the owner machine. No key goes into the repository, a log, or a report.
- Luna prompts stay inside the fitness boundary of D-36 and the usage policies of D-93.
- Each profile and each inventory is synthetic. No personal data and no user photo goes into a fixture or a report (D-53).
- Each report goes into `docs/research/`. It records the date, the model id, the prompt version, the real cost, and the go or no-go.

## 4. Pull requests

The PR-<n> order is the recommended order. PR-3 must merge before PR-4, and PR-5 must merge before PR-6. The other spikes do not depend on each other. The Phase 2 focused roadmap starts at PR-7 (D-96).

| Id | Work area | Title | Paid run |
|---|---|---|---|
| PR-1 | all | `docs: the Phase 1 focused roadmap (PR-1)` | none |
| PR-2 | 1.1 | `feat: the Luna plan spike (PR-2)` | yes, cap 2 USD (D-98) |
| PR-3 | 1.2 | `feat: the recognition test set (PR-3)` | none |
| PR-4 | 1.2 | `feat: the recognition spike (PR-4)` | yes, cap 2 USD (D-94) |
| PR-5 | 1.3 | `feat: the iPhone web platform probe (PR-5)` | none |
| PR-6 | 1.3 | `docs: the iPhone web platform spike report (PR-6)` | none |

### PR-1 - The Phase 1 focused roadmap

Branch: `docs/pr-1-phase-1-roadmap`.

Concerns:

- this file,
- the owner answers to the questions of Phase 1 (D-91 to D-100),
- the branch form of the `one-pr-one-session` skill, aligned with D-86.

Acceptance story: a later session can start the first Phase 1 spike from this file. Every Phase 1 question has an answer, and each paid run has a mark for owner approval.

Checks: `make verify` and `make pr-check`, free.

### PR-2 - The Luna plan spike

Branch: `feat/pr-2-luna-plan-spike`. Work area 1.1.

Concerns:

- a fixed set of 20 synthetic profiles and inventories, with the fields of D-41 (recommendation for the count),
- a JSON schema for the plan output of Luna, and a validator,
- the first draft of the policy rules table (D-23), with an id and a version for each rule,
- a harness that sends each profile to Luna three times, through a role configuration with a fake provider (D-24),
- one paid run, and the report.

The draft rules table covers the reps in reserve of D-37, the rounding of D-65, the load jumps, and the fitness boundary of D-36. Q-92, Q-104, Q-105, and Q-106 stay Phase 3 questions. The draft marks each rule that touches them as a draft, and it does not answer them.

Acceptance story: the report gives the schema pass rate, the policy rejection rate, and the cost per plan. It lists each unsafe proposal that the draft rules catch, and it gives a go or no-go for the Luna plan risk.

Checks:

- `make verify` with the fake provider, free.
- **Paid:** one run of 60 plans (estimate 0.35 USD). The cap is 2 USD (D-98). The owner approves the run at run time (D-25).

Result: go under the bar of D-101. The report is `docs/research/luna-plan-spike.md`, and the harness is in `tools/spikes/luna_plan/`. The real cost was 0.0564 USD. The draft rules also touch Q-101 and Q-102, and they mark those rules as drafts with no answer.

### PR-3 - The recognition test set

Branch: `feat/pr-3-recognition-test-set`. Work area 1.2.

Concerns:

- a manifest with the source URL, the author, the license, the SHA-256, the machine type, the gym, and the split of each image (D-97),
- a license check that refuses each license outside D-95,
- a download script that fills a local cache that Git ignores, and compares each SHA-256,
- a catalog shortlist of machine types, and the test set design of `docs/research/platform-cloud-and-ai.md` section 6.6.

The design of section 6.6 holds hard negatives, machines outside the catalog, degraded photos, and a split by gym. A script makes the degraded photos in the local cache, so Git holds no image.

Owner decisions for this pull request:

- About 200 images, so that the paid run of PR-4 stays below its cap. A smaller real count is acceptable (D-103).
- No image that shows the face of a person, except for a catalog type with few other photos (D-103, D-104).
- Wikimedia Commons first (D-102). Flickr through Openverse for the types with too few Commons images (D-105).

Acceptance story: `make verify` checks the manifest for free. The download script fills the cache with each image of the manifest, and each hash agrees.

Checks: `make verify`, free. The download script uses the network, and it makes no paid call.

### PR-4 - The recognition spike

Branch: `feat/pr-4-recognition-spike`. Work area 1.2. It needs PR-3 on `main`.

Concerns:

- a harness that sends each photo with the catalog shortlist to Luna, through a role configuration with a fake provider (D-24),
- a Luna answer with a machine type and a stated confidence, or an abstention,
- one paid run, and the report.

Acceptance story: the report gives the correct, wrong, and abstained counts, and each wrong answer with a high stated confidence. It gives the measured cost per photo (Q-96), and a go or no-go for the recognition risk.

Checks:

- `make verify` with the fake provider, free.
- **Paid:** one run on the test set of PR-3 (estimate 0.50 USD for 200 photos). The cap is 2 USD (D-94). The owner approves the run at run time (D-25).

Result: no-go under the bar of D-107. 8.6% of the 174 photos had a wrong answer with a high stated confidence, and the bar is 2%. The report is `docs/research/recognition-spike.md`, and the harness is in `tools/spikes/recognition/`. The real cost was 0.0651 USD. The owner then moved photo recognition out of Phase 4, and no phase holds it now (D-110, D-111).

### PR-5 - The iPhone web platform probe

Branch: `feat/pr-5-iphone-platform-probe`. Work area 1.3.

Concerns:

- a probe web app on the React stack of D-84, in `tools/spikes/iphone_probe/`,
- a probe page for each device item: IndexedDB data after an app kill, Screen Wake Lock, install, sign-in, and a startup timer,
- the development project of D-76, with Firebase Hosting and Firebase Authentication only (D-99),
- the Firebase configuration of the probe, with sign-in by email and password (D-75).

The session creates the project only after the owner approves that step. The project uses the free plan with no billing account (D-99). The web configuration of Firebase is not a secret, but no account email goes into the repository.

Acceptance story: the probe builds in CI, and its browser tests pass in the WebKit and Chromium engines. The project exists with Firebase Hosting and Firebase Authentication, and it has no other service.

Checks: `make verify`, free. `make probe` builds the probe and runs the browser tests against the local Auth emulator, free (D-113, D-115). The CI job `verify:probe` runs `make probe`, and it is not a required check (D-114).

### PR-6 - The iPhone web platform spike report

Branch: `docs/pr-6-iphone-platform-report`. Work area 1.3. It needs PR-5 on `main`.

Before the work, deploy the probe by hand from a clean checkout of `main` (D-14). Check the branch and the commit before the deploy. `tools/spikes/iphone_probe/README.md` gives the steps. The owner makes the probe account in the Firebase console. The probe has no form that makes an account, and self sign-up is off (D-117).

Concerns:

- the device checklist on the iPhone of the owner, in Chrome (D-29),
- the report, with a result for each item and a go or no-go.

The checklist holds these items:

1. Write data to IndexedDB, stop the app, and open it again.
2. Keep the screen on with Screen Wake Lock.
3. Add the probe to the Home Screen from Chrome.
4. Sign in with email and password in the Home Screen app.
5. Stop the Home Screen app, open it again, and check the sign-in state.
6. Measure the startup time of the React build (D-84).

The probe has no camera page (D-112).

Acceptance story: the report gives a result for each item and a startup time on the iPhone of the owner.

Checks: `make verify`, free. The owner does the device steps. The report holds no screenshot with personal data.

Result: go under the bar of D-118. Items 1 to 5 passed in the Home Screen app, and the median first contentful paint of 5 cold starts was 33 ms. The owner ran the checklist on an iPhone 16 Pro with iOS 27.0 and Chrome 154, in the Home Screen app only (D-119). The report is `docs/research/iphone-platform-spike.md`.

## 5. Exit of the phase

Phase 1 ends when PR-2, PR-4, and PR-6 give their reports. The owner reads the go or no-go of each report. A no-go changes the high-level roadmap before Phase 2 starts.
