package policy

import "github.com/nkramber/workout-app/go/internal/domain"

// Step is the one load step of the policy: 5 lb (D-65, D-147).
const Step = 5 * domain.Pound

// Direction tells the rounding which way a load changes. It decides a
// value halfway between two steps (D-148).
type Direction int

const (
	// Other is each change that is not an increase or a return after a
	// break. A halfway value rounds up.
	Other Direction = iota
	// Increase is a change to a heavier load. A halfway value rounds
	// down.
	Increase
	// Return is the load of a return after a break. A halfway value
	// rounds down.
	Return
)

// Round gives the nearest multiple of 5 lb (D-65). A value halfway
// between two multiples rounds down on an increase and on a return
// after a break, and up in each other case (D-148). The result can be
// 0 lb. Select then gives the lightest weight of the machine (D-149).
func Round(l domain.Load, d Direction) domain.Load {
	if l <= 0 {
		return 0
	}
	down := l - l%Step
	switch rest := l - down; {
	case rest*2 < Step:
		return down
	case rest*2 > Step:
		return down + Step
	case d == Increase || d == Return:
		return down
	default:
		return down + Step
	}
}

// Select gives the weight of the machine for a rounded load (D-149):
// the heaviest available weight that is not more than the load, or the
// lightest weight when no weight is that light. The weights are in
// ascending order, as an inventory entry holds them. Select gives 0
// when the list is empty.
func Select(rounded domain.Load, available []domain.Load) domain.Load {
	if len(available) == 0 {
		return 0
	}
	out := available[0]
	for _, w := range available {
		if w <= rounded {
			out = w
		}
	}
	return out
}

// Valid tells whether a load is a load that the policy can give on the
// machine: an available weight that Select gives for some multiple of
// 5 lb (D-65, D-149).
func Valid(l domain.Load, available []domain.Load) bool {
	if len(available) == 0 || l <= 0 {
		return false
	}
	if l == available[0] {
		return true
	}
	up := l
	if r := l % Step; r != 0 {
		up = l - r + Step
	}
	return Select(up, available) == l
}

// floor gives the multiple of 5 lb at or below a load.
func floor(l domain.Load) domain.Load {
	if l <= 0 {
		return 0
	}
	return l - l%Step
}
