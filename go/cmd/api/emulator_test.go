//go:build emulator

// The acceptance story of work area 2.1 over the Firebase emulators.
// `make emulator-test` starts the Auth and Firestore emulators of
// firebase.json and runs this file with the tag emulator. The test makes
// no call to a real project: the project id starts with demo-, and the
// emulator variables send each call of the SDKs to 127.0.0.1.
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/allowlist"
)

const emulatorProject = "demo-workout-app"

func requireEmulators(t *testing.T) (authHost string) {
	t.Helper()
	authHost = os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")
	if authHost == "" || os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Fatal("run this test with `make emulator-test`, which sets FIREBASE_AUTH_EMULATOR_HOST and FIRESTORE_EMULATOR_HOST")
	}
	if os.Getenv("GOOGLE_CLOUD_PROJECT") != emulatorProject {
		t.Fatalf("GOOGLE_CLOUD_PROJECT must be %s, so no call reaches a real project", emulatorProject)
	}
	return authHost
}

// signUp makes an account on the Auth emulator and returns its uid and
// ID token. The emulator accepts any API key.
func signUp(t *testing.T, authHost, email string) (uid, token string) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"email": email, "password": "emulator-only-1", "returnSecureToken": true})
	url := fmt.Sprintf("http://%s/identitytoolkit.googleapis.com/v1/accounts:signUp?key=emulator", authHost)
	res, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("signUp = %d %s", res.StatusCode, raw)
	}
	var out struct{ LocalID, IDToken string }
	if err := json.Unmarshal(raw, &out); err != nil || out.LocalID == "" || out.IDToken == "" {
		t.Fatalf("signUp answer %s: %v", raw, err)
	}
	return out.LocalID, out.IDToken
}

// forgedToken is an unsigned token of another project for the uid. The
// emulator mode of the SDK skips the signature, so the test proves that
// the audience check still refuses it.
func forgedToken(uid string) string {
	now := time.Now().Unix()
	enc := func(v any) string {
		raw, _ := json.Marshal(v)
		return base64.RawURLEncoding.EncodeToString(raw)
	}
	claims := map[string]any{
		"iss": "https://securetoken.google.com/demo-other", "aud": "demo-other",
		"sub": uid, "user_id": uid, "iat": now, "exp": now + 3600, "auth_time": now,
	}
	return enc(map[string]any{"alg": "none", "typ": "JWT"}) + "." + enc(claims) + "."
}

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	return fmt.Sprint(l.Addr().(*net.TCPAddr).Port)
}

// startAPI runs the real run function, with the real Firebase verifier
// and the real Firestore allowlist, and waits for the version route. The
// planner calls go to the fake provider (D-24).
func startAPI(t *testing.T) string {
	t.Helper()
	return startAPIWith(t, &ai.Fake{})
}

// startAPIWith runs the API with a provider of the test and the caps of
// D-188.
func startAPIWith(t *testing.T, provider ai.Provider) string {
	t.Helper()
	port := freePort(t)
	env := map[string]string{
		"PORT": port, "ALLOWED_ORIGIN": "http://127.0.0.1:5173",
		ai.EnvUserCap: "1", ai.EnvProjectCap: "2", EnvOpenAIKey: "",
	}
	getenv := func(k string) string {
		if v, ok := env[k]; ok {
			return v
		}
		return os.Getenv(k)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	fake := func(string) (ai.Provider, error) { return provider, nil }
	go func() { done <- run(ctx, slog.New(slog.NewTextHandler(io.Discard, nil)), os.Environ(), getenv, fake) }()
	t.Cleanup(func() {
		cancel()
		if err := <-done; err != nil {
			t.Errorf("run = %v, want a clean stop", err)
		}
	})
	base := "http://127.0.0.1:" + port
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(50 * time.Millisecond) {
		res, err := http.Get(base + VersionPath)
		if err == nil {
			_ = res.Body.Close()
			return base
		}
	}
	t.Fatal("the API did not answer in 10 seconds")
	return ""
}

func TestAcceptanceStory(t *testing.T) {
	authHost := requireEmulators(t)
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	allowedUID, allowedToken := signUp(t, authHost, fmt.Sprintf("allowed-%d@example.test", suffix))
	otherUID, otherToken := signUp(t, authHost, fmt.Sprintf("other-%d@example.test", suffix))

	fs, err := firestore.NewClient(ctx, emulatorProject)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fs.Close() }()
	if _, err := fs.Collection(allowlist.Collection).Doc(allowedUID).Set(ctx, map[string]any{}); err != nil {
		t.Fatal(err)
	}

	base := startAPI(t)
	client := workoutappv1connect.NewUserServiceClient(http.DefaultClient, base)
	cases := []struct {
		name  string
		token string
		code  connect.Code
		uid   string
	}{
		{"no token", "", connect.CodeUnauthenticated, ""},
		{"bad token", "not-a-token", connect.CodeUnauthenticated, ""},
		{"token of another project", forgedToken(allowedUID), connect.CodeUnauthenticated, ""},
		{"uid outside the allowlist", otherToken, connect.CodePermissionDenied, ""},
		{"allowed uid", allowedToken, 0, allowedUID},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := connect.NewRequest(&workoutappv1.GetMeRequest{})
			if c.token != "" {
				req.Header().Set("Authorization", "Bearer "+c.token)
			}
			res, err := client.GetMe(ctx, req)
			if c.code != 0 {
				if connect.CodeOf(err) != c.code {
					t.Fatalf("GetMe error = %v, want code %v", err, c.code)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetMe error = %v", err)
			}
			if res.Msg.GetUid() != c.uid {
				t.Fatalf("GetMe uid = %q, want the allowed uid", res.Msg.GetUid())
			}
		})
	}
	if otherUID == allowedUID {
		t.Fatal("the emulator gave two accounts one uid")
	}

	// The version route answers with no token.
	res, err := http.Get(base + VersionPath)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(raw), `"commit"`) {
		t.Fatalf("version = %d %s", res.StatusCode, raw)
	}
}
