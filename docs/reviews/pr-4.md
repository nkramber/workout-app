# Pull request 4 review

Date: 2026-09-28

## Identity

- PR: 4
- Work area: 1.2 Recognition test set
- Milestone: Add the recognition test set, its manifest checks, source/download/degradation tools, and catalog shortlist.
- Target: `main`
- Base: `89046b4219827aa3dbd5b505c228adcea48cf149`
- Merge base: `89046b4219827aa3dbd5b505c228adcea48cf149`
- Head: `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4`
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

Status: fixed in `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4`.

Open at: `a1bf5dbb73b6f22d72601fe40a1c113779c25c41`.

File: `tools/spikes/recognition_set/manifest_check.py:81-86`.

Trigger: Set `license` to `cc-by-sa-2.0` and `license_url` to `https://creativecommons.org/licenses/by-sa/2.0evil/`, or set `license` to `cc0` and `license_url` to `https://creativecommons.org/publicdomain/zero/1.0-not-a-license/`.

Expected: The manifest check accepts a license URL only when its path identifies the same license as the recorded code (D-95, D-97).

Actual: `license_url_ok` uses `startswith` for versioned paths and broad directory prefixes for CC0 and public domain. The examples above return `True`. The public-domain code also accepts the CC0 URL path.

Consequence: A malformed or different license page can pass the manifest check as license proof.

Correction: Match the complete license path and version, with valid path separators and any supported jurisdiction suffix. Keep CC0 and public-domain paths distinct.

Regression check: The focused tests and `make verify` pass at `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4`. Direct probes reject both trigger URLs and a public-domain code paired with a CC0 path.

### P2-3: License validation accepts unpublished version numbers

Status: open.

Open at: `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4`.

File: `tools/spikes/recognition_set/manifest_check.py:23`.

Trigger: Set `license` to `cc-by-4.5` and `license_url` to `https://creativecommons.org/licenses/by/4.5`.

Expected: The manifest check accepts only an issued license within D-95. Creative Commons lists versions 1.0, 2.0, 2.5, 3.0, and 4.0, plus a 2.1 release for some jurisdictions ([version table](https://wiki.creativecommons.org/wiki/License_Versions), [2.1 Spain license](https://creativecommons.org/licenses/by-sa/2.1/es/), checked 2026-09-28).

Actual: `license_ok` and `license_url_ok` both return true for `cc-by-4.5`. They also accept codes such as `cc-by-sa-3.5` and `cc-by-nc-1.5`. The same check rejects a real 2.1 port, such as `cc-by-sa-2.1-es`.

Consequence: A made-up license code and matching path pass the license gate as valid proof. A valid 2.1 port fails, which can block an otherwise acceptable image.

Correction: Match the issued version numbers and valid jurisdiction ports. Add false cases for unsupported versions and true cases for valid versions, including a 2.1 port.

Regression check: Direct probes reproduce the false accepts and the 2.1 false reject at `c3c5c7a`. The existing 43 recognition-set tests and `make verify` pass, but neither tests these versions.

### P2-4: Manifest attribution names conflict with the privacy rule

Status: open.

Open at: `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4`.

File: `tools/spikes/recognition_set/manifest.json:817`.

Trigger: The manifest publishes an image author name that identifies a person.

Expected: The repository rule says, “Write no ... personal data ... into a file.” D-97 requires the manifest to hold each image author.

Actual: The manifest includes personal names in `author` fields. The cited example is a natural-person name, and other entries have the same form.

Consequence: The manifest follows D-97 and conflicts with the privacy rule. The owner must decide which rule controls before this review can approve the change.

Correction: The owner must define whether public attribution names are an exception. Then the author must apply that decision and update the decision register when required.

Regression check: Read the manifest entry at line 817 and D-97. The conflict reproduces in the committed files.

## Out of scope

Recognition thresholds and the paid model run belong to the later recognition-spike work area 1.3.

## PR comments

No issue comments, review comments, or review threads exist on PR #4.

## Description edits

None.

## Earlier verdicts

The verdict at `81f74d624feebc88bd20f602a707ff45f9735c69` was **Changes required.** P2-1 was open because the source check did not compare source and license URLs.

The verdict at `a1bf5dbb73b6f22d72601fe40a1c113779c25c41` was **Changes required.** P2-1 was fixed. P2-2 was open because malformed license paths passed validation.

## Verification

- `git fetch origin` and `git status --short --branch`: passed. The tree was clean at `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4` on detached HEAD.
- `python3 docs/tools/review_gate.py --effective-head 4`: passed. It returned `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4`.
- `git diff --stat 89046b4219827aa3dbd5b505c228adcea48cf149...c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4`: 16 paths, 4,693 insertions, 42 deletions.
- `git diff --check 89046b4219827aa3dbd5b505c228adcea48cf149...HEAD`: passed.
- `make verify` at `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4`: passed. STE, reference, lifecycle, context, 301 document-tool tests, 44 Luna tests, and 43 recognition-set tests passed. Two image tests skipped because Pillow is not installed.
- `python3 tools/spikes/recognition_set/download.py`: 130 images downloaded, 0 errors. Each SHA-256 agreed.
- `python3 tools/spikes/recognition_set/download.py --offline`: 130 cached images, 0 downloads, 0 errors.
- `python3 tools/spikes/recognition_set/sources.py check`: passed with 0 differences.
- Direct license URL probes: P2-2's malformed paths fail. P2-3's `cc-by-4.5`, `cc-by-sa-3.5`, and `cc-by-nc-1.5` codes and matching URLs pass incorrectly. `cc-by-sa-2.1-es` fails despite its issued license page.
- GitHub checks at `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4`: `pr-contract`, `verify:lint`, and `verify:test` passed. `review-gate` failed because the record held a non-ready verdict.
- PR comments, reviews, and review threads: none. The one-query export is `/tmp/pr4-comments.json`.
- Pillow-dependent image generation: not run because Pillow is unavailable. The author reports a successful run and sample inspection in the pull request body.
- Paid targets: not run, as required by the review skill.
- Push: `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4` was the head of `origin/feat/pr-3-recognition-test-set`, verified with `gh pr view` before this review commit.

## Open questions and accepted risks

Pillow is unavailable in this environment, so the degradation output was not independently inspected. The author reports a successful run and sample inspection in the pull request body.

The owner must resolve this rule conflict: “Write no ... personal data” in `AGENTS.md`, and D-97's requirement to record each image author. The reviewer asked whether public attribution names are an exception.

## Verdict

**Blocked.** This verdict applies to head `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4`.
P2-3 is open because the manifest check accepts unpublished license versions. P2-4 needs an owner decision because the author-attribution requirement conflicts with the privacy rule. The review cannot approve this head until these issues are resolved.
