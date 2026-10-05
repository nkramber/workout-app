package policy

import (
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// The limit of the other working sets after the first set (D-306,
// D-307).
func TestFollow(t *testing.T) {
	lb := domain.Pounds
	five := stack(10, 150, 5)
	calibrated := target("biceps_curl", 3, 12, lb(20))
	calibrated.FirstSetCalibration = true
	table := target("biceps_curl", 3, 12, lb(20))
	table.Calibration = []domain.CalibrationSet{{Reps: 12, Load: lb(20)}}
	mixed := target("biceps_curl", 3, 12, lb(20))
	mixed.Working[2].Load = lb(25)
	db := dumbbellInput(t, "db_biceps_curl").Entry.Available()
	for _, tc := range []struct {
		name      string
		target    domain.PlannedExercise
		available []domain.Load
		want      domain.Load
	}{
		{"next weight of the list", target("biceps_curl", 3, 12, lb(20)), five, lb(25)},
		{"a list of 10 lb steps", target("chest_press", 3, 12, lb(30)), stack(10, 50, 10), lb(40)},
		{"a weight between two weights of the list", target("chest_press", 3, 12, lb(14)), []domain.Load{lb(10), lb(14), lb(20)}, lb(20)},
		{"the top of the list", target("biceps_curl", 3, 12, lb(150)), five, lb(150)},
		{"dumbbells", target("db_biceps_curl", 3, 12, lb(15)), db, lb(20)},
		{"the first-set calibration", calibrated, five, 0},
		{"a calibration set", table, five, 0},
		{"more than one load", mixed, five, 0},
		{"no working set", domain.PlannedExercise{Exercise: "biceps_curl"}, five, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Follow(tc.target, tc.available); got != tc.want {
				t.Errorf("Follow = %s, want %s", got, tc.want)
			}
		})
	}
}

// The load of the other working sets after the first set (D-306 to
// D-308).
func TestFollowed(t *testing.T) {
	lb := domain.Pounds
	limited := target("biceps_curl", 3, 12, lb(20))
	limited.FollowMax = lb(25)
	old := target("biceps_curl", 3, 12, lb(20))
	for _, tc := range []struct {
		name   string
		target domain.PlannedExercise
		weight domain.Load
		want   domain.Load
	}{
		{"heavier, above the limit", limited, lb(30), lb(25)},
		{"heavier, at the limit", limited, lb(25), lb(25)},
		{"the load of the target", limited, lb(20), lb(20)},
		{"lighter, no limit", limited, lb(10), lb(10)},
		{"a target of version 7", old, lb(30), lb(20)},
		{"no working set", domain.PlannedExercise{Exercise: "biceps_curl", FollowMax: lb(25)}, lb(30), 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Followed(tc.target, tc.weight); got != tc.want {
				t.Errorf("Followed = %s, want %s", got, tc.want)
			}
		})
	}
}

// The rules read the load that the other working sets followed (D-309).
// k_follow_capped is the live check of 4bbf6c8: a target of 20 lb and a
// first set of 30 lb. The other sets followed to 25 lb, the limit.
func TestScenarioFollow(t *testing.T) {
	lb := domain.Pounds
	five := stack(10, 150, 5)
	limited := func(p domain.PlannedExercise) domain.PlannedExercise {
		p.FollowMax = Follow(p, five)
		return p
	}
	curl := func(h ...Outcome) Input {
		in := machineInput(t, "biceps_curl", five)
		in.History = h
		return in
	}
	for _, tc := range []struct {
		name string
		in   Input
	}{
		{"k_follow_capped", curl(outcome(limited(target("biceps_curl", 3, 12, lb(20))), set(12, lb(30), 3), set(12, lb(25), 3), set(12, lb(25), 3)))},
		{"k_follow_heavier", curl(outcome(limited(target("biceps_curl", 3, 10, lb(20))), set(10, lb(25), 3), set(10, lb(25), 3), set(10, lb(25), 3)))},
		{"k_follow_lighter", curl(outcome(limited(target("biceps_curl", 3, 12, lb(25))), set(12, lb(15), 3), set(12, lb(15), 3), set(12, lb(15), 3)))},
		{"k_follow_short", curl(outcome(limited(target("biceps_curl", 3, 12, lb(20))), set(12, lb(25), 2), set(9, lb(25), 1), set(8, lb(25), 0)))},
		// A target of version 7 has no limit, so the rules read the
		// load of the target.
		{"k_follow_version_7", curl(outcome(target("biceps_curl", 3, 12, lb(20)), set(12, lb(30), 3), set(12, lb(20), 3), set(12, lb(20), 3)))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.in = dated(tc.in)
			d, err := Next(tc.in)
			if err != nil {
				t.Fatalf("Next: %v", err)
			}
			v, err := Check(d.Target, tc.in)
			if err != nil || len(v) > 0 {
				t.Errorf("Check of the target of Next = %v, %v, want no violation", v, err)
			}
			if want := Follow(d.Target, five); d.Target.FollowMax != want {
				t.Errorf("FollowMax = %s, want %s", d.Target.FollowMax, want)
			}
			golden(t, tc.name, render(d))
		})
	}
}

// Each record of the policy gives the limit, and a start gives none,
// because its first set is the calibration (D-301, D-306).
func TestRecordFollow(t *testing.T) {
	lb := domain.Pounds
	five := stack(10, 150, 5)
	in := machineInput(t, "biceps_curl", five)
	in.Estimate = lb(20)
	r, err := Revise(in)
	if err != nil {
		t.Fatalf("Revise of a start: %v", err)
	}
	if !r.Target.FirstSetCalibration || r.Target.FollowMax != 0 {
		t.Errorf("start: first set %v, FollowMax %s, want the calibration and no limit", r.Target.FirstSetCalibration, r.Target.FollowMax)
	}

	in = machineInput(t, "biceps_curl", five)
	in.History = []Outcome{outcome(target("biceps_curl", 3, 10, lb(20)), set(10, lb(20), 3), set(10, lb(20), 3), set(10, lb(20), 3))}
	in = dated(in)
	r, err = Revise(in)
	if err != nil {
		t.Fatalf("Revise: %v", err)
	}
	if r.Target.FollowMax != lb(25) {
		t.Errorf("Revise: FollowMax %s, want 25 lb", r.Target.FollowMax)
	}
	prop := target("biceps_curl", 3, 10, lb(20))
	prop.RestSeconds = DefaultRest(in.Exercise)
	for i := range prop.Working {
		prop.Working[i].RIR = 3
	}
	r, err = Decide(in, Proposal{Target: &prop})
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if r.Source != SourceLuna || r.Target.FollowMax != lb(25) {
		t.Errorf("Decide: source %q, FollowMax %s, want luna and 25 lb (violations %v)", r.Source, r.Target.FollowMax, r.Violations)
	}
}
