package domain

// The first product catalog of D-155. Change CatalogVersion with each
// change of this data. Never change or reuse an id.

// CatalogVersion is the version of the data of DefaultCatalog.
const CatalogVersion = 1

// DefaultCatalog gives a new copy of the first product catalog, so a
// caller can not change the data of another caller.
func DefaultCatalog() Catalog {
	return Catalog{
		Version: CatalogVersion,
		Machines: []Machine{
			{"leg_press", "Leg press", KindMachine},
			{"leg_extension", "Leg extension", KindMachine},
			{"seated_leg_curl", "Seated leg curl", KindMachine},
			{"lying_leg_curl", "Lying leg curl", KindMachine},
			{"hip_abduction_adduction", "Hip abduction or adduction", KindMachine},
			{"calf_raise", "Calf raise", KindMachine},
			{"chest_press", "Chest press", KindMachine},
			{"shoulder_press", "Shoulder press", KindMachine},
			{"seated_row", "Seated row", KindMachine},
			{"biceps_curl", "Biceps curl", KindMachine},
			{"abdominal_crunch", "Abdominal crunch", KindMachine},
			{"back_extension", "Back extension", KindMachine},
			{"cable_station", "Cable station", KindCable},
			{"dumbbells", "Dumbbells with an adjustable bench", KindDumbbell},
			{"treadmill", "Treadmill", KindCardio},
			{"upright_bike", "Upright bike", KindCardio},
			{"recumbent_bike", "Recumbent bike", KindCardio},
			{"rowing_machine", "Rowing machine", KindCardio},
			{"elliptical", "Elliptical", KindCardio},
			{"stair_climber", "Stair climber", KindCardio},
		},
		Exercises: []Exercise{
			{"leg_press", "Leg press", "leg_press", KindMachine, RegionLowerPush},
			{"leg_extension", "Leg extension", "leg_extension", KindMachine, RegionLowerPush},
			{"seated_leg_curl", "Seated leg curl", "seated_leg_curl", KindMachine, RegionLowerPull},
			{"lying_leg_curl", "Lying leg curl", "lying_leg_curl", KindMachine, RegionLowerPull},
			{"hip_abduction", "Hip abduction", "hip_abduction_adduction", KindMachine, RegionLowerPull},
			{"hip_adduction", "Hip adduction", "hip_abduction_adduction", KindMachine, RegionLowerPush},
			{"calf_raise", "Calf raise", "calf_raise", KindMachine, RegionLowerPush},
			{"chest_press", "Chest press", "chest_press", KindMachine, RegionUpperPush},
			{"shoulder_press", "Shoulder press", "shoulder_press", KindMachine, RegionUpperPush},
			{"seated_row", "Seated row", "seated_row", KindMachine, RegionUpperPull},
			{"biceps_curl", "Biceps curl", "biceps_curl", KindMachine, RegionUpperPull},
			{"abdominal_crunch", "Abdominal crunch", "abdominal_crunch", KindMachine, RegionCore},
			{"back_extension", "Back extension", "back_extension", KindMachine, RegionLowerPull},
			{"lat_pulldown", "Lat pulldown", "cable_station", KindCable, RegionUpperPull},
			{"triceps_pulldown", "Triceps pulldown", "cable_station", KindCable, RegionUpperPush},
			{"db_flat_bench_press", "Dumbbell flat bench press", "dumbbells", KindDumbbell, RegionUpperPush},
			{"db_incline_bench_press", "Dumbbell incline bench press", "dumbbells", KindDumbbell, RegionUpperPush},
			{"db_seated_shoulder_press", "Dumbbell seated shoulder press", "dumbbells", KindDumbbell, RegionUpperPush},
			{"db_one_arm_row", "Dumbbell one-arm row", "dumbbells", KindDumbbell, RegionUpperPull},
			{"db_biceps_curl", "Dumbbell biceps curl", "dumbbells", KindDumbbell, RegionUpperPull},
			{"db_hammer_curl", "Dumbbell hammer curl", "dumbbells", KindDumbbell, RegionUpperPull},
			{"db_lateral_raise", "Dumbbell lateral raise", "dumbbells", KindDumbbell, RegionUpperPush},
			{"db_romanian_deadlift", "Dumbbell Romanian deadlift", "dumbbells", KindDumbbell, RegionLowerPull},
			{"db_goblet_squat", "Dumbbell goblet squat", "dumbbells", KindDumbbell, RegionLowerPush},
			{"treadmill", "Treadmill", "treadmill", KindCardio, RegionCardio},
			{"upright_bike", "Upright bike", "upright_bike", KindCardio, RegionCardio},
			{"recumbent_bike", "Recumbent bike", "recumbent_bike", KindCardio, RegionCardio},
			{"rowing_machine", "Rowing machine", "rowing_machine", KindCardio, RegionCardio},
			{"elliptical", "Elliptical", "elliptical", KindCardio, RegionCardio},
			{"stair_climber", "Stair climber", "stair_climber", KindCardio, RegionCardio},
		},
	}
}
