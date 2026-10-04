# Workout App - the web client

This folder holds the web client of the owner. It holds the installable shell (work area 2.2), the inventory screens (4.1), the onboarding screen (5.1), and the plan screen (5.2). It also holds the workout screen (6.1, 6.2) and the outbox sync (6.3). The stack is the Decktome React stack (D-84): React, Vite, `vite-plugin-pwa`, TanStack Query with Connect Query, and Tailwind. The folder is one npm package (D-135). The app has the phone layout alone (D-20).

| Path | Content |
|---|---|
| `web/src/app.tsx` | The sign-in page for a signed-out owner. For a signed-in owner, the onboarding screen while no profile exists, then the home screen, the workout screen, the plan screen, the inventory screens, and the profile screen |
| `web/src/shell.tsx`, `web/src/sync-line.tsx` | The shell that fills the whole screen (D-120), and the line of the sync below its header (D-276) |
| `web/src/pages` | The sign-in page, the home screen, the onboarding screen of `web/src/pages/profile.tsx`, the plan screen of `web/src/pages/plan.tsx`, and the workout screen of `web/src/pages/workout.tsx` |
| `web/src/pages/inventory` | The inventory screens: the list, the catalog list and the text entry, the weights, and the review screen |
| `web/src/lib/inventory.ts` | The catalog order, the A to Z order of the inventory list, the search of the catalog names, and the checks of the weights, the estimates, and the notes |
| `web/src/lib/inventory-api.ts`, `web/src/lib/errors.ts` | The changes of the inventory in the outbox, the inventory of the phone with the changes that wait on it (D-272, D-273), and the error text of a failed call |
| `web/src/lib/sync.ts`, `web/src/lib/sync-engine.ts` | The sync of the outbox, the offline copies, and the state of the sync (D-274 to D-278) |
| `web/src/lib/profile.ts`, `web/src/lib/profile-api.ts` | The form state and the checks of the profile, the text of the injury warning, and the save of the profile |
| `web/src/lib/plan.ts`, `web/src/lib/plan-api.ts` | The texts of the progress and of the errors of a plan request, the formats of a set and of the rest, and the streams of a plan request and of an exclusion |
| `web/src/lib/workout.ts` | The start of a workout, the set log, the cardio log, the skip of an exercise, the edit of a set, the end of a workout, each with its outbox entry. Also the calibration step, the rest timer, and the steps of the plus and minus buttons |
| `web/src/lib/symptoms.ts`, `web/src/lib/wake-lock.ts` | The list of symptoms and the text of each warning (D-263), and the screen wake lock (D-265, D-283) |
| `web/src/lib/firebase.ts` | Firebase Authentication with email and password (D-75) |
| `web/src/lib/api.ts` | The Connect transport, with the ID token of the owner on each call |
| `web/src/lib/db.ts` | The offline store and the outbox (D-62, D-77, D-132) |
| `web/src/lib/pwa.ts`, `web/src/lib/update-check.ts` | The service worker and its update strategy (D-133) |
| `web/src/lib/storage.ts` | The persistent storage request (D-134) |
| `web/src/gen` | The generated code of the contract. `make proto` writes it, and Git keeps it. |
| `web/e2e` | The browser tests of the acceptance stories of work areas 2.2, 4.1, 5.1, 5.2, 6.1, 6.2, and 6.3 |

## The screen

The viewport meta holds `maximum-scale=1, user-scalable=no`, so a pinch does not zoom the page, as in Decktome (D-120). This fails the WCAG 1.4.4 rule for text resize. The meta also holds `viewport-fit=cover`. In the Home Screen app, the shell takes the screen height, because the dynamic viewport height leaves out the band of the status bar (PR-28).

A browser tab uses the dynamic viewport height. The shell pads itself with the top and side safe areas. The main region is the one part that scrolls, and its bottom padding holds the bottom safe area (`web/src/lib/app-height.ts`).

Since iOS 26, the Home Screen app blurs a band below the status bar, and the page can not turn it off. So the header adds 16 px above the title when a status bar covers the page (D-145).

The app has no form that makes an account. The owner makes the one account in the Firebase console, and self sign-up is off (D-117).

## The inventory screens

The home screen opens the equipment inventory (work area 4.1). The screens read the catalog and the inventory through `InventoryService` of the API, and never through Firestore (D-77).

- The list shows each machine with its state, draft or confirmed, A to Z by name (D-205), and each note.
- The catalog list shows the machines by kind (D-202), and each kind A to Z by name (D-205). A text searches the names of the machines and of their exercises. The owner selects a match, or keeps the text as a note (D-191).
- A machine or the cable station gets a range and a step, and the save makes the list. A later visit adds or removes single weights (D-195, D-245). The dumbbells get the dumbbell set. A cardio machine gets no weights, and its save confirms it (D-246).
- Each exercise of the machine gets one optional estimate, from the lightest to the heaviest weight (D-192, D-198).
- The review screen confirms the weights that it shows (D-201), and removes the machine.

Each load is a whole number of tenths of a pound, as in the contract. The screens read the offline copies of the catalog and of the inventory, so they work with no connection (D-250). Each change goes into the outbox, and shows at once with "Waiting to sync" (D-272). A confirmation with no connection shows the machine as confirmed. When the server refuses it because the weights changed, the machine is a draft again, and the line of the sync shows the refusal (D-273).

## The onboarding screen

After the sign-in, the app reads the profile through `ProfileService`. With no saved profile, the app opens the onboarding screen before the home screen (D-223). The screen has no Back button, and it can sign out. After the first save, the home screen has a "Profile" button that opens the same screen. A save there goes back to the home screen.

One screen holds each input, with one save (D-41, D-42, D-208 to D-211). Most inputs are large buttons, so the owner types little (D-71):

- The experience, the training days from 2 to 4, and the goal template are buttons. A template selects its muscle groups, and the owner can change them (D-210, D-220).
- The injured areas are buttons. After the owner selects an area, the screen shows the injury warning of D-222.
- The age and the weight use the number keypad. The height uses two lists, feet and inches.
- The cardio preference lists the cardio exercises of the catalog (D-217).

`GetProfileOptions` gives the lists, and `GetCatalog` gives the cardio exercises, so the server holds the one copy of each list. The screen checks each bound of the server before the save (D-215). The save is a direct call to the API, as for the inventory (D-196). A browser test saves a profile through the API with `makeOwner` of `web/e2e/support.ts`, so the other tests go past onboarding.

## The plan screen

The home screen has a "Plan" button. The plan screen reads `GetPlan` with the local date of `web/src/lib/today.ts`, so each target is the target on the date of the next session (D-294, D-295). The sync reads the plan copy with the same date. The screen shows the plan of the API in text alone (D-44, D-73):

- The summary of the plan.
- For each session: the warm-up, each exercise, the optional cardio, and the cool-down.
- For each exercise: the calibration sets, the working sets, the rest, and the reason.
- The mobility and recovery items, and the excluded exercises with their reasons.

Each exercise has a "Change target" button. Its form changes the reps and the load of each working set for the next session (D-69, D-293). It needs a reason of 1 to 200 characters, and it calls `OverrideTarget`.

The policy of the server checks the change. The card then shows the override and the reason above the recommendation, and "Use the recommendation" calls `RemoveOverride`. Each call puts the plan in the query cache and in the plan copy, so the next workout starts with the override. An expired override shows a notice, and the workout uses the recommendation. A workout with an override shows "Your change" with the recommended set, and its target copy keeps the recommendation and the reason.

"Make a plan" and "Make a new plan" call `RequestPlan`. Each exercise has an "Exclude" button. It opens a form with an optional reason of 200 characters or fewer, and the form calls `ExcludeExercise` (D-48, D-228).

Both calls are server streams. While a call runs, the screen goes to the top and shows the text of each progress step (D-239). The screen hides the Back button until the call ends. A close of the screen stops the stream, and the server then saves nothing (D-237).

A failed call shows the text of its error, and the screen reads `GetPlan` again (D-240). The server gives `UNAVAILABLE` after the last failed call. So `UNAVAILABLE` after a progress event is the error of no valid plan, and `UNAVAILABLE` with no progress event is an error of the network.

## The workout screen

The home screen has a "Workout" button. With an open workout on the phone, the button is "Continue workout". The screen reads `GetPlan`, `GetCatalog`, and `GetInventory`, and offers the next session of the plan that the owner did not do yet. The owner can also pick another session (D-248). The next session is the first session with the fewest finished workouts, so the week starts again after its last session.

The start copies the targets of the session and the weights of each machine to the phone. After the start, the workout needs no network (D-62). The phone holds one open workout at most.

- The set log shows the next set of the current exercise: the calibration set first, then the working sets. The reps and the weight come from the target (D-249).
- The plus and minus buttons change the reps by 1. They move the weight to the next weight of the list of the machine (D-264).
- A tap on the reps in reserve logs the set. So the owner logs a set in one tap. A working set offers 0, 1, 2, 3, or 4+, and a calibration set offers 0 to 6+ (D-268).
- After the calibration set, each working set gets the load of the calibration table (D-267). The plan holds the 4 loads for each weight of the machine, so the table applies to the weight that the owner logged (D-249). The set log shows the note "The calibration set gave this load". A plan of policy version 3 gives the load of the plan. A weight that the machine did not have at the time of the plan does too.
- The log of a set starts the rest timer with the rest of the target, 60 seconds since policy version 5 (D-59, D-172, D-279). The timer reads a stored end time in the `meta` table. So it is correct after a screen lock and after a stop of the app. "-15 s", "+15 s", and "Dismiss" change it, and it never goes below 0 (D-270). At 0 it shows "Rest done" in another color, with no sound and no notification (D-58, D-61).
- After the last set of an exercise, the screen shows the next machine for 10 seconds, then advances. "Go now" advances at once (D-60, D-269).
- "Skip this exercise" asks for a confirmation, and writes the skip in the header (D-63, D-170). The owner can pick a skipped exercise again in the list.
- Each logged set has "Edit". The edit keeps the id, the kind, and the time of the set, and writes a new outbox entry with the whole new state (D-63).
- "Add pain or a note" shows the optional pain rating from 0 to 10 and a note of 280 characters or fewer (D-57, D-162, D-261). A pain rating of 1 or more shows the pain warning (D-169).
- The cardio card logs the minutes and the effort from 1 to 10. The distance, the resistance level, pain, and a note are optional (D-123, D-165). Each number must fit its `int32` field of the contract.
- "Report a symptom" shows the seven symptoms of D-263. A pick shows the warning. The owner continues after the confirmation, or uses "Finish now" (D-40, D-153, D-251). The phone keeps no symptom report.
- "Finish now" asks for a confirmation when an exercise that the owner did not skip has a set with no log. The workout then ends early, and each exercise with no logged set counts as skipped (D-63).
- After the end of a workout, the screen "Workout done" shows the next target of each exercise of the workout, with its reason (D-290, D-292). `web/src/lib/revision.ts` reads them from the plan copy when the copy holds the revision of the workout. While an entry of the workout waits for the sync, the screen tells the owner that the next targets show after the sync. "Done" goes back to the home screen.

While a workout is open, the plan screen refuses a new plan and an exclusion (D-252). The app holds the screen wake lock from the start of a workout to its end, on each screen. It asks for the lock again at each return to the front, focus, `pageshow` event, and tap. So a return needs no tap (D-283). A release makes no request of its own, and only the newest request sets the state. When the phone refuses the lock, the workout screen shows "The screen can turn off. Tap the screen to try again." with the error name (D-265).

## The offline store

Dexie on IndexedDB holds the local state. Each change and its outbox entry go into one transaction (D-132). An outbox entry holds a UUIDv7 op id, the entity and its id, and the base version. It also holds the payload, the time, the attempts, and the schema version.

Version 2 of the store adds the workouts, the sets, and the cardio logs. The payload of each workout entry is the JSON form of `WorkoutHeader`, `SetEntry`, or `CardioEntry` of `proto/workoutapp/v1/workout_service.proto`, with the whole new state of the entity. A header holds the target of each exercise from the start of the workout (D-291). The payload of an inventory entry is the JSON form of the payload field of `OutboxEntry`, such as `{"saveMachine": {...}}`. The op ids of the phone rise strictly, so the outbox keeps the order of two changes of one millisecond.

Version 3 adds the refused entries and the offline copies. It removes the settings of the skeleton and their outbox entries.

## The sync

`web/src/lib/sync.ts` sends the outbox through `SyncOutbox`, in batches of 100 entries or fewer, in the order of the op ids (D-259, D-275). The server applies each entry one time by its op id. So a batch that the phone sends again after a dropped answer changes nothing (D-257). The phone removes each applied entry, and keeps the server version on its entity. It moves each refused entry to a separate list, and never sends it again (D-274).

The sync runs while the app is open alone, because iOS has no background sync for a web app (D-21). It runs at the open, at each focus and return, at each reconnect, and after each new entry. After a failure, it tries again after 5 s, 15 s, 60 s, and then each 5 minutes (D-277). A plan request and an exclusion run a sync first, so the plan reads each change of the inventory.

After each drain, the sync reads the catalog, the inventory, and the plan again, and keeps a copy of each. The server wins, and the screens put the entries that wait on the new copy (D-258). The plan screen and the profile gate keep their copies too. So with no connection, the app opens, and a workout starts from the copies (D-278).

A sync of a finished workout revises the plan on the server (D-292). So each sync that passes marks the cached `GetPlan` as stale. The plan screen keeps a fresh read of the plan alone in the copy.

The line below the header shows "Synced", the count of the entries that wait, "Offline", "Sync failed", and the count of the refused entries (D-276). A tap opens the detail with "Sync now", and each refused entry with "Dismiss".

The first sign-in on a device asks for persistent storage (D-134). The home screen shows the result and the use of the store.

## Updates

The service worker waits after an update, and the app shows "Update ready" (D-133). The owner applies the update with the button. The app never applies an update during a workout. While a workout is open on the phone, the banner says "Update ready after the workout", with no button. The app checks for an update each hour and at each return to view.

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

The browser tests start the Auth and Firestore emulators of `firebase.json`. They also start the API of `go/` on port 8480, and a build of the app on port 4273. No call reaches a real project (D-115). The API uses the fake provider of Luna, so no test calls OpenAI (D-24). Each call of the fake waits 1 s, so the plan tests can see the progress (D-241).

`makePlanOwner` of `web/e2e/support.ts` saves a profile and confirms machines through the API. A route stub gives the error streams of the cap and of 4 failed calls. The tests make each account on the Auth emulator and write its allowlist document on the Firestore emulator.

Playwright can make a pinch in Chromium alone. So the pinch test runs in Chromium, and WebKit reads the viewport meta. A headless browser has no Home Screen and no status bar. So the device check of PR-11 read the bottom edge and the blur band on the iPhone. A Chromium test sets the safe area through the DevTools protocol, and it reads the space above the title.
