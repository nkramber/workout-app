package plan

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// fullPlan gives a plan with a value in each field, a refused proposal,
// and a cardio.
func fullPlan() Plan {
	prop := domain.PlannedExercise{Exercise: "chest_press", RestSeconds: 90,
		Calibration: []domain.CalibrationSet{{Reps: 8, Load: 700}},
		Working:     []domain.WorkingSet{{Reps: 8, Load: 1500, RIR: 1}}}
	target := domain.PlannedExercise{Exercise: "chest_press", RestSeconds: 120,
		Calibration: []domain.CalibrationSet{{Reps: 8, Load: 700}},
		Working:     []domain.WorkingSet{{Reps: 8, Load: 700, RIR: 3}, {Reps: 8, Load: 700, RIR: 3}}}
	rec := policy.Record{
		PolicyVersion: 3, Exercise: "chest_press", InputHash: "abc", Model: "m", Effort: "medium", PromptHash: "h",
		Proposal: &prop, Violations: []policy.Violation{{Rule: policy.RuleLoadCeiling, Where: "working[0]", Detail: "d"}},
		Source: policy.SourceRules, Cause: policy.CauseRefused, Rules: []policy.RuleID{policy.RuleFallback, policy.RuleStartEstimate},
		Loads: []policy.LoadChange{{Where: "working[0]", Before: 700, After: 700}}, Target: target, Reason: "rules reason",
	}
	return Plan{
		CreatedAt: now, Today: today, Summary: "summary", Guidance: []ai.GuidanceID{"mobility.hips", "recovery.sleep"},
		Filtered: []ai.Filtered{{Where: "summary", Rule: ai.FilterDiet}},
		Sessions: []Session{{
			Title: "Session 1", WarmUp: ai.DefaultWarmUp, CoolDown: ai.DefaultCoolDown,
			Exercises: []Exercise{{Target: target, Reason: "rules reason", Record: rec}},
			Cardio:    &domain.PlannedCardio{Exercise: "treadmill", Minutes: 10},
		}, {Title: "Session 2", WarmUp: "warm_up.light_sets", CoolDown: "cool_down.stretch",
			Exercises: []Exercise{{Target: target, Reason: "r", Record: policy.Record{Source: policy.SourceLuna, Target: target}}}}},
		Model: "m", Effort: "medium", PromptVersion: ai.PromptVersion, PromptHash: "h", SchemaName: ai.SchemaName,
		PolicyVersion: 3, FilterVersion: 1, GuidanceVersion: 1, CatalogVersion: 1, BodyTablesVersion: 1, Attempts: 2,
	}
}

// TestDocRoundTrip: the stored form keeps each field of a plan.
func TestDocRoundTrip(t *testing.T) {
	p := fullPlan()
	if got := encodePlan(p).plan(); !reflect.DeepEqual(got, p) {
		t.Fatalf("round trip:\n got %+v\nwant %+v", got, p)
	}
	x := Exclusions{Items: []Exclusion{{"leg_press", ""}, {"seated_row", "why"}}, Revision: 7}
	d := encodeExclusions(x)
	if d.Revision != 7 || len(d.Items) != 2 || d.Items[1] != (exclusionDoc{"seated_row", "why"}) {
		t.Fatalf("exclusions doc %+v", d)
	}
}

// TestMemoryCopies: the memory store shares no memory with a caller.
func TestMemoryCopies(t *testing.T) {
	ctx := context.Background()
	s := NewMemory()
	p := fullPlan()
	if err := s.Save(ctx, uid, p, Exclusions{Items: []Exclusion{{"leg_press", "a"}}}, true); err != nil {
		t.Fatal(err)
	}
	p.Sessions[0].Exercises[0].Target.Working[0].Load = 1
	got, ok, err := s.Get(ctx, uid)
	if err != nil || !ok || got.Sessions[0].Exercises[0].Target.Working[0].Load != 700 {
		t.Fatalf("the store shares memory: %v, %v", ok, err)
	}
	got.Sessions[0].Exercises[0].Record.Proposal.Working[0].Load = 1
	again, _, _ := s.Get(ctx, uid)
	if again.Sessions[0].Exercises[0].Record.Proposal.Working[0].Load != 1500 {
		t.Fatal("a read shares memory with the store")
	}
	ex, _ := s.Exclusions(ctx, uid)
	ex.Items[0].Reason = "changed"
	if ex2, _ := s.Exclusions(ctx, uid); ex2.Items[0].Reason != "a" || ex2.Revision != 1 {
		t.Fatalf("exclusions %+v", ex2)
	}
}

// TestMemoryRevision: a save with an old revision changes nothing, and a
// save that keeps the list keeps its revision.
func TestMemoryRevision(t *testing.T) {
	ctx := context.Background()
	s := NewMemory()
	if err := s.Save(ctx, uid, Plan{Today: "a"}, Exclusions{}, false); err != nil {
		t.Fatal(err)
	}
	if ex, _ := s.Exclusions(ctx, uid); ex.Revision != 0 {
		t.Fatalf("revision %d, want 0", ex.Revision)
	}
	if err := s.Save(ctx, uid, Plan{Today: "b"}, Exclusions{Items: []Exclusion{{Exercise: "leg_press"}}}, true); err != nil {
		t.Fatal(err)
	}
	err := s.Save(ctx, uid, Plan{Today: "c"}, Exclusions{}, false)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err %v, want ErrConflict", err)
	}
	if p, _, _ := s.Get(ctx, uid); p.Today != "b" {
		t.Fatalf("plan %q, want b", p.Today)
	}
	for _, bad := range []string{"", "a/b"} {
		if _, _, err := s.Get(ctx, bad); err == nil {
			t.Fatalf("uid %q: want an error", bad)
		}
		if err := s.Save(ctx, bad, Plan{}, Exclusions{}, false); err == nil {
			t.Fatalf("uid %q: want an error", bad)
		}
	}
}

// TestErrorDoc: the stored error record holds the fields of D-236.
func TestErrorDoc(t *testing.T) {
	r := ErrorRecord{
		User: uid, Time: now, ExpireAt: now.Add(ErrorRetention), Request: KindExclude, Attempt: 2, MaxAttempts: 4,
		Status: ai.StatusMalformed, Cause: "cause", Model: "m", Effort: "medium", PromptVersion: "v", PromptHash: "h", SchemaName: "s",
		Cost:   ai.CostRecord{Usage: ai.Usage{InputTokens: 10, OutputTokens: 20}, Reserved: 30, Cost: 40, Known: true},
		Output: "output",
	}
	want := errorDoc{
		User: uid, Time: now, ExpireAt: now.Add(ErrorRetention), Request: "exclude", Attempt: 2, MaxAttempts: 4,
		Status: "malformed", Cause: "cause", Model: "m", Effort: "medium", PromptVersion: "v", PromptHash: "h", SchemaName: "s",
		InputTokens: 10, OutputTokens: 20, Reserved: 30, Cost: 40, CostKnown: true, Output: "output",
	}
	if got := encodeError(r); !reflect.DeepEqual(got, want) {
		t.Fatalf("error doc %+v", got)
	}
}
