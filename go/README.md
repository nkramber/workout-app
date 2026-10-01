# Workout App - the Go API

This folder holds the API of work area 2.1. The API serves the contract of `proto/` on Connect-RPC. The layout follows Decktome (D-74).

| Path | Content |
|---|---|
| `go/cmd/api` | The entry point, the routes, and the tests of the acceptance story |
| `go/internal/auth` | The Firebase ID token check, the allowlist check, and CORS |
| `go/internal/allowlist` | The invite allowlist of uids in Firestore (D-131) |
| `go/internal/envguard` | The start guard against an emulator variable on Cloud Run (D-129) |
| `go/internal/usersvc` | The `GetMe` call |
| `go/internal/domain` | The types of the workout domain, the catalog of D-155, and the check of each type (D-157) |
| `go/internal/policy` | The versioned safety policy: the bounds of a target, the rounding of a load, the start and the calibration of a new exercise, the return after a break, the next target, the check of a proposal, the rules fallback, and the decision record (D-23, D-38, D-176) |
| `go/gen` | The generated code. `make proto` writes it, and Git keeps it. |

## Environment

| Variable | Use |
|---|---|
| `PORT` | The listen port. The default is 8080. Cloud Run sets it. |
| `GOOGLE_CLOUD_PROJECT` | The Firebase project of the tokens and of Firestore. The API does not start without it. |
| `ALLOWED_ORIGIN` | The one origin of the web app (D-82). Empty means no cross-origin call. |
| `FIREBASE_AUTH_EMULATOR_HOST`, `FIRESTORE_EMULATOR_HOST` | The local emulators. On Cloud Run, the API refuses each variable with a name that ends in `_EMULATOR_HOST` (D-129). |

The build writes the commit into the binary with `-ldflags "-X main.commit=<sha>"`. The route `GET /version` gives it as `{"commit": "<sha>"}`, with no sign-in. The route does not use `/healthz`, because that path does not answer on a `run.app` URL (`decktome:cloudbuild/api.yaml`).

## The allowlist

Each call of the contract needs a verified Firebase ID token. The uid of the token must have a document in the Firestore collection `allowlist`, with the uid as the document id. The document can be empty. The API keeps each answer for one minute, so a change of the list takes effect in one minute or less.

The refusals:

- No token, a bad token, or a token of another project: `unauthenticated`.
- A uid with no document: `permission_denied`.
- A failed read of the list: `unavailable`.

The live project holds the entry of the owner uid. `docs/setup-gcp.md` gives the step. No uid goes into the repository.

## Emulators

`firebase.json` at the root sets the ports. They do not collide with the ports of Decktome or of the probe.

| Emulator | Port | Decktome | Probe |
|---|---|---|---|
| Auth | 9299 | 9199 | 9099 |
| Firestore | 8381 | 8281 | none |
| Firestore websocket | 9350 | 9150 | none |
| Hub | 4690 | 4490 | 4400 |
| Logging | 4790 | 4590 | 4500 |

The browser tests of `web/` start the API on port 8480, with `ALLOWED_ORIGIN` set to the origin of the test build (`web/README.md`).

The UI of the emulators stays off. `make emulator-test` starts the emulators with the pinned `firebase-tools` of `emulators/package.json`, and runs the Go tests with the build tag `emulator`. The project id is `demo-workout-app`, so no call reaches a real project (D-115). The Firestore emulator needs Java 21.
