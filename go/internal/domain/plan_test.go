package domain

import (
	"errors"
	"testing"
)

func TestWorkingSetCheck(t *testing.T) {
	for _, tc := range []struct {
		name string
		set  WorkingSet
		ok   bool
	}{
		{"10 reps at 12.5 lb, 2 RIR", WorkingSet{10, 125, 2}, true},
		{"0 RIR", WorkingSet{8, Pounds(50), 0}, true},
		{"0 reps", WorkingSet{0, Pounds(50), 2}, false},
		{"load 0", WorkingSet{10, 0, 2}, false},
		{"negative RIR", WorkingSet{10, Pounds(50), -1}, false},
	} {
		err := tc.set.Check()
		if (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("%s: Check() = %v, want ok %v", tc.name, err, tc.ok)
		}
	}
}

func TestCalibrationSetCheck(t *testing.T) {
	for _, tc := range []struct {
		name string
		set  CalibrationSet
		ok   bool
	}{
		{"10 reps at 40 lb", CalibrationSet{10, Pounds(40)}, true},
		{"0 reps", CalibrationSet{0, Pounds(40)}, false},
		{"negative load", CalibrationSet{10, -Pounds(40)}, false},
	} {
		err := tc.set.Check()
		if (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("%s: Check() = %v, want ok %v", tc.name, err, tc.ok)
		}
	}
}

func TestPlannedExerciseCheck(t *testing.T) {
	c := DefaultCatalog()
	work := []WorkingSet{{10, Pounds(50), 2}}
	for _, tc := range []struct {
		name string
		ex   PlannedExercise
		ok   bool
	}{
		{"working sets", PlannedExercise{Exercise: "chest_press", RestSeconds: 90, Working: work}, true},
		{"with calibration", PlannedExercise{Exercise: "db_goblet_squat", Calibration: []CalibrationSet{{10, Pounds(20)}}, Working: work}, true},
		{"unknown exercise", PlannedExercise{Exercise: "pec_fly", Working: work}, false},
		{"cardio exercise", PlannedExercise{Exercise: "treadmill", Working: work}, false},
		{"negative rest", PlannedExercise{Exercise: "chest_press", RestSeconds: -1, Working: work}, false},
		{"no working set", PlannedExercise{Exercise: "chest_press", Calibration: []CalibrationSet{{10, Pounds(20)}}}, false},
		{"bad calibration set", PlannedExercise{Exercise: "chest_press", Calibration: []CalibrationSet{{0, Pounds(20)}}, Working: work}, false},
		{"bad working set", PlannedExercise{Exercise: "chest_press", Working: []WorkingSet{{10, 0, 2}}}, false},
	} {
		err := tc.ex.Check(c)
		if (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("%s: Check() = %v, want ok %v", tc.name, err, tc.ok)
		}
	}
}

func TestPlannedCardioCheck(t *testing.T) {
	c := DefaultCatalog()
	for _, tc := range []struct {
		name   string
		cardio PlannedCardio
		ok     bool
	}{
		{"20 minutes", PlannedCardio{"elliptical", 20}, true},
		{"0 minutes", PlannedCardio{"elliptical", 0}, false},
		{"not cardio", PlannedCardio{"leg_press", 20}, false},
		{"unknown", PlannedCardio{"spin_bike", 20}, false},
	} {
		err := tc.cardio.Check(c)
		if (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("%s: Check() = %v, want ok %v", tc.name, err, tc.ok)
		}
	}
}

func TestPlanCheck(t *testing.T) {
	c := DefaultCatalog()
	press := PlannedExercise{Exercise: "chest_press", RestSeconds: 90, Working: []WorkingSet{{10, Pounds(50), 2}}}
	bike := &PlannedCardio{"upright_bike", 15}
	for _, tc := range []struct {
		name string
		plan Plan
		ok   bool
	}{
		{"one day", Plan{Days: []PlanDay{{Title: "A", Exercises: []PlannedExercise{press}, Cardio: bike}}}, true},
		{"cardio day", Plan{Days: []PlanDay{{Title: "B", Cardio: bike}}}, true},
		{"no day", Plan{}, false},
		{"empty day", Plan{Days: []PlanDay{{Title: "A", Exercises: []PlannedExercise{press}}, {Title: "B"}}}, false},
		{"bad exercise", Plan{Days: []PlanDay{{Exercises: []PlannedExercise{{Exercise: "chest_press"}}}}}, false},
		{"bad cardio", Plan{Days: []PlanDay{{Exercises: []PlannedExercise{press}, Cardio: &PlannedCardio{"chest_press", 15}}}}, false},
	} {
		err := tc.plan.Check(c)
		if (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("%s: Check() = %v, want ok %v", tc.name, err, tc.ok)
		}
	}
}
