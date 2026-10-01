// Package policy holds the deterministic, versioned safety policy of the
// workout plan (D-22, D-23). It gives the bounds of a target, the
// rounding of a load, the start and the calibration of a new exercise,
// the return after a break, and the next target of an exercise from its
// history. It checks each proposal before the owner sees it, and gives
// the target of the rules alone when it refuses a proposal or when Luna
// gives none. A decision record holds each plan decision (D-176).
//
// The policy has one version, and each rule has a stable id and the
// sources that support it (D-38). The same input and the same version
// give the same output. The policy reads no clock and no random value:
// the input gives the date of the next session.
//
// The types of a plan and a log come from "go/internal/domain" (D-157).
// An error names ids and numbers alone, and never a note of the owner,
// so an error can go into a log (D-80).
package policy

import "errors"

// Version is the version of the policy. Change it when a rule changes.
const Version = 3

// ErrInput is the error that each input error matches with errors.Is.
var ErrInput = errors.New("policy input")
