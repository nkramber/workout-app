package policy

import (
	"slices"
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
)

func nextGolden(t *testing.T, name string, in Input) Decision {
	t.Helper()
	d, err := Next(in)
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if v, err := Check(d.Target, in); err != nil || len(v) > 0 {
		t.Errorf("Check of the target of Next = %v, %v, want no violation", v, err)
	}
	golden(t, name, render(d))
	return d
}

func atRIR(p domain.PlannedExercise, rir int) domain.PlannedExercise {
	p.Working = slices.Clone(p.Working)
	for i := range p.Working {
		p.Working[i].RIR = rir
	}
	return p
}

// A missed session: a gap of 7 to 13 days holds the target at 3 reps in
// reserve for one session, and a gap under 7 days changes nothing
// (D-294).
func TestScenarioMissed(t *testing.T) {
	lb := domain.Pounds
	press := func(h ...Outcome) Input {
		in := machineInput(t, "chest_press", stack(10, 150, 5))
		in.History = h
		return in
	}
	before := target("chest_press", 3, 10, lb(100))
	held := atRIR(before, 3)

	for _, tc := range []struct {
		name  string
		in    Input
		rules []RuleID
	}{
		{"g_six_days", dates(press(done(before)), 6), []RuleID{RuleAddReps}},
		{"g_missed_week", dates(press(done(before)), 7), []RuleID{RuleMissed}},
		{"g_missed_13", dates(press(done(before)), 13), []RuleID{RuleMissed}},
		{"g_after_missed", dates(press(done(before), done(held)), 9, 2), []RuleID{RuleAddReps, RuleMissedRestored}},
		// A pain report after a missed week still holds the target, and
		// the reps in reserve go back.
		{"g_after_missed_pain", dates(press(done(before), outcome(held, set(10, lb(100), 3), withPain(set(10, lb(100), 3), 2), set(10, lb(100), 3))), 8, 3), []RuleID{RulePainHold, RuleMissedRestored}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := nextGolden(t, tc.name, tc.in)
			if !slices.Equal(d.Rules, tc.rules) {
				t.Errorf("rules %v, want %v", d.Rules, tc.rules)
			}
			if slices.Contains(tc.rules, RuleMissed) {
				for i, s := range d.Target.Working {
					if s.RIR != 3 || s.Load != before.Working[i].Load || s.Reps != before.Working[i].Reps {
						t.Errorf("working[%d]: %+v, want the same load and reps at 3 reps in reserve", i, s)
					}
				}
			}
			if slices.Contains(tc.rules, RuleMissedRestored) {
				for i, s := range d.Target.Working {
					if s.RIR != before.Working[i].RIR {
						t.Errorf("working[%d]: rir %d, want %d of the target before the missed session", i, s.RIR, before.Working[i].RIR)
					}
				}
			}
		})
	}
}

// deloadHistories gives a curl and a press that each declined in 2
// sessions in a row, at the same load, on 2026-09-01, 09-03, and 09-05.
func deloadHistories(t *testing.T) (curl, press Input) {
	lb := domain.Pounds
	c := target("biceps_curl", 3, 12, lb(25))
	p := target("chest_press", 3, 10, lb(100))
	curl = machineInput(t, "biceps_curl", stack(10, 150, 5))
	curl.History = []Outcome{
		outcome(c, set(12, lb(25), 2), set(12, lb(25), 2), set(12, lb(25), 2)),
		outcome(c, set(12, lb(25), 2), set(12, lb(25), 2), set(10, lb(25), 1)),
		outcome(c, set(12, lb(25), 1), set(11, lb(25), 1), set(9, lb(25), 0)),
	}
	press = machineInput(t, "chest_press", stack(10, 150, 5))
	press.History = []Outcome{
		outcome(p, set(10, lb(100), 2), set(10, lb(100), 2), set(10, lb(100), 2)),
		outcome(p, set(10, lb(100), 2), set(10, lb(100), 2), set(9, lb(100), 1)),
		outcome(p, set(10, lb(100), 1), set(9, lb(100), 1), set(8, lb(100), 1)),
	}
	return dates(curl, 2, 2, 0), dates(press, 2, 2, 0)
}

func histories(ins ...Input) [][]Outcome {
	var out [][]Outcome
	for _, in := range ins {
		out = append(out, in.History)
	}
	return out
}

// The reactive deload: a decline in 2 sessions in a row on 2 or more
// exercises starts 7 days at 0.6 times the sets, the same load, and 3
// reps in reserve. Then the targets of before the deload return
// (D-289, D-295).
func TestScenarioDeload(t *testing.T) {
	lb := domain.Pounds
	curl, press := deloadHistories(t)
	starts, err := Deloads(histories(curl, press))
	if err != nil || !slices.Equal(starts, []string{"2026-09-05"}) {
		t.Fatalf("Deloads = %v, %v, want [2026-09-05]", starts, err)
	}
	curl.Deloads, press.Deloads = starts, starts

	// A third exercise with no decline gets the deload too.
	row := machineInput(t, "seated_row", stack(10, 150, 5))
	row.History = []Outcome{done(target("seated_row", 4, 10, lb(60)))}
	row = dates(row, 0)
	row.History[0].Date = "2026-09-04"
	row.Deloads = starts

	before, err := Next(func() Input { c := curl; c.Deloads = nil; return c }())
	if err != nil {
		t.Fatal(err)
	}

	on := func(in Input, today string, h ...Outcome) Input {
		in.History = append(slices.Clone(in.History), h...)
		in.Today = today
		return in
	}
	deloadSet := target("biceps_curl", 2, 12, lb(25))
	deloadSet = atRIR(deloadSet, 3)
	logged := outcome(deloadSet, set(12, lb(25), 3), set(10, lb(25), 3))
	logged.Date = "2026-09-07"
	logged2 := logged
	logged2.Date = "2026-09-11"

	for _, tc := range []struct {
		name   string
		in     Input
		deload bool
	}{
		{"h_deload_start", curl, true},
		{"h_deload_press", press, true},
		{"h_deload_other", on(row, "2026-09-06"), true},
		{"h_deload_session", on(curl, "2026-09-09", logged), true},
		{"h_deload_last_day", on(curl, "2026-09-12", logged, logged2), true},
		{"h_after_deload", on(curl, "2026-09-13", logged, logged2), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := nextGolden(t, tc.name, tc.in)
			if got := slices.Contains(d.Rules, RuleDeload); got != tc.deload {
				t.Fatalf("deload %v, want %v: rules %v", got, tc.deload, d.Rules)
			}
			for i, s := range d.Target.Working {
				if tc.deload && s.RIR != 3 {
					t.Errorf("working[%d]: rir %d in a deload, want 3", i, s.RIR)
				}
			}
		})
	}

	// After the deload, the target of before the deload returns: the
	// deload sessions are no evidence.
	after, err := Next(on(curl, "2026-09-13", logged, logged2))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(after.Target.Working, before.Target.Working) {
		t.Errorf("after the deload %+v, want the target of before %+v", after.Target.Working, before.Target.Working)
	}

	// The count of sets: 0.6 times, to the nearest set, at least 1.
	for n, want := range map[int]int{1: 1, 2: 1, 3: 2, 4: 2, 5: 3, 6: 4} {
		b := &builder{in: curl, target: target("biceps_curl", n, 12, lb(25)), before: make([]domain.Load, n)}
		b.deload("2026-09-05")
		if len(b.target.Working) != want {
			t.Errorf("%d sets: %d in a deload, want %d", n, len(b.target.Working), want)
		}
	}
}

// The trigger of the deload reads 2 declines in a row on 2 or more
// exercises, recent sessions alone, and no session of a deload.
func TestDeloads(t *testing.T) {
	lb := domain.Pounds
	curl, press := deloadHistories(t)
	shift := func(in Input, days int) Input {
		in.History = slices.Clone(in.History)
		for i := range in.History {
			in.History[i].Date = dateText(day(in.History[i].Date) + days)
		}
		return in
	}
	lighter := curl
	lighter.History = slices.Clone(curl.History)
	lighter.History[2] = outcome(target("biceps_curl", 3, 12, lb(20)), set(12, lb(20), 1), set(11, lb(20), 1), set(9, lb(20), 0))
	lighter.History[2].Date = curl.History[2].Date
	logs := func(id domain.ExerciseID, date string, load domain.Load, reps ...int) Outcome {
		o := Outcome{Date: date, Target: target(id, len(reps), reps[0], load), Log: domain.ExerciseLog{Exercise: id}}
		for _, r := range reps {
			o.Log.Sets = append(o.Log.Sets, set(r, load, 3))
		}
		return o
	}
	// Two deload sessions with fewer reps, then 3 sessions after the
	// deload with 2 declines.
	deloaded := curl
	deloaded.History = append(slices.Clone(curl.History),
		logs("biceps_curl", "2026-09-07", lb(25), 12, 10),
		logs("biceps_curl", "2026-09-10", lb(25), 9),
		logs("biceps_curl", "2026-09-13", lb(25), 12, 12, 12),
		logs("biceps_curl", "2026-09-15", lb(25), 12, 12, 9),
		logs("biceps_curl", "2026-09-17", lb(25), 12, 10, 8))
	pressAfter := press
	pressAfter.History = append(slices.Clone(press.History),
		logs("chest_press", "2026-09-08", lb(100), 8),
		logs("chest_press", "2026-09-13", lb(100), 10, 10, 10),
		logs("chest_press", "2026-09-15", lb(100), 10, 10, 8),
		logs("chest_press", "2026-09-17", lb(100), 10, 8, 8))
	for _, tc := range []struct {
		name string
		in   []Input
		want []string
	}{
		{"one exercise", []Input{curl}, nil},
		{"two exercises", []Input{curl, press}, []string{"2026-09-05"}},
		{"a lighter load is no decline", []Input{lighter, press}, nil},
		{"a decline 14 days before is not recent", []Input{shift(curl, -14), press}, nil},
		{"a decline 13 days before counts", []Input{shift(curl, -13), press}, []string{"2026-09-05"}},
		{"deload sessions start no deload", []Input{deloaded, press}, []string{"2026-09-05"}},
		{"3 new sessions after the deload", []Input{deloaded, pressAfter}, []string{"2026-09-05", "2026-09-17"}},
	} {
		got, err := Deloads(histories(tc.in...))
		if err != nil || !slices.Equal(got, tc.want) {
			t.Errorf("%s: Deloads = %v, %v, want %v", tc.name, got, err, tc.want)
		}
	}
	bad := curl
	bad.History = slices.Clone(curl.History)
	bad.History[1].Date = "2026-9-03"
	if _, err := Deloads(histories(bad)); err == nil {
		t.Error("Deloads of a bad date: no error")
	}
	in := curl
	in.Deloads = []string{"2026-09-05", "2026-09-05"}
	if _, err := Next(in); err == nil {
		t.Error("Next with deloads out of order: no error")
	}
}

// An override changes the load and the reps of each working set alone,
// inside the bounds of the policy (D-293).
func TestCheckOverride(t *testing.T) {
	lb := domain.Pounds
	in := machineInput(t, "chest_press", stack(10, 150, 5))
	rec := target("chest_press", 3, 10, lb(100))
	cal := rec
	cal.Calibration = []domain.CalibrationSet{{Reps: 10, Load: lb(100)}}
	change := func(p domain.PlannedExercise, f func(p *domain.PlannedExercise)) domain.PlannedExercise {
		p.Working = slices.Clone(p.Working)
		p.Calibration = slices.Clone(p.Calibration)
		f(&p)
		return p
	}
	heavier := change(rec, func(p *domain.PlannedExercise) {
		for i := range p.Working {
			p.Working[i].Load, p.Working[i].Reps = lb(110), 8
		}
	})
	for _, tc := range []struct {
		name string
		rec  domain.PlannedExercise
		o    domain.PlannedExercise
		want int
	}{
		{"load and reps", rec, heavier, 0},
		{"one set", rec, change(rec, func(p *domain.PlannedExercise) { p.Working[2].Reps = 12 }), 0},
		{"calibration follows", cal, change(cal, func(p *domain.PlannedExercise) {
			p.Working[0].Load = lb(90)
			p.Calibration[0].Load = lb(90)
		}), 0},
		{"calibration differs", cal, change(cal, func(p *domain.PlannedExercise) { p.Working[0].Load = lb(90) }), 1},
		{"a set fewer", rec, change(rec, func(p *domain.PlannedExercise) { p.Working = p.Working[:2] }), 1},
		{"reps in reserve", rec, change(rec, func(p *domain.PlannedExercise) { p.Working[0].RIR = 1 }), 1},
		{"rest", rec, change(rec, func(p *domain.PlannedExercise) { p.RestSeconds = 90 }), 1},
		{"reps over 20", rec, change(rec, func(p *domain.PlannedExercise) { p.Working[1].Reps = 21 }), 1},
		{"reps under 6", rec, change(rec, func(p *domain.PlannedExercise) { p.Working[1].Reps = 5 }), 1},
		{"load not on the machine", rec, change(rec, func(p *domain.PlannedExercise) { p.Working[0].Load = lb(155) }), 1},
		{"other exercise", rec, change(rec, func(p *domain.PlannedExercise) { p.Exercise = "biceps_curl" }), 1},
	} {
		v, err := CheckOverride(tc.o, tc.rec, in)
		if err != nil || len(v) != tc.want {
			t.Errorf("%s: CheckOverride = %v, %v, want %d violations", tc.name, v, err, tc.want)
		}
		for _, x := range v {
			if x.Rule != RuleOverride {
				t.Errorf("%s: rule %q, want %q", tc.name, x.Rule, RuleOverride)
			}
		}
	}
	if _, err := CheckOverride(heavier, target("biceps_curl", 3, 10, lb(25)), in); err == nil {
		t.Error("CheckOverride of a recommendation of another exercise: no error")
	}
}
