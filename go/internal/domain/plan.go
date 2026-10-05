package domain

// WorkingSet is one target set of a plan: reps, a load, and reps in
// reserve. The policy holds the safe range of each value (D-37).
type WorkingSet struct {
	Reps int
	Load Load
	RIR  int
}

// Check refuses fewer than 1 rep, a load of 0 or less, and reps in
// reserve below 0.
func (s WorkingSet) Check() error {
	if s.Reps < 1 {
		return invalid("reps %d: want 1 or more", s.Reps)
	}
	if err := s.Load.Check(); err != nil {
		return err
	}
	if s.RIR < 0 {
		return invalid("rir %d: want 0 or more", s.RIR)
	}
	return nil
}

// CalibrationSet is one set before the working sets that finds the
// load. It is not a working set, and it stops at 3 to 4 reps in reserve
// (D-150). The policy holds that rule, so the set holds no RIR target.
// Only a target of policy version 6 or earlier has one. From version 7,
// the first working set is the calibration (D-297).
type CalibrationSet struct {
	Reps int
	Load Load
}

// Check refuses fewer than 1 rep and a load of 0 or less.
func (s CalibrationSet) Check() error {
	if s.Reps < 1 {
		return invalid("reps %d: want 1 or more", s.Reps)
	}
	return s.Load.Check()
}

// PlannedExercise is one exercise of a plan day, with its rest, its
// calibration sets, and its working sets. FirstSetCalibration tells
// that the first working set is the calibration: the owner changes the
// weight during its first reps, and the other working sets use the
// weight that the owner logged for it (D-297, D-299). FollowMax is the
// heaviest load that the other working sets can use after a heavier
// first set, and 0 when they keep their loads (D-306, D-307).
type PlannedExercise struct {
	Exercise            ExerciseID
	RestSeconds         int
	Calibration         []CalibrationSet
	Working             []WorkingSet
	FirstSetCalibration bool
	FollowMax           Load
}

// Check refuses an exercise that is not in the catalog, a cardio
// exercise, a rest below 0, no working set, and each bad set.
func (p PlannedExercise) Check(c Catalog) error {
	e, ok := c.Exercise(p.Exercise)
	if !ok {
		return invalid("exercise %q: not in the catalog", p.Exercise)
	}
	if e.Kind == KindCardio {
		return invalid("exercise %q: a cardio exercise with sets", p.Exercise)
	}
	if p.RestSeconds < 0 {
		return invalid("exercise %q rest %d s: want 0 or more", p.Exercise, p.RestSeconds)
	}
	if len(p.Working) == 0 {
		return invalid("exercise %q: no working set", p.Exercise)
	}
	for i, s := range p.Calibration {
		if err := s.Check(); err != nil {
			return invalid("exercise %q calibration[%d]: %v", p.Exercise, i, err)
		}
	}
	for i, s := range p.Working {
		if err := s.Check(); err != nil {
			return invalid("exercise %q working[%d]: %v", p.Exercise, i, err)
		}
	}
	// The limit is 0, or the load of the first working set or more
	// (D-306, D-307).
	if p.FollowMax != 0 && p.FollowMax < p.Working[0].Load {
		return invalid("exercise %q follow max %s: want 0, or %s or more", p.Exercise, p.FollowMax, p.Working[0].Load)
	}
	return nil
}

// PlannedCardio is the optional cardio of a plan day (D-44).
type PlannedCardio struct {
	Exercise ExerciseID
	Minutes  int
}

// Check refuses an exercise that is not a cardio exercise of the
// catalog, and fewer than 1 minute.
func (p PlannedCardio) Check(c Catalog) error {
	e, ok := c.Exercise(p.Exercise)
	if !ok {
		return invalid("cardio %q: not in the catalog", p.Exercise)
	}
	if e.Kind != KindCardio {
		return invalid("cardio %q: kind %q", p.Exercise, e.Kind)
	}
	if p.Minutes < 1 {
		return invalid("cardio %q minutes %d: want 1 or more", p.Exercise, p.Minutes)
	}
	return nil
}

// PlanDay is one session of a plan.
type PlanDay struct {
	Title     string
	Exercises []PlannedExercise
	Cardio    *PlannedCardio
}

// Plan holds the days of a plan. The plan adapts after each session and
// has no fixed block (D-43). The policy checks each target before the
// owner sees it (D-23).
type Plan struct {
	Days []PlanDay
}

// Check refuses a plan with no day, a day with no exercise and no
// cardio, and each bad exercise or cardio.
func (p Plan) Check(c Catalog) error {
	if len(p.Days) == 0 {
		return invalid("plan: no day")
	}
	for d, day := range p.Days {
		if len(day.Exercises) == 0 && day.Cardio == nil {
			return invalid("days[%d]: no exercise and no cardio", d)
		}
		for _, e := range day.Exercises {
			if err := e.Check(c); err != nil {
				return invalid("days[%d]: %v", d, err)
			}
		}
		if day.Cardio != nil {
			if err := day.Cardio.Check(c); err != nil {
				return invalid("days[%d]: %v", d, err)
			}
		}
	}
	return nil
}
