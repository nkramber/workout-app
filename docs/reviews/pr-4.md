# Pull request 4 review

Date: 2026-09-28

## Identity

- PR: 4
- Work area: 1.2 Recognition test set
- Milestone: Add the recognition test set, its manifest checks, source/download/degradation tools, and catalog shortlist.
- Target: `main`
- Base: `89046b4219827aa3dbd5b505c228adcea48cf149`
- Merge base: `89046b4219827aa3dbd5b505c228adcea48cf149`
- Head: `a1bf5dbb73b6f22d72601fe40a1c113779c25c41`
- Branch: `feat/pr-3-recognition-test-set`

## Provider gate

The session record for this work in `docs/session-handoff.md` names Claude Code as the author provider. This session is Codex. The providers differ, so the gate passes.

## Intended behavior and scope

The acceptance story requires a free manifest check and a downloader that fills the ignored local cache with every manifest image whose SHA-256 agrees. The milestone also adds a 26-type catalog, 130-image manifest and 44 degraded-photo recipes, source metadata verification, tests, and updates to the roadmap and research. The relevant contracts are D-45, D-56, D-95, D-97, D-102 to D-105, and section 6.6 of `docs/research/platform-cloud-and-ai.md`.

I inspected all 16 changed paths from the base: `README.md`, `docs/decisions.md`, `docs/questions.md`, `docs/research/platform-cloud-and-ai.md`, `docs/reviews/pr-4-response.md`, `docs/reviews/pr-4.md`, `docs/roadmaps/phase-1-risk-spikes.md`, `docs/session-handoff.md`, and the eight files in `tools/spikes/recognition_set/`. The prior review examined the original 14 paths. This review examined the correction and its consumers. Git tracks no image files.

## Findings

### P2-1: Source verification ignores recorded source and license URLs

Status: fixed in `a1bf5dbb73b6f22d72601fe40a1c113779c25c41`.

Open at: `81f74d624feebc88bd20f602a707ff45f9735c69`.

File: `tools/spikes/recognition_set/sources.py:174-180`.

Trigger: A manifest entry has a wrong `source_url` or `license_url`, while its source record still returns the same license, author, and file URL.

Expected: The manifest's source URL, author, and license must agree with their source (D-97, D-102, D-105). The source check must detect a wrong Flickr photo page or license URL, and a wrong Commons license URL.

Actual: `sources.check` compares Commons `license`, `author`, `file_url`, and `bytes`, and Flickr `license`, `author`, and `file_url`. It does not compare `source_url` for either provider or `license_url` for either provider. The manifest validator checks only that a license URL contains the expected license path, not that it matches the source URL or trusted license host.

Consequence: The check can report no differences while the committed attribution points to a different source page or an unrelated license URL. This weakens the provenance evidence for the licensed test set.

Correction: Compare each recorded source URL and license URL with the source record. Reject a license URL that does not name the matching license at its trusted URL. Add regression cases for wrong Flickr and Commons URLs.

Regression check: `make verify` passed 43 recognition-set tests, including Commons and Flickr URL mismatch cases. The source check returned 0 differences.

### P2-2: License URL paths accept a prefix of the required license

Status: open.

Open at: `a1bf5dbb73b6f22d72601fe40a1c113779c25c41`.

File: `tools/spikes/recognition_set/manifest_check.py:81-86`.

Trigger: Set `license` to `cc-by-sa-2.0` and `license_url` to `https://creativecommons.org/licenses/by-sa/2.0evil/`, or set `license` to `cc0` and `license_url` to `https://creativecommons.org/publicdomain/zero/1.0-not-a-license/`.

Expected: The manifest check accepts a license URL only when its path identifies the same license as the recorded code (D-95, D-97).

Actual: `license_url_ok` uses `startswith` for versioned paths and broad directory prefixes for CC0 and public domain. The examples above return `True`. The public-domain code also accepts the CC0 URL path.

Consequence: A malformed or different license page can pass the manifest check as license proof.

Correction: Match the complete license path and version, with valid path separators and any supported jurisdiction suffix. Keep CC0 and public-domain paths distinct.

Regression check: Add false cases for the URLs above and for a public-domain code paired with a CC0 path. Run the focused recognition-set tests and `make verify`.

## Out of scope

Recognition thresholds and the paid model run belong to the later recognition-spike work area 1.3.

## PR comments

No issue comments, review comments, or review threads exist on PR #4.

## Description edits

None.

## Earlier verdicts

The verdict at `81f74d624feebc88bd20f602a707ff45f9735c69` was **Changes required.** P2-1 was open because the source check did not compare source and license URLs.

## Verification

- `git fetch origin` and `git status --short --branch`: passed. The tree was clean at `a1bf5dbb73b6f22d72601fe40a1c113779c25c41` on detached HEAD.
- `python3 docs/tools/review_gate.py --effective-head 4`: passed. It returned `a1bf5dbb73b6f22d72601fe40a1c113779c25c41`.
- `git diff --stat 89046b4219827aa3dbd5b505c228adcea48cf149...a1bf5dbb73b6f22d72601fe40a1c113779c25c41`: 16 paths, 4,630 insertions, 41 deletions.
- `make verify` at `a1bf5dbb73b6f22d72601fe40a1c113779c25c41`: passed. STE, reference, lifecycle, context, 301 document-tool tests, 44 Luna tests, and 43 recognition-set tests passed. Two image tests skipped because Pillow is not installed.
- `python3 tools/spikes/recognition_set/sources.py check`: passed with 0 differences.
- Direct `license_url_ok` probes: reproduced P2-2. The `2.0evil`, `1.0-not-a-license`, and PD-with-CC0 path cases returned `True`.
- GitHub checks at `a1bf5dbb73b6f22d72601fe40a1c113779c25c41`: `pr-contract`, `verify:lint`, and `verify:test` passed. `review-gate` failed because this verdict is not ready.
- PR comments and review threads: none. One GraphQL query result is at `/tmp/pr4-threads.json`.
- Pillow-dependent image generation: not run because Pillow is unavailable. The author reports a successful run and sample inspection in the pull request body.
- Paid targets: not run, as required by the review skill.
- Push: `45aa695f866f4eb10eef2ef965a1e02bff81f517` was the head of `origin/feat/pr-3-recognition-test-set`, verified with `gh pr view`.

## Open questions and accepted risks

Pillow is unavailable in this environment, so the degradation output was not independently inspected. The author reports a successful run and sample inspection in the pull request body.

## Verdict

**Changes required.** This verdict applies to head `a1bf5dbb73b6f22d72601fe40a1c113779c25c41`.
P2-1 is fixed. P2-2 remains open because malformed license URL paths pass the manifest check.
