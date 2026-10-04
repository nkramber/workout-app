//go:build emulator

// The acceptance story of PR-35 over the Firebase emulators. For each
// scenario of section 5 of the high-level roadmap, the test sends the
// logged sets of each session through SyncOutbox, with the target that
// the owner saw (D-291), and finishes the workout. The sync revises the
// plan (D-292). The plan then holds the target of the rules in each
// session that holds the exercise (D-290), with a reason of the fake
// provider that names a logged set. A reason that names no logged set,
// and a reason with a load jump, give the reason of the rules (D-288).
package main

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"testing"
	"time"

	"cloud.google.com/go/firestore"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/allowlist"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/policy"
	"github.com/nkramber/workout-app/go/internal/revise"
	"github.com/nkramber/workout-app/go/internal/workout"
)

// logged is one logged session of an exercise: its day from the first
// session, the target that the owner saw, and the sets. No set is a
// skip.
type logged struct {
	day    int
	target domain.PlannedExercise
	sets   []domain.SetLog
	ended  bool
}

// revCase is one exercise of a scenario, with the grade of its final
// target against the last target that the owner saw.
type revCase struct {
	id       string
	exercise domain.ExerciseID
	sessions []logged
	safe     func(final, last domain.PlannedExercise) string
}

func pounds(n int64) domain.Load { return domain.Pounds(n) }

func revTarget(id domain.ExerciseID, n, reps int, load int64) domain.PlannedExercise {
	p := domain.PlannedExercise{Exercise: id, RestSeconds: policy.RestDefault}
	for range n {
		p.Working = append(p.Working, domain.WorkingSet{Reps: reps, Load: pounds(load), RIR: 2})
	}
	return p
}

func revSet(reps int, load int64, rir int) domain.SetLog {
	return domain.SetLog{Kind: domain.SetWorking, Reps: reps, Weight: pounds(load), RIR: rir}
}

func revPain(s domain.SetLog, p domain.Pain) domain.SetLog {
	s.Pain = &p
	return s
}

// allDone gives each set of a target at its reps, at 3 reps in reserve.
func allDone(t domain.PlannedExercise) []domain.SetLog {
	var out []domain.SetLog
	for _, s := range t.Working {
		out = append(out, domain.SetLog{Kind: domain.SetWorking, Reps: s.Reps, Weight: s.Load, RIR: 3})
	}
	return out
}

// weeks gives sessions of one target on days 0, 5, 10, and 15, each set
// done. Each exercise of a plan starts as a return (D-238), so the
// first sessions end before the fourth session, and the rules of
// progression apply.
func weeks(t domain.PlannedExercise) []logged {
	var out []logged
	for _, d := range []int{0, 5, 10, 15} {
		out = append(out, logged{day: d, target: t, sets: allDone(t)})
	}
	return out
}

func notHarder(final, last domain.PlannedExercise) string {
	for i, s := range final.Working {
		if i >= len(last.Working) {
			return "more sets"
		}
		if l := last.Working[i]; s.Load > l.Load || s.Reps > l.Reps || s.RIR < l.RIR {
			return fmt.Sprintf("working[%d] is harder than the last target", i)
		}
	}
	return ""
}

// revScenarios gives the scenarios A to F of section 5 of the
// high-level roadmap, as revisions after a logged session. Each case of
// a scenario holds another exercise.
func revScenarios() map[string][]revCase {
	t25 := func(id domain.ExerciseID) domain.PlannedExercise { return revTarget(id, 3, 12, 25) }
	return map[string][]revCase{
		"A": {
			{"a_low_rir", "biceps_curl", []logged{{0, t25("biceps_curl"), []domain.SetLog{revSet(12, 25, 1), revSet(12, 25, 1), revSet(5, 25, 0)}, false}},
				func(f, l domain.PlannedExercise) string {
					if f.Working[0].Load > l.Working[0].Load || f.Working[0].Reps >= 12 {
						return "no lower rep target at the same load"
					}
					return ""
				}},
			{"a_twice", "seated_row", []logged{
				{0, t25("seated_row"), []domain.SetLog{revSet(12, 25, 2), revSet(12, 25, 2), revSet(5, 25, 0)}, false},
				{2, t25("seated_row"), []domain.SetLog{revSet(12, 25, 2), revSet(12, 25, 1), revSet(6, 25, 0)}, false}},
				func(f, l domain.PlannedExercise) string {
					if f.Working[0].Load >= l.Working[0].Load {
						return "no lower load after two shortfalls"
					}
					return ""
				}},
			{"a_pain", "leg_extension", []logged{{0, t25("leg_extension"), []domain.SetLog{revSet(12, 25, 1), revSet(12, 25, 1), revPain(revSet(5, 25, 0), 4)}, false}}, notHarder},
		},
		"B": {
			{"b_load_step", "biceps_curl", weeks(t25("biceps_curl")), func(f, _ domain.PlannedExercise) string {
				if f.Working[0].Load != pounds(30) || f.Working[0].Reps != 8 {
					return "no load step of 5 lb with the reps at the low end"
				}
				return ""
			}},
			{"b_add_reps", "seated_row", weeks(revTarget("seated_row", 3, 10, 25)), func(f, _ domain.PlannedExercise) string {
				if f.Working[0].Load != pounds(25) || f.Working[0].Reps != 12 {
					return "no added reps at the same load"
				}
				return ""
			}},
		},
		"C": {
			{"c_pain_hold", "biceps_curl", []logged{{0, t25("biceps_curl"), []domain.SetLog{revSet(12, 25, 3), revPain(revSet(12, 25, 3), 3), revSet(12, 25, 4)}, false}}, notHarder},
			{"c_pain_easy", "chest_press", []logged{{0, revTarget("chest_press", 3, 12, 60), []domain.SetLog{revPain(revSet(12, 60, 4), 2), revSet(12, 60, 4), revSet(12, 60, 4)}, false}}, notHarder},
		},
		"D": {
			{"d_unlogged", "biceps_curl", []logged{{0, t25("biceps_curl"), []domain.SetLog{revSet(12, 25, 3), revSet(12, 25, 3)}, true}}, func(f, l domain.PlannedExercise) string {
				if f.Working[0].Load < l.Working[0].Load {
					return "the unlogged sets lowered the load"
				}
				return notHarder(f, l)
			}},
			{"d_skipped", "seated_row", []logged{{0, t25("seated_row"), nil, false}}, notHarder},
		},
		"E": {
			{"e_return", "chest_press", []logged{
				{0, revTarget("chest_press", 3, 12, 100), allDone(revTarget("chest_press", 3, 12, 100)), false},
				{21, revTarget("chest_press", 3, 12, 100), allDone(revTarget("chest_press", 3, 12, 100)), false}},
				func(f, l domain.PlannedExercise) string {
					for i, s := range f.Working {
						if s.RIR != 3 || s.Load > l.Working[0].Load {
							return fmt.Sprintf("working[%d]: want 3 reps in reserve and no load increase after a break", i)
						}
					}
					return ""
				}},
			{"e_long", "leg_press", []logged{
				{0, revTarget("leg_press", 3, 12, 150), allDone(revTarget("leg_press", 3, 12, 150)), false},
				{45, revTarget("leg_press", 3, 12, 150), allDone(revTarget("leg_press", 3, 12, 150)), false}},
				func(f, l domain.PlannedExercise) string {
					if f.Working[0].RIR != 3 || f.Working[0].Load > l.Working[0].Load {
						return "want 3 reps in reserve and no load increase after a break"
					}
					return ""
				}},
		},
		"F": {
			{"f_press", "chest_press", []logged{{0, revTarget("chest_press", 3, 12, 100), allDone(revTarget("chest_press", 3, 12, 100)), false}}, notHarder},
			{"f_leg_press", "leg_press", []logged{{0, revTarget("leg_press", 3, 12, 200), allDone(revTarget("leg_press", 3, 12, 200)), false}}, notHarder},
		},
	}
}

// revInventory gives a confirmed machine for each exercise: a stack
// from 10 lb to 300 lb in 5 lb steps, to 500 lb for the leg press.
func revInventory(ids []domain.ExerciseID) inventory.Inventory {
	var inv inventory.Inventory
	for _, id := range ids {
		e, _ := domain.DefaultCatalog().Exercise(id)
		top := int64(300)
		if id == "leg_press" {
			top = 500
		}
		var w []domain.Load
		for x := int64(10); x <= top; x += 5 {
			w = append(w, pounds(x))
		}
		inv.Machines = append(inv.Machines, inventory.Machine{Entry: domain.InventoryEntry{Machine: e.Machine, Weights: w}, Estimates: map[domain.ExerciseID]domain.Load{}, State: inventory.Confirmed})
	}
	return inv
}

// revEnv is one user of the API with a plan of two sessions. Each
// session holds each exercise of the cases.
type revEnv struct {
	t        *testing.T
	ctx      context.Context
	fs       *firestore.Client
	uid      string
	token    string
	workouts workoutappv1connect.WorkoutServiceClient
	plans    workoutappv1connect.PlanServiceClient
	ops      int
	created  time.Time
}

var revUsers int

func newRevEnv(t *testing.T, authHost, base string, fs *firestore.Client, cases []revCase) *revEnv {
	t.Helper()
	revUsers++
	ctx := context.Background()
	uid, token := signUp(t, authHost, fmt.Sprintf("revision-%d-%d@example.test", time.Now().UnixNano(), revUsers))
	if _, err := fs.Collection(allowlist.Collection).Doc(uid).Set(ctx, map[string]any{}); err != nil {
		t.Fatal(err)
	}
	var ids []domain.ExerciseID
	for _, c := range cases {
		ids = append(ids, c.exercise)
	}
	inv := revInventory(ids)
	if _, err := inventory.FromFirestore(fs).Update(ctx, uid, func(inventory.Inventory) (inventory.Inventory, error) { return inv, nil }); err != nil {
		t.Fatal(err)
	}
	e := &revEnv{t: t, ctx: ctx, fs: fs, uid: uid, token: token, created: time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC),
		workouts: workoutappv1connect.NewWorkoutServiceClient(http.DefaultClient, base),
		plans:    workoutappv1connect.NewPlanServiceClient(http.DefaultClient, base)}
	p := plan.Plan{CreatedAt: e.created, Today: "2026-09-01", PolicyVersion: policy.Version}
	for i := range 2 {
		s := plan.Session{Title: fmt.Sprintf("Session %d", i+1), WarmUp: ai.DefaultWarmUp, CoolDown: ai.DefaultCoolDown}
		for _, c := range cases {
			t0 := c.sessions[0].target
			s.Exercises = append(s.Exercises, plan.Exercise{Target: t0, Reason: "This exercise is new.", Record: policy.Record{Source: policy.SourceRules, Target: t0}})
		}
		p.Sessions = append(p.Sessions, s)
	}
	if err := plan.FromFirestore(fs).Save(ctx, uid, p, plan.Exclusions{}, false); err != nil {
		t.Fatal(err)
	}
	return e
}

func (e *revEnv) entry(entity, id string) *workoutappv1.OutboxEntry {
	e.ops++
	return &workoutappv1.OutboxEntry{
		OpId: fmt.Sprintf("01920000-0000-7000-8000-%012x", e.ops), Entity: entity, EntityId: id,
		At: time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC).Add(time.Duration(e.ops) * time.Second).Format(time.RFC3339), SchemaVersion: workout.SchemaVersion,
	}
}

func seenTarget(p domain.PlannedExercise) *workoutappv1.SeenTarget {
	out := &workoutappv1.SeenTarget{ExerciseId: string(p.Exercise), RestSeconds: int32(p.RestSeconds)}
	for _, s := range p.Working {
		out.WorkingSets = append(out.WorkingSets, &workoutappv1.PlannedSet{Reps: int32(s.Reps), LoadTenthLb: int32(s.Load), RirTarget: int32(s.RIR)})
	}
	return out
}

// session sends one workout of one exercise through SyncOutbox: the
// start with the target copy, each set, and the finish. Each entry must
// apply.
func (e *revEnv) session(c revCase, s logged, date string) {
	e.t.Helper()
	id := fmt.Sprintf("01920000-0000-7000-a000-%012x", e.ops+1)
	header := func(finished bool) *workoutappv1.OutboxEntry {
		h := &workoutappv1.WorkoutHeader{
			Date: date, Plan: &workoutappv1.PlanLink{PlanCreatedAt: e.created.Format(time.RFC3339Nano), SessionIndex: 0},
			Targets: []*workoutappv1.SeenTarget{seenTarget(s.target)}, Finished: finished, EndedEarly: finished && s.ended,
		}
		if finished && len(s.sets) == 0 {
			h.SkippedExerciseIds = []string{string(c.exercise)}
		}
		x := e.entry(workout.EntityWorkout, id)
		x.Payload = &workoutappv1.OutboxEntry_Workout{Workout: h}
		return x
	}
	batch := []*workoutappv1.OutboxEntry{header(false)}
	for _, set := range s.sets {
		x := e.entry(workout.EntitySet, fmt.Sprintf("01920000-0000-7000-9000-%012x", e.ops+1))
		se := &workoutappv1.SetEntry{WorkoutId: id, ExerciseId: string(c.exercise), Kind: string(set.Kind), Reps: int32(set.Reps), WeightTenthsLb: int64(set.Weight), Rir: int32(set.RIR)}
		if set.Pain != nil {
			p := int32(*set.Pain)
			se.Pain = &p
		}
		x.Payload = &workoutappv1.OutboxEntry_Set{Set: se}
		batch = append(batch, x)
	}
	batch = append(batch, header(true))
	res, err := e.workouts.SyncOutbox(e.ctx, signed(e.token, &workoutappv1.SyncOutboxRequest{Entries: batch}))
	if err != nil {
		e.t.Fatalf("SyncOutbox = %v", err)
	}
	for i, r := range res.Msg.GetResults() {
		if r.GetStatus() != workoutappv1.EntryResult_STATUS_APPLIED {
			e.t.Fatalf("%s entry %d = %v", c.id, i, r)
		}
	}
}

// run sends each session of each case, and gives the history of each
// case as the revision reads it.
func (e *revEnv) run(cases []revCase) map[string]policy.Input {
	e.t.Helper()
	start, _ := time.Parse(domain.DateLayout, "2026-09-01")
	inv, err := inventory.FromFirestore(e.fs).Get(e.ctx, e.uid)
	if err != nil {
		e.t.Fatal(err)
	}
	pi := inventory.ForPlan(inv)
	out := map[string]policy.Input{}
	for _, c := range cases {
		ex, _ := domain.DefaultCatalog().Exercise(c.exercise)
		entry, _ := pi.Inventory.Entry(ex.Machine)
		in := policy.Input{Exercise: ex, Entry: entry, Returning: true}
		for _, s := range c.sessions {
			date := start.AddDate(0, 0, s.day).Format(domain.DateLayout)
			e.session(c, s, date)
			in.History = append(in.History, policy.Outcome{Date: date, Target: s.target, Log: domain.ExerciseLog{Exercise: c.exercise, Skipped: len(s.sets) == 0, Sets: s.sets}, EndedEarly: s.ended})
			in.Today = date
		}
		out[c.id] = in
	}
	return out
}

// planned gives the exercise of the plan in each session.
func (e *revEnv) planned() *workoutappv1.Plan {
	e.t.Helper()
	res, err := e.plans.GetPlan(e.ctx, signed(e.token, &workoutappv1.GetPlanRequest{}))
	if err != nil {
		e.t.Fatal(err)
	}
	return res.Msg.GetPlan()
}

func targetOf(p *workoutappv1.PlannedExercise) domain.PlannedExercise {
	out := domain.PlannedExercise{Exercise: domain.ExerciseID(p.GetExerciseId()), RestSeconds: int(p.GetRestSeconds())}
	for _, s := range p.GetCalibrationSets() {
		out.Calibration = append(out.Calibration, domain.CalibrationSet{Reps: int(s.GetReps()), Load: domain.Load(s.GetLoadTenthLb())})
	}
	for _, s := range p.GetWorkingSets() {
		out.Working = append(out.Working, domain.WorkingSet{Reps: int(s.GetReps()), Load: domain.Load(s.GetLoadTenthLb()), RIR: int(s.GetRirTarget())})
	}
	return out
}

func TestRevisionAcceptanceStory(t *testing.T) {
	authHost := requireEmulators(t)
	ctx := context.Background()
	fs, err := firestore.NewClient(ctx, emulatorProject)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fs.Close() }()
	scenarios := revScenarios()
	echo := startAPI(t)

	for _, id := range []string{"A", "B", "C", "D", "E", "F"} {
		t.Run("scenario "+id, func(t *testing.T) {
			cases := scenarios[id]
			e := newRevEnv(t, authHost, echo, fs, cases)
			inputs := e.run(cases)
			p := e.planned()
			if p.GetCreatedAt() != e.created.Format(time.RFC3339) || p.GetLastRevision().GetWorkoutId() == "" {
				t.Fatalf("plan %v: want the same plan with a last revision", p)
			}
			for i, c := range cases {
				want, err := policy.Revise(inputs[c.id])
				if err != nil {
					t.Fatal(err)
				}
				last := c.sessions[len(c.sessions)-1]
				for s, sess := range p.GetSessions() {
					got := sess.GetExercises()[i]
					if final := targetOf(got); !reflect.DeepEqual(final, want.Target) {
						t.Fatalf("%s session %d: target %+v, want the rules target %+v", c.id, s, final, want.Target)
					}
					if why := c.safe(targetOf(got), last.target); why != "" {
						t.Errorf("%s: %s", c.id, why)
					}
					if got.GetSource() != "rules" {
						t.Errorf("%s: source %q, want rules", c.id, got.GetSource())
					}
					switch {
					case len(last.sets) == 0:
						if got.GetReasonSource() != "rules" || got.GetReason() != want.Reason {
							t.Errorf("%s: a skip gave the reason %q (%s), want the reason of the rules", c.id, got.GetReason(), got.GetReasonSource())
						}
					default:
						s1 := last.sets[0]
						if got.GetReasonSource() != "luna" || got.GetReason() != ai.EchoReason(s1.Reps, float64(s1.Weight)/10, s1.RIR) {
							t.Errorf("%s: reason %q (%s), want the reason of the fake that names set 1", c.id, got.GetReason(), got.GetReasonSource())
						}
					}
				}
			}
		})
	}

	// A reason that names no logged set, and a reason with a load jump of
	// 50 percent, give the reason of the rules, and the target of the
	// rules stays (D-288, scenario F).
	for _, tc := range []struct {
		name, text, sets, cause string
	}{
		{"no logged set", "Good work today.", `[]`, revise.CauseNoSet},
		{"load jump", "Set 1 was easy, so the load goes up to 150 lb.", `[{"kind":"working","number":1}]`, revise.CauseNumber},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &ai.Fake{Reply: func(ai.Call) (ai.Reply, error) {
				return ai.Reply{Text: fmt.Sprintf(`{"reasons":[{"exercise_id":"chest_press","logged_sets":%s,"reason":%q}]}`, tc.sets, tc.text)}, nil
			}}
			base := startAPIWith(t, fake)
			cases := scenarios["F"][:1]
			e := newRevEnv(t, authHost, base, fs, cases)
			inputs := e.run(cases)
			want, _ := policy.Revise(inputs["f_press"])
			for _, sess := range e.planned().GetSessions() {
				got := sess.GetExercises()[0]
				if !reflect.DeepEqual(targetOf(got), want.Target) || got.GetReasonSource() != "rules" || got.GetReason() != want.Reason {
					t.Fatalf("exercise %v: want the rules target and the reason of the rules", got)
				}
			}
			stored, _, err := plan.FromFirestore(fs).Get(ctx, e.uid)
			if err != nil {
				t.Fatal(err)
			}
			ex := stored.Sessions[0].Exercises[0]
			if ex.ReasonCause != tc.cause || ex.Record.Model == "" || !slices.Equal(ex.Record.Rules, want.Rules) {
				t.Fatalf("stored exercise cause %q, record %+v", ex.ReasonCause, ex.Record)
			}
			if len(fake.Calls()) != 1 {
				t.Fatalf("%d reviser calls, want 1", len(fake.Calls()))
			}
		})
	}

	// A replay of the last batch revises nothing again.
	t.Run("replay", func(t *testing.T) {
		fake := &ai.Fake{}
		base := startAPIWith(t, fake)
		cases := scenarios["F"][:1]
		e := newRevEnv(t, authHost, base, fs, cases)
		e.run(cases)
		before := e.planned()
		e.ops = 0
		e.run(cases)
		after := e.planned()
		if len(fake.Calls()) != 1 || after.GetLastRevision().GetRevisedAt() != before.GetLastRevision().GetRevisedAt() {
			t.Fatalf("a replay revised the plan again: %d calls", len(fake.Calls()))
		}
		// ListWorkouts gives the target copy of each workout (D-291).
		res, err := e.workouts.ListWorkouts(ctx, signed(e.token, &workoutappv1.ListWorkoutsRequest{}))
		if err != nil || len(res.Msg.GetWorkouts()) != 1 || len(res.Msg.GetWorkouts()[0].GetTargets()) != 1 {
			t.Fatalf("ListWorkouts = %v, %v", res, err)
		}
	})
}
