# Workout App - design

This document holds the product thesis, the target experience, the system context, the safety boundaries, and the privacy posture of Workout App. It is the design reference for `docs/roadmaps/high-level-roadmap.md`. The date of this version is 2026-10-04.

Each statement has a label. **Fact** means a verified fact with a source in `docs/research/`. **Decision** means an owner decision in `docs/decisions.md`. **Recommendation** means a proposal that the owner did not accept yet. **Assumption** means a belief that nobody verified yet. **Open** means a question in `docs/questions.md`.

## 1. Product thesis

Workout App is a personal workout app for one person, the owner (Decision, D-67). The owner tells the app the muscles to train, the schedule, the experience, the injuries, and the goals. The owner selects each machine of the gym from a catalog, or enters it as text, and confirms it (Decision, D-49, D-110). No phase of the roadmap holds photo recognition now (Decision, D-111). The app builds a workout plan, guides each workout, records each set, and adapts the next targets from the history.

The core of the app is **a thin LLM over a strict policy** (Decision, D-22, D-23). OpenAI `gpt-6-luna` at xhigh effort proposes each plan and each revision (Decision, D-24, D-253). A deterministic, versioned policy checks every set, load, and change before the owner sees it. When the policy refuses a value of Luna, a rules fallback gives a safe target. When Luna gives no valid plan, the API retries the call, and after the last failure the request gives an error and changes nothing (Decision, D-230). This thesis copies the Decktome thesis, where deterministic code checks every card that the model names.

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

The owner enters the weights of a stack as the lightest weight, the heaviest weight, and the step. The save makes the list, and on a later visit the owner can add or remove a weight (Decision, D-195, D-245). The owner can give a current load estimate for each exercise of the machine (Decision, D-41, D-192).

An estimate is in the range of the weights of the machine, as the policy input reads it. The server refuses a change of the weights that puts an estimate outside that range (Decision, D-198). A weight of a stack is 1,000 lb or less, and a list holds 200 weights or fewer. A note holds 1 to 200 characters, and the inventory holds 50 notes or fewer (Decision, D-199).

The owner saves a machine as a draft, then confirms it on a review screen. A plan reads confirmed machines alone. A change of the weights makes a confirmed machine a draft again (Decision, D-49, D-193). A cardio machine has no weights, so its save confirms it (Decision, D-246).

A second save of a machine replaces its one entry, and a change of the estimates alone keeps the confirmation (Decision, D-200). The confirmation sends the weights that the review screen showed, and the server refuses it when the stored weights are different (Decision, D-201). In Phase 4, each change was a direct call to the API. Since Phase 6, the phone keeps an offline copy of the inventory, and each change goes into the outbox (Decision, D-196, D-250, D-272). A confirmation with no connection waits in the outbox. When the stored weights are different at the sync, the server refuses it, and the machine is a draft again (Decision, D-273).

No phase of the roadmap holds photo recognition now (Decision, D-111). The recognition spike gave a no-go, because Luna gave too many wrong answers with a high stated confidence (Decision, D-107, and `docs/research/recognition-spike.md`). When a later decision adds a photo phase, these rules apply:

- The owner takes one general photo of a machine, and the app asks for more photos only when the result is uncertain (Decision, D-50).
- The app deletes each source photo after the confirmation, unless the owner keeps it (Decision, D-52).
- No user photo goes into an evaluation set (Decision, D-53).

The equipment record holds the identity and the available weights of each machine (Decision, D-54). It also holds the estimates of D-192 and the state of D-193. The owner has one active inventory (Decision, D-46). The API keeps it in one Firestore document at `users/{uid}/inventory/active` (Decision, D-197). The function `inventory.ForPlan` gives the confirmed machines alone to a plan.

The catalog of D-155 gives each machine and each exercise a stable id, a kind, and a region. A machine with two movements gives two exercises (Decision, D-159). The six regions are the four regions of EV-1, core, and cardio (Decision, D-161). The adjustable bench is part of the dumbbell set, and the heaviest dumbbell is at most 100 lb (Decision, D-163, D-166). A load is in pounds, as a count of tenths of a pound, so 12.5 lb stays exact (Decision, D-122, D-160).

### 3.3 Plan

The plan adapts after each session and has no fixed block (Decision, D-43). It holds warm-up, resistance work, rest periods, cooldown, cardio, and mobility and recovery guidance (Decision, D-44, D-255). The guidance stays inside the fitness boundary (Decision, D-36). Instructions are text only (Decision, D-73). The owner can exclude an exercise with an optional reason, and Luna plans again under the policy (Decision, D-48).

A plan holds one session for each training day of one week (Decision, D-211). Each session holds 8 resistance exercises or fewer (Decision, D-233). When the cardio preference names a cardio exercise, each session ends with 20 to 30 minutes of one. An empty preference gives no cardio (Decision, D-217, D-255). A plan starts each new exercise at its estimate, and its first set is the calibration (Decision, D-297, D-300). An exercise with logged history gets its target from that history (Decision, D-301).

The API keeps the plan at `users/{uid}/plan/active` and the exclusions at `users/{uid}/exclusions/active` (Decision, D-226). A new plan replaces the old plan only when its request completes (Decision, D-227). The reason of an exclusion has 200 characters or fewer, and it stays on the server (Decision, D-228, D-229). An exclusion and its new plan save together, or nothing changes (Decision, D-234). A request can take up to 4 calls of Luna, so the API streams each step, and the app shows the progress (Decision, D-231, D-237).

The home screen opens the plan screen. While a request runs, the screen shows the text of each step, with the try number and the cause of a retry (Decision, D-239). Each error of a request has a plain text that says that the plan did not change (Decision, D-240).

Luna writes one plan summary and one short reason for each exercise, with a length limit (Decision, D-182). A new exercise has no logged evidence, so its reason says that it is new (Decision, D-262). Session titles come from a template. The warm-up, the cool-down, and the mobility and recovery texts come from a versioned catalog, and Luna selects each item by id (Decision, D-152). A filter of blocked claims reads each text of Luna, and a template text replaces a blocked text (Decision, D-183).

### 3.4 Guided workout

The workout screen shows one machine at a time. The design targets one-handed use: large targets, few taps, little typing, and tolerance of interruptions (Decision, D-71). The visual style is calm, focused, high-contrast, and minimal (Decision, D-70). Cues are visual only, with no audio and no vibration (Decision, D-58).

1. The owner logs reps, weight, and reps in reserve for each set. Pain (0 to 10) and a note are optional (Decision, D-57, D-162).
2. The rest timer starts when the owner logs a set, with the rest of the target. The owner can adjust it by 15 seconds or dismiss it (Decision, D-59, D-270).
3. After the last set of an exercise, a preview of 10 seconds names the next machine. Then the app advances automatically, and "Go now" advances at once (Decision, D-60, D-269).
4. The owner can edit a prior set or skip an exercise (Decision, D-63).
5. A busy machine gives a skip, not a substitute (Decision, D-47).
6. "Finish now" skips the remaining exercises and records the session as ended early (Decision, D-63).
7. The first set of a new exercise is the calibration. The owner changes the weight during its first reps. The other sets get the logged weight of the first set, with no network (Decision, D-297, D-299).
8. When each exercise is done, the list of the exercises collapses to one line. So the cardio and the end of the workout show near the top (Decision, D-298).

A workout starts from the next session of the plan that the owner did not do yet, or from another session (Decision, D-248). The reps and the weight of a set come from the target, and a tap on the reps in reserve logs the set (Decision, D-249). The plus and minus buttons of the weight move to the next weight of the machine (Decision, D-264). Each workout screen has the button "Report a symptom". It shows seven symptoms, and a pick shows the warning (Decision, D-251, D-263). While a workout is open, the plan screen refuses a new plan and an exclusion (Decision, D-252).

The rest timer shows on the screen only, because the app sends no notifications (Decision, D-61). The timer reads a stored end time, so it is correct after a screen lock. A Screen Wake Lock keeps the screen on during a workout. The app requests the lock again at each return, focus, and tap, so a return needs no tap. When the phone refuses the lock, the screen shows a notice with the error name (Decision, D-265, D-283).

On the iPhone, a test of each method with no tap showed that the Wake Lock API alone keeps the screen on after a return. The owner picked it, and the app uses no video fallback (Decision, D-280, D-283, D-284).

The phone keeps each log and its outbox entry first, and sends the outbox to the workout service later (Decision, D-77, D-132). One call applies 100 entries or fewer, and each entry applies one time alone, keyed by its client op id (Decision, D-259). The server keeps each applied op id with no end date, so a replay changes nothing (Decision, D-257). For a workout entry, the phone wins (Decision, D-258).

The phone sends the outbox while the app is open alone, because iOS has no background sync for a web app (Decision, D-21). The sync runs at the open, at each focus and return, at each reconnect, and after each new entry. After a failure, it tries again after 5 s, 15 s, 60 s, and then each 5 minutes (Decision, D-277). One sync holds the workout entries and the inventory entries in the order of the op ids (Decision, D-275).

The phone moves an entry that the server refused to a separate list, and the owner dismisses it (Decision, D-274). A line below the header of each screen shows the state of the sync (Decision, D-276). The phone keeps copies of the catalog, the inventory, the plan, and the profile, so a workout starts with no connection (Decision, D-278).

The server stores each logged session as one Firestore document at `users/{uid}/workouts/{workoutId}`, with the link to its plan session (Decision, D-248, D-256). It checks each set with the bounds of D-164, and each cardio log with the fields of D-123. A cardio log holds the true duration, with no least time (Decision, D-260). A note has 280 characters or fewer (Decision, D-261). A refused entry changes nothing, and the other entries of the batch still apply.

### 3.5 Adaptation

After a session, the rules of the policy give the next targets (Decision, D-288). They read reps, load, reps in reserve, pain, skipped work, and gaps in the history (Decision, D-64). Luna writes the reason alone. The app shows a concise reason that names the logged evidence (Decision, D-68). A check refuses a reason of Luna that names no logged set, and the reason of the rules then shows (Decision, D-288).

The revision runs on the server when `SyncOutbox` applies a finished workout (Decision, D-292). It reads these inputs:

- The finished workouts of the owner, 100 at most. Each workout holds the target of each exercise that the owner saw at its start (Decision, D-291). A workout of an older phone has no such copy. For it, the revision reads the linked session of a plan with no revision.
- The confirmed machine of each exercise, as for a plan (Decision, D-49, D-193). Each exercise of a plan starts as a return after a long break, so the first sessions count from its start (Decision, D-238).

Each exercise that the workout logged gets the new target of the rules in each session of the plan that holds it (Decision, D-290). A skip counts as a log. The plan keeps its creation time, so each workout keeps its link to the plan. The plan records the last revision, and a replay of the sync does not revise the plan two times.

One reviser call writes the reason of each new target, with the schema `luna_reason_v1` and the prompt `luna-prompt-v6`. The output names the logged sets that each reason uses. The check refuses these reasons (Decision, D-68, D-288):

- a reason that the filter blocks,
- a reason that names no logged set, or a set that the last session did not log,
- a reason with a number that the evidence does not hold.

The reason of the rules then shows, and the plan records the cause. The call has a time limit of 45 s. After a failed call, a call over the cap, or the time limit, the reasons of the rules show, and the targets stay the same. A store failure of a revision fails the sync call. The phone keeps the entries and sends them again, and the revision runs again.

After the workout, the end screen shows the next target of each revised exercise with its reason. With no connection, it tells the owner that the next targets show after the sync. The sync then reads the plan copy again (Decision, D-278, D-292).

A revision gives each target on the date of the finished workout. So the phone sends its local date with each read of the plan. The server then gives each target of the rules on that date, the date of the next session. This read applies the long-break table, the rule of a missed session, and the start and the end of a deload (Decision, D-151, D-179, D-294, D-295). It changes only an exercise that the owner logged under the plan, and it saves nothing. A plan with no revision has no such exercise, so the read reads no store for it.

With no connection, the phone uses the plan copy of its last read.

The owner can override the load and the reps of each working set of a target for the next session, with a reason (Decision, D-69, D-293). The override keeps the count of sets, the reps in reserve, and the rest of the recommendation. Before the save, the policy checks that each set has 6 to 20 reps and a weight of the machine (Decision, D-23). The plan keeps the recommendation, the override, and the reason as separate records. No AI model reads the reason.

At the start of a workout, the phone uses the sets of the override. The target copy of the workout keeps the recommendation and the reason too. The next revision starts from the override, and it removes the override. The override expires when a missed session, a break, or a deload changes the rules of the date after its save. The recommendation then applies, and the owner can override it again.

The engine handles missed sessions and long breaks (Decision, D-66). After a gap of 7 to 13 days, the load and the reps stay the same at 3 reps in reserve for one session. Then the reps in reserve of the target before the gap come back (Decision, D-294).

The reactive deload has these rules (Decision, D-289, D-295, D-296):

- A decline is a session whose working sets have fewer total reps than the session before it. The logged load is the same or heavier.
- A decline in 2 sessions in a row on 2 or more exercises starts a deload on the 7 dates after that session. An exercise counts when its last decline is less than 14 days before the start (Decision, D-296).
- In the deload, each exercise has 0.6 times its sets, at least 1. The load stays, and each set stops at 3 reps in reserve. A load step of the rules waits for the end of the deload (Decision, D-303).
- A deload session is no evidence for the next target. So after the deload, the targets of before it return.

After a gap of 14 days or more, the load goes down by the long-break table (Decision, D-151, D-179). The first sessions back stop at 3 reps in reserve, with rep progression only. A new exercise starts at the estimate of the owner, or at the lightest weight (Decision, D-178, D-300).

The plan has no calibration set (Decision, D-297). The first session of each exercise uses its first set as the calibration (Decision, D-301). A break of 91 days or more gives one more such session. The owner changes the weight during the first reps. The other sets use the logged weight of the first set, and the rules read it as the load of the session (Decision, D-299).

From the second session, the normal rules apply. A new plan reads the logged history, so an exercise with history gets no calibration in it (Decision, D-301).

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
- The rest of each exercise: each plan gives 60 seconds, the leg press too, with policy version 5 (Decision, D-279).
- The rounding of a load to 5 lb, with the halfway rule and the weight of the machine (Decision, D-65, D-148, D-149).
- The next target: double progression with one 5 lb step, missed reps, pain, a lighter weight, and sets with no log (Decision, D-147, D-168 to D-170, D-173, D-174).
- The fixed warning text of a pain report (Decision, D-153, D-169).
- The check of a proposal. The policy refuses a proposal outside a bound (Decision, D-23). Outside the first session of an exercise, it also refuses a proposal that is harder than its target at the same load (Decision, D-186).
- The start of a new exercise with 3 working sets, and the first set as the calibration (Decision, D-178, D-180, D-297, D-299 to D-301).
- The return after a break of 14 days or more, and the first sessions after it (Decision, D-37, D-151, D-179).
- The hold after a missed session of 7 to 13 days, and the reactive deload, with policy version 6 (Decision, D-294, D-295, D-303).
- The check of an override of the owner (Decision, D-23, D-293).
- The rules fallback of one exercise. When the policy refuses a proposal, or Luna gives none for the exercise, the target comes from the rules alone (Decision, D-23). A plan request with no valid output of Luna gets no fallback plan (Decision, D-230).
- The decision record of each plan decision, with the fields of D-176. The record is workout data, so it never goes into a log (Decision, D-80, D-176).

Policy version 6 adds the reactive deload of REC-7 (Decision, D-289, D-295). Policy version 7 removes the calibration sets, and a deload week keeps the load of the last target (Decision, D-297, D-303). The draft in `tools/spikes/luna_plan/policy.py` is the spike record alone.

The role layer is in `go/internal/ai` (Decision, D-157). It holds these parts:

- The planner and the reviser roles on `gpt-6-luna` at xhigh effort (D-253). The role holds the model id, so a call site names a role alone (Decision, D-22, D-24).
- The strict plan schema. Its enums hold the ids of the catalog of D-155 and of the guidance catalog, so a valid output names no unknown id (Decision, D-152).
- The strict reason schema of the reviser. It holds a reason and the logged sets that it names for each exercise, and no target (Decision, D-288).
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

The section "Open questions" of `docs/questions.md` holds the open questions. On 2026-10-04, Q-103 alone is open. The high-level roadmap and the focused roadmaps name the phase or the pull request that needs each answer.
