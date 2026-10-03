package envguard

import (
	"errors"
	"strings"
	"testing"
)

func env(kv ...string) ([]string, func(string) string) {
	m := map[string]string{}
	for _, e := range kv {
		k, v, _ := strings.Cut(e, "=")
		m[k] = v
	}
	return kv, func(k string) string { return m[k] }
}

func TestCheck(t *testing.T) {
	cases := []struct {
		name    string
		environ []string
		refuse  string
	}{
		{"local run with emulators", []string{"FIREBASE_AUTH_EMULATOR_HOST=127.0.0.1:9299", "FIRESTORE_EMULATOR_HOST=127.0.0.1:8381"}, ""},
		{"cloud run clean", []string{"K_SERVICE=api", "PORT=8080"}, ""},
		{"cloud run with auth emulator", []string{"K_SERVICE=api", "FIREBASE_AUTH_EMULATOR_HOST=127.0.0.1:9299"}, "FIREBASE_AUTH_EMULATOR_HOST"},
		{"cloud run with firestore emulator", []string{"K_SERVICE=api", "FIRESTORE_EMULATOR_HOST=127.0.0.1:8381"}, "FIRESTORE_EMULATOR_HOST"},
		{"cloud run with a new emulator", []string{"K_SERVICE=api", "FIREBASE_STORAGE_EMULATOR_HOST=x"}, "FIREBASE_STORAGE_EMULATOR_HOST"},
		{"cloud run with an empty value", []string{"K_SERVICE=api", "FIREBASE_AUTH_EMULATOR_HOST="}, "FIREBASE_AUTH_EMULATOR_HOST"},
		{"cloud run with the fake provider", []string{"K_SERVICE=api", "LUNA_FAKE_PROVIDER=1"}, "LUNA_FAKE_PROVIDER"},
		{"local run with the fake provider", []string{"LUNA_FAKE_PROVIDER=1"}, ""},
		{"cloud run with two", []string{"K_SERVICE=api", "FIRESTORE_EMULATOR_HOST=b", "FIREBASE_AUTH_EMULATOR_HOST=a"}, "FIREBASE_AUTH_EMULATOR_HOST, FIRESTORE_EMULATOR_HOST"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			environ, getenv := env(c.environ...)
			err := Check(environ, getenv)
			if c.refuse == "" {
				if err != nil {
					t.Fatalf("Check = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, ErrEmulatorOnCloudRun) {
				t.Fatalf("Check = %v, want ErrEmulatorOnCloudRun", err)
			}
			if !strings.HasSuffix(err.Error(), ": "+c.refuse) {
				t.Fatalf("Check = %q, want the names %q", err, c.refuse)
			}
		})
	}
}

// TestCheckNamesNoValue keeps a value out of the error and so out of
// the log (D-80).
func TestCheckNamesNoValue(t *testing.T) {
	environ, getenv := env("K_SERVICE=api", "FIRESTORE_EMULATOR_HOST=secret-host:1")
	if err := Check(environ, getenv); err == nil || strings.Contains(err.Error(), "secret-host") {
		t.Fatalf("Check = %v, want a refusal with no value", err)
	}
}
