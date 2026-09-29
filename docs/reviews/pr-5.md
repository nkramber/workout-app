# Pull request 5 review

Date: 2026-09-28

## Identity

- PR: 5
- Work area: 1.2 Recognition spike
- Milestone: Add the recognition harness, run it on the test set, and report the result against D-107.
- Target: `main`
- Base: `8a034c9b18e36ccf6907259cd55a9b8e5448303b`
- Merge base: `8a034c9b18e36ccf6907259cd55a9b8e5448303b`
- Head: `63fe65df968140dead8aae5f497766fd41d8cb47`
- Branch: `feat/pr-4-recognition-spike`

## Provider gate

The current hand-off names Claude Code as the author provider for Session 5. The active reviewer is Codex. The providers differ, so the gate passes.

## Intended behavior and scope

The acceptance story requires the correct, wrong, and abstained counts. It also requires each wrong answer with high stated confidence, the cost per photo, and a go or no-go under D-107. The milestone also requires a fake provider and a paid run under D-25. The owner decisions defer photo recognition after the no-go. D-24, D-25, D-80, D-94, D-100, D-107 to D-112, and Q-96 apply.

I inspected all 20 changed paths from the merge base: `README.md`, `docs/decisions.md`, `docs/design.md`, `docs/questions.md`, `docs/research/platform-cloud-and-ai.md`, `docs/research/recognition-spike.md`, `docs/roadmaps/high-level-roadmap.md`, `docs/roadmaps/phase-1-risk-spikes.md`, `docs/session-handoff.md`, `tools/spikes/luna_plan/providers.py`, `tools/spikes/luna_plan/test_luna_plan.py`, `tools/spikes/recognition/README.md`, `tools/spikes/recognition/photos.py`, `tools/spikes/recognition/recognize.py`, `tools/spikes/recognition/results/paid-run-answers.jsonl`, `tools/spikes/recognition/results/paid-run-summary.json`, `tools/spikes/recognition/results/smoke-summary.json`, `tools/spikes/recognition/roles.json`, `tools/spikes/recognition/test_recognition.py`, and `tools/spikes/recognition/vision.py`.

The shared provider keeps the plan request path. It sends the image request through the same retry loop and budget gate. The recognizer omits the photo id and truth from the request. The harness validates confidence, scores abstentions, and records unknown charges at the worst case. It applies the bar to all requested photos. The report and summary agree.

Fifteen of 174 photos had a high confidence wrong answer. The run had 136 correct answers. The cost was 0.000367 USD per answered photo. The high confidence wrong rate was 8.6%, so the result is no-go. The roadmap and design defer photo recognition under D-110 and D-111.

## Findings

No finding.

## Out of scope

Photo capture, retention, and deletion belong to Deferred - Photo recognition in `docs/roadmaps/high-level-roadmap.md`. Q-103 stays open for that work under D-111.

## PR comments

No issue comments, review comments, or review threads exist on PR #5.

## Description edits

None.

## Verification

- `git fetch origin` and `git status --short --branch`: passed at start. The tree was clean at detached HEAD `63fe65df968140dead8aae5f497766fd41d8cb47`.
- `python3 docs/tools/review_gate.py --effective-head 5`: passed. It returned `63fe65df968140dead8aae5f497766fd41d8cb47`.
- `git diff --stat 8a034c9b18e36ccf6907259cd55a9b8e5448303b...HEAD`: 20 paths, 1,748 insertions, and 68 deletions. I inspected every path.
- GraphQL export of PR #5 review threads and reviews: passed. It returned zero threads and zero reviews.
- `gh pr checks 5 --repo nkramber/workout-app` at head `63fe65df968140dead8aae5f497766fd41d8cb47`: `pr-contract`, `verify:lint`, and `verify:test` passed. `review-gate` failed because the branch did not hold `docs/reviews/pr-5.md`. The failed gate log confirms RG 3 only.
- `make verify`: passed. STE, ref, lifecycle, context, 301 documentation tests, 44 plan tests, 19 recognition tests, and 45 recognition-set tests passed. One recognition image test and two recognition-set image tests skipped because Pillow is not installed.
- `make pr-check`: did not run because this detached review worktree has no pull request association. The `pr-contract` check passed on the PR head.
- Paid targets: not run, as required by the review skill.
- `make hooks`: passed. Hooks are installed in this checkout.
- `make where`: passed before the record commit. It reports detached HEAD. The requested push target is `feat/pr-4-recognition-spike`.
- Push: pending.

## Open questions and accepted risks

The image preparation test skips when Pillow is absent. The paid run records the prepared request outputs, and the report names the skip in CI. Q-103 remains open for the deferred photo work.

## Verdict

**Ready for owner merge.** This verdict applies to head `63fe65df968140dead8aae5f497766fd41d8cb47`.
The provider gate passes. The reviewer inspected all changed paths and the acceptance story. Local verification and substantive CI checks pass. The review-gate check awaits this record on the branch.
