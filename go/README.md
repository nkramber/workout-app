# Workout App - the Go API

This folder holds the API of work area 2.1. The API serves the contract of `proto/` on Connect-RPC. The layout follows Decktome (D-74).

| Path | Content |
|---|---|
| `go/cmd/api` | The entry point, the routes, and the tests of the acceptance story |
| `go/cmd/lunaeval` | The Luna evaluation of Phase 3: synthetic profiles and the scenarios A to F through the layer and the policy (D-184 to D-186) |
| `go/internal/auth` | The Firebase ID token check, the allowlist check, and CORS |
| `go/internal/allowlist` | The invite allowlist of uids in Firestore (D-131) |
| `go/internal/envguard` | The start guard against an emulator variable, the fake provider, or its delay on Cloud Run (D-129, D-241) |
| `go/internal/usersvc` | The `GetMe` call |
| `go/internal/inventory` | The inventory of the owner: the machines and the notes, the checks, the draft and confirmed states, the Firestore store, and `ForPlan` (D-46, D-193, D-197) |
| `go/internal/inventorysvc` | The calls of `InventoryService` |
| `go/internal/profile` | The profile of the owner: the fields, the checks, the Firestore store, and `ForPlan` (D-41, D-208 to D-221) |
| `go/internal/profilesvc` | The calls of `ProfileService` |
| `go/internal/domain` | The types of the workout domain, the catalog of D-155, the injury areas and the muscle groups with their tables (D-218, D-219), and the check of each type (D-157) |
| `go/internal/policy` | The versioned safety policy: the bounds of a target, the rounding of a load, the start and the calibration of a new exercise, the return after a break, the next target, the check of a proposal, the rules fallback, and the decision record (D-23, D-38, D-176) |
| `go/internal/ai` | The Luna role layer: the planner and reviser roles, the plan schema, the prompt, the guidance catalog, the filter of blocked claims, the cost records, the cap hook, the OpenAI provider, and the fake provider (D-24, D-25, D-152, D-183) |
| `go/internal/plan` | The plan of the owner: the planner request, 4 calls at most with the cause of each failure, the policy check of each exercise, the exclusions, the Firestore stores, and the error records (D-226 to D-238) |
| `go/internal/plansvc` | The calls of `PlanService`, with a server stream of the progress (D-237) |
| `go/internal/capstore` | The lasting cap hook: the spend of each calendar month in UTC in Firestore, with a reservation before each call and a charge after it (D-189, D-190, D-224, D-225) |
| `go/gen` | The generated code. `make proto` writes it, and Git keeps it. |

## Environment

| Variable | Use |
|---|---|
| `PORT` | The listen port. The default is 8080. Cloud Run sets it. |
| `GOOGLE_CLOUD_PROJECT` | The Firebase project of the tokens and of Firestore. The API does not start without it. |
| `ALLOWED_ORIGIN` | The one origin of the web app (D-82). Empty means no cross-origin call. |
| `FIREBASE_AUTH_EMULATOR_HOST`, `FIRESTORE_EMULATOR_HOST` | The local emulators. On Cloud Run, the API refuses each variable with a name that ends in `_EMULATOR_HOST` (D-129). |
| `OPENAI_API_KEY` | The key of the planner calls. On Cloud Run it comes from the secret `openai-api-key`. The API does not start without it. The emulator tests give the fake provider in its place. |
| `LUNA_CAP_USER_USD`, `LUNA_CAP_PROJECT_USD` | The monthly AI caps, below. The API does not start without them. |
| `LUNA_FAKE_PROVIDER` | `1` gives the fake provider of Luna in place of OpenAI, so a local run makes no paid call (D-24). The browser tests set it. On Cloud Run, the API refuses it. |
| `LUNA_FAKE_DELAY_MS` | The wait of each call of the fake provider, from 0 to 60000 milliseconds. The browser tests set 1000, so a test can see the progress of a plan request (D-241). On Cloud Run, the API refuses it. |

`go/internal/ai` reads the caps of D-25 from `LUNA_CAP_USER_USD` and `LUNA_CAP_PROJECT_USD`, in US dollars, such as `0.25`. A value that is not set stops the start, and 0 refuses each call. D-188 gives 1 USD for the user and 2 USD for the project, for each calendar month in UTC (D-190). The API reads both variables at its start, and the cap hook of `go/internal/capstore` applies them to each planner call. A plan request over a cap gives an error at once (D-230).

`go/internal/capstore` holds the spend of each month in Firestore (D-189). The paths are `users/{uid}/aiSpend/{YYYY-MM}` for the user and `aiSpend/{YYYY-MM}` for the project (D-224). Each document holds the settled charge and the open reservations, in billionths of a US dollar:

- Before a call, one transaction adds the worst-case cost to the reservations of both documents. When a cap can not cover it, the call does not start.
- After the call, a second transaction moves the reservation to the charge. A failed call charges the worst case (D-225).
- When the API stops between the two, the reservation stays. So the spend can be too high, but never too low.
- A new month uses new documents, so its spend starts at 0.

`go/internal/ai` keeps `MemoryCap` for the tests and for `go/cmd/lunaeval`. A new process starts it at 0.

## The plan

`go/internal/plan` keeps the plan at `users/{uid}/plan/active` and the exclusions at `users/{uid}/exclusions/active` (D-226). A save writes both in one transaction, and it refuses a save when another request changed the exclusions (D-234).

- A request plans the confirmed machines alone, with no exercise of an injured area and no excluded exercise (D-49, D-208, D-229).
- An invalid output gets a retry with its cause and its output, 4 calls at most (D-230, D-231, D-235). A call over the cap ends the request at once.
- Each failed attempt adds a document to the top-level collection `aiErrors`. Its field `expire_at` drives the TTL of 90 days (D-236). The document holds the output of Luna, so no log reads it (D-80).
- The policy decides each exercise of a valid plan, and the plan stores each decision record (D-23, D-176).

## The Luna evaluation

`go/cmd/lunaeval` uses the fake provider unless you give `-live`. A fake run costs nothing:

```bash
cd go && go run ./cmd/lunaeval -cap 2 -out /tmp/report.json
```

CAUTION: a run with `-live` calls OpenAI, and each call costs money. Get the approval of the owner for the run and its cap before the run (D-25). The command reads the key from `OPENAI_API_KEY` alone. Give the key to the one process from Secret Manager (D-187):

```bash
OPENAI_API_KEY="$(gcloud secrets versions access latest --secret=openai-api-key --project=nk-workout-app-prod)" \
  go run ./cmd/lunaeval -live -cap 2 -out ../docs/research/phase-3-check/results.json
```

The flag `-effort` sets the reasoning effort of each call, so two runs can compare two efforts with the same calls. With no flag, each call uses the effort of its role. `docs/research/luna-effort-check.md` used it for medium and xhigh (D-242).

The command writes the report and a summary of ids and numbers. The report holds synthetic data alone.

The build writes the commit into the binary with `-ldflags "-X main.commit=<sha>"`. The route `GET /version` gives it as `{"commit": "<sha>"}`, with no sign-in. The route does not use `/healthz`, because that path does not answer on a `run.app` URL (`decktome:cloudbuild/api.yaml`).

## The allowlist

Each call of the contract needs a verified Firebase ID token. The uid of the token must have a document in the Firestore collection `allowlist`, with the uid as the document id. The document can be empty. The API keeps each answer for one minute, so a change of the list takes effect in one minute or less.

The refusals:

- No token, a bad token, or a token of another project: `unauthenticated`.
- A uid with no document: `permission_denied`.
- A failed read of the list: `unavailable`.

The live project holds the entry of the owner uid. `docs/setup-gcp.md` gives the step. No uid goes into the repository.

## The inventory

`InventoryService` holds the one active inventory of the caller (work area 4.2). Its calls read the catalog and the inventory, and save, confirm, and remove a machine or a note. A load is a whole number of tenths of a pound.

The store keeps one Firestore document for each user at `users/{uid}/inventory/active` (D-197). Each change reads and writes the document in one transaction. A stored machine holds its catalog id, its weights, its estimates, and its state alone. A note holds its id and its text.

The server refuses a bad value with `invalid_argument`:

- A machine outside the catalog, or weights of the wrong kind for the machine.
- A weight above 1,000 lb, or more than 200 weights (D-199). The dumbbells keep the bound of D-166.
- An estimate for an exercise of another machine, or an estimate outside the weights of the machine (D-198).
- A note of 0 or more than 200 characters, or more than 50 notes (D-199).

The other refusals:

- A confirmation of an unknown machine, or a change of an unknown note: `not_found`.
- A confirmation with weights that are not the stored weights: `failed_precondition` (D-201).
- A failed read or write of Firestore: `internal`, with a fixed text.

A removal of an unknown machine or note succeeds, so a retry is safe. An error names ids and numbers alone, and never the text of a note (D-80).

The function `inventory.ForPlan` gives the confirmed machines and their estimates to the plan input of Phase 5. A draft and a note never reach it (D-49, D-191, D-193).

## The profile

`ProfileService` holds the one profile of the caller (work area 5.1). Its calls read the fixed lists, read the profile, and save the whole profile. A caller with no profile gets no profile.

The store keeps one Firestore document for each user at `users/{uid}/profile/active` (D-213). Each save replaces the document. The load estimates of D-41 stay in the inventory (D-192).

Before its check, the server removes the spaces at each end of each text, and puts each list in the order of its fixed list. Then it refuses a bad value with `invalid_argument`:

- An experience other than "intermediate" or "advanced" (D-214), or a template other than "general_fitness" or "strength".
- No muscle group, or a group, an area, or a cardio exercise off its fixed list, or a value two times (D-216, D-217).
- An age, a height, or a weight outside its bound: 18 to 90 years, 48 to 96 in, and 80 to 500 lb (D-215).
- Training days outside 2 to 4 (D-211).
- A free text or an injury text of more than 500 characters (D-215).

A failed read or write of Firestore gives `internal`, with a fixed text. An error names the field and the bound alone, and never the value of the age, the height, the weight, or a text (D-80).

`domain.DefaultBodyTables` holds the areas and the primary groups of each exercise, and the goal templates (D-218 to D-221). A unit test finds each exercise of the catalog in both tables. Section 5.15 of `docs/research/exercise-safety.md` gives the research of each row.

The function `profile.ForPlan` gives the planner input of a profile. It holds the inputs of D-209 and the session count of D-211. It removes each exercise that loads an injured area (D-208). An exercise with no row in the area table loads each area, so a gap removes an exercise. The input type has no field for the age, the height, the weight, the areas, or the injury text.

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
