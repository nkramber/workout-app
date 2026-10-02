# Pull request 24 - author response

Date: 2026-10-02. Review round 1 recorded effective head `fe99aed`, with the verdict "Changes required".

## P1-1: A seated calf raise passes a knee injury filter

**Result: full merit, after an owner decision.**

The trigger reproduces at `fe99aed`. One catalog id, `calf_raise`, covers the standing and the seated machine. The row of D-218 held the lower back and the ankle alone, so `profile.ForPlan` kept `calf_raise` for a knee injury. The seated pad loads the knee, and D-208 removes each exercise that loads an injured area.

The finding conflicted with D-218, which the owner approved with "the close calls stay out of the table". So the author asked the owner (Q-235). The owner chose to add the knee to the row (D-221).

**Correction.**

- `go/internal/domain/body_data.go`: the row of `calf_raise` holds the lower back, the knee, and the ankle (D-221).
- `docs/research/exercise-safety.md`: section 5.15 gives the knee for the seated form, and section 7 records the result.
- `docs/decisions.md` adds D-221 and marks D-218 as amended. `docs/questions.md` adds Q-235.

**Regression checks.**

- `TestCalfRaiseLeavesForAKneeInjury` of `go/internal/profile` fails at `fe99aed`, with "a knee injury keeps calf_raise". It passes after the correction.
- `make go-test` and `make emulator-test` pass.

## P2-1: Invalid list values can echo profile text

**Result: full merit.**

The trigger reproduces at `fe99aed`. `Profile.Check` put an unknown value of the experience, the template, a group, an area, or a cardio exercise into the error. A caller can send any text in these fields, and the service gave the error text to the caller. D-80 and the comment of `ProfileService` permit no profile text in an error.

**Correction.**

- `go/internal/profile/profile.go`: an unknown experience or template gives "not a value of the list" alone. A list error gives the field and the index alone.

**Regression checks.**

- `TestErrorHoldsNoText` of `go/internal/profilesvc` sends a synthetic health text in each field of a fixed list. At `fe99aed`, it fails for each of the five fields. It passes after the correction.
- `make go-test` and `make emulator-test` pass.
