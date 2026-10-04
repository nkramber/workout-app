package ai

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// output gives the echo output of the request of the tests, after a
// change.
func output(t *testing.T, change func(out map[string]any)) string {
	t.Helper()
	req := request(t)
	input, err := userInput(RolePlanner, req)
	if err != nil {
		t.Fatal(err)
	}
	r, err := EchoReply(Call{Input: input})
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(r.Text), &out); err != nil {
		t.Fatal(err)
	}
	change(out)
	return string(encode(out))
}

func session0(out map[string]any) map[string]any {
	return out["sessions"].([]any)[0].(map[string]any)
}

// TestParseMalformed: each output that breaks the schema or the request
// is malformed. The error holds no text and no value of the output.
func TestParseMalformed(t *testing.T) {
	const secret = "secret_value_of_luna"
	for _, tc := range []struct {
		name   string
		change func(out map[string]any)
		text   string
	}{
		{name: "not JSON", text: "Here is your plan: " + secret},
		{name: "text after the JSON", text: `{"summary":"a","sessions":[],"guidance_ids":[]} ` + secret},
		{name: "two objects", text: `{} {}`},
		{name: "array", text: `[]`},
		{name: "no summary", change: func(o map[string]any) { delete(o, "summary") }},
		{name: "no guidance", change: func(o map[string]any) { delete(o, "guidance_ids") }},
		{name: "unknown field", change: func(o map[string]any) { o[secret] = 1 }},
		{name: "no session", change: func(o map[string]any) { o["sessions"] = []any{} }},
		{name: "two sessions", change: func(o map[string]any) { o["sessions"] = append(o["sessions"].([]any), session0(o)) }},
		{name: "no warm-up", change: func(o map[string]any) { delete(session0(o), "warm_up_id") }},
		{name: "unknown warm-up", change: func(o map[string]any) { session0(o)["warm_up_id"] = secret }},
		{name: "cool-down as warm-up", change: func(o map[string]any) { session0(o)["warm_up_id"] = "cool_down.easy_walk" }},
		{name: "warm-up as cool-down", change: func(o map[string]any) { session0(o)["cool_down_id"] = "warm_up.light_sets" }},
		{name: "warm-up as guidance", change: func(o map[string]any) { o["guidance_ids"] = []any{"warm_up.light_sets"} }},
		{name: "unknown guidance", change: func(o map[string]any) { o["guidance_ids"] = []any{secret} }},
		{name: "unknown exercise", change: func(o map[string]any) { exerciseAt(o, 0, 0)["exercise_id"] = secret }},
		{name: "exercise outside the request", change: func(o map[string]any) { exerciseAt(o, 0, 0)["exercise_id"] = "leg_press" }},
		{name: "exercise two times", change: func(o map[string]any) {
			s := session0(o)
			s["exercises"] = append(s["exercises"].([]any), exerciseAt(o, 0, 0))
		}},
		{name: "no reason", change: func(o map[string]any) { delete(exerciseAt(o, 0, 0), "reason") }},
		{name: "no rest", change: func(o map[string]any) { delete(exerciseAt(o, 0, 0), "rest_seconds") }},
		{name: "no working sets", change: func(o map[string]any) { delete(exerciseAt(o, 0, 0), "working_sets") }},
		{name: "no calibration sets", change: func(o map[string]any) { delete(exerciseAt(o, 0, 1), "calibration_sets") }},
		{name: "rir on a calibration set", change: func(o map[string]any) {
			exerciseAt(o, 0, 1)["calibration_sets"].([]any)[0].(map[string]any)["rir_target"] = 3
		}},
		{name: "no rir", change: func(o map[string]any) {
			delete(exerciseAt(o, 0, 0)["working_sets"].([]any)[0].(map[string]any), "rir_target")
		}},
		{name: "no calibration load", change: func(o map[string]any) {
			delete(exerciseAt(o, 0, 1)["calibration_sets"].([]any)[0].(map[string]any), "load_lb")
		}},
		{name: "reps not whole", change: func(o map[string]any) {
			exerciseAt(o, 0, 0)["working_sets"].([]any)[0].(map[string]any)["reps"] = 10.5
		}},
		{name: "load not in tenths", change: func(o map[string]any) {
			exerciseAt(o, 0, 0)["working_sets"].([]any)[0].(map[string]any)["load_lb"] = 100.25
		}},
		{name: "calibration load not in tenths", change: func(o map[string]any) {
			exerciseAt(o, 0, 1)["calibration_sets"].([]any)[0].(map[string]any)["load_lb"] = 5.05
		}},
		{name: "load out of range", change: func(o map[string]any) {
			exerciseAt(o, 0, 0)["working_sets"].([]any)[0].(map[string]any)["load_lb"] = 1e300
		}},
		{name: "load as text", change: func(o map[string]any) {
			exerciseAt(o, 0, 0)["working_sets"].([]any)[0].(map[string]any)["load_lb"] = "100"
		}},
		{name: "no cardio", change: func(o map[string]any) { delete(session0(o), "cardio") }},
		{name: "cardio with no minutes", change: func(o map[string]any) { session0(o)["cardio"] = map[string]any{"exercise_id": ""} }},
		{name: "cardio outside the request", change: func(o map[string]any) {
			session0(o)["cardio"] = map[string]any{"exercise_id": "rowing_machine", "minutes": 10}
		}},
		{name: "strength as cardio", change: func(o map[string]any) {
			session0(o)["cardio"] = map[string]any{"exercise_id": "chest_press", "minutes": 10}
		}},
		{name: "cardio of 0 minutes", change: func(o map[string]any) {
			session0(o)["cardio"] = map[string]any{"exercise_id": "treadmill", "minutes": 0}
		}},
		{name: "minutes and no cardio", change: func(o map[string]any) {
			session0(o)["cardio"] = map[string]any{"exercise_id": "", "minutes": 10}
		}},
		{name: "empty session", change: func(o map[string]any) {
			session0(o)["exercises"] = []any{}
			session0(o)["cardio"] = map[string]any{"exercise_id": "", "minutes": 0}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text := tc.text
			if tc.change != nil {
				text = output(t, tc.change)
			}
			_, err := parse(text, request(t))
			if !errors.Is(err, ErrMalformed) {
				t.Fatalf("err %v: want ErrMalformed", err)
			}
			if strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), "chest_press") {
				t.Fatalf("the error holds a value of the output: %v", err)
			}
		})
	}
}

// TestParseValid: a valid output keeps each value of Luna. The policy,
// not the parser, refuses a value out of bounds (D-23).
func TestParseValid(t *testing.T) {
	text := output(t, func(o map[string]any) {
		w := exerciseAt(o, 0, 0)["working_sets"].([]any)
		w[0].(map[string]any)["load_lb"] = 12.5
		w[1].(map[string]any)["reps"] = -3
		w[2].(map[string]any)["load_lb"] = 0
		exerciseAt(o, 0, 0)["rest_seconds"] = 5
		exerciseAt(o, 0, 0)["reason"] = "  Two   spaces.\n"
		session0(o)["cardio"] = map[string]any{"exercise_id": "treadmill", "minutes": 22}
		session0(o)["exercises"] = session0(o)["exercises"].([]any)[:1]
	})
	p, err := parse(text, request(t))
	if err != nil {
		t.Fatal(err)
	}
	e := p.Sessions[0].Exercises[0]
	if e.Target.Working[0].Load != domain.Load(125) || e.Target.Working[1].Reps != -3 || e.Target.Working[2].Load != 0 || e.Target.RestSeconds != 5 {
		t.Fatalf("target %+v", e.Target)
	}
	if e.Reason != "Two spaces." {
		t.Fatalf("reason %q", e.Reason)
	}
	if len(p.Sessions[0].Exercises) != 1 || p.Sessions[0].Cardio.Minutes != 22 {
		t.Fatalf("session %+v", p.Sessions[0])
	}
	if len(p.Filtered) != 0 {
		t.Fatalf("filtered %v", p.Filtered)
	}
}

// TestParseCardioOnly: a session can hold cardio alone.
func TestParseCardioOnly(t *testing.T) {
	text := output(t, func(o map[string]any) {
		session0(o)["exercises"] = []any{}
		session0(o)["cardio"] = map[string]any{"exercise_id": "treadmill", "minutes": 20}
	})
	p, err := parse(text, request(t))
	if err != nil || len(p.Sessions[0].Exercises) != 0 || p.Sessions[0].Cardio == nil {
		t.Fatalf("plan %+v, err %v", p, err)
	}
}

// TestEchoReplyFull: the default reply of the fake gives each part of a
// plan, so a browser test can show each part (D-44, D-241). With no
// cardio exercise in the input, it gives no cardio.
func TestEchoReplyFull(t *testing.T) {
	req := request(t)
	req.Sessions = 2
	for _, tc := range []struct {
		name   string
		cardio []domain.ExerciseID
		want   *domain.PlannedCardio
	}{
		{"with cardio", []domain.ExerciseID{"treadmill"}, &domain.PlannedCardio{Exercise: "treadmill", Minutes: EchoCardioMinutes}},
		{"no cardio", nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req.Cardio = tc.cardio
			input, err := userInput(RolePlanner, req)
			if err != nil {
				t.Fatal(err)
			}
			r, err := EchoReply(Call{Input: input})
			if err != nil {
				t.Fatal(err)
			}
			p, err := parse(r.Text, req)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(p.Guidance, EchoGuidance) || len(p.Sessions) != 2 {
				t.Fatalf("plan %+v", p)
			}
			for i, s := range p.Sessions {
				if len(s.Exercises) == 0 || (s.Cardio == nil) != (tc.want == nil) || (s.Cardio != nil && *s.Cardio != *tc.want) {
					t.Fatalf("session %d: %+v", i, s)
				}
			}
		})
	}
}
