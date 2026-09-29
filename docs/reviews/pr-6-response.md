# Pull request 6 - author response

Date: 2026-09-29. Review round 2 recorded head `2c20c66`, with the verdict "Changes required".

## P2-1: AGENTS.md says no service exists while Firebase services are active

**Result: full merit.**

The trigger reproduces at `2c20c66`. Line 9 of `AGENTS.md` said "No app or service exists yet", and the next sentence said that the project has Firebase Hosting and Firebase Authentication. Line 9 of `README.md` had the same conflict. A search of the other documents found no other copy of the text.

**Correction.**

- `AGENTS.md`: the stage line says that the workout app and its backend do not exist yet. Then it says that the project `gym-route-dev` exists, with Firebase Hosting and Firebase Authentication only (D-99, D-116).
- `README.md`: the same correction.

**Regression checks.**

- `git grep` for "no app or service", "no service exists", and "cloud resource exists" in the documents outside `docs/reviews/`: no match.
- `make ref-check` and `make verify` pass.
