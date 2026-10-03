package plan

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/policy"
	"github.com/nkramber/workout-app/go/internal/profile"
)

const (
	uid        = "uid-core"
	today      = "2026-10-02"
	injuryText = "secret injury text of the test"
	freeText   = "More back work, please."
)

var now = time.Date(2026, 10, 2, 15, 0, 0, 0, time.UTC)

// coreProfile is a synthetic profile of the core user of D-31: an
// intermediate user, the template "General fitness", 3 training days,
// and the treadmill and the upright bike as the cardio preference.
func coreProfile() profile.Profile {
	return profile.Profile{
		Experience: profile.Intermediate, Template: domain.TemplateGeneralFitness, Groups: domain.MuscleGroups(),
		FreeText: freeText, InjuryText: injuryText, AgeYears: 32, HeightIn: 70, WeightLb: 180,
		Cardio: []domain.ExerciseID{"treadmill", "upright_bike"}, TrainingDays: 3,
	}
}

func weights(lo, hi, step int64) []domain.Load {
	var out []domain.Load
	for w := lo; w <= hi; w += step {
		out = append(out, domain.Pounds(w))
	}
	return out
}

func machine(id domain.MachineID, st inventory.State, est map[domain.ExerciseID]domain.Load) inventory.Machine {
	if est == nil {
		est = map[domain.ExerciseID]domain.Load{}
	}
	return inventory.Machine{Entry: domain.InventoryEntry{Machine: id, Weights: weights(10, 200, 10)}, Estimates: est, State: st}
}

// coreInventory confirms the leg press, the leg extension, the chest
// press, the seated row, the cable station, and the treadmill. The
// shoulder press is a draft, and the upright bike is not in it.
func coreInventory() inventory.Inventory {
	c := inventory.Confirmed
	return inventory.Inventory{Machines: []inventory.Machine{
		machine("leg_press", c, nil),
		machine("leg_extension", c, nil),
		machine("chest_press", c, map[domain.ExerciseID]domain.Load{"chest_press": domain.Pounds(100)}),
		machine("shoulder_press", inventory.Draft, nil),
		machine("seated_row", c, nil),
		machine("cable_station", c, nil),
		{Entry: domain.InventoryEntry{Machine: "treadmill"}, Estimates: map[domain.ExerciseID]domain.Load{}, State: c},
	}}
}

type fixture struct {
	m      *Maker
	fake   *ai.Fake
	errs   *MemoryErrors
	plans  *Memory
	events []Progress
}

func newFixture(t *testing.T, p profile.Profile, inv inventory.Inventory, caps ai.Caps) *fixture {
	t.Helper()
	ctx := context.Background()
	profiles, inventories := profile.NewMemory(), inventory.NewMemory()
	if err := profiles.Save(ctx, uid, p); err != nil {
		t.Fatal(err)
	}
	if _, err := inventories.Update(ctx, uid, func(inventory.Inventory) (inventory.Inventory, error) { return inv, nil }); err != nil {
		t.Fatal(err)
	}
	f := &fixture{fake: &ai.Fake{}, errs: &MemoryErrors{}, plans: NewMemory()}
	f.m = &Maker{
		AI:       &ai.Client{Provider: f.fake, Cap: ai.NewMemoryCap(caps)},
		Profiles: profiles, Inventory: inventories, Plans: f.plans, Errors: f.errs,
		Now: func() time.Time { return now },
	}
	return f
}

func coreFixture(t *testing.T) *fixture {
	return newFixture(t, coreProfile(), coreInventory(), ai.Caps{User: ai.USD, Project: 2 * ai.USD})
}

func (f *fixture) make(t *testing.T, ex *Exclusion) (Plan, error) {
	t.Helper()
	f.events = nil
	return f.m.Make(context.Background(), uid, today, ex, func(p Progress) { f.events = append(f.events, p) })
}

func planned(p Plan) []domain.ExerciseID {
	var out []domain.ExerciseID
	for _, s := range p.Sessions {
		for _, e := range s.Exercises {
			if !slices.Contains(out, e.Target.Exercise) {
				out = append(out, e.Target.Exercise)
			}
		}
	}
	return out
}

func input(t *testing.T, c ai.Call) map[string]any {
	t.Helper()
	var in map[string]any
	if err := json.Unmarshal(c.Input, &in); err != nil {
		t.Fatal(err)
	}
	return in
}

func inputIDs(t *testing.T, c ai.Call) (exercises, cardio []string) {
	t.Helper()
	in := input(t, c)
	for _, e := range in["exercises"].([]any) {
		exercises = append(exercises, e.(map[string]any)["exercise_id"].(string))
	}
	for _, id := range in["cardio_exercises"].([]any) {
		cardio = append(cardio, id.(string))
	}
	return exercises, cardio
}

// TestCorePlan is the acceptance story of the maker: a valid plan for
// the core profile of D-31, from one call of the fake provider.
func TestCorePlan(t *testing.T) {
	f := coreFixture(t)
	p, err := f.make(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Sessions) != 3 || p.Attempts != 1 || p.PromptVersion != ai.PromptVersion || p.PolicyVersion != policy.Version || p.Today != today || !p.CreatedAt.Equal(now) {
		t.Fatalf("plan: %d sessions, %d attempts, prompt %s, policy %d, today %s, time %v", len(p.Sessions), p.Attempts, p.PromptVersion, p.PolicyVersion, p.Today, p.CreatedAt)
	}
	want := []Progress{{StepCall, 1, MaxAttempts, ""}, {Step: StepCheck, MaxAttempts: MaxAttempts}, {Step: StepSave, MaxAttempts: MaxAttempts}}
	if !reflect.DeepEqual(f.events, want) {
		t.Fatalf("progress %v, want %v", f.events, want)
	}

	// The confirmed machines alone: no shoulder press (a draft) and no
	// upright bike (not in the inventory).
	exercises, cardio := inputIDs(t, f.fake.Calls()[0])
	wantIDs := []string{"leg_press", "leg_extension", "chest_press", "seated_row", "lat_pulldown", "triceps_pulldown"}
	if !slices.Equal(exercises, wantIDs) || !slices.Equal(cardio, []string{"treadmill"}) {
		t.Fatalf("the call plans %v and cardio %v", exercises, cardio)
	}
	for _, id := range planned(p) {
		if !slices.Contains(wantIDs, string(id)) {
			t.Fatalf("the plan holds %q", id)
		}
	}

	// The input holds the inputs of D-209 alone.
	in := input(t, f.fake.Calls()[0])
	prof := in["profile"].(map[string]any)
	if prof["free_text"] != freeText || prof["goal_template"] != "general_fitness" || prof["experience"] != "intermediate" {
		t.Fatalf("profile %v", prof)
	}
	raw := string(f.fake.Calls()[0].Input)
	for _, s := range []string{injuryText, "age", "height", "weight_lb\"", "injur", uid} {
		if strings.Contains(raw, s) {
			t.Errorf("the input holds %q", s)
		}
	}

	// The policy accepted each proposal of the echo. Each new exercise
	// starts at 70 percent of its estimate (D-238).
	for _, s := range p.Sessions {
		if s.WarmUp != ai.DefaultWarmUp || s.CoolDown != ai.DefaultCoolDown || s.Title == "" {
			t.Fatalf("session %+v", s)
		}
		for _, e := range s.Exercises {
			if e.Record.Source != policy.SourceLuna || e.Record.Exercise != e.Target.Exercise || e.Reason == "" {
				t.Fatalf("record %+v", e.Record)
			}
			if e.Target.Exercise == "chest_press" && e.Target.Working[0].Load != domain.Pounds(70) {
				t.Fatalf("chest press %+v, want 70 lb from an estimate of 100 lb", e.Target.Working)
			}
		}
	}

	stored, ok, err := f.plans.Get(context.Background(), uid)
	if err != nil || !ok || !reflect.DeepEqual(stored, p) {
		t.Fatalf("stored plan differs: ok %v, err %v", ok, err)
	}
	if n := len(f.errs.Records()); n != 0 {
		t.Fatalf("%d error records, want 0", n)
	}
}

// TestInjuryAndExclusion: the server removes each exercise of an injured
// area (D-208) and each excluded exercise (D-48) before the call, and the
// reason of an exclusion never goes to Luna (D-229).
func TestInjuryAndExclusion(t *testing.T) {
	prof := coreProfile()
	prof.InjuredAreas = []domain.Area{domain.AreaKnee}
	f := newFixture(t, prof, coreInventory(), ai.Caps{User: ai.USD, Project: 2 * ai.USD})
	const reason = "The seat of this machine is broken."
	p, err := f.make(t, &Exclusion{Exercise: "seated_row", Reason: "  " + reason + " "})
	if err != nil {
		t.Fatal(err)
	}
	exercises, cardio := inputIDs(t, f.fake.Calls()[0])
	if !slices.Equal(exercises, []string{"chest_press", "lat_pulldown", "triceps_pulldown"}) || len(cardio) != 0 {
		t.Fatalf("the call plans %v and cardio %v", exercises, cardio)
	}
	for _, id := range planned(p) {
		if id == "seated_row" || id == "leg_press" || id == "leg_extension" {
			t.Fatalf("the plan holds %q", id)
		}
	}
	ex, err := f.plans.Exclusions(context.Background(), uid)
	if err != nil || ex.Revision != 1 || !reflect.DeepEqual(ex.Items, []Exclusion{{"seated_row", reason}}) {
		t.Fatalf("exclusions %+v, err %v", ex, err)
	}

	// A later request keeps the exclusion, and no call holds the reason.
	if _, err := f.make(t, nil); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.fake.Calls() {
		if strings.Contains(string(c.Input), "broken") || strings.Contains(c.Instructions, "broken") {
			t.Fatal("a call holds the reason of an exclusion")
		}
		if ids, _ := inputIDs(t, c); slices.Contains(ids, "seated_row") {
			t.Fatal("a call holds the excluded exercise")
		}
	}
}

// TestRetryWithCause: an invalid output gets a retry that sends the
// cause and the failed output (D-235), and an error record (D-236).
func TestRetryWithCause(t *testing.T) {
	f := coreFixture(t)
	var bad string
	f.fake.Reply = func(c ai.Call) (ai.Reply, error) {
		r, err := ai.EchoReply(c)
		if len(f.fake.Calls()) == 1 {
			var out map[string]any
			_ = json.Unmarshal([]byte(r.Text), &out)
			out["sessions"].([]any)[1].(map[string]any)["cardio"] = map[string]any{"exercise_id": "treadmill", "minutes": 45}
			b, _ := json.Marshal(out)
			r.Text, bad = string(b), string(b)
		}
		return r, err
	}
	p, err := f.make(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Attempts != 2 || len(f.fake.Calls()) != 2 {
		t.Fatalf("%d attempts and %d calls, want 2", p.Attempts, len(f.fake.Calls()))
	}
	if f.events[1] != (Progress{StepCall, 2, MaxAttempts, ai.StatusMalformed}) {
		t.Fatalf("progress %v", f.events)
	}
	prev := input(t, f.fake.Calls()[1])["previous_attempt"].(map[string]any)
	if !strings.Contains(prev["cause"].(string), "sessions[1].cardio: 45 minutes: want 20 to 30") || prev["output"] != bad {
		t.Fatalf("previous_attempt %v", prev)
	}
	if input(t, f.fake.Calls()[0])["previous_attempt"] != nil {
		t.Fatal("the first call holds a previous attempt")
	}
	recs := f.errs.Records()
	if len(recs) != 1 {
		t.Fatalf("%d error records, want 1", len(recs))
	}
	r := recs[0]
	if r.User != uid || r.Request != KindPlan || r.Attempt != 1 || r.MaxAttempts != MaxAttempts || r.Status != ai.StatusMalformed ||
		r.Output != bad || r.Cause != prev["cause"] || r.PromptVersion != ai.PromptVersion || r.PromptHash == "" || r.Model == "" ||
		!r.Time.Equal(now) || !r.ExpireAt.Equal(now.Add(90*24*time.Hour)) || !r.Cost.Known || r.Cost.Cost <= 0 {
		t.Fatalf("error record %+v", r)
	}
}

// TestCardioRuleRetry: when the profile likes a cardio exercise, an
// output with a session of fewer than 20 minutes of cardio, or with no
// cardio, is an invalid output that uses one retry (D-230, D-255). The
// plan of the retry gives each session 20 to 30 minutes of cardio.
func TestCardioRuleRetry(t *testing.T) {
	for _, tc := range []struct {
		name   string
		cardio map[string]any
		cause  string
	}{
		{"19 minutes", map[string]any{"exercise_id": "treadmill", "minutes": 19}, "sessions[0].cardio: 19 minutes: want 20 to 30"},
		{"no cardio", map[string]any{"exercise_id": "", "minutes": 0}, "sessions[0].cardio: no cardio"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := coreFixture(t)
			f.fake.Reply = func(c ai.Call) (ai.Reply, error) {
				r, err := ai.EchoReply(c)
				if len(f.fake.Calls()) == 1 {
					var out map[string]any
					_ = json.Unmarshal([]byte(r.Text), &out)
					out["sessions"].([]any)[0].(map[string]any)["cardio"] = tc.cardio
					b, _ := json.Marshal(out)
					r.Text = string(b)
				}
				return r, err
			}
			p, err := f.make(t, nil)
			if err != nil {
				t.Fatal(err)
			}
			if p.Attempts != 2 || len(f.fake.Calls()) != 2 {
				t.Fatalf("%d attempts and %d calls, want 2", p.Attempts, len(f.fake.Calls()))
			}
			prev := input(t, f.fake.Calls()[1])["previous_attempt"].(map[string]any)
			if !strings.Contains(prev["cause"].(string), tc.cause) {
				t.Fatalf("cause %q, want %q", prev["cause"], tc.cause)
			}
			for i, s := range p.Sessions {
				if s.Cardio == nil || s.Cardio.Minutes < 20 || s.Cardio.Minutes > 30 {
					t.Fatalf("session %d cardio %+v, want 20 to 30 minutes", i, s.Cardio)
				}
			}
		})
	}
}

// TestFourFailures: after 4 failed calls the request gives
// ErrNoValidPlan, and nothing changes: not the old plan, and not the
// exclusions (D-230, D-234).
func TestFourFailures(t *testing.T) {
	f := coreFixture(t)
	old, err := f.make(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	f.fake.Reply = func(ai.Call) (ai.Reply, error) { return ai.Reply{Refusal: true, Text: "refused"}, nil }
	_, err = f.make(t, &Exclusion{Exercise: "chest_press"})
	if !errors.Is(err, ErrNoValidPlan) {
		t.Fatalf("err %v, want ErrNoValidPlan", err)
	}
	if n := len(f.fake.Calls()); n != 1+MaxAttempts {
		t.Fatalf("%d calls, want %d", n, 1+MaxAttempts)
	}
	recs := f.errs.Records()
	if len(recs) != MaxAttempts {
		t.Fatalf("%d error records, want %d", len(recs), MaxAttempts)
	}
	for i, r := range recs {
		if r.Attempt != i+1 || r.Request != KindExclude || r.Status != ai.StatusRefusal || r.Output != "refused" {
			t.Fatalf("record %d: %+v", i, r)
		}
	}
	if n := len(f.events); n != MaxAttempts || f.events[3].Previous != ai.StatusRefusal {
		t.Fatalf("progress %v", f.events)
	}
	got, _, _ := f.plans.Get(context.Background(), uid)
	ex, _ := f.plans.Exclusions(context.Background(), uid)
	if !reflect.DeepEqual(got, old) || len(ex.Items) != 0 || ex.Revision != 0 {
		t.Fatalf("the request changed the store: exclusions %+v", ex)
	}
}

// TestCapped: a call over the cap ends the request at once, with no
// call to the provider and no retry (D-25, D-230).
func TestCapped(t *testing.T) {
	f := newFixture(t, coreProfile(), coreInventory(), ai.Caps{})
	_, err := f.make(t, nil)
	if !errors.Is(err, ErrCapped) || len(f.fake.Calls()) != 0 {
		t.Fatalf("err %v and %d calls, want ErrCapped and 0", err, len(f.fake.Calls()))
	}
	if recs := f.errs.Records(); len(recs) != 1 || recs[0].Status != ai.StatusCapped {
		t.Fatalf("error records %+v", recs)
	}
	if _, ok, _ := f.plans.Get(context.Background(), uid); ok {
		t.Fatal("a capped request saved a plan")
	}
}

// TestRefusedExercise: the policy refuses an unsafe proposal of one
// exercise, and that exercise gets the target of the rules (D-23,
// D-176). The other exercises keep the proposal of Luna.
func TestRefusedExercise(t *testing.T) {
	f := coreFixture(t)
	f.fake.Reply = func(c ai.Call) (ai.Reply, error) {
		r, err := ai.EchoReply(c)
		var out map[string]any
		_ = json.Unmarshal([]byte(r.Text), &out)
		for _, s := range out["sessions"].([]any) {
			for _, e := range s.(map[string]any)["exercises"].([]any) {
				if e := e.(map[string]any); e["exercise_id"] == "chest_press" {
					for _, w := range e["working_sets"].([]any) {
						w.(map[string]any)["load_lb"] = 150
					}
				}
			}
		}
		b, _ := json.Marshal(out)
		r.Text = string(b)
		return r, err
	}
	p, err := f.make(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range p.Sessions {
		for _, e := range s.Exercises {
			rec := e.Record
			if e.Target.Exercise != "chest_press" {
				if rec.Source != policy.SourceLuna {
					t.Fatalf("%s: source %s", e.Target.Exercise, rec.Source)
				}
				continue
			}
			if rec.Source != policy.SourceRules || rec.Cause != policy.CauseRefused || len(rec.Violations) == 0 || rec.Proposal == nil {
				t.Fatalf("chest press record %+v", rec)
			}
			if e.Target.Working[0].Load != domain.Pounds(70) || e.Reason != rec.Reason {
				t.Fatalf("chest press target %+v, reason %q", e.Target, e.Reason)
			}
		}
	}
}

// TestCancel: a request that ends before the save saves nothing (D-237).
func TestCancel(t *testing.T) {
	f := coreFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	_, err := f.m.Make(ctx, uid, today, nil, func(p Progress) {
		if p.Step == StepCheck {
			cancel()
		}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err %v, want context.Canceled", err)
	}
	if _, ok, _ := f.plans.Get(context.Background(), uid); ok {
		t.Fatal("a cancelled request saved a plan")
	}
}

// TestConflict: an exclusion of another request during the request
// stops the save, so no plan misses an exclusion (D-234).
func TestConflict(t *testing.T) {
	f := coreFixture(t)
	ctx := context.Background()
	_, err := f.m.Make(ctx, uid, today, nil, func(p Progress) {
		if p.Step == StepCheck {
			other := Exclusions{Items: []Exclusion{{Exercise: "chest_press"}}}
			if err := f.plans.Save(ctx, uid, Plan{Today: "other"}, other, true); err != nil {
				t.Fatal(err)
			}
		}
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err %v, want ErrConflict", err)
	}
	if got, _, _ := f.plans.Get(ctx, uid); got.Today != "other" {
		t.Fatal("the request saved over the plan of the other request")
	}
}

// TestRequestErrors: a bad date, a bad exclusion, no profile, and no
// allowed exercise end the request before a call.
func TestRequestErrors(t *testing.T) {
	ctx := context.Background()
	f := coreFixture(t)
	for _, tc := range []struct {
		name  string
		today string
		ex    *Exclusion
	}{
		{"bad date", "2026-13-01", nil},
		{"date with no zeros", "2026-10-2", nil},
		{"date 2 days later", "2026-10-04", nil},
		{"date 2 days earlier", "2026-09-30", nil},
		{"unknown exercise", today, &Exclusion{Exercise: "squat_rack"}},
		{"long reason", today, &Exclusion{Exercise: "chest_press", Reason: strings.Repeat("é", MaxReasonRunes+1)}},
		{"reason not UTF-8", today, &Exclusion{Exercise: "chest_press", Reason: "\xff"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.m.Make(ctx, uid, tc.today, tc.ex, nil)
			if !errors.Is(err, domain.ErrInvalid) {
				t.Fatalf("err %v, want ErrInvalid", err)
			}
		})
	}
	for _, d := range []string{"2026-10-01", "2026-10-03"} {
		if _, err := f.m.Make(ctx, uid, d, nil, nil); err != nil {
			t.Fatalf("date %s: %v", d, err)
		}
	}
	if _, err := f.m.Make(ctx, uid, today, &Exclusion{Exercise: "chest_press", Reason: strings.Repeat("é", MaxReasonRunes)}, nil); err != nil {
		t.Fatalf("a reason at the limit: %v", err)
	}
	if _, err := f.m.Make(ctx, "uid-none", today, nil, nil); !errors.Is(err, ErrNoProfile) {
		t.Fatalf("err %v, want ErrNoProfile", err)
	}
	drafts := newFixture(t, coreProfile(), inventory.Inventory{Machines: []inventory.Machine{machine("chest_press", inventory.Draft, nil)}}, ai.Caps{User: ai.USD, Project: ai.USD})
	if _, err := drafts.make(t, nil); !errors.Is(err, ErrNothingToPlan) {
		t.Fatalf("err %v, want ErrNothingToPlan", err)
	}
	if n := len(f.fake.Calls()) + len(drafts.fake.Calls()); n != 3 {
		t.Fatalf("%d calls, want 3 from the valid requests", n)
	}
}

// TestExclusionsWith: a new exclusion goes into the order of the
// catalog, and a second exclusion of an exercise replaces its reason.
func TestExclusionsWith(t *testing.T) {
	c := domain.DefaultCatalog()
	x := Exclusions{Revision: 4}
	var err error
	for _, e := range []Exclusion{{"seated_row", "a"}, {"leg_press", ""}, {"seated_row", " b "}} {
		if x, err = x.With(e, c); err != nil {
			t.Fatal(err)
		}
	}
	want := Exclusions{Items: []Exclusion{{"leg_press", ""}, {"seated_row", "b"}}, Revision: 4}
	if !reflect.DeepEqual(x, want) || !x.Has("leg_press") || x.Has("chest_press") {
		t.Fatalf("exclusions %+v", x)
	}
}
