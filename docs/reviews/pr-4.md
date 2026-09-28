# Pull request 4 review

Date: 2026-09-28

## Identity

- PR: 4
- Work area: 1.2 Recognition test set
- Milestone: Add the recognition test set, its manifest checks, source/download/degradation tools, and catalog shortlist.
- Target: `main`
- Base: `89046b4219827aa3dbd5b505c228adcea48cf149`
- Merge base: `89046b4219827aa3dbd5b505c228adcea48cf149`
- Head: `d87e7ce36df325ac50745c3a777b10555f8317cc`
- Branch: `feat/pr-3-recognition-test-set`

## Provider gate

The session record for this work in `docs/session-handoff.md` names Claude Code as the author provider. This session is Codex. The providers differ, so the gate passes.

## Intended behavior and scope

The acceptance story requires a free manifest check and a downloader that fills the ignored local cache with every manifest image whose SHA-256 agrees. The milestone also adds a 26-type catalog, 130-image manifest and 44 degraded-photo recipes, source metadata verification, tests, and updates to the roadmap and research. The relevant contracts are D-45, D-56, D-95, D-97, D-102 to D-105, and section 6.6 of `docs/research/platform-cloud-and-ai.md`.

I inspected all 18 changed paths from the base: `AGENTS.md`, `README.md`, `docs/decisions.md`, `docs/design.md`, `docs/questions.md`, `docs/research/platform-cloud-and-ai.md`, `docs/reviews/pr-4-response.md`, `docs/reviews/pr-4.md`, `docs/roadmaps/phase-1-risk-spikes.md`, `docs/session-handoff.md`, and the eight files in `tools/spikes/recognition_set/`. The earlier review examined the original 16 paths. This review examined the latest corrections and their consumers. Git tracks no image files.

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

Status: fixed in `642a16a`.

Open at: `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4`.

File: `tools/spikes/recognition_set/manifest_check.py:23`.

Trigger: Set `license` to `cc-by-4.5` and `license_url` to `https://creativecommons.org/licenses/by/4.5`.

Expected: The manifest check accepts only an issued license within D-95. Creative Commons lists versions 1.0, 2.0, 2.5, 3.0, and 4.0, plus a 2.1 release for some jurisdictions ([version table](https://wiki.creativecommons.org/wiki/License_Versions), [2.1 Spain license](https://creativecommons.org/licenses/by-sa/2.1/es/), checked 2026-09-28).

Actual: `license_ok` and `license_url_ok` both return true for `cc-by-4.5`. They also accept codes such as `cc-by-sa-3.5` and `cc-by-nc-1.5`. The same check rejects a real 2.1 port, such as `cc-by-sa-2.1-es`.

Consequence: A made-up license code and matching path pass the license gate as valid proof. A valid 2.1 port fails, which can block an otherwise acceptable image.

Correction: Match the issued version numbers and valid jurisdiction ports. Add false cases for unsupported versions and true cases for valid versions, including a 2.1 port.

Regression check: `LicenseTest.test_only_issued_versions_and_ports_pass` and `LicenseTest.test_the_url_of_an_unissued_license_fails` pass. The checks accept issued version and port pairs and reject unsupported pairs, including each trigger at `c3c5c7a`.

### P2-4: Manifest attribution names conflict with the privacy rule

Status: fixed in `d87e7ce`.

Open at: `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4`.

File: `tools/spikes/recognition_set/manifest.json:817`.

Trigger: The manifest publishes an image author name that identifies a person.

Expected: The repository rule says, “Write no ... personal data ... into a file.” D-97 requires the manifest to hold each image author.

Actual: The manifest includes personal names in `author` fields. The cited example is a natural-person name, and other entries have the same form.

Consequence: The manifest follows D-97 and conflicts with the privacy rule. The owner must decide which rule controls before this review can approve the change.

Correction: The owner must define whether public attribution names are an exception. Then the author must apply that decision and update the decision register when required.

Regression check: Q-124 records the owner's answer as D-106. `AGENTS.md` and section 6 of `docs/design.md` now permit only the author credit that an image license requires and the source publishes. `make ref-check` passes.

## Out of scope

Recognition thresholds and the paid model run belong to the later recognition-spike work area 1.3.

## PR comments

No issue comments, review comments, or review threads exist on PR #4.

## Description edits

None.

## Earlier verdicts

The verdict at `81f74d624feebc88bd20f602a707ff45f9735c69` was **Changes required.** P2-1 was open because the source check did not compare source and license URLs.

The verdict at `a1bf5dbb73b6f22d72601fe40a1c113779c25c41` was **Changes required.** P2-1 was fixed. P2-2 was open because malformed license paths passed validation.

The verdict at `c3c5c7a1b4c49857bf4a6d676c8704ef2e1ed8d4` was **Blocked.** P2-1 and P2-2 were fixed. P2-3 was open, and P2-4 needed an owner decision.

## Verification

- `git fetch origin` and `git status --short --branch`: passed at start, with a clean tree at `318da6853649b6d9d3057bbb10f94bd2a91665c4` on detached HEAD.
- `python3 docs/tools/review_gate.py --effective-head 4`: passed. It returned `d87e7ce36df325ac50745c3a777b10555f8317cc`.
- `git diff --stat 89046b4219827aa3dbd5b505c228adcea48cf149...HEAD`: 18 paths, 4,830 insertions, 43 deletions.
- `git diff --check 89046b4219827aa3dbd5b505c228adcea48cf149..HEAD`: passed.
- `make verify` at the effective head: passed. STE, reference, lifecycle, context, 301 document-tool tests, 44 Luna tests, and 45 recognition-set tests passed. Two image tests skipped because Pillow was absent from the default environment.
- `PYTHONPATH=/tmp/pr4-pillow python3 -m unittest -v test_recognition_set.DegradeTest test_recognition_set.DegradeOrientationTest`: passed all 5 tests with Pillow 11.3.0 in a temporary target.
- `python3 tools/spikes/recognition_set/manifest_check.py`: passed. It reported 130 images, 44 degraded-photo recipes, and 0 errors.
- `python3 tools/spikes/recognition_set/sources.py check`: passed with 0 differences.
- `python3 -m unittest -v test_recognition_set` from the repository root: failed to import `degrade` because the test module expects its folder on the import path. The correct invocation from `tools/spikes/recognition_set/` passed all 45 tests.
- `make pr-check`: failed because this detached review worktree has no pull request association. The explicit body/title check also failed because it treated `HEAD` as the branch. `python3 docs/tools/pr_check.py pr --body-file /tmp/pr4-body.md --title "feat: the recognition test set (PR-3)" --head feat/pr-3-recognition-test-set --base origin/main`: passed with 0 contract errors.
- `python3 tools/spikes/recognition_set/download.py`: 130 images downloaded, 0 errors. Each SHA-256 agreed.
- `python3 tools/spikes/recognition_set/download.py --offline`: 130 cached images, 0 downloads, 0 errors.
- `python3 tools/spikes/recognition_set/sources.py check`: passed with 0 differences.
- P2-3 correction probes: seven issued code/port pairs pass, nine unsupported pairs fail, and license URL probes reject the prior false accepts. The official Creative Commons site lists the 2.1 Spain deed and legal code ([deed](https://creativecommons.org/licenses/by-sa/2.1/es/deed.en), [legal code](https://creativecommons.org/licenses/by-sa/2.1/es/legalcode.es)).
- GitHub checks at tip `318da6853649b6d9d3057bbb10f94bd2a91665c4`: `pr-contract`, `verify:lint`, and `verify:test` passed. `review-gate` failed because the published record still held the prior blocked verdict.
- Paid targets: not run, as required by the review skill.
- PR comments, review comments, and review threads: none. The one-query export is `/tmp/pr4-comments.json`.

## Open questions and accepted risks

None. The owner resolved the attribution question in Q-124 and D-106.

## Verdict

**Ready for owner merge.** This verdict applies to head `d87e7ce36df325ac50745c3a777b10555f8317cc`.
The earlier findings are fixed. The focused checks pass, and the owner resolved the attribution question.
