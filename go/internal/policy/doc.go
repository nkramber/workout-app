// Package policy holds the deterministic, versioned safety policy of the
// workout plan (D-22, D-23). It gives the bounds of a target, the
// rounding of a load, and the next target of an exercise from its
// history. It checks each proposal before the owner sees it.
//
// The policy has one version, and each rule has a stable id and the
// sources that support it (D-38). The same input and the same version
// give the same output. The policy reads no clock and no random value.
//
// The types of a plan and a log come from "go/internal/domain" (D-157).
// An error names ids and numbers alone, and never a note of the owner,
// so an error can go into a log (D-80).
package policy

import "errors"

// Version is the version of the policy. Change it when a rule changes.
const Version = 1

// ErrInput is the error that each input error matches with errors.Is.
var ErrInput = errors.New("policy input")

// ErrNoHistory tells that an exercise has no history. The start of a new
// exercise is the calibration of D-150, which this package does not hold
// yet.
var ErrNoHistory = errors.New("policy: no history")
