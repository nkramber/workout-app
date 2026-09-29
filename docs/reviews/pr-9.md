# Pull request 9 review

Date: 2026-09-29

## Identity

- PR: 9
- Work area: Phase 2 platform skeleton, work area 2.1
- Milestone: the contract and the API skeleton
- Target: `main`
- Base: `7aa7bd42382c4978c262fabbfa9f400104e53368`
- Merge base: `7aa7bd42382c4978c262fabbfa9f400104e53368`
- Head: `3c3db409fdb93178897473a404769cec7a39bd61`
- Branch: `feat/pr-8-api-skeleton`

## Provider gate

The hand-off names Claude Code as the author provider in Session 9. The pull request body and the hand-off name that session as the author. The active environment identifies this reviewer as Codex. The providers differ, so the gate passes.

## Intended behavior and scope

The milestone adds a buf contract and a Go Connect-RPC API. The API checks Firebase tokens and uids against an allowlist. It adds CORS, a build version route, local emulators, product targets, CI jobs, and required checks. The acceptance story requires unauthenticated and uninvited calls to fail. An allowed uid returns from `GetMe`.

The roadmap assigns this work to Phase 2 work area 2.1. The pull request cites D-75, D-82, and D-126 to D-131. The matching decisions and answered questions agree with the code. D-23, D-24, and D-80 do not change in this milestone. No model call or workout prescription exists in this API.

I inspected each changed path from the merge base to the effective head: `.github/rulesets/review-gate.json`, `.github/workflows/verify.yml`, `.gitignore`, `AGENTS.md`, `Makefile`, `README.md`, `buf.gen.yaml`, `buf.yaml`, `docs/decisions.md`, `docs/design.md`, `docs/questions.md`, `docs/research/platform-cloud-and-ai.md`, `docs/session-handoff.md`, `docs/tools/test_ruleset_check.py`, `emulators/.nvmrc`, `emulators/package-lock.json`, `emulators/package.json`, `firebase.json`, `go/README.md`, `go/cmd/api/emulator_test.go`, `go/cmd/api/main.go`, `go/cmd/api/main_test.go`, `go/gen/gymroute/v1/gymroutev1connect/user_service.connect.go`, `go/gen/gymroute/v1/user_service.pb.go`, `go/go.mod`, `go/go.sum`, `go/internal/allowlist/allowlist.go`, `go/internal/allowlist/allowlist_test.go`, `go/internal/auth/auth.go`, `go/internal/auth/auth_test.go`, `go/internal/auth/cors.go`, `go/internal/envguard/envguard.go`, `go/internal/envguard/envguard_test.go`, `go/internal/usersvc/usersvc.go`, `go/internal/usersvc/usersvc_test.go`, and `proto/gymroute/v1/user_service.proto`.

## Findings

No finding.

## Out of scope

The billing account and budget alert belong to work area 2.3. Q-142 remains open for that work area.

## PR comments

No comments or review threads exist on PR #9.

## Description edits

None.

## Verification

- `make verify` at `8b7f34709e456e4de71b99f424e1abc70f63c2fa`: passed. STE, references, lifecycle, context budget, and all Python tests passed.
- `make contract` at `8b7f34709e456e4de71b99f424e1abc70f63c2fa`: passed. Buf lint, generation, and generated-code checks passed. `origin/main` has no `proto/` folder, so there was no prior contract to break.
- `make go-test` at `8b7f34709e456e4de71b99f424e1abc70f63c2fa`: passed. Formatting, module tidiness, vet, and Go unit tests passed.
- `make emulator-test` at `8b7f34709e456e4de71b99f424e1abc70f63c2fa`: passed over the Auth and Firestore emulators.
- `make ruleset-check` at `8b7f34709e456e4de71b99f424e1abc70f63c2fa`: passed. The live ruleset matches the repository file.
- GitHub checks at head `8b7f34709e456e4de71b99f424e1abc70f63c2fa`: `pr-contract`, `verify:contract`, `verify:emulator`, `verify:go`, `verify:lint`, `verify:probe`, and `verify:test` passed. `review-gate` failed because this record was not yet on the branch.
- GitHub checks at published head `dd4ff6ccf46bec19a6b2e21045bf7963d9b25838`: `pr-contract`, `review-gate`, `verify:contract`, `verify:emulator`, `verify:go`, `verify:lint`, `verify:probe`, and `verify:test` passed.
- The single GitHub query for review threads returned none. The pull request view has no comments or submitted reviews.
- Inspected all 37 changed paths from base `7aa7bd42382c4978c262fabbfa9f400104e53368` to effective head `3c3db409fdb93178897473a404769cec7a39bd61`.
- Push: `dd4ff6ccf46bec19a6b2e21045bf7963d9b25838` is the head of `origin/feat/pr-8-api-skeleton`, verified with `gh pr view`.

## Open questions and accepted risks

Q-142 stays open for work area 2.3. The review found no other open question or accepted risk.

## Verdict

**Ready for owner merge.** This verdict applies to head `3c3db409fdb93178897473a404769cec7a39bd61`.
The acceptance story holds, the provider gate passes, and each required product check passes. The review-gate check awaits this record.
