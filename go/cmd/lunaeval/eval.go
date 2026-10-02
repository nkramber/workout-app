package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// PlannerSessions is the number of sessions of each planner call.
const PlannerSessions = 3

//go:embed profiles.json
var profilesJSON []byte

// Profile is one synthetic profile of the planner run. The exercises
// have no history, so the policy gives the start of D-150.
type Profile struct {
	ID        string `json:"profile_id"`
	Returning bool   `json:"returning"`
	Exercises []struct {
		ID       domain.ExerciseID `json:"exercise_id"`
		Estimate float64           `json:"estimate_lb"`
	} `json:"exercises"`
	Cardio []domain.ExerciseID `json:"cardio"`
}

// Profiles gives the synthetic profiles of profiles.json.
func Profiles() ([]Profile, error) {
	var doc struct {
		Profiles []Profile `json:"profiles"`
	}
	if err := json.Unmarshal(profilesJSON, &doc); err != nil {
		return nil, fmt.Errorf("profiles.json: %w", err)
	}
	return doc.Profiles, nil
}

// Request gives the planner request of a profile.
func (p Profile) Request() ai.Request {
	r := ai.Request{User: "eval", Today: Today, Sessions: PlannerSessions, Cardio: p.Cardio}
	for _, x := range p.Exercises {
		e := catalogExercise(x.ID)
		r.Exercises = append(r.Exercises, policy.Input{
			Exercise: e, Entry: entry(e), Today: Today, Estimate: domain.Load(math.Round(x.Estimate * float64(domain.Pound))), Returning: p.Returning,
		})
	}
	return r
}

// Report is the result of one evaluation. It holds synthetic data
// alone, and no key.
type Report struct {
	Provider      string            `json:"provider"`
	Cap           string            `json:"cap"`
	Model         string            `json:"model"`
	Effort        string            `json:"effort"`
	PromptVersion string            `json:"prompt_version"`
	PromptHashes  map[string]string `json:"prompt_hashes"`
	SchemaName    string            `json:"schema"`
	PolicyVersion int               `json:"policy_version"`
	FilterVersion int               `json:"filter_version"`
	Repeats       int               `json:"scenario_repeats"`
	Calls         []Call            `json:"calls"`
	Totals        Totals            `json:"totals"`
	Scenarios     []ScenarioTotal   `json:"scenarios"`
}

// Call is the result of one call of a role.
type Call struct {
	Role     ai.RoleName   `json:"role"`
	Item     string        `json:"item"`
	Repeat   int           `json:"repeat"`
	Status   ai.Status     `json:"status"`
	Seconds  float64       `json:"seconds"`
	Usage    ai.Usage      `json:"usage"`
	Cost     ai.NanoUSD    `json:"cost_nano_usd"`
	Known    bool          `json:"cost_known"`
	Summary  string        `json:"summary,omitempty"`
	Filtered []ai.Filtered `json:"filtered,omitempty"`
	// Cardio is each cardio item of the plan. The policy has no cardio
	// rule, so no rule checks it.
	Cardio     []string   `json:"cardio,omitempty"`
	NotPlanned []string   `json:"not_planned,omitempty"`
	Decisions  []Decision `json:"decisions"`
}

// Decision is the policy decision of one exercise of a call.
type Decision struct {
	Case       string   `json:"case,omitempty"`
	Exercise   string   `json:"exercise"`
	Session    int      `json:"session"`
	Source     string   `json:"source"`
	Cause      string   `json:"cause,omitempty"`
	Violations []string `json:"violations,omitempty"`
	Proposal   string   `json:"proposal,omitempty"`
	Rules      string   `json:"rules_target"`
	Final      string   `json:"final_target"`
	Reason     string   `json:"reason"`
	// Safe and Why are the grade of a scenario case. Why is empty for a
	// safe target.
	Safe *bool  `json:"safe,omitempty"`
	Why  string `json:"why,omitempty"`
	// Jump is the result of the 50 percent jump of scenario F: true when
	// the policy refused it and gave the rules target.
	Jump *bool `json:"jump_refused,omitempty"`
}

// Totals are the counts of a report.
type Totals struct {
	Calls        int               `json:"calls"`
	ByStatus     map[ai.Status]int `json:"by_status"`
	SchemaPass   int               `json:"schema_pass"`
	Decisions    int               `json:"decisions"`
	Proposals    int               `json:"luna_proposals"`
	Accepted     int               `json:"accepted"`
	Refused      int               `json:"refused"`
	NoProposal   int               `json:"no_proposal"`
	NotPlanned   int               `json:"not_planned"`
	ByRule       map[string]int    `json:"refusals_by_rule"`
	Filtered     int               `json:"filtered_texts"`
	Cardio       int               `json:"cardio_items"`
	Cost         ai.NanoUSD        `json:"cost_nano_usd"`
	CostUnknown  int               `json:"calls_with_unknown_cost"`
	MaxSeconds   float64           `json:"max_seconds"`
	InputTokens  int64             `json:"input_tokens"`
	OutputTokens int64             `json:"output_tokens"`
	Reasoning    int64             `json:"reasoning_tokens"`
}

// ScenarioTotal is the grade of one scenario over each repeat.
type ScenarioTotal struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Cases    int    `json:"cases"`
	Safe     int    `json:"safe"`
	Accepted int    `json:"accepted"`
	Refused  int    `json:"refused"`
	Jumps    int    `json:"jumps"`
	JumpsOK  int    `json:"jumps_refused"`
}

// Pass tells that each case of the scenario gave the safe behavior.
func (s ScenarioTotal) Pass() bool { return s.Cases > 0 && s.Safe == s.Cases && s.JumpsOK == s.Jumps }

// job is one call of the evaluation.
type job struct {
	role     ai.RoleName
	item     string
	repeat   int
	req      ai.Request
	scenario *Scenario
}

// Run sends each profile through the planner one time, and each
// scenario through the reviser repeats times. Then the policy decides
// each exercise (D-23). Workers is the count of calls at the same time.
func Run(ctx context.Context, c *ai.Client, profiles []Profile, scenarios []Scenario, repeats, workers int) (Report, error) {
	var jobs []job
	for _, p := range profiles {
		jobs = append(jobs, job{role: ai.RolePlanner, item: p.ID, req: p.Request()})
	}
	for i := range scenarios {
		for r := range repeats {
			jobs = append(jobs, job{role: ai.RoleReviser, item: scenarios[i].ID, repeat: r + 1, req: scenarios[i].Request(), scenario: &scenarios[i]})
		}
	}
	calls := make([]Call, len(jobs))
	errs := make([]error, len(jobs))
	next := make(chan int)
	var wg sync.WaitGroup
	for range max(workers, 1) {
		wg.Go(func() {
			for i := range next {
				calls[i], errs[i] = runJob(ctx, c, jobs[i])
			}
		})
	}
	for i := range jobs {
		next <- i
	}
	close(next)
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			return Report{}, fmt.Errorf("%s %s repeat %d: %w", jobs[i].role, jobs[i].item, jobs[i].repeat, err)
		}
	}
	rep := Report{
		Model: ai.Planner().Model, Effort: ai.Planner().Effort, PromptVersion: ai.PromptVersion,
		PromptHashes: map[string]string{}, SchemaName: ai.SchemaName,
		PolicyVersion: policy.Version, FilterVersion: ai.FilterVersion, Repeats: repeats, Calls: calls,
	}
	for _, r := range ai.Roles() {
		rep.PromptHashes[string(r.Name)] = ai.PromptHash(r)
	}
	rep.Totals, rep.Scenarios = total(calls, scenarios)
	return rep, nil
}

func runJob(ctx context.Context, c *ai.Client, j job) (Call, error) {
	start := time.Now()
	var res ai.Result
	var err error
	if j.role == ai.RolePlanner {
		res, err = c.Plan(ctx, j.req)
	} else {
		res, err = c.Revise(ctx, j.req)
	}
	if err != nil {
		return Call{}, err
	}
	out := Call{
		Role: j.role, Item: j.item, Repeat: j.repeat, Status: res.Status,
		Seconds: time.Since(start).Seconds(), Usage: res.Cost.Usage, Cost: res.Cost.Cost, Known: res.Cost.Known,
		Decisions: []Decision{},
	}
	if res.Plan != nil {
		out.Summary, out.Filtered = res.Plan.Summary, res.Plan.Filtered
		for i, s := range res.Plan.Sessions {
			if s.Cardio != nil {
				out.Cardio = append(out.Cardio, fmt.Sprintf("session %d: %s, %d min", i, s.Cardio.Exercise, s.Cardio.Minutes))
			}
		}
	}
	if j.scenario != nil {
		for _, cs := range j.scenario.Cases {
			d, rec, rules, err := decide(res, 0, cs.Input)
			if err != nil {
				return Call{}, err
			}
			d.Case = cs.ID
			why := cs.Safe(rec.Target, rules)
			ok := why == ""
			d.Safe, d.Why = &ok, why
			if j.scenario.ID == "F" {
				jump, err := jumpRefused(cs.Input, rec, rules)
				if err != nil {
					return Call{}, err
				}
				d.Jump = &jump
			}
			out.Decisions = append(out.Decisions, d)
		}
		return out, nil
	}
	// A new exercise has no history, so its next session is its first
	// session in the plan. The policy checks the proposal of that
	// session.
	for _, in := range j.req.Exercises {
		session := firstSession(res, in.Exercise.ID)
		if session < 0 && res.Status == ai.StatusOK {
			out.NotPlanned = append(out.NotPlanned, string(in.Exercise.ID))
			continue
		}
		d, _, _, err := decide(res, max(session, 0), in)
		if err != nil {
			return Call{}, err
		}
		out.Decisions = append(out.Decisions, d)
	}
	return out, nil
}

func firstSession(res ai.Result, id domain.ExerciseID) int {
	if res.Plan == nil {
		return -1
	}
	for i := range res.Plan.Sessions {
		if res.Proposal(i, id).Target != nil {
			return i
		}
	}
	return -1
}

// decide gives the policy decision of one exercise, with the record
// and the rules target.
func decide(res ai.Result, session int, in policy.Input) (Decision, policy.Record, domain.PlannedExercise, error) {
	rec, err := policy.Decide(in, res.Proposal(session, in.Exercise.ID))
	if err != nil {
		return Decision{}, policy.Record{}, domain.PlannedExercise{}, err
	}
	rules, err := policy.Next(in)
	if err != nil {
		return Decision{}, policy.Record{}, domain.PlannedExercise{}, err
	}
	d := Decision{
		Exercise: string(in.Exercise.ID), Session: session, Source: string(rec.Source), Cause: string(rec.Cause),
		Rules: render(rules.Target), Final: render(rec.Target), Reason: res.Reason(session, rec),
	}
	if rec.Proposal != nil {
		d.Proposal = render(*rec.Proposal)
	}
	for _, v := range rec.Violations {
		d.Violations = append(d.Violations, string(v.Rule)+" "+v.Where)
	}
	return d, rec, rules.Target, nil
}

// jumpRefused proves scenario F on the output of the live model. It
// raises each load of the proposal of Luna by 50 percent, or of the
// rules target when Luna gave none. The policy must refuse the jump and
// give the rules target.
func jumpRefused(in policy.Input, rec policy.Record, rules domain.PlannedExercise) (bool, error) {
	base := rules
	if rec.Proposal != nil {
		base = *rec.Proposal
	}
	jump := base
	jump.Calibration = slices.Clone(base.Calibration)
	jump.Working = slices.Clone(base.Working)
	for i := range jump.Calibration {
		jump.Calibration[i].Load = up50(jump.Calibration[i].Load)
	}
	for i := range jump.Working {
		jump.Working[i].Load = up50(jump.Working[i].Load)
	}
	j, err := policy.Decide(in, policy.Proposal{Model: rec.Model, Effort: rec.Effort, PromptHash: rec.PromptHash, Target: &jump})
	if err != nil {
		return false, err
	}
	return j.Cause == policy.CauseRefused && j.Source == policy.SourceRules && reflect.DeepEqual(j.Target, rules), nil
}

// up50 gives a load 50 percent higher, down to a multiple of 5 lb.
func up50(l domain.Load) domain.Load {
	step := domain.Pounds(5)
	return (l * 3 / 2) / step * step
}

// render gives the text of a target, such as "3 x 12 at 25 lb, 2 RIR".
func render(p domain.PlannedExercise) string {
	var parts []string
	for _, s := range p.Calibration {
		parts = append(parts, fmt.Sprintf("calibration %d at %s", s.Reps, s.Load))
	}
	for _, s := range p.Working {
		parts = append(parts, fmt.Sprintf("%d at %s, %d RIR", s.Reps, s.Load, s.RIR))
	}
	return fmt.Sprintf("%s, rest %d s: %s", p.Exercise, p.RestSeconds, strings.Join(parts, "; "))
}

func total(calls []Call, scenarios []Scenario) (Totals, []ScenarioTotal) {
	t := Totals{ByStatus: map[ai.Status]int{}, ByRule: map[string]int{}}
	byID := map[string]*ScenarioTotal{}
	var sc []ScenarioTotal
	for _, s := range scenarios {
		sc = append(sc, ScenarioTotal{ID: s.ID, Title: s.Title})
	}
	for i := range sc {
		byID[sc[i].ID] = &sc[i]
	}
	for _, c := range calls {
		t.Calls++
		t.ByStatus[c.Status]++
		if c.Status == ai.StatusOK {
			t.SchemaPass++
		}
		t.Cost += c.Cost
		if !c.Known {
			t.CostUnknown++
		}
		t.MaxSeconds = max(t.MaxSeconds, c.Seconds)
		t.InputTokens += c.Usage.InputTokens
		t.OutputTokens += c.Usage.OutputTokens
		t.Reasoning += c.Usage.ReasoningTokens
		t.Filtered += len(c.Filtered)
		t.Cardio += len(c.Cardio)
		t.NotPlanned += len(c.NotPlanned)
		s := byID[c.Item]
		if c.Role != ai.RoleReviser {
			s = nil
		}
		for _, d := range c.Decisions {
			t.Decisions++
			switch {
			case d.Source == string(policy.SourceLuna):
				t.Proposals++
				t.Accepted++
			case d.Cause == string(policy.CauseRefused):
				t.Proposals++
				t.Refused++
				for _, v := range d.Violations {
					rule, _, _ := strings.Cut(v, " ")
					t.ByRule[rule]++
				}
			default:
				t.NoProposal++
			}
			if s == nil {
				continue
			}
			s.Cases++
			if d.Safe != nil && *d.Safe {
				s.Safe++
			}
			if d.Source == string(policy.SourceLuna) {
				s.Accepted++
			} else if d.Cause == string(policy.CauseRefused) {
				s.Refused++
			}
			if d.Jump != nil {
				s.Jumps++
				if *d.Jump {
					s.JumpsOK++
				}
			}
		}
	}
	return t, sc
}
