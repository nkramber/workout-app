# Pull request 39 response

Date: 2026-10-05

This file answers the findings of round 1 of `docs/reviews/pr-39.md`, at `dcd23494a8212291ddc4788ceca8d85cb0f1de29`.

## P2-1: Stage 8 starts before the Phase 7 exit check

Result: partial merit.

Evidence: the trigger reproduced. `AGENTS.md:9` named Phase 8 as the stage, and the three iPhone checks of Phase 7 did not run. The roadmap put the checks before the work of PR-39, but no line said that the work of work area 8.1 waits for them. No line said what a failed check does.

The order itself stays. The owner put the fix and this roadmap before the check, and the check at the start of PR-39 (D-310). The owner approved that milestone with those terms before the first edit (D-12). So the finding has no merit where it asks to move the check out of PR-39.

Correction:

- `AGENTS.md` names the exit of Phase 7 as the stage, and says that no Phase 8 work starts before the iPhone check passes (D-310).
- `docs/roadmaps/phase-8-personal-use-operations.md` adds the rule to section 3, and a gate to PR-39. When a check fails, the session stops and asks the owner for the fix. The work of PR-39 starts only after a fix passes the three checks.
- `docs/roadmaps/high-level-roadmap.md` adds the same gate to the exit note of Phase 7.

Regression check: `make ste-check`, `make ref-check`, and `make verify` passed. The stage line, section 1.1, section 3, and the PR-39 gate of the roadmap now agree. Each one keeps the work of work area 8.1 after the three checks.
