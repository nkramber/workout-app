//go:build emulator

// The acceptance story of PR-36 over the Firebase emulators. The test
// sends the logged sessions through SyncOutbox, and reads the plan on the
// date of the next session. It proves the targets after a missed week
// and after a break of 14 days or more (D-151, D-179, D-294), and the
// deload after a decline (D-295). An override of the owner shows in the
// plan of the next workout, and the recommendation and the reason stay
// as separate records in the plan and in the workout log (D-69, D-293).
package main

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// runFrom sends each session of each case as run does, with day 0 on
// the date start. It gives the history of each case.
func (e *revEnv) runFrom(cases []revCase, start time.Time) map[string]policy.Input {
	e.t.Helper()
	inv, err := inventory.FromFirestore(e.fs).Get(e.ctx, e.uid)
	if err != nil {
		e.t.Fatal(err)
	}
	pi := inventory.ForPlan(inv)
	out := map[string]policy.Input{}
	for _, c := range cases {
		ex, _ := domain.DefaultCatalog().Exercise(c.exercise)
		entry, _ := pi.Inventory.Entry(ex.Machine)
		in := policy.Input{Exercise: ex, Entry: entry}
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

// on gives the plan on a date.
func (e *revEnv) on(today string) *workoutappv1.Plan {
	e.t.Helper()
	res, err := e.plans.GetPlan(e.ctx, signed(e.token, &workoutappv1.GetPlanRequest{Today: today}))
	if err != nil {
		e.t.Fatal(err)
	}
	return res.Msg.GetPlan()
}

func exerciseOf(p *workoutappv1.Plan, s int, id domain.ExerciseID) *workoutappv1.PlannedExercise {
	for _, x := range p.GetSessions()[s].GetExercises() {
		if x.GetExerciseId() == string(id) {
			return x
		}
	}
	return nil
}

func setsOf(in []*workoutappv1.PlannedSet) []domain.WorkingSet {
	var out []domain.WorkingSet
	for _, s := range in {
		out = append(out, domain.WorkingSet{Reps: int(s.GetReps()), Load: domain.Load(s.GetLoadTenthLb()), RIR: int(s.GetRirTarget())})
	}
	return out
}

func TestDisruptionAcceptanceStory(t *testing.T) {
	authHost := requireEmulators(t)
	ctx := context.Background()
	fs, err := firestore.NewClient(ctx, emulatorProject)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fs.Close() }()
	base := startAPI(t)
	now := time.Now().UTC()
	today := now.Format(domain.DateLayout)
	midnight, _ := time.Parse(domain.DateLayout, today)
	daysAgo := func(n int) time.Time { return midnight.AddDate(0, 0, -n) }

	// A missed week and a break: the plan on the date of the next session
	// holds the target of the rules on that date. The plan with no date
	// keeps the target of the revision.
	t.Run("missed week and break", func(t *testing.T) {
		press := revTarget("chest_press", 3, 10, 100)
		legs := revTarget("leg_press", 3, 12, 150)
		missedCase := revCase{"missed", "chest_press", weeks(press), nil}
		breakCase := revCase{"break", "leg_press", weeks(legs), nil}
		cases := []revCase{missedCase, breakCase}
		e := newRevEnv(t, authHost, base, fs, cases)
		inputs := map[string]policy.Input{}
		// The last session of the press is 7 days ago, and of the leg
		// press 20 days ago.
		for k, v := range e.runFrom(cases[:1], daysAgo(22)) {
			inputs[k] = v
		}
		for k, v := range e.runFrom(cases[1:], daysAgo(35)) {
			inputs[k] = v
		}
		stored, dated := e.planned(), e.on(today)
		for _, tc := range []struct {
			c    revCase
			rule policy.RuleID
		}{{missedCase, policy.RuleMissed}, {breakCase, policy.RuleBreakShort}} {
			in := inputs[tc.c.id]
			then, _ := policy.Revise(in)
			in.Today = today
			want, err := policy.Revise(in)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Contains(want.Rules, tc.rule) {
				t.Fatalf("%s: rules %v, want %s", tc.c.id, want.Rules, tc.rule)
			}
			for s := range dated.GetSessions() {
				got := exerciseOf(dated, s, tc.c.exercise)
				if !reflect.DeepEqual(targetOf(got), want.Target) || got.GetReason() != want.Reason || got.GetReasonSource() != "rules" {
					t.Fatalf("%s session %d: %v, want the target and the reason of the rules on %s: %+v", tc.c.id, s, got, today, want.Target)
				}
				if old := exerciseOf(stored, s, tc.c.exercise); !reflect.DeepEqual(targetOf(old), then.Target) {
					t.Fatalf("%s session %d: the plan with no date %+v, want the revision %+v", tc.c.id, s, targetOf(old), then.Target)
				}
			}
			last := tc.c.sessions[len(tc.c.sessions)-1].target
			for i, s := range want.Target.Working {
				if s.RIR != 3 || s.Load > last.Working[0].Load {
					t.Errorf("%s working[%d]: %+v, want 3 reps in reserve and no load increase", tc.c.id, i, s)
				}
			}
			if tc.rule == policy.RuleMissed && !reflect.DeepEqual(want.Target.Working[0].Reps, last.Working[0].Reps) {
				t.Errorf("missed: reps %d, want the reps of the last target %d", want.Target.Working[0].Reps, last.Working[0].Reps)
			}
			if tc.rule == policy.RuleBreakShort && (len(want.Target.Working) != len(last.Working)-1 || want.Target.Working[0].Load != pounds(135)) {
				t.Errorf("break: %+v, want one set fewer at 10 percent less load", want.Target.Working)
			}
		}
		// The plan read saves nothing.
		if p, _, err := plan.FromFirestore(fs).Get(ctx, e.uid); err != nil || p.Revisions != 8 {
			t.Fatalf("stored plan: %d revisions, %v, want 8 and no change", p.Revisions, err)
		}
	})

	// A decline in 2 sessions in a row on 2 exercises starts the deload on
	// the next date, and the plan on that date gives it to each exercise.
	t.Run("deload after a decline", func(t *testing.T) {
		curl := revTarget("biceps_curl", 3, 12, 25)
		row := revTarget("seated_row", 3, 10, 60)
		// The leg extension is at the top of its range, so the rules give
		// a load step. The deload keeps the load of the last target, and
		// the step waits for the end of the deload (D-303).
		ext := revTarget("leg_extension", 3, 12, 40)
		cases := []revCase{
			{"curl", "biceps_curl", []logged{
				{0, curl, []domain.SetLog{revSet(12, 25, 2), revSet(12, 25, 2), revSet(12, 25, 2)}, false},
				{2, curl, []domain.SetLog{revSet(12, 25, 2), revSet(12, 25, 1), revSet(10, 25, 1)}, false},
				{4, curl, []domain.SetLog{revSet(11, 25, 1), revSet(10, 25, 1), revSet(9, 25, 0)}, false}}, nil},
			{"row", "seated_row", []logged{
				{0, row, []domain.SetLog{revSet(10, 60, 2), revSet(10, 60, 2), revSet(10, 60, 2)}, false},
				{2, row, []domain.SetLog{revSet(10, 60, 2), revSet(10, 60, 1), revSet(9, 60, 1)}, false},
				{4, row, []domain.SetLog{revSet(10, 60, 1), revSet(8, 60, 1), revSet(8, 60, 1)}, false}}, nil},
			{"ext", "leg_extension", []logged{{1, ext, allDone(ext), false}}, nil},
		}
		e := newRevEnv(t, authHost, base, fs, cases)
		// The last sessions are yesterday, so today is the first date of
		// the deload (D-295).
		inputs := e.runFrom(cases, daysAgo(5))
		var hs [][]policy.Outcome
		for _, c := range cases {
			hs = append(hs, inputs[c.id].History)
		}
		starts, err := policy.Deloads(hs)
		yesterday := daysAgo(1).Format(domain.DateLayout)
		if err != nil || !slices.Equal(starts, []string{yesterday}) {
			t.Fatalf("Deloads = %v, %v, want [%s]", starts, err, yesterday)
		}
		stored, dated := e.planned(), e.on(today)
		for _, c := range cases {
			in := inputs[c.id]
			in.Deloads = starts
			in.Today = today
			want, err := policy.Revise(in)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Contains(want.Rules, policy.RuleDeload) || len(want.Target.Working) != 2 {
				t.Fatalf("%s: %v %+v, want the deload with 2 sets", c.id, want.Rules, want.Target.Working)
			}
			for s := range dated.GetSessions() {
				got := exerciseOf(dated, s, c.exercise)
				if !reflect.DeepEqual(targetOf(got), want.Target) {
					t.Fatalf("%s session %d: %+v, want the deload %+v", c.id, s, targetOf(got), want.Target)
				}
				for i, w := range targetOf(got).Working {
					if w.RIR != 3 || w.Load != c.sessions[len(c.sessions)-1].target.Working[0].Load {
						t.Errorf("%s working[%d]: %+v, want the same load at 3 reps in reserve", c.id, i, w)
					}
				}
			}
			// Each revision gave the target on the date of its workout,
			// and the deload starts on the next date. So only the plan on
			// the date gives the deload.
			if old := targetOf(exerciseOf(stored, 0, c.exercise)); len(old.Working) != 3 {
				t.Errorf("%s: the plan with no date has %d sets, want 3", c.id, len(old.Working))
			}
		}
	})

	// An override shows in the plan of the next workout. The plan and the
	// workout log keep the recommendation, the override, and the reason
	// as separate records, and the next revision starts from the override.
	t.Run("override", func(t *testing.T) {
		press := revTarget("chest_press", 3, 10, 100)
		c := revCase{"press", "chest_press", []logged{{0, press, allDone(press), false}}, nil}
		e := newRevEnv(t, authHost, base, fs, []revCase{c})
		inputs := e.runFrom([]revCase{c}, daysAgo(3))
		rec := targetOf(exerciseOf(e.on(today), 0, "chest_press"))
		var sets []*workoutappv1.PlannedSet
		for _, w := range rec.Working {
			sets = append(sets, &workoutappv1.PlannedSet{Reps: int32(w.Reps), LoadTenthLb: int32(w.Load + pounds(10))})
		}
		const reason = "The last session felt easy."

		for _, bad := range []struct {
			name string
			req  *workoutappv1.OverrideTargetRequest
		}{
			{"no reason", &workoutappv1.OverrideTargetRequest{Today: today, ExerciseId: "chest_press", WorkingSets: sets}},
			{"no such load", &workoutappv1.OverrideTargetRequest{Today: today, ExerciseId: "chest_press", Reason: reason,
				WorkingSets: []*workoutappv1.PlannedSet{{Reps: 10, LoadTenthLb: 1003}, sets[1], sets[2]}}},
			{"a set fewer", &workoutappv1.OverrideTargetRequest{Today: today, ExerciseId: "chest_press", Reason: reason, WorkingSets: sets[:2]}},
			{"not in the plan", &workoutappv1.OverrideTargetRequest{Today: today, ExerciseId: "leg_press", Reason: reason, WorkingSets: sets}},
		} {
			_, err := e.plans.OverrideTarget(ctx, signed(e.token, bad.req))
			var ce *connect.Error
			if !errors.As(err, &ce) || (ce.Code() != connect.CodeInvalidArgument && ce.Code() != connect.CodeFailedPrecondition) {
				t.Fatalf("%s: OverrideTarget = %v, want INVALID_ARGUMENT or FAILED_PRECONDITION", bad.name, err)
			}
		}

		res, err := e.plans.OverrideTarget(ctx, signed(e.token, &workoutappv1.OverrideTargetRequest{Today: today, ExerciseId: "chest_press", Reason: reason, WorkingSets: sets}))
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range []*workoutappv1.Plan{res.Msg.GetPlan(), e.on(today)} {
			for s := range p.GetSessions() {
				x := exerciseOf(p, s, "chest_press")
				o := x.GetOverride()
				if o == nil || o.GetReason() != reason {
					t.Fatalf("session %d: override %v, want the override with its reason", s, o)
				}
				if !reflect.DeepEqual(targetOf(x), rec) || !reflect.DeepEqual(setsOf(o.GetRecommendedWorkingSets()), rec.Working) {
					t.Fatalf("session %d: the recommendation changed: %+v", s, targetOf(x))
				}
				for i, w := range setsOf(o.GetWorkingSets()) {
					if w.Load != rec.Working[i].Load+pounds(10) || w.Reps != rec.Working[i].Reps || w.RIR != rec.Working[i].RIR {
						t.Fatalf("override working[%d]: %+v", i, w)
					}
				}
			}
		}
		stored, _, err := plan.FromFirestore(fs).Get(ctx, e.uid)
		if err != nil {
			t.Fatal(err)
		}
		so := stored.Sessions[0].Exercises[0].Override
		if so == nil || so.Reason != reason || !reflect.DeepEqual(so.Recommendation, rec) || reflect.DeepEqual(so.Target, rec) || !reflect.DeepEqual(stored.Sessions[0].Exercises[0].Target, rec) {
			t.Fatalf("stored override %+v: want three separate records", so)
		}

		// RemoveOverride shows the recommendation again, and a new
		// override saves.
		rm, err := e.plans.RemoveOverride(ctx, signed(e.token, &workoutappv1.RemoveOverrideRequest{Today: today, ExerciseId: "chest_press"}))
		if err != nil || exerciseOf(rm.Msg.GetPlan(), 0, "chest_press").GetOverride() != nil {
			t.Fatalf("RemoveOverride = %v, %v", rm, err)
		}
		if _, err := e.plans.OverrideTarget(ctx, signed(e.token, &workoutappv1.OverrideTargetRequest{Today: today, ExerciseId: "chest_press", Reason: reason, WorkingSets: sets})); err != nil {
			t.Fatal(err)
		}

		// The next workout shows the override: the phone copies its sets
		// into the target copy, with the recommendation and the reason.
		over := so.Target
		seen := seenTarget(over)
		for _, w := range rec.Working {
			seen.RecommendedWorkingSets = append(seen.RecommendedWorkingSets, &workoutappv1.PlannedSet{Reps: int32(w.Reps), LoadTenthLb: int32(w.Load), RirTarget: int32(w.RIR)})
		}
		seen.OverrideReason = reason
		s := logged{3, over, allDone(over), false}
		e.sessionSeen(c, s, today, seen)

		in := inputs["press"]
		in.History = append(in.History, policy.Outcome{Date: today, Target: over, Log: domain.ExerciseLog{Exercise: "chest_press", Sets: s.sets}})
		in.Today = today
		want, err := policy.Revise(in)
		if err != nil {
			t.Fatal(err)
		}
		after := e.on(today)
		for i := range after.GetSessions() {
			x := exerciseOf(after, i, "chest_press")
			if x.GetOverride() != nil || !reflect.DeepEqual(targetOf(x), want.Target) {
				t.Fatalf("session %d after the workout: %+v, override %v, want the rules target from the override %+v", i, targetOf(x), x.GetOverride(), want.Target)
			}
		}
		if want.Target.Working[0].Load < over.Working[0].Load {
			t.Errorf("the next target %+v starts from the recommendation, want the override", want.Target.Working)
		}

		list, err := e.workouts.ListWorkouts(ctx, signed(e.token, &workoutappv1.ListWorkoutsRequest{}))
		if err != nil {
			t.Fatal(err)
		}
		var found bool
		for _, w := range list.Msg.GetWorkouts() {
			if w.GetDate() != today {
				continue
			}
			tg := w.GetTargets()[0]
			found = tg.GetOverrideReason() == reason && reflect.DeepEqual(setsOf(tg.GetRecommendedWorkingSets()), rec.Working) && reflect.DeepEqual(setsOf(tg.GetWorkingSets()), over.Working)
		}
		if !found {
			t.Fatalf("ListWorkouts: no workout of %s with the override, the recommendation, and the reason", today)
		}
	})
}
