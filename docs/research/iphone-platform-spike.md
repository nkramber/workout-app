# Gym Route - iPhone web platform spike report

Status: spike report of work area 1.3 of `docs/roadmaps/high-level-roadmap.md`. It holds measurements and recommendations. The owner decisions live in `docs/decisions.md`.

Date of the device checklist: 2026-09-29. The probe is in `tools/spikes/iphone_probe/`. The owner ran each item on the iPhone of the owner. The session read the results from screenshots of the probe pages and from the notes of the owner.

## 1. Result

**Go.** The iPhone web platform risk passes each part of the go bar of D-118. For that bar, a cold start is a new page load of the Home Screen app after an app stop (D-121).

| Item | Result | Go bar (D-118) | Pass |
|---|---|---|---|
| 1. IndexedDB data after an app stop | 3 of 3 records of the earlier launch stayed | The item passes in the Home Screen app | Yes |
| 2. Screen Wake Lock | The screen stayed on past the Auto-Lock time, and the lock came back after the background | The item passes in the Home Screen app | Yes |
| 3. Add to the Home Screen from Chrome | The app opened in the `standalone` display mode | The item passes | Yes |
| 4. Sign-in in the Home Screen app | Signed in with email and password, with no error | The item passes in the Home Screen app | Yes |
| 5. Sign-in state after an app stop | Signed in after the stop, and the session read took 358 ms | The item passes in the Home Screen app | Yes |
| 6. Startup time of the React build | Median first contentful paint of 33 ms over 5 cold starts (D-121) | 2500 ms or less | Yes |

The startup time measures from the start of the navigation. It does not include the time from the tap on the icon to that start. Section 5 gives this limit.

The owner also saw two display faults of the probe: the layout sat too high, and a pinch zoomed the page. They are not items of the bar. Section 3.6 gives them, and D-120 makes the app of the roadmap fix them.

## 2. Method

### 2.1 The device and the build

| Fact | Value |
|---|---|
| Device | iPhone 16 Pro |
| iOS | 27.0 |
| Chrome | 154.0.8037.55 |
| Engine | WebKit of iOS 27.0, because Chrome on iOS uses WebKit (D-29) |
| Probe build | `d6e3c54`, the merge commit of the probe on `main` |
| Site | `https://gym-route-dev.web.app`, Firebase Hosting of the project `gym-route-dev` (D-99) |
| Network | Cellular 5G |

The session deployed the probe by hand from a clean worktree of `main` at `d6e3c54` (D-14). Before the deploy, `make where` gave the branch `main`, a clean tree, and no difference from `origin/main`. After the deploy, the live page gave the header `build d6e3c54`. Each screenshot of the owner shows the same build id.

### 2.2 The procedure

The owner made the probe account in the Firebase console, because self sign-up is off (D-117). Then the owner did these steps:

1. Open the site in Chrome, and add it to the Home Screen with the share button.
2. Open the probe from the Home Screen. Read the Install page.
3. Write three records on the Storage page, and ask for persistent storage.
4. Stop the app in the app switcher, open it again, and read the Storage page.
5. Turn on the wake lock on the Wake page, and wait past the Auto-Lock time.
6. Send the app to the background, and open it again. Read the log of the Wake page.
7. Sign in on the Sign-in page.
8. Stop the app, open it again, and read the Sign-in page.
9. Stop the app and open it again five times. Open the Startup page in each launch.

Each item ran in the Home Screen app, as D-119 sets. Item 3 started in Chrome, because Chrome adds the app to the Home Screen.

### 2.3 The measures

- The probe gives each page load a random launch id. A new launch id shows that the app loaded again from the start.
- The Storage page counts the records that an earlier launch wrote.
- The Startup page takes each time from the start of the navigation. It gives four times: the inline script, the bundle start, the first React render, and the first contentful paint.
- The Startup page records one launch each time the owner opens it in a new launch. It gives the median paint of the stored launches in each display mode.
- A cold start is a new page load of the Home Screen app after an app stop in the app switcher (D-121). The page can not see whether iOS ended the web view process.

## 3. Results for each item

### 3.1 Item 3: add to the Home Screen from Chrome

The share button of Chrome gave "Add to Home Screen", and the app opened from the icon. This result closes the unresolved Chrome install row of `docs/research/platform-cloud-and-ai.md`.

| Row of the Install page | Value |
|---|---|
| Display mode | `standalone` |
| Manifest | Gym Route probe, `standalone`, 3 icons |
| Service worker | `activated` |
| Page controlled by the worker | no, in the first launch |
| Install prompt available | no |

The first launch of the Home Screen app registered the service worker. The worker did not control that first page, because the worker configuration has no `clientsClaim`. In the fifth launch of section 3.5, the worker served the page. The screenshots do not show this value for the other launches.

iOS has no install prompt event. So "no" is the expected value of the last row.

### 3.2 Item 1: IndexedDB data after an app stop

The owner wrote three records in launch `g5moeir4`, between 13:45:43 and 13:45:44 UTC. After the stop, launch `mbnk4xjq` read the records again.

| Row of the Storage page | Value |
|---|---|
| Records | 3 |
| Records from an earlier launch | 3 |
| Persistent storage | yes |
| Storage used | 525 KB |

WebKit gave persistent storage to the Home Screen app with no prompt. This agrees with the storage policy of PC-4 in `docs/research/platform-cloud-and-ai.md`.

### 3.3 Item 2: Screen Wake Lock

The owner reported two results:

- The screen stayed on past the Auto-Lock time of the phone.
- The log showed that the lock came back after the app went to the background and returned.

So the lock works in the Home Screen app on iOS 27.0, and the request on `visibilitychange` works (PC-2, PC-13). The owner did not send the Auto-Lock time or the "Held for" time. Section 5 gives this limit.

### 3.4 Items 4 and 5: sign-in, then an app stop

| Launch | Step | State | Session read in |
|---|---|---|---|
| `mbnk4xjq` | After the sign-in with email and password | signed in | 92 ms |
| `00adc34c` | After the app stop | signed in | 358 ms |

The sign-in gave no error. The new launch id shows a new load of the app, and the app read the stored session with no second sign-in. The report holds no email and no user id (D-80).

### 3.5 Item 6: startup time of the React build

Five cold starts under D-121, between 13:49:12 and 13:49:23 UTC. Each launch was in the `standalone` display mode, with the navigation type `navigate`:

| Launch | First React render (ms) | First contentful paint (ms) |
|---|---|---|
| 1 | 30 | 34 |
| 2 | 31 | 33 |
| 3 | 31 | 30 |
| 4 | 31 | 32 |
| 5 | 30 | 33 |

The median first contentful paint is 33 ms. The fifth launch gave these other times:

- The inline script ran at 14 ms.
- The app bundle started at 22 ms.
- DOMContentLoaded ended at 23 ms.

The service worker served the page.

The research target of section 4.3 of `docs/research/platform-cloud-and-ai.md` was a first interaction under about 2.5 s on a cold start. The measured paint is far below that target and the bar of D-118.

### 3.6 Display faults of the probe

The owner saw two faults in the Home Screen app. They are not items of the checklist, and they do not change the go.

- The layout sat too high. A gap stayed below the bar of page links, at the bottom edge of the screen.
- A pinch zoomed the page in and out.

In the screenshots, the gap is about 66 points high. The top safe area of the iPhone 16 Pro has about the same height. The probe uses `height: 100%` for its shell, with the `black-translucent` status bar and `viewport-fit=cover`. So the probable cause is a shell of the screen height minus the status bar, which starts below the status bar (unverified). The viewport meta of the probe does not block the pinch zoom.

The owner decided that the probe can keep both faults, and that the app of the roadmap fixes both (D-120). Decktome solved the same faults with `h-dvh` for the shell and `maximum-scale=1, user-scalable=no` in the viewport meta (decktome:D-624, decktome:D-625). Work area 2.2 of `docs/roadmaps/high-level-roadmap.md` holds the fix.

## 4. Findings for the app of the roadmap

- Chrome on iPhone can add the app to the Home Screen, and the app opens in the `standalone` display mode. The install screen of the app can give the steps of the Install page.
- The Home Screen app keeps its IndexedDB data after an app stop, and WebKit marks the storage as persistent. This supports REC-5 of `docs/research/platform-cloud-and-ai.md`.
- Firebase Authentication with email and password keeps its session in the Home Screen app after an app stop.
- Screen Wake Lock works in the Home Screen app. The request again on each return to view, as REC-6 says, works.
- The React build of D-84 gives a first paint of about 33 ms from the service worker cache on this device.
- A shell of `height: 100%` with a translucent status bar left a gap at the bottom edge. The shell of the app must fill the whole screen (D-120).

## 5. Limits

- One device only. The iPhone 16 Pro is a fast phone, so a slower phone can give a longer startup time.
- The startup time starts at the start of the navigation. The page can not see the time from the tap on the icon to that start, when iOS starts the web view.
- The five launches came in 11 seconds. After an app stop, iOS can keep parts of the web view in memory. D-121 counts these launches as cold starts, but a launch after a phone restart or a long pause can be slower.
- The probe is small. Its main bundle is 236.71 KB before compression. The app of the roadmap will have more code, so its startup time will be longer.
- The storage test covers an app stop, not seven days with no use or a phone restart (PC-5).
- The wake lock result comes from the notes of the owner, with no Auto-Lock time and no "Held for" time.
- The worker served the fifth launch, so the cellular network had no effect on it. The screenshots do not show this value for the other four launches.

## 6. Recommendations (not owner decisions)

- Keep the React stack of D-84. The startup time gives no reason to change it.
- Add `clientsClaim` to the worker of the app, so that the first launch also gets the worker cache (Recommendation).
- Measure the startup time of the app again when its first screens exist. Measure from the tap to the first paint, after a long pause (Recommendation).
