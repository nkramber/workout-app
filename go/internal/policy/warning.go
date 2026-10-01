package policy

import "github.com/nkramber/workout-app/go/internal/domain"

// PainWarning is the fixed text of the warning of a pain report (D-40,
// D-153). It names the symptom plainly and tells the user to stop the
// exercise. It holds no diagnosis, no referral, no first aid step, and
// no emergency number (D-36). The user can continue after a
// confirmation (D-40).
const PainWarning = "You reported pain. Stop this exercise."

// Warning gives the warning of a logged set: PainWarning for a pain
// rating of 1 or more (D-169), and false for a set with no report or a
// rating of 0.
func Warning(s domain.SetLog) (string, bool) {
	if s.Pain != nil && *s.Pain >= 1 {
		return PainWarning, true
	}
	return "", false
}
