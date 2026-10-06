package main

import (
	"slices"
	"strconv"

	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
	"github.com/nkramber/workout-app/go/internal/revise"
)

// Report is the diff report of a replay. It holds counts and rule ids
// alone, and no uid, load, rep, note, or reason (D-80).
type Report struct {
	PolicyVersion int          `json:"policy_version"`
	Users         int          `json:"users"`
	UsersWithPlan int          `json:"users_with_plan"`
	Records       RecordReport `json:"records"`
	Copies        CopyReport   `json:"copies"`
}

// RecordReport compares each decision record of the active plans with
// its replay (D-176).
//
// ChangedByRule counts each changed target under each rule id that
// explains the change: a rule that only one of the two records names,
// and the rule of each changed field that no record names, such as
// "follow.first-set" for the limit of D-306. With no such rule, the
// target counts under the rules of the replay, or under "luna" for an
// accepted proposal of Luna.
type RecordReport struct {
	Total         int            `json:"total"`
	ByVersion     map[string]int `json:"by_version"`
	ByOrigin      map[string]int `json:"by_origin"`
	NotRebuilt    int            `json:"not_rebuilt"`
	SameInputHash int            `json:"same_input_hash"`
	Unchanged     int            `json:"unchanged"`
	Changed       int            `json:"changed"`
	SourceChanged int            `json:"source_changed"`
	ChangedByRule map[string]int `json:"changed_by_rule"`
	ChangedFields map[string]int `json:"changed_fields"`
}

// CopyReport compares each target copy of a finished workout with the
// target of the rules of the current version on its date (D-291).
// StaleHistory counts each copy that passes the bounds only without the
// newest workouts of its plan, because its phone started the workout
// before their revision. OutsideBounds counts each copy that breaks a
// bound of the current version with each such history, and
// ViolationsByRule counts each broken rule.
type CopyReport struct {
	Total              int            `json:"total"`
	Overrides          int            `json:"overrides"`
	NotRebuilt         int            `json:"not_rebuilt"`
	SameAsRules        int            `json:"same_as_rules"`
	DifferentFromRules int            `json:"different_from_rules"`
	StaleHistory       int            `json:"stale_history"`
	OutsideBounds      int            `json:"outside_bounds"`
	ViolationsByRule   map[string]int `json:"violations_by_rule"`
}

// RuleLuna is the key of ChangedByRule for an accepted proposal of
// Luna, which names no rule.
const RuleLuna = "luna"

// fieldRules gives the rule of a field that the policy sets with no
// rule id in the record.
var fieldRules = map[string]policy.RuleID{
	"follow_max":            policy.RuleFollowFirstSet,
	"first_set_calibration": policy.RuleCalibrationFirstSet,
}

func newReport() Report {
	return Report{
		PolicyVersion: policy.Version,
		Records:       RecordReport{ByVersion: map[string]int{}, ByOrigin: map[string]int{}, ChangedByRule: map[string]int{}, ChangedFields: map[string]int{}},
		Copies:        CopyReport{ViolationsByRule: map[string]int{}},
	}
}

// add adds the replay of one user to the report.
func (r *Report) add(rp revise.Replay) {
	r.Users++
	if rp.HasPlan {
		r.UsersWithPlan++
	}
	for _, x := range rp.Records {
		r.Records.add(x)
	}
	for _, c := range rp.Copies {
		r.Copies.add(c)
	}
}

func (r *RecordReport) add(x revise.Replayed) {
	r.Total++
	r.ByVersion[strconv.Itoa(x.Stored.PolicyVersion)]++
	r.ByOrigin[string(x.Origin)]++
	if !x.Rebuilt {
		r.NotRebuilt++
		return
	}
	if x.Stored.InputHash == x.Now.InputHash {
		r.SameInputHash++
	}
	if x.Stored.Source != x.Now.Source {
		r.SourceChanged++
	}
	fields := changedFields(x.Stored.Target, x.Now.Target)
	if len(fields) == 0 {
		r.Unchanged++
		return
	}
	r.Changed++
	for _, f := range fields {
		r.ChangedFields[f]++
	}
	for _, id := range changeRules(x.Stored, x.Now, fields) {
		r.ChangedByRule[id]++
	}
}

func (c *CopyReport) add(x revise.CopyReplayed) {
	c.Total++
	if x.Override {
		c.Overrides++
	}
	if !x.Rebuilt {
		c.NotRebuilt++
		return
	}
	if len(changedFields(x.Copy, x.Rules.Target)) == 0 {
		c.SameAsRules++
	} else {
		c.DifferentFromRules++
	}
	if x.Lag > 0 {
		c.StaleHistory++
	}
	if len(x.Violations) > 0 {
		c.OutsideBounds++
	}
	var seen []policy.RuleID
	for _, v := range x.Violations {
		if !slices.Contains(seen, v.Rule) {
			seen = append(seen, v.Rule)
			c.ViolationsByRule[string(v.Rule)]++
		}
	}
}

// changedFields names each field of b that differs from a, in a fixed
// order. A nil and an empty list are the same.
func changedFields(a, b domain.PlannedExercise) []string {
	var out []string
	add := func(name string, differ bool) {
		if differ {
			out = append(out, name)
		}
	}
	add("exercise", a.Exercise != b.Exercise)
	add("rest", a.RestSeconds != b.RestSeconds)
	add("calibration", !slices.Equal(a.Calibration, b.Calibration))
	add("sets", len(a.Working) != len(b.Working))
	n := min(len(a.Working), len(b.Working))
	load, reps, rir := false, false, false
	for i := range n {
		load = load || a.Working[i].Load != b.Working[i].Load
		reps = reps || a.Working[i].Reps != b.Working[i].Reps
		rir = rir || a.Working[i].RIR != b.Working[i].RIR
	}
	add("load", load)
	add("reps", reps)
	add("rir", rir)
	add("first_set_calibration", a.FirstSetCalibration != b.FirstSetCalibration)
	add("follow_max", a.FollowMax != b.FollowMax)
	return out
}

// changeRules gives the rule ids that explain the change of a target,
// sorted.
func changeRules(stored, now policy.Record, fields []string) []string {
	var ids []string
	add := func(id string) {
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	for _, id := range now.Rules {
		if !slices.Contains(stored.Rules, id) {
			add(string(id))
		}
	}
	for _, id := range stored.Rules {
		if !slices.Contains(now.Rules, id) {
			add(string(id))
		}
	}
	for _, f := range fields {
		if id, ok := fieldRules[f]; ok && !slices.Contains(now.Rules, id) && !slices.Contains(stored.Rules, id) {
			add(string(id))
		}
	}
	if len(ids) == 0 {
		for _, id := range now.Rules {
			add(string(id))
		}
	}
	if len(ids) == 0 {
		add(RuleLuna)
	}
	slices.Sort(ids)
	return ids
}
