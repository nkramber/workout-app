# Workout App - Phase 2 device and deploy check

Status: check report of work areas 2.2 and 2.3 of `docs/roadmaps/high-level-roadmap.md`. It holds the exit evidence of Phase 2. The owner decisions live in `docs/decisions.md`.

Date of the checks: 2026-09-30 (UTC). The owner ran each device item on the iPhone of the owner. The session read the live endpoints, the builds, and the logs with `curl`, `gcloud`, and the Firebase Rules API.

## 1. Result

**Pass.** Each item of the device check of PR-11 passes, and the live API, web app, and rules name the merge commit of PR-10.

| Item | Result | Pass |
|---|---|---|
| Deploy: `/version` of the API | `df79c169ae38326c0a876ac678a5465f80103d13` | Yes |
| Deploy: `/version.json` of the web app | `df79c169ae38326c0a876ac678a5465f80103d13` | Yes |
| Deploy: the Firestore rules | The build `deploy-rules` of `df79c16` released the deny-all rules, after the repair of section 3 | Yes |
| 1. Add the app to the Home Screen from Chrome | The app opened as a Home Screen app | Yes |
| 2. Sign in with email and password | The owner signed in, and the API answered `GetMe` with HTTP 200 | Yes |
| 3. The home screen with the answer of the API | The home screen showed the uid from the live API | Yes |
| 4. A gap at the bottom edge | No gap | Yes |
| 5. A pinch zoom | No zoom | Yes |

The owner also saw a blur on the top half of the title "Workout App". It is not an item of the check. Section 4 gives the cause and the fix (D-145).

## 2. The deploy check

The merge of PR-10 started the three builds at 2026-09-30T02:17:29Z. `deploy-api` and `deploy-web` passed. `deploy-rules` failed (section 3).

At 02:29:42Z and again at 03:54:47Z, the two endpoints gave the same commit:

```
GET https://api-665413986587.us-central1.run.app/version  -> {"commit":"df79c169ae38326c0a876ac678a5465f80103d13"}
GET https://nk-workout-app-prod.web.app/version.json      -> {"commit":"df79c169ae38326c0a876ac678a5465f80103d13"}
```

The sign-in of item 2 reached the revision `api-00002-6f2`, at 03:15:41Z. The log holds the status and the path alone (D-80).

## 3. The repair of the rules build

The build `deploy-rules` of `df79c16` stopped with this error of `firebase-tools` 15.32.0:

```
HTTP Error: 403, Permission denied to get service [firestore.googleapis.com]
```

Before the release, `firebase-tools` reads the state of the Firestore API through `serviceusage.googleapis.com`. The account `rules-deployer` held `roles/firebaserules.admin` and `roles/logging.logWriter` alone. That set has no permission to read the state of a service.

The repair, with the approval of the owner at run time (D-146):

1. The session gave `rules-deployer` the role `roles/serviceusage.serviceUsageViewer` on the project. The role reads the state of a service, and it can not turn a service on or off.
2. The session started the failed build again through the `builds/{id}:retry` call of the Cloud Build API. The build `60bcf32b` of `df79c16` ran as `rules-deployer`, and it passed.
3. The session read the release `cloud.firestore` back. It names the ruleset `e324b2f9`, and its update time is 02:33:00Z, inside the build time of 02:31:52Z to 02:33:03Z.

The ruleset has the same text as `firestore.rules`. `firebase-tools` found no change of the text, so it released the same ruleset again. The installed `gcloud` has no `builds retry` command, so the session used the API call. The console "Rebuild" button uses the same call.

## 4. The blur band below the status bar

### 4.1 The fault

In the Home Screen app, the top half of the title "Workout App" showed a blur. The status bar is `black-translucent`, and the header starts at the safe area. The web app has no blur, no `backdrop-filter`, and no `filter` of its own.

### 4.2 The cause

Since iOS 26, the Home Screen app draws a Liquid Glass "scroll edge effect" over a band of the page below the status bar. The WebKit source, read on 2026-09-30, gives these facts:

- `WKWebView` has three reasons to hide the top blur: an element in full screen, a Mac rule, and a list of sites.
- Else, it hides the blur when it shows a flat color view at the top edge. It shows that view only when the host app gives a top inset (`_obscuredInsets.top`). The bars of Safari give such an inset. A Home Screen app that fills the screen gives none.
- `LocalFrameView::fixedContainerEdges` reads the color of a fixed or sticky box at the top edge. That color goes to the flat color view, so it has no effect when the view does not show.

Sources: `Source/WebKit/UIProcess/API/Cocoa/WKWebView.mm`, `Source/WebKit/UIProcess/API/ios/WKWebViewIOS.mm`, and `Source/WebCore/page/LocalFrameView.cpp` of the WebKit repository, branch `main`. Apple gives no document for this behavior (unverified beyond the source).

### 4.3 The tests on the iPhone

Only `main` deploys to the live project (D-14). So before the merge, the Mac of the owner served each build of the branch on the local network. The owner added each build to the Home Screen from Chrome. iOS opened each one as a Home Screen app, with no browser bars. The sign-in does not work there, because the browser key allows the sites of the project alone (D-117). The sign-in page has the same header, so the check needs no sign-in.

| Build | Change | Result on the iPhone |
|---|---|---|
| A | The whole shell is `position: fixed` at the top, with the opaque color `#0b1220` | The blur stayed |
| B | A thin fixed bar of `#0b1220` covers the top edge, and the title stays at the live position | The blur stayed |
| C | `apple-mobile-web-app-status-bar-style` is `default` | The blur stayed |
| D | The header adds 16 px above the title when a status bar covers the page | The title is sharp |

A and B agree with section 4.2: a page can not turn off the blur of a Home Screen app. The owner did not test a build of 12 px or less. The owner chose 16 px and asked for no more space (D-145).

### 4.4 The fix

`web/src/shell.tsx` gives the header a top padding of `calc(0.75rem + min(env(safe-area-inset-top), 16px))`. With a top safe area, the title moves 16 px lower. With no top safe area, as in a browser tab, the title does not move.

A browser test of `web/e2e/shell.spec.ts` sets a top safe area of 59 px through the DevTools protocol of Chromium. It reads a space of 28 px above the title, and 12 px with no safe area. The test failed on the old header. Chromium draws no Liquid Glass blur, so the owner reads the blur on the iPhone.

Decktome has the same shell and the same blur. This repository can not change Decktome.

## 5. The old project

The owner signed in on `nk-workout-app-prod`. So the condition of D-137 holds, and the owner approved the shutdown. At 2026-09-30T03:19:47Z, the session ran `gcloud projects delete gym-route-dev`. The project state is `DELETE_REQUESTED`. Google Cloud keeps the project for 30 days, and `gcloud projects undelete gym-route-dev` can restore it in that time.

At 03:54:47Z, `https://gym-route-dev.web.app` still gave HTTP 200. The Hosting site of a project stops some time after the shutdown (unverified: the session did not find the delay in a Google document).

## 6. The package rules

`buf.yaml` now holds `PACKAGE_NO_DELETE` and `PACKAGE_SERVICE_NO_DELETE` (D-138). The old rules compare the files by path, so they passed the move of PR-10 from `gymroute.v1` to `workoutapp.v1`. The new rules refuse that move:

```
<input>:1:1:Previously present package "gymroute.v1" was deleted.
```

`make contract` now runs `scripts/package_move_probe.sh`. The probe moves the contract to a new package and path in a temporary tree, and it expects `buf breaking` to refuse the move. The probe fails with the old rules of `buf.yaml`.
