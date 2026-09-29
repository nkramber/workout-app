# iPhone web platform probe

This folder holds the probe of the iPhone web platform spike, work area 1.3 of `docs/roadmaps/high-level-roadmap.md`. The probe is a small web app on the React stack of D-84. It stays outside the product code. PR-6 runs the device checklist on the iPhone of the owner and writes the report.

The probe has one page for each device item. It has no camera page (D-112).

| Page | Device item | What the page shows |
|---|---|---|
| Storage | 1. IndexedDB data after an app stop | The records, the records of an earlier launch, and the persistent storage state |
| Wake | 2. Screen Wake Lock | The API, the lock state, the time that the lock stays on, and a log |
| Install | 3. Install from Chrome | The display mode, the manifest, the service worker, and the install prompt |
| Sign-in | 4 and 5. Sign-in, then an app stop | The sign-in state, a short user id, and the time to read the stored session |
| Startup | 6. Startup time | The times of this launch, and the median paint time of each display mode |

The header shows the build id, the launch id, and the display mode. The build id is the short commit of the build, so the report names the build that it tested.

## Files

| File | Content |
|---|---|
| `src/` | The app. `src/pages/` holds one file for each page. `src/lib/` holds the IndexedDB store, the startup timer, and the Firebase module. |
| `src/lib/firebase-config.ts` | The public web configuration of the development project (D-76, D-99). |
| `e2e/probe.spec.ts` | The browser tests. They run in WebKit and in Chromium (D-113). |
| `test_probe_config.py` | The configuration tests. `make test` runs them. |
| `firebase.json` | Firebase Hosting, the email and password provider (D-75), and the Auth emulator. |
| `scripts/make-icons.mjs` | It renders the PNG icons from `public/favicon.svg`. |

## Checks

`make probe` builds the probe and runs the browser tests. It needs Node 22, the version of `.nvmrc`. It downloads the WebKit and Chromium browsers of Playwright when they are not on the machine. The CI job `verify:probe` runs the same target (D-113). The job is not a required check (D-114).

The browser tests use the real Firebase Auth SDK against the local Auth emulator of `firebase-tools` (D-115). They make no call to the real project. A test makes its account on the emulator with an address of the reserved domain `example.com`.

`make test` runs `test_probe_config.py`. It needs no Node. It checks these rules:

- `firebase.json` names Firebase Hosting and Firebase Authentication only (D-99).
- The app imports no Firebase module other than `firebase/app` and `firebase/auth`.
- Each package has an exact version, and the lock file agrees.
- No file holds an email address outside `example.com`.
- The app does not call the camera (D-112).

## Sign-in

The probe has no form that makes an account, and self sign-up is off in the project (D-117). The owner makes the account in the Firebase console. The page shows the user id only, and it never shows or logs the email (D-80).

The probe reads the stored session from IndexedDB first, and then from local storage. So the device checklist tests the storage that the app of the roadmap uses.

## The project settings

The project `gym-route-dev` holds these settings outside the repository (D-116, D-117). The owner set them with approval, and `firebase.json` does not hold them.

| Setting | Value |
|---|---|
| Billing | Off. The project stays on the free plan (D-99). |
| The APIs of the browser key | `identitytoolkit.googleapis.com` and `securetoken.googleapis.com` only |
| The sites of the browser key | `https://gym-route-dev.web.app/*`, `https://gym-route-dev.firebaseapp.com/*`, and ports 4173 and 5173 of `localhost` and `127.0.0.1` |
| Self sign-up | Off. A sign-up call gives `ADMIN_ONLY_OPERATION`. |

The browser key is in `src/lib/firebase-config.ts`. GitHub secret scanning flags it, but each browser that loads the probe gets the key. The settings above limit what the key can do. A new port for the probe needs a new site on the key.

## Deploy

PR-6 deploys the probe by hand from a clean checkout of `main` (D-14, D-99). Do these steps in this folder:

1. Run `make where`, and make sure that the branch is `main` and that the tree is clean.
2. Run `npm ci`, then `npm run build`.
3. Run `npx firebase deploy --only hosting --project gym-route-dev`.

The deploy sends the folder `dist` to Firebase Hosting. It changes no other service.

## Notes from the build

- WebKit refuses a page on port 4190, because the browser holds that port on its list of restricted ports. The preview server uses port 4173, the default of Vite.
- The Auth emulator answers a wrong password with `auth/wrong-password`. The real service answers `auth/invalid-credential`. The test accepts both codes.
