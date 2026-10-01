# Workout App - Phase 3 focused roadmap: workout domain and safety policy

This roadmap splits Phase 3 of `docs/roadmaps/high-level-roadmap.md` into pull requests. `docs/roadmaps/README.md` gives the rules. The high-level roadmap keeps the objective, the order, and the exit evidence of the phase.

The date of this version is 2026-09-30.

## 1. Scope and start state

Phase 3 builds the part that makes every prescription safe: the domain model, the deterministic policy, the rules fallback, and the Luna role layer (D-23, D-24). At the end of the phase, a Go package takes a Luna proposal and returns accepted targets, corrected targets, or a refusal with a reason. Phase 3 adds no screen and no API call.

The exit of Phase 2 holds. `docs/research/phase-2-check.md` holds its evidence. The Go module is in `go/`, and the contract is in `proto/`.

Phase 1 gives these inputs:

- `docs/research/exercise-safety.md` holds the evidence register, the synthesis, and the recommendations REC-1 to REC-22.
- `docs/research/luna-plan-spike.md` holds the plan schema and the go result of Luna (D-101).
- `tools/spikes/luna_plan/` holds the spike harness: the schema, the prompt, a Python policy, and the fake provider. It is a reference, not product code.

The owner changed the scope in the session of PR-12 (D-154). Dumbbells and two exercises on a cable station come into scope. So PR-12 adds the free-weight research to `docs/research/exercise-safety.md` (D-156).

## 2. Owner answers for this phase

| Question | Answer | Decision |
|---|---|---|
| Q-92, a large 5 lb step | No percent limit. At the top of the rep range, add one 5 lb step and reset the reps. At most one step for each exercise in each session. | D-147 |
| Q-104, a halfway value | Down on an increase and on a return after a break. Up in other cases. | D-148 |
| Q-105, a load that the machine does not have | The heaviest available weight at or below the rounded load, or the lightest weight. | D-149 |
| Q-106, calibration sets | Stop at 3 to 4 reps in reserve, with the calibration table of REC-5. | D-150 |
| Q-102, a long break | A gap of 2 weeks or more. The long-break table of REC-8. The first 3 sessions or 14 days, the longer of the two, at 3 reps in reserve (REC-4). | D-151 |
| Q-101, mobility and recovery text | A versioned catalog of texts in the repository. Luna selects items by id. | D-152 |
| Q-107, the warning boundary | The warning names the symptom and tells the user to stop the exercise. No referral text and no emergency advice. | D-153 |
| Q-166 and Q-167, the scope | Dumbbells, the cable lat pulldown, and the cable triceps pulldown come into scope. | D-154 |
| Q-165 and Q-168, the first catalog | The machines, cable exercises, dumbbell exercises, and cardio machines of D-155. | D-155 |
| Q-169, the free-weight research | In PR-12. | D-156 |
| Q-170, the place of the types | Go types alone in Phase 3. | D-157 |
| Q-171, the split | PR-12 to PR-17, in section 4. | D-158 |
| Q-172 to Q-179, the domain model | Two exercises for one machine, loads in tenths of a pound, six regions, pain from 0 to 10, the bench in the dumbbell set, set log bounds, distance in tenths of a mile, and a 100 lb dumbbell at most. | D-159 to D-166 |

Q-95 and Q-100 have an answer from earlier phases (D-93, D-123). No question of the phase stays open. Section 4 names each recommendation that a later pull request of the phase asks the owner about.

## 3. Rules for each pull request of this phase

- The domain model goes in "go/internal/domain", the policy in "go/internal/policy", and the role layer in "go/internal/ai" (D-157). Phase 3 adds no proto message.
- Each rule of the policy has a rule id, and the policy has one version. Each rule cites the `EV-<n>` ids or the decision that supports it (D-38).
- The same history and the same policy version give the same targets. The policy reads no clock and no random value.
- A number with no direct source stays a recommendation or an assumption. It becomes a rule only when the owner answers it in the session of the pull request.
- No model id appears at a call site. The role layer holds it (D-24).
- Tests make no network call. The fake provider stands in for Luna (D-24).
- Logs, metrics, and error reports hold ids only (D-80).
- A paid run needs the approval of the owner at run time (D-25). Each step that costs money has a **Paid** mark.
- Each pull request with code gets the review of the other provider (D-15).
- `make go-test` and the CI job `verify:go` run the Go tests. `make verify` stays Python only (D-126).

## 4. Pull requests

The PR-<n> order is the order of work. Each pull request needs the one before it on `main`.

| Id | Work area | Title | Paid step |
|---|---|---|---|
| PR-12 | all | `docs: the Phase 3 focused roadmap (PR-12)` | none |
| PR-13 | 3.1 | `feat: the workout domain model and catalog (PR-13)` | none |
| PR-14 | 3.2 | `feat: the policy bounds and progression (PR-14)` | none |
| PR-15 | 3.2 | `feat: the re-entry, calibration, and rules fallback (PR-15)` | none |
| PR-16 | 3.3 | `feat: the Luna role layer (PR-16)` | none |
| PR-17 | 3.2, 3.3 | `test: the Phase 3 Luna evaluation (PR-17)` | yes, Luna calls |

### PR-12 - The Phase 3 focused roadmap

Branch: `docs/pr-12-phase-3-roadmap`.

Concerns:

- this file,
- the owner answers of section 2 (D-147 to D-158),
- the change of scope of D-154 in `docs/design.md` and `docs/roadmaps/high-level-roadmap.md`, with the step rule of D-147,
- the free-weight research in `docs/research/exercise-safety.md` (D-156),
- the row of this file in `docs/roadmaps/README.md`, and the Phase 3 stage in `AGENTS.md`.

Acceptance story: the owner reads this roadmap. Each pull request of it names its work area, concerns, acceptance story, checks, and the questions it needs. No open question blocks PR-13.

Checks: `make verify` and `make pr-check`, free. Codex reviews PR-12, because it sets the safety rules that the policy code must follow.

### PR-13 - The workout domain model and catalog

Branch: `feat/pr-13-domain-model`. Work area 3.1. It needs PR-12 on `main`.

Concerns:

- the types of the domain: machine, exercise, inventory entry, dumbbell set, plan, session, working set, calibration set, set log (D-57), and cardio log (D-123),
- loads in pounds alone (D-122). An inventory entry holds the identity and the available weights alone (D-54). A dumbbell load is the load of one dumbbell (D-155),
- the catalog of D-155 as data with stable ids. Each exercise names its kind (machine, cable, dumbbell, or cardio) and its region,
- the lookups of the catalog: by id, by kind, and by region,
- the checks of each type, such as reps in a set log and weights of a stack.

The owner answered the open points in the session of PR-13. One machine with two movements gives two exercises (D-159). A load is a count of tenths of a pound (D-160). The catalog has six regions (D-161), and the adjustable bench is part of the dumbbell set (D-163). Pain is a rating from 0 to 10 (D-162). A set log has reps and reps in reserve of 0 or more (D-164), and a cardio distance is in tenths of a mile (D-165).

The heaviest dumbbell of a set is at most 100 lb (D-166).

Acceptance story: `make go-test` runs table tests for every type and for each lookup of the catalog. The catalog holds each item of D-155. A value outside the rules of its type fails its check.

Checks: `make go-test` and `make verify`, free. Codex reviews PR-13.

Questions for the session: none open.

### PR-14 - The policy bounds and progression

Branch: `feat/pr-14-policy-progression`. Work area 3.2. It needs PR-13 on `main`.

Concerns:

- the policy package, with its version and a rule id for each rule,
- the bounds of reps, reps in reserve (D-37), rest, and load (D-54),
- the rounding of D-65, with the tie rule of D-148 and the machine weight of D-149. A dumbbell rounds for each dumbbell,
- double progression with the one 5 lb step of D-147 (scenario B),
- missed reps (scenario A), a pain report (scenario C), and an early end (scenario D, D-63, D-64),
- the fixed warning text of D-153 for a pain report,
- the free-weight rules of `docs/research/exercise-safety.md` section 5.14 that the owner adopts.

Acceptance story: the golden tests of scenarios A to D pass. Property tests prove each property of section 6.3 of the high-level roadmap for every input.

Checks: `make go-test` and `make verify`, free. Codex reviews PR-14.

Questions for the session: which of REC-6, REC-9, REC-14, REC-15, and REC-17 to REC-22 the policy adopts. REC-9 holds the pain hold for two sessions, and D-153 already declined its referral text. The rep range limits of section 5.2 are recommendations too.

### PR-15 - The re-entry, calibration, and rules fallback

Branch: `feat/pr-15-policy-fallback`. Work area 3.2. It needs PR-14 on `main`.

Concerns:

- the long-break table and the first sessions after a break (D-151), with no failure in those sessions (D-37),
- the calibration of a new exercise (D-150),
- the rules fallback: a target from the rules alone when Luna fails or the policy refuses a proposal (D-23),
- the decision record of each plan decision. It holds the policy version, the model id, a hash of the prompt, the rules that applied, and the loads before and after the rounding,
- the golden tests of scenarios E and F.

Acceptance story: the golden tests of scenarios E and F pass. A proposal with a 50 percent load jump becomes a refusal and a fallback target, and the decision record names the refusal.

Checks: `make go-test` and `make verify`, free. Codex reviews PR-15.

Questions for the session: the reactive deload triggers (REC-7) and the fields of the decision record (REC-12). The session also asks about the fallback target for an exercise with no history.

### PR-16 - The Luna role layer

Branch: `feat/pr-16-luna-role-layer`. Work area 3.3. It needs PR-15 on `main`.

Concerns:

- the planner and reviser roles on `gpt-6-luna` at medium effort (D-22, D-24), with the strict plan schema of the spike,
- the fake provider for tests (D-24),
- a cost record for each call, and a cap hook that refuses a call over the cap (D-25). Q-98 gives the cap values in Phase 4, so the hook reads them from the configuration,
- a prompt that keeps each text inside the fitness boundary (D-36), with the dated copy of the usage policies (D-93),
- the versioned catalog of mobility and recovery texts, which Luna selects by id (D-152).

Acceptance story: the fake-provider tests cover a valid proposal, a malformed proposal, an unsafe proposal, and a timeout. The policy turns the unsafe proposal into a refusal and a fallback target. No test calls OpenAI.

Checks: `make go-test` and `make verify`, free. Codex reviews PR-16.

Questions for the session: the blocked-claims filter on Luna text (REC-11), and which shown text Luna writes and which text comes from templates.

### PR-17 - The Phase 3 Luna evaluation

Branch: `test/pr-17-luna-evaluation`. Work areas 3.2 and 3.3. It needs PR-16 on `main`.

Concerns:

- an evaluation command that sends test profiles through the planner, the reviser, and the policy,
- **Paid:** a limited run on `gpt-6-luna`, after the owner approves the run and its cap at run time (D-25),
- the OpenAI key in Secret Manager of `nk-workout-app-prod`, after the owner approves the change,
- a report "docs/research/phase-3-check.md" with the schema pass rate, the rate of policy refusals, the result of each scenario, and the cost.

Acceptance story: the report gives each count and the cost. Each scenario of section 5 of the high-level roadmap gives the safe behavior with the live model.

Checks: `make go-test` and `make verify`, free. Codex reviews PR-17. **Paid:** the spike of PR-2 used a price of 0.10 USD for each million input tokens (`docs/research/luna-plan-spike.md`). So a run of the same size costs cents (assumption).

Questions for the session: the cap of the run, and the number of profiles.

## 5. Exit of the phase

Phase 3 ends when PR-17 merges. The golden tests, the property tests, and the fake-provider tests of PR-13 to PR-16 pass in CI. The report of PR-17 gives the result of each scenario with the live model.
