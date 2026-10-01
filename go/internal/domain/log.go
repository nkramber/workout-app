package domain

import "time"

// Pain is a pain rating from 0 to 10 (D-162). A log holds it as an
// optional value: nil means no report.
type Pain int

// PainMax is the top of the pain rating (D-162).
const PainMax Pain = 10

// Check refuses a rating outside 0 to 10.
func (p Pain) Check() error {
	if p < 0 || p > PainMax {
		return invalid("pain %d: want 0 to %d", p, PainMax)
	}
	return nil
}

func checkPain(p *Pain) error {
	if p == nil {
		return nil
	}
	return p.Check()
}

// SetKind tells a working set from a calibration set in a log.
type SetKind string

const (
	SetWorking     SetKind = "working"
	SetCalibration SetKind = "calibration"
)

// SetLog is one logged set: reps, the weight, and reps in reserve, with
// an optional pain rating and an optional note (D-57). The note is text
// of the owner, so no check error and no log line holds it (D-80).
type SetLog struct {
	Kind   SetKind
	Reps   int
	Weight Load
	RIR    int
	Pain   *Pain
	Note   string
}

// Check refuses an unknown kind, reps or reps in reserve below 0, a
// weight of 0 or less, and a bad pain rating. A set of 0 reps is valid,
// because it tells that the load was too heavy (D-164).
func (s SetLog) Check() error {
	if s.Kind != SetWorking && s.Kind != SetCalibration {
		return invalid("set kind %q: want %q or %q", s.Kind, SetWorking, SetCalibration)
	}
	if s.Reps < 0 {
		return invalid("reps %d: want 0 or more", s.Reps)
	}
	if err := s.Weight.Check(); err != nil {
		return err
	}
	if s.RIR < 0 {
		return invalid("rir %d: want 0 or more", s.RIR)
	}
	return checkPain(s.Pain)
}

// ExerciseLog holds the logged sets of one exercise in a session. A
// skipped exercise holds no set (D-47, D-63).
type ExerciseLog struct {
	Exercise ExerciseID
	Skipped  bool
	Sets     []SetLog
}

// Check refuses an exercise that is not in the catalog, a cardio
// exercise, sets on a skipped exercise, no set on an exercise that is
// not skipped, and each bad set.
func (l ExerciseLog) Check(c Catalog) error {
	e, ok := c.Exercise(l.Exercise)
	if !ok {
		return invalid("exercise %q: not in the catalog", l.Exercise)
	}
	if e.Kind == KindCardio {
		return invalid("exercise %q: a cardio exercise with sets", l.Exercise)
	}
	if l.Skipped && len(l.Sets) > 0 {
		return invalid("exercise %q: skipped with %d sets", l.Exercise, len(l.Sets))
	}
	if !l.Skipped && len(l.Sets) == 0 {
		return invalid("exercise %q: no set and not skipped", l.Exercise)
	}
	for i, s := range l.Sets {
		if err := s.Check(); err != nil {
			return invalid("exercise %q sets[%d]: %v", l.Exercise, i, err)
		}
	}
	return nil
}

// EffortMin and EffortMax bound the effort rating of a cardio log
// (D-123).
const (
	EffortMin = 1
	EffortMax = 10
)

// Distance is a distance in tenths of a mile (D-165).
type Distance int

// CardioLog is one logged cardio exercise. The duration and the effort
// rating are necessary. The distance, the resistance level, the pain
// rating, and the note are optional. The log holds no calories and no
// heart rate (D-123). The note is text of the owner (D-80).
type CardioLog struct {
	Exercise   ExerciseID
	Duration   time.Duration
	Effort     int
	Distance   *Distance
	Resistance *int
	Pain       *Pain
	Note       string
}

// Check refuses an exercise that is not a cardio exercise of the
// catalog, a duration of 0 or less, an effort outside 1 to 10, a
// distance or a resistance level below 0, and a bad pain rating.
func (l CardioLog) Check(c Catalog) error {
	e, ok := c.Exercise(l.Exercise)
	if !ok {
		return invalid("cardio %q: not in the catalog", l.Exercise)
	}
	if e.Kind != KindCardio {
		return invalid("cardio %q: kind %q", l.Exercise, e.Kind)
	}
	if l.Duration <= 0 {
		return invalid("cardio %q duration %s: want more than 0", l.Exercise, l.Duration)
	}
	if l.Effort < EffortMin || l.Effort > EffortMax {
		return invalid("cardio %q effort %d: want %d to %d", l.Exercise, l.Effort, EffortMin, EffortMax)
	}
	if l.Distance != nil && *l.Distance < 0 {
		return invalid("cardio %q distance %d tenths of a mile: want 0 or more", l.Exercise, *l.Distance)
	}
	if l.Resistance != nil && *l.Resistance < 0 {
		return invalid("cardio %q resistance %d: want 0 or more", l.Exercise, *l.Resistance)
	}
	if err := checkPain(l.Pain); err != nil {
		return invalid("cardio %q: %v", l.Exercise, err)
	}
	return nil
}

// DateLayout is the form of the date of a session: a calendar date with
// no time and no zone.
const DateLayout = "2006-01-02"

// Session is one logged workout: its date, the exercise logs, and the
// cardio logs. EndedEarly records a "finish now" action, which skips
// each remaining exercise (D-63).
type Session struct {
	Date       string
	EndedEarly bool
	Exercises  []ExerciseLog
	Cardio     []CardioLog
}

// Check refuses a date that is not a calendar date, a session with no
// log, and each bad log. An error holds no date, because the date is
// data of the log (D-80).
func (s Session) Check(c Catalog) error {
	if _, err := time.Parse(DateLayout, s.Date); err != nil {
		return invalid("session date: want the form %s", DateLayout)
	}
	if len(s.Exercises) == 0 && len(s.Cardio) == 0 {
		return invalid("session: no log")
	}
	for _, l := range s.Exercises {
		if err := l.Check(c); err != nil {
			return invalid("session: %v", err)
		}
	}
	for _, l := range s.Cardio {
		if err := l.Check(c); err != nil {
			return invalid("session: %v", err)
		}
	}
	return nil
}
