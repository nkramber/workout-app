package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// ErrMalformed is the error that each parse error matches with
// errors.Is. A parse error names a field and a position, and never a
// text or a value of the output, so it can go into a log (D-80).
var ErrMalformed = errors.New("ai: malformed output")

type parseError string

func (e parseError) Error() string { return "ai: malformed output: " + string(e) }
func (parseError) Is(t error) bool { return t == ErrMalformed }

func malformed(format string, args ...any) error {
	return parseError(fmt.Sprintf(format, args...))
}

// The bounds of a session of Luna. A session has MaxSessionExercises
// exercises with sets or fewer (D-233). When the request has a cardio
// exercise, each session has MinCardioMinutes to MaxCardioMinutes
// minutes of one (D-255). When it has none, no session has cardio. An
// output outside a bound is malformed, and the plan API retries the
// call (D-230).
const (
	MaxSessionExercises = 8
	MinCardioMinutes    = 20
	MaxCardioMinutes    = 30
)

// Plan is a valid output of Luna. The policy has not checked a target
// yet (D-23). The plan holds text of Luna and targets of the owner, so
// it never goes into a log (D-80).
type Plan struct {
	Summary  string
	Sessions []Session
	Guidance []GuidanceID
	// Filtered names each text of Luna that the filter replaced with a
	// template text (D-183).
	Filtered []Filtered
}

// Session is one session of a plan. The title comes from a template,
// and the warm-up and the cool-down are items of the guidance catalog
// (D-152, D-182).
type Session struct {
	Title     string
	WarmUp    GuidanceID
	CoolDown  GuidanceID
	Exercises []Exercise
	Cardio    *domain.PlannedCardio
}

// Exercise is the proposal of Luna for one exercise, with the reason
// of Luna after the filter.
type Exercise struct {
	Target domain.PlannedExercise
	Reason string
}

// Filtered names one text that the filter replaced: where it was, such
// as "sessions[0].exercises[1].reason", and the rule that blocked it.
type Filtered struct {
	Where string
	Rule  FilterRule
}

// The wire form of the output. A pointer tells a missing field from a
// zero value, because the schema requires each field.
type outPlan struct {
	Summary  *string       `json:"summary"`
	Sessions *[]outSession `json:"sessions"`
	Guidance *[]GuidanceID `json:"guidance_ids"`
}

type outSession struct {
	WarmUp    *GuidanceID    `json:"warm_up_id"`
	Exercises *[]outExercise `json:"exercises"`
	Cardio    *outCardio     `json:"cardio"`
	CoolDown  *GuidanceID    `json:"cool_down_id"`
}

type outExercise struct {
	Exercise    *domain.ExerciseID `json:"exercise_id"`
	Rest        *int               `json:"rest_seconds"`
	Calibration *[]outCalSet       `json:"calibration_sets"`
	Working     *[]outWorkSet      `json:"working_sets"`
	Reason      *string            `json:"reason"`
}

type outCalSet struct {
	Reps *int            `json:"reps"`
	Load json.RawMessage `json:"load_lb"`
}

type outWorkSet struct {
	Reps *int            `json:"reps"`
	Load json.RawMessage `json:"load_lb"`
	RIR  *int            `json:"rir_target"`
}

type outCardio struct {
	Exercise *domain.ExerciseID `json:"exercise_id"`
	Minutes  *int               `json:"minutes"`
}

// parse reads the output text of a call into a plan. It refuses an
// output that breaks the schema, a session count that is not the count
// of the request, and an exercise that is not in the request. It does
// not check a bound of a target: the policy does that (D-23).
func parse(text string, req Request) (Plan, error) {
	dec := json.NewDecoder(strings.NewReader(text))
	dec.DisallowUnknownFields()
	var out outPlan
	if err := dec.Decode(&out); err != nil {
		return Plan{}, malformed("not the JSON of the schema")
	}
	if _, err := dec.Token(); err != io.EOF {
		return Plan{}, malformed("text after the JSON")
	}
	if out.Summary == nil || out.Sessions == nil || out.Guidance == nil {
		return Plan{}, malformed("a field of the plan is missing")
	}
	if len(*out.Sessions) != req.Sessions {
		return Plan{}, malformed("%d sessions: want %d", len(*out.Sessions), req.Sessions)
	}
	known := map[domain.ExerciseID]bool{}
	for _, x := range req.Exercises {
		known[x.Exercise.ID] = true
	}
	cardio := map[domain.ExerciseID]bool{}
	for _, id := range req.Cardio {
		cardio[id] = true
	}

	var p Plan
	p.Summary = p.text("summary", *out.Summary, SummaryMax, SummaryTemplate)
	for i, g := range *out.Guidance {
		if err := checkGuidance(g, KindMobility, KindRecovery); err != nil {
			return Plan{}, malformed("guidance_ids[%d]: %v", i, err)
		}
		p.Guidance = append(p.Guidance, g)
	}
	for i, s := range *out.Sessions {
		where := fmt.Sprintf("sessions[%d]", i)
		if s.WarmUp == nil || s.Exercises == nil || s.Cardio == nil || s.CoolDown == nil {
			return Plan{}, malformed("%s: a field is missing", where)
		}
		if err := checkGuidance(*s.WarmUp, KindWarmUp); err != nil {
			return Plan{}, malformed("%s.warm_up_id: %v", where, err)
		}
		if err := checkGuidance(*s.CoolDown, KindCoolDown); err != nil {
			return Plan{}, malformed("%s.cool_down_id: %v", where, err)
		}
		if n := len(*s.Exercises); n > MaxSessionExercises {
			return Plan{}, malformed("%s: %d exercises: want %d or fewer", where, n, MaxSessionExercises)
		}
		sess := Session{Title: fmt.Sprintf("Session %d", i+1), WarmUp: *s.WarmUp, CoolDown: *s.CoolDown}
		seen := map[domain.ExerciseID]bool{}
		for j, e := range *s.Exercises {
			at := fmt.Sprintf("%s.exercises[%d]", where, j)
			t, err := target(e)
			if err != nil {
				return Plan{}, malformed("%s: %v", at, err)
			}
			if !known[t.Exercise] {
				return Plan{}, malformed("%s: not an exercise of the request", at)
			}
			if seen[t.Exercise] {
				return Plan{}, malformed("%s: the exercise is in the session two times", at)
			}
			seen[t.Exercise] = true
			reason := p.text(at+".reason", *e.Reason, ReasonMax, ReasonTemplate)
			sess.Exercises = append(sess.Exercises, Exercise{t, reason})
		}
		c := s.Cardio
		if c.Exercise == nil || c.Minutes == nil {
			return Plan{}, malformed("%s.cardio: a field is missing", where)
		}
		switch {
		case *c.Exercise == "" && *c.Minutes == 0 && len(cardio) > 0:
			return Plan{}, malformed("%s.cardio: no cardio: want %d to %d minutes of a cardio exercise of the request", where, MinCardioMinutes, MaxCardioMinutes)
		case *c.Exercise == "" && *c.Minutes == 0:
		case !cardio[*c.Exercise]:
			return Plan{}, malformed("%s.cardio: not a cardio exercise of the request", where)
		case *c.Minutes < MinCardioMinutes || *c.Minutes > MaxCardioMinutes:
			return Plan{}, malformed("%s.cardio: %d minutes: want %d to %d", where, *c.Minutes, MinCardioMinutes, MaxCardioMinutes)
		default:
			sess.Cardio = &domain.PlannedCardio{Exercise: *c.Exercise, Minutes: *c.Minutes}
		}
		if len(sess.Exercises) == 0 && sess.Cardio == nil {
			return Plan{}, malformed("%s: no exercise and no cardio", where)
		}
		p.Sessions = append(p.Sessions, sess)
	}
	return p, nil
}

// text gives a text of Luna after the filter: the text with its spaces
// made single, or the template when a rule blocks it (D-183).
func (p *Plan) text(where, s string, max int, template string) string {
	if rule, bad := Filter(s, max); bad {
		p.Filtered = append(p.Filtered, Filtered{where, rule})
		return template
	}
	return strings.Join(strings.Fields(s), " ")
}

func checkGuidance(id GuidanceID, kinds ...GuidanceKind) error {
	g, ok := GuidanceItemOf(id)
	if !ok {
		return errors.New("not an id of the guidance catalog")
	}
	for _, k := range kinds {
		if g.Kind == k {
			return nil
		}
	}
	return fmt.Errorf("an id of kind %q", g.Kind)
}

func target(e outExercise) (domain.PlannedExercise, error) {
	if e.Exercise == nil || e.Rest == nil || e.Calibration == nil || e.Working == nil || e.Reason == nil {
		return domain.PlannedExercise{}, errors.New("a field is missing")
	}
	t := domain.PlannedExercise{Exercise: *e.Exercise, RestSeconds: *e.Rest}
	for i, s := range *e.Calibration {
		if s.Reps == nil || s.Load == nil {
			return t, fmt.Errorf("calibration_sets[%d]: a field is missing", i)
		}
		l, err := load(s.Load)
		if err != nil {
			return t, fmt.Errorf("calibration_sets[%d]: %v", i, err)
		}
		t.Calibration = append(t.Calibration, domain.CalibrationSet{Reps: *s.Reps, Load: l})
	}
	for i, s := range *e.Working {
		if s.Reps == nil || s.Load == nil || s.RIR == nil {
			return t, fmt.Errorf("working_sets[%d]: a field is missing", i)
		}
		l, err := load(s.Load)
		if err != nil {
			return t, fmt.Errorf("working_sets[%d]: %v", i, err)
		}
		t.Working = append(t.Working, domain.WorkingSet{Reps: *s.Reps, Load: l, RIR: *s.RIR})
	}
	return t, nil
}

// load reads a JSON number of lb into tenths of a pound. It refuses a
// load that is not a whole number of tenths, because a change of the
// value of Luna would hide it from the policy. A load out of the bounds
// stays: the policy refuses it (D-23).
func load(raw json.RawMessage) (domain.Load, error) {
	if len(raw) == 0 || (raw[0] != '-' && (raw[0] < '0' || raw[0] > '9')) {
		return 0, errors.New("load_lb: not a number")
	}
	f, err := strconv.ParseFloat(string(raw), 64)
	if err != nil || math.Abs(f) > 100_000 {
		return 0, errors.New("load_lb: not a number in range")
	}
	t := math.Round(f * 10)
	if math.Abs(f*10-t) > 1e-6 {
		return 0, errors.New("load_lb: not a whole number of tenths of a pound")
	}
	return domain.Load(t), nil
}

// encode gives the output text of a plan, in the form of the schema.
// The fake provider uses it.
func encode(v any) []byte {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
	return bytes.TrimSpace(b.Bytes())
}

// Reason is the reason of the reviser for one exercise (D-288). Sets
// names each logged set of the last session that the reason uses. Text
// is the reason with its spaces made single, or "" when the filter
// blocked it, and Blocked then names the rule. The text of Luna never
// goes into a log (D-80).
type Reason struct {
	Exercise domain.ExerciseID
	Sets     []SetRef
	Text     string
	Blocked  FilterRule
}

// SetRef names one logged set of a session: its kind, and its number
// inside its kind, from 1.
type SetRef struct {
	Kind   domain.SetKind
	Number int
}

type outReasons struct {
	Reasons *[]outReason `json:"reasons"`
}

type outReason struct {
	Exercise *domain.ExerciseID `json:"exercise_id"`
	Sets     *[]outSetRef       `json:"logged_sets"`
	Reason   *string            `json:"reason"`
}

type outSetRef struct {
	Kind   *domain.SetKind `json:"kind"`
	Number *int            `json:"number"`
}

// parseReasons reads the output text of a reviser call. It refuses an
// output that breaks the schema, an exercise that is not in the
// request, and an exercise with two reasons. An exercise with no reason
// is valid: the reason of the rules shows for it. It does not check a
// set against the log: the caller does that (D-288).
func parseReasons(text string, req Request) ([]Reason, error) {
	dec := json.NewDecoder(strings.NewReader(text))
	dec.DisallowUnknownFields()
	var out outReasons
	if err := dec.Decode(&out); err != nil {
		return nil, malformed("not the JSON of the schema")
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, malformed("text after the JSON")
	}
	if out.Reasons == nil {
		return nil, malformed("a field of the reasons is missing")
	}
	known := map[domain.ExerciseID]bool{}
	for _, x := range req.Exercises {
		known[x.Exercise.ID] = true
	}
	seen := map[domain.ExerciseID]bool{}
	var list []Reason
	for i, r := range *out.Reasons {
		at := fmt.Sprintf("reasons[%d]", i)
		if r.Exercise == nil || r.Sets == nil || r.Reason == nil {
			return nil, malformed("%s: a field is missing", at)
		}
		if !known[*r.Exercise] {
			return nil, malformed("%s: not an exercise of the request", at)
		}
		if seen[*r.Exercise] {
			return nil, malformed("%s: the exercise has two reasons", at)
		}
		seen[*r.Exercise] = true
		reason := Reason{Exercise: *r.Exercise}
		for j, s := range *r.Sets {
			if s.Kind == nil || s.Number == nil {
				return nil, malformed("%s.logged_sets[%d]: a field is missing", at, j)
			}
			if *s.Kind != domain.SetWorking && *s.Kind != domain.SetCalibration {
				return nil, malformed("%s.logged_sets[%d]: not a kind of set", at, j)
			}
			reason.Sets = append(reason.Sets, SetRef{*s.Kind, *s.Number})
		}
		if rule, bad := Filter(*r.Reason, ReasonMax); bad {
			reason.Blocked = rule
		} else {
			reason.Text = strings.Join(strings.Fields(*r.Reason), " ")
		}
		list = append(list, reason)
	}
	return list, nil
}
