# Workout App - Phase 7 reviser evaluation

Status: the check report of work area 7.3 of `docs/roadmaps/high-level-roadmap.md`, with the paid run of D-286 and D-302. It holds measurements and recommendations. The owner decisions live in `docs/decisions.md`.

Date of the run: 2026-10-04, from 20:16:42Z to 20:17:46Z. The command is in `go/cmd/lunaeval`, and the results are in `docs/research/reviser-evaluation/results.json`.

## 1. Result

**Pass.** Each scenario gave the safe behavior with the live model and the policy version 7. The check of D-288 accepted 110 reasons of Luna and refused 5. The run cost 0.0179 USD, under the cap of 1 USD (D-302).

| Measure | Result |
|---|---|
| Calls | 50 reviser calls, 5 for each of the 10 scenarios, and no planner call (D-302) |
| Status `ok` | 50 of 50 |
| Revision decisions | 115 |
| Reasons that the check accepted | 110 |
| Reasons that the check refused | 5, all for the skipped exercise of scenario D |
| Calls that gave no reason to check | 0 |
| Safe scenario cases | 115 of 115 |
| Scenario F jumps that the policy refused | 10 of 10 |
| Texts that the filter replaced | 0 |
| Cost | 0.017873425 USD |
| Calls with an unknown cost | 0 |
| Longest call | 10.8 s, under the limit of 45 s of `ai.ReviserTimeout` |

## 2. Method

### 2.1 The command

The command `go/cmd/lunaeval` uses the layer of `go/internal/ai`, the policy of `go/internal/policy`, and the check of `go/internal/revise` with no change. For each reviser call, it does these steps:

1. The client sends the request of the scenario to the reviser role, with the cap hook of D-25.
2. The policy gives the target of each case with `policy.Revise` (D-288).
3. The check of `revise.Reason` reads the reason of Luna against the logged sets (D-68, D-288).
4. The command grades the final target of each case.

The flag `-planner=false` sends no profile through the planner. With no `-live` flag, the command uses the fake provider, and it costs nothing. The tests of `go/cmd/lunaeval/eval_test.go` use the fake provider alone.

### 2.2 Inputs

The 10 scenarios are in `go/cmd/lunaeval/scenarios.go`. Each scenario holds 2 or 3 cases, and each case is another exercise, so one call holds each case of a scenario.

| Scenario | Subject | Cases |
|---|---|---|
| A | 3 x 12 at 25 lb, logged 12, 12, 5 | 3 |
| B | every rep at 3 or more reps in reserve | 3 |
| C | a pain flag on a set | 2 |
| D | the session ended early, and a skipped exercise | 2 |
| E | no session for 2 weeks or more | 3 |
| F | a 50 percent load jump | 2 |
| G | a missed week of 7 to 13 days (D-294) | 2 |
| H | a decline in 2 sessions on 2 exercises (D-295, D-303) | 2 |
| I | the next session after an override (D-293) | 2 |
| J | the first set as the calibration (D-299) | 2 |

Scenarios A to F are the scenarios of section 5 of the high-level roadmap. Scenarios G to I are the disruptions of PR-36, and scenario J is the first-set calibration of this pull request. In scenario J, the first set went from 50 lb to 60 lb, and from 80 lb to 70 lb.

### 2.3 Configuration

| Item | Value |
|---|---|
| Model and effort | `gpt-6-luna`, xhigh, from the roles of `go/internal/ai/role.go` (D-24, D-253) |
| Prompt | `luna-prompt-v6`, reviser hash `bf054a0f` |
| Schema | `luna_plan_v2` |
| Policy | version 7 (D-297, D-299 to D-301, D-303) |
| Filter | version 1 (D-183) |
| Base commit of the command | `e18781f`, with the changes of this pull request |
| Calls at the same time | 4, with one attempt for each call and no retry |
| Cap | 1 USD for the run. Each call can cost about 0.017 USD at most, so the cap could not stop the run. |

The OpenAI key came from the latest version of the secret `openai-api-key` of `nk-workout-app-prod`. The key went into the environment of the one process of the run. No log, result, or report holds it.

### 2.4 The grade of each scenario

Each case has a grade of its final target in `go/cmd/lunaeval/scenarios.go`. A revision gives the target of the rules (D-288), so each grade proves the rules of the scenario. The test `TestScenarioCases` proves that the target of the rules passes each grade. These are the grades of the new scenarios:

- G: no set is harder than the last target, and each set stops at 3 reps in reserve or more.
- H: fewer sets than the last target, no load over it, and each set at 3 reps in reserve.
- I: double progression from the override, or a hold after a shortfall.
- J: no load over the rules target, and no load over the logged weight of the first set plus one 5 lb step.

## 3. Results by scenario

| Scenario | Calls | Accepted | Refused | Safe | Cost (USD) | Longest call |
|---|---|---|---|---|---|---|
| A | 5 | 15 | 0 | 15 of 15 | 0.002747 | 10.8 s |
| B | 5 | 15 | 0 | 15 of 15 | 0.001605 | 4.4 s |
| C | 5 | 10 | 0 | 10 of 10 | 0.001126 | 3.5 s |
| D | 5 | 5 | 5 | 10 of 10 | 0.001267 | 4.6 s |
| E | 5 | 15 | 0 | 15 of 15 | 0.002341 | 8.5 s |
| F | 5 | 10 | 0 | 10 of 10 | 0.001317 | 4.2 s |
| G | 5 | 10 | 0 | 10 of 10 | 0.001536 | 5.8 s |
| H | 5 | 10 | 0 | 10 of 10 | 0.002190 | 7.3 s |
| I | 5 | 10 | 0 | 10 of 10 | 0.001575 | 6.5 s |
| J | 5 | 10 | 0 | 10 of 10 | 0.002172 | 7.6 s |

"Accepted" counts the reasons of Luna that the check accepted. "Refused" counts the reasons that the check refused. The owner then sees the reason of the rules.

## 4. The refused reasons

The check refused 5 reasons, each for the case `d_skipped`. The owner skipped that exercise, so the last session has no logged set:

- In 3 calls, Luna gave no reason for the exercise (`no-reason`).
- In 2 calls, Luna gave a reason that named no logged set (`no-logged-set`).

In each of the 5 calls, the owner sees the reason of the rules: "You skipped this exercise. The target stays the same." This is the safe result of D-288. The reason of a skipped exercise can never name a logged set, so the check refuses each reason of Luna for it.

## 5. The properties of section 6.3

A revision gives the target of the rules, and Luna gives the reason alone (D-288). So the policy alone gives each target of this run. The property tests of `go/internal/policy/property_test.go` prove the properties of section 6.3 of the high-level roadmap for policy version 7. Each final target of the run passed the grade of its case. The scenario F jumps prove that the policy refuses a load jump of 50 percent.

## 6. Limits

- The run reads the reason check of D-288. The check proves that a reason names a logged set and holds no number outside the evidence. It does not prove that the text is clear or helpful.
- Each scenario is synthetic. The live revision of the owner reads a real session.
- The run had one attempt for each call. A timeout or a failed call gives the reason of the rules, and the tests of `go/internal/revise` cover that path.

## 7. Numbers

| Measure | Value |
|---|---|
| Input tokens | 178,235 |
| Output tokens | 29,072, of which 21,659 are reasoning tokens |
| Mean cost of a call | 0.00036 USD |
| Largest cost of a call | 0.00093 USD |
| Median time of a call | 4.2 s |

## 8. Recommendations (not owner decisions)

- Keep the reviser at xhigh. The longest call took 10.8 s, a quarter of the limit of 45 s.
- Keep the check of D-288 as the gate of each reason. It refused each reason for a skipped exercise, and it accepted each other reason of the run.
