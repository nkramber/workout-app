package revise

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/policy"
	"github.com/nkramber/workout-app/go/internal/workout"
)

const uid = "uid-test-1"

var (
	created = time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)
	lb      = domain.Pounds
)

// fixture holds the stores of one test: a plan of two sessions that
// both hold the chest press, and the seated row in session 0. The
// inventory confirms both machines, with weights from 10 lb to 300 lb.
type fixture struct {
	t        *testing.T
	plans    *plan.Memory
	workouts *workout.Memory
	inv      *inventory.Memory
	fake     *ai.Fake
	reviser  *Reviser
	ops      int
}

func target(id domain.ExerciseID, n, reps int, load domain.Load) domain.PlannedExercise {
	p := domain.PlannedExercise{Exercise: id, RestSeconds: policy.RestDefault}
	for range n {
		p.Working = append(p.Working, domain.WorkingSet{Reps: reps, Load: load, RIR: 2})
	}
	return p
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{t: t, plans: plan.NewMemory(), workouts: workout.NewMemory(), inv: inventory.NewMemory(), fake: &ai.Fake{}}
	var weights []domain.Load
	for w := int64(10); w <= 300; w += 5 {
		weights = append(weights, lb(w))
	}
	machine := func(id domain.MachineID) inventory.Machine {
		return inventory.Machine{Entry: domain.InventoryEntry{Machine: id, Weights: weights}, Estimates: map[domain.ExerciseID]domain.Load{}, State: inventory.Confirmed}
	}
	if _, err := f.inv.Update(context.Background(), uid, func(inventory.Inventory) (inventory.Inventory, error) {
		return inventory.Inventory{Machines: []inventory.Machine{machine("chest_press"), machine("seated_row")}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	ex := func(t domain.PlannedExercise) plan.Exercise {
		return plan.Exercise{Target: t, Reason: "This exercise is new.", Record: policy.Record{Source: policy.SourceRules}}
	}
	p := plan.Plan{CreatedAt: created, Today: "2026-10-01", PolicyVersion: policy.Version, Sessions: []plan.Session{
		{Title: "Session 1", Exercises: []plan.Exercise{ex(target("chest_press", 3, 12, lb(25))), ex(target("seated_row", 3, 12, lb(40)))}},
		{Title: "Session 2", Exercises: []plan.Exercise{ex(target("chest_press", 3, 12, lb(25)))}},
	}}
	if err := f.plans.Save(context.Background(), uid, p, plan.Exclusions{}, false); err != nil {
		t.Fatal(err)
	}
	f.reviser = &Reviser{
		AI:    &ai.Client{Provider: f.fake, Cap: ai.NewMemoryCap(ai.Caps{User: ai.USD, Project: ai.USD})},
		Plans: f.plans, Workouts: f.workouts, Inventory: f.inv,
		Now: func() time.Time { return time.Date(2026, 10, 4, 9, 0, 0, 0, time.UTC) },
	}
	return f
}

func (f *fixture) apply(e workout.Entry) {
	f.t.Helper()
	f.ops++
	e.OpID = fmt.Sprintf("01920000-0000-7000-8000-%012x", f.ops)
	e.At = time.Date(2026, 10, 1, 10, 0, f.ops, 0, time.UTC)
	e.SchemaVersion = workout.SchemaVersion
	if _, err := f.workouts.Apply(context.Background(), uid, e); err != nil {
		f.t.Fatal(err)
	}
}

// log is the log of one exercise in a workout.
type log struct {
	target domain.PlannedExercise
	sets   []domain.SetLog
}

func set(reps int, load domain.Load, rir int) domain.SetLog {
	return domain.SetLog{Kind: domain.SetWorking, Reps: reps, Weight: load, RIR: rir}
}

// workout logs a finished workout of session 0 on a date. copies tells
// whether the header holds the target copies (D-291).
func (f *fixture) workout(n int, date string, copies bool, logs ...log) string {
	f.t.Helper()
	id := fmt.Sprintf("01920000-0000-7000-a000-%012x", n)
	h := workout.Header{Date: date, Plan: workout.PlanLink{PlanCreatedAt: created}}
	if copies {
		for _, l := range logs {
			h.Targets = append(h.Targets, l.target)
		}
	}
	start := h
	f.apply(workout.Entry{Entity: workout.EntityWorkout, EntityID: id, Header: &start})
	for i, l := range logs {
		if len(l.sets) == 0 {
			h.Skipped = append(h.Skipped, l.target.Exercise)
		}
		for j, s := range l.sets {
			s := s
			f.apply(workout.Entry{Entity: workout.EntitySet, EntityID: fmt.Sprintf("01920000-0000-7000-b%03x-%012x", n, i*10+j), WorkoutID: id, Set: &s, SetExercise: l.target.Exercise})
		}
	}
	h.Finished = true
	f.apply(workout.Entry{Entity: workout.EntityWorkout, EntityID: id, Header: &h})
	return id
}

func (f *fixture) plan() plan.Plan {
	f.t.Helper()
	p, ok, err := f.plans.Get(context.Background(), uid)
	if err != nil || !ok {
		f.t.Fatalf("plan: %v, %v", ok, err)
	}
	return p
}

func (f *fixture) revise(id string) Result {
	f.t.Helper()
	r, err := f.reviser.Revise(context.Background(), uid, id)
	if err != nil {
		f.t.Fatal(err)
	}
	return r
}

// input gives the policy input that a revision reads for the chest
// press after the outcomes h.
func pressInput(t *testing.T, f *fixture, h ...policy.Outcome) policy.Input {
	t.Helper()
	e, _ := domain.DefaultCatalog().Exercise("chest_press")
	inv, _ := f.inv.Get(context.Background(), uid)
	entry, _ := inventory.ForPlan(inv).Inventory.Entry(e.Machine)
	return policy.Input{Exercise: e, Entry: entry, History: h, Today: h[len(h)-1].Date}
}

// TestRevise: after a finished workout, the chest press gets the target
// of the rules in both sessions, with the reason of the fake, which
// names set 1. The seated row was not logged, so it keeps its target.
// A replay of the same workout changes nothing (D-290, D-292).
func TestRevise(t *testing.T) {
	f := newFixture(t)
	seen := target("chest_press", 3, 12, lb(25))
	sets := []domain.SetLog{set(12, lb(25), 1), set(12, lb(25), 1), set(5, lb(25), 0)}
	id := f.workout(1, "2026-10-02", true, log{seen, sets})

	res := f.revise(id)
	if !res.Revised || !slices.Equal(res.Exercises, []domain.ExerciseID{"chest_press"}) || res.Status != ai.StatusOK || res.Luna != 1 {
		t.Fatalf("result %+v", res)
	}
	in := pressInput(t, f, policy.Outcome{Date: "2026-10-02", Target: seen, Log: domain.ExerciseLog{Exercise: "chest_press", Sets: sets}})
	want, err := policy.Revise(in)
	if err != nil {
		t.Fatal(err)
	}
	p := f.plan()
	for _, s := range []int{0, 1} {
		got := p.Sessions[s].Exercises[0]
		if !reflect.DeepEqual(got.Target, want.Target) || got.ReasonSource != policy.SourceLuna || got.ReasonCause != "" {
			t.Fatalf("session %d: %+v, want the target %+v with a reason of Luna", s, got, want.Target)
		}
		if got.Reason != ai.EchoReason(12, 25, 1) {
			t.Errorf("session %d reason %q", s, got.Reason)
		}
		r := got.Record
		if r.Source != policy.SourceRules || r.Cause != policy.CauseNone || !slices.Equal(r.Rules, want.Rules) || r.InputHash != want.InputHash || r.Reason != want.Reason || r.Model == "" || r.PromptHash != ai.PromptHash(ai.Reviser()) {
			t.Errorf("session %d record %+v", s, r)
		}
	}
	if row := p.Sessions[0].Exercises[1]; row.Target.Working[0].Load != lb(40) || row.Reason != "This exercise is new." {
		t.Errorf("the seated row changed: %+v", row)
	}
	if p.Revisions != 1 || p.LastRevision == nil || p.LastRevision.WorkoutID != id || !slices.Equal(p.RevisedWorkouts, []string{id}) || !p.CreatedAt.Equal(created) {
		t.Fatalf("revision fields %+v %v", p.LastRevision, p.RevisedWorkouts)
	}
	if again := f.revise(id); again.Revised || len(f.fake.Calls()) != 1 {
		t.Fatalf("a replay revised the plan again: %+v, %d calls", again, len(f.fake.Calls()))
	}
}

// TestReviseHistory: the history holds each finished workout of the
// exercise, the oldest first, with the target that the owner saw. The
// same shortfall in two sessions lowers the load (scenario A).
func TestReviseHistory(t *testing.T) {
	f := newFixture(t)
	seen := target("chest_press", 3, 12, lb(25))
	short := []domain.SetLog{set(12, lb(25), 2), set(12, lb(25), 2), set(5, lb(25), 0)}
	first := f.workout(1, "2026-10-02", true, log{seen, short})
	f.revise(first)
	second := f.workout(2, "2026-10-04", true, log{seen, short})
	f.revise(second)
	in := pressInput(t, f,
		policy.Outcome{Date: "2026-10-02", Target: seen, Log: domain.ExerciseLog{Exercise: "chest_press", Sets: short}},
		policy.Outcome{Date: "2026-10-04", Target: seen, Log: domain.ExerciseLog{Exercise: "chest_press", Sets: short}})
	want, _ := policy.Revise(in)
	got := f.plan().Sessions[1].Exercises[0]
	if !reflect.DeepEqual(got.Target, want.Target) || got.Target.Working[0].Load >= lb(25) || !slices.Contains(want.Rules, policy.RuleShortfallTwice) {
		t.Fatalf("target %+v, want %+v from rules %v", got.Target, want.Target, want.Rules)
	}
	if f.plan().Revisions != 2 || f.plan().LastRevision.WorkoutID != second {
		t.Fatalf("revisions %d", f.plan().Revisions)
	}
}

// TestReviseRulesReason: each failure of the reviser gives the reason
// of the rules with its cause, and the targets of the rules stay
// (D-288, D-292).
func TestReviseRulesReason(t *testing.T) {
	noSet := func(c ai.Call) (ai.Reply, error) {
		return ai.Reply{Text: `{"reasons":[{"exercise_id":"chest_press","logged_sets":[],"reason":"Good work today."}]}`}, nil
	}
	for _, tc := range []struct {
		name  string
		setup func(f *fixture)
		cause string
	}{
		{"no logged set", func(f *fixture) { f.fake.Reply = noSet }, CauseNoSet},
		{"capped", func(f *fixture) { f.reviser.AI.Cap = ai.NewMemoryCap(ai.Caps{}) }, CauseCapped},
		{"timeout", func(f *fixture) { f.fake.Hang = true; f.reviser.AI.Timeout = 10 * time.Millisecond }, "call-timeout"},
		{"malformed", func(f *fixture) { f.fake.Reply = func(ai.Call) (ai.Reply, error) { return ai.Reply{Text: "no"}, nil } }, "call-malformed"},
		{"no reason", func(f *fixture) {
			f.fake.Reply = func(ai.Call) (ai.Reply, error) { return ai.Reply{Text: `{"reasons":[]}`}, nil }
		}, CauseNoReason},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			tc.setup(f)
			seen := target("chest_press", 3, 12, lb(25))
			sets := []domain.SetLog{set(12, lb(25), 3), set(12, lb(25), 3), set(12, lb(25), 3)}
			id := f.workout(1, "2026-10-02", true, log{seen, sets})
			if res := f.revise(id); !res.Revised || res.Luna != 0 {
				t.Fatalf("result %+v", res)
			}
			want, _ := policy.Revise(pressInput(t, f, policy.Outcome{Date: "2026-10-02", Target: seen, Log: domain.ExerciseLog{Exercise: "chest_press", Sets: sets}}))
			got := f.plan().Sessions[0].Exercises[0]
			if !reflect.DeepEqual(got.Target, want.Target) || got.Reason != want.Reason || got.ReasonSource != policy.SourceRules || got.ReasonCause != tc.cause {
				t.Fatalf("exercise %+v, want the rules target and reason with cause %q", got, tc.cause)
			}
			if capped := tc.cause == CauseCapped; (got.Record.Model == "") != capped {
				t.Errorf("record model %q: want none for a capped call alone", got.Record.Model)
			}
		})
	}
}

// TestReviseOlderPhone: a workout with no target copy reads the target
// of its linked session while the plan has no revision. After a
// revision, such a workout adds nothing to the history (D-291).
func TestReviseOlderPhone(t *testing.T) {
	f := newFixture(t)
	sets := []domain.SetLog{set(12, lb(25), 3), set(12, lb(25), 3), set(12, lb(25), 3)}
	id := f.workout(1, "2026-10-02", false, log{target: target("chest_press", 3, 12, lb(25)), sets: sets})
	if res := f.revise(id); !res.Revised || len(res.Exercises) != 1 {
		t.Fatalf("result %+v", res)
	}
	want, _ := policy.Revise(pressInput(t, f, policy.Outcome{Date: "2026-10-02", Target: target("chest_press", 3, 12, lb(25)), Log: domain.ExerciseLog{Exercise: "chest_press", Sets: sets}}))
	if got := f.plan().Sessions[0].Exercises[0].Target; !reflect.DeepEqual(got, want.Target) {
		t.Fatalf("target %+v, want %+v", got, want.Target)
	}
	later := f.workout(2, "2026-10-04", false, log{target: target("chest_press", 3, 12, lb(25)), sets: sets})
	res := f.revise(later)
	if !res.Revised || len(res.Exercises) != 0 || len(f.fake.Calls()) != 1 {
		t.Fatalf("a workout with no copy after a revision: %+v, %d calls", res, len(f.fake.Calls()))
	}
}

// TestReviseNothing: no plan, an unknown or open workout, a new plan,
// and a machine that is not confirmed change nothing.
func TestReviseNothing(t *testing.T) {
	f := newFixture(t)
	if res := f.revise("01920000-0000-7000-a000-0000000000ff"); res.Revised {
		t.Fatal("an unknown workout revised the plan")
	}
	open := "01920000-0000-7000-a000-0000000000fe"
	f.apply(workout.Entry{Entity: workout.EntityWorkout, EntityID: open, Header: &workout.Header{Date: "2026-10-02", Plan: workout.PlanLink{PlanCreatedAt: created}}})
	if res := f.revise(open); res.Revised {
		t.Fatal("an open workout revised the plan")
	}
	if res, err := (&Reviser{AI: f.reviser.AI, Plans: plan.NewMemory(), Workouts: f.workouts, Inventory: f.inv}).Revise(context.Background(), uid, open); err != nil || res.Revised {
		t.Fatalf("no plan: %+v, %v", res, err)
	}

	// The owner removes the chest press, so its target stays.
	if _, err := f.inv.Update(context.Background(), uid, func(inv inventory.Inventory) (inventory.Inventory, error) {
		inv.Machines = inv.Machines[1:]
		return inv, nil
	}); err != nil {
		t.Fatal(err)
	}
	id := f.workout(1, "2026-10-02", true, log{target("chest_press", 3, 12, lb(25)), []domain.SetLog{set(12, lb(25), 3)}})
	if res := f.revise(id); !res.Revised || len(res.Exercises) != 0 || len(f.fake.Calls()) != 0 {
		t.Fatalf("no machine: %+v, %d calls", res, len(f.fake.Calls()))
	}

	// A new plan during the revision wins.
	g := newFixture(t)
	g.fake.Reply = func(c ai.Call) (ai.Reply, error) {
		p := g.plan()
		p.CreatedAt = created.Add(time.Hour)
		if err := g.plans.Save(context.Background(), uid, p, plan.Exclusions{}, false); err != nil {
			t.Fatal(err)
		}
		return ai.EchoReply(c)
	}
	id = g.workout(1, "2026-10-02", true, log{target("chest_press", 3, 12, lb(25)), []domain.SetLog{set(12, lb(25), 3)}})
	if res := g.revise(id); res.Revised || g.plan().Revisions != 0 {
		t.Fatalf("a replaced plan: %+v", res)
	}
}

// TestReviseSkipped: a skipped exercise gets the target and the reason
// of the rules, because the last session logged no set.
func TestReviseSkipped(t *testing.T) {
	f := newFixture(t)
	id := f.workout(1, "2026-10-02", true,
		log{target("chest_press", 3, 12, lb(25)), []domain.SetLog{set(12, lb(25), 3)}},
		log{target: target("seated_row", 3, 12, lb(40))})
	if res := f.revise(id); !slices.Equal(res.Exercises, []domain.ExerciseID{"chest_press", "seated_row"}) || res.Luna != 1 {
		t.Fatalf("result %+v", res)
	}
	row := f.plan().Sessions[0].Exercises[1]
	if row.ReasonSource != policy.SourceRules || row.ReasonCause != CauseNoReason || !slices.Contains(row.Record.Rules, policy.RuleSkipped) {
		t.Fatalf("skipped row %+v", row)
	}
	var in struct {
		Exercises []struct {
			ID          string `json:"exercise_id"`
			RulesReason string `json:"rules_reason"`
		} `json:"exercises"`
	}
	if err := json.Unmarshal(f.fake.Calls()[0].Input, &in); err != nil || len(in.Exercises) != 2 || in.Exercises[1].RulesReason != row.Reason {
		t.Fatalf("reviser input %+v, %v", in, err)
	}
}

// A new plan reads each workout (P2-1 and P2-2 of PR-38): an exercise
// that the newest 2000 workouts omit keeps its history, so its target is
// the return of D-179 and not a new start (D-301).
func TestHistoryBeyondRevisionWindow(t *testing.T) {
	f := newFixture(t)
	press := target("chest_press", 3, 12, lb(100))
	f.workout(1, "2025-06-02", true, log{press, []domain.SetLog{set(12, lb(100), 3), set(12, lb(100), 3), set(12, lb(100), 3)}})
	row := target("seated_row", 3, 12, lb(40))
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	const newer = 2100
	for n := 2; n <= newer+1; n++ {
		date := start.AddDate(0, 0, n).Format(domain.DateLayout)
		f.workout(n, date, true, log{row, []domain.SetLog{set(12, lb(40), 3)}})
	}
	h, err := f.reviser.History(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	if got := h.Outcomes["chest_press"]; len(got) != 1 || got[0].Date != "2025-06-02" {
		t.Fatalf("chest press history %+v, want the workout of 2025-06-02", got)
	}
	if n := len(h.Outcomes["seated_row"]); n != newer {
		t.Fatalf("seated row history of %d workouts, want %d", n, newer)
	}
	in := pressInput(t, f, h.Outcomes["chest_press"]...)
	in.Today = "2026-10-05"
	d, err := policy.Next(in)
	if err != nil {
		t.Fatal(err)
	}
	if d.Rules[0] != policy.RuleBreakRecalibrate || d.Target.Working[0].Load != lb(70) || !d.Target.FirstSetCalibration {
		t.Fatalf("target %+v by %v, want the return at 70 lb with the first set as the calibration", d.Target, d.Rules)
	}
}

// Two syncs of one finished workout at the same time make one reviser
// call (D-304). The first revision claims the workout before its call,
// so the second one stops before its call, and the plan gets one
// revision.
func TestReviseClaim(t *testing.T) {
	f := newFixture(t)
	id := f.workout(1, "2026-10-02", true, log{target("chest_press", 3, 12, lb(25)), []domain.SetLog{set(12, lb(25), 3), set(12, lb(25), 3), set(12, lb(25), 3)}})
	started, release := make(chan struct{}), make(chan struct{})
	f.fake.Reply = func(c ai.Call) (ai.Reply, error) {
		close(started)
		<-release
		return ai.EchoReply(c)
	}
	first := make(chan Result)
	go func() {
		r, err := f.reviser.Revise(context.Background(), uid, id)
		if err != nil {
			t.Error(err)
		}
		first <- r
	}()
	<-started
	if p := f.plan(); len(p.Claims) != 1 || p.Claims[0].WorkoutID != id {
		t.Fatalf("claims %+v during the call, want the claim of the workout", p.Claims)
	}
	if res := f.revise(id); res.Revised || res.Status != "" {
		t.Fatalf("the second sync gave %+v, want no revision and no call", res)
	}
	close(release)
	if res := <-first; !res.Revised {
		t.Fatalf("the first sync gave %+v, want the revision", res)
	}
	if p := f.plan(); len(f.fake.Calls()) != 1 || p.Revisions != 1 || len(p.Claims) != 0 {
		t.Fatalf("%d reviser calls, %d revisions, claims %+v: want 1, 1, and none", len(f.fake.Calls()), p.Revisions, p.Claims)
	}
}

// A claim lasts until the end of its lease. A revision that failed
// before its save leaves its claim, and a replay after the lease
// revises the plan (D-304).
func TestReviseClaimLease(t *testing.T) {
	f := newFixture(t)
	id := f.workout(1, "2026-10-02", true, log{target("chest_press", 3, 12, lb(25)), []domain.SetLog{set(12, lb(25), 3), set(12, lb(25), 3), set(12, lb(25), 3)}})
	now := f.reviser.Now()
	if err := f.plans.Update(context.Background(), uid, created, func(p *plan.Plan) error {
		p.Claim(id, now, now.Add(Timeout))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if res := f.revise(id); res.Revised || len(f.fake.Calls()) != 0 {
		t.Fatalf("a live claim gave %+v and %d calls, want no revision and no call", res, len(f.fake.Calls()))
	}
	f.reviser.Now = func() time.Time { return now.Add(Timeout) }
	if res := f.revise(id); !res.Revised || len(f.fake.Calls()) != 1 || len(f.plan().Claims) != 0 {
		t.Fatalf("after the lease: %+v, %d calls, claims %+v", res, len(f.fake.Calls()), f.plan().Claims)
	}
}
