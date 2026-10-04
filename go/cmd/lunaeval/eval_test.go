package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

func caseByID(t *testing.T, id string) Case {
	t.Helper()
	for _, s := range Scenarios() {
		for _, c := range s.Cases {
			if c.ID == id {
				return c
			}
		}
	}
	t.Fatalf("no case %q", id)
	return Case{}
}

func rulesOf(t *testing.T, c Case) domain.PlannedExercise {
	t.Helper()
	d, err := policy.Next(c.Input)
	if err != nil {
		t.Fatalf("%s: Next: %v", c.ID, err)
	}
	return d.Target
}

// change gives a copy of a target with a change to each working set.
func change(p domain.PlannedExercise, f func(*domain.WorkingSet)) domain.PlannedExercise {
	p.Working = append([]domain.WorkingSet(nil), p.Working...)
	for i := range p.Working {
		f(&p.Working[i])
	}
	return p
}

// Each scenario holds cases with another exercise each, and the target
// of the rules alone is safe for each case. So a fallback is safe.
func TestScenarioCases(t *testing.T) {
	ids := map[string]bool{}
	for _, s := range Scenarios() {
		seen := map[domain.ExerciseID]bool{}
		if len(s.Cases) < 2 {
			t.Errorf("scenario %s: %d cases, want 2 or more", s.ID, len(s.Cases))
		}
		for _, c := range s.Cases {
			if seen[c.Input.Exercise.ID] || ids[c.ID] {
				t.Errorf("scenario %s: case %s repeats an exercise or an id", s.ID, c.ID)
			}
			seen[c.Input.Exercise.ID], ids[c.ID] = true, true
			rules := rulesOf(t, c)
			if why := c.Safe(rules, rules); why != "" {
				t.Errorf("%s: the rules target %s is not safe: %s", c.ID, render(rules), why)
			}
		}
	}
	if got := len(Scenarios()); got != 6 {
		t.Fatalf("%d scenarios, want A to F", got)
	}
}

// Each grade finds a target that breaks its scenario, also a target
// that the policy accepts.
func TestGradesFindUnsafeTargets(t *testing.T) {
	lb := domain.Pounds
	for _, tc := range []struct {
		id string
		f  func(*domain.WorkingSet)
	}{
		{"a_low_rir", func(s *domain.WorkingSet) { s.Reps = 12 }},
		{"a_low_rir", func(s *domain.WorkingSet) { s.Load = lb(30) }},
		{"a_twice", func(s *domain.WorkingSet) { s.Load = lb(25) }},
		{"a_pain", func(s *domain.WorkingSet) { s.Reps = 13 }},
		{"b_load_step", func(s *domain.WorkingSet) { s.Reps = 12 }},
		{"b_load_step", func(s *domain.WorkingSet) { s.Load = lb(35) }},
		{"b_add_reps", func(s *domain.WorkingSet) { s.Reps = 13 }},
		{"c_pain_hold", func(s *domain.WorkingSet) { s.RIR = 1 }},
		{"c_pain_easy", func(s *domain.WorkingSet) { s.Load = lb(65) }},
		{"d_unlogged", func(s *domain.WorkingSet) { s.Load = lb(20) }},
		{"d_skipped", func(s *domain.WorkingSet) { s.Reps = 14 }},
		{"e_short", func(s *domain.WorkingSet) { s.RIR = 2 }},
		{"e_long", func(s *domain.WorkingSet) { s.Load += lb(5) }},
		{"f_press", func(s *domain.WorkingSet) { s.Load = lb(150) }},
	} {
		c := caseByID(t, tc.id)
		rules := rulesOf(t, c)
		if why := c.Safe(change(rules, tc.f), rules); why == "" {
			t.Errorf("%s: the grade accepts an unsafe target", tc.id)
		}
	}
}

// fakeOutput changes each set of the echo reply of the fake.
func fakeOutput(f func(set map[string]any)) func(ai.Call) (ai.Reply, error) {
	return func(c ai.Call) (ai.Reply, error) {
		r, err := ai.EchoReply(c)
		if err != nil {
			return r, err
		}
		var out map[string]any
		if err := json.Unmarshal([]byte(r.Text), &out); err != nil {
			return r, err
		}
		// A reviser output holds reasons alone, so it stays as it is.
		sessions, _ := out["sessions"].([]any)
		for _, s := range sessions {
			for _, e := range s.(map[string]any)["exercises"].([]any) {
				for _, key := range []string{"calibration_sets", "working_sets"} {
					for _, w := range e.(map[string]any)[key].([]any) {
						f(w.(map[string]any))
					}
				}
			}
		}
		b, _ := json.Marshal(out)
		r.Text = string(b)
		return r, nil
	}
}

func evaluate(t *testing.T, p ai.Provider, caps ai.Caps, timeout time.Duration) Report {
	t.Helper()
	profiles, err := Profiles()
	if err != nil {
		t.Fatal(err)
	}
	c := &ai.Client{Provider: p, Cap: ai.NewMemoryCap(caps), Timeout: timeout}
	rep, err := Run(context.Background(), c, profiles, Scenarios(), 2, 3)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return rep
}

var twoUSD = ai.Caps{User: 2 * ai.USD, Project: 2 * ai.USD}

// The echo reply of the fake gives the targets of the rules, so the
// policy accepts each planner proposal, and each scenario passes. The
// echo reason names set 1, so each case with a logged working set gets
// the reason of Luna. The skipped case of scenario D gets the reason of
// the rules (D-288).
func TestRunEcho(t *testing.T) {
	rep := evaluate(t, &ai.Fake{}, twoUSD, 0)
	tt := rep.Totals
	if tt.Calls != 20+6*2 || tt.SchemaPass != tt.Calls || tt.Refused != 0 || tt.NoProposal != 0 {
		t.Fatalf("totals %+v", tt)
	}
	if tt.Decisions != tt.Accepted || tt.Cost <= 0 || tt.CostUnknown != 0 {
		t.Fatalf("totals %+v", tt)
	}
	if tt.Revisions != 2*15 || tt.LunaReasons != tt.Revisions-2 || tt.ReasonCauses["no-reason"] != 2 {
		t.Fatalf("revision totals %d, %d, %v", tt.Revisions, tt.LunaReasons, tt.ReasonCauses)
	}
	for _, s := range rep.Scenarios {
		if !s.Pass() || s.Cases == 0 {
			t.Errorf("scenario %+v: want a pass", s)
		}
		if s.ID == "F" && (s.Jumps != s.Cases || s.JumpsOK != s.Jumps) {
			t.Errorf("scenario F %+v: want each jump refused", s)
		}
	}
	if rep.PromptHashes["planner"] != ai.PromptHash(ai.Planner()) || rep.PolicyVersion != policy.Version {
		t.Errorf("report %+v: want the prompt hashes and the policy version", rep)
	}
}

// A load jump of 50 percent breaks the load ceiling. The policy refuses
// each planner proposal, and the rules target applies. A revision has
// no proposal, so each scenario passes with the targets of the rules.
func TestRunJump(t *testing.T) {
	rep := evaluate(t, &ai.Fake{Reply: fakeOutput(func(s map[string]any) {
		s["load_lb"] = s["load_lb"].(float64) * 1.5
	})}, twoUSD, 0)
	tt := rep.Totals
	if tt.Refused != tt.Decisions || tt.Accepted != 0 || tt.ByRule[string(policy.RuleLoadCeiling)] == 0 {
		t.Fatalf("totals %+v: want each proposal refused for %s", tt, policy.RuleLoadCeiling)
	}
	for _, s := range rep.Scenarios {
		if !s.Pass() {
			t.Errorf("scenario %+v: want a pass", s)
		}
	}
}

// One more rep is valid in a calibration session (D-177), and the
// planner plans calibration sessions alone, so the policy accepts each
// planner proposal. Each scenario passes with the targets of the rules.
func TestRunMoreReps(t *testing.T) {
	rep := evaluate(t, &ai.Fake{Reply: fakeOutput(func(s map[string]any) {
		s["reps"] = s["reps"].(float64) + 1
	})}, twoUSD, 0)
	for _, s := range rep.Scenarios {
		if !s.Pass() {
			t.Errorf("scenario %+v: want a pass", s)
		}
	}
	for _, c := range rep.Calls {
		for _, d := range c.Decisions {
			if c.Role == ai.RolePlanner && d.Source != string(policy.SourceLuna) {
				t.Errorf("%s %s: a calibration session with more reps gave %s", c.Item, d.Exercise, d.Cause)
			}
		}
	}
}

// A malformed output, a time-out, and the cap each give the rules
// fallback, with no Luna proposal.
func TestRunFailures(t *testing.T) {
	for _, tc := range []struct {
		name    string
		p       *ai.Fake
		caps    ai.Caps
		timeout time.Duration
		status  ai.Status
	}{
		{"malformed", &ai.Fake{Reply: func(ai.Call) (ai.Reply, error) { return ai.Reply{Text: "{}"}, nil }}, twoUSD, 0, ai.StatusMalformed},
		{"timeout", &ai.Fake{Hang: true}, twoUSD, time.Millisecond, ai.StatusTimeout},
		{"capped", &ai.Fake{}, ai.Caps{}, 0, ai.StatusCapped},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rep := evaluate(t, tc.p, tc.caps, tc.timeout)
			tt := rep.Totals
			if tt.ByStatus[tc.status] != tt.Calls || tt.SchemaPass != 0 || tt.NoProposal != tt.Decisions || tt.Proposals != 0 {
				t.Fatalf("totals %+v: want each call %s with the fallback", tt, tc.status)
			}
			for _, s := range rep.Scenarios {
				if !s.Pass() {
					t.Errorf("scenario %+v: the fallback must be safe", s)
				}
			}
			if tc.status == ai.StatusCapped && (tt.Cost != 0 || len(tc.p.Calls()) != 0) {
				t.Errorf("a capped run cost %s and made %d calls", tt.Cost, len(tc.p.Calls()))
			}
		})
	}
}

// Each profile gives a valid planner request with each exercise of the
// catalog.
func TestProfiles(t *testing.T) {
	ps, err := Profiles()
	if err != nil || len(ps) != 20 {
		t.Fatalf("%d profiles, %v: want 20", len(ps), err)
	}
	for _, p := range ps {
		r := p.Request()
		if len(r.Exercises) == 0 {
			t.Errorf("%s: no exercise", p.ID)
		}
		for _, in := range r.Exercises {
			if _, err := policy.Next(in); err != nil {
				t.Errorf("%s: %v", p.ID, err)
			}
		}
		for _, c := range r.Cardio {
			if e, ok := domain.DefaultCatalog().Exercise(c); !ok || e.Kind != domain.KindCardio {
				t.Errorf("%s: cardio %q is not a cardio exercise", p.ID, c)
			}
		}
	}
}

// The command writes the report and the counts. A live run needs the
// key, and each run needs a cap.
func TestCommand(t *testing.T) {
	out := filepath.Join(t.TempDir(), "report.json")
	env := func(string) string { return "" }
	var stdout bytes.Buffer
	if err := run(context.Background(), []string{"-cap", "2", "-repeats", "1", "-out", out}, env, &stdout); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var rep Report
	if err := json.Unmarshal(b, &rep); err != nil || rep.Provider != "fake" || rep.Cap != "2 USD" || rep.Totals.Calls != 26 {
		t.Fatalf("report %+v, %v", rep.Totals, err)
	}
	if rep.Effort != ai.Planner().Effort {
		t.Errorf("effort %q: want the effort of the role, %q", rep.Effort, ai.Planner().Effort)
	}
	if !strings.Contains(stdout.String(), "scenario F: pass true") {
		t.Errorf("summary:\n%s", stdout.String())
	}

	// The -effort flag sets the effort of the report and of each call.
	stdout.Reset()
	if err := run(context.Background(), []string{"-cap", "2", "-repeats", "1", "-effort", "xhigh", "-out", out}, env, &stdout); err != nil {
		t.Fatal(err)
	}
	if b, err = os.ReadFile(out); err != nil {
		t.Fatal(err)
	}
	rep = Report{}
	if err := json.Unmarshal(b, &rep); err != nil || rep.Effort != "xhigh" {
		t.Fatalf("effort %q, %v: want xhigh", rep.Effort, err)
	}
	if !strings.Contains(stdout.String(), "effort xhigh") {
		t.Errorf("summary:\n%s", stdout.String())
	}

	for _, args := range [][]string{
		{"-live", "-cap", "2", "-out", out},
		{"-out", out},
		{"-cap", "2"},
		{"-cap", "two", "-out", out},
		{"-cap", "2", "-effort", "extra-high", "-out", out},
	} {
		if err := run(context.Background(), args, env, &bytes.Buffer{}); err == nil {
			t.Errorf("run %v: want an error", args)
		}
	}
}
