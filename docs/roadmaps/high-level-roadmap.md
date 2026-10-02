# Workout App - high-level roadmap

This roadmap gives the path from a blank repository to a safe, useful Workout App for its one user, the owner (D-67). It names the phases, their order, the risk work, the outcomes, the work areas of pull request size, and the exit evidence of each phase. It is not a plan of tasks. A focused roadmap turns one phase into tasks later. `docs/roadmaps/README.md` gives the rules for focused roadmaps.

The date of this version is 2026-10-02. `docs/design.md` holds the product design. `docs/decisions.md` and `docs/questions.md` hold every decision and question that this roadmap cites.

## 1. Rules of this roadmap

- A work area has the size of one pull request. It holds one milestone with one acceptance story, and it can hold two, three, or more concerns (D-10, D-12).
- Work areas have no pull request numbers (D-9). A focused roadmap names the pull requests.
- One clean session works on one pull request. The owner approves the work before it starts (D-12).
- A phase starts only after the exit evidence of each phase that it depends on is complete.
- A paid check runs only after the owner approves it (D-25). Section 6 lists the free and the paid checks.
- The owner ended the D-4 period on 2026-09-29 (D-125). A cross-provider review is necessary for a change of code or safety behavior (D-15).
- The owner confirms every merge (D-13). Only `main` deploys (D-14).

## 2. Scope that shapes the phases

These decisions remove whole areas of work from the roadmap.

| Decision | Effect on the roadmap |
|---|---|
| D-17 | No native apps, no app store, no store review, no app signing, no TestFlight, and no Play tracks. |
| D-67 | One user. No invite beta and no public release. Section 7 gives the gate that reopens this scope. |
| D-78, D-79, D-81 | No user data controls, minimum compliance, and no public policy pages. |
| D-39 | No qualified human review gate. |
| D-61 | No notifications and no Web Push work. |
| D-72 | Accessibility work is deferred. Semantic markup stays a code-review item, not a phase. |
| D-47 | No exercise substitution. A busy machine gives a skip. |
| D-76 | One development project. No production project. |

## 3. Phase map

| Phase | Name | Depends on | Main outcome |
|---|---|---|---|
| 0 | Foundation | none | Registers, research, process tooling, and this roadmap on `main` |
| 1 | Risk spikes | 0 | Measured answers about Luna plans, photo recognition, and the iPhone web platform |
| 2 | Platform skeleton | 1 | A signed-in, installable app shell on the development project, deployed from `main` |
| 3 | Workout domain and safety policy | 1, 2 | A versioned policy engine and the Luna role layer, with golden scenarios |
| 4 | Equipment inventory | 2, 3 | Confirmed machines from manual selection or text entry (D-110) |
| 5 | Onboarding and plan generation | 3, 4 | A validated plan from the profile and the confirmed inventory |
| 6 | Guided workout and offline logging | 5 | A fast, one-handed workout flow that survives a lost connection |
| 7 | Adaptation loop | 6 | Validated next-session targets with reasons, overrides, and break handling |
| 8 | Personal-use operations | 7 | Backups, alerts, cost caps, incident steps, and version migration |

```text
0 -> 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7 -> 8
          \______^
```

Phase 3 depends on Phase 1 and Phase 2. Phase 1 gives the policy evidence and the Luna spike. Phase 2 gives the Go module and the contract.

## 4. Phases

### Phase 0 - Foundation

**Objective.** Put the registers, the research, the process tooling, and this roadmap on `main`.

**Outcome.** A later session can start from `docs/session-handoff.md` and find every rule, decision, and open question. The process gates of Decktome run in this repository without Gitar (D-3, D-6).

| Work area | Concerns | Exit evidence |
|---|---|---|
| 0.1 Foundation documents and process tooling | Registers and design. Research and this roadmap. STE, reference, contract, and review gates with tests. | This pull request merges with a green CI and a Codex approval. |
| 0.2 Live ruleset | Apply the committed ruleset of `main` after the owner approves (Q-94). | `make ruleset-check` passes against the live ruleset. |

**Decisions and questions.** D-1 to D-15, D-83, D-91, D-92, Q-93, Q-94.

### Phase 1 - Risk spikes

**Objective.** Measure the three largest risks before the project commits to an architecture or to content.

**Depends on.** Phase 0.

**Outcome.** Three short spike reports with numbers, and a go or no-go for each risk. Spike code lives in a tools folder, not in the product. The spike code has a fake provider for each paid call.

| Work area | Concerns | Exit evidence |
|---|---|---|
| 1.1 Luna plan spike | Structured plan output from a profile and an inventory. The first draft of the policy rules table. Rejection counts. | A report on a fixed set of profiles: schema pass rate, policy rejection rate, cost per plan, and the unsafe proposals that the draft rules catch. |
| 1.2 Recognition spike | A test set of public or licensed machine images (D-56, Q-97). Luna identification against a catalog shortlist. Cost per photo (Q-96). | A report with correct, wrong, and abstained counts, the wrong answers with high stated confidence, and the cost per photo. |
| 1.3 iPhone web platform spike | IndexedDB durability after an app kill, Screen Wake Lock, install and sign-in in the Home Screen app from Chrome on iPhone (D-29). Startup time of a React build (D-84). The development project of D-76 with Firebase Hosting and Firebase Authentication only (D-99). | A device checklist with a result for each item on the owner's iPhone, and a startup time. |

**Decisions and questions.** D-22 to D-24, D-29, D-53, D-56, D-84, D-93 to D-100, Q-95, Q-96, Q-97.

**Focused roadmap.** `docs/roadmaps/phase-1-risk-spikes.md`.

**Gate.** The owner approves each paid spike run (D-25). A no-go result changes this roadmap before Phase 2 starts.

### Phase 2 - Platform skeleton

**Objective.** Build the smallest running system on the Decktome stack (D-74), with the React web client of D-84.

**Depends on.** Phase 1.

**Outcome.** The owner installs the app on the iPhone, signs in, and reaches an empty home screen. The API runs on Cloud Run in the development project (D-76). A merge to `main` deploys both parts (D-14, D-18).

| Work area | Concerns | Exit evidence |
|---|---|---|
| 2.1 Contract and API skeleton | Protobuf contract and generated code. Go API with the Firebase token check and the invite allowlist (D-75). Emulators and fakes for local work. | Product `make` targets run contract, Go, and emulator tests for free. `make verify` stays Python only (D-113, D-126). |
| 2.2 Installable web shell | Phone layout only (D-20). A shell that fills the whole screen and blocks the pinch zoom (D-120), with the title out of the iOS blur band (D-145). Sign-in. Offline store and outbox skeleton (D-62, D-77). | Browser tests pass in the WebKit and Chromium engines. The owner installs and signs in on the iPhone, and sees no gap at the bottom edge and no pinch zoom. |
| 2.3 Development project and deploy | Firestore, Cloud Run, Secret Manager, and a budget alert in `us-central1`, in the one project of D-137, which replaced the project of work area 1.3. Cloud Build deploy from `main`. Point-in-time recovery and a daily backup (D-124). The billing link of Q-142. | A merge deploys. The live version endpoint names the merged commit. |

**Decisions and questions.** D-17, D-18, D-20, D-62, D-74 to D-77, D-80, D-82, D-84, D-99, D-120, D-124 to D-128, D-137, D-138, D-144 to D-146, Q-99, Q-142.

**Focused roadmap.** `docs/roadmaps/phase-2-platform-skeleton.md`.

### Phase 3 - Workout domain and safety policy

**Objective.** Build the part that makes every prescription safe: the domain model, the deterministic policy, the rules fallback, and the Luna role layer (D-23, D-24).

**Depends on.** Phase 1 and Phase 2.

**Outcome.** A Go package takes a Luna proposal and returns accepted targets, corrected targets, or a refusal with a reason. Every rule has a version and cites evidence ids from `docs/research/exercise-safety.md` (D-38). Each plan decision records the policy version, the model id, and a hash of the prompt.

| Work area | Concerns | Exit evidence |
|---|---|---|
| 3.1 Domain model and catalog | Machines, exercises, inventory, plan, session, and set log. The catalog of D-155: machines, two cable exercises, dumbbells, and cardio machines (D-154). The cardio log fields of D-123. Machines with pound markings only (D-122). | Table tests for every type and for the catalog lookups. |
| 3.2 Policy engine and fallback | Rep, reps-in-reserve, rest, and load bounds (D-37). The 5 lb rounding of D-65, with D-148 and D-149, and the one 5 lb step of D-147. The rules fallback. | Golden scenario tests pass, section 5 scenarios included. Property tests prove that no output breaks a bound. |
| 3.3 Luna role layer | Planner and reviser roles on `gpt-6-luna` at medium effort. The fake provider. Cost records, cap hooks, and a prompt that keeps text inside the fitness boundary (D-36, Q-95, Q-101). | Fake-provider tests cover a valid proposal, a malformed proposal, an unsafe proposal, and a timeout. |

**Decisions and questions.** D-22 to D-25, D-30, D-32, D-36 to D-38, D-40, D-43, D-45, D-64 to D-66, D-122, D-123, D-147 to D-158, D-184 to D-187, Q-92, Q-95, Q-100 to Q-102, Q-104 to Q-107.

**Gate.** A limited paid evaluation of the planner and the reviser runs only with owner approval (D-25).

### Phase 4 - Equipment inventory

**Objective.** Turn manual selections and text entries into confirmed machines (D-49, D-51, D-55). The recognition spike gave a no-go, so this phase has no photo recognition (D-110). Section "Deferred - Photo recognition" holds the photo work.

**Depends on.** Phase 2 and Phase 3.

**Outcome.** The owner builds the one active inventory (D-46). Every machine carries an identity, the available weights, and the owner's load estimate for each exercise (D-41, D-54, D-192).

| Work area | Concerns | Exit evidence |
|---|---|---|
| 4.1 Selection and text entry | The owner selects each machine from the catalog of D-155, in the order of D-202 and D-205, or enters it as text that searches the catalog (D-51, D-55, D-191). The owner confirms each machine before a plan uses it (D-49, D-193). | Browser UI tests add a machine by selection and by text entry. A plan can not use a machine that the owner did not confirm. After the merge, the owner confirms one machine in the live app on the iPhone (D-203, D-204). |
| 4.2 Machine details | Identity and available weights only (D-54, D-195). The load estimates of the owner, in the range of the weights (D-41, D-192, D-198). Pound markings alone (D-122). The store of D-197, with the bounds of D-199 and the confirmation of D-200 and D-201. | An emulator test proves that a stored machine holds only the fields of D-54, the load estimates, and its state (D-193). |

**Decisions and questions.** D-41, D-45, D-46, D-154, D-155, D-49, D-51, D-54, D-55, D-110, D-122, D-191 to D-206, Q-91, Q-205 to Q-220.

### Phase 5 - Onboarding and plan generation

**Objective.** Produce the first validated plan from the profile and the confirmed inventory.

**Depends on.** Phase 3 and Phase 4.

**Outcome.** The owner completes onboarding (D-34, D-41, D-42) and sees a plan that the policy accepted (D-23). The plan holds warm-up, work sets, rest, cooldown, optional cardio, and mobility and recovery text (D-44, D-73).

| Work area | Concerns | Exit evidence |
|---|---|---|
| 5.1 Onboarding | Profile inputs, with the training days of D-211. The injury warning of D-35 and the injury areas of D-208. The muscle groups and the goal templates of D-210, with free text (D-42). | UI tests cover each input and the injury warning. |
| 5.2 Plan generation and view | The planner call with the input of D-209, the policy check, the fallback, and exclusions (D-48). The lasting store of the monthly AI caps of D-188 and D-190, before the first live call (D-189). | An end-to-end test with the fake provider returns a valid plan for the core profile of D-31. A cap test refuses a call over the cap after a restart of the API. After the deploy, the owner approves the cost and sees a plan in the live app on the iPhone (D-207, D-212). |

**Decisions and questions.** D-25, D-31 to D-36, D-41 to D-44, D-48, D-73, D-188 to D-190, D-207 to D-212, Q-221 to Q-226.

### Phase 6 - Guided workout and offline logging

**Objective.** Deliver the workout flow for one-handed use in a gym with poor signal (D-62, D-71).

**Depends on.** Phase 5.

**Outcome.** The owner completes a full workout on the iPhone with no connection. The logs reach Firestore when the app is open and online again (D-77).

| Work area | Concerns | Exit evidence |
|---|---|---|
| 6.1 Workout screen and set log | Set log of D-57. Visual cues only (D-58). Warning flow of D-40. Wake lock. | The owner logs a set in three taps or fewer. UI tests pass. |
| 6.2 Rest timer and automatic advance | Timer from a stored end time (D-59). Preview, then advance (D-60). Edit, skip, and "finish now" (D-63). | UI tests prove that the timer survives a screen lock and that the advance happens after the last set. |
| 6.3 Outbox sync | Client operation ids, idempotent unary sync, sync on open, on focus, and on reconnect. | Offline tests replay a full workout with a dropped connection and an app kill, and the server holds each set once. |

**Decisions and questions.** D-21, D-40, D-57 to D-63, D-70, D-71, D-77.

### Phase 7 - Adaptation loop

**Objective.** Adapt the next session from the logged history under the policy (D-43, D-64).

**Depends on.** Phase 6.

**Outcome.** After each session, the owner sees validated targets for the next session, each with a concise reason (D-68). Overrides keep the recommendation and the reason (D-69). Missed sessions and long breaks change the targets (D-66).

| Work area | Concerns | Exit evidence |
|---|---|---|
| 7.1 Revision after a session | The reviser call and the policy check. Reasons that cite logged sets. | The section 5 scenarios pass end to end with the fake provider. |
| 7.2 Overrides and disruptions | Separate override records. Missed-session and long-break rules (Q-102). | Scenario tests for a missed week and for a break of the Q-102 length. |

**Decisions and questions.** D-37, D-43, D-64 to D-69, Q-92, Q-102, Q-202.

**Gate.** A paid evaluation of the reviser on the section 5 scenarios runs only with owner approval (D-25).

### Phase 8 - Personal-use operations

**Objective.** Keep the one copy of the owner's history safe, the cost bounded, and the model current.

**Depends on.** Phase 7.

**Outcome.** The system runs for weeks with no data loss, no surprise bill, and a known path to a new model or a new policy version.

| Work area | Concerns | Exit evidence |
|---|---|---|
| 8.1 Backups and recovery | The backups of D-124 from work area 2.3. A restore drill into a new database. A rollback procedure for the API and the web app. | A restore drill report. |
| 8.2 Alerts and cost caps | Error and job-failure alerts. Spend caps on the development project. Work area 5.2 holds the lasting cap store (D-189). | A test alert reaches the owner. |
| 8.3 Version migration | Policy and evidence version changes. A model change when `gpt-6-luna` retires. An incident runbook. | A replay of stored sessions under a new policy version gives a diff report. |

**Decisions and questions.** D-25, D-76, D-80, D-124, D-188, D-189, Q-98, Q-99.

**Exit for the roadmap.** Four weeks of owner use with no lost set, no refused valid sync, and no policy breach in the decision log.

### Deferred - Photo recognition

The recognition spike of work area 1.2 gave a no-go under D-107. Luna gave a wrong answer with a high stated confidence for 8.6% of the photos (`docs/research/recognition-spike.md`). So no phase holds photo recognition (D-110, D-111). A later owner decision adds a phase for it. Before that phase starts, a recognition method must pass the bar of D-107 on the recognition test set.

The deferred work areas:

| Work area | Concerns | Exit evidence |
|---|---|---|
| Capture and upload | One photo first, more on doubt (D-50). Metadata removal on the phone. Short-lived upload URL. | Unit tests prove that an uploaded image has no location metadata. |
| Identification and confirmation | Luna identification against the catalog. Confirm, correct, retry, manual selection, and text entry on every screen (D-51, D-55). | The recognition test set passes the bar of D-107. |
| Photo deletion | Deletion after confirmation (D-52). A scheduled purge job. The kept-photo rule of Q-103. | An emulator test proves that no source photo outlives its confirmation, except a kept photo. |

**Decisions and questions.** D-50, D-52, D-53, D-56, D-107, D-111, Q-103.

## 5. Adaptation scenarios

The policy of Phase 3 must pass these scenarios. The numbers come from the synthesis in `docs/research/exercise-safety.md`. The owner fixed the step, the rounding, the calibration, and the break rules in the session of the Phase 3 focused roadmap (D-147 to D-151). In work area 3.2, the owner fixed the numbers of scenarios A to D (D-167 to D-174), and of scenarios E and F (D-175 to D-180). The other numbers stay recommendations until a later pull request fixes them with the owner.

| Id | Prescribed | Logged | Safe next-session behavior |
|---|---|---|---|
| Scenario A | 3 x 12 at 25 lb | 12, 12, 5 | No load increase. If the owner reported pain, the pain rule applies first. If sets 1 and 2 had one rep in reserve or less, keep 25 lb and lower the rep target, or lower the load one 5 lb step. If the same shortfall occurs in two sessions, lower the load. The reason names set 3. |
| Scenario B | 3 x 12 at a load | 12, 12, 12, each with three or more reps in reserve | Progress with double progression (D-147). Below the top of the rep range, add reps. At the top of the range, add one 5 lb step and reset the reps to the low end of the range. |
| Scenario C | Any | A pain flag on a set | No progression on that exercise in the next session. The warning of D-40 shows, with the text of D-153. The text never diagnoses (D-36). |
| Scenario D | Any | The session ended early | Unlogged sets count as skipped work, not as failed reps (D-63, D-64). |
| Scenario E | Any | No session for 2 weeks or more (D-151) | Lower the load with the long-break table of D-151, and use 3 reps in reserve with rep progression only for the first sessions back. No failure in those sessions (D-37). |
| Scenario F | Luna proposes a 50 percent load jump | Any | The policy refuses the proposal, the fallback target applies, and the decision log records the refusal. |

## 6. Test and evaluation strategy

### 6.1 Checks

| Check | What it proves | Cost | Phase |
|---|---|---|---|
| STE, reference, contract, and context checks | Documents follow the rules, and each id and path resolves | Free | 0 and later |
| Deterministic policy tests | Table, golden, and property tests of every rule | Free | 3 |
| Fake-provider tests | Behavior for valid, malformed, unsafe, and slow Luna output | Free | 3 and later |
| Emulator integration tests | API, Firestore, Storage, and Auth behavior on local emulators | Free | 2 and later |
| Offline and sync tests | A replayed workout with lost connections gives one copy of each set | Free | 6 |
| Browser UI tests | Flows in the WebKit and Chromium engines with phone emulation | Free | 2 and later |
| Security and privacy checks | No secrets or personal data in the repository, no sensitive fields in logs | Free | 2 and later |
| Real iPhone checklist | Home Screen install, storage, wake lock on the owner's phone | Free, manual | 1, 6 |
| Codex review | Cross-provider review of each pull request (D-4, D-15) | Codex plan (D-8) | All |
| Luna plan and revision evaluation | Schema pass rate, policy rejection rate, scenario results with the live model | Paid, owner approval | 1, 3, 7 |
| Recognition evaluation | Correct, wrong, and abstained counts on licensed images | Paid, owner approval | 1, and the deferred photo work |

D-72 defers accessibility tests.

### 6.2 Recognition evaluation

The test set holds public or licensed images only (D-56). The main risk is a confident wrong identification. The spike report and the deferred photo work count these results for each image:

- Correct identity.
- Wrong identity. A separate count holds each wrong identity that Luna gave with high stated confidence.
- Abstention, with a request for a new photo or manual entry (D-55).
- Confusion between similar machines, as a list of pairs.

The owner confirms every machine (D-49), so a wrong identity costs time, not safety. The gate still requires a low wrong-identity count, because a wrong machine gives a wrong load range.

### 6.3 Adaptation evaluation

The Phase 3 and Phase 7 suites prove these properties for every input:

- No load increase follows a set with missed reps.
- No output adds more than one 5 lb step for each exercise in each session (D-147).
- No output exceeds the rep bounds or the reps-in-reserve bounds.
- Every load is a multiple of 5 lb (D-65), or the machine weight that D-149 selects. A dumbbell load is the load of one dumbbell (D-155).
- A pain flag blocks progression on that exercise in the next session.
- Every Luna proposal outside the bounds becomes a refusal and a fallback, never a displayed target.
- Outside a calibration session, no accepted proposal is harder than the policy target at the same load (D-186).
- The same history and the same policy version give the same validated targets.

## 7. Scope reopening gate

The launch prompt asked for the evidence to move from personal use to an invite beta and then to a public release. D-67 closes both moves. If the owner opens the app to another person, a new owner decision must first reopen these items:

| Item | Current decision | Evidence needed before another user |
|---|---|---|
| Readiness screen | D-34 | A readiness screen with stop and refer rules |
| Warning symptoms | D-40 | A stop rule for cardiac warning signs |
| Expert review | D-39 | A review of the policy by a qualified person |
| Data controls | D-78 | Export, photo deletion, history deletion, and account deletion |
| Compliance | D-79 | A current check of US health-data laws for the new users |
| Public pages | D-81 | A privacy notice on the default URL |
| Accessibility | D-72 | An accessibility baseline and tests |
| Production project | D-76 | A separate production project with backups |

## 8. Completion and focused roadmaps

This high-level roadmap is complete when Phase 0 merges and the owner confirms the phase order. Both conditions hold on 2026-09-28 (D-91, D-92). After that, each phase gets one focused roadmap before its first pull request. `docs/roadmaps/README.md` gives the format. A focused roadmap can split or merge work areas. A change of phase order or of scope needs an owner decision and an update of this file.
