package ai

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// Status names the outcome of one call.
type Status string

const (
	// StatusOK is a valid output. The policy still checks each target.
	StatusOK Status = "ok"
	// StatusMalformed is an output that breaks the schema or names an
	// exercise that is not in the request.
	StatusMalformed Status = "malformed"
	// StatusRefusal is a refusal of the model.
	StatusRefusal Status = "refusal"
	// StatusIncomplete is an output that stopped early.
	StatusIncomplete Status = "incomplete"
	// StatusTimeout is a call over the time limit of its role.
	StatusTimeout Status = "timeout"
	// StatusError is a call that failed for another cause.
	StatusError Status = "error"
	// StatusCapped is a call that the cap hook refused. The layer sent
	// nothing (D-25).
	StatusCapped Status = "capped"
)

// Request is the input of one call. Exercises holds the policy input
// of each exercise that the call can plan, and Cardio holds each cardio
// exercise of the inventory. User is the uid for the cap hook alone,
// and the layer never sends it.
type Request struct {
	User      string
	Today     string
	Sessions  int
	Exercises []policy.Input
	Cardio    []domain.ExerciseID
}

// Client calls the roles of Luna through a provider, with a cap hook.
// Record, when it is not nil, gets the cost record of each call (D-25).
// Timeout, when it is more than 0, replaces the time limit of the role.
type Client struct {
	Provider Provider
	Cap      CapHook
	Record   func(CostRecord)
	Timeout  time.Duration
}

// Result is the outcome of one call. Plan is nil unless Status is
// StatusOK.
type Result struct {
	Role       RoleName
	Model      string
	Effort     string
	PromptHash string
	Status     Status
	Plan       *Plan
	Cost       CostRecord
}

// Plan calls the planner for the next sessions of a request.
func (c *Client) Plan(ctx context.Context, req Request) (Result, error) {
	return c.call(ctx, Planner(), req)
}

// Revise calls the reviser for the next session after a logged
// session. The request has 1 session.
func (c *Client) Revise(ctx context.Context, req Request) (Result, error) {
	return c.call(ctx, Reviser(), req)
}

// call gives an error only for a bad request or a client with no
// provider or no cap hook. Each failure of the call gives a result with
// no plan, and the policy then gives the rules fallback (D-23).
func (c *Client) call(ctx context.Context, role Role, req Request) (Result, error) {
	if c.Provider == nil || c.Cap == nil {
		return Result{}, errors.New("ai: the client needs a provider and a cap hook")
	}
	if req.Sessions < 1 || req.Sessions > role.MaxSessions {
		return Result{}, fmt.Errorf("ai: %d sessions: want 1 to %d for the %s", req.Sessions, role.MaxSessions, role.Name)
	}
	if len(req.Exercises) == 0 {
		return Result{}, errors.New("ai: the request has no exercise")
	}
	seen := map[domain.ExerciseID]bool{}
	for _, x := range req.Exercises {
		if seen[x.Exercise.ID] {
			return Result{}, fmt.Errorf("ai: exercise %q is in the request two times", x.Exercise.ID)
		}
		seen[x.Exercise.ID] = true
	}
	input, err := userInput(req)
	if err != nil {
		return Result{}, err
	}
	call := Call{Role: role, Instructions: Instructions(role), Input: input, SchemaName: SchemaName, Schema: Schema()}
	r := Result{Role: role.Name, Model: role.Model, Effort: role.Effort, PromptHash: PromptHash(role)}
	rec := CostRecord{User: req.User, Role: role.Name, Model: role.Model, Effort: role.Effort, PromptHash: r.PromptHash}
	done := func(st Status) (Result, error) {
		r.Status, rec.Status = st, st
		r.Cost = rec
		if c.Record != nil {
			c.Record(rec)
		}
		return r, nil
	}

	size := len(call.Instructions) + len(call.Input) + len(call.Schema)
	if size > role.MaxRequestBytes {
		return Result{}, fmt.Errorf("ai: a request of %d bytes: want %d or fewer", size, role.MaxRequestBytes)
	}
	worst := role.worst(size)
	settle, err := c.Cap.Reserve(req.User, worst)
	if errors.Is(err, ErrCap) {
		rec.Known = true
		return done(StatusCapped)
	}
	if err != nil {
		return Result{}, err
	}
	rec.Reserved = worst

	limit := role.Timeout
	if c.Timeout > 0 {
		limit = c.Timeout
	}
	callCtx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	reply, err := c.Provider.Send(callCtx, call)
	if err != nil {
		// The request can reach OpenAI before the failure, so the
		// charge is not known, and the cap keeps the worst case.
		settle(worst)
		rec.Cost = worst
		if errors.Is(err, context.DeadlineExceeded) && ctx.Err() == nil {
			return done(StatusTimeout)
		}
		return done(StatusError)
	}
	rec.Usage, rec.Cost, rec.Known = reply.Usage, role.Prices.Cost(reply.Usage), true
	settle(rec.Cost)
	switch {
	case reply.Refusal:
		return done(StatusRefusal)
	case reply.Incomplete:
		return done(StatusIncomplete)
	}
	p, err := parse(reply.Text, req)
	if err != nil {
		return done(StatusMalformed)
	}
	r.Plan = &p
	return done(StatusOK)
}

// Proposal gives the proposal of Luna for one exercise of one session,
// for policy.Decide (D-23). The input of the policy is the input of the
// next session, so the policy checks a later session of the planner
// when that session is the next one. With no plan, or with no such exercise in
// the session, the proposal has no target, and the policy gives the
// rules fallback. A capped call sent nothing, so it gives a zero
// proposal.
func (r Result) Proposal(session int, id domain.ExerciseID) policy.Proposal {
	if r.Status == StatusCapped {
		return policy.Proposal{}
	}
	p := policy.Proposal{Model: r.Model, Effort: r.Effort, PromptHash: r.PromptHash}
	if e, ok := r.exercise(session, id); ok {
		t := e.Target
		t.Calibration = append([]domain.CalibrationSet(nil), t.Calibration...)
		t.Working = append([]domain.WorkingSet(nil), t.Working...)
		p.Target = &t
	}
	return p
}

// Reason gives the reason that the owner sees for the decision record
// of one exercise of one session. When the policy accepted the
// proposal of Luna, it is the reason of Luna after the filter (D-182,
// D-183). Otherwise it is the reason of the rules.
func (r Result) Reason(session int, rec policy.Record) string {
	if rec.Source != policy.SourceLuna {
		return rec.Reason
	}
	if e, ok := r.exercise(session, rec.Exercise); ok {
		return e.Reason
	}
	return ReasonTemplate
}

func (r Result) exercise(session int, id domain.ExerciseID) (Exercise, bool) {
	if r.Plan == nil || session < 0 || session >= len(r.Plan.Sessions) {
		return Exercise{}, false
	}
	for _, e := range r.Plan.Sessions[session].Exercises {
		if e.Target.Exercise == id {
			return e, true
		}
	}
	return Exercise{}, false
}
