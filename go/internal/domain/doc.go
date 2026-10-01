// Package domain holds the types of the workout domain and the catalog
// (D-157). The types are plain Go values with no proto message, no
// Firestore tag, and no clock. Each type has a Check method that reads
// the structure of a value: the ids, the units, and the bounds that a
// decision gives. The safety bounds of a prescription, such as 1 to 3
// reps in reserve, are not here. The policy of "go/internal/policy"
// holds them (D-23).
//
// A check error names the field and the number, and never a note or
// other text of the owner, so an error can go into a log (D-80). Each
// check error matches ErrInvalid.
package domain

import (
	"errors"
	"fmt"
)

// ErrInvalid is the error that each check error matches with errors.Is.
var ErrInvalid = errors.New("invalid")

type checkError string

func (e checkError) Error() string { return string(e) }
func (checkError) Is(t error) bool { return t == ErrInvalid }

func invalid(format string, args ...any) error {
	return checkError(fmt.Sprintf(format, args...))
}
