# Workout App - Luna effort check: medium and xhigh

This report compares two reasoning efforts of `gpt-6-luna` with the same calls. The owner asked for the check after the live check of Phase 5 (D-242). The date of the runs is 2026-10-03.

## 1. Result

Both efforts gave the safe behavior in each call. xhigh cost 1.74 times as much as medium, and it took about 1.6 times as long. The schema passes, the policy refusals, and the grades of the scenarios did not change. The one visible change: xhigh put cardio in more plans.

The owner then chose xhigh for both roles (D-253). The two runs cost 0.1166 USD in total, under the cap of 2 USD for each run.

## 2. Method

### 2.1 The command

The command `go/cmd/lunaeval` of the Phase 3 check ran two times (`docs/research/phase-3-check.md`). PR-28 added the flag `-effort`. The flag sets the effort of each call through the field `Effort` of `ai.Client`, and the prompt and its hash do not change. The flag accepts only the efforts of the model page: none, low, medium, high, xhigh, and max (PC-61 of `docs/research/platform-cloud-and-ai.md`).

The commands, from the folder `go`:

```bash
OPENAI_API_KEY="$(gcloud secrets versions access latest --secret=openai-api-key --project=nk-workout-app-prod)" \
  go run ./cmd/lunaeval -live -cap 2 -effort medium -out ../docs/research/luna-effort-check/medium.json
OPENAI_API_KEY="$(gcloud secrets versions access latest --secret=openai-api-key --project=nk-workout-app-prod)" \
  go run ./cmd/lunaeval -live -cap 2 -effort xhigh -out ../docs/research/luna-effort-check/xhigh.json
```

The owner approved both runs and their caps at run time (D-25). The key went into the environment of each process alone (D-187). No log, result, or report holds it. The runs used the cap of the command in memory, so they did not change the monthly spend of the app (D-188).

### 2.2 Inputs and configuration

| Item | Value |
|---|---|
| Calls of each run | 50: 20 planner calls for the synthetic profiles, and 30 reviser calls for the scenarios A to F, 5 times each |
| Model | `gpt-6-luna` |
| Prompt | `luna-prompt-v3` |
| Schema | `luna_plan_v2` |
| Policy | version 3 |
| Filter | version 1 |
| Code | The tree of PR-28 on the base `9d6f8eb`, with the `-effort` flag and the role at medium |
| Calls at the same time | 4, with one attempt for each call and no retry |
| Time limit of a call | 90 s, from the role |
| Output limit of a call | 32,000 tokens, from the role |
| Cap | 2 USD for each run |
| Medium run | 04:21:36Z to 04:24:33Z |
| xhigh run | 04:24:39Z to 04:29:48Z |

The results are in `docs/research/luna-effort-check/medium.json` and `docs/research/luna-effort-check/xhigh.json`. They hold synthetic data alone.

## 3. Results

### 3.1 Cost, time, and tokens

| Measure | medium | xhigh | Ratio |
|---|---|---|---|
| Cost of the run | 0.0426 USD | 0.0740 USD | 1.74 |
| Mean cost of a planner call | 0.00151 USD | 0.00261 USD | 1.73 |
| Largest cost of a planner call | 0.00234 USD | 0.00426 USD | 1.82 |
| Mean cost of a reviser call | 0.00041 USD | 0.00073 USD | 1.78 |
| Mean time of a planner call | 25.0 s | 40.4 s | 1.62 |
| Longest planner call | 39.5 s | 63.6 s | 1.61 |
| Mean time of a reviser call | 6.3 s | 13.6 s | 2.16 |
| Longest reviser call | 9.5 s | 25.1 s | 2.64 |
| Mean output tokens of a planner call | 2,602 | 4,765 | 1.83 |
| Largest output of a call | 4,254 tokens | 7,351 tokens | 1.73 |
| Mean reasoning tokens of a planner call | 856 | 2,570 | 3.00 |
| Input tokens of the run | 188,912 | 188,912 | 1.00 |
| Calls with an unknown cost | 0 | 0 | - |

A reasoning token bills as an output token (PC-64). So the cost grows with the reasoning tokens, and the input cost stays the same.

### 3.2 Safety and content

| Measure | medium | xhigh |
|---|---|---|
| Schema passes | 50 of 50 | 50 of 50 |
| Timeouts | 0 | 0 |
| Outputs cut off | 0 | 0 |
| Policy decisions | 196 | 196 |
| Proposals that the policy refused | 0 | 0 |
| Proposals equal to the target of the rules | 194 | 196 |
| Filtered texts | 0 | 0 |
| Scenarios A to F | Each passes | Each passes |
| Jumps that the policy refused in scenario F | 10 of 10 | 10 of 10 |
| Cardio items | 9 | 27 |
| Planner calls with cardio | 5 of 20 | 9 of 20 |

The two medium proposals that differ from the rules came from scenario C, a pain report. Luna gave 3 reps in reserve in place of the 2 of the rules, at the same load. The policy accepted both, because more reserve is not a risk.

xhigh put cardio in more plans, and in more sessions of each plan. Each cardio item stayed inside the bound of 5 to 30 minutes (D-232).

### 3.3 The live plan of the same day

The live check of Phase 5 used medium effort (D-212). The second plan request of the owner made 1 call. The call cost 0.00104 USD, and `RequestPlan` took 19.6 s on the server. With the ratios of section 3.1, the same request at xhigh can cost about 0.0018 USD and take about 32 s (assumption).

## 4. Effect on the limits

- **Time limit.** The longest xhigh call took 63.6 s of the limit of 90 s. A request has 4 calls at most (D-231). 4 times 90 s is 360 s, under the request timeout of 420 s of the service `api`. So the limits stay (D-253).
- **Output limit.** The largest xhigh output used 7,351 of 32,000 tokens.
- **Cap.** The cap reserves the worst case of the output limit, so the reservation does not change. One plan request at xhigh costs less than 0.01 USD, against the monthly cap of 1 USD for the user (D-188).

## 5. Limits of this check

- Each run is small: 20 planner calls and 5 calls for each scenario.
- The inputs of the planner hold no profile of D-209, and each call plans 3 sessions. The live app sends the profile and the training days of the owner.
- In both runs, Luna copied the targets of the rules almost each time. So the check shows no gain of quality from xhigh in the targets. Q-202 stays open for Phase 7.
- The check measured medium and xhigh alone. It did not measure high or max.
