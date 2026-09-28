# Luna plan spike

This folder holds the harness of the Luna plan spike, work area 1.1 of `docs/roadmaps/high-level-roadmap.md`. The report is `docs/research/luna-plan-spike.md`. The spike code stays outside the product code (D-100).

## Files

| File | Content |
|---|---|
| `profiles.json` | 20 synthetic profiles with the onboarding inputs of D-41, and the inventory of each profile. No field holds personal data (D-53). |
| `machines.json` | The synthetic machine catalog: identity and available weights only (D-54). |
| `plan_schema.json` | The JSON schema of the plan output, in the strict form of the OpenAI structured outputs. |
| `schema_check.py` | A small validator for the keywords of the schema. The spike adds no dependency. |
| `policy.py` | The first draft of the policy rules table (D-23). Each rule has an id, a version, its sources, and its open questions. |
| `prompt.py` | The prompt, with a version and a hash. |
| `roles.json` | The role configuration. It holds the model id, the effort, the prices, and the cap of the run (D-24, D-98). |
| `providers.py` | The fake provider and the OpenAI provider. |
| `harness.py` | The harness. It sends each profile to the plan role three times, and checks each answer against the schema and the policy. |
| `test_luna_plan.py` | The unit tests. `make test` runs them with the fake provider only. |

## Free run

The fake provider makes no network call. It starts from a plan that obeys every rule, and it breaks some plans in a known way. So a free run shows each part of the summary.

```bash
python3 tools/spikes/luna_plan/harness.py
```

## Paid run

CAUTION: A paid run spends the OpenAI API account of the owner. The owner approves each paid run at run time (D-25). The cap in `roles.json` is 2 USD (D-98), and the harness refuses a higher cap.

1. Get the approval of the owner for the run.
2. Put `OPENAI_API_KEY` in the environment of the harness process only. Do not write the key into a log, a result, or a report.
3. Run the harness with the approval flag:

```bash
python3 tools/spikes/luna_plan/harness.py --provider openai --owner-approved
```

Before each attempt, retries included, the harness reserves the worst-case cost of the attempt. It skips each plan that the cap can not cover, and it records the plan as `cap_skip`. It stops the retries of a plan when the cap can not cover one more attempt. An attempt with an unknown charge, such as a timeout or a server error, counts at its worst-case cost. Each request sets `store` to false.

## Results

The harness writes each run to `.local/spikes/luna_plan/<run id>/`, which Git ignores:

- `plans.jsonl`: one record for each plan, with the raw text, the schema errors, the violations, the tokens, and the cost.
- `summary.json`: the rates, the counts for each rule, the cost, and the go result of D-101. It also gives the attempts with an unknown charge, and the spend bound.
- `summary.md`: the same summary as Markdown tables.
