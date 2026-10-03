package policy

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// done gives an outcome with each planned set logged at its target
// reps and load, at 3 reps in reserve.
func done(p domain.PlannedExercise) Outcome {
	var sets []domain.SetLog
	for _, s := range p.Calibration {
		sets = append(sets, domain.SetLog{Kind: domain.SetCalibration, Reps: s.Reps, Weight: s.Load, RIR: 3})
	}
	for _, s := range p.Working {
		sets = append(sets, set(s.Reps, s.Load, 3))
	}
	return outcome(p, sets...)
}

// renderRecord gives the text of a decision record that a golden file
// holds. The input hash is not in the text, because a change of the
// form of an input changes it. TestRecord checks the hash.
func renderRecord(r Record) string {
	var b strings.Builder
	fmt.Fprintf(&b, "policy version: %d\n", r.PolicyVersion)
	fmt.Fprintf(&b, "exercise: %s\n", r.Exercise)
	fmt.Fprintf(&b, "model: %q, effort: %q, prompt hash: %q\n", r.Model, r.Effort, r.PromptHash)
	if r.Proposal != nil {
		b.WriteString("proposal:\n  ")
		var p strings.Builder
		renderTarget(&p, *r.Proposal, nil)
		b.WriteString(strings.ReplaceAll(strings.TrimSuffix(p.String(), "\n"), "\n", "\n  ") + "\n")
	}
	for _, v := range r.Violations {
		fmt.Fprintf(&b, "violation: %s\n", v)
	}
	fmt.Fprintf(&b, "source: %s, cause: %q\n", r.Source, r.Cause)
	fmt.Fprintf(&b, "rules: %s\n", ruleList(r.Rules))
	renderTarget(&b, r.Target, r.Loads)
	fmt.Fprintf(&b, "reason: %s\n", r.Reason)
	return b.String()
}

// Scenario E of section 5 of the high-level roadmap: no session for 2
// weeks or more (D-151, D-179). The return lowers the load by the
// long-break table. The first sessions stop at 3 reps in reserve, with
// rep progression only, and no target is failure (D-37).
func TestScenarioE(t *testing.T) {
	lb := domain.Pounds
	press := func(h ...Outcome) Input {
		in := machineInput(t, "chest_press", stack(10, 150, 5))
		in.History = h
		return in
	}
	before := done(target("chest_press", 3, 12, lb(100)))
	ret := target("chest_press", 2, 12, lb(90))
	for i := range ret.Working {
		ret.Working[i].RIR = 3
	}
	curl := done(target("biceps_curl", 3, 12, lb(25)))
	curlIn := machineInput(t, "biceps_curl", stack(10, 150, 5))
	curlIn.History = []Outcome{curl}
	newIn := machineInput(t, "chest_press", stack(10, 150, 5))
	newIn.Estimate, newIn.Returning = lb(100), true
	lower := ret
	lower.Working = slices.Clone(ret.Working)
	for i := range lower.Working {
		lower.Working[i].Reps = 10
		lower.Working[i].RIR = 2
	}

	for _, tc := range []struct {
		name string
		in   Input
	}{
		// The return session of each row of the table.
		{"e_short", dates(press(before), 21)},
		{"e_short_halfway", dates(curlIn, 14)},
		{"e_long", dates(press(before), 45)},
		{"e_recalibrate", dates(press(before), 4*365)},
		{"e_new_returning", newIn},
		// The first sessions after the return: 3 sessions or 14 days.
		{"e_first_reps", dates(press(before, done(lower)), 21, 2)},
		{"e_first_hold", dates(press(before, done(ret)), 21, 2)},
		{"e_first_third", dates(press(before, done(ret), done(ret)), 21, 2, 2)},
		{"e_first_days", dates(press(before, done(ret), done(ret), done(ret)), 21, 2, 2, 2)},
		{"e_restored", dates(press(before, done(ret), done(ret), done(ret)), 21, 7, 7, 2)},
		// A return in a calibration keeps the calibration set.
		{"e_recalibrate_after", dates(press(before, outcome(func() domain.PlannedExercise {
			p := target("chest_press", 3, 12, lb(70))
			p.Calibration = []domain.CalibrationSet{{Reps: 12, Load: lb(70)}}
			for i := range p.Working {
				p.Working[i].RIR = 3
			}
			return p
		}(), domain.SetLog{Kind: domain.SetCalibration, Reps: 12, Weight: lb(70), RIR: 5},
			set(12, lb(75), 3), set(12, lb(75), 3), set(12, lb(75), 3))), 400, 3)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := Next(tc.in)
			if err != nil {
				t.Fatalf("Next: %v", err)
			}
			if v, err := Check(d.Target, tc.in); err != nil || len(v) > 0 {
				t.Errorf("Check of the target of Next = %v, %v, want no violation", v, err)
			}
			if tc.in.firstSessions(tc.in.pause()) {
				for i, s := range d.Target.Working {
					if s.RIR != 3 {
						t.Errorf("working[%d]: rir %d in the first sessions after a break, want 3", i, s.RIR)
					}
				}
			}
			golden(t, tc.name, render(d))
		})
	}
}

// Scenario F of section 5 of the high-level roadmap: Luna proposes a 50
// percent load jump. The policy refuses it, the fallback target
// applies, and the decision record names the refusal (D-23, D-176).
func TestScenarioF(t *testing.T) {
	lb := domain.Pounds
	in := machineInput(t, "chest_press", stack(10, 200, 5))
	in.History = []Outcome{done(target("chest_press", 3, 12, lb(100)))}
	in = dated(in)
	jump := target("chest_press", 3, 8, lb(150))
	step := target("chest_press", 3, 8, lb(105))
	for i := range step.Working {
		step.Working[i].RIR = 3
	}
	other := target("leg_press", 3, 8, lb(105))
	luna := func(p *domain.PlannedExercise) Proposal {
		return Proposal{Model: "fake-model", Effort: "medium", PromptHash: "prompt-hash", Target: p}
	}
	start := machineInput(t, "chest_press", stack(10, 200, 5))

	for _, tc := range []struct {
		name   string
		in     Input
		p      Proposal
		source Source
		cause  Cause
	}{
		{"f_refused", in, luna(&jump), SourceRules, CauseRefused},
		{"f_no_proposal", in, luna(nil), SourceRules, CauseNoProposal},
		{"f_other_exercise", in, luna(&other), SourceRules, CauseRefused},
		{"f_accepted", in, luna(&step), SourceLuna, CauseNone},
		{"f_start_lightest", start, Proposal{}, SourceRules, CauseNoProposal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Decide(tc.in, tc.p)
			if err != nil {
				t.Fatalf("Decide: %v", err)
			}
			if r.Source != tc.source || r.Cause != tc.cause {
				t.Errorf("source %q, cause %q, want %q, %q", r.Source, r.Cause, tc.source, tc.cause)
			}
			if tc.cause == CauseRefused && (len(r.Violations) == 0 || reflect.DeepEqual(r.Target, *r.Proposal)) {
				t.Errorf("a refusal with violations %v and target %+v", r.Violations, r.Target)
			}
			if v, err := Check(r.Target, tc.in); err != nil || len(v) > 0 {
				t.Errorf("Check of the final target = %v, %v, want no violation", v, err)
			}
			golden(t, tc.name, renderRecord(r))
		})
	}
}

// The scenario F record names the refusal of the 50 percent jump, with
// a violation of RuleLoadCeiling on each set.
func TestScenarioFRefusal(t *testing.T) {
	lb := domain.Pounds
	in := machineInput(t, "chest_press", stack(10, 200, 5))
	in.History = []Outcome{done(target("chest_press", 3, 12, lb(100)))}
	in = dated(in)
	jump := target("chest_press", 3, 8, lb(150))
	r, err := Decide(in, Proposal{Model: "fake-model", Target: &jump})
	if err != nil {
		t.Fatal(err)
	}
	next, _ := Next(in)
	if r.Cause != CauseRefused || r.Rules[0] != RuleFallback || !reflect.DeepEqual(r.Target, next.Target) {
		t.Fatalf("record %+v: want a refusal with the target of Next", r)
	}
	n := 0
	for _, v := range r.Violations {
		if v.Rule == RuleLoadCeiling {
			n++
		}
	}
	if n != 3 {
		t.Errorf("violations %v: want %s on each of 3 sets", r.Violations, RuleLoadCeiling)
	}
	// The record holds a copy of the proposal, so a later change of the
	// proposal does not change the record.
	jump.Working[0].Load = lb(10)
	if r.Proposal.Working[0].Load != lb(150) {
		t.Error("the record shares the proposal of the caller")
	}
}

// The start of a new exercise (D-150, D-178, D-179, D-180).
func TestStart(t *testing.T) {
	lb := domain.Pounds
	weights := []domain.Load{lb(10), lb(14), lb(20), lb(25), lb(30), lb(40)}
	for _, tc := range []struct {
		name      string
		estimate  domain.Load
		returning bool
		want      domain.Load
		before    domain.Load
		rule      RuleID
	}{
		{"no estimate", 0, false, lb(10), lb(10), RuleStartLightest},
		{"estimate", lb(25), false, lb(25), lb(25), RuleStartEstimate},
		{"halfway up", lb(22) + 5, false, lb(25), lb(22) + 5, RuleStartEstimate},
		{"machine weight", lb(15), false, lb(14), lb(15), RuleStartEstimate},
		{"returning", lb(40), true, lb(30), lb(28), RuleStartEstimate},             // 28 lb rounds to 30 lb
		{"returning halfway", lb(25), true, lb(14), lb(17) + 5, RuleStartEstimate}, // 17.5 lb rounds down to 15 lb, and the stack gives 14 lb
		{"returning lightest", lb(10), true, lb(10), lb(7), RuleStartEstimate},
	} {
		in := machineInput(t, "leg_extension", weights)
		in.Estimate, in.Returning = tc.estimate, tc.returning
		d, err := Next(in)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		p := d.Target
		if len(p.Calibration) != 1 || len(p.Working) != StartSets || d.Rules[0] != tc.rule {
			t.Fatalf("%s: target %+v, rules %v", tc.name, p, d.Rules)
		}
		if p.Calibration[0].Load != tc.want || p.Calibration[0].Reps != 8 {
			t.Errorf("%s: calibration %+v, want 8 reps at %s", tc.name, p.Calibration[0], tc.want)
		}
		for i, s := range p.Working {
			if s.Load != tc.want || s.Reps != 8 || s.RIR != 3 {
				t.Errorf("%s: working[%d] %+v, want 8 reps at %s, 3 RIR", tc.name, i, s, tc.want)
			}
		}
		if d.Loads[0].Before != tc.before || d.Loads[0].After != tc.want {
			t.Errorf("%s: loads %v, want %s before and %s after", tc.name, d.Loads, tc.before, tc.want)
		}
		if p.RestSeconds != DefaultRest(in.Exercise) {
			t.Errorf("%s: rest %d s", tc.name, p.RestSeconds)
		}
		if v, err := Check(p, in); err != nil || len(v) > 0 {
			t.Errorf("%s: Check of the start = %v, %v", tc.name, v, err)
		}
	}
}

// The calibration sessions: the first 3 sessions of a new exercise
// start with a calibration set, and the load does not go up by double
// progression in them (D-177). The working load of a calibration
// session is the load that its calibration set (D-150).
func TestCalibrationSessions(t *testing.T) {
	lb := domain.Pounds
	in := machineInput(t, "leg_extension", stack(10, 150, 5))
	in.Estimate = lb(50)
	var loads []domain.Load
	var cals []int
	for range 5 {
		d, err := Next(in)
		if err != nil {
			t.Fatal(err)
		}
		loads = append(loads, d.Target.Working[0].Load)
		cals = append(cals, len(d.Target.Calibration))
		// In the first session, the owner does the calibration set at 5
		// reps in reserve, so the working sets use one 5 lb step more.
		// Each working set has 12 reps at 3 reps in reserve.
		o := done(d.Target)
		for i := range o.Log.Sets {
			if o.Log.Sets[i].Kind == domain.SetCalibration {
				o.Log.Sets[i].RIR = 3
				if len(in.History) == 0 {
					o.Log.Sets[i].RIR = 5
				}
			} else {
				o.Log.Sets[i].Weight = loads[len(loads)-1]
				if len(in.History) == 0 {
					o.Log.Sets[i].Weight += Step
				}
				o.Log.Sets[i].Reps = 12
			}
		}
		in.History = append(in.History, o)
		in = dated(in)
	}
	// The fourth session is not a calibration session, so its target
	// adds one 5 lb step at the top of the range.
	wantLoads := []domain.Load{lb(50), lb(55), lb(55), lb(60), lb(60)}
	wantCals := []int{1, 1, 1, 0, 0}
	if !slices.Equal(loads, wantLoads) || !slices.Equal(cals, wantCals) {
		t.Errorf("loads %v and calibration sets %v, want %v and %v", loads, cals, wantLoads, wantCals)
	}
}

// A log with more than one calibration set, from a policy before
// version 4, gives the working load of its first calibration set alone
// (D-267).
func TestEffectiveFirstCalibrationSet(t *testing.T) {
	lb := domain.Pounds
	in := machineInput(t, "leg_extension", stack(10, 150, 5))
	in.Estimate = lb(50)
	d, err := Next(in)
	if err != nil {
		t.Fatal(err)
	}
	o := done(d.Target)
	if o.Log.Sets[0].Kind != domain.SetCalibration {
		t.Fatalf("first set %+v, want a calibration set", o.Log.Sets[0])
	}
	o.Log.Sets[0].RIR = 5
	extra := o.Log.Sets[0]
	extra.RIR = 9
	o.Log.Sets = append([]domain.SetLog{o.Log.Sets[0], extra}, o.Log.Sets[1:]...)
	in.History = []Outcome{o}
	if got := in.effective()[0].Target.Working[0].Load; got != lb(55) {
		t.Fatalf("working load %s, want 55 lb from the first calibration set", got)
	}

	// The table applies to the weight that the owner logged: 60 lb at 5
	// reps in reserve gives 65 lb, not 55 lb from the target of 50 lb
	// (D-249, D-267).
	o.Log.Sets[0].Weight = lb(60)
	in.History = []Outcome{o}
	if got := in.effective()[0].Target.Working[0].Load; got != lb(65) {
		t.Fatalf("working load %s, want 65 lb from the logged weight of 60 lb", got)
	}
}

// The calibration table of D-150 gives the load of the working sets
// after the one calibration set of a session, for each weight of the
// machine (D-267).
func TestCalibrationTable(t *testing.T) {
	lb := domain.Pounds
	in := machineInput(t, "leg_extension", []domain.Load{lb(10), lb(14), lb(20), lb(25), lb(30), lb(35), lb(40), lb(45), lb(50)})
	table, err := CalibrationTable(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(table) != 9 {
		t.Fatalf("%d rows, want one for each of the 9 weights", len(table))
	}
	row := func(w domain.Load) CalibrationLoads {
		for _, c := range table {
			if c.Weight == w {
				return c
			}
		}
		t.Fatalf("no row for %s", w)
		return CalibrationLoads{}
	}
	for _, want := range []CalibrationLoads{
		{Weight: lb(25), Down: lb(20), Keep: lb(25), UpOne: lb(30), UpTwo: lb(35)},
		{Weight: lb(20), Down: lb(14), Keep: lb(20), UpOne: lb(25), UpTwo: lb(30)},
		{Weight: lb(10), Down: lb(10), Keep: lb(10), UpOne: lb(14), UpTwo: lb(20)},
		{Weight: lb(50), Down: lb(45), Keep: lb(50), UpOne: lb(50), UpTwo: lb(50)},
		{Weight: lb(45), Down: lb(40), Keep: lb(45), UpOne: lb(50), UpTwo: lb(50)},
	} {
		if got := row(want.Weight); got != want {
			t.Errorf("row %s = %+v, want %+v", want.Weight, got, want)
		}
	}
	if _, err := CalibrationTable(Input{}); !errors.Is(err, ErrInput) {
		t.Errorf("an empty input: error %v, want ErrInput", err)
	}

	c := row(lb(25))
	cal := func(rir int) domain.SetLog {
		return domain.SetLog{Kind: domain.SetCalibration, Reps: 8, Weight: lb(25), RIR: rir}
	}
	pain := func(rir int, p domain.Pain) domain.SetLog { s := cal(rir); s.Pain = &p; return s }
	for _, tc := range []struct {
		name string
		set  domain.SetLog
		want domain.Load
	}{
		{"rir 0", cal(0), lb(20)},
		{"rir 2", cal(2), lb(20)},
		{"rir 3", cal(3), lb(25)},
		{"rir 4", cal(4), lb(25)},
		{"rir 5", cal(5), lb(30)},
		{"rir 6", cal(6), lb(35)},
		{"rir 9", cal(9), lb(35)},
		{"pain 0 keeps", pain(4, 0), lb(25)},
		{"pain 1 at rir 6", pain(6, 1), lb(20)},
	} {
		if got := c.For(tc.set); got != tc.want {
			t.Errorf("%s: For = %s, want %s", tc.name, got, tc.want)
		}
	}
}

// A record is the same for the same input, and the input hash ignores
// a note alone (D-80, D-176).
func TestRecord(t *testing.T) {
	lb := domain.Pounds
	in := machineInput(t, "chest_press", stack(10, 200, 5))
	in.History = []Outcome{done(target("chest_press", 3, 12, lb(100)))}
	in = dated(in)
	a, err := Decide(in, Proposal{})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := Decide(in, Proposal{})
	if !reflect.DeepEqual(a, b) || len(a.InputHash) != 64 {
		t.Fatalf("records %+v and %+v: want the same record with a SHA-256", a, b)
	}
	noted := in
	noted.History = slices.Clone(in.History)
	noted.History[0].Log.Sets = slices.Clone(in.History[0].Log.Sets)
	noted.History[0].Log.Sets[0].Note = "a note of the owner"
	if h, _ := InputHash(noted); h != a.InputHash {
		t.Error("a note changed the input hash")
	}
	if in.History[0].Log.Sets[0].Note != "" {
		t.Error("InputHash changed the input")
	}
	other := in
	other.Today = "2026-09-30"
	if h, _ := InputHash(other); h == a.InputHash {
		t.Error("another date gave the same input hash")
	}
	if _, err := Decide(Input{}, Proposal{}); !errors.Is(err, ErrInput) {
		t.Errorf("Decide of a bad input: %v, want ErrInput", err)
	}
}
