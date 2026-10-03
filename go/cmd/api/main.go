// Command api is the Workout App API on Connect-RPC (work area 2.1). The
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
//   - OPENAI_API_KEY: the key of the planner calls, from the secret
//     openai-api-key. The API refuses to start without it.
//   - LUNA_CAP_USER_USD, LUNA_CAP_PROJECT_USD: the monthly AI caps of
//     D-188. The API refuses to start without them.
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

	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/allowlist"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/capstore"
	"github.com/nkramber/workout-app/go/internal/envguard"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/inventorysvc"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/plansvc"
	"github.com/nkramber/workout-app/go/internal/profile"
	"github.com/nkramber/workout-app/go/internal/profilesvc"
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

// EnvOpenAIKey names the configuration value of the OpenAI key.
const EnvOpenAIKey = "OPENAI_API_KEY"

// ProviderFunc gives the provider of the planner calls from the key.
// The API gives OpenAI, and a test gives the fake provider (D-24).
type ProviderFunc func(key string) (ai.Provider, error)

func openAI(key string) (ai.Provider, error) { return ai.NewOpenAI(key, "", nil) }

const (
	readHeaderTimeout = 5 * time.Second
	idleTimeout       = 120 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, logger, os.Environ(), os.Getenv, openAI)
	stop()
	if err != nil {
		logger.Error("api failed", "err", err)
		os.Exit(1)
	}
	logger.Info("api stopped")
}

// run checks the environment, starts the server, and blocks until ctx
// ends or the server fails. The guard runs first, so a refused
// environment makes no client and no network call. A missing key or cap
// stops the start, so the API never calls Luna with no cap (D-25).
func run(ctx context.Context, logger *slog.Logger, environ []string, getenv func(string) string, newProvider ProviderFunc) error {
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
	caps, err := ai.CapsFromEnv(getenv)
	if err != nil {
		return err
	}
	provider, err := newProvider(getenv(EnvOpenAIKey))
	if err != nil {
		return err
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

	inventories, profiles := inventory.FromFirestore(fs), profile.FromFirestore(fs)
	maker := &plan.Maker{
		AI: &ai.Client{Provider: provider, Cap: capstore.New(fs, caps), Record: func(r ai.CostRecord) {
			// A cost record holds ids and numbers alone (D-80).
			logger.Info("ai call", "uid", r.User, "role", string(r.Role), "status", string(r.Status),
				"reserved_nano_usd", int64(r.Reserved), "cost_nano_usd", int64(r.Cost), "known", r.Known, "unsettled", r.Unsettled)
		}},
		Profiles: profiles, Inventory: inventories, Plans: plan.FromFirestore(fs), Errors: plan.ErrorsFromFirestore(fs), Log: logger,
	}
	// The server sets no write timeout, because a plan request streams
	// for up to 4 calls of Luna (D-231). The request timeout of the
	// service ends a request that runs too long.
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           newHandler(verifier, allowlist.FromFirestore(fs), inventories, profiles, maker, origin, commit),
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

// newHandler builds the routes. Each call of UserService,
// InventoryService, ProfileService, and PlanService needs a token of a
// uid on the allowlist. The version route needs no sign-in, and it gives
// the commit alone.
func newHandler(v auth.Verifier, a auth.Allowlist, store inventory.Store, profiles profile.Store, maker *plan.Maker, origin, buildCommit string) http.Handler {
	mux := http.NewServeMux()
	signedIn := connect.WithInterceptors(auth.Interceptor(v, a))
	mux.Handle(workoutappv1connect.NewUserServiceHandler(usersvc.Server{}, signedIn))
	mux.Handle(workoutappv1connect.NewInventoryServiceHandler(inventorysvc.New(store), signedIn))
	mux.Handle(workoutappv1connect.NewProfileServiceHandler(profilesvc.New(profiles), signedIn))
	mux.Handle(workoutappv1connect.NewPlanServiceHandler(plansvc.New(maker), signedIn))
	body, _ := json.Marshal(map[string]string{"commit": buildCommit})
	mux.HandleFunc("GET "+VersionPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(body)
	})
	return auth.CORS(origin, mux)
}
