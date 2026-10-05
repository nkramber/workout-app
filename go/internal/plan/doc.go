package plan

import (
	"time"

	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// The stored documents. Each load is a whole number of tenths of a
// pound, as domain.Load. An empty list is stored as an empty array, and
// it reads back as nil.

type exclusionsDoc struct {
	Items    []exclusionDoc `firestore:"items"`
	Revision int64          `firestore:"revision"`
}

type exclusionDoc struct {
	Exercise string `firestore:"exercise_id"`
	Reason   string `firestore:"reason"`
}

type planDoc struct {
	CreatedAt         time.Time     `firestore:"created_at"`
	Today             string        `firestore:"today"`
	Summary           string        `firestore:"summary"`
	Sessions          []sessionDoc  `firestore:"sessions"`
	Guidance          []string      `firestore:"guidance_ids"`
	Filtered          []filteredDoc `firestore:"filtered"`
	Model             string        `firestore:"model"`
	Effort            string        `firestore:"effort"`
	PromptVersion     string        `firestore:"prompt_version"`
	PromptHash        string        `firestore:"prompt_hash"`
	SchemaName        string        `firestore:"schema"`
	PolicyVersion     int64         `firestore:"policy_version"`
	FilterVersion     int64         `firestore:"filter_version"`
	GuidanceVersion   int64         `firestore:"guidance_version"`
	CatalogVersion    int64         `firestore:"catalog_version"`
	BodyTablesVersion int64         `firestore:"body_tables_version"`
	Attempts          int64         `firestore:"attempts"`
	Revisions         int64         `firestore:"revisions"`
	LastRevision      *revisionDoc  `firestore:"last_revision"`
	RevisedWorkouts   []string      `firestore:"revised_workouts"`
	Claims            []claimDoc    `firestore:"revision_claims,omitempty"`
}

// claimDoc is a revision that runs now (D-304). A plan of an older
// version has none.
type claimDoc struct {
	WorkoutID string    `firestore:"workout_id"`
	Until     time.Time `firestore:"until"`
}

// revisionDoc is the last revision of a plan. A plan of an older
// version has none.
type revisionDoc struct {
	WorkoutID string    `firestore:"workout_id"`
	At        time.Time `firestore:"at"`
	Exercises []string  `firestore:"exercise_ids"`
}

type filteredDoc struct {
	Where string `firestore:"where"`
	Rule  string `firestore:"rule"`
}

type sessionDoc struct {
	Title     string        `firestore:"title"`
	WarmUp    string        `firestore:"warm_up_id"`
	CoolDown  string        `firestore:"cool_down_id"`
	Exercises []exerciseDoc `firestore:"exercises"`
	Cardio    *cardioDoc    `firestore:"cardio"`
}

type cardioDoc struct {
	Exercise string `firestore:"exercise_id"`
	Minutes  int64  `firestore:"minutes"`
}

type exerciseDoc struct {
	Target       targetDoc        `firestore:"target"`
	Reason       string           `firestore:"reason"`
	Record       recordDoc        `firestore:"record"`
	Calibration  []calibrationDoc `firestore:"calibration_loads"`
	ReasonSource string           `firestore:"reason_source"`
	ReasonCause  string           `firestore:"reason_cause"`
	Override     *overrideDoc     `firestore:"override"`
}

type overrideDoc struct {
	Target         targetDoc `firestore:"target"`
	Recommendation targetDoc `firestore:"recommendation"`
	Reason         string    `firestore:"reason"`
	At             time.Time `firestore:"at"`
	Today          string    `firestore:"today"`
}

// calibrationDoc is one row of the table of D-267. A plan of policy
// version 3 and an exercise with no calibration set have none, and they
// read back as nil.
type calibrationDoc struct {
	Weight int64 `firestore:"weight_tenth_lb"`
	Down   int64 `firestore:"down_tenth_lb"`
	Keep   int64 `firestore:"keep_tenth_lb"`
	UpOne  int64 `firestore:"up_one_tenth_lb"`
	UpTwo  int64 `firestore:"up_two_tenth_lb"`
}

func encodeCalibration(table []policy.CalibrationLoads) []calibrationDoc {
	if len(table) == 0 {
		return nil
	}
	out := make([]calibrationDoc, 0, len(table))
	for _, c := range table {
		out = append(out, calibrationDoc{int64(c.Weight), int64(c.Down), int64(c.Keep), int64(c.UpOne), int64(c.UpTwo)})
	}
	return out
}

func decodeCalibration(rows []calibrationDoc) []policy.CalibrationLoads {
	if len(rows) == 0 {
		return nil
	}
	out := make([]policy.CalibrationLoads, 0, len(rows))
	for _, d := range rows {
		out = append(out, policy.CalibrationLoads{Weight: domain.Load(d.Weight), Down: domain.Load(d.Down), Keep: domain.Load(d.Keep), UpOne: domain.Load(d.UpOne), UpTwo: domain.Load(d.UpTwo)})
	}
	return out
}

type targetDoc struct {
	Exercise    string   `firestore:"exercise_id"`
	Rest        int64    `firestore:"rest_seconds"`
	Calibration []setDoc `firestore:"calibration_sets"`
	Working     []setDoc `firestore:"working_sets"`
	// The first working set is the calibration (D-297). A target of
	// policy version 6 or earlier has no such field.
	FirstSet bool `firestore:"first_set_calibration,omitempty"`
	// The limit of the other working sets after the first set (D-306,
	// D-307). A target of policy version 7 or earlier has no such field.
	FollowMax int64 `firestore:"follow_max_tenth_lb,omitempty"`
}

// setDoc is a set. A calibration set stores an RIR of 0, which no
// reader uses (D-150).
type setDoc struct {
	Reps int64 `firestore:"reps"`
	Load int64 `firestore:"load_tenth_lb"`
	RIR  int64 `firestore:"rir_target"`
}

// recordDoc is the decision record of D-176.
type recordDoc struct {
	PolicyVersion int64          `firestore:"policy_version"`
	Exercise      string         `firestore:"exercise_id"`
	InputHash     string         `firestore:"input_hash"`
	Model         string         `firestore:"model"`
	Effort        string         `firestore:"effort"`
	PromptHash    string         `firestore:"prompt_hash"`
	Proposal      *targetDoc     `firestore:"proposal"`
	Violations    []violationDoc `firestore:"violations"`
	Source        string         `firestore:"source"`
	Cause         string         `firestore:"cause"`
	Rules         []string       `firestore:"rules"`
	Loads         []loadDoc      `firestore:"loads"`
	Target        targetDoc      `firestore:"target"`
	Reason        string         `firestore:"reason"`
}

type violationDoc struct {
	Rule   string `firestore:"rule"`
	Where  string `firestore:"where"`
	Detail string `firestore:"detail"`
}

type loadDoc struct {
	Where  string `firestore:"where"`
	Before int64  `firestore:"before_tenth_lb"`
	After  int64  `firestore:"after_tenth_lb"`
}

func encodePlan(p Plan) planDoc {
	d := planDoc{
		CreatedAt: p.CreatedAt.UTC(), Today: p.Today, Summary: p.Summary,
		Sessions: []sessionDoc{}, Guidance: strs(p.Guidance), Filtered: []filteredDoc{},
		Model: p.Model, Effort: p.Effort, PromptVersion: p.PromptVersion, PromptHash: p.PromptHash, SchemaName: p.SchemaName,
		PolicyVersion: int64(p.PolicyVersion), FilterVersion: int64(p.FilterVersion), GuidanceVersion: int64(p.GuidanceVersion),
		CatalogVersion: int64(p.CatalogVersion), BodyTablesVersion: int64(p.BodyTablesVersion), Attempts: int64(p.Attempts),
		Revisions: int64(p.Revisions), RevisedWorkouts: strs(p.RevisedWorkouts),
	}
	if r := p.LastRevision; r != nil {
		d.LastRevision = &revisionDoc{WorkoutID: r.WorkoutID, At: r.At.UTC(), Exercises: strs(r.Exercises)}
	}
	for _, c := range p.Claims {
		d.Claims = append(d.Claims, claimDoc{c.WorkoutID, c.Until.UTC()})
	}
	for _, f := range p.Filtered {
		d.Filtered = append(d.Filtered, filteredDoc{f.Where, string(f.Rule)})
	}
	for _, s := range p.Sessions {
		sd := sessionDoc{Title: s.Title, WarmUp: string(s.WarmUp), CoolDown: string(s.CoolDown), Exercises: []exerciseDoc{}}
		for _, e := range s.Exercises {
			ed := exerciseDoc{encodeTarget(e.Target), e.Reason, encodeRecord(e.Record), encodeCalibration(e.Calibration), string(e.ReasonSource), e.ReasonCause, nil}
			if o := e.Override; o != nil {
				ed.Override = &overrideDoc{encodeTarget(o.Target), encodeTarget(o.Recommendation), o.Reason, o.At.UTC(), o.Today}
			}
			sd.Exercises = append(sd.Exercises, ed)
		}
		if c := s.Cardio; c != nil {
			sd.Cardio = &cardioDoc{string(c.Exercise), int64(c.Minutes)}
		}
		d.Sessions = append(d.Sessions, sd)
	}
	return d
}

func (d planDoc) plan() Plan {
	p := Plan{
		CreatedAt: d.CreatedAt.UTC(), Today: d.Today, Summary: d.Summary, Guidance: ids[ai.GuidanceID](d.Guidance),
		Model: d.Model, Effort: d.Effort, PromptVersion: d.PromptVersion, PromptHash: d.PromptHash, SchemaName: d.SchemaName,
		PolicyVersion: int(d.PolicyVersion), FilterVersion: int(d.FilterVersion), GuidanceVersion: int(d.GuidanceVersion),
		CatalogVersion: int(d.CatalogVersion), BodyTablesVersion: int(d.BodyTablesVersion), Attempts: int(d.Attempts),
		Revisions: int(d.Revisions), RevisedWorkouts: ids[string](d.RevisedWorkouts),
	}
	if r := d.LastRevision; r != nil {
		p.LastRevision = &Revision{WorkoutID: r.WorkoutID, At: r.At.UTC(), Exercises: ids[domain.ExerciseID](r.Exercises)}
	}
	for _, c := range d.Claims {
		p.Claims = append(p.Claims, Claim{c.WorkoutID, c.Until.UTC()})
	}
	for _, f := range d.Filtered {
		p.Filtered = append(p.Filtered, ai.Filtered{Where: f.Where, Rule: ai.FilterRule(f.Rule)})
	}
	for _, sd := range d.Sessions {
		s := Session{Title: sd.Title, WarmUp: ai.GuidanceID(sd.WarmUp), CoolDown: ai.GuidanceID(sd.CoolDown)}
		for _, e := range sd.Exercises {
			x := Exercise{e.Target.target(), e.Reason, e.Record.record(), decodeCalibration(e.Calibration), policy.Source(e.ReasonSource), e.ReasonCause, nil}
			if o := e.Override; o != nil {
				x.Override = &Override{Target: o.Target.target(), Recommendation: o.Recommendation.target(), Reason: o.Reason, At: o.At.UTC(), Today: o.Today}
			}
			s.Exercises = append(s.Exercises, x)
		}
		if c := sd.Cardio; c != nil {
			s.Cardio = &domain.PlannedCardio{Exercise: idOf(c.Exercise), Minutes: int(c.Minutes)}
		}
		p.Sessions = append(p.Sessions, s)
	}
	return p
}

func encodeTarget(t domain.PlannedExercise) targetDoc {
	d := targetDoc{Exercise: string(t.Exercise), Rest: int64(t.RestSeconds), Calibration: []setDoc{}, Working: []setDoc{}, FirstSet: t.FirstSetCalibration, FollowMax: int64(t.FollowMax)}
	for _, s := range t.Calibration {
		d.Calibration = append(d.Calibration, setDoc{Reps: int64(s.Reps), Load: int64(s.Load)})
	}
	for _, s := range t.Working {
		d.Working = append(d.Working, setDoc{int64(s.Reps), int64(s.Load), int64(s.RIR)})
	}
	return d
}

func (d targetDoc) target() domain.PlannedExercise {
	t := domain.PlannedExercise{Exercise: idOf(d.Exercise), RestSeconds: int(d.Rest), FirstSetCalibration: d.FirstSet, FollowMax: domain.Load(d.FollowMax)}
	for _, s := range d.Calibration {
		t.Calibration = append(t.Calibration, domain.CalibrationSet{Reps: int(s.Reps), Load: domain.Load(s.Load)})
	}
	for _, s := range d.Working {
		t.Working = append(t.Working, domain.WorkingSet{Reps: int(s.Reps), Load: domain.Load(s.Load), RIR: int(s.RIR)})
	}
	return t
}

func encodeRecord(r policy.Record) recordDoc {
	d := recordDoc{
		PolicyVersion: int64(r.PolicyVersion), Exercise: string(r.Exercise), InputHash: r.InputHash,
		Model: r.Model, Effort: r.Effort, PromptHash: r.PromptHash, Violations: []violationDoc{},
		Source: string(r.Source), Cause: string(r.Cause), Rules: strs(r.Rules), Loads: []loadDoc{},
		Target: encodeTarget(r.Target), Reason: r.Reason,
	}
	if r.Proposal != nil {
		t := encodeTarget(*r.Proposal)
		d.Proposal = &t
	}
	for _, v := range r.Violations {
		d.Violations = append(d.Violations, violationDoc{string(v.Rule), v.Where, v.Detail})
	}
	for _, l := range r.Loads {
		d.Loads = append(d.Loads, loadDoc{l.Where, int64(l.Before), int64(l.After)})
	}
	return d
}

func (d recordDoc) record() policy.Record {
	r := policy.Record{
		PolicyVersion: int(d.PolicyVersion), Exercise: idOf(d.Exercise), InputHash: d.InputHash,
		Model: d.Model, Effort: d.Effort, PromptHash: d.PromptHash,
		Source: policy.Source(d.Source), Cause: policy.Cause(d.Cause), Rules: ids[policy.RuleID](d.Rules),
		Target: d.Target.target(), Reason: d.Reason,
	}
	if d.Proposal != nil {
		t := d.Proposal.target()
		r.Proposal = &t
	}
	for _, v := range d.Violations {
		r.Violations = append(r.Violations, policy.Violation{Rule: policy.RuleID(v.Rule), Where: v.Where, Detail: v.Detail})
	}
	for _, l := range d.Loads {
		r.Loads = append(r.Loads, policy.LoadChange{Where: l.Where, Before: domain.Load(l.Before), After: domain.Load(l.After)})
	}
	return r
}

// clone gives a deep copy through the stored form.
func (p Plan) clone() Plan { return encodePlan(p).plan() }

func idOf(s string) domain.ExerciseID { return domain.ExerciseID(s) }

// strs gives a new list of strings. An empty list gives an empty list,
// not nil, so Firestore stores an empty array.
func strs[T ~string](in []T) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		out = append(out, string(v))
	}
	return out
}

// ids gives the values as type T, or nil for an empty list.
func ids[T ~string](in []string) []T {
	var out []T
	for _, v := range in {
		out = append(out, T(v))
	}
	return out
}
