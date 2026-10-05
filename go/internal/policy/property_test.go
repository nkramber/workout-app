package policy

import (
	"math/rand/v2"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// The property tests prove the properties of section 6.3 of the
// high-level roadmap over many generated inputs. A fixed seed keeps
// each run the same, so a failure repeats.
const propertyRuns = 20000

type gen struct {
	r         *rand.Rand
	exercises []domain.Exercise
}

func newGen(seed uint64) *gen {
	var ex []domain.Exercise
	for _, e := range domain.DefaultCatalog().Exercises {
		if e.Kind != domain.KindCardio {
			ex = append(ex, e)
		}
	}
	return &gen{r: rand.New(rand.NewPCG(seed, 14)), exercises: ex}
}

func (g *gen) pick(n int) int { return g.r.IntN(n) }

// entry gives an inventory entry: a dumbbell set of D-166, or a stack
// with uneven steps, so the weight selection of D-149 applies.
func (g *gen) entry(e domain.Exercise) domain.InventoryEntry {
	if e.Kind == domain.KindDumbbell {
		steps := []domain.Load{25, 50, 100}
		step := steps[g.pick(len(steps))]
		lightest := step * domain.Load(1+g.pick(4))
		heaviest := lightest + step*domain.Load(g.pick(int((domain.DumbbellMax-lightest)/step)+1))
		return domain.InventoryEntry{Machine: e.Machine, Dumbbells: &domain.DumbbellSet{Lightest: lightest, Heaviest: heaviest, Step: step}}
	}
	steps := []domain.Load{25, 50, 50, 50, 75, 100, 125, 150, 200, 30, 40}
	w := domain.Load(10 * (1 + g.pick(30)))
	var ws []domain.Load
	for range 3 + g.pick(25) {
		ws = append(ws, w)
		w += steps[g.pick(len(steps))]
	}
	return domain.InventoryEntry{Machine: e.Machine, Weights: ws}
}

func validLoads(available []domain.Load) []domain.Load {
	var out []domain.Load
	for _, w := range available {
		if Valid(w, available) {
			out = append(out, w)
		}
	}
	return out
}

// target gives a target at a valid load. With wild true, the reps,
// the reps in reserve, and the rest can be outside the bounds, as a
// target of an older policy version can be.
func (g *gen) target(e domain.Exercise, loads []domain.Load, wild bool) domain.PlannedExercise {
	p := domain.PlannedExercise{Exercise: e.ID, RestSeconds: 60 + g.pick(121)}
	reps, rir := RepRange(e), RIRRange(e)
	load := loads[g.pick(len(loads))]
	for range 1 + g.pick(4) {
		s := domain.WorkingSet{Reps: reps.Min + g.pick(reps.Max-reps.Min+1), Load: load, RIR: rir.Min + g.pick(rir.Max-rir.Min+1)}
		if wild {
			s.Reps, s.RIR = 1+g.pick(25), g.pick(5)
		}
		p.Working = append(p.Working, s)
	}
	switch g.pick(10) {
	case 0:
		// A target of policy version 6 or earlier (D-267).
		p.Calibration = []domain.CalibrationSet{{Reps: p.Working[0].Reps, Load: load}}
	case 1:
		p.FirstSetCalibration = true
	}
	if wild {
		p.RestSeconds = g.pick(400)
		if g.pick(4) == 0 {
			// A load that the machine does not have, as after a change
			// of the inventory (RuleLoadRepair).
			l := domain.Load(1 + g.pick(1200))
			for i := range p.Working {
				p.Working[i].Load = l
			}
		}
	}
	return p
}

func (g *gen) log(p domain.PlannedExercise, available []domain.Load) (domain.ExerciseLog, bool) {
	l := domain.ExerciseLog{Exercise: p.Exercise}
	if g.pick(10) == 0 {
		l.Skipped = true
		return l, g.pick(2) == 0
	}
	if len(p.Calibration) > 0 || g.pick(8) == 0 {
		w := available[0]
		if len(p.Calibration) > 0 {
			w = p.Calibration[0].Load
		}
		for range g.pick(5) {
			s := domain.SetLog{Kind: domain.SetCalibration, Reps: 1 + g.pick(12), Weight: w, RIR: g.pick(8)}
			if g.pick(8) == 0 {
				p := domain.Pain(1 + g.pick(10))
				s.Pain = &p
			}
			l.Sets = append(l.Sets, s)
		}
	}
	n := len(p.Working)
	early := g.pick(4) == 0
	if early {
		n = g.pick(n + 1)
	} else if g.pick(6) == 0 {
		n++
	}
	for i := range n {
		t := p.Working[min(i, len(p.Working)-1)]
		s := domain.SetLog{Kind: domain.SetWorking, Reps: t.Reps, Weight: t.Load, RIR: 3 + g.pick(2)}
		switch g.pick(5) {
		case 0:
			s.Reps = g.pick(t.Reps + 1)
		case 1:
			s.Reps += g.pick(3)
		}
		if g.pick(3) == 0 {
			s.RIR = g.pick(5)
		}
		if g.pick(8) == 0 {
			s.Weight = available[g.pick(len(available))]
		}
		if g.pick(6) == 0 {
			p := domain.Pain(g.pick(11))
			s.Pain = &p
		}
		l.Sets = append(l.Sets, s)
	}
	return l, early
}

// gap gives the days between two sessions: most often a few days, and
// sometimes each row of the long-break table (D-151, D-179).
func (g *gen) gap() int {
	switch g.pick(40) {
	case 0:
		return BreakDays + g.pick(LongBreakDays-BreakDays)
	case 1:
		return LongBreakDays + g.pick(RecalibrateDays-LongBreakDays)
	case 2:
		return RecalibrateDays + g.pick(1500)
	case 3:
		return 10 + g.pick(8) // near the limit of 14 days
	}
	return g.pick(5)
}

func dateOf(d int) string {
	return time.Unix(int64(d)*86400, 0).UTC().Format(domain.DateLayout)
}

// input gives an exercise with 1 to 7 sessions. Most targets after the
// first are the targets that the policy gave, as in the product.
func (g *gen) input(wild bool) Input {
	e := g.exercises[g.pick(len(g.exercises))]
	in := Input{Exercise: e, Entry: g.entry(e), Returning: g.pick(12) == 0}
	available := in.Entry.Available()
	loads := validLoads(available)
	p := g.target(e, loads, wild)
	d := 20000 + g.pick(1000)
	for range 1 + g.pick(7) {
		l, early := g.log(p, available)
		in.History = append(in.History, Outcome{Date: dateOf(d), Target: p, Log: l, EndedEarly: early})
		d += g.gap()
		in.Today = dateOf(d)
		if g.pick(3) != 0 {
			if next, err := Next(in); err == nil {
				p = next.Target
				continue
			}
		}
		p = g.target(e, loads, wild)
	}
	return in
}

func forInputs(t *testing.T, seed uint64, wild bool, f func(t *testing.T, in Input, d Decision)) {
	t.Helper()
	g := newGen(seed)
	for i := range propertyRuns {
		in := g.input(wild)
		d, err := Next(in)
		if err != nil {
			t.Fatalf("run %d: Next: %v", i, err)
		}
		f(t, in, d)
		if t.Failed() {
			t.Fatalf("run %d: input %+v\ndecision %+v", i, in, d)
		}
	}
}

// last gives the last outcome with the load that its calibration set,
// as the rules read it.
func last(in Input) Outcome {
	h := in.effective()
	return h[len(h)-1]
}

// lastSet gives the set of the last target at the index of a set of the
// next target. A next target after a break can have more sets.
func lastSet(o Outcome, i int) domain.WorkingSet {
	return o.Target.Working[min(i, len(o.Target.Working)-1)]
}

func missedReps(o Outcome) bool {
	total, _ := o.shortfall()
	return total > 0
}

// No load increase follows a set with missed reps.
func TestPropertyNoIncreaseAfterMissedReps(t *testing.T) {
	forInputs(t, 1, false, func(t *testing.T, in Input, d Decision) {
		o := last(in)
		if !missedReps(o) {
			return
		}
		for i, s := range d.Target.Working {
			if s.Load > lastSet(o, i).Load {
				t.Errorf("working[%d]: load %s after missed reps, last load %s", i, s.Load, lastSet(o, i).Load)
			}
		}
	})
}

// No output adds more than one 5 lb step for each exercise in each
// session (D-147).
func TestPropertyOneStep(t *testing.T) {
	forInputs(t, 2, false, func(t *testing.T, in Input, d Decision) {
		o := last(in)
		for i, s := range d.Target.Working {
			if s.Load > lastSet(o, i).Load+Step {
				t.Errorf("working[%d]: load %s, last load %s: more than one 5 lb step", i, s.Load, lastSet(o, i).Load)
			}
		}
	})
}

// No output exceeds the rep bounds or the reps-in-reserve bounds, and
// each output is inside each bound of Check, also when the last target
// was outside the bounds.
func TestPropertyBounds(t *testing.T) {
	forInputs(t, 3, true, func(t *testing.T, in Input, d Decision) {
		reps, rir := RepRange(in.Exercise), RIRRange(in.Exercise)
		if !RestLimits.Has(d.Target.RestSeconds) {
			t.Errorf("rest %d s: outside %v", d.Target.RestSeconds, RestLimits)
		}
		if len(d.Target.Working) == 0 {
			t.Error("no working set")
		}
		for i, s := range d.Target.Working {
			if !reps.Has(s.Reps) || !RepLimits.Has(s.Reps) {
				t.Errorf("working[%d]: reps %d outside %v", i, s.Reps, reps)
			}
			if !rir.Has(s.RIR) || s.RIR < 1 || s.RIR > 3 {
				t.Errorf("working[%d]: rir %d outside %v", i, s.RIR, rir)
			}
		}
		if v, err := Check(d.Target, in); err != nil || len(v) > 0 {
			t.Errorf("Check of the target = %v, %v, want no violation", v, err)
		}
	})
}

// Every load is a multiple of 5 lb (D-65), or the machine weight that
// D-149 selects. A dumbbell load is the load of one dumbbell, at most
// 100 lb (D-155, D-166).
func TestPropertyLoads(t *testing.T) {
	forInputs(t, 4, true, func(t *testing.T, in Input, d Decision) {
		available := in.Entry.Available()
		for i, s := range d.Target.Working {
			if !slices.Contains(available, s.Load) {
				t.Errorf("working[%d]: load %s is not an available weight", i, s.Load)
			}
			if s.Load%Step != 0 && !selected(s.Load, available) {
				t.Errorf("working[%d]: load %s is not a multiple of 5 lb or a D-149 weight", i, s.Load)
			}
			if in.Exercise.Kind == domain.KindDumbbell && s.Load > domain.DumbbellMax {
				t.Errorf("working[%d]: dumbbell load %s above %s", i, s.Load, domain.DumbbellMax)
			}
		}
	})
}

// selected tells whether D-149 gives a load for some multiple of 5 lb,
// by a search of each multiple up to the heaviest weight. It does not
// use Valid, so it checks Valid too.
func selected(l domain.Load, available []domain.Load) bool {
	top := available[len(available)-1]
	for r := domain.Load(0); r <= top+Step; r += Step {
		if Select(r, available) == l {
			return true
		}
	}
	return false
}

// A pain flag blocks progression on that exercise in the next session.
func TestPropertyPainHold(t *testing.T) {
	forInputs(t, 5, false, func(t *testing.T, in Input, d Decision) {
		o := last(in)
		if !o.pain() {
			return
		}
		reps := RepRange(in.Exercise)
		for i, s := range d.Target.Working {
			was := lastSet(o, i)
			if s.Load > was.Load || s.Reps > reps.clamp(was.Reps) {
				t.Errorf("working[%d]: %d reps at %s after pain, last %d reps at %s", i, s.Reps, s.Load, was.Reps, was.Load)
			}
		}
	})
}

// An unlogged set is skipped work, not a missed rep: with no missed rep
// on a logged set, no load goes down (D-63, D-64, D-170). A break of 14
// days or more lowers the load by the long-break table, so this
// property reads the sessions with no such gap.
func TestPropertyUnloggedIsNotMissed(t *testing.T) {
	forInputs(t, 6, false, func(t *testing.T, in Input, d Decision) {
		o := last(in)
		if missedReps(o) || in.pause().gap >= BreakDays {
			return
		}
		for i, s := range d.Target.Working {
			if s.Load < lastSet(o, i).Load {
				t.Errorf("working[%d]: load %s below last load %s with no missed rep", i, s.Load, lastSet(o, i).Load)
			}
		}
	})
}

// The same history and the same policy version give the same targets.
func TestPropertyDeterministic(t *testing.T) {
	forInputs(t, 7, true, func(t *testing.T, in Input, d Decision) {
		again, err := Next(in)
		if err != nil || !reflect.DeepEqual(d, again) {
			t.Errorf("Next gave %+v, then %+v, %v", d, again, err)
		}
		if d.Version != Version || len(d.Rules) == 0 || d.Reason == "" {
			t.Errorf("decision %+v: want the version, a rule, and a reason", d)
		}
		for _, r := range d.Rules {
			if !slices.ContainsFunc(Rules(), func(x Rule) bool { return x.ID == r }) {
				t.Errorf("rule %q: not in the rules", r)
			}
		}
	})
}

// Every proposal outside the bounds is refused: Check gives a
// violation. A proposal inside the bounds gets none.
// TestPropertyFallback proves the fallback of each refusal.
func TestPropertyRefusal(t *testing.T) {
	g := newGen(8)
	for i := range propertyRuns {
		in := g.input(false)
		d, err := Next(in)
		if err != nil {
			t.Fatalf("run %d: Next: %v", i, err)
		}
		p := g.mutate(d.Target, in)
		v, err := Check(p, in)
		if err != nil {
			t.Fatalf("run %d: Check: %v", i, err)
		}
		if out := outside(p, d.Target, in); out != (len(v) > 0) {
			t.Fatalf("run %d: outside = %v, violations %v\nproposal %+v\ntarget %+v", i, out, v, p, d.Target)
		}
	}
}

// mutate changes one value of a target, often past a bound.
func (g *gen) mutate(t domain.PlannedExercise, in Input) domain.PlannedExercise {
	p := t
	p.Working = slices.Clone(t.Working)
	p.Calibration = slices.Clone(t.Calibration)
	i := g.pick(len(p.Working))
	available := in.Entry.Available()
	switch g.pick(13) {
	case 0:
		p.Working[i].Reps = g.pick(30)
	case 1:
		p.Working[i].RIR = g.pick(6)
	case 2:
		p.RestSeconds = g.pick(400)
	case 3:
		p.Working[i].Load = available[g.pick(len(available))]
	case 4:
		p.Working[i].Load += domain.Load(g.pick(200) - 100)
	case 5:
		p.Working[i].Load = p.Working[i].Load * 3 / 2 // a 50 percent jump, scenario F
	case 6:
		// The rules decide the calibration, so a proposal can not
		// change it (D-301).
		p.FirstSetCalibration = !p.FirstSetCalibration
	case 7:
		p.Calibration = append(p.Calibration, domain.CalibrationSet{Reps: p.Working[0].Reps, Load: available[g.pick(len(available))]})
	case 8:
		p.Working = append(p.Working, p.Working[i])
	case 9:
		p.Working[i].RIR = 2
	case 10:
		p.Working[i].Reps += 1 + g.pick(3)
	case 11:
		p.Calibration = append(p.Calibration, domain.CalibrationSet{Reps: p.Working[0].Reps, Load: p.Working[0].Load})
	}
	return p
}

// outside is the oracle of the bounds, written apart from Check.
func outside(p, ceiling domain.PlannedExercise, in Input) bool {
	if !RestLimits.Has(p.RestSeconds) {
		return true
	}
	available := in.Entry.Available()
	bad := func(reps int, l domain.Load) bool {
		return reps < 6 || reps > 20 || !slices.Contains(available, l) || (l%Step != 0 && !selected(l, available))
	}
	// The first working set is the calibration (D-297).
	if len(p.Calibration) > 0 {
		return true
	}
	first := afterBreak(in)
	if first && len(p.Working) > len(ceiling.Working) {
		return true
	}
	// The effort ceiling of D-186 applies outside a session of the
	// first-set calibration (D-301).
	effort := !ceiling.FirstSetCalibration
	if effort && len(p.Working) > len(ceiling.Working) {
		return true
	}
	rir := RIRRange(in.Exercise)
	for i, s := range p.Working {
		c := ceiling.Working[min(i, len(ceiling.Working)-1)]
		switch {
		case bad(s.Reps, s.Load),
			s.RIR < rir.Min || s.RIR > 3,
			first && s.RIR < 3,
			s.Load > c.Load,
			effort && s.Load == c.Load && (s.Reps > c.Reps || s.RIR < c.RIR):
			return true
		}
	}
	return false
}

// afterBreak is the oracle of the first sessions after a break (D-151),
// written apart from pause. It reads the dates of the sessions with a
// logged working set.
func afterBreak(in Input) bool {
	var days []int
	for _, o := range in.History {
		for _, s := range o.Log.Sets {
			if s.Kind == domain.SetWorking {
				t, _ := time.Parse(domain.DateLayout, o.Date)
				days = append(days, int(t.Unix()/86400))
				break
			}
		}
	}
	if len(days) == 0 {
		return in.Returning
	}
	t, _ := time.Parse(domain.DateLayout, in.Today)
	today := int(t.Unix() / 86400)
	if today-days[len(days)-1] >= 14 {
		return true
	}
	from := -1
	for i := len(days) - 1; i > 0 && from < 0; i-- {
		if days[i]-days[i-1] >= 14 {
			from = i
		}
	}
	if from < 0 {
		if !in.Returning {
			return false
		}
		from = 0
	}
	return len(days)-from+1 <= 3 || today-days[from] < 14
}

// A last load that the machine does not have becomes the weight that
// D-149 selects for the multiple of 5 lb at or below it.
func TestLoadRepair(t *testing.T) {
	lb := domain.Pounds
	weights := []domain.Load{lb(10), lb(12), lb(14), lb(25)}
	for _, tc := range []struct {
		last, want domain.Load
	}{
		{lb(12), lb(10)}, // on the stack, but no multiple of 5 lb selects it
		{lb(24), lb(14)}, // not on the stack: Select(20 lb)
		{lb(30), lb(25)},
		{lb(5), lb(10)}, // below the lightest weight
	} {
		in := machineInput(t, "chest_press", weights)
		p := target("chest_press", 2, 10, tc.last)
		in.History = []Outcome{outcome(p, set(10, tc.last, 2), set(10, tc.last, 2))}
		d, err := Next(dated(in))
		if err != nil {
			t.Fatal(err)
		}
		if got := d.Target.Working[0].Load; got != tc.want || d.Rules[0] != RuleLoadRepair {
			t.Errorf("last load %s: next load %s, rules %v, want %s first by %s", tc.last, got, d.Rules, tc.want, RuleLoadRepair)
		}
	}
}

// A 50 percent load jump is refused (scenario F).
func TestRefuseLargeJump(t *testing.T) {
	lb := domain.Pounds
	in := machineInput(t, "biceps_curl", stack(10, 150, 5))
	in.History = []Outcome{outcome(target("biceps_curl", 3, 12, lb(40)), set(12, lb(40), 3), set(12, lb(40), 3), set(12, lb(40), 3))}
	in = dated(in)
	p := target("biceps_curl", 3, 8, lb(60))
	v, err := Check(p, in)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 3 || v[0].Rule != RuleLoadCeiling {
		t.Fatalf("Check of a 50 percent jump = %v, want 3 violations of %s", v, RuleLoadCeiling)
	}
	p = target("biceps_curl", 3, 8, lb(45))
	if v, _ := Check(p, in); len(v) > 0 {
		t.Fatalf("Check of one 5 lb step = %v, want none", v)
	}
}

// Every Luna proposal outside the bounds becomes a refusal and a
// fallback, never a displayed target (section 6.3 of the high-level
// roadmap, D-23). A proposal inside the bounds is the target. No
// proposal gives the fallback too. Each final target is inside the
// bounds, and the record names the refusal.
func TestPropertyFallback(t *testing.T) {
	g := newGen(9)
	for i := range propertyRuns {
		in := g.input(false)
		next, err := Next(in)
		if err != nil {
			t.Fatalf("run %d: Next: %v", i, err)
		}
		var p Proposal
		if g.pick(10) != 0 {
			prop := g.mutate(next.Target, in)
			p = Proposal{Model: "fake-model", Effort: "medium", PromptHash: "prompt-hash", Target: &prop}
		}
		r, err := Decide(in, p)
		if err != nil {
			t.Fatalf("run %d: Decide: %v", i, err)
		}
		if v, err := Check(r.Target, in); err != nil || len(v) > 0 {
			t.Fatalf("run %d: Check of the final target = %v, %v", i, v, err)
		}
		switch {
		case p.Target == nil:
			if r.Source != SourceRules || r.Cause != CauseNoProposal || r.Proposal != nil {
				t.Fatalf("run %d: no proposal gave %+v", i, r)
			}
		case outside(*p.Target, next.Target, in):
			if r.Source != SourceRules || r.Cause != CauseRefused || len(r.Violations) == 0 {
				t.Fatalf("run %d: a proposal outside the bounds gave source %q, cause %q, violations %v", i, r.Source, r.Cause, r.Violations)
			}
		default:
			// The rules decide the calibration (D-301).
			want := *p.Target
			want.FirstSetCalibration = next.Target.FirstSetCalibration
			// The policy gives the limit of the other sets (D-306).
			want.FollowMax = Follow(want, in.Entry.Available())
			if r.Source != SourceLuna || r.Cause != CauseNone || len(r.Violations) > 0 || !reflect.DeepEqual(r.Target, want) || !reflect.DeepEqual(*r.Proposal, *p.Target) {
				t.Fatalf("run %d: a proposal inside the bounds gave %+v", i, r)
			}
		}
		if r.Source == SourceRules {
			if !reflect.DeepEqual(r.Target, next.Target) || r.Rules[0] != RuleFallback || !slices.Equal(r.Rules[1:], next.Rules) || r.Reason != next.Reason {
				t.Fatalf("run %d: fallback %+v, want the decision of Next %+v", i, r, next)
			}
		}
		if len(r.Loads) != len(r.Target.Calibration)+len(r.Target.Working) || r.PolicyVersion != Version || r.InputHash == "" {
			t.Fatalf("run %d: record %+v: want each load, the version, and the hash", i, r)
		}
		again, _ := Decide(in, p)
		if !reflect.DeepEqual(r, again) {
			t.Fatalf("run %d: Decide gave %+v, then %+v", i, r, again)
		}
	}
}

// No failure target appears in the first sessions after a break: each
// working set stops at 3 reps in reserve, and the load does not go up
// (D-37, D-151). The return session has no more sets than the last
// target, and a lower or equal load.
func TestPropertyFirstSessions(t *testing.T) {
	n := 0
	forInputs(t, 10, false, func(t *testing.T, in Input, d Decision) {
		if !afterBreak(in) {
			return
		}
		n++
		o := last(in)
		for i, s := range d.Target.Working {
			if s.RIR != 3 {
				t.Errorf("working[%d]: rir %d in the first sessions after a break", i, s.RIR)
			}
			if s.Load > lastSet(o, i).Load {
				t.Errorf("working[%d]: load %s above the last load %s in the first sessions after a break", i, s.Load, lastSet(o, i).Load)
			}
		}
		if in.pause().gap >= BreakDays && len(d.Target.Working) > len(o.Target.Working) {
			t.Errorf("return: %d sets, last target %d", len(d.Target.Working), len(o.Target.Working))
		}
	})
	if n < propertyRuns/10 {
		t.Errorf("%d inputs in the first sessions after a break: want more", n)
	}
}

// Each start of a new exercise is inside the bounds, with 3 working sets
// at 3 reps in reserve, the first set as the calibration, and no load
// above the estimate or the lightest weight (D-178, D-180, D-300).
func TestPropertyStart(t *testing.T) {
	g := newGen(11)
	for i := range propertyRuns {
		e := g.exercises[g.pick(len(g.exercises))]
		in := Input{Exercise: e, Entry: g.entry(e), Returning: g.pick(3) == 0}
		available := in.Entry.Available()
		if g.pick(4) != 0 {
			in.Estimate = available[0] + domain.Load(g.pick(int(available[len(available)-1]-available[0])+1))
		}
		d, err := Next(in)
		if err != nil {
			t.Fatalf("run %d: Next: %v", i, err)
		}
		p := d.Target
		top := available[0]
		if in.Estimate > 0 {
			top = max(Round(in.Estimate, Other), available[0])
		}
		if len(p.Calibration) != 0 || !p.FirstSetCalibration || len(p.Working) != StartSets {
			t.Fatalf("run %d: start %+v", i, p)
		}
		for j, s := range p.Working {
			if s.RIR != 3 || s.Reps != RepRange(e).Min || s.Load > top {
				t.Fatalf("run %d: working[%d] %+v, estimate %s", i, j, s, in.Estimate)
			}
		}
		if v, err := Check(p, in); err != nil || len(v) > 0 {
			t.Fatalf("run %d: Check of the start = %v, %v", i, v, err)
		}
	}
}

// Each row of the calibration table is for an available weight. Each
// load is an available weight that D-149 selects, and each change is at
// most two 5 lb steps in its direction (D-150, D-267).
func TestPropertyCalibrationTable(t *testing.T) {
	g := newGen(12)
	for i := range propertyRuns {
		e := g.exercises[g.pick(len(g.exercises))]
		in := Input{Exercise: e, Entry: g.entry(e)}
		available := in.Entry.Available()
		table, err := CalibrationTable(in)
		if err != nil {
			t.Fatalf("run %d: CalibrationTable: %v", i, err)
		}
		if len(table) != len(available) {
			t.Fatalf("run %d: %d rows for %d weights", i, len(table), len(available))
		}
		for k, c := range table {
			if c.Weight != available[k] {
				t.Fatalf("run %d: row %d is for %s, want %s", i, k, c.Weight, available[k])
			}
			for _, l := range []domain.Load{c.Down, c.Keep, c.UpOne, c.UpTwo} {
				if !slices.Contains(available, l) || !Valid(l, available) {
					t.Fatalf("run %d: load %s is not a valid weight", i, l)
				}
			}
			if Valid(c.Weight, available) && c.Keep != c.Weight {
				t.Fatalf("run %d: keep %s for the valid weight %s", i, c.Keep, c.Weight)
			}
			if c.Down > c.Keep || c.UpOne < c.Keep || c.UpTwo < c.UpOne || c.UpOne > c.Keep+Step || c.UpTwo > c.Keep+2*Step {
				t.Fatalf("run %d: loads %+v", i, c)
			}
		}
	}
}
