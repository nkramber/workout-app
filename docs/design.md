# Gym Route - design

This document holds the product thesis, the target experience, the system context, the safety boundaries, and the privacy posture of Gym Route. It is the design reference for `docs/roadmaps/high-level-roadmap.md`. The date of this version is 2026-09-28.

Each statement has a label. **Fact** means a verified fact with a source in `docs/research/`. **Decision** means an owner decision in `docs/decisions.md`. **Recommendation** means a proposal that the owner did not accept yet. **Assumption** means a belief that nobody verified yet. **Open** means a question in `docs/questions.md`.

## 1. Product thesis

Gym Route is a personal workout app for one person, the owner (Decision, D-67). The owner tells the app the muscles to train, the schedule, the experience, the injuries, and the goals. The owner photographs each machine of the gym. The app identifies each machine, and the owner confirms it (Decision, D-49). The app builds a workout plan, guides each workout, records each set, and adapts the next targets from the history.

The core of the app is **a thin LLM over a strict policy** (Decision, D-22, D-23). OpenAI `gpt-6-luna` at medium effort proposes each plan and each revision (Decision, D-24). A deterministic, versioned policy checks every set, load, and change before the owner sees it. When Luna fails or proposes a value that the policy refuses, a rules fallback gives a safe target. This thesis copies the Decktome thesis, where deterministic code checks every card that the model names.

Gym Route is an installable, phone-first web app on a default Firebase Hosting URL (Decision, D-17). It never goes to an app store, and no native app exists. It has one phone layout, and nobody designs or tests a desktop layout (Decision, D-20).

## 2. Target user and scope

| Item | Value | Label |
|---|---|---|
| User | The owner alone, ever | Decision, D-67 |
| Core profile | A 180 lb, 32-year-old man with prior machine experience, away from the gym for several years | Decision, D-31 |
| Experience level | Intermediate to advanced. Similar path for each, with more weight for advanced lifters. The core user starts as intermediate with his own load estimates. | Decision, D-30, D-32 |
| Out of scope | People under 18, people with no weight training | Decision, D-26, D-33 |
| Goals | General fitness and strength | Decision, D-31 |
| Equipment | Fixed-path resistance machines and cardio machines. No cable stations, free weights, bodyweight exercises, or bands. | Decision, D-45 |
| Locale | US English, pounds only | Decision, D-28 |
| Lead platform | Chrome on iPhone. It uses the WebKit engine. | Decision, D-29 |
| Business model | Free, no ads, no sale of data | Decision, D-27 |

## 3. Target experience

### 3.1 Onboarding

The owner enters experience, goals, injuries and restrictions, age, height, weight, and cardio preference (Decision, D-41). The owner picks muscles directly or starts from a goal template, and can add free text for Luna (Decision, D-42). The screen asks about injuries and experience only. It has no readiness questions (Decision, D-34). An injury answer gives a warning, and the plan avoids the injured area (Decision, D-35).

### 3.2 Equipment capture

The owner takes one general photo of a machine (Decision, D-50). The app asks for more photos only when the result is uncertain. Every screen also offers manual selection and text entry (Decision, D-51, D-55). The owner confirms or corrects each machine before a plan uses it (Decision, D-49). The owner then gives a current load estimate for the machine (Decision, D-41).

The app deletes each source photo after the confirmation, unless the owner keeps it (Decision, D-52). No user photo goes into an evaluation set (Decision, D-53).

The equipment record holds the identity and the available weights of each machine (Decision, D-54). The owner has one active inventory (Decision, D-46).

### 3.3 Plan

The plan adapts after each session and has no fixed block (Decision, D-43). It holds warm-up, resistance work, rest periods, cooldown, optional cardio, and mobility and recovery guidance (Decision, D-44). The guidance stays inside the fitness boundary (Decision, D-36). Instructions are text only (Decision, D-73). The owner can exclude an exercise with an optional reason, and Luna plans again under the policy (Decision, D-48).

### 3.4 Guided workout

The workout screen shows one machine at a time. The design targets one-handed use: large targets, few taps, little typing, and tolerance of interruptions (Decision, D-71). The visual style is calm, focused, high-contrast, and minimal (Decision, D-70). Cues are visual only, with no audio and no vibration (Decision, D-58).

1. The owner logs reps, weight, and reps in reserve for each set. Pain and a note are optional (Decision, D-57).
2. The rest timer starts when the owner logs a set. The owner can adjust or dismiss it (Decision, D-59).
3. After the last set of an exercise, a brief preview names the next machine. Then the app advances automatically (Decision, D-60).
4. The owner can edit a prior set or skip an exercise (Decision, D-63).
5. A busy machine gives a skip, not a substitute (Decision, D-47).
6. "Finish now" skips the remaining exercises and records the session as ended early (Decision, D-63).

The rest timer shows on the screen only, because the app sends no notifications (Decision, D-61). A Screen Wake Lock keeps the screen on during a workout (Recommendation, from `docs/research/platform-cloud-and-ai.md`).

### 3.5 Adaptation

After a session, Luna proposes the next targets from reps, load, reps in reserve, pain, skipped work, and gaps in the history (Decision, D-64). The policy checks the proposal. The app shows a concise reason that names the logged evidence (Decision, D-68). The owner can override a target, and the app keeps the recommendation, the override, and the reason as separate records (Decision, D-69). The engine handles missed sessions and long breaks (Decision, D-66).

Targets use one to three reps in reserve. Failure is rare, and it never occurs in the first sessions after a break (Decision, D-37). Loads round to the nearest 5 lb, up or down (Decision, D-65). A rounded jump can exceed a validated target, and Q-92 asks how the policy bounds it (Open).

## 4. System context

```text
 Phone (installable web app, Chrome or Safari on iPhone)
   UI, IndexedDB store and outbox, camera input, wake lock
        |  HTTPS, Connect-RPC unary calls, Firebase ID token
        v
 Cloud Run API (Go, default run.app URL, CORS to the app origin)
   auth interceptor -> handlers -> policy engine -> role layer -> OpenAI gpt-6-luna
        |                    |
        v                    v
 Firestore (server only)   Cloud Storage (short-lived photos)
 Cloud Scheduler -> Cloud Run jobs (photo purge, maintenance)
 Firebase Hosting (static app), Firebase Authentication (email and password)
```

| Part | Role | Sensitivity | Label |
|---|---|---|---|
| Web app | Screens, offline store, outbox, camera input | Holds workout history and profile data on the phone | Decision, D-17, D-62, D-77 |
| API | Auth check, sync, plan calls, policy | Reads and writes all user data | Decision, D-18, D-82 |
| Policy engine | Checks every prescription | Safety-critical | Decision, D-23 |
| Role layer | Model choice, cost records, fake provider | Sends profile and workout data to OpenAI | Decision, D-24 |
| Firestore | Source of record after sync | Injuries, body data, workout history | Decision, D-77 |
| Cloud Storage | Photos until confirmation | Photos can show other people | Decision, D-52 |
| Firebase Auth | Email and password, invite allowlist | Email address | Decision, D-75 |
| GCP project | One development project in `us-central1` | All of the above | Decision, D-76 |

The web client stack follows Decktome by default (Decision, D-74). After research on alternatives, the owner chose the Decktome React stack with Vite 8 (Decision, D-84). The native framework comparison of the launch prompt does not apply (Decision, D-17).

### 4.1 Limits of a web app on iOS

The owner accepted these limits (Decision, D-21). The facts come from `docs/research/platform-cloud-and-ai.md`.

| Limit | Effect on Gym Route | Label |
|---|---|---|
| No Vibration API in Safari on iOS | No haptic cues. D-58 chose visual cues only. | Fact |
| No Background Sync on iOS | The app syncs when it is open, visible, and online. | Fact |
| Web Push only for Home Screen apps | Not relevant. D-61 chose no notifications. | Fact |
| Home Screen app storage is separate from Safari storage | The owner signs in again inside the installed app. | Fact |
| The OS can evict web storage under storage pressure | The server copy stays the durable record. | Fact |
| No HealthKit, Health Connect, watch, or Live Activities | Q-49 is not applicable. | Fact |

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
9. The app strips photo metadata and deletes each photo after the confirmation (Decision, D-52, and Recommendation for metadata).
10. A logged set survives a lost connection and an app restart (Decision, D-62).

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

### 5.3 Medical boundary

The app states that it gives fitness guidance only (Decision, D-36). The FDA general wellness guidance of January 2026 treats claims about strength and muscle size as wellness claims (Fact). Claims about disease, treatment, or rehabilitation fall outside it (Fact). The policy and the prompts must keep Luna text inside that line. Q-95 asks what the OpenAI usage policies add (Open).

## 6. Privacy posture

Gym Route stores data about one person, the owner (Decision, D-67). The owner chose the minimum compliance posture (Decision, D-79) and no user data controls (Decision, D-78). The project hosts no public policy pages (Decision, D-81). These rules still hold:

- The repository is public. No personal data, email address, photo, or workout log goes into it.
- Telemetry holds ids only (Decision, D-80).
- The API sends OpenAI requests with the response store turned off (Recommendation, from `docs/research/platform-cloud-and-ai.md`).
- Photos have their metadata removed before upload, and the server deletes them after the confirmation (Decision, D-52).
- Secrets live in Secret Manager, never in the repository.

A change of audience reopens D-78, D-79, D-81, D-34, D-39, and D-40. The high-level roadmap names this reopening gate.

## 7. Open questions

Q-91 to Q-107 in `docs/questions.md` hold the open questions. Q-90 has an answer (D-84). The high-level roadmap names the phase that needs each answer.
