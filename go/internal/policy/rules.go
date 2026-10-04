package policy

// RuleID is the stable id of one rule. A decision and a violation name
// the rules that applied. Do not reuse an id for a different rule.
type RuleID string

// The rules of the bounds of a target.
const (
	RuleRepBounds     RuleID = "reps.bounds"
	RuleRIRBounds     RuleID = "rir.bounds"
	RuleRIRPress      RuleID = "rir.dumbbell-press"
	RuleRestBounds    RuleID = "rest.bounds"
	RuleLoadAvailable RuleID = "load.available"
	RuleLoadRounding  RuleID = "load.rounding"
	RuleLoadCeiling   RuleID = "load.ceiling"
	RuleEffortCeiling RuleID = "effort.ceiling"
	RuleLoadRepair    RuleID = "load.repair"
)

// The rules of the next target.
const (
	RuleSkipped         RuleID = "progress.skipped"
	RulePainHold        RuleID = "progress.pain-hold"
	RuleShortfallSmall  RuleID = "progress.shortfall-small"
	RuleShortfallTwice  RuleID = "progress.shortfall-twice"
	RuleShortfallLowRIR RuleID = "progress.shortfall-low-rir"
	RuleShortfallRepeat RuleID = "progress.shortfall-repeat"
	RuleIncomplete      RuleID = "progress.incomplete"
	RuleLighterHold     RuleID = "progress.lighter-hold"
	RuleFailureHold     RuleID = "progress.failure-hold"
	RulePressHold       RuleID = "progress.press-hold"
	RuleEffortHold      RuleID = "progress.effort-hold"
	RuleAddReps         RuleID = "progress.add-reps"
	RuleLoadStep        RuleID = "progress.load-step"
	RuleNoHeavier       RuleID = "progress.no-heavier"
)

// The rules of the start and the calibration of a new exercise.
const (
	RuleStartEstimate    RuleID = "start.estimate"
	RuleStartLightest    RuleID = "start.lightest"
	RuleCalibrationSet   RuleID = "calibration.set"
	RuleCalibrationTable RuleID = "calibration.table"
	RuleCalibrationHold  RuleID = "calibration.reps-only"
)

// The rules of a return after a break.
const (
	RuleBreakShort       RuleID = "break.short"
	RuleBreakLong        RuleID = "break.long"
	RuleBreakRecalibrate RuleID = "break.recalibrate"
	RuleBreakFirst       RuleID = "break.first-sessions"
	RuleBreakSets        RuleID = "break.sets-restored"
)

// The rules of a missed session, of the reactive deload, and of an
// override of the owner.
const (
	RuleMissed         RuleID = "missed.hold"
	RuleMissedRestored RuleID = "missed.rir-restored"
	RuleDeload         RuleID = "deload.reactive"
	RuleOverride       RuleID = "override.bounds"
)

// The rules of a proposal and of the rules fallback.
const (
	RuleProposalExercise RuleID = "proposal.exercise"
	RuleFallback         RuleID = "fallback.rules"
)

// The rule of the warning of a pain report.
const RulePainWarning RuleID = "warning.pain"

// Rule is one rule of the policy: its id, what it requires, and the
// decisions and evidence ids that support it (D-38).
type Rule struct {
	ID      RuleID
	Text    string
	Sources []string
}

var rules = []Rule{
	{RuleRepBounds, "Each target set has 6 to 20 reps. The rep range of the exercise sets the target inside these limits.", []string{"D-167", "EV-20", "EV-21", "EV-27"}},
	{RuleRIRBounds, "Each target set has 1 to 3 reps in reserve. No target is failure.", []string{"D-37", "EV-1", "EV-23"}},
	{RuleRIRPress, "The flat, incline, and seated dumbbell press target 2 to 3 reps in reserve.", []string{"D-171", "EV-27", "EV-28", "EV-76", "EV-77", "EV-79"}},
	{RuleRestBounds, "Each exercise rests 60 to 180 seconds. Each plan gives 60 seconds to each exercise, the leg press too.", []string{"D-172", "D-279", "D-59", "EV-35", "EV-36", "EV-37"}},
	{RuleLoadAvailable, "Each load is an available weight of the machine. A dumbbell load is the load of one dumbbell, and the input refuses a dumbbell set above 100 lb.", []string{"D-54", "D-149", "D-155", "D-166"}},
	{RuleLoadRounding, "Each load is a multiple of 5 lb, or the weight that D-149 selects for a multiple of 5 lb.", []string{"D-65", "D-148", "D-149"}},
	{RuleLoadCeiling, "A proposed load is not more than the load of the next target of the policy. So a proposal adds at most one 5 lb step, and adds no load when the policy holds or lowers it.", []string{"D-23", "D-147"}},
	{RuleEffortCeiling, "A proposal has no more working sets than the next target of the policy. A proposed set at the load of the set of the same position of that target has no more reps and no fewer reps in reserve. A set at a lower load can have more reps, inside the rep bounds.", []string{"D-23", "D-147", "D-168", "D-186"}},
	{RuleLoadRepair, "When the last load is not a valid load of the machine, the next load is the weight that D-149 selects for the multiple of 5 lb at or below it. A repair raises a load only to the lightest weight of the machine.", []string{"D-65", "D-149"}},
	{RuleSkipped, "A skipped exercise repeats its target.", []string{"D-63", "D-64", "D-170"}},
	{RulePainHold, "A pain rating of 1 or more on a set gives no progression on that exercise in the next session. The target repeats.", []string{"D-40", "D-169", "EV-54", "EV-55"}},
	{RuleShortfallSmall, "A total shortfall of 1 or 2 reps repeats the target.", []string{"D-168", "EV-2"}},
	{RuleShortfallTwice, "A shortfall of more than 2 reps in 2 sessions in a row lowers the load one 5 lb step, with the same rep target. At the lightest weight, the reps go to the bottom of the range.", []string{"D-168", "D-174", "EV-2"}},
	{RuleShortfallLowRIR, "A shortfall of more than 2 reps, after sets at 0 or 1 reps in reserve, keeps the load and lowers the reps to the bottom of the range.", []string{"D-168", "EV-2", "EV-38"}},
	{RuleShortfallRepeat, "A shortfall of more than 2 reps with no other cause repeats the target.", []string{"D-168", "EV-2"}},
	{RuleIncomplete, "An exercise with a planned set that has no log gets no progression. An unlogged set is skipped work, not a missed rep.", []string{"D-63", "D-64", "D-170"}},
	{RuleLighterHold, "A logged weight below the target load gives no progression. The target repeats.", []string{"D-64", "D-173"}},
	{RuleFailureHold, "A logged set at 0 reps in reserve is a failure event. Failure events in 2 sessions in a row hold progression.", []string{"D-37", "D-168", "EV-23", "EV-33", "EV-34"}},
	{RulePressHold, "A logged set at 0 or 1 reps in reserve on a dumbbell press holds the load.", []string{"D-171", "EV-27", "EV-79"}},
	{RuleEffortHold, "Progression needs 3 or more reps in reserve on each set. Otherwise the target repeats.", []string{"D-147", "D-168"}},
	{RuleAddReps, "Each target set below the top of the rep range adds 2 reps, up to the top of the range.", []string{"D-147", "D-168", "EV-38"}},
	{RuleLoadStep, "At the top of the rep range, the load adds one 5 lb step and the reps go to the bottom of the range. The load rounds to the nearest 5 lb, down at a halfway value, and never more than 5 lb above the last load.", []string{"D-147", "D-148", "D-149", "EV-2"}},
	{RuleNoHeavier, "At the top of the rep range with no heavier weight inside one 5 lb step, the target repeats.", []string{"D-147", "D-149"}},
	{RuleStartEstimate, "A new exercise starts with one calibration set and 3 working sets at the bottom of the rep range, at 3 reps in reserve. The load is the estimate of the owner, rounded to the nearest 5 lb, with the weight that D-149 selects. After a break of 91 days or more before the estimate, the load is 70 percent of the estimate, and a halfway value rounds down. The input refuses an estimate outside the weights of the machine.", []string{"D-41", "D-148", "D-149", "D-150", "D-179", "D-180", "EV-22", "EV-48"}},
	{RuleStartLightest, "A new exercise with no estimate starts at the lightest weight of the machine, with one calibration set and 3 working sets at the bottom of the rep range, at 3 reps in reserve.", []string{"D-150", "D-178", "D-180"}},
	{RuleCalibrationSet, "The first 3 sessions of an exercise, and the first 3 sessions after a new calibration, start with one calibration set at the reps and the load of the first working set. A proposal for such a session needs one calibration set, at the reps and the load of its first working set, and at a load that is not more than the load of the policy. A proposal for another session has no calibration set.", []string{"D-150", "D-177", "D-181", "EV-27", "EV-83"}},
	{RuleCalibrationTable, "After a calibration set, 6 or more reps in reserve adds two 5 lb steps, 5 adds one step, 3 or 4 keeps the load, and 2 or less or a pain report removes one step. A session has one calibration set, so the table applies one time, and the working sets use its load.", []string{"D-149", "D-150", "D-177", "D-267", "EV-22", "EV-27", "EV-48"}},
	{RuleCalibrationHold, "In a calibration session, the reps can go up to the top of the range, but the load does not go up. The reps in reserve of these sessions are practice.", []string{"D-177", "EV-83"}},
	{RuleBreakShort, "After 14 to 27 days with no logged set of the exercise, the load goes down 10 percent and the target has one set fewer, at 3 reps in reserve. A halfway value rounds down.", []string{"D-148", "D-151", "D-179", "EV-43", "EV-46"}},
	{RuleBreakLong, "After 28 to 90 days with no logged set of the exercise, the load goes down 20 percent and the target has one set fewer, at 3 reps in reserve. A halfway value rounds down.", []string{"D-148", "D-151", "D-179", "EV-43", "EV-44"}},
	{RuleBreakRecalibrate, "After 91 days or more with no logged set of the exercise, the load goes down to 70 percent of the last load, at 3 reps in reserve, and the calibration starts again. A halfway value rounds down.", []string{"D-148", "D-150", "D-151", "D-179", "EV-43", "EV-44", "EV-47", "EV-53"}},
	{RuleBreakFirst, "The first sessions after a break of 14 days or more are the first 3 sessions or the first 14 days, the longer of the two. In them, each working set stops at 3 reps in reserve, the load does not go up, and the target keeps the sets of the return. The policy refuses a proposal with fewer reps in reserve or more sets.", []string{"D-37", "D-151", "D-179", "EV-43", "EV-53"}},
	{RuleBreakSets, "After the first sessions after a break, the target gets back the number of sets of the target before the break.", []string{"D-151", "D-179"}},
	{RuleMissed, "After 7 to 13 days with no logged set of the exercise, the load and the reps stay the same, and each working set stops at 3 reps in reserve, for that one session.", []string{"D-66", "D-294", "EV-46"}},
	{RuleMissedRestored, "After the session that followed 7 to 13 days with no logged set, each working set gets back the reps in reserve of the target before that session.", []string{"D-294"}},
	{RuleDeload, "A decline is a session with fewer total reps of working sets than the session before it, at the same logged load or a heavier load. A decline in 2 sessions in a row on 2 or more exercises starts a deload of 7 days. In it, each exercise has 0.6 times its sets, rounded to the nearest whole set and at least 1, at the same load and 3 reps in reserve. A deload session is no evidence for the next target, so after the deload the targets of before it return.", []string{"D-43", "D-289", "D-295", "EV-40", "EV-41", "EV-42"}},
	{RuleOverride, "An override of the owner changes the load and the reps of each working set alone. Each set has 6 to 20 reps and an available load of the machine, and a calibration set has the reps and the load of the first working set.", []string{"D-23", "D-69", "D-293"}},
	{RuleProposalExercise, "A proposal is for the exercise of the decision.", []string{"D-23"}},
	{RuleFallback, "When Luna gives no proposal, or the policy refuses its proposal, the target is the next target of the rules alone. The decision record names the cause and each violation.", []string{"D-23", "D-176", "D-178"}},
	{RulePainWarning, "A pain report shows the fixed warning text. It names the symptom and tells the user to stop the exercise, with no diagnosis and no referral.", []string{"D-36", "D-40", "D-153"}},
}

// Rules gives each rule of the policy, in a fixed order.
func Rules() []Rule {
	out := make([]Rule, len(rules))
	for i, r := range rules {
		out[i] = Rule{r.ID, r.Text, append([]string(nil), r.Sources...)}
	}
	return out
}
