# Workout App - design

This document holds the product thesis, the target experience, the system context, the safety boundaries, and the privacy posture of Workout App. It is the design reference for `docs/roadmaps/high-level-roadmap.md`. The date of this version is 2026-10-02.

Each statement has a label. **Fact** means a verified fact with a source in `docs/research/`. **Decision** means an owner decision in `docs/decisions.md`. **Recommendation** means a proposal that the owner did not accept yet. **Assumption** means a belief that nobody verified yet. **Open** means a question in `docs/questions.md`.

## 1. Product thesis

Workout App is a personal workout app for one person, the owner (Decision, D-67). The owner tells the app the muscles to train, the schedule, the experience, the injuries, and the goals. The owner selects each machine of the gym from a catalog, or enters it as text, and confirms it (Decision, D-49, D-110). No phase of the roadmap holds photo recognition now (Decision, D-111). The app builds a workout plan, guides each workout, records each set, and adapts the next targets from the history.

The core of the app is **a thin LLM over a strict policy** (Decision, D-22, D-23). OpenAI `gpt-6-luna` at medium effort proposes each plan and each revision (Decision, D-24). A deterministic, versioned policy checks every set, load, and change before the owner sees it. When the policy refuses a value of Luna, a rules fallback gives a safe target. When Luna gives no valid plan, the API retries the call, and after the last failure the request gives an error and changes nothing (Decision, D-230). This thesis copies the Decktome thesis, where deterministic code checks every card that the model names.

Workout App is an installable, phone-first web app on a default Firebase Hosting URL (Decision, D-17). It never goes to an app store, and no native app exists. It has one phone layout, and nobody designs or tests a desktop layout (Decision, D-20).

## 2. Target user and scope

| Item | Value | Label |
|---|---|---|
| User | The owner alone, ever | Decision, D-67 |
| Core profile | A 180 lb, 32-year-old man with prior machine experience, away from the gym for several years | Decision, D-31 |
| Experience level | Intermediate to advanced. Similar path for each, with more weight for advanced lifters. The core user starts as intermediate with his own load estimates. | Decision, D-30, D-32 |
| Out of scope | People under 18, people with no weight training | Decision, D-26, D-33 |
| Goals | General fitness and strength | Decision, D-31 |
| Equipment | Fixed-path resistance machines, dumbbells with an adjustable bench, the cable lat pulldown and triceps pulldown, and cardio machines. No other cable exercise, no barbell, no bodyweight exercise, and no band. D-155 gives the first catalog. | Decision, D-154, D-155 |
| Locale | US English, pounds only | Decision, D-28 |
| Lead platform | Chrome on iPhone. It uses the WebKit engine. | Decision, D-29 |
| Business model | Free, no ads, no sale of data | Decision, D-27 |

## 3. Target experience

### 3.1 Onboarding

The owner enters experience, goals, injuries and restrictions, age, height, weight, and cardio preference (Decision, D-41). The owner picks muscles directly or starts from a goal template, and can add free text for Luna (Decision, D-42). The screen asks about injuries and experience only. It has no readiness questions (Decision, D-34).

An injury answer gives a warning, and the plan avoids the injured area (Decision, D-35). The owner selects each injured area from a fixed list. A versioned table gives the areas that each exercise loads, and the server removes each such exercise before the call to Luna. The policy refuses it too (Decision, D-208).

The owner selects from ten muscle groups, or starts from the template "General fitness" or "Strength" (Decision, D-210). Onboarding asks the training days in each week, from 2 to 4, and a plan holds that count of sessions for one week (Decision, D-211).

The experience is "intermediate" or "advanced" (Decision, D-214). The server accepts an age of 18 to 90 years, a height of 48 to 96 in, and a weight of 80 to 500 lb. Each text has 500 characters or fewer (Decision, D-215). The cardio preference is a list of the cardio exercises that the owner likes (Decision, D-217). The template "Strength" selects the chest, the back, the shoulders, the quadriceps, the hamstrings, and the glutes (Decision, D-220).

After the sign-in, the app opens onboarding before the home screen while no profile exists. The home screen then opens the same screen to change the profile (Decision, D-223). After the owner selects an injured area, the screen shows: "Your plan avoids exercises that load the areas you selected. This app gives fitness guidance only, and it does not diagnose or treat an injury." (Decision, D-222).

The API stores the profile in one Firestore document at `users/{uid}/profile/active` (Decision, D-213). Section 5.15 of `docs/research/exercise-safety.md` gives the values of both tables and their research (Decision, D-218, D-219, D-221).

### 3.2 Equipment capture

The owner selects each machine from the catalog, or enters it as text (Decision, D-51, D-55, D-110). A text searches the names of the catalog, and the owner selects the match. A text with no match stays as a note in the inventory, and no plan uses a note (Decision, D-191). The catalog list shows the machines by kind: the machines, the cable station, the dumbbells, and the cardio machines (Decision, D-202). Each kind group and the inventory list show the machines A to Z by name (Decision, D-205).

The owner enters the weights of a stack as the lightest weight, the heaviest weight, and the step. The app makes the list, and the owner can add or remove a weight (Decision, D-195). The owner can give a current load estimate for each exercise of the machine (Decision, D-41, D-192).

An estimate is in the range of the weights of the machine, as the policy input reads it. The server refuses a change of the weights that puts an estimate outside that range (Decision, D-198). A weight of a stack is 1,000 lb or less, and a list holds 200 weights or fewer. A note holds 1 to 200 characters, and the inventory holds 50 notes or fewer (Decision, D-199).

The owner saves a machine as a draft, then confirms it on a review screen. A plan reads confirmed machines alone. A change of the weights makes a confirmed machine a draft again (Decision, D-49, D-193). A second save of a machine replaces its one entry, and a change of the estimates alone keeps the confirmation (Decision, D-200). The confirmation sends the weights that the review screen showed, and the server refuses it when the stored weights are different (Decision, D-201). In Phase 4, each change is a direct call to the API, and Phase 6 adds the offline copy (Decision, D-196).

No phase of the roadmap holds photo recognition now (Decision, D-111). The recognition spike gave a no-go, because Luna gave too many wrong answers with a high stated confidence (Decision, D-107, and `docs/research/recognition-spike.md`). When a later decision adds a photo phase, these rules apply:

- The owner takes one general photo of a machine, and the app asks for more photos only when the result is uncertain (Decision, D-50).
- The app deletes each source photo after the confirmation, unless the owner keeps it (Decision, D-52).
- No user photo goes into an evaluation set (Decision, D-53).

The equipment record holds the identity and the available weights of each machine (Decision, D-54). It also holds the estimates of D-192 and the state of D-193. The owner has one active inventory (Decision, D-46). The API keeps it in one Firestore document at `users/{uid}/inventory/active` (Decision, D-197). The function `inventory.ForPlan` gives the confirmed machines alone to a plan.

The catalog of D-155 gives each machine and each exercise a stable id, a kind, and a region. A machine with two movements gives two exercises (Decision, D-159). The six regions are the four regions of EV-1, core, and cardio (Decision, D-161). The adjustable bench is part of the dumbbell set, and the heaviest dumbbell is at most 100 lb (Decision, D-163, D-166). A load is in pounds, as a count of tenths of a pound, so 12.5 lb stays exact (Decision, D-122, D-160).

### 3.3 Plan

The plan adapts after each session and has no fixed block (Decision, D-43). It holds warm-up, resistance work, rest periods, cooldown, optional cardio, and mobility and recovery guidance (Decision, D-44). The guidance stays inside the fitness boundary (Decision, D-36). Instructions are text only (Decision, D-73). The owner can exclude an exercise with an optional reason, and Luna plans again under the policy (Decision, D-48).

A plan holds one session for each training day of one week (Decision, D-211). Each session holds 8 resistance exercises or fewer, and an optional cardio of 5 to 30 minutes (Decision, D-232, D-233). A plan starts each new exercise at 70 percent of its estimate, because the profile does not record a break (Decision, D-238).

The API keeps the plan at `users/{uid}/plan/active` and the exclusions at `users/{uid}/exclusions/active` (Decision, D-226). A new plan replaces the old plan only when its request completes (Decision, D-227). The reason of an exclusion has 200 characters or fewer, and it stays on the server (Decision, D-228, D-229). An exclusion and its new plan save together, or nothing changes (Decision, D-234). A request can take up to 4 calls of Luna, so the API streams each step, and the app shows the progress (Decision, D-231, D-237).

Luna writes one plan summary and one short reason for each exercise, with a length limit (Decision, D-182). Session titles come from a template. The warm-up, the cool-down, and the mobility and recovery texts come from a versioned catalog, and Luna selects each item by id (Decision, D-152). A filter of blocked claims reads each text of Luna, and a template text replaces a blocked text (Decision, D-183).

### 3.4 Guided workout

The workout screen shows one machine at a time. The design targets one-handed use: large targets, few taps, little typing, and tolerance of interruptions (Decision, D-71). The visual style is calm, focused, high-contrast, and minimal (Decision, D-70). Cues are visual only, with no audio and no vibration (Decision, D-58).

1. The owner logs reps, weight, and reps in reserve for each set. Pain (0 to 10) and a note are optional (Decision, D-57, D-162).
2. The rest timer starts when the owner logs a set. The owner can adjust or dismiss it (Decision, D-59).
3. After the last set of an exercise, a brief preview names the next machine. Then the app advances automatically (Decision, D-60).
4. The owner can edit a prior set or skip an exercise (Decision, D-63).
5. A busy machine gives a skip, not a substitute (Decision, D-47).
6. "Finish now" skips the remaining exercises and records the session as ended early (Decision, D-63).

The rest timer shows on the screen only, because the app sends no notifications (Decision, D-61). A Screen Wake Lock keeps the screen on during a workout (Recommendation, from `docs/research/platform-cloud-and-ai.md`).

### 3.5 Adaptation

After a session, Luna proposes the next targets from reps, load, reps in reserve, pain, skipped work, and gaps in the history (Decision, D-64). The policy checks the proposal. The app shows a concise reason that names the logged evidence (Decision, D-68). The owner can override a target, and the app keeps the recommendation, the override, and the reason as separate records (Decision, D-69). The engine handles missed sessions and long breaks (Decision, D-66).

After a gap of 14 days or more, the load goes down by the long-break table (Decision, D-151, D-179). The first sessions back stop at 3 reps in reserve, with rep progression only. A new exercise starts with a calibration set from the estimate of the owner, or from the lightest weight (Decision, D-150, D-177, D-178).

Targets use one to three reps in reserve. Failure is rare, and it never occurs in the first sessions after a break (Decision, D-37). Loads round to the nearest 5 lb, up or down (Decision, D-65). A rounded jump can exceed a validated target, so the policy adds at most one 5 lb step for each exercise in each session (Decision, D-147).

## 4. System context

```text
 Phone (installable web app, Chrome or Safari on iPhone)
   UI, IndexedDB store and outbox, wake lock
        |  HTTPS, Connect-RPC unary calls, Firebase ID token
        v
 Cloud Run API (Go, default run.app URL, CORS to the app origin)
   auth interceptor -> handlers -> policy engine -> role layer -> OpenAI gpt-6-luna
        |
        v
 Firestore (server only)
 Cloud Scheduler -> Cloud Run jobs (maintenance)
 Firebase Hosting (static app), Firebase Authentication (email and password)
```

| Part | Role | Sensitivity | Label |
|---|---|---|---|
| Web app | Screens, offline store, outbox | Holds workout history and profile data on the phone | Decision, D-17, D-62, D-77 |
| API | Auth check, sync, plan calls, policy. It refuses to start on Cloud Run with an emulator variable. | Reads and writes all user data | Decision, D-18, D-82, D-129 |
| Policy engine | Checks every prescription | Safety-critical | Decision, D-23 |
| Role layer | Model choice, the plan schema, the prompt, cost records, the cap hook, the text filter, and the fake provider | Sends workout data to OpenAI | Decision, D-24, D-25, D-183 |
| Firestore | Source of record after sync | Injuries, body data, workout history, and the equipment inventory with its notes | Decision, D-77, D-197 |
| Firebase Auth | Email and password, and an invite allowlist of uids in Firestore | Email address | Decision, D-75, D-131 |
| GCP project | One project, `nk-workout-app-prod`, in `us-central1` | All of the above | Decision, D-76, D-137 |

The deferred photo work adds camera input, Cloud Storage for short-lived photos, and a photo purge job (Decision, D-52, D-111).

The web client stack follows Decktome by default (Decision, D-74). After research on alternatives, the owner chose the Decktome React stack with Vite 8 (Decision, D-84). The native framework comparison of the launch prompt does not apply (Decision, D-17).

### 4.1 Limits of a web app on iOS

The owner accepted these limits (Decision, D-21). The facts come from `docs/research/platform-cloud-and-ai.md`.

| Limit | Effect on Workout App | Label |
|---|---|---|
| No Vibration API in Safari on iOS | No haptic cues. D-58 chose visual cues only. | Fact |
| No Background Sync on iOS | The app syncs when it is open, visible, and online. | Fact |
| Web Push only for Home Screen apps | Not relevant. D-61 chose no notifications. | Fact |
| Home Screen app storage is separate from Safari storage | The owner signs in again inside the installed app. | Fact |
| The OS can evict web storage under storage pressure | The server copy stays the durable record. | Fact |
| No HealthKit, Health Connect, watch, or Live Activities | Q-49 is not applicable. | Fact |
| A Home Screen app with a translucent status bar can end short of the bottom edge, and a pinch can zoom the page | The shell fills the whole screen, and the viewport meta blocks the pinch zoom (D-120). | Fact, from `docs/research/iphone-platform-spike.md` |
| Since iOS 26, a Home Screen app blurs a band of the page below the status bar | The header adds 16 px above the title when a status bar covers the page (D-145). | Fact, from `docs/research/phase-2-check.md` |

## 5. Safety boundaries

### 5.1 Guardrails

These rules hold for every phase. The label names the source of each rule.

1. Every set, load, and change passes the policy before the owner sees it. The policy wins over Luna (Decision, D-23).
2. No model id appears at a call site. Models come from the role layer (Decision, D-24).
3. Every external service ships with a local fake in the same pull request (Recommendation, from the Decktome pattern).
4. The app gives fitness guidance only. Luna text never diagnoses, treats, or prescribes rehabilitation (Decision, D-36).
5. The owner confirms every machine before a plan uses it (Decision, D-49).
6. Each paid AI run in development needs owner approval. Production has caps (Decision, D-25).
7. Logs, metrics, and error reports hold ids only. No workout text, photos, prompts, or health details (Decision, D-80).
8. Only `main` deploys (Decision, D-14).
9. When a later phase adds photos, the app strips photo metadata and deletes each photo after the confirmation (Decision, D-52, D-111).
10. A logged set survives a lost connection and an app restart (Decision, D-62).

The policy is in `go/internal/policy` (Decision, D-157). It has one version, and each rule has an id and the decisions and evidence that support it (Decision, D-38). It holds these rules:

- The bounds of a target: the reps, the reps in reserve, the rest, and a weight of the machine (Decision, D-37, D-54, D-167, D-171, D-172).
- The rounding of a load to 5 lb, with the halfway rule and the weight of the machine (Decision, D-65, D-148, D-149).
- The next target: double progression with one 5 lb step, missed reps, pain, a lighter weight, and sets with no log (Decision, D-147, D-168 to D-170, D-173, D-174).
- The fixed warning text of a pain report (Decision, D-153, D-169).
- The check of a proposal. The policy refuses a proposal outside a bound (Decision, D-23). Outside a calibration session, it also refuses a proposal that is harder than its target at the same load (Decision, D-186).
- The start of a new exercise with 3 working sets, and the calibration of its first 3 sessions (Decision, D-150, D-177, D-178, D-180).
- The return after a break of 14 days or more, and the first sessions after it (Decision, D-37, D-151, D-179).
- The rules fallback of one exercise. When the policy refuses a proposal, or Luna gives none for the exercise, the target comes from the rules alone (Decision, D-23). A plan request with no valid output of Luna gets no fallback plan (Decision, D-230).
- The decision record of each plan decision, with the fields of D-176. The record is workout data, so it never goes into a log (Decision, D-80, D-176).

The policy has no reactive deload (Decision, D-175). The draft in `tools/spikes/luna_plan/policy.py` is the spike record alone.

The role layer is in `go/internal/ai` (Decision, D-157). It holds these parts:

- The planner and the reviser roles on `gpt-6-luna` at medium effort. The role holds the model id, so a call site names a role alone (Decision, D-22, D-24).
- The strict plan schema. Its enums hold the ids of the catalog of D-155 and of the guidance catalog, so a valid output names no unknown id (Decision, D-152).
- The prompt, with the boundary of D-36 and the dated copy of the usage policies of D-93, and each rule of the policy.
- A cost record for each call, and a cap hook that reserves the worst-case cost before the call. The hook refuses a call over the cap. The cap values come from the configuration. The caps are 1 USD for the user and 2 USD for the project, for each calendar month in UTC (Decision, D-25, D-188, D-190). The API uses the lasting cap hook of `go/internal/capstore`. It keeps the spend of each month in Firestore, so a new instance of the API does not reset it (Decision, D-189, D-224). A failed call charges the reserved worst-case cost (Decision, D-225).
- The filter of blocked claims on each text of Luna (Decision, D-183).
- The fake provider for tests. No test calls OpenAI (Decision, D-24).

A malformed output, a refusal of the model, a time-out, or an error gives no plan. The plan API of `go/internal/plan` then retries the call. Each retry sends the cause and the failed output, and Luna makes a fresh plan (Decision, D-235). A request makes 4 calls at most. After the last failure, or after a call over the cap, the request gives an error and no plan changes (Decision, D-230, D-231).

Each failed attempt gets an error record in the top-level collection `aiErrors` (Decision, D-236). In a valid plan, the policy decides each exercise, and a refused proposal gets the rules target (Decision, D-23, D-176).

The command `go/cmd/lunaeval` sends synthetic profiles and the scenarios of section 5 of the high-level roadmap through the layer and the policy. A live run needs the approval of the owner (Decision, D-25). `docs/research/phase-3-check.md` holds the result of the Phase 3 run.

### 5.2 Accepted risks

The owner chose these options against the launch prompt recommendations. The research in `docs/research/exercise-safety.md` records the evidence behind each recommendation.

| Decision | What the evidence recommends | What the app does |
|---|---|---|
| D-34 | A readiness screen, such as PAR-Q+, that finds reasons to stop and refer | Asks about injuries and experience only |
| D-35 | Stop plan generation for listed risk answers | Warns and continues with a conservative plan |
| D-39 | Review of the policy by a qualified person | No qualified human review |
| D-40 | Stop the workout for chest pain, fainting, or severe breathlessness | Warns, and the owner can continue after a confirmation |
| D-65 | Load steps that match each machine | Rounds to the nearest 5 lb, up or down |
| D-72 | Accessibility from the start | Accessibility work is deferred |
| D-169 | A hold of 2 sessions after a pain report (REC-9) | Holds progression for the next session alone |
| D-171 | A return to the last load when the first set after a dumbbell step is too hard (REC-18) | Keeps the new load, and the rules of D-168 apply |

### 5.3 Medical boundary

The app states that it gives fitness guidance only (Decision, D-36). The FDA general wellness guidance of January 2026 treats claims about strength and muscle size as wellness claims (Fact). Claims about disease, treatment, or rehabilitation fall outside it (Fact). The policy and the prompts must keep Luna text inside that line.

The prompt follows the dated copy of the OpenAI usage policies of D-93 (Decision, D-93). Luna writes short texts alone, and the filter of blocked claims replaces a text with a medical, diet, or emergency claim (Decision, D-182, D-183). The filter is a guard after the prompt, not a proof: a claim in other words can pass it.

## 6. Privacy posture

Workout App stores data about one person, the owner (Decision, D-67). The owner chose the minimum compliance posture (Decision, D-79) and no user data controls (Decision, D-78). The project hosts no public policy pages (Decision, D-81). These rules still hold:

- The repository is public. No personal data, email address, photo, or workout log goes into it. The author credit that the license of a test image requires is the one exception (Decision, D-106).
- Telemetry holds ids only (Decision, D-80).
- A planner call sends the experience, the goal template, the muscle groups, the free text, and the cardio preference alone. The age, the height, the weight, the injury text, and the reason of an exclusion stay on the server (Decision, D-209, D-229).
- An error record holds the output of a failed call of Luna, which can repeat the free text. So no log, metric, or error report reads it, and a Firestore TTL deletes it after 90 days (Decision, D-80, D-236).
- The API sends OpenAI requests with the response store turned off (Recommendation, from `docs/research/platform-cloud-and-ai.md`). The OpenAI provider of `go/internal/ai` does this. A call sends no note of the owner and no user id.
- The app takes no photo now (Decision, D-110). In a later photo phase, the app removes photo metadata before upload, and the server deletes each photo after the confirmation (Decision, D-52).
- Secrets live in Secret Manager, never in the repository.

A change of audience reopens D-78, D-79, D-81, D-34, D-39, and D-40. The high-level roadmap names this reopening gate.

## 7. Open questions

Q-91 to Q-107 in `docs/questions.md` hold the open questions. Q-90 has an answer (D-84). The high-level roadmap names the phase that needs each answer.
