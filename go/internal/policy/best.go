package policy

import "github.com/nkramber/workout-app/go/internal/domain"

// LastBest gives the best working set of the last session of the
// history with a logged working set, or nil (D-340). The best set has
// the highest estimated one-rep maximum of Epley: weight x (1 + reps /
// 30). Of two sets with the same estimate, the heavier set wins, and
// then the earlier set. The set logger shows it.
func LastBest(history []Outcome) *domain.BestSet {
	for i := len(history) - 1; i >= 0; i-- {
		var best *domain.BestSet
		for _, s := range history[i].working() {
			if best == nil || better(s, *best) {
				best = &domain.BestSet{Reps: s.Reps, Weight: s.Weight, RIR: s.RIR}
			}
		}
		if best != nil {
			return best
		}
	}
	return nil
}

// better tells whether s beats b. The estimate of Epley times 30 is
// weight x (30 + reps), so the compare needs no fraction.
func better(s domain.SetLog, b domain.BestSet) bool {
	es, eb := int64(s.Weight)*int64(30+s.Reps), int64(b.Weight)*int64(30+b.Reps)
	return es > eb || (es == eb && s.Weight > b.Weight)
}
