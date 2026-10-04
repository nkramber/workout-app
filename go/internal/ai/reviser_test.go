package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// reviseRequest gives a reviser request with the press history and the
// new squat.
func reviseRequest(t testing.TB) Request {
	return Request{User: "uid-test-1", Today: "2026-10-01", Sessions: 1, Exercises: []policy.Input{pressInput(t), squatInput(t)}}
}

// TestReviseEcho: the reviser call sends the reason schema, and the
// echo of the fake gives a reason for the press alone, which names its
// set 1. The squat has no history, so it gets no reason (D-288).
func TestReviseEcho(t *testing.T) {
	fake := &Fake{}
	c := &Client{Provider: fake, Cap: bigCap()}
	res, err := c.Revise(context.Background(), reviseRequest(t))
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusOK || res.Plan != nil || res.Role != RoleReviser || res.PromptHash != PromptHash(Reviser()) {
		t.Fatalf("result %+v", res)
	}
	want := []Reason{{Exercise: "chest_press", Sets: []SetRef{{domain.SetWorking, 1}}, Text: "Set 1: 10 reps at 100 lb, 3 in reserve."}}
	if !slices.EqualFunc(res.Reasons, want, func(a, b Reason) bool {
		return a.Exercise == b.Exercise && slices.Equal(a.Sets, b.Sets) && a.Text == b.Text && a.Blocked == b.Blocked
	}) {
		t.Fatalf("reasons %+v, want %+v", res.Reasons, want)
	}
	calls := fake.Calls()
	if len(calls) != 1 || calls[0].SchemaName != ReasonSchemaName || string(calls[0].Schema) != string(ReasonSchema()) || calls[0].Role.Timeout != ReviserTimeout {
		t.Fatalf("calls %+v", calls)
	}
}

// TestReviserInput: the reviser input holds the target that the owner
// saw, each logged set with its number inside its kind, the next target
// of the rules, and the rules with their reason. It holds no note.
func TestReviserInput(t *testing.T) {
	req := reviseRequest(t)
	cal := domain.PlannedExercise{Exercise: "chest_press", RestSeconds: 60,
		Calibration: []domain.CalibrationSet{{Reps: 10, Load: domain.Pounds(100)}},
		Working:     []domain.WorkingSet{{Reps: 10, Load: domain.Pounds(100), RIR: 3}}}
	last := &req.Exercises[0].History[2]
	last.Target = cal
	last.Log.Sets = append([]domain.SetLog{{Kind: domain.SetCalibration, Reps: 10, Weight: domain.Pounds(100), RIR: 4}}, last.Log.Sets...)
	input, err := userInput(RoleReviser, req)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(input), "private note") {
		t.Fatal("the input holds a note of the owner")
	}
	var in wireInput
	if err := json.Unmarshal(input, &in); err != nil {
		t.Fatal(err)
	}
	press := in.Exercises[0]
	d, _ := policy.Next(req.Exercises[0])
	if press.RulesReason != d.Reason || !slices.Equal(press.Rules, strs(d.Rules)) || press.RulesReason == "" {
		t.Fatalf("rules %v and reason %q, want %v and %q", press.Rules, press.RulesReason, d.Rules, d.Reason)
	}
	h := press.History[2]
	if len(h.Target.Calibration) != 1 || h.Target.Working[0].RIR != 3 {
		t.Fatalf("history target %+v", h.Target)
	}
	var numbers []string
	for _, s := range h.Sets {
		numbers = append(numbers, fmt.Sprintf("%s %d", s.Kind, s.Number))
	}
	if !slices.Equal(numbers, []string{"calibration 1", "working 1", "working 2", "working 3"}) {
		t.Fatalf("set numbers %q", numbers)
	}
	planner, err := userInput(RolePlanner, req)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(planner), "rules_reason") {
		t.Fatal("the planner input holds the reason of the rules")
	}
}

// TestParseReasons: each output that breaks the schema or the request
// is malformed, and the error holds no text of the output. A blocked
// reason has no text. An exercise with no reason is valid.
func TestParseReasons(t *testing.T) {
	const secret = "secret_value_of_luna"
	req := reviseRequest(t)
	good := `{"reasons":[{"exercise_id":"chest_press","logged_sets":[{"kind":"working","number":2}],"reason":"Set 2:  10 reps at 100 lb."}]}`
	r, err := parseReasons(good, req)
	if err != nil || len(r) != 1 || r[0].Text != "Set 2: 10 reps at 100 lb." || r[0].Sets[0] != (SetRef{domain.SetWorking, 2}) {
		t.Fatalf("parse = %+v, %v", r, err)
	}
	if r, err := parseReasons(`{"reasons":[]}`, req); err != nil || len(r) != 0 {
		t.Fatalf("no reason: %+v, %v", r, err)
	}
	blocked := `{"reasons":[{"exercise_id":"chest_press","logged_sets":[{"kind":"working","number":1}],"reason":"This set heals the knee."}]}`
	if r, err := parseReasons(blocked, req); err != nil || r[0].Text != "" || r[0].Blocked != FilterMedical {
		t.Fatalf("blocked: %+v, %v", r, err)
	}
	for _, text := range []string{
		"Here are the reasons: " + secret,
		`{"reasons":[]} ` + secret,
		`{"reasons":[],"` + secret + `":1}`,
		`{}`,
		`{"reasons":[{"exercise_id":"leg_press","logged_sets":[],"reason":"` + secret + `"}]}`,
		`{"reasons":[{"exercise_id":"chest_press","logged_sets":[],"reason":"a"},{"exercise_id":"chest_press","logged_sets":[],"reason":"b"}]}`,
		`{"reasons":[{"exercise_id":"chest_press","logged_sets":[{"kind":"` + secret + `","number":1}],"reason":"a"}]}`,
		`{"reasons":[{"exercise_id":"chest_press","logged_sets":[{"kind":"working"}],"reason":"a"}]}`,
		`{"reasons":[{"exercise_id":"chest_press","reason":"a"}]}`,
	} {
		_, err := parseReasons(text, req)
		if !errors.Is(err, ErrMalformed) || strings.Contains(err.Error(), secret) {
			t.Errorf("parse of %q: %v, want ErrMalformed with no text of the output", text, err)
		}
	}
}

// TestReviseMalformed: an output in the form of the plan schema is
// malformed for the reviser, so the reason of the rules shows.
func TestReviseMalformed(t *testing.T) {
	fake := &Fake{Reply: func(c Call) (Reply, error) {
		c.SchemaName = SchemaName
		return EchoReply(c)
	}}
	res, err := (&Client{Provider: fake, Cap: bigCap()}).Revise(context.Background(), reviseRequest(t))
	if err != nil || res.Status != StatusMalformed || res.Reasons != nil {
		t.Fatalf("result %+v, %v", res, err)
	}
}

// TestReasonSchema: the schema is strict, and it holds no target.
func TestReasonSchema(t *testing.T) {
	text := string(ReasonSchema())
	for _, want := range []string{`"additionalProperties":false`, `"logged_sets"`, `"reason"`, `"chest_press"`} {
		if !strings.Contains(text, want) {
			t.Errorf("the schema lacks %s", want)
		}
	}
	for _, bad := range []string{"load_lb", "working_sets", "treadmill"} {
		if strings.Contains(text, bad) {
			t.Errorf("the schema holds %s", bad)
		}
	}
}
