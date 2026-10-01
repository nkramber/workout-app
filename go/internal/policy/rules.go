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
	{RuleRestBounds, "Each exercise rests 60 to 180 seconds. The default is 120 seconds, and 180 seconds for the leg press.", []string{"D-172", "D-59", "EV-35", "EV-36", "EV-37"}},
	{RuleLoadAvailable, "Each load is an available weight of the machine. A dumbbell load is the load of one dumbbell, and the input refuses a dumbbell set above 100 lb.", []string{"D-54", "D-149", "D-155", "D-166"}},
	{RuleLoadRounding, "Each load is a multiple of 5 lb, or the weight that D-149 selects for a multiple of 5 lb.", []string{"D-65", "D-148", "D-149"}},
	{RuleLoadCeiling, "A proposed load is not more than the load of the next target of the policy. So a proposal adds at most one 5 lb step, and adds no load when the policy holds or lowers it.", []string{"D-23", "D-147"}},
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
