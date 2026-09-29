// Command api is the Gym Route API on Connect-RPC (work area 2.1). The
// routes follow decktome:go/cmd/api/main.go.
//
// Environment:
//   - PORT: the listen port, 8080 when empty. Cloud Run sets it.
//   - GOOGLE_CLOUD_PROJECT: the Firebase project of the ID tokens and of
//     Firestore. The API refuses to start without it.
//   - ALLOWED_ORIGIN: the one origin of the web app (D-82), or empty for
//     no cross-origin call.
//   - FIREBASE_AUTH_EMULATOR_HOST, FIRESTORE_EMULATOR_HOST: the local
//     emulators. The API refuses them on Cloud Run (D-129).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	"connectrpc.com/connect"

	"github.com/nkramber/workout-app/go/gen/gymroute/v1/gymroutev1connect"
	"github.com/nkramber/workout-app/go/internal/allowlist"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/envguard"
	"github.com/nkramber/workout-app/go/internal/usersvc"
)

// commit is the build commit. The build sets it with
// -ldflags "-X main.commit=<sha>", and the deploy check of work area 2.3
// reads it on VersionPath.
var commit = "unknown"

// VersionPath names the build commit. Decktome reads its commit on
// /readyz, and /healthz does not answer on a run.app URL
// (decktome:cloudbuild/api.yaml), so the path is another one.
const VersionPath = "/version"

const (
	readHeaderTimeout = 5 * time.Second
	idleTimeout       = 120 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, logger, os.Environ(), os.Getenv)
	stop()
	if err != nil {
		logger.Error("api failed", "err", err)
		os.Exit(1)
	}
	logger.Info("api stopped")
}

// run checks the environment, starts the server, and blocks until ctx
// ends or the server fails. The guard runs first, so a refused
// environment makes no client and no network call.
func run(ctx context.Context, logger *slog.Logger, environ []string, getenv func(string) string) error {
	if err := envguard.Check(environ, getenv); err != nil {
		return err
	}
	project := getenv("GOOGLE_CLOUD_PROJECT")
	if project == "" {
		return errors.New("GOOGLE_CLOUD_PROJECT is not set")
	}
	origin, err := auth.ParseOrigin(getenv("ALLOWED_ORIGIN"))
	if err != nil {
		return err
	}
	port := getenv("PORT")
	if port == "" {
		port = "8080"
	}

	verifier, err := auth.NewFirebase(ctx, project)
	if err != nil {
		return err
	}
	fs, err := firestore.NewClient(ctx, project)
	if err != nil {
		return fmt.Errorf("firestore: %w", err)
	}
	defer func() { _ = fs.Close() }()

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           newHandler(verifier, allowlist.FromFirestore(fs), origin, commit),
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
	}
	serveErr := make(chan error, 1)
	go func() {
		logger.Info("api listening", "port", port, "commit", commit)
		serveErr <- srv.ListenAndServe()
	}()
	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// newHandler builds the routes. Each call of UserService needs a token
// of a uid on the allowlist. The version route needs no sign-in, and it
// gives the commit alone.
func newHandler(v auth.Verifier, a auth.Allowlist, origin, buildCommit string) http.Handler {
	mux := http.NewServeMux()
	path, h := gymroutev1connect.NewUserServiceHandler(usersvc.Server{},
		connect.WithInterceptors(auth.Interceptor(v, a)))
	mux.Handle(path, h)
	body, _ := json.Marshal(map[string]string{"commit": buildCommit})
	mux.HandleFunc("GET "+VersionPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(body)
	})
	return auth.CORS(origin, mux)
}
