# Pull request 3 review

Date: 2026-09-28

## Identity

- PR: 3
- Work area: 1.1 Luna plan spike
- Milestone: Measure Luna plan output against the first draft of the schema and policy.
- Target: `main`
- Base: `ddd0f49992fccf5e17be0349c912c127fd2e786a`
- Merge base: `ddd0f49992fccf5e17be0349c912c127fd2e786a`
- Head: `eb8164b99dd9bb8d82c002ad3d5b7d53abf408d9`
- Branch: `feat/pr-2-luna-plan-spike`

## Provider gate

The hand-off session record names Claude Code as the author provider. This Codex session is the reviewer. The providers differ, so the gate passes.

## Intended behavior and scope

The milestone evaluates 20 synthetic profiles with three plan requests each. Its acceptance story requires schema and policy rates, cost per plan, caught unsafe proposals, and a go or no-go under D-101. The change adds a spike harness, its schema, its draft policy, test fixtures, results, and reports. It also adds spike tests to `make test` and updates the related documents.

The review applies D-23, D-24, D-25, D-36, D-41, D-53, D-54, D-65, D-80, D-98, D-100, and D-101. The policy rules remain drafts. Phase 3 owns the product policy and the revision path.

Inspected every changed path from the base: `AGENTS.md`, `Makefile`, `README.md`, `docs/decisions.md`, `docs/design.md`, `docs/questions.md`, `docs/research/luna-plan-spike.md`, `docs/reviews/pr-3-response.md`, `docs/reviews/pr-3.md`, `docs/roadmaps/phase-1-risk-spikes.md`, `docs/session-handoff.md`, `tools/spikes/luna_plan/README.md`, `tools/spikes/luna_plan/harness.py`, `tools/spikes/luna_plan/machines.json`, `tools/spikes/luna_plan/plan_schema.json`, `tools/spikes/luna_plan/policy.py`, `tools/spikes/luna_plan/profiles.json`, `tools/spikes/luna_plan/prompt.py`, `tools/spikes/luna_plan/providers.py`, `tools/spikes/luna_plan/results/paid-run-plans.jsonl`, `tools/spikes/luna_plan/results/paid-run-summary.json`, `tools/spikes/luna_plan/results/smoke-summary.json`, `tools/spikes/luna_plan/roles.json`, `tools/spikes/luna_plan/schema_check.py`, and `tools/spikes/luna_plan/test_luna_plan.py`.

## Earlier verdicts

- **Changes required.** This verdict applied to head `365c2045ec80d68fe5f97914e05e74abe99b599e`. Finding P2-1 stated that retries exceeded the reserved cap.

## Findings

### P2-1: Retries can spend beyond the reserved cap

Status: fixed in `f5abf9ada0be99d93b699bdd3cc976998b29f14d`.

Open at: `365c2045ec80d68fe5f97914e05e74abe99b599e`.

File: `tools/spikes/luna_plan/harness.py:65-84`, `tools/spikes/luna_plan/providers.py:243-276`.

Trigger: A paid request times out after the provider accepts it, then the retry loop sends more requests.

Expected: The run cap of D-98 bounds total provider charges.

Actual: The prior code reserved once for the logical plan request. The provider sent up to three billable requests. The correction reserves before each attempt, counts an unknown charge at its worst-case cost, and stops when the cap can not fund another attempt.

Consequence: The prior code incurred charges beyond the cap under this trigger. The corrected gate bounds retries for the fixed request inputs. The estimate exceeds reported input usage for all 20 committed profiles, and the output reservation uses the configured maximum.

Correction: Reserve each attempt before the provider sends it. Settle known usage or the worst-case cost for an unknown charge. Stop retries when the gate refuses a reservation.

Regression check: `PYTHONPATH=tools/spikes/luna_plan python3 -m unittest test_luna_plan.RetryCapTest` at `eb8164b99dd9bb8d82c002ad3d5b7d53abf408d9`: passed, 4 tests. `make verify` at the same head also passed.

## Out of scope

The word-list policy does not prove the meaning of all generated text. The report states this limit, and Phase 3 owns the product policy and its evidence.

## PR comments

No issue comments, review comments, or review threads exist on PR #3.

## Description edits

None.

## Verification

- `git fetch origin main`: passed. `origin/main` is `ddd0f49992fccf5e17be0349c912c127fd2e786a`.
- `python3 docs/tools/review_gate.py --effective-head 3`: passed and returned `eb8164b99dd9bb8d82c002ad3d5b7d53abf408d9`.
- `git diff --stat ddd0f49992fccf5e17be0349c912c127fd2e786a..eb8164b99dd9bb8d82c002ad3d5b7d53abf408d9`: 25 paths, 2,599 insertions, and 13 deletions.
- `make verify` at `eb8164b99dd9bb8d82c002ad3d5b7d53abf408d9`: passed. STE, reference, lifecycle, and context checks passed. All 301 document-tool tests and 44 spike tests passed.
- `PYTHONPATH=tools/spikes/luna_plan python3 -m unittest test_luna_plan.RetryCapTest`: passed, 4 tests.
- Input reservation comparison on the 20 committed profile requests: passed. Each estimate was 1.46 to 1.49 times reported input usage.
- `gh pr checks 3 --repo nkramber/workout-app`: `pr-contract`, `verify:lint`, and `verify:test` passed. `review-gate` failed because the record still named the earlier head and verdict. A fresh result remains pending until GitHub checks this record.
- Paid target: not run. The review skill forbids paid targets.
- Push: pending publication of this review record and the hand-off.

## Open questions and accepted risks

The fresh `review-gate` result for the updated record remains pending publication.

## Verdict

**Blocked.** This verdict applies to head `eb8164b99dd9bb8d82c002ad3d5b7d53abf408d9`.
The earlier finding is fixed, and local verification passes. The current published head has no passing `review-gate` result for this verdict.
