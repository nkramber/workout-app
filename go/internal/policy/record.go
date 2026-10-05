package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// Proposal is the output of Luna for one exercise, with the model id,
// the effort, and the hash of the prompt template that the role layer
// used (D-24). Target is nil when Luna gave no valid proposal, for
// example after a time-out or a schema error. A zero Proposal tells
// that no Luna call occurred.
type Proposal struct {
	Model      string
	Effort     string
	PromptHash string
	Target     *domain.PlannedExercise
}

// Source names the origin of the target of a record.
type Source string

const (
	// SourceLuna is a Luna proposal that the policy accepted.
	SourceLuna Source = "luna"
	// SourceRules is the target of the rules alone (D-23).
	SourceRules Source = "rules"
)

// Cause tells why the rules fallback gave the target.
type Cause string

const (
	// CauseNone is a record with no fallback.
	CauseNone Cause = ""
	// CauseNoProposal is a fallback when Luna gave no proposal.
	CauseNoProposal Cause = "no-proposal"
	// CauseRefused is a fallback when the policy refused the proposal.
	CauseRefused Cause = "refused"
)

// Record is the decision record of one plan decision for one exercise
// (D-176): the policy version, the hash of the input, the model id, the
// effort, the prompt template hash, the proposal, each violation, the
// source and the cause, the rules that applied, each load before and
// after the rounding, and the final target with its reason.
//
// The record is workout data of the owner. The product stores it with
// the plan, and it never goes into a log, a metric, or an error report
// (D-80). A log line can hold the exercise id, the source, the cause,
// and the rule ids alone.
type Record struct {
	PolicyVersion int
	Exercise      domain.ExerciseID
	InputHash     string
	Model         string
	Effort        string
	PromptHash    string
	Proposal      *domain.PlannedExercise
	Violations    []Violation
	Source        Source
	Cause         Cause
	Rules         []RuleID
	Loads         []LoadChange
	Target        domain.PlannedExercise
	Reason        string
}

// Decide gives the decision record of one exercise for its next session
// (D-23, D-176). The policy checks the proposal of Luna. With no
// violation, the proposal is the target, and the record has no rule,
// because the rules did not set the target. A refused proposal, or no
// proposal, gives the target of Next, and the record names
// RuleFallback first, then the rules of Next. The owner never sees a
// refused proposal. The same input, proposal, and policy version give
// the same record.
func Decide(in Input, p Proposal) (Record, error) {
	d, err := Next(in)
	if err != nil {
		return Record{}, err
	}
	hash, err := InputHash(in)
	if err != nil {
		return Record{}, err
	}
	r := Record{
		PolicyVersion: Version,
		Exercise:      in.Exercise.ID,
		InputHash:     hash,
		Model:         p.Model,
		Effort:        p.Effort,
		PromptHash:    p.PromptHash,
		Cause:         CauseNoProposal,
	}
	if p.Target != nil {
		prop := clonePlan(*p.Target)
		r.Proposal = &prop
		if prop.Exercise != in.Exercise.ID {
			r.Violations = []Violation{{RuleProposalExercise, "exercise", fmt.Sprintf("exercise %q: want %q", prop.Exercise, in.Exercise.ID)}}
		} else if r.Violations, err = Check(prop, in); err != nil {
			return Record{}, err
		}
		if len(r.Violations) == 0 {
			r.Source, r.Cause = SourceLuna, CauseNone
			r.Target = clonePlan(prop)
			// The rules decide the calibration, and a proposal can not
			// change it (D-301).
			r.Target.FirstSetCalibration = d.Target.FirstSetCalibration
			// The policy gives the limit of the other working sets too
			// (D-306).
			r.Target.FollowMax = Follow(r.Target, in.Entry.Available())
			r.Loads = sameLoads(prop)
			return r, nil
		}
		r.Cause = CauseRefused
	}
	r.Source = SourceRules
	r.Rules = append([]RuleID{RuleFallback}, d.Rules...)
	r.Loads = d.Loads
	r.Target = d.Target
	r.Reason = d.Reason
	return r, nil
}

// Revise gives the decision record of one exercise after a logged
// session. The rules of Next give the target, and no proposal exists
// (D-288). So the source is SourceRules with no cause, and the rules
// are the rules of Next with no RuleFallback. The caller adds the model
// id, the effort, and the prompt hash of the reviser call that wrote
// the reason, when one occurred. The same input and policy version give
// the same record.
func Revise(in Input) (Record, error) {
	d, err := Next(in)
	if err != nil {
		return Record{}, err
	}
	hash, err := InputHash(in)
	if err != nil {
		return Record{}, err
	}
	return Record{
		PolicyVersion: Version,
		Exercise:      in.Exercise.ID,
		InputHash:     hash,
		Source:        SourceRules,
		Rules:         d.Rules,
		Loads:         d.Loads,
		Target:        d.Target,
		Reason:        d.Reason,
	}, nil
}

func clonePlan(p domain.PlannedExercise) domain.PlannedExercise {
	p.Calibration = slices.Clone(p.Calibration)
	p.Working = slices.Clone(p.Working)
	return p
}

// sameLoads gives the loads of a target that no rule computed.
func sameLoads(p domain.PlannedExercise) []LoadChange {
	var out []LoadChange
	for i, s := range p.Calibration {
		out = append(out, LoadChange{fmt.Sprintf("calibration[%d]", i), s.Load, s.Load})
	}
	for i, s := range p.Working {
		out = append(out, LoadChange{fmt.Sprintf("working[%d]", i), s.Load, s.Load})
	}
	return out
}

// InputHash gives the SHA-256 of the JSON form of an input, in hex. A
// record holds the hash in place of a copy of the logs, and a replay of
// the same input proves the decision (D-176). The hash does not read
// the note of a set, because no rule reads it.
func InputHash(in Input) (string, error) {
	c := in
	c.History = slices.Clone(in.History)
	for i := range c.History {
		sets := slices.Clone(c.History[i].Log.Sets)
		for j := range sets {
			sets[j].Note = ""
		}
		c.History[i].Log.Sets = sets
	}
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("policy: input hash: %w", err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
