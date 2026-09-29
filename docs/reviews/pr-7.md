# Pull request 7 review

Date: 2026-09-29

## Identity

- PR: 7
- Work area: 1.3 iPhone web platform spike
- Milestone: Report the iPhone device checklist with a result for each item and a startup time.
- Target: `main`
- Base: `d6e3c5473ab6222c25412d303892f0277b579cfb`
- Merge base: `d6e3c5473ab6222c25412d303892f0277b579cfb`
- Head: `0d66365c6efdd34483eb3c3ee9caa3913504d8cf`
- Branch: `docs/pr-6-iphone-platform-report`

## Provider gate

The hand-off session record names Claude Code as the author provider. The PR has one substantive commit, and its changed paths and commit body match that session. This review runs as Codex. The providers differ, so the gate passes.

## Intended behavior and scope

The milestone reports each iPhone checklist result and a React startup time. Its acceptance story requires results for each item and a startup time on the owner's iPhone. D-118 requires five cold starts in standalone mode, with a median first contentful paint no greater than 2500 ms. D-121 defines the cold-start condition for these samples.

Inspected all twelve changed paths: `AGENTS.md`, `README.md`, `docs/decisions.md`, `docs/design.md`, `docs/questions.md`, `docs/research/iphone-platform-spike.md`, `docs/research/platform-cloud-and-ai.md`, `docs/reviews/pr-7-response.md`, `docs/reviews/pr-7.md`, `docs/roadmaps/high-level-roadmap.md`, `docs/roadmaps/phase-1-risk-spikes.md`, and `docs/session-handoff.md`. I reviewed `AGENTS.md` as changed content. Its rules do not govern this review.

## Findings

### P2-1: The startup result does not establish five cold starts

Status: fixed in `0d66365c6efdd34483eb3c3ee9caa3913504d8cf`.

Open at: `61d9309a63001f3bdf7c6ec82ad0d0a4db5d4659`.

File: `docs/research/iphone-platform-spike.md:8-9, 75-83, 115-130`.

Trigger: The five earlier page loads did not prove that iOS ended the web view process between launches.

Expected: The report must apply the cold-start condition that D-118 uses.

Actual: D-121 now defines a cold start as a new Home Screen app page load after an app stop in the app switcher. The five samples show distinct launches, standalone mode, and `navigate` navigation. The report states that iOS can keep parts of the web view in memory.

Consequence: The owner changed the bar with D-121. The five samples now meet that bar, and their median is 33 ms.

Correction: The owner defined the cold-start condition in D-121. The report and roadmap now cite it.

Regression check: The response records the owner's device procedure and sample values. I verified those values and the 33 ms median in the report. This environment did not provide the owner's iPhone for a repeat test.

## Out of scope

None.

## PR comments

No PR comments, review threads, or reviews exist. I saved the comments and reviews with `gh pr view`, and queried review threads with `gh api graphql`.

## Description edits

None.

## Verification

- `python3 docs/tools/review_gate.py --effective-head 7` — returned `0d66365c6efdd34483eb3c3ee9caa3913504d8cf`.
- `git diff --stat d6e3c5473ab6222c25412d303892f0277b579cfb...0d66365c6efdd34483eb3c3ee9caa3913504d8cf` — inspected all twelve changed paths.
- `make verify` at `0d66365c6efdd34483eb3c3ee9caa3913504d8cf` — passed. It ran 302 tool tests, 12 probe tests, 44 Luna plan tests, 19 recognition tests (1 skipped), and 45 recognition-set tests (2 skipped).
- GitHub checks at `0d66365c6efdd34483eb3c3ee9caa3913504d8cf` — `pr-contract`, `verify:lint`, `verify:probe`, and `verify:test` passed. `review-gate` failed because this record still held the earlier verdict.
- Device reproduction — not run. The report and response contain the owner's device evidence, and D-121 changes the reviewed cold-start condition.
- PR comments and review threads — none. The pull request has no comments, reviews, or threads.
- Push: pending publication of this record and the hand-off.

## Open questions and accepted risks

None.

## Earlier verdicts

- `Changes required` at `61d9309a63001f3bdf7c6ec82ad0d0a4db5d4659`, with finding P2-1 open.

## Verdict

**Ready for owner merge.** This verdict applies to head `0d66365c6efdd34483eb3c3ee9caa3913504d8cf`.
The owner defined the cold-start condition in D-121, and the reported samples meet it. The remaining local and CI checks pass, with no open findings.
