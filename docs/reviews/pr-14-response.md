# Pull request 14 - author response

Date: 2026-09-30. Review round 1 recorded effective head `6da346c`, with the verdict "Changes required".

## P2-1: Exported kind and region lists can change validation

**Result: full merit.**

The trigger reproduces at `6da346c`. `Kind.known` and `Region.known` read the exported slices `Kinds` and `Regions`. A caller that appended "barbell" to `Kinds` made `Catalog.Check` accept a machine of that kind, against D-155 and D-161.

**Correction.**

- `go/internal/domain/catalog.go`: `known` uses a fixed `switch` over the four kinds and the six regions.
- `Kinds()` and `Regions()` are functions, and each call gives a new list. A change of a list changes no check and no later list.

**Regression checks.**

- `TestKindAndRegionListsAreCopies` changes and extends the lists, then checks a catalog with the kind "barbell" and a catalog with the region "arms". `Check` refuses each one.
- The test fails on the code of `6da346c`: `Check() of kind barbell = <nil>, want ErrInvalid`.

## P2-2: A valid extreme dumbbell range can make weight expansion loop forever

**Result: full merit.**

The trigger reproduces at `6da346c`. `DumbbellSet{Lightest: 1, Heaviest: math.MaxInt64, Step: 1}` passed `Check`, and `Weights` did not end, because `w += d.Step` overflowed to a negative value. A loop that stops before the overflow is not enough, because that range holds about 9 x 10^18 weights.

No decision gave an upper bound, so the author asked the owner (Q-179). The owner set the heaviest dumbbell of a set to 100 lb at most (D-166).

**Correction.**

- `go/internal/domain/inventory.go`: the constant `DumbbellMax` is 100 lb, and `DumbbellSet.Check` refuses a heaviest weight above it. The step is 0.1 lb or more, so a set holds 1,000 weights or fewer, and the loop of `Weights` can not overflow.
- `docs/decisions.md`, `docs/questions.md`, `docs/design.md`, and `docs/roadmaps/phase-3-workout-domain.md` record D-166.

**Regression checks.**

- `TestDumbbellSetCheck` refuses a heaviest weight of 100.5 lb and a range up to `math.MaxInt64`, and accepts 100 lb.
- `TestDumbbellSetWeights` gives nil at once for the range up to `math.MaxInt64`, and 1,000 weights for 0.1 lb to 100 lb.
- On the code of `6da346c`, the `Check` cases fail, and `TestDumbbellSetWeights` stops at the 30 s limit of `go test -timeout 30s`.
- `make go-test` and `make verify` pass, and the coverage of `go/internal/domain` stays 100%.
