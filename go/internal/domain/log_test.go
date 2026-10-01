package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }

func TestPainCheck(t *testing.T) {
	for _, tc := range []struct {
		pain Pain
		ok   bool
	}{
		{0, true}, {5, true}, {10, true}, {-1, false}, {11, false},
	} {
		err := tc.pain.Check()
		if (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("Pain(%d).Check() = %v, want ok %v", tc.pain, err, tc.ok)
		}
	}
}

func TestSetLogCheck(t *testing.T) {
	for _, tc := range []struct {
		name string
		set  SetLog
		ok   bool
	}{
		{"working set", SetLog{Kind: SetWorking, Reps: 10, Weight: Pounds(50), RIR: 2}, true},
		{"calibration set at 12.5 lb", SetLog{Kind: SetCalibration, Reps: 10, Weight: 125, RIR: 4}, true},
		{"0 reps", SetLog{Kind: SetWorking, Reps: 0, Weight: Pounds(50), RIR: 0}, true},
		{"high reps and RIR", SetLog{Kind: SetWorking, Reps: 60, Weight: Pounds(5), RIR: 12}, true},
		{"pain and note", SetLog{Kind: SetWorking, Reps: 8, Weight: Pounds(50), RIR: 2, Pain: ptr(Pain(3)), Note: "left knee"}, true},
		{"no kind", SetLog{Reps: 10, Weight: Pounds(50), RIR: 2}, false},
		{"unknown kind", SetLog{Kind: "warmup", Reps: 10, Weight: Pounds(50), RIR: 2}, false},
		{"negative reps", SetLog{Kind: SetWorking, Reps: -1, Weight: Pounds(50), RIR: 2}, false},
		{"weight 0", SetLog{Kind: SetWorking, Reps: 10, RIR: 2}, false},
		{"negative RIR", SetLog{Kind: SetWorking, Reps: 10, Weight: Pounds(50), RIR: -1}, false},
		{"pain 11", SetLog{Kind: SetWorking, Reps: 10, Weight: Pounds(50), RIR: 2, Pain: ptr(Pain(11))}, false},
	} {
		err := tc.set.Check()
		if (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("%s: Check() = %v, want ok %v", tc.name, err, tc.ok)
		}
	}
}

func TestExerciseLogCheck(t *testing.T) {
	c := DefaultCatalog()
	sets := []SetLog{{Kind: SetWorking, Reps: 10, Weight: Pounds(50), RIR: 2}}
	for _, tc := range []struct {
		name string
		log  ExerciseLog
		ok   bool
	}{
		{"sets", ExerciseLog{Exercise: "seated_row", Sets: sets}, true},
		{"skipped", ExerciseLog{Exercise: "seated_row", Skipped: true}, true},
		{"skipped with sets", ExerciseLog{Exercise: "seated_row", Skipped: true, Sets: sets}, false},
		{"no set, not skipped", ExerciseLog{Exercise: "seated_row"}, false},
		{"unknown exercise", ExerciseLog{Exercise: "pec_fly", Sets: sets}, false},
		{"cardio exercise", ExerciseLog{Exercise: "rowing_machine", Sets: sets}, false},
		{"bad set", ExerciseLog{Exercise: "seated_row", Sets: []SetLog{{Kind: SetWorking, Reps: -2, Weight: Pounds(50)}}}, false},
	} {
		err := tc.log.Check(c)
		if (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("%s: Check() = %v, want ok %v", tc.name, err, tc.ok)
		}
	}
}

func TestCardioLogCheck(t *testing.T) {
	c := DefaultCatalog()
	base := func() CardioLog { return CardioLog{Exercise: "treadmill", Duration: 20 * time.Minute, Effort: 5} }
	for _, tc := range []struct {
		name   string
		change func(*CardioLog)
		ok     bool
	}{
		{"necessary fields alone", func(*CardioLog) {}, true},
		{"each optional field", func(l *CardioLog) {
			l.Distance, l.Resistance, l.Pain, l.Note = ptr(Distance(23)), ptr(4), ptr(Pain(0)), "easy"
		}, true},
		{"effort 1", func(l *CardioLog) { l.Effort = 1 }, true},
		{"effort 10", func(l *CardioLog) { l.Effort = 10 }, true},
		{"effort 0", func(l *CardioLog) { l.Effort = 0 }, false},
		{"effort 11", func(l *CardioLog) { l.Effort = 11 }, false},
		{"duration 0", func(l *CardioLog) { l.Duration = 0 }, false},
		{"negative distance", func(l *CardioLog) { l.Distance = ptr(Distance(-1)) }, false},
		{"negative resistance", func(l *CardioLog) { l.Resistance = ptr(-1) }, false},
		{"pain 11", func(l *CardioLog) { l.Pain = ptr(Pain(11)) }, false},
		{"not cardio", func(l *CardioLog) { l.Exercise = "leg_press" }, false},
		{"unknown", func(l *CardioLog) { l.Exercise = "ski_erg" }, false},
	} {
		l := base()
		tc.change(&l)
		err := l.Check(c)
		if (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("%s: Check() = %v, want ok %v", tc.name, err, tc.ok)
		}
	}
}

func TestSessionCheck(t *testing.T) {
	c := DefaultCatalog()
	row := ExerciseLog{Exercise: "seated_row", Sets: []SetLog{{Kind: SetWorking, Reps: 10, Weight: Pounds(50), RIR: 2}}}
	bike := CardioLog{Exercise: "upright_bike", Duration: 10 * time.Minute, Effort: 4}
	for _, tc := range []struct {
		name    string
		session Session
		ok      bool
	}{
		{"full", Session{Date: "2026-09-30", Exercises: []ExerciseLog{row}, Cardio: []CardioLog{bike}}, true},
		{"ended early", Session{Date: "2026-09-30", EndedEarly: true, Exercises: []ExerciseLog{row, {Exercise: "chest_press", Skipped: true}}}, true},
		{"cardio alone", Session{Date: "2026-09-30", Cardio: []CardioLog{bike}}, true},
		{"no log", Session{Date: "2026-09-30"}, false},
		{"no date", Session{Exercises: []ExerciseLog{row}}, false},
		{"date with a time", Session{Date: "2026-09-30T10:00:00Z", Exercises: []ExerciseLog{row}}, false},
		{"no such day", Session{Date: "2026-02-30", Exercises: []ExerciseLog{row}}, false},
		{"bad exercise log", Session{Date: "2026-09-30", Exercises: []ExerciseLog{{Exercise: "seated_row"}}}, false},
		{"bad cardio log", Session{Date: "2026-09-30", Cardio: []CardioLog{{Exercise: "upright_bike", Effort: 4}}}, false},
	} {
		err := tc.session.Check(c)
		if (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("%s: Check() = %v, want ok %v", tc.name, err, tc.ok)
		}
	}
}

// A check error holds ids and numbers alone, and never a note or a date
// of the owner (D-80).
func TestCheckErrorHoldsNoOwnerText(t *testing.T) {
	c := DefaultCatalog()
	const note, date = "private note", "2026-09-30"
	s := Session{Date: date, Exercises: []ExerciseLog{{Exercise: "seated_row", Sets: []SetLog{
		{Kind: SetWorking, Reps: -1, Weight: Pounds(50), Note: note},
	}}}, Cardio: []CardioLog{{Exercise: "treadmill", Effort: 0, Note: note}}}
	err := s.Check(c)
	if err == nil {
		t.Fatal("Check() = nil, want an error")
	}
	if msg := err.Error(); strings.Contains(msg, note) || strings.Contains(msg, date) {
		t.Errorf("Check() = %q holds owner text", msg)
	}
	s.Exercises = nil
	if err := s.Check(c); err == nil || strings.Contains(err.Error(), note) || strings.Contains(err.Error(), date) {
		t.Errorf("Check() of the cardio log = %v, want an error with no owner text", err)
	}
}
