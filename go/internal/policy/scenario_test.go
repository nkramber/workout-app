package policy

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/nkramber/workout-app/go/internal/domain"
)

var update = flag.Bool("update", false, "write the golden files of testdata")

func exercise(t testing.TB, id domain.ExerciseID) domain.Exercise {
	t.Helper()
	e, ok := domain.DefaultCatalog().Exercise(id)
	if !ok {
		t.Fatalf("exercise %q: not in the catalog", id)
	}
	return e
}

// stack gives the weights from lo to hi in steps of step, in pounds.
func stack(lo, hi, step int64) []domain.Load {
	var out []domain.Load
	for w := lo; w <= hi; w += step {
		out = append(out, domain.Pounds(w))
	}
	return out
}

func machineInput(t testing.TB, id domain.ExerciseID, weights []domain.Load) Input {
	e := exercise(t, id)
	return Input{Exercise: e, Entry: domain.InventoryEntry{Machine: e.Machine, Weights: weights}}
}

func dumbbellInput(t testing.TB, id domain.ExerciseID) Input {
	e := exercise(t, id)
	set := &domain.DumbbellSet{Lightest: domain.Pounds(5), Heaviest: domain.Pounds(50), Step: domain.Pounds(5)}
	return Input{Exercise: e, Entry: domain.InventoryEntry{Machine: e.Machine, Dumbbells: set}}
}

// target gives n working sets of reps at a load in pounds, at 2 reps
// in reserve, with a rest of 120 s.
func target(id domain.ExerciseID, n, reps int, load domain.Load) domain.PlannedExercise {
	p := domain.PlannedExercise{Exercise: id, RestSeconds: 120}
	for range n {
		p.Working = append(p.Working, domain.WorkingSet{Reps: reps, Load: load, RIR: 2})
	}
	return p
}

// set is one logged working set.
func set(reps int, load domain.Load, rir int) domain.SetLog {
	return domain.SetLog{Kind: domain.SetWorking, Reps: reps, Weight: load, RIR: rir}
}

func withPain(s domain.SetLog, p domain.Pain) domain.SetLog {
	s.Pain = &p
	return s
}

func outcome(p domain.PlannedExercise, sets ...domain.SetLog) Outcome {
	return Outcome{Target: p, Log: domain.ExerciseLog{Exercise: p.Exercise, Sets: sets}}
}

// dates gives the dates of a history. The first session is on
// 2026-09-01, and each gap is a count of days: the gap before each
// later session, then the gap before today.
func dates(in Input, gaps ...int) Input {
	d := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	h := slices.Clone(in.History)
	for i := range h {
		if i > 0 {
			d = d.AddDate(0, 0, gaps[i-1])
		}
		h[i].Date = d.Format(domain.DateLayout)
	}
	in.History = h
	in.Today = d.AddDate(0, 0, gaps[len(h)-1]).Format(domain.DateLayout)
	return in
}

// dated gives each history a session each 2 days, with today 2 days
// after the last session, so no break applies.
func dated(in Input) Input {
	gaps := make([]int, len(in.History)+1)
	for i := range gaps {
		gaps[i] = 2
	}
	return dates(in, gaps...)
}

func ruleList(rs []RuleID) string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = string(r)
	}
	return strings.Join(out, ", ")
}

// renderTarget gives the text of a target. A load line shows each load
// that the rounding changed.
func renderTarget(b *strings.Builder, p domain.PlannedExercise, loads []LoadChange) {
	fmt.Fprintf(b, "target: %s, rest %d s\n", p.Exercise, p.RestSeconds)
	for i, s := range p.Calibration {
		fmt.Fprintf(b, "  calibration[%d]: %d reps at %s\n", i, s.Reps, s.Load)
	}
	for i, s := range p.Working {
		fmt.Fprintf(b, "  working[%d]: %d reps at %s, %d RIR\n", i, s.Reps, s.Load, s.RIR)
	}
	if p.FirstSetCalibration {
		fmt.Fprintf(b, "  working[0] is the calibration\n")
	}
	for _, l := range loads {
		if l.Before != l.After {
			fmt.Fprintf(b, "  load %s: %s before the rounding, %s after\n", l.Where, l.Before, l.After)
		}
	}
}

// render gives the text of a decision that a golden file holds.
func render(d Decision) string {
	var b strings.Builder
	fmt.Fprintf(&b, "policy version: %d\n", d.Version)
	fmt.Fprintf(&b, "rules: %s\n", ruleList(d.Rules))
	renderTarget(&b, d.Target, d.Loads)
	fmt.Fprintf(&b, "reason: %s\n", d.Reason)
	return b.String()
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v (run go test with -update)", path, err)
	}
	if got != string(want) {
		t.Errorf("%s:\ngot:\n%s\nwant:\n%s", path, got, want)
	}
}

// The scenarios of section 5 of the high-level roadmap.
func TestScenarios(t *testing.T) {
	lb := domain.Pounds
	curl := target("biceps_curl", 3, 12, lb(25))
	machine := func(h ...Outcome) Input {
		in := machineInput(t, "biceps_curl", stack(10, 150, 5))
		in.History = h
		return in
	}
	early := func(o Outcome) Outcome {
		o.EndedEarly = true
		return o
	}
	skipped := Outcome{Target: curl, Log: domain.ExerciseLog{Exercise: "biceps_curl", Skipped: true}}
	painful := outcome(curl, set(12, lb(25), 3), withPain(set(12, lb(25), 3), 3), set(12, lb(25), 4))
	db := func(id domain.ExerciseID, h ...Outcome) Input {
		in := dumbbellInput(t, id)
		in.History = h
		return in
	}
	odd := func(weights []domain.Load, h ...Outcome) Input {
		in := machineInput(t, "chest_press", weights)
		in.History = h
		return in
	}
	press := target("chest_press", 3, 12, lb(10))

	for _, tc := range []struct {
		name string
		in   Input
	}{
		// Scenario A: 3 x 12 at 25 lb, logged 12, 12, 5.
		{"a_low_rir", machine(outcome(curl, set(12, lb(25), 1), set(12, lb(25), 1), set(5, lb(25), 0)))},
		{"a_repeat", machine(outcome(curl, set(12, lb(25), 2), set(12, lb(25), 2), set(5, lb(25), 0)))},
		{"a_twice", machine(
			outcome(curl, set(12, lb(25), 2), set(12, lb(25), 2), set(5, lb(25), 0)),
			outcome(curl, set(12, lb(25), 2), set(12, lb(25), 1), set(6, lb(25), 0)))},
		{"a_twice_lightest", machine(
			outcome(target("biceps_curl", 3, 12, lb(10)), set(12, lb(10), 2), set(12, lb(10), 2), set(5, lb(10), 0)),
			outcome(target("biceps_curl", 3, 12, lb(10)), set(12, lb(10), 2), set(12, lb(10), 2), set(5, lb(10), 0)))},
		{"a_pain", machine(outcome(curl, set(12, lb(25), 1), set(12, lb(25), 1), withPain(set(5, lb(25), 0), 4)))},
		{"a_small", machine(outcome(curl, set(12, lb(25), 2), set(12, lb(25), 2), set(10, lb(25), 1)))},
		{"a_failure_twice", machine(
			outcome(curl, set(12, lb(25), 3), set(12, lb(25), 3), set(12, lb(25), 0)),
			outcome(curl, set(12, lb(25), 3), set(12, lb(25), 3), set(12, lb(25), 0)))},
		// Scenario B: every rep at 3 or more reps in reserve.
		{"b_load_step", machine(outcome(curl, set(12, lb(25), 3), set(12, lb(25), 3), set(12, lb(25), 4)))},
		{"b_add_reps", machine(outcome(target("biceps_curl", 3, 10, lb(25)), set(10, lb(25), 3), set(10, lb(25), 3), set(10, lb(25), 3)))},
		{"b_add_reps_cap", machine(outcome(target("biceps_curl", 3, 11, lb(25)), set(11, lb(25), 3), set(11, lb(25), 3), set(11, lb(25), 3)))},
		{"b_effort_hold", machine(outcome(curl, set(12, lb(25), 3), set(12, lb(25), 2), set(12, lb(25), 2)))},
		{"b_lighter", machine(outcome(curl, set(12, lb(20), 3), set(12, lb(20), 3), set(12, lb(20), 4)))},
		{"b_dumbbell_reps", db("db_biceps_curl", outcome(target("db_biceps_curl", 3, 12, lb(15)), set(12, lb(15), 3), set(12, lb(15), 3), set(12, lb(15), 3)))},
		{"b_dumbbell_step", db("db_biceps_curl", outcome(target("db_biceps_curl", 3, 20, lb(15)), set(20, lb(15), 3), set(20, lb(15), 3), set(20, lb(15), 3)))},
		{"b_press_hold", db("db_flat_bench_press", outcome(target("db_flat_bench_press", 3, 10, lb(30)), set(10, lb(30), 3), set(10, lb(30), 3), set(10, lb(30), 1)))},
		{"b_machine_weight", odd([]domain.Load{lb(10), lb(14), lb(20)}, outcome(press, set(12, lb(10), 3), set(12, lb(10), 3), set(12, lb(10), 3)))},
		{"b_no_heavier", odd(stack(10, 50, 10), outcome(target("chest_press", 3, 12, lb(30)), set(12, lb(30), 3), set(12, lb(30), 3), set(12, lb(30), 3)))},
		// Scenario C: a pain report holds progression for one session.
		{"c_pain_hold", machine(painful)},
		{"c_after_hold", machine(painful, outcome(curl, set(12, lb(25), 3), set(12, lb(25), 3), set(12, lb(25), 3)))},
		// Scenario D: the session ended early.
		{"d_unlogged", machine(early(outcome(curl, set(12, lb(25), 3), set(12, lb(25), 3))))},
		{"d_short", machine(early(outcome(curl, set(12, lb(25), 3), set(6, lb(25), 0))))},
		{"d_skipped", machine(skipped)},
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
			golden(t, tc.name, render(d))
		})
	}
}

// Scenario C: the warning text of D-153 shows for a pain report alone.
func TestWarning(t *testing.T) {
	lb := domain.Pounds
	for _, tc := range []struct {
		name string
		set  domain.SetLog
		want bool
	}{
		{"no report", set(12, lb(25), 3), false},
		{"rating 0", withPain(set(12, lb(25), 3), 0), false},
		{"rating 1", withPain(set(12, lb(25), 3), 1), true},
		{"rating 10", withPain(set(12, lb(25), 3), 10), true},
	} {
		text, ok := Warning(tc.set)
		if ok != tc.want || (ok && text != PainWarning) || (!ok && text != "") {
			t.Errorf("%s: Warning = %q, %v, want %v", tc.name, text, ok, tc.want)
		}
	}
	// The text stays inside D-36 and D-153: it names the symptom, tells
	// the user to stop, and holds no medical or emergency word.
	low := strings.ToLower(PainWarning)
	if !strings.Contains(low, "pain") || !strings.Contains(low, "stop") {
		t.Errorf("PainWarning %q: want the symptom and the stop", PainWarning)
	}
	for _, w := range []string{"doctor", "physician", "911", "emergency", "diagnos", "injur", "treat", "medical", "call"} {
		if strings.Contains(low, w) {
			t.Errorf("PainWarning %q holds %q", PainWarning, w)
		}
	}
}
