# Roadmaps

This folder holds the roadmaps of Workout App. The date of this version is 2026-10-02.

## Files

| File | Purpose |
|---|---|
| `docs/roadmaps/high-level-roadmap.md` | The one high-level roadmap: phases, order, work areas, exit evidence |
| `docs/roadmaps/README.md` | This file: the rules for all roadmaps |
| `docs/roadmaps/phase-1-risk-spikes.md` | The focused roadmap of Phase 1: PR-1 to PR-6 |
| `docs/roadmaps/phase-2-platform-skeleton.md` | The focused roadmap of Phase 2: PR-7 to PR-11 |
| `docs/roadmaps/phase-3-workout-domain.md` | The focused roadmap of Phase 3: PR-12 to PR-17 |
| `docs/roadmaps/phase-4-equipment-inventory.md` | The focused roadmap of Phase 4: PR-18 to PR-21 |

A focused roadmap for one phase gets a file named `phase-<n>-<short-name>.md` in this folder.

## Rules for all roadmaps

- Write in ASD-STE100 (D-83). `make ste-check` checks each file.
- Cite each decision as a D- id and each question as a Q- id. `make ref-check` fails on an id that no register defines.
- The high-level roadmap names work areas with no pull request ids (D-9). A focused roadmap gives each pull request a PR-<n> id. One sequence of ids runs across all focused roadmaps, from PR-1 (D-86, D-96).
- Name each pull request `<type>: <summary> (PR-<n>)` on the branch `<type>/pr-<n>-<slug>` (D-86).
- A work area holds one milestone with one acceptance story. It can hold two, three, or more concerns (D-10, D-12).
- Every work area names its exit evidence. The evidence is a check, a test, a report, or a device result that a reader can see.
- Mark each paid check. A paid check runs only after the owner approves it (D-25).
- Label each statement that is not an owner decision as a recommendation, an assumption, or an open question.

## How a focused roadmap starts

1. Start a clean session for the focused roadmap. The owner approves the work first (D-12).
2. Read `docs/session-handoff.md`, the phase in the high-level roadmap, and each cited decision and question.
3. Ask the owner each open question that the phase names. Record each answer in `docs/decisions.md`.
4. Split each work area into pull requests of one milestone each.
5. Give each pull request its concerns, its acceptance story, its checks, and its documentation impact.
6. Open one pull request with the focused roadmap and the hand-off. It can use the `review-override` label (D-15, D-125).

## When a roadmap changes

A focused roadmap can split or merge work areas of its phase. A change of phase order, scope, or exit evidence changes the high-level roadmap too. That change needs an owner decision in `docs/decisions.md`.

The high-level roadmap is complete when Phase 0 merges and the owner confirms the phase order. After that, the high-level roadmap changes only through an owner decision.
