package policy

import (
	"slices"
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// mirrorCase is one case of the mirror (D-341 to D-346). The phone test
// "gives the targets of the mirror as the policy does" of
// `web/src/lib/workout.test.ts` holds the same cases.
type mirrorCase struct {
	name               string
	calibration        bool
	followMax          domain.Load
	minReps, maxReps   int
	logs               []domain.SetLog
	shown, read, start []domain.WorkingSet
}

func ws(reps int, load domain.Load) domain.WorkingSet {
	return domain.WorkingSet{Reps: reps, Load: load, RIR: 2}
}

func mirrorCases() []mirrorCase {
	lb := domain.Pounds
	t := ws(10, lb(100))
	return []mirrorCase{
		{name: "the first set goes up", followMax: lb(105), minReps: 8, maxReps: 12,
			logs:  []domain.SetLog{set(12, lb(105), 2), set(12, lb(105), 2)},
			shown: []domain.WorkingSet{t, ws(12, lb(105)), ws(12, lb(105))},
			read:  []domain.WorkingSet{ws(12, lb(105)), ws(12, lb(105)), ws(12, lb(105))},
			start: []domain.WorkingSet{ws(12, lb(105)), ws(12, lb(105)), ws(12, lb(105))}},
		{name: "0 reps in reserve changes nothing", followMax: lb(105), minReps: 8, maxReps: 12,
			logs:  []domain.SetLog{set(12, lb(105), 0)},
			shown: []domain.WorkingSet{t, t, t}, read: []domain.WorkingSet{t, t, t}, start: []domain.WorkingSet{t, t, t}},
		{name: "a later set goes up", followMax: lb(105), minReps: 8, maxReps: 12,
			logs:  []domain.SetLog{set(10, lb(100), 3), set(11, lb(100), 2), set(11, lb(100), 2)},
			shown: []domain.WorkingSet{t, t, ws(11, lb(100))},
			read:  []domain.WorkingSet{t, ws(11, lb(100)), ws(11, lb(100))},
			start: []domain.WorkingSet{ws(11, lb(100)), ws(11, lb(100)), ws(11, lb(100))}},
		{name: "heavier with fewer reps gives the whole set", followMax: lb(105), minReps: 8, maxReps: 12,
			logs:  []domain.SetLog{set(8, lb(105), 2)},
			shown: []domain.WorkingSet{t, ws(8, lb(105)), ws(8, lb(105))},
			read:  []domain.WorkingSet{ws(8, lb(105)), ws(8, lb(105)), ws(8, lb(105))},
			start: []domain.WorkingSet{ws(8, lb(105)), ws(8, lb(105)), ws(8, lb(105))}},
		{name: "the limits apply", followMax: lb(105), minReps: 8, maxReps: 12,
			logs:  []domain.SetLog{set(5, lb(110), 1), set(16, lb(105), 2)},
			shown: []domain.WorkingSet{t, ws(8, lb(105)), ws(12, lb(105))},
			read:  []domain.WorkingSet{ws(8, lb(105)), ws(12, lb(105)), ws(12, lb(105))},
			start: []domain.WorkingSet{ws(12, lb(105)), ws(12, lb(105)), ws(12, lb(105))}},
		{name: "a lighter set at 0 reps in reserve", followMax: lb(105), minReps: 8, maxReps: 12,
			logs:  []domain.SetLog{set(10, lb(95), 0)},
			shown: []domain.WorkingSet{t, ws(10, lb(95)), ws(10, lb(95))},
			read:  []domain.WorkingSet{ws(10, lb(95)), ws(10, lb(95)), ws(10, lb(95))},
			start: []domain.WorkingSet{ws(10, lb(95)), ws(10, lb(95)), ws(10, lb(95))}},
		{name: "pain blocks the mirror", followMax: lb(105), minReps: 8, maxReps: 12,
			logs:  []domain.SetLog{withPain(set(12, lb(105), 2), 3)},
			shown: []domain.WorkingSet{t, t, t}, read: []domain.WorkingSet{t, t, t}, start: []domain.WorkingSet{t, t, t}},
		{name: "the first-set calibration mirrors the reps alone", calibration: true, minReps: 8, maxReps: 12,
			logs:  []domain.SetLog{set(10, lb(90), 3), set(12, lb(110), 2)},
			shown: []domain.WorkingSet{t, ws(10, lb(90)), ws(12, lb(90))},
			read:  []domain.WorkingSet{ws(10, lb(90)), ws(12, lb(90)), ws(12, lb(90))},
			start: []domain.WorkingSet{ws(12, lb(90)), ws(12, lb(90)), ws(12, lb(90))}},
		{name: "version 9 reads the first set alone", followMax: lb(105),
			logs:  []domain.SetLog{set(10, lb(100), 3), set(12, lb(105), 2)},
			shown: []domain.WorkingSet{t, t, t}, read: []domain.WorkingSet{t, t, t}, start: []domain.WorkingSet{t, t, t}},
		{name: "version 9 follows a heavier first set at 0 reps in reserve", followMax: lb(105),
			logs:  []domain.SetLog{set(10, lb(110), 0)},
			shown: []domain.WorkingSet{t, ws(10, lb(105)), ws(10, lb(105))},
			read:  []domain.WorkingSet{ws(10, lb(105)), ws(10, lb(105)), ws(10, lb(105))},
			start: []domain.WorkingSet{ws(10, lb(105)), ws(10, lb(105)), ws(10, lb(105))}},
	}
}

func TestMirrored(t *testing.T) {
	for _, tc := range mirrorCases() {
		t.Run(tc.name, func(t *testing.T) {
			p := target("biceps_curl", 3, 10, domain.Pounds(100))
			p.FirstSetCalibration, p.FollowMax, p.FollowMinReps, p.FollowMaxReps = tc.calibration, tc.followMax, tc.minReps, tc.maxReps
			shown, read, start := Mirrored(p, tc.logs)
			if !slices.Equal(shown, tc.shown) {
				t.Errorf("shown = %v, want %v", shown, tc.shown)
			}
			if !slices.Equal(read, tc.read) {
				t.Errorf("read = %v, want %v", read, tc.read)
			}
			if !slices.Equal(start, tc.start) {
				t.Errorf("start = %v, want %v", start, tc.start)
			}
		})
	}
}

// Limits gives the rep range of the exercise as the rep limits, also
// for the first-set calibration, and none for a calibration set.
func TestLimits(t *testing.T) {
	five := stack(10, 150, 5)
	in := machineInput(t, "biceps_curl", five)
	p := target("biceps_curl", 3, 10, domain.Pounds(100))
	Limits(&p, in.Exercise, five)
	if p.FollowMax != domain.Pounds(105) || p.FollowMinReps != 8 || p.FollowMaxReps != 12 {
		t.Errorf("Limits = %s, %d to %d, want 105 lb, 8 to 12", p.FollowMax, p.FollowMinReps, p.FollowMaxReps)
	}
	p.FirstSetCalibration = true
	Limits(&p, in.Exercise, five)
	if p.FollowMax != 0 || p.FollowMinReps != 8 || p.FollowMaxReps != 12 {
		t.Errorf("Limits of the first-set calibration = %s, %d to %d, want 0, 8 to 12", p.FollowMax, p.FollowMinReps, p.FollowMaxReps)
	}
	p.FirstSetCalibration = false
	p.Calibration = []domain.CalibrationSet{{Reps: 5, Load: domain.Pounds(50)}}
	Limits(&p, in.Exercise, five)
	if p.FollowMax != 0 || p.FollowMinReps != 0 || p.FollowMaxReps != 0 {
		t.Errorf("Limits of a calibration set = %s, %d to %d, want 0, 0 to 0", p.FollowMax, p.FollowMinReps, p.FollowMaxReps)
	}
}

// The best set has the highest estimate of Epley, from the last session
// with a logged working set (D-340).
func TestLastBest(t *testing.T) {
	lb := domain.Pounds
	p := target("biceps_curl", 3, 10, lb(100))
	cal := domain.SetLog{Kind: domain.SetCalibration, Reps: 12, Weight: lb(200), RIR: 4}
	for _, tc := range []struct {
		name    string
		history []Outcome
		want    *domain.BestSet
	}{
		{"no history", nil, nil},
		// 12 x 100 is 140, and 8 x 110 is 139.3.
		{"the highest estimate", []Outcome{outcome(p, set(8, lb(110), 2), set(12, lb(100), 1), set(10, lb(100), 3))}, &domain.BestSet{Reps: 12, Weight: lb(100), RIR: 1}},
		// 12 x 100 and 10 x 105 both give 140.
		{"a tie goes to the heavier set", []Outcome{outcome(p, set(12, lb(100), 2), set(10, lb(105), 2))}, &domain.BestSet{Reps: 10, Weight: lb(105), RIR: 2}},
		{"a tie of the same set goes to the first", []Outcome{outcome(p, set(10, lb(100), 3), set(10, lb(100), 1))}, &domain.BestSet{Reps: 10, Weight: lb(100), RIR: 3}},
		{"the last session wins", []Outcome{outcome(p, set(12, lb(150), 2)), outcome(p, set(8, lb(100), 2))}, &domain.BestSet{Reps: 8, Weight: lb(100), RIR: 2}},
		{"a skipped session and a calibration set do not count", []Outcome{outcome(p, set(9, lb(100), 2)), outcome(p, cal), {Target: p, Log: domain.ExerciseLog{Exercise: p.Exercise, Skipped: true}}}, &domain.BestSet{Reps: 9, Weight: lb(100), RIR: 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := LastBest(tc.history)
			if (got == nil) != (tc.want == nil) || (got != nil && *got != *tc.want) {
				t.Errorf("LastBest = %v, want %v", got, tc.want)
			}
		})
	}
}

// The rules read the targets of the mirror, and the next session starts
// from the last mirrored values (D-345). The cases of the acceptance
// story are k_mirror_up and k_mirror_zero_rir.
func TestScenarioMirror(t *testing.T) {
	lb := domain.Pounds
	five := stack(10, 150, 5)
	curl := func(sets ...domain.SetLog) Input {
		in := machineInput(t, "biceps_curl", five)
		p := target("biceps_curl", 3, 10, lb(100))
		Limits(&p, in.Exercise, five)
		in.History = []Outcome{outcome(p, sets...)}
		return dated(in)
	}
	for _, tc := range []struct {
		name string
		in   Input
	}{
		{"k_mirror_up", curl(set(12, lb(105), 2), set(12, lb(105), 2), set(12, lb(105), 2))},
		{"k_mirror_zero_rir", curl(set(12, lb(105), 0), set(10, lb(100), 2), set(10, lb(100), 2))},
		{"k_mirror_later_set", curl(set(10, lb(100), 3), set(12, lb(105), 3), set(12, lb(105), 3))},
		{"k_mirror_clamped", curl(set(5, lb(110), 1), set(8, lb(105), 1), set(8, lb(105), 1))},
		{"k_mirror_pain", curl(withPain(set(12, lb(105), 2), 3), set(10, lb(100), 3), set(10, lb(100), 3))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := Next(tc.in)
			if err != nil {
				t.Fatalf("Next: %v", err)
			}
			v, err := Check(d.Target, tc.in)
			if err != nil || len(v) > 0 {
				t.Errorf("Check of the target of Next = %v, %v, want no violation", v, err)
			}
			golden(t, tc.name, render(d))
		})
	}
}
