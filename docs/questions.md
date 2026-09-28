# Gym Route - questions

This file is the durable question register. It keeps every question, its answer, its status, and the decision that closes it. Do not delete an answered question. A later answer adds a row or a note, and the earlier answer stays.

The ids have three ranges:

- Q-1 to Q-73 are the questions of the launch prompt of the first session.
- Q-74 to Q-89 and Q-108 to Q-112 are follow-up questions of the first session.
- Q-113 to Q-118 are questions of the Phase 1 roadmap session.
- Q-119 is a question of the Luna plan spike session.
- Q-120 to Q-123 are questions of the recognition test set session.
- Q-90 to Q-107 are open questions. Q-90 and Q-94 to Q-97 have an answer. Each open question names the phase of `docs/roadmaps/high-level-roadmap.md` that needs the answer.

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

## Open questions

| # | Question | Why it matters | Ask when | Status |
|---|---|---|---|---|
| Q-90 | Which web client stack does the app use? | Q-52 asked for research first. Answer on 2026-09-28: the Decktome React stack (D-84). | Before this PR closes | Answered |
| Q-91 | How does the app handle equipment with kilogram markings under pounds only (D-28)? | A wrong conversion gives a wrong load. | Phase 2 | Open |
| Q-92 | When nearest-5-lb rounding (D-65) makes a load jump larger than the policy limit, does the policy hold the load or add reps instead? | On a 25 lb stack, one 5 lb step is a 20 percent jump. | Phase 3 | Open |
| Q-93 | When does the D-4 period end, and which focused roadmaps must exist first? | The `review-override` label stays off until then. | After the first focused roadmaps merge | Open |
| Q-94 | Does the owner apply the committed ruleset of `main` on GitHub after this PR merges? | The ruleset makes the review gate and the contract check mandatory. It changes live GitHub settings. Answer on 2026-09-28: yes. The session applied both files, and `make ruleset-check` passed (D-91). | Right after this PR merges | Answered |
| Q-95 | What do the current OpenAI usage policies say about fitness and health advice? | The research could not read the policy page. The prompts of Luna must obey it. Answer on 2026-09-28: use the dated copy of the policy page (D-93). | Phase 1 | Answered |
| Q-96 | What does one equipment photo cost on `gpt-6-luna`? | OpenAI does not publish the image token rate of this model. A paid probe needs owner approval (D-25). Answer on 2026-09-28: the recognition spike measures it, with a cap of 2 USD (D-94). | Phase 1 | Answered |
| Q-97 | Which image licenses are acceptable for the recognition test set (D-56)? | Test images must be lawful to copy and store. Answer on 2026-09-28: CC0, public domain, CC BY, CC BY-SA, and CC BY-NC (D-95). | Phase 1 | Answered |
| Q-98 | What are the monthly AI caps for the one user and for the project (D-25)? | A cap stops a runaway loop from a large bill. | Phase 4 | Open |
| Q-99 | Does the one development project get Firestore backups and point-in-time recovery? | D-76 leaves one project. It holds the only copy of the workout history. | Phase 4 | Open |
| Q-100 | Which fields does a cardio machine log hold? | D-57 defines the set log for resistance machines only. D-45 adds cardio machines. | Phase 2 | Open |
| Q-101 | Where does mobility and recovery guidance come from, as text only (D-73) and inside the fitness boundary (D-36)? | Luna can write this text, and the policy can not check prose for medical claims as easily as numbers. | Phase 3 | Open |
| Q-102 | How long a gap counts as a long break (D-66)? | The re-entry rule needs a threshold. | Phase 3 | Open |
| Q-103 | Where do the photos that a user chooses to keep (D-52) live, and for how long? | Retention needs a place and a limit. | Phase 2 | Open |
| Q-104 | Which way does D-65 round a value exactly halfway between two 5 lb steps, such as 22.5 lb? | "Nearest" does not decide a tie, and the policy needs one answer. | Phase 3 | Open |
| Q-105 | What happens when a rounded load does not exist on the machine? | D-54 stores the available weights, and D-65 rounds to a global 5 lb step. The two can disagree. | Phase 3 | Open |
| Q-106 | Do calibration sets for a new machine use three to four reps in reserve, when D-37 sets targets at one to three? | The research recommends more reserve for the first sets on an unknown machine. | Phase 3 | Open |
| Q-107 | Where does the D-40 warning end and emergency advice, which D-36 excludes, begin? | A warning for chest pain needs text that stays inside the fitness boundary. | Phase 3 | Open |
