# Pull request 3 review

Date: 2026-09-28

## Identity

- PR: 3
- Work area: 1.1 Luna plan spike
- Milestone: Measure Luna plan output against the first draft of the schema and policy.
- Target: `main`
- Base: `ddd0f49992fccf5e17be0349c912c127fd2e786a`
- Merge base: `ddd0f49992fccf5e17be0349c912c127fd2e786a`
- Head: `365c2045ec80d68fe5f97914e05e74abe99b599e`
- Branch: `feat/pr-2-luna-plan-spike`

## Provider gate

The hand-off session record names Claude Code as the author provider. This Codex session is the reviewer. The providers differ, so the gate passes.

## Intended behavior and scope

The milestone evaluates 20 synthetic profiles with three plan requests each. Its acceptance story requires schema and policy rates, cost per plan, caught unsafe proposals, and a go or no-go under D-101. The change adds a spike harness, its schema, its draft policy, test fixtures, results, and reports. It also adds spike tests to `make test` and updates the related documents.

The review applies D-23, D-24, D-25, D-36, D-41, D-53, D-54, D-65, D-80, D-98, D-100, and D-101. The policy rules remain drafts. Phase 3 owns the product policy and the revision path.

Inspected all changed paths: `AGENTS.md`, `Makefile`, `README.md`, `docs/decisions.md`, `docs/design.md`, `docs/questions.md`, `docs/research/luna-plan-spike.md`, `docs/roadmaps/phase-1-risk-spikes.md`, `docs/session-handoff.md`, `tools/spikes/luna_plan/README.md`, `tools/spikes/luna_plan/harness.py`, `tools/spikes/luna_plan/machines.json`, `tools/spikes/luna_plan/plan_schema.json`, `tools/spikes/luna_plan/policy.py`, `tools/spikes/luna_plan/profiles.json`, `tools/spikes/luna_plan/prompt.py`, `tools/spikes/luna_plan/providers.py`, `tools/spikes/luna_plan/results/paid-run-plans.jsonl`, `tools/spikes/luna_plan/results/paid-run-summary.json`, `tools/spikes/luna_plan/results/smoke-summary.json`, `tools/spikes/luna_plan/roles.json`, `tools/spikes/luna_plan/schema_check.py`, and `tools/spikes/luna_plan/test_luna_plan.py`.

## Findings

### P2-1: Retries can spend beyond the reserved cap

Status: open.

Open at: `365c2045ec80d68fe5f97914e05e74abe99b599e`.

File: `tools/spikes/luna_plan/harness.py:74-81`, `tools/spikes/luna_plan/providers.py:220-241`.

Trigger: A paid request times out after the provider accepts it, then the retry loop sends up to two more requests.

Expected: The run cap of D-98 bounds total provider charges. The README says the harness reserves the worst-case cost before each call.

Actual: The harness reserves once for the logical plan request. The provider can send three billable requests. It returns usage only for the final response, and the harness settles only that usage.

Consequence: Repeated ambiguous timeouts can incur charges that the budget neither reserves nor counts. With the committed role and a representative profile, one reservation is about $0.0163. Sixty reservations are about $0.98, while three charged attempts per request can exceed the $2 cap.

Correction: Reserve the worst-case amount for every possible billable attempt, or stop retries when the remaining cap cannot cover another attempt. Count all known attempt usage in the result.

Regression check: Add a fake timeout-after-acceptance case. Prove that retries stop within the cap. I did not run it because this review changes no code.

## Out of scope

The word-list policy does not prove the meaning of all generated text. The report states this limit, and Phase 3 owns the product policy and its evidence.

## PR comments

No comments, review comments, or review threads exist on PR #3.

## Description edits

None.

## Verification

- `git fetch origin`: passed. `origin/main` is `ddd0f49992fccf5e17be0349c912c127fd2e786a`.
- `git status --short --branch`: At review start, the status showed a clean detached worktree at the PR head.
- `gh pr view 3 --repo nkramber/workout-app --json ...`: PR #3 is open. Its base is `main`, branch is `feat/pr-2-luna-plan-spike`, and head is `365c2045ec80d68fe5f97914e05e74abe99b599e`.
- `git diff --stat ddd0f49992fccf5e17be0349c912c127fd2e786a..HEAD`: 23 paths, 2,379 insertions, and 13 deletions.
- `gh pr checks 3 --repo nkramber/workout-app`: `pr-contract`, `verify:lint`, and `verify:test` passed. `review-gate` failed because the review record did not yet exist.
- `make verify` at `365c2045ec80d68fe5f97914e05e74abe99b599e`: passed. STE and reference checks each found 0 issues. All 301 document-tool tests and 40 spike tests passed.
- Paid retry-cap reproducer: not run. The causal path is the single reservation in `harness.py` and the three-attempt loop in `providers.py`.
- Paid target: not run. The review skill forbids paid targets.
- Push: pending.

## Open questions and accepted risks

The owner must decide whether the paid-call retry budget needs a different cap rule. Finding P2-1 stays open until a correction passes its regression check.

## Verdict

**Changes required.** This verdict applies to head `365c2045ec80d68fe5f97914e05e74abe99b599e`.
The retry loop can make more billable calls than the harness reserves under the stated cap. The correction needs a focused regression check before approval.
