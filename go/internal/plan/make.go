package plan

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/policy"
	"github.com/nkramber/workout-app/go/internal/profile"
)

// MaxAttempts is the count of calls of one request at most: the first
// call and 3 retries (D-231).
const MaxAttempts = 4

// MaxErrorOutput is the largest output text of an error record, in
// bytes. It keeps the record inside the document limit of Firestore.
const MaxErrorOutput = 256_000

// recordTimeout limits the write of one error record. The write runs
// also when the request ended first, so the record is not lost.
const recordTimeout = 10 * time.Second

// Step names a step of a request (D-231).
type Step string

const (
	StepCall  Step = "call"
	StepCheck Step = "check"
	StepSave  Step = "save"
)

// Progress is one step of a request. Previous is the status of the
// failed call before a retry, or "" for the first call.
type Progress struct {
	Step        Step
	Attempt     int
	MaxAttempts int
	Previous    ai.Status
}

// Maker makes the plan of a user. AI holds the provider and the cap
// hook. Now gives the time, and Log gets a line of ids and numbers for
// each failed attempt. A nil Log writes no line.
type Maker struct {
	AI        *ai.Client
	Profiles  profile.Store
	Inventory inventory.Store
	Plans     Store
	Errors    ErrorLog
	Now       func() time.Time
	Log       *slog.Logger
}

// Make makes a new plan of the user for the date today, and saves it.
// When exclude is not nil, the exclusion and the plan save together
// (D-234). Each step calls progress, when it is not nil. Make saves
// nothing when ctx ends first (D-237).
//
// An error of a bad date or a bad exclusion matches domain.ErrInvalid.
// The other errors of the request are ErrNoProfile, ErrNothingToPlan,
// ErrCapped, ErrNoValidPlan, and ErrConflict.
func (m *Maker) Make(ctx context.Context, uid, today string, exclude *Exclusion, progress func(Progress)) (Plan, error) {
	if progress == nil {
		progress = func(Progress) {}
	}
	catalog, tables := domain.DefaultCatalog(), domain.DefaultBodyTables()
	if err := m.checkToday(today); err != nil {
		return Plan{}, err
	}
	prof, ok, err := m.Profiles.Get(ctx, uid)
	if err != nil {
		return Plan{}, fmt.Errorf("plan: the profile: %w", err)
	}
	if !ok {
		return Plan{}, ErrNoProfile
	}
	inv, err := m.Inventory.Get(ctx, uid)
	if err != nil {
		return Plan{}, fmt.Errorf("plan: the inventory: %w", err)
	}
	ex, err := m.Plans.Exclusions(ctx, uid)
	if err != nil {
		return Plan{}, fmt.Errorf("plan: the exclusions: %w", err)
	}
	kind := KindPlan
	if exclude != nil {
		kind = KindExclude
		if ex, err = ex.With(*exclude, catalog); err != nil {
			return Plan{}, err
		}
	}

	req, inputs := Request(uid, today, prof, inventory.ForPlan(inv), ex, catalog, tables)
	if len(req.Exercises) == 0 {
		return Plan{}, ErrNothingToPlan
	}

	var res ai.Result
	var previous ai.Status
	attempt := 1
	for ; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return Plan{}, err
		}
		progress(Progress{StepCall, attempt, MaxAttempts, previous})
		if res, err = m.AI.Plan(ctx, req); err != nil {
			return Plan{}, fmt.Errorf("plan: the call: %w", err)
		}
		if res.Status == ai.StatusOK {
			break
		}
		m.record(ctx, uid, kind, attempt, res)
		switch {
		case res.Status == ai.StatusCapped:
			return Plan{}, ErrCapped
		case ctx.Err() != nil:
			return Plan{}, ctx.Err()
		case attempt == MaxAttempts:
			return Plan{}, ErrNoValidPlan
		}
		req.Retry = &ai.Retry{Cause: res.Cause, Output: res.Output}
		previous = res.Status
	}

	progress(Progress{Step: StepCheck, MaxAttempts: MaxAttempts})
	p, err := build(res, inputs, today, m.now(), attempt)
	if err != nil {
		return Plan{}, err
	}
	if err := ctx.Err(); err != nil {
		return Plan{}, err
	}
	progress(Progress{Step: StepSave, MaxAttempts: MaxAttempts})
	if err := m.Plans.Save(ctx, uid, p, ex, exclude != nil); err != nil {
		if errors.Is(err, ErrConflict) {
			return Plan{}, ErrConflict
		}
		return Plan{}, fmt.Errorf("plan: the save: %w", err)
	}
	return p, nil
}

func (m *Maker) now() time.Time {
	if m.Now == nil {
		return time.Now()
	}
	return m.Now()
}

// checkToday refuses a date that is not in the form of
// domain.DateLayout, or that is more than one day from the date in UTC.
// The owner gives the local date, and the time zones of the world are
// inside one day of UTC.
func (m *Maker) checkToday(today string) error {
	d, err := time.Parse(domain.DateLayout, today)
	if err != nil || d.Format(domain.DateLayout) != today {
		return invalid("today: want a date as YYYY-MM-DD")
	}
	utc, _ := time.Parse(domain.DateLayout, m.now().UTC().Format(domain.DateLayout))
	if diff := d.Sub(utc); diff < -24*time.Hour || diff > 24*time.Hour {
		return invalid("today: want a date one day or less from the date in UTC")
	}
	return nil
}

// record adds the error record of a failed attempt (D-236). A failed
// write gives a log line and does not stop the request.
func (m *Maker) record(ctx context.Context, uid string, kind Kind, attempt int, res ai.Result) {
	now := m.now()
	r := ErrorRecord{
		User: uid, Time: now, ExpireAt: now.Add(ErrorRetention), Request: kind,
		Attempt: attempt, MaxAttempts: MaxAttempts, Status: res.Status, Cause: res.Cause,
		Model: res.Model, Effort: res.Effort, PromptVersion: ai.PromptVersion, PromptHash: res.PromptHash,
		SchemaName: ai.SchemaName, Cost: res.Cost, Output: ai.Clip(res.Output, MaxErrorOutput),
	}
	wctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), recordTimeout)
	defer cancel()
	err := m.Errors.Add(wctx, r)
	if m.Log != nil {
		// The line holds ids and numbers alone (D-80).
		m.Log.Warn("plan attempt failed", "uid", uid, "request", string(kind), "attempt", attempt,
			"status", string(res.Status), "cost_nano_usd", int64(res.Cost.Cost), "record_failed", err != nil)
	}
}

// Request gives the planner request of a user, and the policy input of
// each exercise of the request. The request holds the inputs of D-209
// and the training days of D-211. An exercise is in it when its machine
// is confirmed (D-49, D-193), it loads no injured area (D-208), and it
// is not excluded (D-48). A cardio exercise of the preference is in it
// on the same terms. Each new exercise starts as a return after a long
// break (D-238). The exclusion reasons stay on the server (D-229).
func Request(uid, today string, p profile.Profile, inv inventory.PlanInput, ex Exclusions, c domain.Catalog, t domain.BodyTables) (ai.Request, map[domain.ExerciseID]policy.Input) {
	pf := profile.ForPlan(p, c, t)
	req := ai.Request{
		User: uid, Today: today, Sessions: pf.Sessions,
		Profile: &ai.Profile{
			Experience: string(pf.Experience), GoalTemplate: string(pf.Template),
			MuscleGroups: strs(pf.Groups), FreeText: pf.FreeText,
		},
	}
	allowed := func(id domain.ExerciseID) (domain.Exercise, domain.InventoryEntry, bool) {
		e, ok := c.Exercise(id)
		if !ok || ex.Has(id) {
			return domain.Exercise{}, domain.InventoryEntry{}, false
		}
		entry, ok := inv.Inventory.Entry(e.Machine)
		return e, entry, ok
	}
	inputs := map[domain.ExerciseID]policy.Input{}
	for _, id := range pf.Exercises {
		e, entry, ok := allowed(id)
		if !ok {
			continue
		}
		in := policy.Input{Exercise: e, Entry: entry, Today: today, Estimate: inv.Estimates[id], Returning: true}
		req.Exercises = append(req.Exercises, in)
		inputs[id] = in
	}
	for _, id := range pf.Cardio {
		if _, _, ok := allowed(id); ok {
			req.Cardio = append(req.Cardio, id)
		}
	}
	return req, inputs
}

// build gives the plan of a valid output. The policy decides each
// exercise of each session (D-23). With no history between the sessions
// of the plan, each session of an exercise gets the decision of its next
// session. A refused proposal gets the target of the rules, and the
// record names the cause (D-176).
func build(res ai.Result, inputs map[domain.ExerciseID]policy.Input, today string, now time.Time, attempts int) (Plan, error) {
	out := Plan{
		CreatedAt: now.UTC(), Today: today, Summary: res.Plan.Summary,
		Guidance: append([]ai.GuidanceID(nil), res.Plan.Guidance...), Filtered: append([]ai.Filtered(nil), res.Plan.Filtered...),
		Model: res.Model, Effort: res.Effort, PromptVersion: ai.PromptVersion, PromptHash: res.PromptHash, SchemaName: ai.SchemaName,
		PolicyVersion: policy.Version, FilterVersion: ai.FilterVersion, GuidanceVersion: ai.GuidanceVersion,
		CatalogVersion: domain.CatalogVersion, BodyTablesVersion: domain.BodyTablesVersion, Attempts: attempts,
	}
	for i, s := range res.Plan.Sessions {
		sess := Session{Title: s.Title, WarmUp: s.WarmUp, CoolDown: s.CoolDown}
		if s.Cardio != nil {
			c := *s.Cardio
			sess.Cardio = &c
		}
		for _, e := range s.Exercises {
			in, ok := inputs[e.Target.Exercise]
			if !ok {
				// The parse of the layer refuses an exercise that is not
				// in the request, so this is a fault of the code.
				return Plan{}, fmt.Errorf("plan: exercise %q is not in the request", e.Target.Exercise)
			}
			rec, err := policy.Decide(in, res.Proposal(i, e.Target.Exercise))
			if err != nil {
				return Plan{}, fmt.Errorf("plan: the policy: %w", err)
			}
			sess.Exercises = append(sess.Exercises, Exercise{Target: rec.Target, Reason: res.Reason(i, rec), Record: rec})
		}
		out.Sessions = append(out.Sessions, sess)
	}
	return out, nil
}
