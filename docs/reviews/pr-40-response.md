# Pull request 40 response

Date: 2026-10-05

This file answers the findings of round 1 of `docs/reviews/pr-40.md`, at `97a8d40a65d5d77dc6be2db7af8cd8b32e9f7d3f`.

## P2-1: The restore report gives the wrong collection-path total

Result: partial merit.

Evidence: the two totals count different sets, and each total is correct. The count of the restored database gave 9 collection paths, and the count of `(default)` gave 10. The table of section 3.2 lists the 10 paths of both databases. The path `users/{uid}/history` is in `(default)` alone, because `DeleteHistory` wrote its fence document after the backup. But the summary row did not tell the reader why the totals differ, so the report gave two totals with no reason.

Correction: the summary row of `docs/research/restore-drill.md` now gives 10 paths in `(default)` and 9 in the restore. It names `users/{uid}/history` as the path in `(default)` alone.

Regression check: the table of section 3.2 has 10 rows. The output of the count script has 10 lines for `(default)` and 9 lines for the restore. The summary row now gives the same totals. `make verify` passed.

## Round 2

This part answers the findings of round 2, at `a019d5caa7b7711fe327be94afaa6e62235a458b`.

### P2-2: The restore procedure omits the operation name

Result: full merit.

Evidence: the trigger reproduced. Step 5 gave `gcloud firestore operations describe` with no operation name, and the command needs the name. In the drill, the session gave the full name of the restore operation to that command.

Correction: step 4 of section 6 of `docs/deploy-and-rollback.md` now prints the operation name with `--format="value(name)"`. Step 5 sets `OP` to that name, and step 6 gives `"$OP"` to `gcloud firestore operations describe`. The later steps moved one number, and each reference to them in `docs/deploy-and-rollback.md` and `docs/research/restore-drill.md` agrees.

Regression check: the drill read `name.basename()` of the restore response as the operation id, so `name` is the full operation name. With that full name, `gcloud firestore operations describe` gave `True SUCCESSFUL` at 19:29:13Z. `make verify` passed.
