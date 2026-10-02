# Set up the project on Google Cloud

This document gives the live state of the project `nk-workout-app-prod`, and the steps that make it again. The structure follows `decktome:docs/setup-gcp.md`. `docs/deploy-and-rollback.md` gives the deploy and the rollback.

The date of this version is 2026-10-02. The session of work area 2.3 made each part with the approval of the owner at run time, and it read each part back. The session of PR-21 changed the role of `api-runtime` (D-206).

CAUTION: do not write an account email, a uid, or a secret value into this file. The repository is public (D-106).

## 1. What the project holds

| Part | Value | Decision |
|---|---|---|
| Project | `nk-workout-app-prod`, number 665413986587, display name "Workout App", no organization | D-137 |
| Billing | The one open billing account of the owner | D-139 |
| Budget | 10 USD each month, this project alone, alerts at 50%, 90%, and 100%. It does not stop spend. | D-139 |
| Firebase Authentication | Email and password alone, self sign-up off, one account of the owner | D-75, D-117 |
| Browser key | `identitytoolkit` and `securetoken` alone, from the two Firebase domains of the project and ports 4173 and 5173 of `localhost` and `127.0.0.1` | D-117, D-137 |
| Firestore | Default database, Standard edition, Native mode, `us-central1`, delete protection on | D-76, D-140 |
| Firestore rules | `firestore.rules`: each client read and write is refused | D-77 |
| Point-in-time recovery | On, with a window of 7 days | D-124 |
| Backups | A daily schedule, and each backup stays 10 days | D-124 |
| Allowlist | One document in `allowlist`, with the uid of the owner as its id | D-75, D-131 |
| Cloud Run | Service `api` in `us-central1`, request billing, min instances 0, max instances 2, CPU boost | D-141 |
| Secret Manager | Secret `openai-api-key`. The owner added version 1 on 2026-10-01 in a local terminal. | D-24, D-187 |
| Artifact Registry | Docker repository `workout-app` in `us-central1` | - |
| Deploy lock | Bucket `nk-workout-app-prod-deploy-lock` in `us-central1`, with public access prevention, and a rule that deletes each object after one day | D-143 |
| Cloud Build | The connection `github` to `nkramber/workout-app`, and the triggers `deploy-api`, `deploy-web`, and `deploy-rules` | D-14, D-142 |
| Hosting | The site `nk-workout-app-prod`, at `https://nk-workout-app-prod.web.app` | D-18 |

## 2. The service accounts

Each account holds the roles of one job, and no account holds a basic role.

| Account | Roles | Use |
|---|---|---|
| `api-runtime` | `roles/datastore.user`, and the accessor role on `openai-api-key` alone | The Cloud Run service `api` runs as this account. |
| `api-deployer` | `roles/run.developer`, `roles/artifactregistry.writer`, `roles/logging.logWriter`, and `roles/iam.serviceAccountUser` on `api-runtime` alone | The trigger `deploy-api` |
| `web-deployer` | `roles/firebasehosting.admin`, `roles/logging.logWriter` | The trigger `deploy-web` |
| `rules-deployer` | `roles/firebaserules.admin`, `roles/serviceusage.serviceUsageViewer`, `roles/logging.logWriter` | The trigger `deploy-rules` |

Each deployer account also holds `roles/storage.objectUser` on the bucket `nk-workout-app-prod-deploy-lock` alone (D-143).

Before a release of the rules, `firebase-tools` reads the state of the Firestore API. So `rules-deployer` holds the viewer role of Service Usage, which can not change a service (D-146).

The web build installs npm code, so `web-deployer` can release Hosting alone. A bad package can not deploy the API, and it can not open the rules (D-142).

Note: `api-runtime` reads the allowlist, and it reads and writes the inventory of PR-19. Until 2026-10-02 it held `roles/datastore.viewer` alone, so each save gave HTTP 500 in the live app. The emulator applies no IAM, so the emulator tests did not find the fault (D-206).

CAUTION: a new write path of the API needs a write role on the live project. The emulator tests can not prove the role. Read the role of `api-runtime` before the live check of a new write path.

## 3. Before you start

- Sign in to `gcloud` with the account of the owner. The Firebase CLI uses the same account (D-116).
- Use the pinned `firebase-tools` of `emulators/`. Run `npm ci --prefix emulators` one time.
- Add `--project nk-workout-app-prod` to each command. The default configuration of `gcloud` can name another project.
- Read each part back after its step.

## 4. Make the project again

The steps below made the project on 2026-09-29. A new project id needs a change of each file that names the project. Run `git grep nk-workout-app-prod` to find them.

### 4.1 The project and Firebase

1. Run `gcloud projects create nk-workout-app-prod --name="Workout App"`.
2. Run `firebase projects:addfirebase nk-workout-app-prod`.
3. Run `firebase apps:create web "Workout App" --project nk-workout-app-prod`.
4. Run `firebase apps:sdkconfig WEB <app id>`, and write the four values into `web/src/lib/firebase-config.ts`.
5. In a folder with a `firebase.json` that holds `{"auth":{"providers":{"emailPassword":true}}}`, run `firebase deploy --only auth`.
6. Set `client.permissions.disabledUserSignup` to `true` with a PATCH of the Identity Toolkit `admin/v2` config.
7. Limit the browser key with `gcloud services api-keys update`, as section 1 gives.
8. The owner adds the one account in the Firebase console.

CAUTION: read each value before you write it into a file or a command. A placeholder in angle brackets writes itself as the value (`decktome:docs/deploy-and-rollback.md`).

### 4.2 Billing and the budget

1. Run `gcloud billing projects link nk-workout-app-prod --billing-account=<account id>`.
2. Run `gcloud services enable billingbudgets.googleapis.com --project nk-workout-app-prod`.
3. Run `gcloud billing budgets create` with `--budget-amount=10USD` and `--filter-projects=projects/nk-workout-app-prod`.
4. Add `--threshold-rule=percent=0.5`, `0.9`, and `1.0` to the command of step 3.

### 4.3 Firestore

CAUTION: the location of a database is permanent. Check `us-central1` before the create step.

1. Run `gcloud firestore databases create --location=us-central1 --type=firestore-native --edition=standard`.
2. Run `firebase deploy --only firestore:rules` from a clean checkout of `main`.
3. Run `gcloud firestore databases update --database='(default)' --enable-pitr`.
4. Run `gcloud firestore databases update --database='(default)' --delete-protection`.
5. Run `gcloud firestore backups schedules create --database='(default)' --recurrence=daily --retention=10d`.

### 4.4 The API side

1. Turn on the Run, Artifact Registry, Cloud Build, Secret Manager, and IAM APIs.
2. Make the four accounts of section 2, and give each account its roles.
3. Run `gcloud secrets create openai-api-key --replication-policy=user-managed --locations=us-central1`.
4. Run `gcloud artifacts repositories create workout-app --location=us-central1 --repository-format=docker`.
5. Make the bucket `nk-workout-app-prod-deploy-lock` with `--uniform-bucket-level-access` and `--public-access-prevention`.
6. Give each deployer account `roles/storage.objectUser` on that bucket.
7. Make the service `api` with the `hello` image of Google, as the block below gives.

```
gcloud run deploy api --project nk-workout-app-prod --region us-central1 \
  --image us-docker.pkg.dev/cloudrun/container/hello \
  --service-account api-runtime@nk-workout-app-prod.iam.gserviceaccount.com \
  --allow-unauthenticated --cpu-throttling --min-instances 0 --max-instances 2 --cpu-boost \
  --set-env-vars GOOGLE_CLOUD_PROJECT=nk-workout-app-prod,ALLOWED_ORIGIN=https://nk-workout-app-prod.web.app
```

The first image is a placeholder, because the API deploys from `main` alone (D-14). The API checks each token itself, so the service lets each caller in (D-82). The fixed URL of the service is `https://api-665413986587.us-central1.run.app`, and `cloudbuild/web.yaml` names it.

### 4.5 The allowlist entry

The API reads the document `allowlist/<uid>` (D-131). Write it with the uid of the owner. Keep the uid out of the terminal output:

```
P=nk-workout-app-prod
TOKEN=$(gcloud auth print-access-token)
H=(-H "Authorization: Bearer $TOKEN" -H "x-goog-user-project: $P")
UID_=$(curl -s "${H[@]}" "https://identitytoolkit.googleapis.com/v1/projects/$P/accounts:batchGet?maxResults=10" \
  | python3 -c 'import json,sys; u=json.load(sys.stdin)["users"]; assert len(u)==1; print(u[0]["localId"])')
curl -s -o /dev/null -w '%{http_code}\n' -X PATCH "${H[@]}" -H 'Content-Type: application/json' \
  "https://firestore.googleapis.com/v1/projects/$P/databases/(default)/documents/allowlist/$UID_" -d '{"fields":{}}'
```

### 4.6 Cloud Build

`docs/deploy-and-rollback.md` section 2 gives the connection and the triggers.

## 5. Cost

The owner expects about 1 USD each month in Phase 2. The main parts are Firestore storage, the point-in-time recovery storage, the backups, and the images. Cloud Run at min instances 0 costs nothing while it waits. Cloud Build gives 2,500 build-minutes each month at no cost, and three builds of one merge use about 10 minutes (assumption, not measured).

The budget sends an email to the billing admins, and it does not stop spend (PC-82). The owner skipped the Cloud Run spend cap (D-141). The cap is a Preview feature that the console alone sets (PC-83, read on 2026-09-29).

## 6. The old project

The owner signed in on the new project, so the session shut down the project `gym-route-dev` of D-116 on 2026-09-30 (D-137). Its state is `DELETE_REQUESTED`. Google Cloud keeps it for 30 days, and `gcloud projects undelete gym-route-dev` can restore it in that time. `docs/research/phase-2-check.md` section 5 gives the read back.
