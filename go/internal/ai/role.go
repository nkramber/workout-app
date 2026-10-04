package ai

import (
	"slices"
	"time"
)

// RoleName names one role of the layer.
type RoleName string

const (
	// RolePlanner plans the next sessions of the owner (D-22).
	RolePlanner RoleName = "planner"
	// RoleReviser writes the reason of each next target after the owner
	// logs a session. The rules of the policy give each target (D-64,
	// D-288).
	RoleReviser RoleName = "reviser"
)

// Role is the model configuration of one role. This file holds the one
// model id of the module (D-24). A test fails when another Go file of
// the module names a model.
type Role struct {
	Name            RoleName
	Model           string
	Effort          string
	MaxOutputTokens int
	// MaxSessions is the largest number of sessions in one call.
	MaxSessions int
	// MaxRequestBytes is the largest request. It keeps each call in
	// the short context of the model, so the prices of Prices apply.
	MaxRequestBytes int
	Timeout         time.Duration
	Prices          Prices
}

// The configuration of `gpt-6-luna` (D-24). The prices come from
// "docs/research/platform-cloud-and-ai.md", read 2026-10-01: 0.10 USD,
// 0.01 USD, 0.125 USD, and 0.50 USD for each million input, cached
// input, cache-write, and output tokens. These are the prices of the
// short context, 272,000 input tokens or fewer. A longer request bills
// at higher prices, so the layer refuses a request of more than
// 272,000 bytes, because one token holds one byte or more. The output
// limit is the limit of the plan spike, and one plan of the spike used
// 2,579 output tokens at most. One call of
// the spike took 29.7 seconds at most, so the time limit is 3 times
// that, rounded up (assumption). The effort is xhigh (D-253). At xhigh,
// one call of "docs/research/luna-effort-check.md" used 7,351 output
// tokens and took 63.6 seconds at most, so both limits stay.
const (
	lunaModel       = "gpt-6-luna"
	lunaEffort      = "xhigh"
	lunaMaxOutput   = 32000
	lunaTimeout     = 90 * time.Second
	lunaMaxRequest  = 272_000
	plannerSessions = 7
)

// ReviserTimeout is the time limit of a reviser call. The call runs in
// the sync of a finished workout, and after the limit the reason of the
// rules shows (D-292). A reason call has a smaller output than a plan,
// so the limit is half the limit of a plan (assumption). The live
// revision of work area 7.2 measures it.
const ReviserTimeout = 45 * time.Second

var lunaPrices = Prices{Input: 100, CachedInput: 10, CacheWrite: 125, Output: 500}

// Efforts gives each reasoning effort of `gpt-6-luna`, from the model
// page, read 2026-09-28 ("docs/research/platform-cloud-and-ai.md").
var Efforts = []string{"none", "low", "medium", "high", "xhigh", "max"}

// ValidEffort tells that e is one of Efforts.
func ValidEffort(e string) bool { return slices.Contains(Efforts, e) }

// Planner gives the role that plans the next sessions, at most one week
// of them.
func Planner() Role {
	return Role{RolePlanner, lunaModel, lunaEffort, lunaMaxOutput, plannerSessions, lunaMaxRequest, lunaTimeout, lunaPrices}
}

// Reviser gives the role that writes the reason of each target of the
// next session.
func Reviser() Role {
	return Role{RoleReviser, lunaModel, lunaEffort, lunaMaxOutput, 1, lunaMaxRequest, ReviserTimeout, lunaPrices}
}

// Roles gives each role, in a fixed order.
func Roles() []Role { return []Role{Planner(), Reviser()} }
