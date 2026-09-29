// Package auth names the caller of each call. A Connect interceptor reads
// the Firebase ID token from the Authorization header, verifies it, asks
// the invite allowlist about its uid (D-75, D-131), and puts the uid in
// the context. The services read it with UserID. The package follows
// decktome:go/internal/auth/auth.go, with two changes: the allowlist key
// is the uid and not the email, and no debug user exists (D-129).
package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"connectrpc.com/connect"
	firebase "firebase.google.com/go/v4"
	fbauth "firebase.google.com/go/v4/auth"
)

// Verifier checks one ID token and returns the uid that it names.
type Verifier interface {
	Verify(ctx context.Context, idToken string) (uid string, err error)
}

// VerifierFunc adapts a function to Verifier.
type VerifierFunc func(ctx context.Context, idToken string) (string, error)

// Verify implements Verifier.
func (f VerifierFunc) Verify(ctx context.Context, idToken string) (string, error) {
	return f(ctx, idToken)
}

// Allowlist says whether a uid can use the API (D-131).
type Allowlist interface {
	Allowed(ctx context.Context, uid string) (bool, error)
}

// Firebase verifies tokens with the Firebase Admin SDK. With
// FIREBASE_AUTH_EMULATOR_HOST set, the SDK trusts the unsigned tokens of
// the emulator, so the envguard package refuses that variable on Cloud
// Run (D-129).
type Firebase struct {
	client *fbauth.Client
}

// NewFirebase builds the verifier for one project.
func NewFirebase(ctx context.Context, projectID string) (*Firebase, error) {
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID})
	if err != nil {
		return nil, fmt.Errorf("auth: firebase app: %w", err)
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth: firebase auth client: %w", err)
	}
	return &Firebase{client: client}, nil
}

// Verify implements Verifier.
func (f *Firebase) Verify(ctx context.Context, idToken string) (string, error) {
	tok, err := f.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", err
	}
	if tok.UID == "" {
		return "", errors.New("token carries no uid")
	}
	return tok.UID, nil
}

type ctxKey struct{}

// UserID reads the uid that the interceptor stored, or "" when none.
func UserID(ctx context.Context) string {
	uid, _ := ctx.Value(ctxKey{}).(string)
	return uid
}

// WithUserID returns ctx with uid set. Tests and the interceptor use it.
func WithUserID(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, ctxKey{}, uid)
}

// The sentences that a refused caller reads. None names the reason of
// the verifier, because that reason can name the project or a key id.
var (
	errNoToken    = errors.New("a bearer token is required")
	errBadToken   = errors.New("the bearer token was refused")
	errNotInvited = errors.New("this app is open to invited users alone")
	errNoList     = errors.New("the invite list could not be read")
)

type interceptor struct {
	verify Verifier
	allow  Allowlist
}

// Interceptor returns the Connect interceptor. Each unary and streaming
// handler needs a verified token of a uid on the allowlist. The client
// side stays as it is.
func Interceptor(v Verifier, a Allowlist) connect.Interceptor {
	return &interceptor{verify: v, allow: a}
}

// resolve reads the Authorization header and returns the context that
// carries the uid, or the Connect error of the refusal.
func (i *interceptor) resolve(ctx context.Context, authorization string) (context.Context, error) {
	token := bearer(authorization)
	if token == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errNoToken)
	}
	uid, err := i.verify.Verify(ctx, token)
	if err != nil || uid == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errBadToken)
	}
	ok, err := i.allow.Allowed(ctx, uid)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, errNoList)
	}
	if !ok {
		return nil, connect.NewError(connect.CodePermissionDenied, errNotInvited)
	}
	return WithUserID(ctx, uid), nil
}

// bearer returns the token of "Bearer <token>", or "" when the header is
// absent, names another scheme, or holds no token. The scheme is case
// insensitive. Only a space separates the scheme and the token.
func bearer(authorization string) string {
	const scheme = "bearer "
	v := strings.TrimSpace(authorization)
	if len(v) < len(scheme) || !strings.EqualFold(v[:len(scheme)], scheme) {
		return ""
	}
	token := strings.TrimSpace(v[len(scheme):])
	if strings.IndexFunc(token, unicode.IsSpace) >= 0 {
		return ""
	}
	return token
}

func (i *interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			return next(ctx, req)
		}
		ctx, err := i.resolve(ctx, req.Header().Get("Authorization"))
		if err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

func (i *interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, err := i.resolve(ctx, conn.RequestHeader().Get("Authorization"))
		if err != nil {
			return err
		}
		return next(ctx, conn)
	}
}
