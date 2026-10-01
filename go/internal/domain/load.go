package domain

import "strconv"

// Load is a weight in tenths of a pound. Pounds are the one unit (D-28,
// D-122). An integer count keeps 12.5 lb and 2.5 lb exact, so no float
// rounding changes a comparison of two loads (D-160). A dumbbell load
// is the load of one dumbbell (D-155).
type Load int64

const (
	// Tenth is 0.1 lb.
	Tenth Load = 1
	// Pound is 1 lb.
	Pound Load = 10
)

// Pounds gives the load of a whole number of pounds.
func Pounds(lb int64) Load { return Load(lb) * Pound }

// String gives the load in pounds, such as "12.5 lb" or "40 lb".
func (l Load) String() string {
	whole, tenth := int64(l)/10, int64(l)%10
	if tenth < 0 {
		tenth = -tenth
	}
	s := strconv.FormatInt(whole, 10)
	if l < 0 && whole == 0 {
		s = "-0"
	}
	if tenth != 0 {
		s += "." + strconv.FormatInt(tenth, 10)
	}
	return s + " lb"
}

// Check refuses a load of 0 or less. A weight on a stack, a dumbbell,
// a target, and a logged weight are each more than 0.
func (l Load) Check() error {
	if l <= 0 {
		return invalid("load %s: want more than 0", l)
	}
	return nil
}
