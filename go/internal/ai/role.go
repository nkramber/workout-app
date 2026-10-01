package ai

import "time"

// RoleName names one role of the layer.
type RoleName string

const (
	// RolePlanner plans the next sessions of the owner (D-22).
	RolePlanner RoleName = "planner"
	// RoleReviser gives the targets of the next session after the owner
	// logs a session (D-22, D-64).
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
	Timeout     time.Duration
	Prices      Prices
}

// The configuration of `gpt-6-luna` (D-24). The prices come from
// "docs/research/platform-cloud-and-ai.md", read 2026-09-28: 0.10 USD,
// 0.01 USD, and 0.50 USD for each million input, cached input, and
// output tokens. The output limit is the limit of the plan spike, and
// one plan of the spike used 2,579 output tokens at most. One call of
// the spike took 29.7 seconds at most, so the time limit is 3 times
// that, rounded up (assumption).
const (
	lunaModel       = "gpt-6-luna"
	lunaEffort      = "medium"
	lunaMaxOutput   = 32000
	lunaTimeout     = 90 * time.Second
	plannerSessions = 7
)

var lunaPrices = Prices{Input: 100, CachedInput: 10, Output: 500}

// Planner gives the role that plans the next sessions, at most one week
// of them.
func Planner() Role {
	return Role{RolePlanner, lunaModel, lunaEffort, lunaMaxOutput, plannerSessions, lunaTimeout, lunaPrices}
}

// Reviser gives the role that gives the targets of the next session.
func Reviser() Role {
	return Role{RoleReviser, lunaModel, lunaEffort, lunaMaxOutput, 1, lunaTimeout, lunaPrices}
}

// Roles gives each role, in a fixed order.
func Roles() []Role { return []Role{Planner(), Reviser()} }
