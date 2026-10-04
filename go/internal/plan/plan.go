// Package plan makes and stores the one plan of the owner (work area
// 5.2). Luna proposes the plan through the role layer of
// "go/internal/ai", and the policy of "go/internal/policy" checks each
// target before the owner sees it (D-23).
//
// A plan reads the profile and the confirmed machines alone (D-49,
// D-193). The server removes each exercise of an injured area (D-208)
// and each excluded exercise (D-48, D-229) before the call. A call that
// gives no valid plan gets a retry with the cause of the failure, and
// the fourth failure gives ErrNoValidPlan with no change (D-230, D-231,
// D-235). Each failed attempt gets an error record (D-236).
//
// A plan holds the text of Luna and the targets of the owner, and an
// exclusion can hold a reason of the owner, so neither goes into a log
// (D-80). An error names ids and numbers alone.
package plan

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// Plan is the plan of one week, one session for each training day
// (D-211). The fields after Guidance name the versions that made it.
type Plan struct {
	CreatedAt time.Time
	// Today is the date of the first session, in the form of
	// domain.DateLayout.
	Today    string
	Summary  string
	Sessions []Session
	// Guidance holds the mobility and recovery items (D-152).
	Guidance []ai.GuidanceID
	// Filtered names each text of Luna that the filter replaced (D-183).
	Filtered []ai.Filtered

	Model             string
	Effort            string
	PromptVersion     string
	PromptHash        string
	SchemaName        string
	PolicyVersion     int
	FilterVersion     int
	GuidanceVersion   int
	CatalogVersion    int
	BodyTablesVersion int
	// Attempts is the count of calls that the plan needed.
	Attempts int

	// Revisions counts the revisions after a finished workout (D-292).
	// A revision keeps CreatedAt, so a workout keeps its link to the
	// plan (D-248). LastRevision is nil before the first revision.
	Revisions    int
	LastRevision *Revision
	// RevisedWorkouts holds the id of each finished workout that revised
	// the plan, the oldest first, MaxRevisedWorkouts at most. A replay of
	// the sync does not revise the plan two times.
	RevisedWorkouts []string
}

// MaxRevisedWorkouts is the largest count of workout ids that a plan
// keeps. A replay of the sync comes soon after its first call, so a
// short list is enough.
const MaxRevisedWorkouts = 50

// Revision is the last revision of a plan: the finished workout that
// started it, its time, and each exercise that got a new target, in
// the order of the plan (D-290).
type Revision struct {
	WorkoutID string
	At        time.Time
	Exercises []domain.ExerciseID
}

// Revised tells whether a finished workout revised the plan.
func (p Plan) Revised(workoutID string) bool { return slices.Contains(p.RevisedWorkouts, workoutID) }

// Session is one session of a plan (D-44). The title comes from a
// template, and the warm-up and the cool-down from the guidance catalog
// (D-152, D-182).
type Session struct {
	Title     string
	WarmUp    ai.GuidanceID
	CoolDown  ai.GuidanceID
	Exercises []Exercise
	Cardio    *domain.PlannedCardio
}

// Exercise is the final target of one exercise of a session, with the
// reason for the owner and the decision record of the policy (D-176).
// Calibration holds the loads of the working sets after a calibration
// set at each weight of the machine, and it is nil when the target has no
// calibration set (D-267).
//
// ReasonSource tells whether the reason is the reason of Luna or of the
// rules. ReasonCause tells why the reason of the rules shows after a
// revision, or it is "" (D-288). A plan of an older version has no
// source, and ReasonSourceOf gives it from the record.
//
// Override is the override of the owner for the next session of the
// exercise, or nil (D-69, D-293). Target stays the recommendation.
type Exercise struct {
	Target       domain.PlannedExercise
	Reason       string
	Record       policy.Record
	Calibration  []policy.CalibrationLoads
	ReasonSource policy.Source
	ReasonCause  string
	Override     *Override
}

// Override is an override of the owner for the next session of one
// exercise (D-69, D-293). Target is the target that the owner chose,
// Recommendation is the target of the plan that it replaced, and Reason
// is the reason of the owner. The three stay separate records. The
// policy checked Target before the save (D-23). A revision of the
// exercise removes the override, because an override is for one
// session. The reason is data of the owner, so it never goes into a log
// (D-80).
type Override struct {
	Target         domain.PlannedExercise
	Recommendation domain.PlannedExercise
	Reason         string
	At             time.Time
}

// SetOverride gives the override to each session of the plan that holds
// the exercise. A nil override removes it. It tells whether a session
// holds the exercise.
func (p *Plan) SetOverride(id domain.ExerciseID, o *Override) bool {
	found := false
	for i := range p.Sessions {
		for j := range p.Sessions[i].Exercises {
			e := &p.Sessions[i].Exercises[j]
			if e.Target.Exercise != id {
				continue
			}
			found = true
			e.Override = nil
			if o != nil {
				c := *o
				e.Override = &c
			}
		}
	}
	return found
}

// OverrideReason gives the reason of an override with no space at
// either end. It refuses an empty reason, a reason that is not UTF-8,
// and a reason over MaxReasonRunes characters. An error names no reason
// (D-80).
func OverrideReason(r string) (string, error) {
	if !utf8.ValidString(r) {
		return "", invalid("override reason: the text is not UTF-8")
	}
	r = strings.TrimSpace(r)
	switch k := utf8.RuneCountInString(r); {
	case k == 0:
		return "", invalid("override reason: want a reason")
	case k > MaxReasonRunes:
		return "", invalid("override reason: %d characters, want %d or fewer", k, MaxReasonRunes)
	}
	return r, nil
}

// ReasonSourceOf gives the source of the reason. A plan reason of Luna
// shows when the policy accepted the proposal of Luna, so a stored
// exercise with no source has the source of its record.
func (e Exercise) ReasonSourceOf() policy.Source {
	if e.ReasonSource != "" {
		return e.ReasonSource
	}
	return e.Record.Source
}

// MaxReasonRunes is the length limit of the reason of an exclusion and
// of an override, in characters (D-228, D-293).
const MaxReasonRunes = 200

// Exclusion is one excluded exercise with the optional reason of the
// owner (D-48). No model reads the reason (D-229).
type Exclusion struct {
	Exercise domain.ExerciseID
	Reason   string
}

// Exclusions is the list of the excluded exercises of the owner, in the
// order of the catalog. Revision counts each change of the list, so a
// save can find a change of another request (D-234).
type Exclusions struct {
	Items    []Exclusion
	Revision int64
}

// Has tells whether the list holds an exercise.
func (x Exclusions) Has(id domain.ExerciseID) bool {
	return slices.ContainsFunc(x.Items, func(e Exclusion) bool { return e.Exercise == id })
}

// With gives a new list with the exclusion added, in the order of the
// catalog. An exercise that is in the list already gets the new reason.
// The reason gets no space at either end. With refuses an exercise that
// is not in the catalog, and a reason that is not UTF-8 or is over its
// limit. An error names no reason (D-80).
func (x Exclusions) With(e Exclusion, c domain.Catalog) (Exclusions, error) {
	if _, ok := c.Exercise(e.Exercise); !ok {
		return Exclusions{}, invalid("exclusion: the exercise is not in the catalog")
	}
	if !utf8.ValidString(e.Reason) {
		return Exclusions{}, invalid("exclusion reason: the text is not UTF-8")
	}
	e.Reason = strings.TrimSpace(e.Reason)
	if k := utf8.RuneCountInString(e.Reason); k > MaxReasonRunes {
		return Exclusions{}, invalid("exclusion reason: %d characters, want %d or fewer", k, MaxReasonRunes)
	}
	out := Exclusions{Revision: x.Revision}
	for _, old := range x.Items {
		if old.Exercise != e.Exercise {
			out.Items = append(out.Items, old)
		}
	}
	out.Items = append(out.Items, e)
	rank := func(id domain.ExerciseID) int {
		return slices.IndexFunc(c.Exercises, func(ex domain.Exercise) bool { return ex.ID == id })
	}
	slices.SortStableFunc(out.Items, func(a, b Exclusion) int { return rank(a.Exercise) - rank(b.Exercise) })
	return out, nil
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", domain.ErrInvalid, fmt.Sprintf(format, args...))
}

// The errors of Make. Each one names no data of the owner.
var (
	// ErrNoProfile is a request of a user with no saved profile.
	ErrNoProfile = errors.New("plan: save a profile first")
	// ErrNothingToPlan is a request with no allowed resistance exercise:
	// no confirmed machine gives an exercise that loads no injured area
	// and is not excluded.
	ErrNothingToPlan = errors.New("plan: no confirmed machine gives an allowed exercise")
	// ErrCapped is a call that the monthly AI cap refused (D-25). No
	// retry follows it (D-230).
	ErrCapped = errors.New("plan: the monthly AI cap can not cover the call")
	// ErrNoValidPlan is a request whose last attempt gave no valid plan
	// (D-230).
	ErrNoValidPlan = errors.New("plan: Luna gave no valid plan")
	// ErrConflict is a save after another request changed the
	// exclusions. Nothing changed, and the owner can try again.
	ErrConflict = errors.New("plan: the exclusions changed during the request")
)
