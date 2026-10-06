package ai

import (
	"context"
	"strings"
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// gymRequest gives a planner request of new machine exercises of the
// owner's gym: push, pull, legs, calves, and core, with each group
// selected and no cardio.
func gymRequest(t testing.TB, sessions int, ids ...domain.ExerciseID) Request {
	if len(ids) == 0 {
		ids = []domain.ExerciseID{"chest_press", "shoulder_press", "triceps_pulldown", "seated_row", "lat_pulldown",
			"biceps_curl", "leg_press", "leg_extension", "seated_leg_curl", "calf_raise", "abdominal_crunch"}
	}
	var weights []domain.Load
	for w := int64(10); w <= 150; w += 5 {
		weights = append(weights, domain.Pounds(w))
	}
	req := Request{User: "uid-test-1", Today: "2026-10-01", Sessions: sessions,
		Profile: &Profile{Experience: "beginner", GoalTemplate: "general_fitness", MuscleGroups: strs(domain.MuscleGroups())}}
	for _, id := range ids {
		e := exerciseOf(t, id)
		req.Exercises = append(req.Exercises, policy.Input{Exercise: e, Entry: domain.InventoryEntry{Machine: e.Machine, Weights: weights}, Today: req.Today})
	}
	return req
}

func layout(p *Plan) [][]domain.ExerciseID {
	out := make([][]domain.ExerciseID, len(p.Sessions))
	for i, s := range p.Sessions {
		for _, e := range s.Exercises {
			out[i] = append(out[i], e.Target.Exercise)
		}
	}
	return out
}

// TestEchoSplitObeysTheRotation: the fake gives a plan that each
// rotation rule accepts, for each count of sessions of D-211.
func TestEchoSplitObeysTheRotation(t *testing.T) {
	for _, n := range []int{2, 3, 4} {
		req := gymRequest(t, n)
		res, err := (&Client{Provider: &Fake{}, Cap: bigCap()}).Plan(context.Background(), req)
		if err != nil || res.Status != StatusOK {
			t.Fatalf("%d sessions: status %q, cause %q, err %v", n, res.Status, res.Cause, err)
		}
		r := RotationOf(req)
		if !r.Applies() {
			t.Fatalf("%d sessions: the rotation does not apply", n)
		}
		l := layout(res.Plan)
		if v := policy.CheckRotation(r, l); len(v) != 0 {
			t.Fatalf("%d sessions: violations %v", n, v)
		}
		for i, s := range l {
			if len(s) == 0 {
				t.Fatalf("%d sessions: session %d has no exercise", n, i)
			}
		}
	}
}

// TestParseRefusesABrokenRotation: the acceptance story of the layer. A
// plan with the same exercises in two sessions in a row is malformed,
// and the cause names the rule and the groups alone (D-80).
func TestParseRefusesABrokenRotation(t *testing.T) {
	fake := &Fake{Reply: edit(t, func(out map[string]any) {
		s := out["sessions"].([]any)
		s[1].(map[string]any)["exercises"] = s[0].(map[string]any)["exercises"]
	})}
	res, err := (&Client{Provider: fake, Cap: bigCap()}).Plan(context.Background(), gymRequest(t, 2))
	if err != nil || res.Status != StatusMalformed {
		t.Fatalf("status %q, err %v: want malformed", res.Status, err)
	}
	if !strings.Contains(res.Cause, "rotation.no-repeat sessions[0] and sessions[1]: group chest in both sessions") ||
		!strings.Contains(res.Cause, "rotation.cover") {
		t.Fatalf("cause %q: want the repeat of chest and a cover violation", res.Cause)
	}
}

// TestPlannerInputGivesTheRotation: the planner input tells if the
// rotation applies, and gives the groups of each exercise. The reviser
// input holds neither.
func TestPlannerInputGivesTheRotation(t *testing.T) {
	req := gymRequest(t, 2, "chest_press", "leg_press", "hip_adduction")
	in, err := userInput(RolePlanner, req)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"rotation":true`, `"exercise_id":"chest_press","name":"Chest press","kind":"machine","region":"upper_push","muscle_groups":["chest","triceps"]`} {
		if !strings.Contains(string(in), want) {
			t.Errorf("planner input lacks %s", want)
		}
	}
	if !strings.Contains(string(in), `"exercise_id":"hip_adduction","name":"Hip adduction","kind":"machine","region":"lower_push","available_weights_lb"`) {
		t.Error("an exercise with no group gives muscle_groups")
	}
	in, err = userInput(RoleReviser, req)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(in), `"rotation"`) || strings.Contains(string(in), `"region":"upper_push","muscle_groups"`) {
		t.Errorf("reviser input %s: want no rotation and no groups", in)
	}
}

// TestNoRotationWhenNoSplitIsPossible: with one unit of groups and no
// filler, the rules do not apply, and the same exercise can be in each
// session.
func TestNoRotationWhenNoSplitIsPossible(t *testing.T) {
	req := gymRequest(t, 2, "leg_press", "leg_extension")
	in, err := userInput(RolePlanner, req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(in), `"rotation":false`) {
		t.Fatalf("planner input lacks \"rotation\":false")
	}
	res, err := (&Client{Provider: &Fake{}, Cap: bigCap()}).Plan(context.Background(), req)
	if err != nil || res.Status != StatusOK {
		t.Fatalf("status %q, cause %q, err %v", res.Status, res.Cause, err)
	}
	if l := layout(res.Plan); len(l[0]) != 2 || len(l[1]) != 2 {
		t.Fatalf("layout %v: want each exercise in each session", l)
	}
}
