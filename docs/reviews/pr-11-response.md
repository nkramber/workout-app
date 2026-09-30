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

## P2-1, round 2: The deploy guard leaves a check-to-deploy race

Review round 2 recorded effective head `5f9d991`, with the verdict "Changes required".

**Result: full merit.**

The trigger reproduces at `5f9d991`. The guard read `main` in one step, and the deploy ran in the next step. So a newer build had a gap in which it passed its guard and deployed. The older build then deployed its commit. The response of round 1 named this gap as unlikely, and an unlikely gap still breaks the contract.

**Correction.**

- The owner chose a lock for each part in a Cloud Storage bucket (Q-160, D-143). The session made the bucket `nk-workout-app-prod-deploy-lock` with the approval of the owner. It gave each deployer account `roles/storage.objectUser` on that bucket alone.
- `docs/tools/deploy_order.py`: the command `deploy` makes one critical section. It gets the lock of the part, reads `main`, runs the deploy command or skips, and then removes the lock. Cloud Storage creates the lock object with `ifGenerationMatch=0` for one caller alone. A lock older than 20 minutes is stale, and the deploy command stops after 10 minutes.
- `cloudbuild/api.yaml`, `cloudbuild/web.yaml`, `cloudbuild/rules.yaml`: the separate step `guard` is gone. Each deploy command runs after `--` of `deploy_order.py deploy`, so no step deploys outside the lock.
- `docs/deploy-and-rollback.md` and `docs/setup-gcp.md` give the lock and the bucket.

**Regression checks.**

- `TwoBuilds.test_an_older_build_paused_after_its_guard_can_not_move_the_part_back` in `docs/tools/test_deploy_order.py` does the check of the review. The older build pauses inside its critical section. A newer change merges, and its build starts. The newer build waits for the lock, and it deploys after the older build. The last deploy is the newer commit.
- The same test fails with the lock turned off. The control replaced `acquire` with a function that returns at once.
- The lock tests prove the release after a failed deploy and the wait limit. They also prove the removal of a stale lock, and one lock for each part.
- `docs/tools/test_deploy_config.py` proves that each deploy command runs inside `deploy_order.py deploy`, with the lock of its part and the paths of its trigger.
- A smoke test ran the `Bucket` class on the live bucket. The second create got no lock, and a delete with a wrong generation kept the lock. `deploy` removed its lock at the end.

## P2-2: A slow history read can outlive the stale-lock limit

Review round 3 recorded effective head `9ec3819`, with the verdict "Changes required". Codex marked P2-1 as fixed at that head.

**Result: full merit.**

The trigger reproduces at `9ec3819`. The clone of `read_main` had no timeout, and only the deploy command had a limit of 10 minutes. So a slow read kept a live lock past the stale limit of 20 minutes. A newer build then removed the lock.

**Correction.**

- `docs/tools/deploy_order.py`: the read of `main` has one deadline, `READ_TIMEOUT` of 5 minutes. The clone and each git call get the rest of that deadline. A read that ends after the deadline makes the build stop with no deploy, and the lock goes back. The deploy command keeps `DEPLOY_TIMEOUT` of 10 minutes.
- So a live build holds its lock for `HOLD` of 15 minutes at most, and `STALE` is 20 minutes. The 5 minutes between the two limits give a deploy on the server side time to end after its command stops.
- The three deploy steps get a step timeout of 2400 s. The wait and the hold take 30 minutes at most. So Cloud Build does not stop a build that holds the lock. The config test found this gap with the old timeout of 1800 s.
- `docs/deploy-and-rollback.md` gives the two limits.

**Regression checks.**

- `Bounds.test_a_read_past_its_limit_deploys_nothing_and_frees_the_lock` holds the live lock with a read past its limit. The build does not deploy, and the next build gets the lock and deploys. The test fails with the check after the read removed.
- `Bounds.test_read_main_stops_at_its_timeout` records each git process of the read. The clone, `merge-base`, and `rev-list` each get a timeout inside the deadline. The test fails with the timeout of the clone removed.
- `Bounds.test_a_timed_out_read_frees_the_lock` and `Bounds.test_the_read_and_the_deploy_get_their_limits` prove the release after a timeout and the two limits.
- `docs/tools/test_deploy_config.py` proves that `HOLD` is less than `STALE`, and that the wait and the hold end before the step timeout.
