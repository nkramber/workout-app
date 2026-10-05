# Deploy after a merge, and roll back

This document gives the deploy of the project `nk-workout-app-prod` and the rollback of each part. The structure follows `decktome:docs/deploy-and-rollback.md`. `docs/setup-gcp.md` gives the project and its accounts.

The date of this version is 2026-09-29.

## 1. A merge to main deploys itself

Cloud Build builds and releases each merge to `main` (D-14, D-18). No person runs a deploy step. Three triggers read the diff of the merge, and each trigger deploys one part.

| Trigger | It fires on | Account | It does |
|---|---|---|---|
| `deploy-api`, `cloudbuild/api.yaml` | `go/**`, `docker/**`, `cloudbuild/api.yaml` | `api-deployer` | Builds the image, deploys the service `api`, and waits for `/version` |
| `deploy-web`, `cloudbuild/web.yaml` | `web/**`, `firebase.json`, `cloudbuild/web.yaml` | `web-deployer` | Builds the web app, releases Hosting, and reads `/version.json` |
| `deploy-rules`, `cloudbuild/rules.yaml` | `firestore.rules`, `cloudbuild/rules.yaml` | `rules-deployer` | Releases the Firestore rules |

A merge of documents alone starts no build. A pull request starts no build, because each trigger reads `^main$` alone.

Each check reads the commit of its build:

- `GET /version` of the API returns `{"commit":"<sha>"}`. The Dockerfile gives the commit to the Go build.
- `/version.json` of the web app returns the same form. The web build writes it after the Vite build, so the service worker does not keep it.

The three triggers run apart, so the build of an older merge can reach its deploy after the build of a newer merge. So each build deploys its part inside one critical section. `docs/tools/deploy_order.py` holds the rule:

1. The build gets the lock of its part: the object `api.lock`, `web.lock`, or `rules.lock` in the bucket `nk-workout-app-prod-deploy-lock`.
2. Cloud Storage creates the object for one caller alone, so two builds of one part can not deploy together.
3. The build reads `main`. When a commit after its own commit changes a path of the trigger, the build skips.
4. Else the build runs its deploy command, and then it removes the lock.

A skipped build writes `/workspace/.deploy-skip`, and its check step does nothing. The build of the newer commit deploys the part. A newer commit that changes no path of the trigger starts no build of the trigger, so it never makes a build skip.

A build waits up to 15 minutes for the lock, and then it fails. A live build holds the lock for 15 minutes at most:

- The read of `main` stops after 5 minutes. A read that ends later makes the build stop with no deploy.
- The deploy command stops after 10 minutes.

A lock older than 20 minutes is stale, and the next build removes it. The 5 minutes between the two limits give a deploy on the server side time to end after its command stops. The bucket deletes each object after one day.

When the build of the newer merge fails, the part stays at an older commit until a fix merges.

Note: the web build and the API build do not wait for each other. After a merge that changes the contract, the new web app can call the old API for a short time.

## 2. The one-time setup

The setup ran on 2026-09-29, and the three triggers are live.

1. The owner links the repository in the console. Open Cloud Build, then Repositories, then the 2nd gen tab.
2. Make a host connection in `us-central1` with the name `github`. Leave the KMS key empty.
3. Authorize GitHub, and install the app on `nkramber/workout-app` alone.
4. Link the repository `nkramber/workout-app`.
5. Make each trigger with the command below. Change the name, the account, the file, and the paths.

```
P=nk-workout-app-prod
REPO=projects/$P/locations/us-central1/connections/github/repositories/nkramber-workout-app
gcloud builds triggers create github --project $P --name=deploy-api --region=us-central1 \
  --repository=$REPO --branch-pattern='^main$' \
  --service-account=projects/$P/serviceAccounts/api-deployer@$P.iam.gserviceaccount.com \
  --build-config=cloudbuild/api.yaml --included-files='go/**,docker/**,cloudbuild/api.yaml'
```

6. Run `gcloud builds triggers list --region=us-central1 --project nk-workout-app-prod`, and compare it with the table of section 1.

CAUTION: a trigger of a 2nd-gen repository needs `--service-account` and `--region`. Without them the API refuses the trigger, or the trigger finds no connection (`decktome:docs/deploy-and-rollback.md`).

## 3. Before a deploy by hand

A deploy by hand is for a repair only. Obey these rules (D-14):

1. Run `git fetch origin` and `git switch --detach origin/main` in a clean worktree.
2. Run `git status`, and make sure that the tree is clean.
3. Run the build of the part again from the Cloud Build history, if you can.

The "Rebuild" button of the console and the `builds/{id}:retry` call of the Cloud Build API do the same step. On 2026-09-30, Google Cloud SDK 533.0.0 had no `gcloud builds retry` command. So PR-11 used the API call:

```
T=$(gcloud auth print-access-token)
curl -s -X POST -H "Authorization: Bearer $T" -H 'Content-Type: application/json' -d '{}' \
  "https://cloudbuild.googleapis.com/v1/projects/nk-workout-app-prod/locations/us-central1/builds/<build id>:retry"
```

## 4. Roll back

### 4.1 The API

Cloud Run keeps each revision. A rollback moves the traffic, and it needs no build.

1. Run `gcloud run revisions list --service api --region us-central1 --project nk-workout-app-prod`.
2. Read the name of the last good revision.
3. Run `gcloud run services update-traffic api --region us-central1 --to-revisions=REVISION=100`.
4. Run `curl https://api-665413986587.us-central1.run.app/version`, and read the commit.

Run `gcloud run services update-traffic api --region us-central1 --to-latest` after the fix merges.

CAUTION: the pin of step 3 holds until `--to-latest`. A deploy during the pin makes a revision that gets no traffic. The check of that deploy then fails, because `/version` names the old commit.

### 4.2 The web app

The Hosting console keeps each release. Open the Firebase console, then Hosting, then the release history. Select the menu of a good release, then Rollback. Each release names its commit in its message.

The REST API of Hosting does the same rollback with no console:

1. Set `S` to `https://firebasehosting.googleapis.com/v1beta1/sites/nk-workout-app-prod`.
2. Set `H` to the header `Authorization: Bearer $(gcloud auth print-access-token)`.
3. Set `Q` to the header `x-goog-user-project: nk-workout-app-prod`.
4. Run `curl -H "$H" -H "$Q" "$S/channels/live/releases?pageSize=3"`, and read each version and message.
5. Set `V` to the last part of the version name of the good release.
6. Run `curl -X POST -H "$H" -H "$Q" "$S/channels/live/releases?versionName=sites/nk-workout-app-prod/versions/$V"`.

The new release has the type `ROLLBACK`. In the drill of 2026-10-05, each release took less than 1 s (`docs/research/restore-drill.md`).

Then read `https://nk-workout-app-prod.web.app/version.json`. The service worker of the app shows "Update ready", and the owner applies it (D-133).

### 4.3 The Firestore rules

1. Run `git revert` on the commit that changed `firestore.rules`, in a pull request.
2. Merge the pull request. The trigger `deploy-rules` releases the rules.

The rules refuse each client call (D-77), and no client reads Firestore. So a rollback of the rules is for a wrong change only.

## 5. What a rollback does not repair

- A Firestore document that the new version wrote stays as it is.
- An image in Artifact Registry stays until you remove it.
- Section 6 restores the Firestore data.

## 6. Restore the Firestore data

Two copies of the data exist (D-124):

- Point-in-time recovery keeps each version of the data for 7 days. A read with a `readTime` in that window gives the old document.
- The daily backup keeps each backup for 10 days.

A restore of a backup writes a new database. It never writes over `(default)`. Do these steps:

1. Run `gcloud firestore backups list --project=nk-workout-app-prod --format="value(name,snapshotTime,state)"`.
2. Select the newest READY backup before the damage.
3. Set `BACKUP` to its full name, and `DEST` to a new id, for example `restore-20261001`.
4. Run `gcloud firestore databases restore --project=nk-workout-app-prod --source-backup="$BACKUP" --destination-database="$DEST"`.
5. Read the operation with `gcloud firestore operations describe` until it gives `SUCCESSFUL`.
6. Compare each damaged document in `$DEST` with `(default)`.
7. Copy each damaged document back into `(default)` with a script that the owner reads first.
8. Run `gcloud firestore databases update --project=nk-workout-app-prod --database="$DEST" --no-delete-protection`.
9. Run `gcloud firestore databases delete --project=nk-workout-app-prod --database="$DEST"` after the repair.

CAUTION: give `--database="$DEST"` to step 8. The same step on `(default)` removes its delete protection.

The restored database gets the delete protection of `(default)`. So without step 8, the delete of step 9 fails with `FAILED_PRECONDITION`. In the drill of 2026-10-05, the restore of the database of one user took 8 min 48 s.

For the count of step 6, walk the tree with the REST call `listCollectionIds`, and run a `count` aggregation for each collection. `docs/research/restore-drill.md` gives the counts of the drill.

Note: `gcloud` 533.0.0 has no `firestore databases clone` command in its GA group. The clone of a database at a past time needs another track or the console (unverified).

## 7. Refresh the image digests

Each build file and the Dockerfile name each image with its digest. The test `docs/tools/test_deploy_config.py` refuses an image with no digest. The Go image follows `go/go.mod`, and the Node image follows `web/.nvmrc`. Change the tag and the digest together, and read the digest from the registry.

## 8. Sources and dates

| Fact | Date read | Source |
|---|---|---|
| A 2nd-gen trigger needs `--service-account` and `--region`. | 2026-09-29 | `decktome:docs/deploy-and-rollback.md` |
| A budget does not stop spend. | 2026-09-28 | PC-82 of `docs/research/platform-cloud-and-ai.md` |
| The spend cap is Preview, the console alone sets it, and it pauses Cloud Run at 100%. | 2026-09-29 | Google Cloud, "Spend cap budgets", updated 2026-09-24 |
| The service `api` answers on `https://api-665413986587.us-central1.run.app`. | 2026-09-29 | `gcloud run deploy` |
| A restored database has delete protection, and its delete fails with `FAILED_PRECONDITION` until step 8 of section 6. | 2026-10-05 | `docs/research/restore-drill.md` |
| A release of an earlier version through the Hosting REST API has the type `ROLLBACK`. | 2026-10-05 | `docs/research/restore-drill.md` |
