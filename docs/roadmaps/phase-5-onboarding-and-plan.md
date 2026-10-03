# Workout App - Phase 5 focused roadmap: onboarding and plan generation

This roadmap splits Phase 5 of `docs/roadmaps/high-level-roadmap.md` into pull requests. `docs/roadmaps/README.md` gives the rules. The high-level roadmap keeps the objective, the order, and the exit evidence of the phase.

The date of this version is 2026-10-02.

## 1. Scope and start state

Phase 5 produces the first validated plan from the profile and the confirmed inventory. At the end of the phase, the owner completes onboarding and sees a plan that the policy accepted (D-23, D-34, D-41, D-42). The plan holds warm-up, work sets, rest, cooldown, optional cardio, and mobility and recovery text (D-44, D-73).

Phase 5 puts the live planner on the service `api`. Each live check of a planner call is a paid step, and the owner approves it with its expected cost (D-25, D-212). The tests use the fake provider, and no test calls OpenAI (D-24).

### 1.1 The exit evidence of Phase 4

The exit of Phase 4 holds. This session read these facts on 2026-10-02:

- PR-21 merged to `main` as `e010746`. It changed `web/` alone.
- The build `deploy-web` `5d3916f0` of `e010746` gave SUCCESS at 17:03:50Z. At 17:22:00Z, the live `/version.json` named `e010746`.
- At 17:22:00Z, the live `/version` named `b3484b6`, because no API deploy followed PR-21.
- At 17:22:00Z, `api-runtime` held `roles/datastore.user` and no other project role (D-206).
- The owner added one machine in the live app on the iPhone and confirmed it. The lists showed the machines A to Z. The owner gave this result in this session (D-203, D-204, D-205).

The emulator tests of PR-19 and the browser tests of PR-20 and PR-21 passed in CI before each merge.

### 1.2 Inputs from the earlier phases

- `go/internal/inventory` gives the confirmed machines alone to a plan, through `ForPlan` (D-49, D-193).
- `go/internal/ai` holds the planner role, the prompt `luna-prompt-v2`, the strict plan schema, the filter of blocked claims, and the fake provider. Its cap hook `MemoryCap` holds the spend in memory alone.
- `go/internal/policy` holds the policy, the rules fallback, and the decision record of D-176.
- The planner input holds no profile. The input JSON of `go/internal/ai/prompt.go` holds the date, the session count, the exercises, and the cardio exercises.
- The API of `go/cmd/api` does not use the role layer. The service `api` reads no OpenAI key and no cap value. `api-runtime` holds the accessor role on the secret `openai-api-key` (`docs/setup-gcp.md`).
- The web app has the shell and the inventory screens in `web/`. It never reads or writes Firestore itself (D-77).

## 2. Owner answers for this phase

| Question | Answer | Decision |
|---|---|---|
| Q-98, the monthly AI caps | 1 USD for the user and 2 USD for the project, for each month. | D-188 |
| Q-203, the lasting cap store | Phase 5, in work area 5.2, before the first live call of the planner. | D-189 |
| Q-204, the cap period | The calendar month in UTC. | D-190 |
| Q-221, the split | PR-23 to PR-27, as section 4 gives them. | D-207 |
| Q-222, the injured area | Areas from a fixed list, and a versioned table of the areas of each exercise. The server removes these exercises before the call to Luna. | D-208 |
| Q-223, the input of Luna | The experience, the goal template, the muscle groups, the free text, and the cardio preference. | D-209 |
| Q-224, the muscle groups | Ten fixed groups, a table of the primary groups of each exercise, and the templates "General fitness" and "Strength". | D-210 |
| Q-225, the sessions of a plan | The training days in each week, from 2 to 4, for one week. | D-211 |
| Q-226, the live check | A paid step. The owner approves each live check with its expected cost. | D-212 |

No open question blocks PR-23. Section 4 names the questions that each session asks. Q-194 and Q-202 stay open for Phase 7. Q-103 stays open for the deferred photo work.

## 3. Rules for each pull request of this phase

- The API alone reads and writes the profile and the plan in Firestore (D-77). The rules of `firestore.rules` refuse each client read and write.
- A plan reads the confirmed machines alone, through `ForPlan` (D-49, D-193).
- Every set and load passes the policy before the owner sees it (D-23). When Luna gives no proposal, or the policy refuses it, the rules fallback gives the target.
- No model id appears at a call site (D-24). The tests use the fake provider.
- A planner call sends the inputs of D-209 alone. The server removes each exercise of an injured area before the call (D-208).
- No log, metric, or error report holds the age, the height, the weight, the injury text, or the free text (D-80).
- The text of a plan is fitness guidance alone, in text alone (D-36, D-73). The filter of blocked claims reads each text of Luna (D-183).
- Before each live check of a planner call, the session states the expected cost and asks the owner (D-25, D-212).
- After a change in `proto/`, run `make proto`, and commit the generated code. `make contract` runs `buf breaking` against `main`.
- The tests use synthetic data alone. The repository is public.
- Each pull request with code gets the review of the other provider (D-15).
- A merge that changes `go/` deploys the API, and a merge that changes `web/` deploys the web app (D-137). The next session reads each deploy first.
- A new secret or a new configuration value of the service `api` goes into `cloudbuild/api.yaml` or the service, and `docs/setup-gcp.md` records it (recommendation).

## 4. Pull requests

The PR-<n> order is the order of work. Each pull request needs the one before it on `main`.

| Id | Work area | Title | Paid step |
|---|---|---|---|
| PR-22 | all | `docs: the Phase 5 focused roadmap (PR-22)` | none |
| PR-23 | 5.1 | `feat: the profile API and store (PR-23)` | none |
| PR-24 | 5.1 | `feat: the onboarding screens (PR-24)` | none |
| PR-25 | 5.2 | `feat: the lasting AI cap store (PR-25)` | none |
| PR-26 | 5.2 | `feat: the plan API (PR-26)` | none |
| PR-27 | 5.2 | `feat: the plan screens (PR-27)` | the live check, after the deploy (D-212) |

### PR-22 - The Phase 5 focused roadmap

Branch: `docs/pr-22-phase-5-roadmap`.

Concerns:

- this file, with the exit evidence of Phase 4 in section 1.1 (D-203, D-204),
- the owner answers of section 2 (D-207 to D-212), with their change of the Phase 5 work areas in `docs/roadmaps/high-level-roadmap.md` and of `docs/design.md`,
- the row of this file in `docs/roadmaps/README.md`, and the Phase 5 stage in `AGENTS.md`.

Acceptance story: the owner reads this roadmap. Each pull request of it names its work area, concerns, acceptance story, checks, and the questions it needs. Section 1.1 holds the exit evidence of Phase 4, and no open question blocks PR-23.

Checks: `make verify` and `make pr-check`, free. Codex reviews PR-22, because it sets the injury rule of D-208 and the data that goes to OpenAI (D-209).

### PR-23 - The profile API and store

Branch: `feat/pr-23-profile-api`. Work area 5.1. It needs PR-22 on `main`.

Concerns:

- a profile service in `proto/workoutapp/v1`. Its calls read and save the one profile of the owner,
- the Firestore store of the profile, with the inputs of D-41, D-42, and D-208 to D-211,
- the checks of the server: the bound of each field, and each area and each group against its fixed list,
- the versioned tables of D-208 and D-210 in `go/internal/domain`: the areas and the primary muscle groups of each exercise of the catalog. The research goes into `docs/research/exercise-safety.md` first (D-38),
- a Go function that gives the planner input of a profile. It holds the inputs of D-209 alone, and no exercise of an injured area (D-208).

Acceptance story: the emulator tests save and read a profile through the API, and the server refuses a field outside its bound. A unit test finds each exercise of the catalog in both tables. The function for the planner input holds the inputs of D-209 alone, and no exercise that loads an injured area.

Checks: `make contract`, `make go-test`, `make emulator-test`, and `make verify`, free. Codex reviews PR-23.

Questions for the session: the Firestore path of the profile, the bounds of each field, and the values of the experience field (D-30, D-33). Also the final list of areas, and the values of both tables. The owner answered them as Q-227 to Q-234 (D-213 to D-220). Codex finding P1-1 gave Q-235 (D-221).

Calls for PR-24: `GetProfileOptions` gives the experience values, the goal templates with their groups, the ten groups, and the seven areas. `GetProfile` gives the profile, or no profile. `SaveProfile` replaces the whole profile, and it refuses a field outside its bound with `invalid_argument`.

### PR-24 - The onboarding screens

Branch: `feat/pr-24-onboarding-screens`. Work area 5.1. It needs PR-23 on `main`.

Concerns:

- the onboarding screens for each input of PR-23, with large targets and little typing (D-71),
- the injury warning of D-35: after the owner selects an area, the screen warns that the plan avoids it. The text stays inside the fitness boundary (D-36),
- the selection of muscle groups, the two goal templates, and the free text (D-42, D-210),
- the training days in each week, from 2 to 4 (D-211),
- a direct call to the API for each save, and an error with no connection, as D-196 gives for the inventory (recommendation).

Acceptance story: the browser tests fill each input, show the injury warning, select a template, change its groups, and save the profile. A second visit shows the saved profile.

Checks: `make web`, `make verify`, and the Go checks of PR-23 when `go/` changes, free. Codex reviews PR-24.

Questions for the session: the groups of the template "Strength" (D-210), and the text of the injury warning. D-220 answered the groups. The owner answered the text of the warning and the way to onboarding as Q-236 and Q-237 (D-222, D-223).

### PR-25 - The lasting AI cap store

Branch: `feat/pr-25-cap-store`. Work area 5.2. It needs PR-24 on `main`.

Concerns:

- a cap hook that holds the spend of each period in Firestore, so a new instance of the API does not reset it (D-189),
- the calendar month in UTC (D-190), and the caps of D-188 for the user and for the project,
- the reservation of the worst-case cost before a call, and the charge after it, in one transaction, as `MemoryCap` does in memory,
- the comments of `go/internal/ai/cost.go` and the cap text of `go/README.md`.

Acceptance story: after a restart of the API, an emulator test refuses a call over the cap. The new store reads the spend of the month. The spend of a new month starts at 0. Two calls at the same time never pass the cap together.

Checks: `make go-test`, `make emulator-test`, and `make verify`, free. Codex reviews PR-25.

Questions for the session: the Firestore path of the spend, and the charge of a failed call with an unknown cost. The owner answered them as Q-238 and Q-239 (D-224, D-225).

### PR-26 - The plan API

Branch: `feat/pr-26-plan-api`. Work area 5.2. It needs PR-25 on `main`.

Concerns:

- a plan service in `proto/workoutapp/v1`. Its calls request a plan, read the current plan, and exclude an exercise with an optional reason (D-48),
- the planner call through the role layer, with the planner input of PR-23 and the confirmed machines. A new prompt version adds the inputs of D-209 and the session count of D-211,
- the policy check of each exercise, the rules fallback, and the decision record (D-23, D-176). A call over the cap gives the rules fallback,
- the plan content of D-44, D-152, and D-182. The warm-up, the cooldown, and the mobility and recovery items come by id, and the session titles from templates,
- the exclusion of D-48: Luna plans again, and the policy checks the result,
- the Firestore store of the plan,
- the service `api` reads the secret `openai-api-key`, the caps of D-188, and the cap store of PR-25. `cloudbuild/api.yaml` and `docs/setup-gcp.md` record the change.

Acceptance story: an end-to-end test with the fake provider returns a valid plan for the core profile of D-31. A failed call and a refused proposal give the rules fallback. No plan holds an exercise of an injured area, an excluded exercise, or a machine that the owner did not confirm.

Checks: `make contract`, `make go-test`, `make emulator-test`, and `make verify`, free. The session makes no live call. Codex reviews PR-26.

Questions for the session: the Firestore path of the plan, and the bound of an exclusion reason. Also the effect of a new plan on the old plan.

### PR-27 - The plan screens

Branch: `feat/pr-27-plan-screens`. Work area 5.2. It needs PR-26 on `main`.

Concerns:

- the plan screen. Each session shows the warm-up, the work sets, the rest, the cooldown, the optional cardio, and the mobility and recovery text (D-44, D-73),
- the request of a plan, with a wait state and an error state,
- the exclusion of an exercise, with an optional reason (D-48),
- the live check on the iPhone after the deploy of the merge (D-212).

Acceptance story: the browser tests show each part of a plan from the API, exclude an exercise, and show the new plan. After the `deploy-web` build of the merge, the owner approves the cost and requests one plan on the iPhone. The live app shows a plan that the policy accepted.

A deploy comes from `main` alone (D-14). So the browser tests are the evidence of the merge, and the owner does the live check after the deploy. The Phase 6 roadmap session records the result (recommendation, as D-203 did for Phase 4).

Checks: `make web`, `make verify`, and the Go checks of PR-26 when `go/` changes, free. Codex reviews PR-27.

Questions for the session: the text that the screen shows for a plan from the rules fallback.

## 5. Exit of the phase

Phase 5 ends when PR-27 merges and the live check passes. These items give the exit evidence of `docs/roadmaps/high-level-roadmap.md`:

- The browser tests of PR-24 cover each input and the injury warning.
- The end-to-end test of PR-26 with the fake provider returns a valid plan for the core profile of D-31.
- The cap test of PR-25 refuses a call over the cap after a restart of the API.
- After the deploy of PR-27, the owner approves the cost. The live app on the iPhone then shows a plan that the policy accepted (D-212).
