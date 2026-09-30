package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/envguard"
)

const webOrigin = "https://nk-workout-app-prod.web.app"

// fakeVerifier accepts the tokens of its map, and refuses each other.
type fakeVerifier map[string]string

func (f fakeVerifier) Verify(_ context.Context, token string) (string, error) {
	if uid, ok := f[token]; ok {
		return uid, nil
	}
	return "", errors.New("unknown token")
}

type fakeList struct {
	uids map[string]bool
	err  error
}

func (f fakeList) Allowed(_ context.Context, uid string) (bool, error) {
	return f.uids[uid], f.err
}

func server(t *testing.T, list auth.Allowlist) *httptest.Server {
	t.Helper()
	v := fakeVerifier{"token-a": "uid-a", "token-b": "uid-b", "token-empty": ""}
	srv := httptest.NewServer(newHandler(v, list, webOrigin, "abc123"))
	t.Cleanup(srv.Close)
	return srv
}

func getMe(t *testing.T, srv *httptest.Server, authorization string) (*connect.Response[workoutappv1.GetMeResponse], error) {
	t.Helper()
	client := workoutappv1connect.NewUserServiceClient(srv.Client(), srv.URL)
	req := connect.NewRequest(&workoutappv1.GetMeRequest{})
	if authorization != "" {
		req.Header().Set("Authorization", authorization)
	}
	return client.GetMe(context.Background(), req)
}

// TestGetMe is the acceptance story over fakes. The emulator test of
// emulator_test.go runs the same story over the Firebase emulators.
func TestGetMe(t *testing.T) {
	srv := server(t, fakeList{uids: map[string]bool{"uid-a": true}})
	cases := []struct {
		name          string
		authorization string
		code          connect.Code
		uid           string
	}{
		{"no token", "", connect.CodeUnauthenticated, ""},
		{"empty bearer", "Bearer ", connect.CodeUnauthenticated, ""},
		{"other scheme", "Basic token-a", connect.CodeUnauthenticated, ""},
		{"bad token", "Bearer forged", connect.CodeUnauthenticated, ""},
		{"token with no uid", "Bearer token-empty", connect.CodeUnauthenticated, ""},
		{"uid off the list", "Bearer token-b", connect.CodePermissionDenied, ""},
		{"allowed uid", "Bearer token-a", 0, "uid-a"},
		{"allowed uid, lower-case scheme", "bearer token-a", 0, "uid-a"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := getMe(t, srv, c.authorization)
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
				t.Fatalf("GetMe uid = %q, want %q", res.Msg.GetUid(), c.uid)
			}
		})
	}
}

// TestGetMeListUnreadable refuses the call when the list can not be
// read. A read error never lets a caller in.
func TestGetMeListUnreadable(t *testing.T) {
	srv := server(t, fakeList{uids: map[string]bool{"uid-a": true}, err: errors.New("firestore down")})
	_, err := getMe(t, srv, "Bearer token-a")
	if connect.CodeOf(err) != connect.CodeUnavailable {
		t.Fatalf("GetMe error = %v, want Unavailable", err)
	}
	if strings.Contains(err.Error(), "firestore down") {
		t.Fatalf("GetMe error %q names the internal reason", err)
	}
}

func TestVersion(t *testing.T) {
	srv := server(t, fakeList{})
	res, err := srv.Client().Get(srv.URL + VersionPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK || body["commit"] != "abc123" || len(body) != 1 {
		t.Fatalf("version = %d %v, want 200 and the commit alone", res.StatusCode, body)
	}
	if got := res.Header.Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func TestCORS(t *testing.T) {
	srv := server(t, fakeList{uids: map[string]bool{"uid-a": true}})
	procedure := srv.URL + workoutappv1connect.UserServiceGetMeProcedure
	do := func(method, origin string) *http.Response {
		t.Helper()
		req, err := http.NewRequest(method, procedure, strings.NewReader("{}"))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Origin", origin)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer token-a")
		if method == http.MethodOptions {
			req.Header.Set("Access-Control-Request-Method", "POST")
			req.Header.Set("Access-Control-Request-Headers", "authorization, content-type")
		}
		res, err := srv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
		return res
	}

	pre := do(http.MethodOptions, webOrigin)
	if pre.StatusCode != http.StatusNoContent || pre.Header.Get("Access-Control-Allow-Origin") != webOrigin {
		t.Fatalf("preflight = %d %q, want 204 and the origin", pre.StatusCode, pre.Header.Get("Access-Control-Allow-Origin"))
	}
	if !strings.Contains(pre.Header.Get("Access-Control-Allow-Headers"), "Authorization") {
		t.Fatalf("preflight allows %q, want Authorization", pre.Header.Get("Access-Control-Allow-Headers"))
	}
	post := do(http.MethodPost, webOrigin)
	if post.StatusCode != http.StatusOK || post.Header.Get("Access-Control-Allow-Origin") != webOrigin {
		t.Fatalf("call = %d %q, want 200 and the origin", post.StatusCode, post.Header.Get("Access-Control-Allow-Origin"))
	}
	for _, method := range []string{http.MethodOptions, http.MethodPost} {
		other := do(method, "https://evil.example")
		if got := other.Header.Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("%s from another origin allows %q, want no header", method, got)
		}
		if method == http.MethodOptions && other.StatusCode == http.StatusNoContent {
			t.Fatalf("preflight from another origin = 204, want a refusal")
		}
	}
}

// TestRunRefuses proves that run refuses a bad environment before it
// makes a client, so these cases need no network.
func TestRunRefuses(t *testing.T) {
	cases := []struct {
		name    string
		environ []string
		want    string
	}{
		{"emulator on cloud run", []string{"K_SERVICE=api", "GOOGLE_CLOUD_PROJECT=p", "FIREBASE_AUTH_EMULATOR_HOST=127.0.0.1:9299"}, envguard.ErrEmulatorOnCloudRun.Error()},
		{"no project", []string{"PORT=0"}, "GOOGLE_CLOUD_PROJECT"},
		{"two origins", []string{"GOOGLE_CLOUD_PROJECT=p", "ALLOWED_ORIGIN=https://a.web.app,https://b.web.app"}, "ALLOWED_ORIGIN"},
		{"origin with a path", []string{"GOOGLE_CLOUD_PROJECT=p", "ALLOWED_ORIGIN=https://a.web.app/"}, "ALLOWED_ORIGIN"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			env := map[string]string{}
			for _, kv := range c.environ {
				k, v, _ := strings.Cut(kv, "=")
				env[k] = v
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			err := run(context.Background(), logger, c.environ, func(k string) string { return env[k] })
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("run = %v, want an error with %q", err, c.want)
			}
		})
	}
}
