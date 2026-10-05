# Workout App - questions

This file is the durable question register. It keeps every question, its answer, its status, and the decision that closes it. Do not delete an answered question. A later answer adds a row or a note, and the earlier answer stays.

The ids have three ranges:

- Q-1 to Q-73 are the questions of the launch prompt of the first session.
- Q-74 to Q-89 and Q-108 to Q-112 are follow-up questions of the first session.
- Q-113 to Q-118 are questions of the Phase 1 roadmap session.
- Q-119 is a question of the Luna plan spike session.
- Q-120 to Q-124 are questions of the recognition test set session.
- Q-125 to Q-130 are questions of the recognition spike session.
- Q-140 to Q-144 are questions of the Phase 2 roadmap session.
- Q-145 to Q-147 are questions of the API skeleton session.
- Q-152 to Q-160 are questions of the deploy session.
- Q-161 to Q-164 are questions of the Phase 2 check session.
- Q-165 to Q-171 are questions of the Phase 3 roadmap session.
- Q-172 to Q-179 are questions of the domain model session.
- Q-180 to Q-187 are questions of the policy session.
- Q-90 to Q-107 are open questions. Q-90 to Q-97, Q-99 to Q-102, and Q-104 to Q-107 have an answer. Each open question names the phase of `docs/roadmaps/high-level-roadmap.md` that needs the answer.

The owner answered through the answer controls of the session. The first session asked Q-1 to Q-46 and Q-74 to Q-88 on 2026-09-27. It asked the other questions on 2026-09-28. The column "Rec." tells whether the owner chose the option that the launch prompt marked as recommended. A "No" in that column is an owner decision, not a mistake.

## Status values

| Status | Meaning |
|---|---|
| Answered | The owner answered. The Decision column names the register row. |
| Superseded | A later answer replaced the answer. The Decision column names both rows. |
| Not applicable | A decision removed the question. The Answer column gives the reason. |
| Waiting on research | The owner asked for research before the answer. |
| Open | Nobody asked the question yet, or the owner did not answer yet. |

## A. Product and release

| # | Question | Answer | Rec. | Status | Decision |
|---|---|---|---|---|---|
| Q-1 | What is the app's working name? | "Permanent name - workshop some now. This is intended for a SMALL audience, mainly me. This won't be monetized, so the name can be trademarked + it does not have to be overly clever." Follow-up Q-77 chose "Gym Route". | No | Answered | D-16 |
| Q-2 | Who should the first usable release serve? | The owner and a small invite-only beta. | Yes | Superseded | D-19, D-67 |
| Q-3 | What distribution target should the high-level roadmap include? | "See answer to first question": the app never goes to an app store (Q-77). | No | Answered | D-17 |
| Q-4 | What does "hosted on GCP" mean? | "Follow whatever Decktome does". | No | Answered | D-18 |
| Q-5 | What age range is in scope? | Adults 18 and older only. | Yes | Answered | D-26 |
| Q-6 | What is the business model? | Free, privacy-respecting product with no ads or data monetization. | Yes | Answered | D-27 |
| Q-7 | What initial language, market, and units should the app support? | US English and pounds only. | No | Answered | D-28 |
| Q-8 | Which mobile platform leads? | "iPhone Chrome". Chrome on iOS uses WebKit. | No | Answered | D-29 |

## B. Safety and workout science

| # | Question | Answer | Rec. | Status | Decision |
|---|---|---|---|---|---|
| Q-9 | What training experience is initially in scope? | "Intermediate to advanced. Similar path for each, more weight for advanced". Follow-ups Q-81 and Q-82. | No | Answered | D-30 |
| Q-10 | Which goals should the first release support? | "Fitness + Strength. The CORE use case is for a 180lb 32 year old male who has some experience lifting weight via machines like a chest press, but has not gone to a gym in several years." | No | Answered | D-31 |
| Q-11 | Which screening inputs should onboarding collect? | Injuries and experience only. | No | Answered | D-34 |
| Q-12 | May screening stop automatic plan generation? | Warn the user but continue with conservative programming. | No | Answered | D-35 |
| Q-13 | What medical boundary should the product state? | Fitness guidance only. No diagnosis, treatment, rehabilitation prescription, or emergency advice. | Yes | Answered | D-36 |
| Q-14 | How should proximity to failure work? | Reps in reserve with rare failure. No failure in the first sessions back. | Yes | Answered | D-37 |
| Q-15 | What should make safety-critical workout decisions? | Luna proposes, and deterministic code validates every result. This is the hybrid option of the launch prompt. | No | Answered | D-23 |
| Q-16 | What evidence standard should apply? | Current authoritative guidance plus peer-reviewed systematic evidence, fully cited and dated. | Yes | Answered | D-38 |
| Q-17 | Is qualified human review required before broader release? | Not required. | No | Answered | D-39 |
| Q-18 | What should happen when a user reports a warning symptom? | Warn and let the user continue after confirmation. Follow-up Q-83 kept this answer for every symptom. | No | Answered | D-40 |

## C. Plan generation

| # | Question | Answer | Rec. | Status | Decision |
|---|---|---|---|---|---|
| Q-19 | Which onboarding inputs are required beyond muscles, days, and equipment? | "Experience, goals, injuries/restrictions, current load estimates (per machine, after uploading images of machines), age, height, weight, cardio preference." | No | Answered | D-41 |
| Q-20 | How should users express what they want to train? | Both direct muscle selection and goal-based templates. | Yes | Answered | D-42 |
| Q-21 | How long does a plan last? | Continuous adaptation with no explicit block. | No | Answered | D-43 |
| Q-22 | What supporting content belongs in the plan? | Add mobility and recovery prescriptions to warm-up, rest, cooldown, and optional cardio. | No | Answered | D-44 |
| Q-23 | Which exercise modalities are initially supported? | "Resistance machines + cardio machines". Follow-up Q-84 excluded cable stations. | No | Answered | D-45 |
| Q-24 | Should a user have multiple equipment profiles? | One active equipment inventory only. | No | Answered | D-46 |
| Q-25 | What happens when planned equipment is occupied or unavailable? | Let the user skip the exercise. | No | Answered | D-47 |
| Q-26 | May users exclude exercises or movements? | Yes, with optional reasons and safe replanning. | Yes | Answered | D-48 |

## D. Equipment photography and modeling

| # | Question | Answer | Rec. | Status | Decision |
|---|---|---|---|---|---|
| Q-27 | How does equipment recognition become trusted data? | Automatic identification followed by mandatory user confirmation or correction. | Yes | Answered | D-49 |
| Q-28 | Which photographs should the capture flow request? | One general photograph first, with follow-up requests when uncertain. | No | Answered | D-50 |
| Q-29 | Is a complete manual equipment path required? | Yes. | Yes | Answered | D-51 |
| Q-30 | What is the default image-retention policy? | Delete source images after confirmed extraction unless the user explicitly retains them. | Yes | Answered | D-52 |
| Q-31 | May user images improve a shared recognition system? | First answer: "Yes, no opt in". Follow-up Q-85 replaced it: delete photos, no evaluations. | No | Superseded | D-53 |
| Q-32 | Which machine details should the equipment model store? | The first answer repeated the Q-85 answer. The re-ask chose identity and available weights only. | No | Answered | D-54 |
| Q-33 | What should happen below the recognition confidence threshold? | "ALWAYS allow manual selection/text entry OR a photo upload. If photo fails, prompt for a new one OR manual selection/text entry". | No | Answered | D-55 |

## E. Workout flow and adaptation

| # | Question | Answer | Rec. | Status | Decision |
|---|---|---|---|---|---|
| Q-34 | What is the minimum set log? | "Reps, weight, RIR, optional pain, optional note". | No | Answered | D-57 |
| Q-35 | When does the rest timer start? | Automatically when a set is logged, with easy adjustment or dismissal. | Yes | Answered | D-59 |
| Q-36 | How should the confirmed automatic move to the next exercise appear? | A brief completion and next-machine preview, then an automatic advance. | Yes | Answered | D-60 |
| Q-37 | Which hands-free cues are needed? | Visual only. | No | Answered | D-58 |
| Q-38 | Must workout logging work offline? | Yes, with durable local state and later synchronization. | Yes | Answered | D-62 |
| Q-39 | Which in-workout corrections must be supported? | Edit and skip only. Follow-up Q-87 added "finish now". | No | Answered | D-63 |
| Q-40 | Which signals may affect progression? | Reps, load, effort or reps in reserve, pain, skipped work, and history quality. | Yes | Answered | D-64 |
| Q-41 | Must recommendations respect actual machine increments? | Round to a global unit increment. Follow-up Q-88 chose the nearest 5 lb. | No | Answered | D-65 |
| Q-42 | Which recovery disruptions should the engine support? | Missed sessions and long breaks only. | No | Answered | D-66 |
| Q-43 | Should adaptation be explainable? | Yes. A concise reason tied to prior logged evidence. | Yes | Answered | D-68 |
| Q-44 | May users override prescriptions? | Yes. Keep the recommendation, the override, and the reason separately. | Yes | Answered | D-69 |

## F. Mobile design

| # | Question | Answer | Rec. | Status | Decision |
|---|---|---|---|---|---|
| Q-45 | What visual direction should lead? | Calm, focused, high-contrast, and minimal. | Yes | Answered | D-70 |
| Q-46 | Is one-handed gym use a formal requirement? | Yes. | Yes | Answered | D-71 |
| Q-47 | Which accessibility baseline applies? | Defer accessibility until after the core flow. | No | Answered | D-72 |
| Q-48 | What exercise-instruction media is permitted? | Text only initially. | No | Answered | D-73 |
| Q-49 | Which ecosystem integrations enter the roadmap now? | Not applicable. A web app has no access to HealthKit, Health Connect, watch apps, or Live Activities (D-17). The session did not ask it. | n/a | Not applicable | D-17 |
| Q-50 | Which notifications should exist? | No notifications. | No | Answered | D-61 |

## G. Architecture, identity, privacy, and operation

| # | Question | Answer | Rec. | Status | Decision |
|---|---|---|---|---|---|
| Q-51 | How strongly should Decktome's technical stack bind this project? | The Decktome stack is the default, and each deviation needs evidence. | Yes | Answered | D-74 |
| Q-52 | How should the mobile framework be selected? | The native comparison no longer applies (D-17). The owner chose "Research alternatives" for the web client stack. After the research, Q-90 chose the Decktome React stack. | n/a | Answered | D-84 |
| Q-53 | What is the initial authentication approach? | Firebase email and password with an invite allowlist, as in Decktome. | Yes | Answered | D-75 |
| Q-54 | How should GCP environments and region work? | "Just dev, us-central1". | No | Answered | D-76 |
| Q-55 | Which AI-provider policy applies? | `gpt-6-luna` through a role layer with a fake provider. | Yes | Answered | D-24 |
| Q-56 | How are paid AI calls controlled? | Owner approval for development runs plus production caps for each user and for the project. | Yes | Answered | D-25 |
| Q-57 | Which data authority model should lead? | Offline-first local workout state synchronized to a cloud source of record. | Yes | Answered | D-77 |
| Q-58 | Which user data controls are required? | "None, this project is just for me". Follow-up Q-89 confirmed a single user. | No | Answered | D-78 |
| Q-59 | What compliance posture should the roadmap assume? | Minimum app-store and general consumer requirements. No app store applies (D-17). | No | Answered | D-79 |
| Q-60 | What telemetry is allowed? | Operational metrics and error reports with no sensitive workout text, images, prompts, or health details. | Yes | Answered | D-80 |

## H. Public web surface

| # | Question | Answer | Rec. | Status | Decision |
|---|---|---|---|---|---|
| Q-61 | May the project host required policy and support pages on a default provider URL? | None. No store or other user needs them. | n/a | Answered | D-81 |
| Q-62 | How should the API be addressed? | The default Cloud Run URL, with no custom domain. | Yes | Answered | D-82 |

## I. Repository workflow and first-session boundaries

| # | Question | Answer | Rec. | Status | Decision |
|---|---|---|---|---|---|
| Q-63 | What may this first session implement? | Add minimal non-product validation tooling. The owner statement with Q-64 extended the scope. | No | Answered | D-1, D-3 |
| Q-64 | Which foundation artifacts should it create? | The complete expected artifact set. The owner added a statement about Codex review, the AGENTS.md and CLAUDE.md link, ASD-STE100, Gitar, and the `review-override` label. | Yes | Answered | D-5, D-3, D-4 |
| Q-65 | How should the two-or-three-concern PR rule work? | The concerns form one cohesive milestone with one combined acceptance story. | Yes | Answered | D-10 |
| Q-66 | Does one clean session per PR remain? | "One clean session per PR, but PRs can still contain 2/3/more concerns depending on scope. Owner approval before beginning work". | No | Answered | D-12 |
| Q-67 | What cross-provider review rule applies? | Cross-provider review only for code or safety-affecting changes. | No | Answered | D-15 |
| Q-68 | Who authorizes merges? | The owner explicitly authorizes every merge after current checks and review. | Yes | Answered | D-13 |
| Q-69 | Which Decktome Git and delivery rules carry over? | The standard set of the launch prompt. | Yes | Answered | D-14 |
| Q-70 | Which documentation and gate style applies? | The owner statement with Q-64 requires ASD-STE100, and Q-74 chose the gates. The session did not ask Q-70 separately. | n/a | Answered | D-83 |
| Q-71 | What Git operations may the first session perform? | Create a branch, commit, push, and open the documentation PR. Do not merge. | Yes | Answered | D-2 |
| Q-72 | How should the roadmap represent future PRs? | Broad PR-sized work areas without permanent PR numbers. | Yes | Answered | D-9 |
| Q-73 | Where should the standalone launch prompt live after this session starts? | Outside the repository. | Yes | Answered | D-11 |

## Follow-up questions of the first session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-74 | Besides Codex review, the AGENTS.md and CLAUDE.md mirror, and STE, which Decktome process items does the first PR port? | Core gates, the ruleset of `main`, the session hooks, and the skills. | Answered | D-6 |
| Q-75 | What does the CLAUDE.md and AGENTS.md mirror mean? | "One points to other": `AGENTS.md` holds the rules, and `CLAUDE.md` points to it. | Answered | D-7 |
| Q-76 | Can the session run `make codex-review` on its own PR? | "Yes. Codex reviews must be run AUTOMATICALLY by you after CI is green." | Answered | D-8 |
| Q-77 | Which permanent name? | "This will NEVER be on an app store, it will live as an installable web app, the same as Decktome. Gym Route works." | Answered | D-16, D-17 |
| Q-78 | Confirm the product form: native apps or an installable web app? | Installable web app. | Answered | D-17 |
| Q-79 | What does a desktop browser show? | The phone layout only. | Answered | D-20 |
| Q-80 | How does the roadmap treat the iOS web limits? | "Accept, design around." The owner added the statement that D-22 records: Luna is key to planning and revision. | Answered | D-21, D-22 |
| Q-81 | How does the engine treat the core user? | Intermediate from the first day, with the user's load estimates. | Answered | D-32 |
| Q-82 | Are people with no weight training in scope? | Out of scope. | Answered | D-33 |
| Q-83 | Keep "warn, allow continue" for every symptom, or split by symptom type? | Warn and continue for every symptom. | Answered | D-40 |
| Q-84 | Do cable stations count as resistance machines? | No. Fixed-path machines only. | Answered | D-45 |
| Q-85 | How does photo reuse fit with deletion after confirmation? | Delete the photos. No evaluations use user photos. | Answered | D-53 |
| Q-86 | With no evaluations from user photos, where do test photos come from? | Public or licensed images. | Answered | D-56 |
| Q-87 | How does a user stop a workout partway? | "Finish now" skips the remaining exercises, and the session records as ended early. | Answered | D-63 |
| Q-88 | Which global rounding rule applies to prescribed loads? | The nearest 5 lb, up or down. | Answered | D-65 |
| Q-89 | Which audience statement holds: Q-2 (invite beta) or Q-58 (just for me)? | "Only you, ever". | Answered | D-67 |
| Q-108 | The three role-model repositories name pull requests in different forms. Which form do pull requests after #1 use? | The form of the-thing-below: `<type>: <summary> (PR-<n>)` and `<type>/pr-<n>-<slug>`. | Answered | D-86 |
| Q-110 | Codex finding P1-2: the author loop sends a pull request that Codex writes to a Codex review. Can the Codex author session run a Claude Code review automatically? | Yes, automatically, as D-8 does for Codex. | Answered | D-88 |
| Q-111 | Codex finding P1-3: the provider gate trusts the `Author provider` line of the author. How is it resolved? | Extend the accepted risk of D-87. The merge question names the author provider. | Answered | D-89 |
| Q-112 | Codex finding P1-5: Dependabot pull requests skip the review gate, and the repository has no Dependabot. What happens to the exemption? | Remove it. | Answered | D-90 |
| Q-109 | Codex finding P1-1: the review gate can not prove that Codex wrote the review record. How is it resolved? | The rule what-you-carry:D-198: an accepted risk, with the commit of the record in the gate output. | Answered | D-87 |

## Questions of the Phase 1 roadmap session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-113 | Section 8 of the high-level roadmap needs an owner confirmation of the phase order. Does the owner confirm the order Phase 0 to Phase 8? | Yes, as written. | Answered | D-92 |
| Q-114 | D-86 assigns PR-<n> ids "from PR-1". Does each phase start again at PR-1? | No. One sequence runs across all phases. | Answered | D-96 |
| Q-115 | The repository is public. Where do the images of the recognition test set live? | A manifest in Git, and the images in a local cache outside Git. | Answered | D-97 |
| Q-116 | Which cap does the paid run of the Luna plan spike get? | 2 USD. | Answered | D-98 |
| Q-117 | The iPhone web platform spike needs a trusted HTTPS origin and an auth project. How does it serve its probe? | Create the development project early, with Firebase Hosting and Firebase Authentication only. | Answered | D-99 |
| Q-118 | Which language do the Luna spike harnesses use? | Python. | Answered | D-100 |

## Questions of the Luna plan spike session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-119 | No decision defines the go bar of the Luna plan risk. Which bar does the report of the spike apply? | A schema pass rate of 95% or more, a policy rejection rate of 25% or less, and a mean cost per plan of 0.01 USD or less. | Answered | D-101 |

## Questions of the recognition test set session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-120 | Where does the recognition test set look for images? | Wikimedia Commons first. Add another source only when a machine type has too few images, and record the license proof of each image. | Answered | D-102 |
| Q-121 | Licensed photos of gym machines are few, and many name no gym. The owner asked why the set needs so many images. Which size does the set get? | Keep about 200 licensed images. The images are one-photo trials that measure the rate of confident wrong answers. When a photo names no gym, the photographer and the upload event give the gym key. Accept a smaller real count, and record it. | Answered | D-103 |
| Q-122 | Few face-free photos exist for many strength machines. Can the set use a photo that shows the face of a person? | Yes, when no other photo or few other photos exist for that machine type. | Answered | D-104 |
| Q-123 | Four catalog types have no image on Commons, and six types have images from one source only. How does the set fill these gaps? | Add a source outside Commons for these types, with the license proof of each image. | Answered | D-105 |
| Q-124 | The Codex review of #4 found that the author names of the manifest (D-97) conflict with the privacy rule of `AGENTS.md`. Which rule controls? | D-97 controls for the author credit. The license requires the credit, and the source publishes it. The privacy rule still covers each other personal datum. | Answered | D-106 |

## Questions of the recognition spike session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-125 | No decision defines the go bar of the recognition risk. D-101 is the bar of the Luna plan risk only. Which bar does the report of the spike apply? | Safety first. A stated confidence of 0.8 or more is high. Go when the high-confidence wrong answers are 2% or less of all photos, the correct answers are 70% or more, and the mean cost per photo is 0.01 USD or less. | Answered | D-107 |
| Q-126 | Which photos go into the paid run? | All 174: the 130 originals and the 44 degraded photos, with the counts of each group in the report. | Answered | D-108 |
| Q-127 | The OpenAI API key is not in the environment of the session. How does the harness get it? | The owner first chose a new key file in `.local` of the worktree. Then the owner named the `.env` file of the main checkout, which holds the key and which Git ignores. The session loads that file into the command of the harness only. | Answered | D-109 |
| Q-128 | The recognition spike gave a no-go: 8.6% of the photos had a wrong answer with a high stated confidence, and the bar is 2%. The roadmap needs an owner decision for a no-go. How does it change? | Manual entry first. Phase 4 gives manual selection and text entry only, and photo recognition moves out of Phase 4. | Answered | D-110 |
| Q-129 | Which phase holds photo recognition after D-110? | No phase. The roadmap records it as deferred, and a later owner decision adds a phase. | Answered | D-111 |
| Q-130 | The iPhone probe has a camera page. Does it keep the page after D-110? | No. The probe tests only the device items that the manual app needs. | Answered | D-112 |

## Questions of the iPhone probe session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-131 | The probe needs Node 22 and the Playwright browsers. How do its checks connect to `make` and CI? | A separate target `make probe` builds the probe and runs the browser tests. A new CI job `verify:probe` runs it. `make verify` stays Python only. A Python test in the probe folder checks the configuration. | Answered | D-113 |
| Q-132 | Is the CI job `verify:probe` a required check of the ruleset of `main`? | No. The job runs on each pull request and each push to `main`, and the ruleset does not change. | Answered | D-114 |
| Q-133 | How do the browser tests check sign-in with no network access to the real project? | The tests use the real Firebase Auth SDK against the local Auth emulator, with a pinned `firebase-tools`. | Answered | D-115 |
| Q-134 | Which project id and which Google account does the development project of D-76 use? | The id `gym-route-dev`, and the account that the Firebase CLI of the owner machine uses. The account stays out of the repository. | Answered | D-116 |
| Q-135 | GitHub secret scanning flags the Firebase browser key of `gym-route-dev` in the public repository. How does the project handle it? | Keep the key in the repository, because each browser that loads the app gets it. Limit the key to the Auth APIs and to the sites of the probe, and turn off self sign-up. The owner closes the alert. | Answered | D-117 |

## Questions of the iPhone report session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-136 | Which bar gives a go for the iPhone web platform risk? | Go when items 1 to 5 of the device checklist pass in the Home Screen app, and the median first contentful paint of 5 cold starts in standalone mode is 2500 ms or less. Else no-go. | Answered | D-118 |
| Q-137 | Where does the owner run the items of the device checklist? | In the Home Screen app only. Item 3 starts in Chrome, because Chrome adds the app to the Home Screen. | Answered | D-119 |
| Q-138 | The probe showed two display faults on the iPhone: the layout sat too high, and a pinch zoomed the page. What do they mean for the app of the roadmap? | The owner stated on 2026-09-29: the probe can keep them, but the app of the roadmap must fix both. It blocks the pinch zoom as Decktome does. | Answered | D-120 |
| Q-139 | Codex finding P2-1 of pull request 7: the page can not prove that iOS ended the web view between the five launches. Does the bar of D-118 need a new definition of a cold start? | Yes. A cold start is a new page load of the Home Screen app after an app stop in the app switcher. The report keeps the limit that iOS can keep the web view in memory. | Answered | D-121 |

## Questions of the Phase 2 roadmap session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-140 | The exit evidence of work area 2.1 puts the contract, Go, and emulator tests in `make verify`. D-113 keeps `make verify` Python only. Which rule wins? | D-113. The product checks get their own `make` targets and CI jobs. | Answered | D-126 |
| Q-141 | The owner ended the D-4 period. Which pull request changes `OVERRIDE_ENABLED` and the rule text? | PR-7. | Answered | D-125 |
| Q-142 | Work area 2.3 needs a billing account for Cloud Run and for the backups of D-124. D-99 kept the project with no billing account. Which billing account and which budget alert apply? | The session of PR-10 read the billing state first: one open account. The owner chose that account and an alerts-only budget of 10 USD each month. | Answered | D-139 |
| Q-143 | Are the CI jobs of the product code required checks of `main`? | Yes. Each pull request that adds a job adds it to the ruleset. | Answered | D-127 |
| Q-144 | How does Phase 2 split into pull requests? | PR-8 to PR-11, as the focused roadmap gives them. | Answered | D-128 |

## Questions of the API skeleton session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-145 | How does the Go API guard against an emulator variable on Cloud Run (REC-17)? | It refuses to start on Cloud Run when any `_EMULATOR_HOST` variable exists. It has no debug user. | Answered | D-129 |
| Q-146 | Which Go and buf versions does the product code pin? | Go 1.27.1 and buf v1.73.0, through `go/go.mod`. | Answered | D-130 |
| Q-147 | D-75 names an allowlist as in Decktome, which keys the list by email. The PR-8 story names a uid. Which key applies? | The uid. | Answered | D-131 |

## Questions of the web shell session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-148 | Which local store and which outbox form does the web client use (REC-1)? | Dexie on IndexedDB, and an outbox with UUIDv7 op ids. The change and its outbox entry go into one transaction. | Answered | D-132 |
| Q-149 | How does the service worker apply an update (REC-3)? | The `prompt` mode. The owner applies the update, and never during a workout. | Answered | D-133 |
| Q-150 | Does the app ask for persistent storage after the first sign-in (REC-5)? | Yes, one time on each device, and the home screen shows the result. | Answered | D-134 |
| Q-151 | Does `web/` use the pnpm workspace of Decktome? | No. One npm package, with the generated code in `web/src/gen`. | Answered | D-135 |

## Questions of the deploy session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-152 | Why is the project `gym-route-dev`, and what is the name of the app? | The app is "Workout App", and the identifiers use `workout-app`. | Answered | D-136 |
| Q-153 | Google Cloud gave "already in use" for `workout-app-prod`. Which project id applies? | `nk-workout-app-prod`. | Answered | D-137 |
| Q-154 | What happens to the old project `gym-route-dev`? | It stays until the new project serves the app and the owner signs in. Then the owner shuts it down. | Answered | D-137 |
| Q-155 | Do the past records change to the new name? | No. Each current file changes, and the past records keep the old name. | Answered | D-136 |
| Q-156 | Does the contract package change to `workoutapp.v1`? And does `buf.yaml` get the package rules that see a moved file? | Yes to the package. The package rules come in a later pull request, after this change is on `main`. | Answered | D-138 |
| Q-157 | Which Firestore edition and mode apply (REC-12)? | Standard, Native mode, `us-central1`. | Answered | D-140 |
| Q-158 | Which Cloud Run cost settings apply (REC-15)? | Request billing, min 0, max 2, and the CPU boost. The owner skipped the spend cap. | Answered | D-141 |
| Q-159 | How does a change of `firestore.rules` deploy? | A third trigger with its own service account. | Answered | D-142 |
| Q-160 | The guard before each deploy leaves a gap before the deploy (P2-1, round 2). How do the builds keep their order? | A lock for each part in a small Cloud Storage bucket. | Answered | D-143 |

## Questions of the Phase 2 check session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-161 | PR-11 adds the package rules to `buf.yaml`. Does the type stay `docs`? | Yes. | Answered | D-144 |
| Q-162 | The roadmap gives PR-11 both the review of the other provider and the `review-override` label. Which one applies? | The review of the other provider, with no label. | Answered | D-144 |
| Q-163 | How does the app handle the iOS blur below the status bar? | The header adds 16 px above the title, and no more. | Answered | D-145 |
| Q-164 | Can `rules-deployer` get `roles/serviceusage.serviceUsageViewer`, and can the failed rules build run again? | Yes, at run time. | Answered | D-146 |

## Questions of the Phase 3 roadmap session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-165 | Which machine types does the first product catalog hold? | 12 selectorized machines, 2 cable exercises, dumbbells, and 6 cardio machines. | Answered | D-155 |
| Q-166 | D-45 excludes free weights. Do dumbbells come into scope? | Yes. D-45 changes. | Answered | D-154 |
| Q-167 | Is the triceps pulldown a cable exercise or a fixed-path machine? | A cable exercise, and the lat pulldown too. | Answered | D-154 |
| Q-168 | Which dumbbell exercises does the catalog hold? | 9 exercises with an adjustable bench. | Answered | D-155 |
| Q-169 | Where does the safety research for free weights go? | Into PR-12. | Answered | D-156 |
| Q-170 | Where do the domain types live in Phase 3? | Go types alone. Proto messages come in Phase 4. | Answered | D-157 |
| Q-171 | How does Phase 3 split, and does it end with a check pull request? | PR-12 to PR-17, with a paid evaluation in PR-17. | Answered | D-158 |

## Questions of the domain model session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-172 | How does the catalog hold one machine with two movements? | One machine and two exercises. | Answered | D-159 |
| Q-173 | Which form holds a load such as 12.5 lb exactly? | Integer tenths of a pound. | Answered | D-160 |
| Q-174 | Which regions does the catalog use? | The four regions of EV-1, core, and cardio. | Answered | D-161 |
| Q-175 | Which form does the optional pain value have? | An optional rating from 0 to 10. | Answered | D-162 |
| Q-176 | How does the model hold the adjustable bench? | As part of the dumbbell set. | Answered | D-163 |
| Q-177 | Which bounds does the check of a set log apply? | Reps and reps in reserve of 0 or more, and a weight above 0. | Answered | D-164 |
| Q-178 | Which unit holds the distance of a cardio log? | Tenths of a mile. | Answered | D-165 |
| Q-179 | Codex finding P2-2 of PR 14: which upper bound does the dumbbell set have? | At most 100 lb for each dumbbell. | Answered | D-166 |
| Q-180 | Which rep ranges does the policy use? | 8 to 12 on a machine, REC-19 for dumbbells and cables, and limits of 6 to 20. | Answered | D-167 |
| Q-181 | How does the policy handle missed reps and progression? | REC-6 and REC-15, with the thresholds of section 5.8. | Answered | D-168 |
| Q-182 | What is a pain report, and how long does it hold progression? | A rating of 1 or more, and a hold of one session. | Answered | D-169 |
| Q-183 | How does the policy read an exercise with sets that have no log? | It reads the logged sets. Progression needs a log of each planned set. | Answered | D-170 |
| Q-184 | Which free-weight rules of section 5.14 does the policy adopt? | REC-17 alone. | Answered | D-171 |
| Q-185 | Which rest does the policy permit? | 60 to 180 seconds, with a default of 120 seconds. | Answered | D-172 |
| Q-186 | What does the policy do when a logged weight is not the target load? | A lighter weight gives no progression. | Answered | D-173 |
| Q-187 | Which reps go with the lower load of a shortfall in 2 sessions in a row? | The same rep target. | Answered | D-174 |

## Questions of the policy fallback session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-188 | Does the policy of work area 3.2 adopt the reactive deload triggers of REC-7? | No. Phase 7 reads REC-7 again. | Answered | D-175 |
| Q-189 | Which fields does the record of each plan decision hold (REC-12)? | REC-12, with the proposal, the violations, the source, the cause, the target, and an input hash. | Answered | D-176 |
| Q-190 | How does the calibration of D-150 apply to each new exercise (REC-22)? | REC-22 for each kind, with 3 calibration sessions. | Answered | D-177 |
| Q-191 | Which fallback target does an exercise with no history and no estimate get? | A calibration from the lightest weight. | Answered | D-178 |
| Q-192 | Which values does the long-break table of D-151 use? | 10 and 20 percent less load with one set fewer, and 70 percent after 91 days. | Answered | D-179 |
| Q-193 | How many working sets does the start of a new exercise from the rules alone have? | 3. | Answered | D-180 |
| Q-195 | Codex round 3 of PR 16: does the full contract of a calibration set in a proposal go in PR-15? | Yes. One set in a calibration session alone, at the reps and the load of the first working set. | Answered | D-181 |

## Questions of the Luna role layer session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-196 | Which shown text does Luna write, and which text comes from templates? | Luna writes one plan summary and one short reason for each exercise. Templates and the guidance catalog give the other texts. | Answered | D-182 |
| Q-197 | Does the role layer apply the blocked-claims filter of REC-11? | Yes. A versioned filter reads each text of Luna, and a template text replaces a blocked text. | Answered | D-183 |

## Questions of the Luna evaluation session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-198 | How many profiles and calls does the paid run of the Phase 3 Luna evaluation have? | 20 planner calls and 5 reviser calls for each of the 6 scenarios. | Answered | D-184 |
| Q-199 | Which cap does the paid run get? | 2 USD. | Answered | D-185 |
| Q-200 | The policy accepts more reps or fewer reps in reserve at the load of its target. Does it refuse such a proposal? | Yes. A new rule goes in before the paid run. | Answered | D-186 |
| Q-201 | How does the OpenAI key reach Secret Manager? | The owner adds the version in a local terminal. | Answered | D-187 |

## Questions of the Phase 4 roadmap session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-203 | Which phase builds the lasting store of the monthly AI caps? The cap hook of PR-16 holds the spend in memory, so a new instance of the API resets it. | Phase 5, before the first live call of the planner. | Answered | D-189 |
| Q-204 | Which period does a monthly AI cap use? | The calendar month in UTC. | Answered | D-190 |
| Q-205 | What happens to a machine that the owner enters as text? | A search of the catalog. A text with no match stays as a note that no plan uses. | Answered | D-191 |
| Q-206 | Does the load estimate belong to a machine or to an exercise? | To an exercise, and it is optional. | Answered | D-192 |
| Q-207 | What is the confirmation of a machine that the owner enters? | A confirmation on a review screen after a draft. A change of the weights makes a draft again. | Answered | D-193 |
| Q-208 | How does Phase 4 split into pull requests? | PR-19 for the API and the store, then PR-20 for the screens. | Answered | D-194 |
| Q-209 | How does the owner enter the weights of a stack? | A range and a step, then a change of single weights. | Answered | D-195 |
| Q-210 | How does a change of the inventory reach the server in Phase 4? | A direct call to the API. The outbox comes in Phase 6. | Answered | D-196 |

## Questions of the inventory API session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-211 | Which Firestore path holds the inventory? | `users/{uid}/inventory/active`, one document. | Answered | D-197 |
| Q-212 | Which load estimate does the server accept? | An estimate in the range of the weights of the machine. | Answered | D-198 |
| Q-213 | Which bounds does the server apply to the inventory? | 1,000 lb and 200 weights for a list, and 200 characters and 50 notes for the notes. | Answered | D-199 |
| Q-214 | What happens when the owner saves a machine that the inventory holds? | The save replaces the entry. A change of the estimates alone keeps the confirmation. | Answered | D-200 |
| Q-215 | What does the confirmation of a machine send? | The weights that the review screen showed. The server refuses a confirmation of other weights. | Answered | D-201 |

## Questions of the inventory screens session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-216 | Which order does the catalog list of the inventory screen use? | By kind: the machines, the cable station, the dumbbells, then the cardio machines, each group in catalog order. | Answered | D-202 |
| Q-217 | When does the owner add and confirm one machine in the live app on the iPhone? | After the merge of PR-20 and its web deploy. The Phase 5 roadmap session records the result as the exit evidence of Phase 4. | Answered | D-203 |

## Questions of the inventory writes session

The live check of D-203 failed on 2026-10-02. The app showed "The API did not answer. The change is not saved." `SaveMachine` gave HTTP 500, because `api-runtime` held `roles/datastore.viewer` alone.

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-218 | The live check failed. What comes before the Phase 5 roadmap? | A Phase 4 correction in this session as PR-21. The Phase 5 roadmap becomes PR-22. | Answered | D-204 |
| Q-219 | Which lists show the machines A to Z? | Both: the inventory list, and each kind group of the catalog list. | Answered | D-205 |
| Q-220 | Which role lets the API write Firestore? | `roles/datastore.user` in place of `roles/datastore.viewer`, applied in this session. A server fault shows "The server failed." | Answered | D-206 |

## Questions of the Phase 5 roadmap session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-221 | How does Phase 5 split into pull requests? | Five: the profile API, the onboarding screens, the lasting cap store, the plan API, and the plan screens. | Answered | D-207 |
| Q-222 | How does an injury answer make the plan avoid the injured area? The catalog has six regions alone, and no table of the joints of each exercise. | Areas from a fixed list, and a versioned table of the areas of each exercise. The server removes these exercises before the call. | Answered | D-208 |
| Q-223 | Which profile inputs does a planner call send to Luna? | The training inputs alone. The age, the height, the weight, and the injury text stay on the server. | Answered | D-209 |
| Q-224 | Which muscle groups and goal templates does onboarding give? | Ten fixed groups, a table of the primary groups of each exercise, and the templates "General fitness" and "Strength". | Answered | D-210 |
| Q-225 | How many sessions does a plan hold? D-41 has no training frequency. | Onboarding asks the training days in each week, from 2 to 4. | Answered | D-211 |
| Q-226 | Is the live check of a planner call in the app a paid development run? | Yes. The owner approves each live check with its expected cost. | Answered | D-212 |

## Questions of the profile API session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-227 | Which Firestore path holds the one profile? | `users/{uid}/profile/active`. | Answered | D-213 |
| Q-228 | Which values does the experience field have? | Intermediate and advanced. | Answered | D-214 |
| Q-229 | Which bounds does the server apply? | Age 18 to 90, height 48 to 96 in, weight 80 to 500 lb, and texts of 500 characters or fewer. | Answered | D-215 |
| Q-230 | Which fixed list of injury areas does onboarding give? | The seven areas of D-208. | Answered | D-216 |
| Q-231 | Which form does the cardio preference take? | A list of the cardio exercises that the owner likes. | Answered | D-217 |
| Q-232 | Which areas does each exercise load? | The research table, with the close calls left out. | Answered | D-218 |
| Q-233 | Which primary groups does each exercise have? The adductors are not one of the ten groups. | The research table. Hip adduction and the cardio exercises get no group. | Answered | D-219 |
| Q-234 | Which groups does the template "Strength" select? | The six prime-mover groups. | Answered | D-220 |
| Q-235 | Codex finding P1-1 of PR 24: one id covers the standing and the seated calf raise, and the seated pad loads the knee. Does a knee injury remove it? | Yes. The row adds the knee. | Answered | D-221 |


## Questions of the onboarding screens session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-236 | Which text does the injury warning show? | The plan avoids the areas, and the app gives fitness guidance only, with no diagnosis or treatment. | Answered | D-222 |
| Q-237 | How does the owner get to onboarding? | The app opens it before the home screen while no profile exists. | Answered | D-223 |

## Questions of the cap store session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-238 | Which Firestore path holds the monthly AI spend? | `users/{uid}/aiSpend/{YYYY-MM}` for the user and `aiSpend/{YYYY-MM}` for the project, in one transaction. | Answered | D-224 |
| Q-239 | What does a failed AI call with an unknown cost charge? | The reserved worst-case cost. | Answered | D-225 |

## Questions of the plan API session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-240 | Which Firestore paths hold the plan and the exclusions? | `users/{uid}/plan/active` and `users/{uid}/exclusions/active`. | Answered | D-226 |
| Q-241 | What does a new plan do to the old plan? | It replaces the old plan, and no history stays. | Answered | D-227 |
| Q-242 | Which bound applies to the reason of an exclusion? | 200 characters or fewer. | Answered | D-228 |
| Q-243 | Does Luna get the reason of an exclusion? | No. The reason stays on the server. | Answered | D-229 |
| Q-244 | Which sessions does the rules fallback build when Luna gives no valid plan? | No fallback plan. The server retries, then gives an error. The owner confirmed this change of D-23. | Answered | D-230 |
| Q-245 | How many calls does a plan request make? | The first call and 3 retries at 90 s each, a request timeout of 420 s, and progress for the owner. | Answered | D-231 |
| Q-246 | Which bound applies to the optional cardio of a plan? | 5 to 30 minutes. Luna must stay inside it. | Answered | D-232 |
| Q-247 | Which bound applies to the size of one session? | 8 resistance exercises or fewer. More gives a new plan from Luna. | Answered | D-233 |
| Q-248 | What happens to an exclusion when each attempt fails? | All or nothing. Each retry sends the failed output and its cause, and an error record keeps each failure. | Answered | D-234, D-235, D-236 |
| Q-249 | How does the API give the progress to the app? | A Connect server stream. | Answered | D-237 |
| Q-250 | What does an error record hold, and for how long? | The ids, the cause, the cost, and the output of Luna, for 90 days. | Answered | D-236 |
| Q-251 | Does the plan API treat a new exercise as a return after a long break? | Yes, always: 70 percent of each estimate. | Answered | D-238 |

## Questions of the plan screens session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-252 | Which text does the plan screen show for each step of the progress? | Texts that name Luna, the try number, and the cause of a retry. | Answered | D-239 |
| Q-253 | Which text does the plan screen show for each error of a plan request? | Plain and specific texts, each with "Your plan did not change." | Answered | D-240 |
| Q-254 | How do the browser tests get a full plan and a visible progress from the local API? | A fuller reply of the fake, a local delay switch, and a route stub for two errors. | Answered | D-241 |
| Q-255 | Does the error of no allowed exercise say that the plan did not change? | Yes. The owner added "Your plan did not change." after the review of PR-27. | Answered | D-240 |

## Questions of the Phase 6 roadmap session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-256 | Where does the comparison of the medium and xhigh efforts go? | Into PR-28, with a paid run at each effort. | Answered | D-242 |
| Q-257 | Where does a change of the Luna role go? | Into PR-28. | Answered | D-243 |
| Q-258 | Does PR-28 also hold the fixes of the live check? | Yes, all in PR-28, over a split under D-10. | Answered | D-244 |
| Q-259 | When does the app make the weight list? | At the save, with no button. | Answered | D-245 |
| Q-260 | How does the owner confirm a cardio machine? | The save confirms it. | Answered | D-246 |
| Q-261 | How does Phase 6 split into pull requests? | Four, with the API first: PR-29 to PR-32. | Answered | D-247 |
| Q-262 | Which session of the plan does a workout start? | The next one that the owner did not do, and the owner can pick another. | Answered | D-248 |
| Q-263 | How does the owner log a set in three taps or fewer? | Reps and weight from the target, and one tap on the reps in reserve. | Answered | D-249 |
| Q-264 | Do the changes of the inventory go through the outbox? | Yes, in PR-32. | Answered | D-250 |
| Q-265 | How does the owner report a warning symptom during a workout? | A button "Report a symptom" on each workout screen. | Answered | D-251 |
| Q-266 | What happens to a plan request during a workout? | The plan screen refuses it until the workout ends. | Answered | D-252 |
| Q-267 | Which reasoning effort does Luna use? | xhigh, for both roles. | Answered | D-253 |
| Q-268 | Where does the change of the cardio rule go? | Into PR-29, with a wider milestone. | Answered | D-254 |
| Q-269 | What does "20 to 30 minutes of cardio" mean for a plan? | Each session has 20 to 30 minutes when the owner likes a cardio exercise. | Answered | D-255 |
| Q-270 | Which Firestore paths hold the logged sessions and the applied op ids? | One document for each session, and one for each op id. | Answered | D-256 |
| Q-271 | How long does the server keep an applied op id? | With no end date. | Answered | D-257 |
| Q-272 | Which rule applies to a conflict with `baseVersion`? | The phone wins for a workout entry. | Answered | D-258 |
| Q-273 | What is the size limit of one sync batch? | 100 entries. | Answered | D-259 |
| Q-274 | Does the 20-minute least time apply to a cardio log? | No, the log holds the true duration. | Answered | D-260 |
| Q-275 | What is the length limit of a note? | 280 characters. | Answered | D-261 |

## Questions of the workout screen session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-276 | Each exercise of the live plan said "The exercise log has a gap". How does the app fix the reason of a new exercise? | Prompt v5, in PR-30. | Answered | D-262 |
| Q-277 | Which symptoms does "Report a symptom" show, and which text does each warning have (D-153)? | Seven symptoms, each with the warning "You reported <symptom>. Stop this exercise." | Answered | D-263 |
| Q-278 | Which step do the plus and minus buttons of the weight use (D-249)? | The next weight of the list of the machine. | Answered | D-264 |
| Q-279 | How does the app keep the screen on during a workout? | The Screen Wake Lock API, with a notice when the phone refuses it. | Answered | D-265 |

## Questions of the rest timer session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-280 | The calibration set does not change the load of the working sets on the workout screen. Where does the app apply the step of the policy? | In PR-31. | Answered | D-266 |
| Q-281 | After a calibration set that is not on target, does the owner do one more calibration set (D-150)? | No. One calibration set, and the table gives the load of the working sets one time. | Answered | D-267 |
| Q-282 | The set log offers 0 to 4+ reps in reserve, and the table of D-150 needs 5 and 6+. Which buttons does a calibration set show? | 0 to 6+ on a calibration set alone. | Answered | D-268 |
| Q-283 | How long does the preview of the next machine show before the automatic advance (D-60)? | 10 seconds, with "Go now". | Answered | D-269 |
| Q-284 | Which controls does the rest timer have (D-59)? | "-15 s", "+15 s", and "Dismiss". | Answered | D-270 |
| Q-285 | The screen showed "The screen can turn off." after the owner left the app and came back. How does the app get the lock again? | At each tap, focus, `pageshow` event, and return, with the error name in the notice. | Answered | D-271 |

## Questions of the outbox sync session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-286 | `SyncOutbox` reads workout, set, and cardio entries alone, and a replay of `SaveNote` with an empty id adds a second note. How do the inventory entries of the outbox reach the server (D-250)? | Through `SyncOutbox`, with each entry applied one time by its op id. | Answered | D-272 |
| Q-287 | How does the confirmation of a machine work with no connection (D-201)? | The outbox keeps it with the weights of the review screen. | Answered | D-273 |
| Q-288 | Where on the phone does an entry that the server refused go? | Into a separate local list, with a count and a detail view. | Answered | D-274 |
| Q-289 | In one sync, what is the order of the inventory entries and the workout entries? | One order, the order of the op ids. | Answered | D-275 |
| Q-290 | Where does the state of the sync show on the screen? | In a line of the shell on each screen. | Answered | D-276 |
| Q-291 | After a failed sync while the app is open, when does the phone try again? | At each open, focus, and reconnect, and on a timer. | Answered | D-277 |
| Q-292 | The start of a workout reads the plan from the server. With no connection at the open of the app, no workout can start. Does PR-32 add an offline copy of the plan? | Yes. | Answered | D-278 |
| Q-293 | The owner asked for 60 seconds of rest between sets. Which exercises does it cover (D-172)? | Every exercise, the leg press too. | Answered | D-279 |
| Q-294 | On the iPhone, the screen lock needs a tap after a return to the app (D-271). What comes next? | A probe of each method with no tap, then the owner picks one. | Answered | D-280 |
| Q-295 | Where do the change of the rest and the probe of the screen lock go? | Into PR-32. | Answered | D-281 |
| Q-296 | The probe needs an HTTPS page, and a deploy comes from `main` alone (D-14). How does the probe run? | A "Screen lock test" screen in PR-32. After the deploy, the owner runs it, and the next pull request applies the method. | Answered | D-282 |

## Questions of the screen lock session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-297 | Which method of the "Screen lock test" does the workout use (D-282)? | "Wake Lock, no tap". | Answered | D-283 |
| Q-298 | What happens to the "Screen lock test" screen? | Remove it, and put the method into the workout. | Answered | D-284 |
| Q-299 | The iPhone check needs a deploy from `main` (D-14). Which evidence proves the milestone? | The browser tests. The owner checks the iPhone after the deploy. | Answered | D-285 |

## Questions of the Phase 7 roadmap session

| # | Question | Answer | Status | Decision |
|---|---|---|---|---|
| Q-300 | The high-level roadmap has two work areas in Phase 7. Does Phase 7 get a third work area? | Yes. Work area 7.3, "Reviser evaluation". | Answered | D-286 |
| Q-301 | How does Phase 7 split into pull requests? | Three: PR-35 to PR-37. | Answered | D-287 |

## Open questions

| # | Question | Why it matters | Ask when | Status |
|---|---|---|---|---|
| Q-90 | Which web client stack does the app use? | Q-52 asked for research first. Answer on 2026-09-28: the Decktome React stack (D-84). | Before this PR closes | Answered |
| Q-91 | How does the app handle equipment with kilogram markings under pounds only (D-28)? | A wrong conversion gives a wrong load. Answer on 2026-09-29: such a machine is out of scope, and the app makes no conversion (D-122). | Phase 2 | Answered |
| Q-92 | When nearest-5-lb rounding (D-65) makes a load jump larger than the policy limit, does the policy hold the load or add reps instead? | On a 25 lb stack, one 5 lb step is a 20 percent jump. Answer on 2026-09-30: no percent limit. At the top of the rep range, the policy adds one 5 lb step and resets the reps (D-147). | Phase 3 | Answered |
| Q-93 | When does the D-4 period end, and which focused roadmaps must exist first? | The `review-override` label stays off until then. Answer on 2026-09-29: the period ends now, and PR-7 turns on the label (D-125). | After the first focused roadmaps merge | Answered |
| Q-94 | Does the owner apply the committed ruleset of `main` on GitHub after this PR merges? | The ruleset makes the review gate and the contract check mandatory. It changes live GitHub settings. Answer on 2026-09-28: yes. The session applied both files, and `make ruleset-check` passed (D-91). | Right after this PR merges | Answered |
| Q-95 | What do the current OpenAI usage policies say about fitness and health advice? | The research could not read the policy page. The prompts of Luna must obey it. Answer on 2026-09-28: use the dated copy of the policy page (D-93). | Phase 1 | Answered |
| Q-96 | What does one equipment photo cost on `gpt-6-luna`? | OpenAI does not publish the image token rate of this model. A paid probe needs owner approval (D-25). Answer on 2026-09-28: the recognition spike measures it, with a cap of 2 USD (D-94). Measured on 2026-09-28: 0.00037 USD per photo, with 2,815 input tokens and 368 output tokens on average (`docs/research/recognition-spike.md`). | Phase 1 | Answered |
| Q-97 | Which image licenses are acceptable for the recognition test set (D-56)? | Test images must be lawful to copy and store. Answer on 2026-09-28: CC0, public domain, CC BY, CC BY-SA, and CC BY-NC (D-95). | Phase 1 | Answered |
| Q-98 | What are the monthly AI caps for the one user and for the project (D-25)? | A cap stops a runaway loop from a large bill. Answer on 2026-10-02: 1 USD for the user and 2 USD for the project, for each calendar month in UTC (D-188, D-190). | Phase 4 | Answered |
| Q-99 | Does the one development project get Firestore backups and point-in-time recovery? | D-76 leaves one project. It holds the only copy of the workout history. Answer on 2026-09-29: yes, point-in-time recovery and a daily backup from work area 2.3 (D-124). | Phase 4 | Answered |
| Q-100 | Which fields does a cardio machine log hold? | D-57 defines the set log for resistance machines only. D-45 adds cardio machines. Answer on 2026-09-29: duration and an effort rating, with optional distance, level, pain, and note (D-123). | Phase 2 | Answered |
| Q-101 | Where does mobility and recovery guidance come from, as text only (D-73) and inside the fitness boundary (D-36)? | Luna can write this text, and the policy can not check prose for medical claims as easily as numbers. Answer on 2026-09-30: a versioned catalog of texts in the repository, which Luna selects by id (D-152). | Phase 3 | Answered |
| Q-102 | How long a gap counts as a long break (D-66)? | The re-entry rule needs a threshold. Answer on 2026-09-30: a gap of 2 weeks or more, with the long-break table and the first 3 sessions or 14 days at 3 reps in reserve (D-151). | Phase 3 | Answered |
| Q-103 | Where do the photos that a user chooses to keep (D-52) live, and for how long? | Retention needs a place and a limit. | The deferred photo work (D-111) | Open |
| Q-104 | Which way does D-65 round a value exactly halfway between two 5 lb steps, such as 22.5 lb? | "Nearest" does not decide a tie, and the policy needs one answer. Answer on 2026-09-30: down on an increase and on a return after a break, up in other cases (D-148). | Phase 3 | Answered |
| Q-105 | What happens when a rounded load does not exist on the machine? | D-54 stores the available weights, and D-65 rounds to a global 5 lb step. The two can disagree. Answer on 2026-09-30: the heaviest available weight at or below the rounded load, or the lightest weight (D-149). | Phase 3 | Answered |
| Q-106 | Do calibration sets for a new machine use three to four reps in reserve, when D-37 sets targets at one to three? | The research recommends more reserve for the first sets on an unknown machine. Answer on 2026-09-30: 3 to 4 reps in reserve, with the calibration table of REC-5 (D-150). | Phase 3 | Answered |
| Q-107 | Where does the D-40 warning end and emergency advice, which D-36 excludes, begin? | A warning for chest pain needs text that stays inside the fitness boundary. Answer on 2026-09-30: the warning names the symptom and tells the user to stop the exercise, with no referral text (D-153). | Phase 3 | Answered |
| Q-194 | Does the policy adopt the reactive deload triggers of REC-7 (D-175)? | A decline on 2 or more exercises in 2 or more sessions can need a deload of the full day. The rules of D-168 read one exercise at a time. Answer on 2026-10-04: yes, in work area 7.2, with a new policy version (D-289). | Phase 7 | Answered |
| Q-202 | Does Luna keep a role in the targets? | In the Phase 3 evaluation, Luna copied the target of the rules in each of 197 decisions (`docs/research/phase-3-check.md`). Answer on 2026-10-04: for a revision, Luna writes the reason alone, and the rules give each target (D-288). | Phase 7 | Answered |
| Q-302 | After a finished workout, which targets does a revision change: the next session of the plan alone, or each exercise of the plan that the new log changes? | The plan holds one week of sessions, and the phone repeats them (D-211). An exercise can be in more than one session. Answer on 2026-10-04: each exercise that the workout logged, in each session of the plan that holds it (D-290). | PR-35 | Answered |
| Q-303 | Where does the target that the owner saw stay for the history of an exercise? | The rules read the target that the owner saw. A workout log holds only a link to the plan, and a new plan replaces the old plan with no history (D-227). Answer on 2026-10-04: the phone copies the target of each exercise into the workout log, and the sync sends it (D-291). | PR-35 | Answered |
| Q-304 | When does a revision run, and what shows when the phone has no connection at that time? | The policy runs on the server, and the phone logs a workout with no connection (D-77). Answer on 2026-10-04: in the sync of the finished workout, with a time limit for the reviser call. With no connection, the offline plan keeps the old targets until a sync completes (D-292). | PR-35 | Answered |
| Q-305 | Which fields of a target can an override change, and does the next revision start from the override or from the recommendation? | D-69 keeps the recommendation, the override, and the reason as separate records. The rules need one target for the next step. Answer on 2026-10-04: the load and the reps of each working set, and the next revision starts from the override (D-293). | PR-36 | Answered |
| Q-306 | What does a missed session change when the gap is shorter than 14 days? | D-66 names missed sessions. Today the rules change nothing for a gap under 14 days (D-179). Answer on 2026-10-04: a gap of 7 to 13 days holds the load and the reps at 3 reps in reserve for one session (D-294). | PR-36 | Answered |
| Q-307 | Which numbers does the reactive deload use (D-289)? | REC-7 of `docs/research/exercise-safety.md` gives 30 to 50 percent fewer sets and 3 reps in reserve for one week. The policy needs one value for each. Answer on 2026-10-04: a decline in 2 sessions in a row on 2 or more exercises, then 7 days at 0.6 times the sets, the same load, and 3 reps in reserve (D-295). | PR-36 | Answered |
| Q-308 | Which size and which cap does the paid reviser evaluation use? | The Phase 3 evaluation used 50 calls with a cap of 2 USD (D-184, D-185). A paid run needs the approval of the owner (D-25). Answer on 2026-10-04: 50 reviser calls with a cap of 1 USD (D-302). | PR-37 | Answered |
| Q-309 | How recent must a decline be to count toward a deload? | D-295 gives no age. With no limit, an old decline of one exercise and a new decline of another exercise start a deload together. Answer on 2026-10-04: less than 14 days before the start (D-296). | PR-36 | Answered |
| Q-310 | How does the first set of an exercise act as the calibration (D-297)? | The policy needs one rule: which load the other sets of the first session get from the weight of the first set, and what replaces the start as a return of D-238 in a new plan. Answer on 2026-10-04: the other sets use the weight of the first set, the first set is a working set, a new exercise starts at the estimate, the first session of each exercise calibrates, and the normal rules start in the second session (D-299 to D-301). | PR-37 | Answered |
| Q-311 | Which load does a deload week use when the rules give a load step? | The deload of D-295 applies to the next target of the rules. With the normal rules from the second session (D-301), that target can have a load step. Answer on 2026-10-04: the load of the last target, with no load step (D-303). | PR-37 | Answered |
| Q-312 | How does PR-37 handle two syncs of one finished workout that each call the reviser? | The live check of PR-37 found the defect of PR-35. A second sync of a finished workout reached the reviser while the first sync waited for its call. Answer on 2026-10-04: fix it in PR-37 with a claim of the workout before the call (D-304). | PR-37 | Answered |
| Q-313 | In a session with no calibration, which load do the later working sets get when the owner logs the first set at another weight? | The live check of `4bbf6c8` found it. On an exercise with history, the owner logged 30 lb at 3 reps in reserve for the first set, and the next set showed the target of 20 lb (D-301). Answer on 2026-10-05: the weight of the first set, inside a limit of the policy (D-306). | PR-38 | Answered |
| Q-314 | How far above the target can the later working sets follow a heavier first set? | A limit keeps each load inside the check of the policy (D-23). Answer on 2026-10-05: one weight of the list (D-307). | PR-38 | Answered |
| Q-315 | Do the later working sets follow a lighter first set? | A lower load is safe (D-186). Answer on 2026-10-05: yes, with no limit (D-308). | PR-38 | Answered |
| Q-316 | After a session in which the later working sets followed the first set, which load do the rules read? | The next target starts from that load. Answer on 2026-10-05: the followed load (D-309). | PR-38 | Answered |
| Q-317 | How does Phase 8 split into pull requests? | Each pull request holds one milestone (D-10). Answer on 2026-10-05: PR-39 to PR-41 for work areas 8.1 to 8.3, and PR-42 for the report of the four weeks (D-311). | PR-38 | Answered |
| Q-318 | Which spend guard does work area 8.2 add? | The budget sends email alone, and the owner skipped the spend cap of Cloud Run (D-139, D-141). Answer on 2026-10-05: no new cap, and alerts (D-312). | PR-38 | Answered |
| Q-319 | Where do the alerts of work area 8.2 go? | The app sends no notifications (D-61). Answer on 2026-10-05: a Cloud Monitoring email channel (D-313). | PR-38 | Answered |
