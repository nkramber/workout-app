package revise

import (
	"testing"

	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// TestCheck: a reason passes when it names a logged set of the last
// session, and each number of its text is in the evidence (D-68,
// D-288). Scenario A: 3 x 12 at 25 lb, logged 12, 12, 5, with pain 3
// on set 2. The reason of the rules names set 3.
func TestCheck(t *testing.T) {
	seen := target("chest_press", 3, 12, lb(25))
	pain := domain.Pain(3)
	sets := []domain.SetLog{
		{Kind: domain.SetCalibration, Reps: 12, Weight: lb(20), RIR: 4},
		set(12, lb(25), 1), {Kind: domain.SetWorking, Reps: 12, Weight: lb(25), RIR: 1, Pain: &pain}, set(5, lb(25), 0),
	}
	e, _ := domain.DefaultCatalog().Exercise("chest_press")
	entry := domain.InventoryEntry{Machine: e.Machine, Weights: []domain.Load{lb(20), lb(25), lb(30)}}
	in := policy.Input{Exercise: e, Entry: entry, Today: "2026-10-02",
		History: []policy.Outcome{{Date: "2026-10-02", Target: seen, Log: domain.ExerciseLog{Exercise: e.ID, Sets: sets}}}}
	rec, err := policy.Revise(in)
	if err != nil {
		t.Fatal(err)
	}
	w := func(n ...int) []ai.SetRef {
		var out []ai.SetRef
		for _, x := range n {
			out = append(out, ai.SetRef{Kind: domain.SetWorking, Number: x})
		}
		return out
	}
	for _, tc := range []struct {
		name  string
		r     ai.Reason
		cause string
	}{
		{"set number", ai.Reason{Sets: w(3), Text: "Set 3 ended 7 reps short, so the target stays at 25 lb."}, ""},
		{"set values", ai.Reason{Sets: w(1), Text: "You did 12 reps at 25 lb with 1 rep in reserve."}, ""},
		{"pain", ai.Reason{Sets: w(2), Text: "You reported pain 3 on set 2."}, ""},
		{"calibration set", ai.Reason{Sets: []ai.SetRef{{Kind: domain.SetCalibration, Number: 1}}, Text: "Calibration set 1 had 4 in reserve at 20 lb."}, ""},
		{"blocked", ai.Reason{Sets: w(3), Blocked: ai.FilterMedical}, "blocked-text.medical"},
		{"no set", ai.Reason{Text: "Good work on set 3."}, CauseNoSet},
		{"unknown set", ai.Reason{Sets: w(4), Text: "Set 4 ended short."}, CauseUnknownSet},
		{"unknown calibration set", ai.Reason{Sets: []ai.SetRef{{Kind: domain.SetCalibration, Number: 2}}, Text: "Set 2."}, CauseUnknownSet},
		{"unnamed set", ai.Reason{Sets: w(3), Text: "The target stays the same."}, CauseUnnamedSet},
		{"another set named", ai.Reason{Sets: w(3), Text: "Set 1 was hard."}, CauseUnnamedSet},
		{"number not logged", ai.Reason{Sets: w(3), Text: "Set 3 ended at 6 reps."}, CauseNumber},
		{"load jump", ai.Reason{Sets: w(3), Text: "Set 3 was easy, so the load goes up to 37.5 lb."}, CauseNumber},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.r.Exercise = e.ID
			if got := Check(in, rec, tc.r); got != tc.cause {
				t.Errorf("Check = %q, want %q", got, tc.cause)
			}
		})
	}
	if got := Check(policy.Input{Exercise: e}, rec, ai.Reason{Sets: w(1), Text: "Set 1."}); got != CauseUnknownSet {
		t.Errorf("no history: %q", got)
	}
}
