package domain

// The body tables of D-208 and D-210, with the values of D-218 to D-220.
// Change BodyTablesVersion with each change of this data. Section 5.15
// of docs/research/exercise-safety.md gives the research of each row
// (D-38).
//
// An area is in a row when the joint moves under the load, holds a large
// moment of it, or takes it directly (D-218). Hip adduction and the
// cardio exercises have no primary group (D-219).

// BodyTablesVersion is the version of the data of DefaultBodyTables.
const BodyTablesVersion = 1

// DefaultBodyTables gives a new copy of the tables, so a caller can not
// change the data of another caller.
func DefaultBodyTables() BodyTables {
	const (
		sh, el, wr, lb, hp, kn, an = AreaShoulder, AreaElbow, AreaWrist, AreaLowerBack, AreaHip, AreaKnee, AreaAnkle
	)
	const (
		ch, bk, sd, bi, tr = GroupChest, GroupBack, GroupShoulders, GroupBiceps, GroupTriceps
		qu, hs, gl, ca, co = GroupQuadriceps, GroupHamstrings, GroupGlutes, GroupCalves, GroupCore
	)
	rows := []struct {
		id     ExerciseID
		areas  []Area
		groups []MuscleGroup
	}{
		{"leg_press", []Area{lb, hp, kn, an}, []MuscleGroup{qu, gl}},
		{"leg_extension", []Area{kn}, []MuscleGroup{qu}},
		{"seated_leg_curl", []Area{kn}, []MuscleGroup{hs}},
		{"lying_leg_curl", []Area{kn}, []MuscleGroup{hs}},
		{"hip_abduction", []Area{hp}, []MuscleGroup{gl}},
		{"hip_adduction", []Area{hp}, nil},
		{"calf_raise", []Area{lb, an}, []MuscleGroup{ca}},
		{"chest_press", []Area{sh, el, wr}, []MuscleGroup{ch, tr}},
		{"shoulder_press", []Area{sh, el, wr}, []MuscleGroup{sd, tr}},
		{"seated_row", []Area{sh, el, wr}, []MuscleGroup{bk, bi}},
		{"biceps_curl", []Area{el, wr}, []MuscleGroup{bi}},
		{"abdominal_crunch", []Area{lb}, []MuscleGroup{co}},
		{"back_extension", []Area{lb, hp}, []MuscleGroup{bk}},
		{"lat_pulldown", []Area{sh, el, wr}, []MuscleGroup{bk, bi}},
		{"triceps_pulldown", []Area{el, wr}, []MuscleGroup{tr}},
		{"db_flat_bench_press", []Area{sh, el, wr}, []MuscleGroup{ch, tr}},
		{"db_incline_bench_press", []Area{sh, el, wr}, []MuscleGroup{ch, sd, tr}},
		{"db_seated_shoulder_press", []Area{sh, el, wr}, []MuscleGroup{sd, tr}},
		{"db_one_arm_row", []Area{sh, el, wr, lb}, []MuscleGroup{bk, bi}},
		{"db_biceps_curl", []Area{el, wr}, []MuscleGroup{bi}},
		{"db_hammer_curl", []Area{el, wr}, []MuscleGroup{bi}},
		{"db_lateral_raise", []Area{sh, el, wr}, []MuscleGroup{sd}},
		{"db_romanian_deadlift", []Area{wr, lb, hp}, []MuscleGroup{hs, gl}},
		{"db_goblet_squat", []Area{el, wr, lb, hp, kn, an}, []MuscleGroup{qu, gl}},
		{"treadmill", []Area{hp, kn, an}, nil},
		{"upright_bike", []Area{hp, kn, an}, nil},
		{"recumbent_bike", []Area{hp, kn, an}, nil},
		{"rowing_machine", Areas(), nil},
		{"elliptical", []Area{hp, kn, an}, nil},
		{"stair_climber", []Area{hp, kn, an}, nil},
	}
	t := BodyTables{
		Version: BodyTablesVersion,
		Areas:   map[ExerciseID][]Area{},
		Groups:  map[ExerciseID][]MuscleGroup{},
		Templates: []GoalTemplate{
			{TemplateGeneralFitness, "General fitness", MuscleGroups()},
			{TemplateStrength, "Strength", []MuscleGroup{ch, bk, sd, qu, hs, gl}},
		},
	}
	for _, r := range rows {
		t.Areas[r.id] = append([]Area{}, r.areas...)
		t.Groups[r.id] = append([]MuscleGroup{}, r.groups...)
	}
	return t
}
