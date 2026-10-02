# Workout App - the web client

This folder holds the installable web shell of work area 2.2 and the inventory screens of work area 4.1. The stack is the Decktome React stack (D-84): React, Vite, `vite-plugin-pwa`, TanStack Query with Connect Query, and Tailwind. The folder is one npm package (D-135). The app has the phone layout alone (D-20).

| Path | Content |
|---|---|
| `web/src/app.tsx` | The sign-in page for a signed-out owner, and the home screen and the inventory screens for a signed-in owner |
| `web/src/shell.tsx` | The shell that fills the whole screen (D-120) |
| `web/src/pages` | The sign-in page and the home screen |
| `web/src/pages/inventory` | The inventory screens: the list, the catalog list and the text entry, the weights, and the review screen |
| `web/src/lib/inventory.ts` | The catalog order, the A to Z order of the inventory list, the search of the catalog names, and the checks of the weights, the estimates, and the notes |
| `web/src/lib/inventory-api.ts`, `web/src/lib/errors.ts` | The calls that change the inventory, and the error text of a failed call |
| `web/src/lib/firebase.ts` | Firebase Authentication with email and password (D-75) |
| `web/src/lib/api.ts` | The Connect transport, with the ID token of the owner on each call |
| `web/src/lib/db.ts` | The offline store and the outbox (D-62, D-77, D-132) |
| `web/src/lib/pwa.ts`, `web/src/lib/update-check.ts` | The service worker and its update strategy (D-133) |
| `web/src/lib/storage.ts` | The persistent storage request (D-134) |
| `web/src/gen` | The generated code of the contract. `make proto` writes it, and Git keeps it. |
| `web/e2e` | The browser tests of the acceptance stories of work areas 2.2 and 4.1 |

## The screen

The viewport meta holds `maximum-scale=1, user-scalable=no`, so a pinch does not zoom the page, as in Decktome (D-120). This fails the WCAG 1.4.4 rule for text resize. The meta also holds `viewport-fit=cover`. The shell uses the dynamic viewport height and pads itself with the safe areas. The main region is the one part that scrolls.

Since iOS 26, the Home Screen app blurs a band below the status bar, and the page can not turn it off. So the header adds 16 px above the title when a status bar covers the page (D-145).

The app has no form that makes an account. The owner makes the one account in the Firebase console, and self sign-up is off (D-117).

## The inventory screens

The home screen opens the equipment inventory (work area 4.1). The screens read the catalog and the inventory through `InventoryService` of the API, and never through Firestore (D-77).

- The list shows each machine with its state, draft or confirmed, A to Z by name (D-205), and each note.
- The catalog list shows the machines by kind (D-202), and each kind A to Z by name (D-205). A text searches the names of the machines and of their exercises. The owner selects a match, or keeps the text as a note (D-191).
- A machine or the cable station gets a range and a step, then single weights (D-195). The dumbbells get the dumbbell set. A cardio machine gets no weights.
- Each exercise of the machine gets one optional estimate, from the lightest to the heaviest weight (D-192, D-198).
- The review screen confirms the weights that it shows (D-201), and removes the machine.

Each load is a whole number of tenths of a pound, as in the contract. Each change is a direct call to the API. With no connection, the screen shows an error, and the server keeps no change (D-196). A server fault, the code `internal`, shows "The server failed." (D-206). The inventory has no offline copy until Phase 6.

## The offline store

Dexie on IndexedDB holds the local state. Each change and its outbox entry go into one transaction (D-132). An outbox entry holds a UUIDv7 op id, the entity and its id, and the base version. It also holds the payload, the time, the attempts, and the schema version. The one entity of Phase 2 is a setting. Phase 3 adds the entities of the domain model, and Phase 6 adds the sync call.

The first sign-in on a device asks for persistent storage (D-134). The home screen shows the result and the use of the store.

## Updates

The service worker waits after an update, and the app shows "Update ready" (D-133). The owner applies the update with the button. The app never applies an update during a workout. Phase 4 adds the workout state. The app checks for an update each hour and at each return to view.

## Environment

| Variable | Use |
|---|---|
| `VITE_API_BASE_URL` | The origin of the API at build time, such as the Cloud Run URL that `cloudbuild/web.yaml` names. Empty means the origin of the page. |
| `VITE_AUTH_EMULATOR_HOST` | The local Auth emulator. The browser tests set it. `npm run dev` uses `127.0.0.1:9299`. A build with no value uses the project `nk-workout-app-prod`. |

## Checks

`make web` runs the checks of the CI job `verify:web` (D-126). It needs Node 22, Go, and Java 21:

1. Type-check the code with `tsc`.
2. Run the unit tests of `web/src` with Vitest.
3. Build the app.
4. Run the browser tests in WebKit and in Chromium, each with phone emulation.

The browser tests start the Auth and Firestore emulators of `firebase.json`. They also start the API of `go/` on port 8480, and a build of the app on port 4273. No call reaches a real project (D-115). The tests make each account on the Auth emulator and write its allowlist document on the Firestore emulator.

Playwright can make a pinch in Chromium alone. So the pinch test runs in Chromium, and WebKit reads the viewport meta. A headless browser has no Home Screen and no status bar. So the device check of PR-11 read the bottom edge and the blur band on the iPhone. A Chromium test sets the safe area through the DevTools protocol, and it reads the space above the title.
