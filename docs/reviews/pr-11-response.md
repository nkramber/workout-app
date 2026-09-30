# Pull request 11 - author response

Date: 2026-09-29. Review round 1 recorded effective head `e5f724b`, with the verdict "Changes required".

## P2-1: An older build can replace a newer deploy

**Result: full merit.**

The trigger reproduces at `e5f724b`. Each build deploys its own commit, and no step reads `main` before the deploy. The old text of `docs/deploy-and-rollback.md` gave the race as a known risk, and a known risk does not stop the older deploy.

**Correction.**

- `docs/tools/deploy_order.py`: a new guard. It reads the history of `main` from the public repository. It exits 3 and writes the skip file when a commit after the commit of the build changes a watched path. A newer commit that changes no watched path starts no build of the trigger, so it never makes a build skip.
- `cloudbuild/api.yaml`, `cloudbuild/web.yaml`, `cloudbuild/rules.yaml`: the step `guard` runs just before the deploy step, with `allowExitCodes: [3]`. Each later step does nothing when the skip file exists. After the wait of a check, a newer change of the watched paths lets the build pass, because the newer build deployed first.
- `docs/deploy-and-rollback.md`: section 1 gives the guard and the one case that stays open. A newer merge can come after the guard of an older build. The newer build must then do all its steps before the older build ends its deploy step. The newer build does more steps, so the case is very unlikely.

**Regression checks.**

- `docs/tools/test_deploy_order.py` models two builds that finish out of order in a local repository. The older build gets exit 3 and the skip file after a newer merge of its paths. A newer merge of other paths does not stop it. A commit outside `main` stops the build.
- `docs/tools/test_deploy_config.py` proves that the guard runs just before each deploy step. The guard watches the paths of its trigger in `docs/deploy-and-rollback.md`, and each later step reads the skip file.
- `python3 docs/tools/deploy_order.py --commit d9b192e go` against the real repository gave exit 3, and it named `6117952` as the newer change of `go/`.
- The `cloud-sdk` image of the builds holds `git` 2.47.3 and Python 3.13.5, read with `docker run`. The `allowExitCodes` field is in the Cloud Build v1 API of `gcloud` 533.0.0.
