// Package envguard refuses a start on Cloud Run with a local-only
// variable (D-129, REC-17). With FIREBASE_AUTH_EMULATOR_HOST set, the
// Firebase Admin SDK accepts an unsigned ID token, so one wrong variable
// opens the API to a forged token. The guard refuses each variable whose
// name ends in _EMULATOR_HOST, so a new emulator needs no change here.
// It also refuses each name of LocalOnly, such as the switch to the fake
// provider of Luna (D-24).
package envguard

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
)

// ErrEmulatorOnCloudRun is the refusal of Check.
var ErrEmulatorOnCloudRun = errors.New("an emulator or local-only variable is set on Cloud Run")

// LocalOnly holds the other names that only a local run can set.
var LocalOnly = []string{"LUNA_FAKE_PROVIDER"}

// OnCloudRun reports whether the process runs on Cloud Run. Cloud Run
// sets K_SERVICE in each service container.
func OnCloudRun(getenv func(string) string) bool {
	return getenv("K_SERVICE") != ""
}

// Check returns an error when the environment is Cloud Run and holds an
// emulator variable. A variable with an empty value counts too, because
// the refusal is about the deploy config and not about the value. The
// error names the variables, and never a value.
func Check(environ []string, getenv func(string) string) error {
	if !OnCloudRun(getenv) {
		return nil
	}
	var names []string
	for _, kv := range environ {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasSuffix(name, "_EMULATOR_HOST") || slices.Contains(LocalOnly, name) {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil
	}
	sort.Strings(names)
	return fmt.Errorf("%w: %s", ErrEmulatorOnCloudRun, strings.Join(names, ", "))
}
