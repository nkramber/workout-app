# Pull request 4 review

Date: 2026-09-28

## Identity

- PR: 4
- Work area: 1.2 Recognition test set
- Milestone: Add the recognition test set, its manifest checks, source/download/degradation tools, and catalog shortlist.
- Target: `main`
- Base: `89046b4219827aa3dbd5b505c228adcea48cf149`
- Merge base: `89046b4219827aa3dbd5b505c228adcea48cf149`
- Head: `81f74d624feebc88bd20f602a707ff45f9735c69`
- Branch: `feat/pr-3-recognition-test-set`

## Provider gate

The session record for this work in `docs/session-handoff.md` names Claude Code as the author provider. This session is Codex. The providers differ, so the gate passes.

## Intended behavior and scope

The acceptance story requires a free manifest check and a downloader that fills the ignored local cache with every manifest image whose SHA-256 agrees. The milestone also adds a 26-type catalog, 130-image manifest and 44 degraded-photo recipes, source metadata verification, tests, and updates to the roadmap and research. The relevant contracts are D-45, D-56, D-95, D-97, D-102 to D-105, and section 6.6 of `docs/research/platform-cloud-and-ai.md`.

I inspected all 14 changed paths from the base: `README.md`, `docs/decisions.md`, `docs/questions.md`, `docs/research/platform-cloud-and-ai.md`, `docs/roadmaps/phase-1-risk-spikes.md`, `docs/session-handoff.md`, and the eight files in `tools/spikes/recognition_set/`. I ran the manifest validator and the live source check. Git tracks no image files.

## Findings

### P2-1: Source verification ignores recorded source and license URLs

Status: open.

Open at: `81f74d624feebc88bd20f602a707ff45f9735c69`.

File: `tools/spikes/recognition_set/sources.py:174-180`.

Trigger: A manifest entry has a wrong `source_url` or `license_url`, while its source record still returns the same license, author, and file URL.

Expected: The manifest's source URL, author, and license must agree with their source (D-97, D-102, D-105). The source check must detect a wrong Flickr photo page or license URL, and a wrong Commons license URL.

Actual: `sources.check` compares Commons `license`, `author`, `file_url`, and `bytes`, and Flickr `license`, `author`, and `file_url`. It does not compare `source_url` for either provider or `license_url` for either provider. The manifest validator checks only that a license URL contains the expected license path, not that it matches the source URL or trusted license host.

Consequence: The check can report no differences while the committed attribution points to a different source page or an unrelated license URL. This weakens the provenance evidence for the licensed test set.

Correction: Compare each recorded source URL and license URL with the source record. Reject a license URL that does not name the matching license at its trusted URL. Add regression cases for wrong Flickr and Commons URLs.

Regression check: Change each field in a source-check test. Keep the source response fixed. The check must fail each test. No such regression test exists at this head.

## Out of scope

Recognition thresholds and the paid model run belong to the later recognition-spike work area 1.3.

## PR comments

No issue comments, review comments, or review threads exist on PR #4.

## Description edits

None.

## Verification

- `git fetch origin` and `git status --short --branch`: passed before review. Git had no current branch at `81f74d624feebc88bd20f602a707ff45f9735c69`. The tree was clean.
- `python3 docs/tools/review_gate.py --effective-head 4`: passed. It returned `81f74d624feebc88bd20f602a707ff45f9735c69`.
- `git diff --stat 89046b4219827aa3dbd5b505c228adcea48cf149...81f74d624feebc88bd20f602a707ff45f9735c69`: 14 paths, 4,487 insertions, 37 deletions.
- `make verify` at `81f74d624feebc88bd20f602a707ff45f9735c69`: passed. STE, reference, lifecycle, context, 301 document-tool tests, 44 Luna tests, and 40 recognition-set tests passed. Two image tests skipped because Pillow is not installed.
- `python3 tools/spikes/recognition_set/download.py`: passed at the reviewed head. It downloaded 130 images and returned 0 errors. Each SHA-256 matched.
- `python3 tools/spikes/recognition_set/download.py --offline`: passed. It found 130 images in cache, downloaded 0, and returned 0 errors.
- `python3 tools/spikes/recognition_set/manifest_check.py`: passed. It found 130 images, 44 degraded photos, and 0 errors.
- `python3 tools/spikes/recognition_set/sources.py check`: passed with 0 differences. It does not cover the URL fields in P2-1.
- `PR_BODY_FILE=/tmp/pr4-body.md PR_TITLE='feat: the recognition test set (PR-3)' make pr-check`: did not pass in this review worktree. The checker needs a branch name, but the PR head is `HEAD`. It reported branch and session-branch mismatches.
- GitHub checks at the reviewed head: `pr-contract`, `verify:lint`, and `verify:test` passed. `review-gate` failed RG 3 because `docs/reviews/pr-4.md` did not exist. The log confirms this.
- PR comments and review threads: none. I read `gh pr view` and saved the GraphQL `reviewThreads` query at `/tmp/pr4-review-threads.json`.
- Pillow-dependent visual generation: not run because Pillow is unavailable. The two related unit tests skipped. The author reports running the degrade script and examining samples in the PR body.
- Paid target: not run, as required by the review skill.
- Push: pending.

## Open questions and accepted risks

Pillow is unavailable in this environment, so the degradation output was not independently inspected. The author reports a successful run and sample inspection in the pull request body.

## Verdict

**Changes required.** This verdict applies to head `81f74d624feebc88bd20f602a707ff45f9735c69`.
The source verifier does not validate required source and license URLs. Correct P2-1 and rerun its URL mismatch regression checks before approval.
