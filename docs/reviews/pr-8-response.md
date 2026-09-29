# Pull request 8 - author response

Date: 2026-09-29. Review round 1 recorded effective head `f1a9321`, with the verdict "Changes required".

## P2-1: The public records disclose owner billing details

**Result: full merit.**

The trigger reproduces at `f1a9321`. Two rows gave the trial state of the billing account of the owner, its credit, and its time left. One row was Q-142 of `docs/questions.md`, and one row was in the facts table of `docs/session-handoff.md`. `AGENTS.md` forbids personal data in this public repository. The roadmap does not need these details, because PR-10 reads the billing state before it asks Q-142.

**Correction.**

- `docs/questions.md`: the Q-142 answer says only that the owner asked for a check of the billing account first. The session of PR-10 reads the billing state before it asks. Q-142 stays open.
- `docs/session-handoff.md`: the author removed the row of the billing statement.
- `docs/roadmaps/phase-2-platform-skeleton.md`: the PR-10 text no longer names a free trial. It says that the owner reads the credits of the account in the console.

The earlier commit `f1a9321` stays in the history of the branch. The squash merge puts only the corrected text on `main`.

**Regression checks.**

- A search of `docs/questions.md`, `docs/session-handoff.md`, and `docs/roadmaps/` for "free trial", "280 USD", and "70 days" finds nothing.
- `make ste-check`, `make ref-check`, and `make verify` pass.
