package ai

import (
	"strings"
	"testing"
)

func TestFilterBlocks(t *testing.T) {
	for _, tc := range []struct {
		text string
		rule FilterRule
	}{
		{"", FilterEmpty},
		{" \n\t ", FilterEmpty},
		{strings.Repeat("a", ReasonMax+1), FilterLength},
		{"This press can treat your knee.", FilterMedical},
		{"A TREATMENT plan.", FilterMedical},
		{"Rehab for the shoulder.", FilterMedical},
		{"This will heal the injury.", FilterMedical},
		{"Training cures back pain.", FilterMedical},
		{"I diagnose a strain.", FilterMedical},
		{"Your doctor can help.", FilterMedical},
		{"Physical therapy work.", FilterMedical},
		{"Your medication can change this.", FilterMedical},
		{"This is a medical plan.", FilterMedical},
		{"I prescribe 3 sets.", FilterMedical},
		{"A symptom of fatigue.", FilterMedical},
		{"Arthritis needs care.", FilterMedical},
		{"Keep a low calorie diet.", FilterDiet},
		{"Good for weight   loss.", FilterDiet},
		{"Lose weight fast.", FilterDiet},
		{"Take a supplement.", FilterDiet},
		{"Call 911 now.", FilterEmergency},
		{"Chest pain is an emergency.", FilterEmergency},
		{"Stop for chest pain.", FilterEmergency},
	} {
		rule, bad := Filter(tc.text, ReasonMax)
		if !bad || rule != tc.rule {
			t.Errorf("Filter(%q) = %q %v: want %q", tc.text, rule, bad, tc.rule)
		}
	}
}

// TestFilterPasses: plain fitness text passes, and a word that holds a
// blocked word inside it passes.
func TestFilterPasses(t *testing.T) {
	for _, text := range []string{
		"You did 3 sets of 12 reps at 100 lb with 3 reps in reserve, so the load goes up 5 lb.",
		"You reported pain on the last set, so the target repeats.",
		"Two weeks with no session, so the load goes down.",
		"Use the treadmill at an easy pace.",
		"A secure grip on the handles.",
		"Each repetition is slow and smooth.",
		strings.Repeat("a", ReasonMax),
	} {
		if rule, bad := Filter(text, ReasonMax); bad {
			t.Errorf("Filter(%q) blocks it with %q", text, rule)
		}
	}
}

// TestFixedTextsPass: each template and each item of the guidance
// catalog passes the filter, so a fixed text never needs a template.
func TestFixedTextsPass(t *testing.T) {
	texts := []string{SummaryTemplate, ReasonTemplate}
	for _, g := range Guidance() {
		texts = append(texts, g.Text)
	}
	for _, text := range texts {
		if rule, bad := Filter(text, SummaryMax); bad {
			t.Errorf("Filter(%q) blocks it with %q", text, rule)
		}
	}
	if rule, bad := Filter(ReasonTemplate, ReasonMax); bad {
		t.Errorf("the reason template is blocked with %q", rule)
	}
}

// TestSpikeWordsBlocked: each blocked word of the prompt of the plan
// spike is blocked (`tools/spikes/luna_plan/prompt.py`).
func TestSpikeWordsBlocked(t *testing.T) {
	for _, w := range strings.Split("diagnose, diagnosis, treat, treatment, rehab, rehabilitation, therapy, cure, heal, prescribe, prescription, medication", ", ") {
		if _, bad := Filter("This "+w+" now.", ReasonMax); !bad {
			t.Errorf("%q passes", w)
		}
	}
}
