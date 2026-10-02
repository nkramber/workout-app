//go:build emulator

package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

// TestFirestoreRulesRefuseEachClientCall proves the rules of D-77 over the
// Firestore emulator. The emulator loads firestore.rules through
// firebase.json. A signed-in client can not read or write a document, and
// the allowlist document of its own uid is no exception. The control call
// as the emulator owner, which skips the rules as the Admin SDK does,
// proves that the refusals come from the rules.
func TestFirestoreRulesRefuseEachClientCall(t *testing.T) {
	authHost := requireEmulators(t)
	uid, token := signUp(t, authHost, "rules@example.com")
	docs := fmt.Sprintf("http://%s/v1/projects/%s/databases/(default)/documents",
		os.Getenv("FIRESTORE_EMULATOR_HOST"), emulatorProject)

	call := func(method, url, bearer string, body []byte) (int, string) {
		t.Helper()
		req, err := http.NewRequest(method, url, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = res.Body.Close() }()
		raw, _ := io.ReadAll(res.Body)
		return res.StatusCode, string(raw)
	}

	own := docs + "/allowlist/" + uid
	other := docs + "/settings/" + uid
	inv := docs + "/users/" + uid + "/inventory/active"
	prof := docs + "/users/" + uid + "/profile/active"
	empty := []byte(`{"fields":{}}`)

	if code, raw := call(http.MethodPatch, own, "owner", empty); code != http.StatusOK {
		t.Fatalf("control write as owner = %d %s", code, raw)
	}
	cases := []struct {
		name, method, url, bearer string
		body                      []byte
	}{
		{"read of the own allowlist entry", http.MethodGet, own, token, nil},
		{"write of the own allowlist entry", http.MethodPatch, own, token, empty},
		{"write of another document", http.MethodPatch, other, token, empty},
		{"read of the own inventory", http.MethodGet, inv, token, nil},
		{"write of the own inventory", http.MethodPatch, inv, token, empty},
		{"read of the own profile", http.MethodGet, prof, token, nil},
		{"write of the own profile", http.MethodPatch, prof, token, empty},
		{"read with no sign-in", http.MethodGet, own, "", nil},
		{"write with no sign-in", http.MethodPatch, other, "", empty},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if code, raw := call(c.method, c.url, c.bearer, c.body); code != http.StatusForbidden {
				t.Fatalf("%s %s = %d %s, want 403", c.method, c.url, code, raw)
			}
		})
	}
	if code, raw := call(http.MethodGet, own, "owner", nil); code != http.StatusOK {
		t.Fatalf("control read as owner = %d %s", code, raw)
	}
}
