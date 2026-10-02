# Workout App - Phase 4 focused roadmap: equipment inventory

This roadmap splits Phase 4 of `docs/roadmaps/high-level-roadmap.md` into pull requests. `docs/roadmaps/README.md` gives the rules. The high-level roadmap keeps the objective, the order, and the exit evidence of the phase.

The date of this version is 2026-10-02.

## 1. Scope and start state

Phase 4 turns manual selections and text entries into confirmed machines (D-49, D-51, D-55). At the end of the phase, the owner builds the one active inventory in the live app (D-46). Each machine holds its identity, its available weights, and the load estimates of the owner (D-41, D-54, D-192).

Phase 4 has no photo recognition and no AI call (D-110). So the phase has no paid step.

The exit of Phase 3 holds. `docs/research/phase-3-check.md` holds the Luna evaluation. The deploy of `9d0e5d0` passed on 2026-10-02: the build `deploy-api` `46b8f059` gave SUCCESS, and the live `/version` names `9d0e5d0`.

Phase 3 gives these inputs:

- `go/internal/domain` holds the catalog of D-155, the inventory entry, and the dumbbell set. An entry holds the catalog id and the weights alone (D-54).
- `go/internal/policy` reads the estimate of one exercise and the inventory entry of its machine.
- `go/internal/ai` holds the cap hook of D-25. It keeps the spend in memory and reads the cap values from the configuration.

Phase 2 gives the contract in `proto/`, the API in `go/`, the web shell in `web/`, and the deploy from `main` (D-137). The web app never reads or writes Firestore itself (D-77).

## 2. Owner answers for this phase

| Question | Answer | Decision |
|---|---|---|
| Q-91, kilogram markings | A machine with kilogram markings alone is out of scope. The app makes no conversion. | D-122 |
| Q-98, the monthly AI caps | 1 USD for the user and 2 USD for the project, for each month. | D-188 |
| Q-203, the lasting cap store | Phase 5, in work area 5.2, before the first live call of the planner. | D-189 |
| Q-204, the cap period | The calendar month in UTC. | D-190 |
| Q-205, text entry | The text searches the names of the catalog. A text with no match stays as a note that no plan uses. | D-191 |
| Q-206, the load estimate | One optional estimate for each exercise, so a machine with two exercises gets two. | D-192 |
| Q-207, the confirmation | A draft, then a confirmation on a review screen. A change of the weights makes a draft again. | D-193 |
| Q-208, the split | PR-19 for the API and the store, then PR-20 for the screens. | D-194 |
| Q-209, the weights of a stack | A range and a step, then a change of single weights. | D-195 |
| Q-210, the write path | A direct call to the API for each change. Phase 6 adds the outbox. | D-196 |
| Q-211, the inventory path | `users/{uid}/inventory/active`, one document. | D-197 |
| Q-212, the highest estimate | An estimate in the range of the weights of the machine. | D-198 |
| Q-213, the bounds | 1,000 lb and 200 weights for a list. 200 characters and 50 notes for the notes. | D-199 |
| Q-214, a second save | The save replaces the entry. A change of the estimates alone keeps the confirmation. | D-200 |
| Q-215, the confirmation | The weights that the review screen showed. The server refuses other weights. | D-201 |
| Q-216, the catalog list | By kind: the machines, the cable station, the dumbbells, and the cardio machines. | D-202 |
| Q-217, the live check | After the merge of PR-20 and its web deploy. The Phase 5 roadmap session records it. | D-203 |

No question of the phase stays open. PR-20 asked Q-216 and Q-217 (D-202, D-203). Q-194 and Q-202 stay open for Phase 7. Q-103 stays open for the deferred photo work.

The answers of Q-98, Q-203, and Q-204 do not change Phase 4. They change work areas 5.2 and 8.2 of the high-level roadmap (D-189). The pull request of Phase 5 that adds the lasting cap store also changes the comments of `go/internal/ai/cost.go` and the cap text of `go/README.md`.

## 3. Rules for each pull request of this phase

- The API alone reads and writes the inventory in Firestore (D-77). The rules of `firestore.rules` refuse each client read and write.
- The catalog of `go/internal/domain` is the one source of the catalog. The web app reads it through the API (recommendation).
- A stored machine holds its catalog id, its weights, the estimates of its exercises, and its state alone (D-54, D-192, D-193). A note holds its text alone (D-191).
- Each load is in pounds (D-122). An estimate is a load of more than 0 (recommendation).
- After a change in `proto/`, run `make proto`, and commit the generated code. `make contract` runs `buf breaking` against `main`.
- Logs, metrics, and error reports hold ids only (D-80). A note text never goes into a log.
- Tests use synthetic data alone. The repository is public.
- Each pull request with code gets the review of the other provider (D-15).
- A merge that changes `go/` deploys the API, and a merge that changes `web/` deploys the web app (D-137). The next session reads each deploy first.

## 4. Pull requests

The PR-<n> order is the order of work. Each pull request needs the one before it on `main`.

| Id | Work area | Title | Paid step |
|---|---|---|---|
| PR-18 | all | `docs: the Phase 4 focused roadmap (PR-18)` | none |
| PR-19 | 4.2 | `feat: the inventory API and store (PR-19)` | none |
| PR-20 | 4.1 | `feat: the inventory screens (PR-20)` | none |

### PR-18 - The Phase 4 focused roadmap

Branch: `docs/pr-18-phase-4-roadmap`.

Concerns:

- this file,
- the owner answers of section 2 (D-188 to D-196), with the change of work areas 5.2 and 8.2 of D-189 in `docs/roadmaps/high-level-roadmap.md`,
- the equipment rules of D-191 to D-193 and D-195 in `docs/design.md`,
- the row of this file in `docs/roadmaps/README.md`, and the Phase 4 stage in `AGENTS.md`.

Acceptance story: the owner reads this roadmap. Each pull request of it names its work area, concerns, acceptance story, checks, and the questions it needs. No open question blocks PR-19.

Checks: `make verify` and `make pr-check`, free. Codex reviews PR-18, because it sets the AI caps of D-25.

### PR-19 - The inventory API and store

Branch: `feat/pr-19-inventory-api`. Work area 4.2. It needs PR-18 on `main`.

Concerns:

- an inventory service in `proto/workoutapp/v1`. Its calls read the catalog and the inventory, and save, confirm, and remove a machine or a note,
- the Firestore store of the one active inventory of the owner (D-46),
- the checks of the server: each entry against the catalog, and one entry for each machine,
- a draft state after a change of the weights of a confirmed machine (D-193),
- a Go function that gives the confirmed machines alone to the plan input of Phase 5 (D-49, D-193).

Acceptance story: the emulator tests save, confirm, change, and remove machines and notes through the API. A stored machine holds only its identity, its weights, its estimates, and its state. The function for the plan input gives the confirmed machines alone. A draft and a note never reach it.

Checks: `make contract`, `make go-test`, `make emulator-test`, and `make verify`, free. Codex reviews PR-19.

Questions for the session: the Firestore path of the inventory, and the highest estimate that the server accepts. The owner answered them as Q-211 and Q-212. The session also asked Q-213 to Q-215 (D-197 to D-201).

### PR-20 - The inventory screens

Branch: `feat/pr-20-inventory-screens`. Work area 4.1. It needs PR-19 on `main`.

Concerns:

- the inventory screen, with each machine and its state, and each note,
- the selection of a machine from the catalog of `GetCatalog` (D-51),
- the text entry: a search of the catalog names, then a match or a note (D-55, D-191),
- the weights: a range and a step, then a change of single weights (D-195). The dumbbells use the dumbbell set, and a cardio machine has no weights,
- the optional estimate of each exercise (D-192),
- the review screen, the confirmation with the shown weights, and the removal of a machine (D-193, D-201),
- a direct call to the API for each change, and an error with no connection (D-196).

Acceptance story: the browser tests add a machine by selection and a machine by text entry, and confirm both. A change of the weights makes a confirmed machine a draft again. A text with no match stays as a note. The owner adds and confirms one machine in the live app on the iPhone.

A deploy comes from `main` alone (D-14). So the browser tests are the evidence of the merge, and the owner does the iPhone step after the `deploy-web` build of the merge (D-203).

Checks: `make web`, `make verify`, and the Go checks of PR-19 when `go/` changes, free. Codex reviews PR-20.

Questions for the session: the order of the catalog list, by region or by kind. The owner answered by kind (Q-216, D-202). The session also asked Q-217 (D-203).

## 5. Exit of the phase

Phase 4 ends when PR-20 merges. The emulator tests of PR-19 and the browser tests of PR-20 pass in CI. After the deploy of the merge, the owner confirms one machine in the live app on the iPhone. The Phase 5 roadmap session records the result (D-203). The function of PR-19 gives the confirmed machines alone. So a plan of Phase 5 can not use a machine that the owner did not confirm.
