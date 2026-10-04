package ai

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

func exerciseOf(t testing.TB, id domain.ExerciseID) domain.Exercise {
	t.Helper()
	e, ok := domain.DefaultCatalog().Exercise(id)
	if !ok {
		t.Fatalf("exercise %q: not in the catalog", id)
	}
	return e
}

// pressInput gives the chest press after 3 logged sessions of 3 sets
// of 10 reps at 100 lb, at 3 reps in reserve, so the calibration is
// done. Each set holds a note, and no note may leave the layer.
func pressInput(t testing.TB) policy.Input {
	e := exerciseOf(t, "chest_press")
	var weights []domain.Load
	for w := int64(10); w <= 150; w += 5 {
		weights = append(weights, domain.Pounds(w))
	}
	in := policy.Input{Exercise: e, Entry: domain.InventoryEntry{Machine: e.Machine, Weights: weights}, Today: "2026-10-01"}
	target := domain.PlannedExercise{Exercise: e.ID, RestSeconds: 120}
	for range 3 {
		target.Working = append(target.Working, domain.WorkingSet{Reps: 10, Load: domain.Pounds(100), RIR: 2})
	}
	for _, d := range []string{"2026-09-22", "2026-09-24", "2026-09-26"} {
		var sets []domain.SetLog
		for range 3 {
			sets = append(sets, domain.SetLog{Kind: domain.SetWorking, Reps: 10, Weight: domain.Pounds(100), RIR: 3, Note: "private note of the owner"})
		}
		in.History = append(in.History, policy.Outcome{Date: d, Target: target, Log: domain.ExerciseLog{Exercise: e.ID, Sets: sets}})
	}
	return in
}

// squatInput gives a new dumbbell exercise with no history and no
// estimate, so the policy target is a calibration from the lightest
// dumbbell (D-178).
func squatInput(t testing.TB) policy.Input {
	e := exerciseOf(t, "db_goblet_squat")
	set := &domain.DumbbellSet{Lightest: domain.Pounds(5), Heaviest: domain.Pounds(50), Step: domain.Pounds(5)}
	return policy.Input{Exercise: e, Entry: domain.InventoryEntry{Machine: e.Machine, Dumbbells: set}, Today: "2026-10-01"}
}

func request(t testing.TB) Request {
	return Request{User: "uid-test-1", Today: "2026-10-01", Sessions: 1,
		Exercises: []policy.Input{pressInput(t), squatInput(t)}, Cardio: []domain.ExerciseID{"treadmill"}}
}

func bigCap() *MemoryCap { return NewMemoryCap(Caps{User: USD, Project: USD}) }

// edit gives a reply func that changes the echo output of a call.
func edit(t testing.TB, change func(out map[string]any)) func(Call) (Reply, error) {
	return func(c Call) (Reply, error) {
		r, err := EchoReply(c)
		if err != nil {
			return r, err
		}
		var out map[string]any
		if err := json.Unmarshal([]byte(r.Text), &out); err != nil {
			t.Fatal(err)
		}
		change(out)
		r.Text = string(encode(out))
		return r, nil
	}
}

// exerciseAt gives exercise j of session i of a decoded output.
func exerciseAt(out map[string]any, i, j int) map[string]any {
	s := out["sessions"].([]any)[i].(map[string]any)
	return s["exercises"].([]any)[j].(map[string]any)
}

// TestAcceptance is the acceptance story of PR-16. The fake provider
// gives a valid proposal, a malformed proposal, an unsafe proposal, and
// a timeout. policy.Decide turns the unsafe proposal into a refusal and
// a fallback target, and each failure into the rules fallback. No test
// calls OpenAI: the fake makes no network call.
func TestAcceptance(t *testing.T) {
	unsafe := edit(t, func(out map[string]any) {
		// A 50 percent load jump on the chest press, at 0 reps in reserve.
		for _, s := range exerciseAt(out, 0, 0)["working_sets"].([]any) {
			s.(map[string]any)["load_lb"] = 150
			s.(map[string]any)["rir_target"] = 0
		}
	})
	type want struct {
		status Status
		source map[domain.ExerciseID]policy.Source
		cause  map[domain.ExerciseID]policy.Cause
	}
	luna := map[domain.ExerciseID]policy.Source{"chest_press": policy.SourceLuna, "db_goblet_squat": policy.SourceLuna}
	rules := map[domain.ExerciseID]policy.Source{"chest_press": policy.SourceRules, "db_goblet_squat": policy.SourceRules}
	none := map[domain.ExerciseID]policy.Cause{"chest_press": policy.CauseNoProposal, "db_goblet_squat": policy.CauseNoProposal}
	for _, tc := range []struct {
		name string
		fake *Fake
		want want
	}{
		{"valid", &Fake{}, want{StatusOK, luna, map[domain.ExerciseID]policy.Cause{"chest_press": policy.CauseNone, "db_goblet_squat": policy.CauseNone}}},
		{"malformed", &Fake{Reply: func(Call) (Reply, error) {
			return Reply{Text: `{"summary": "A plan", "sessions": [`, Usage: Usage{InputTokens: 10, OutputTokens: 5}}, nil
		}}, want{StatusMalformed, rules, none}},
		{"unsafe", &Fake{Reply: unsafe}, want{StatusOK,
			map[domain.ExerciseID]policy.Source{"chest_press": policy.SourceRules, "db_goblet_squat": policy.SourceLuna},
			map[domain.ExerciseID]policy.Cause{"chest_press": policy.CauseRefused, "db_goblet_squat": policy.CauseNone}}},
		{"timeout", &Fake{Hang: true}, want{StatusTimeout, rules, none}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := request(t)
			c := &Client{Provider: tc.fake, Cap: bigCap(), Timeout: 20 * time.Millisecond}
			res, err := c.Plan(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			if res.Status != tc.want.status {
				t.Fatalf("status %q: want %q", res.Status, tc.want.status)
			}
			if n := len(tc.fake.Calls()); n != 1 {
				t.Fatalf("%d calls of the fake: want 1", n)
			}
			if (res.Plan != nil) != (res.Status == StatusOK) {
				t.Fatalf("plan %v with status %q", res.Plan != nil, res.Status)
			}
			for _, in := range req.Exercises {
				id := in.Exercise.ID
				prop := res.Proposal(0, id)
				if prop.Model != Planner().Model || prop.Effort != Planner().Effort || prop.PromptHash != PromptHash(Planner()) {
					t.Errorf("%s: proposal names model %q, effort %q, hash %q", id, prop.Model, prop.Effort, prop.PromptHash)
				}
				rec, err := policy.Decide(in, prop)
				if err != nil {
					t.Fatalf("%s: %v", id, err)
				}
				if rec.Source != tc.want.source[id] || rec.Cause != tc.want.cause[id] {
					t.Fatalf("%s: source %q cause %q: want %q %q, violations %v", id, rec.Source, rec.Cause, tc.want.source[id], tc.want.cause[id], rec.Violations)
				}
				next, err := policy.Next(in)
				if err != nil {
					t.Fatal(err)
				}
				if rec.Source == policy.SourceRules && !equalTarget(rec.Target, next.Target) {
					t.Errorf("%s: fallback target %+v: want the target of the rules %+v", id, rec.Target, next.Target)
				}
				shown := res.Reason(0, rec)
				switch rec.Source {
				case policy.SourceLuna:
					if shown != "The target follows your last logged sets." {
						t.Errorf("%s: reason %q: want the reason of Luna", id, shown)
					}
				default:
					if shown != rec.Reason || shown == "" {
						t.Errorf("%s: reason %q: want the reason of the rules %q", id, shown, rec.Reason)
					}
				}
			}
			if tc.name == "unsafe" {
				rec, _ := policy.Decide(req.Exercises[0], res.Proposal(0, "chest_press"))
				var ids []policy.RuleID
				for _, v := range rec.Violations {
					ids = append(ids, v.Rule)
				}
				for _, want := range []policy.RuleID{policy.RuleRIRBounds, policy.RuleLoadCeiling} {
					if !slices.Contains(ids, want) {
						t.Errorf("violations %v: want %s", ids, want)
					}
				}
				if rec.Rules[0] != policy.RuleFallback {
					t.Errorf("rules %v: want %s first", rec.Rules, policy.RuleFallback)
				}
			}
		})
	}
}

func equalTarget(a, b domain.PlannedExercise) bool {
	return a.Exercise == b.Exercise && a.RestSeconds == b.RestSeconds &&
		slices.Equal(a.Calibration, b.Calibration) && slices.Equal(a.Working, b.Working)
}

// TestCapRefuses: a call over the cap sends nothing, and the policy
// gives the rules fallback with no model in the record (D-25).
func TestCapRefuses(t *testing.T) {
	fake := &Fake{}
	var recs []CostRecord
	c := &Client{Provider: fake, Cap: NewMemoryCap(Caps{User: 0, Project: USD}), Record: func(r CostRecord) { recs = append(recs, r) }}
	req := request(t)
	res, err := c.Plan(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusCapped || len(fake.Calls()) != 0 {
		t.Fatalf("status %q with %d calls: want %q with no call", res.Status, len(fake.Calls()), StatusCapped)
	}
	if p := res.Proposal(0, "chest_press"); p != (policy.Proposal{}) {
		t.Fatalf("proposal %+v: want a zero proposal", p)
	}
	rec, err := policy.Decide(req.Exercises[0], res.Proposal(0, "chest_press"))
	if err != nil {
		t.Fatal(err)
	}
	if rec.Source != policy.SourceRules || rec.Cause != policy.CauseNoProposal || rec.Model != "" {
		t.Fatalf("record source %q cause %q model %q", rec.Source, rec.Cause, rec.Model)
	}
	if len(recs) != 1 || recs[0].Status != StatusCapped || recs[0].Cost != 0 || recs[0].Reserved != 0 {
		t.Fatalf("cost records %+v: want one capped record with no cost", recs)
	}
}

// TestCapProject: the project cap refuses the call of a second user.
func TestCapProject(t *testing.T) {
	worst := Planner().worst(1)
	cap := NewMemoryCap(Caps{User: USD, Project: 1})
	if _, err := cap.Reserve(context.Background(), "a", worst); !errors.Is(err, ErrCap) {
		t.Fatalf("err %v: want ErrCap", err)
	}
}

// TestCostRecord: each call gives one cost record. A known usage gives
// its cost, and a failed call keeps the worst case in the cap.
func TestCostRecord(t *testing.T) {
	usage := Usage{InputTokens: 4000, CachedInputTokens: 3000, OutputTokens: 2000, ReasoningTokens: 800}
	for _, tc := range []struct {
		name   string
		fake   *Fake
		status Status
		known  bool
	}{
		{"ok", &Fake{Reply: func(c Call) (Reply, error) { r, err := EchoReply(c); r.Usage = usage; return r, err }}, StatusOK, true},
		{"refusal", &Fake{Reply: func(Call) (Reply, error) { return Reply{Refusal: true, Usage: usage}, nil }}, StatusRefusal, true},
		{"incomplete", &Fake{Reply: func(Call) (Reply, error) { return Reply{Incomplete: true, Usage: usage}, nil }}, StatusIncomplete, true},
		{"error", &Fake{Reply: func(Call) (Reply, error) { return Reply{}, errors.New("openai: HTTP 500") }}, StatusError, false},
		{"timeout", &Fake{Hang: true}, StatusTimeout, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var recs []CostRecord
			cap := bigCap()
			c := &Client{Provider: tc.fake, Cap: cap, Timeout: 20 * time.Millisecond, Record: func(r CostRecord) { recs = append(recs, r) }}
			res, err := c.Plan(context.Background(), request(t))
			if err != nil {
				t.Fatal(err)
			}
			if res.Status != tc.status || len(recs) != 1 || recs[0] != res.Cost {
				t.Fatalf("status %q, %d records: want %q and one record", res.Status, len(recs), tc.status)
			}
			r := recs[0]
			if r.User != "uid-test-1" || r.Role != RolePlanner || r.Model != Planner().Model || r.Effort != Planner().Effort || r.Status != tc.status || r.Known != tc.known {
				t.Fatalf("record %+v", r)
			}
			want := r.Reserved
			if tc.known {
				want = Planner().Prices.Cost(usage)
				// 1000 fresh input, 3000 cached, and 2000 output tokens.
				if want != 1000*100+3000*10+2000*500 {
					t.Fatalf("cost %s", want)
				}
			}
			if r.Cost != want || r.Reserved <= 0 {
				t.Fatalf("cost %s reserved %s: want cost %s", r.Cost, r.Reserved, want)
			}
			if u, p := cap.Spent("uid-test-1"); u != want || p != want {
				t.Fatalf("spent %s and %s: want %s", u, p, want)
			}
			b, _ := json.Marshal(r)
			if strings.Contains(string(b), "logged sets") || strings.Contains(string(b), "private note") {
				t.Fatalf("the cost record holds text: %s", b)
			}
		})
	}
}

// failingCap is a cap hook whose store fails. With reserveErr, Reserve
// fails. Otherwise settle fails.
type failingCap struct{ reserveErr error }

func (f failingCap) Reserve(context.Context, string, NanoUSD) (func(NanoUSD) error, error) {
	if f.reserveErr != nil {
		return nil, f.reserveErr
	}
	return func(NanoUSD) error { return errors.New("store: unavailable") }, nil
}

// TestCapStoreFailure: a Reserve that fails stops the call, and the
// client sends nothing. A settle that fails marks the record, and the
// reservation stays in the cap.
func TestCapStoreFailure(t *testing.T) {
	fake := &Fake{}
	down := errors.New("store: unavailable")
	_, err := (&Client{Provider: fake, Cap: failingCap{reserveErr: down}}).Plan(context.Background(), request(t))
	if !errors.Is(err, down) || len(fake.Calls()) != 0 {
		t.Fatalf("err %v with %d calls: want the store error and no call", err, len(fake.Calls()))
	}
	for _, tc := range []struct {
		name string
		fake *Fake
	}{
		{"ok", &Fake{}},
		{"error", &Fake{Reply: func(Call) (Reply, error) { return Reply{}, errors.New("openai: HTTP 500") }}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res, err := (&Client{Provider: tc.fake, Cap: failingCap{}}).Plan(context.Background(), request(t))
			if err != nil || !res.Cost.Unsettled {
				t.Fatalf("err %v, record %+v: want an unsettled record", err, res.Cost)
			}
		})
	}
}

// TestFilterReplacesText: a blocked claim of Luna never reaches the
// owner. The template text replaces it, and the plan names the rule
// (D-183).
func TestFilterReplacesText(t *testing.T) {
	fake := &Fake{Reply: edit(t, func(out map[string]any) {
		out["summary"] = "This plan will cure your back pain."
		exerciseAt(out, 0, 0)["reason"] = "Heavy presses heal the shoulder injury."
		exerciseAt(out, 0, 1)["reason"] = strings.Repeat("Long text. ", 20)
	})}
	req := request(t)
	res, err := (&Client{Provider: fake, Cap: bigCap()}).Plan(context.Background(), req)
	if err != nil || res.Status != StatusOK {
		t.Fatalf("status %q, err %v", res.Status, err)
	}
	if res.Plan.Summary != SummaryTemplate {
		t.Errorf("summary %q: want the template", res.Plan.Summary)
	}
	want := []Filtered{
		{"summary", FilterMedical},
		{"sessions[0].exercises[0].reason", FilterMedical},
		{"sessions[0].exercises[1].reason", FilterLength},
	}
	if !slices.Equal(res.Plan.Filtered, want) {
		t.Fatalf("filtered %v: want %v", res.Plan.Filtered, want)
	}
	rec, err := policy.Decide(req.Exercises[0], res.Proposal(0, "chest_press"))
	if err != nil || rec.Source != policy.SourceLuna {
		t.Fatalf("source %q, err %v", rec.Source, err)
	}
	if got := res.Reason(0, rec); got != ReasonTemplate {
		t.Fatalf("reason %q: want the template", got)
	}
}

// TestPlanShape: the plan takes the session title from a template, the
// guidance from the catalog, and the cardio of the output.
func TestPlanShape(t *testing.T) {
	fake := &Fake{Reply: edit(t, func(out map[string]any) {
		out["guidance_ids"] = []any{"mobility.hips", "recovery.rest_day"}
		s := out["sessions"].([]any)[0].(map[string]any)
		s["cardio"] = map[string]any{"exercise_id": "treadmill", "minutes": 25}
		s["cool_down_id"] = "cool_down.stretch"
		out["sessions"].([]any)[1].(map[string]any)["cardio"] = map[string]any{"exercise_id": "treadmill", "minutes": 30}
	})}
	req := request(t)
	req.Sessions = 2
	res, err := (&Client{Provider: fake, Cap: bigCap()}).Plan(context.Background(), req)
	if err != nil || res.Status != StatusOK {
		t.Fatalf("status %q, err %v", res.Status, err)
	}
	p := res.Plan
	if len(p.Sessions) != 2 || p.Sessions[0].Title != "Session 1" || p.Sessions[1].Title != "Session 2" {
		t.Fatalf("sessions %+v", p.Sessions)
	}
	if p.Sessions[0].WarmUp != DefaultWarmUp || p.Sessions[0].CoolDown != "cool_down.stretch" {
		t.Fatalf("session 0 guidance %q %q", p.Sessions[0].WarmUp, p.Sessions[0].CoolDown)
	}
	if c := p.Sessions[0].Cardio; c == nil || *c != (domain.PlannedCardio{Exercise: "treadmill", Minutes: 25}) ||
		p.Sessions[1].Cardio == nil || p.Sessions[1].Cardio.Minutes != 30 {
		t.Fatalf("cardio %+v %+v", p.Sessions[0].Cardio, p.Sessions[1].Cardio)
	}
	if !slices.Equal(p.Guidance, []GuidanceID{"mobility.hips", "recovery.rest_day"}) {
		t.Fatalf("guidance %v", p.Guidance)
	}
	if got := res.Proposal(1, "chest_press"); got.Target == nil {
		t.Fatal("session 1 gives no proposal")
	}
	for _, s := range []int{-1, 2} {
		if got := res.Proposal(s, "chest_press"); got.Target != nil {
			t.Fatalf("session %d gives a proposal", s)
		}
	}
	if got := res.Proposal(0, "leg_press"); got.Target != nil {
		t.Fatal("an exercise outside the plan gives a proposal")
	}
}

// TestProposalCopy: a change of a proposal does not change the plan.
func TestProposalCopy(t *testing.T) {
	res, err := (&Client{Provider: &Fake{}, Cap: bigCap()}).Plan(context.Background(), request(t))
	if err != nil {
		t.Fatal(err)
	}
	p := res.Proposal(0, "chest_press")
	p.Target.Working[0].Reps = 99
	if res.Proposal(0, "chest_press").Target.Working[0].Reps == 99 {
		t.Fatal("the proposal shares the sets of the plan")
	}
}

// TestRevise: the reviser makes one session, with its own role.
func TestRevise(t *testing.T) {
	fake := &Fake{}
	res, err := (&Client{Provider: fake, Cap: bigCap()}).Revise(context.Background(), request(t))
	if err != nil || res.Status != StatusOK || res.Role != RoleReviser {
		t.Fatalf("status %q role %q err %v", res.Status, res.Role, err)
	}
	c := fake.Calls()[0]
	if c.Role != Reviser() || c.Instructions != Instructions(Reviser()) || res.PromptHash != PromptHash(Reviser()) {
		t.Fatal("the call does not use the reviser")
	}
}

// TestEffort: the effort of the client replaces the effort of the role
// in the call, the result, and the cost record. The prompt and its hash
// do not change.
func TestEffort(t *testing.T) {
	fake := &Fake{}
	var rec CostRecord
	c := &Client{Provider: fake, Cap: bigCap(), Effort: "high", Record: func(r CostRecord) { rec = r }}
	res, err := c.Plan(context.Background(), request(t))
	if err != nil || res.Status != StatusOK {
		t.Fatalf("status %q err %v", res.Status, err)
	}
	call := fake.Calls()[0]
	if call.Role.Effort != "high" || res.Effort != "high" || rec.Effort != "high" {
		t.Fatalf("effort: call %q, result %q, record %q: want high", call.Role.Effort, res.Effort, rec.Effort)
	}
	if call.Instructions != Instructions(Planner()) || res.PromptHash != PromptHash(Planner()) {
		t.Fatal("the effort changed the prompt")
	}
	if Planner().Effort == "high" {
		t.Fatal("the client changed the role")
	}
	for _, e := range Efforts {
		if !ValidEffort(e) {
			t.Errorf("ValidEffort(%q) = false", e)
		}
	}
	for _, e := range []string{"extra-high", "Medium", "minimal"} {
		if ValidEffort(e) {
			t.Errorf("ValidEffort(%q) = true", e)
		}
	}
}

// TestRequestErrors: a bad request or a client with no provider or no
// cap hook makes no call.
func TestRequestErrors(t *testing.T) {
	fake := &Fake{}
	ok := &Client{Provider: fake, Cap: bigCap()}
	dup := request(t)
	dup.Exercises = append(dup.Exercises, pressInput(t))
	noEx := request(t)
	noEx.Exercises = nil
	many := request(t)
	many.Sessions = 8
	zero := request(t)
	zero.Sessions = 0
	two := request(t)
	two.Sessions = 2
	badIn := request(t)
	badIn.Exercises[0].Today = "not a date"
	for _, tc := range []struct {
		name   string
		c      *Client
		revise bool
		req    Request
	}{
		{"no provider", &Client{Cap: bigCap()}, false, request(t)},
		{"no cap hook", &Client{Provider: fake}, false, request(t)},
		{"duplicate exercise", ok, false, dup},
		{"no exercise", ok, false, noEx},
		{"8 sessions", ok, false, many},
		{"0 sessions", ok, false, zero},
		{"2 sessions of the reviser", ok, true, two},
		{"bad policy input", ok, false, badIn},
		{"unknown effort", &Client{Provider: fake, Cap: bigCap(), Effort: "extra-high"}, false, request(t)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			call := tc.c.Plan
			if tc.revise {
				call = tc.c.Revise
			}
			if _, err := call(context.Background(), tc.req); err == nil {
				t.Fatal("no error")
			}
		})
	}
	if n := len(fake.Calls()); n != 0 {
		t.Fatalf("%d calls: want none", n)
	}
}

// TestInputPrivacy: the call holds no note of the owner and no user id,
// and it sends the last sessions of each exercise alone.
func TestInputPrivacy(t *testing.T) {
	fake := &Fake{}
	req := request(t)
	more := req.Exercises[0].History[0]
	more.Date = "2026-09-20"
	req.Exercises[0].History = append([]policy.Outcome{more}, req.Exercises[0].History...)
	if _, err := (&Client{Provider: fake, Cap: bigCap()}).Plan(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	c := fake.Calls()[0]
	for _, s := range []string{"private note", "uid-test-1", "2026-09-20"} {
		if strings.Contains(string(c.Input), s) || strings.Contains(c.Instructions, s) {
			t.Errorf("the call holds %q", s)
		}
	}
	var in wireInput
	if err := json.Unmarshal(c.Input, &in); err != nil {
		t.Fatal(err)
	}
	if len(in.Exercises[0].History) != HistorySessions || in.Exercises[0].Target.Working[0].Load != 100 {
		t.Fatalf("input %+v", in.Exercises[0])
	}
}

// TestParentCancel: a call that the caller cancels is an error, not a
// time-out.
func TestParentCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	res, err := (&Client{Provider: &Fake{Hang: true}, Cap: bigCap()}).Plan(ctx, request(t))
	if err != nil || res.Status != StatusError {
		t.Fatalf("status %q err %v: want %q", res.Status, err, StatusError)
	}
}

// TestCapBoundary: the cap permits a call when the cap covers its
// reservation at the highest input rate, and refuses it one billionth
// of a dollar below.
func TestCapBoundary(t *testing.T) {
	req := request(t)
	input, err := userInput(RolePlanner, req)
	if err != nil {
		t.Fatal(err)
	}
	worst := Planner().worst(len(Instructions(Planner())) + len(input) + len(Schema()))
	for _, tc := range []struct {
		cap  NanoUSD
		want Status
	}{{worst, StatusOK}, {worst - 1, StatusCapped}} {
		fake := &Fake{}
		res, err := (&Client{Provider: fake, Cap: NewMemoryCap(Caps{User: tc.cap, Project: tc.cap})}).Plan(context.Background(), req)
		if err != nil || res.Status != tc.want || res.Cost.Reserved > tc.cap {
			t.Fatalf("cap %s: status %q reserved %s err %v: want %q", tc.cap, res.Status, res.Cost.Reserved, err, tc.want)
		}
	}
}

// TestRequestLimit: a request over the short context of the model makes
// no call, because a longer request bills at higher prices.
func TestRequestLimit(t *testing.T) {
	req := request(t)
	in := &req.Exercises[0]
	last := in.History[len(in.History)-1]
	for len(last.Log.Sets) < 5000 {
		last.Log.Sets = append(last.Log.Sets, last.Log.Sets[0])
	}
	in.History[len(in.History)-1] = last
	fake := &Fake{}
	if _, err := (&Client{Provider: fake, Cap: bigCap()}).Plan(context.Background(), req); err == nil || len(fake.Calls()) != 0 {
		t.Fatalf("err %v with %d calls: want an error and no call", err, len(fake.Calls()))
	}
}
