package policy

import (
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
)

func TestRound(t *testing.T) {
	for _, tc := range []struct {
		load domain.Load
		dir  Direction
		want domain.Load
	}{
		{domain.Pounds(25), Other, domain.Pounds(25)},
		{domain.Pounds(43) + 2, Increase, domain.Pounds(45)}, // 43.2 lb, section 5.8
		{domain.Pounds(42), Other, domain.Pounds(40)},
		{domain.Pounds(43), Other, domain.Pounds(45)},
		// A halfway value: down on an increase and a return, up in other cases (D-148).
		{225, Increase, domain.Pounds(20)},
		{225, Return, domain.Pounds(20)},
		{225, Other, domain.Pounds(25)},
		{25, Increase, 0},
		{25, Other, domain.Pounds(5)},
		{24, Other, 0},
		{0, Other, 0},
		{-5, Other, 0},
	} {
		if got := Round(tc.load, tc.dir); got != tc.want {
			t.Errorf("Round(%s, %d) = %s, want %s", tc.load, tc.dir, got, tc.want)
		}
	}
}

func TestSelect(t *testing.T) {
	stack := []domain.Load{domain.Pounds(10), 175, domain.Pounds(25), domain.Pounds(40)}
	for _, tc := range []struct {
		rounded domain.Load
		want    domain.Load
	}{
		{domain.Pounds(25), domain.Pounds(25)},
		{domain.Pounds(20), 175}, // the heaviest weight at or below the load (D-149)
		{domain.Pounds(35), domain.Pounds(25)},
		{domain.Pounds(100), domain.Pounds(40)},
		{domain.Pounds(5), domain.Pounds(10)}, // no weight that light: the lightest weight
		{0, domain.Pounds(10)},
	} {
		if got := Select(tc.rounded, stack); got != tc.want {
			t.Errorf("Select(%s) = %s, want %s", tc.rounded, got, tc.want)
		}
	}
	if got := Select(domain.Pounds(5), nil); got != 0 {
		t.Errorf("Select with no weight = %s, want 0", got)
	}
}

func TestValid(t *testing.T) {
	stack := []domain.Load{domain.Pounds(10), domain.Pounds(12), domain.Pounds(14), 175, domain.Pounds(25)}
	for _, tc := range []struct {
		load domain.Load
		want bool
	}{
		{domain.Pounds(10), true},  // the lightest weight
		{domain.Pounds(12), false}, // Select(15 lb) gives 14 lb, so no multiple of 5 lb selects 12 lb
		{domain.Pounds(14), true},  // Select(15 lb)
		{175, true},                // Select(20 lb)
		{domain.Pounds(25), true},
		{domain.Pounds(15), false}, // not on the stack
		{0, false},
	} {
		if got := Valid(tc.load, stack); got != tc.want {
			t.Errorf("Valid(%s) = %v, want %v", tc.load, got, tc.want)
		}
	}
	if Valid(domain.Pounds(10), nil) {
		t.Error("Valid with no weight = true, want false")
	}
}
