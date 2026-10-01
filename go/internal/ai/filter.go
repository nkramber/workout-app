package ai

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// FilterVersion is the version of the filter of blocked claims (D-183).
// Change it with each change of a pattern or a limit.
const FilterVersion = 1

// FilterRule is the stable id of one rule of the filter.
type FilterRule string

const (
	// FilterMedical blocks a claim of diagnosis, treatment,
	// rehabilitation, or medical care (D-36, D-93).
	FilterMedical FilterRule = "text.medical"
	// FilterDiet blocks diet and weight-loss advice (D-36, D-93).
	FilterDiet FilterRule = "text.diet"
	// FilterEmergency blocks emergency advice (D-36).
	FilterEmergency FilterRule = "text.emergency"
	// FilterEmpty blocks a text with no words.
	FilterEmpty FilterRule = "text.empty"
	// FilterLength blocks a text over its length limit (D-182).
	FilterLength FilterRule = "text.length"
)

// The length limits of the texts of Luna, in characters (D-182).
const (
	SummaryMax = 280
	ReasonMax  = 160
)

// The template texts that replace a blocked text (D-183).
const (
	SummaryTemplate = "Your next sessions. The safety policy checked each target."
	ReasonTemplate  = "The safety policy checked this target."
)

// The patterns of the blocked claims. Each one matches whole words, in
// any case. The patterns are a guard after the prompt, not a proof: a
// claim in other words can pass. So the texts stay short (D-182).
var blocked = []struct {
	rule FilterRule
	re   *regexp.Regexp
}{
	{FilterMedical, regexp.MustCompile(`(?i)\b(diagnos\w*|treat(s|ed|ing|ment|ments)?|therap\w*|rehab\w*|cur(e|es|ed|ing)|heal(s|ed|ing)?|prescri\w*|medic\w*|injur\w*|diseases?|arthritis|surgery|surgical|physio\w*|doctors?|physicians?|symptoms?)\b`)},
	{FilterDiet, regexp.MustCompile(`(?i)\b(diet\w*|calori\w*|fasting|weight loss|lose weight|fat loss|burns? fat|supplements?)\b`)},
	{FilterEmergency, regexp.MustCompile(`(?i)\b(emergenc\w*|911|ambulance|chest pain|heart attack)\b`)},
}

// Filter gives the first rule that blocks a text of Luna with a length
// limit of max characters, or false when no rule blocks it.
func Filter(text string, max int) (FilterRule, bool) {
	words := strings.Join(strings.Fields(text), " ")
	if words == "" {
		return FilterEmpty, true
	}
	if utf8.RuneCountInString(words) > max {
		return FilterLength, true
	}
	for _, b := range blocked {
		if b.re.MatchString(words) {
			return b.rule, true
		}
	}
	return "", false
}
