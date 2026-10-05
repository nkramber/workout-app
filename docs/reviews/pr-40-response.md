# Pull request 40 response

Date: 2026-10-05

This file answers the findings of round 1 of `docs/reviews/pr-40.md`, at `97a8d40a65d5d77dc6be2db7af8cd8b32e9f7d3f`.

## P2-1: The restore report gives the wrong collection-path total

Result: partial merit.

Evidence: the two totals count different sets, and each total is correct. The count of the restored database gave 9 collection paths, and the count of `(default)` gave 10. The table of section 3.2 lists the 10 paths of both databases. The path `users/{uid}/history` is in `(default)` alone, because `DeleteHistory` wrote its fence document after the backup. But the summary row did not tell the reader why the totals differ, so the report gave two totals with no reason.

Correction: the summary row of `docs/research/restore-drill.md` now gives 10 paths in `(default)` and 9 in the restore. It names `users/{uid}/history` as the path in `(default)` alone.

Regression check: the table of section 3.2 has 10 rows. The output of the count script has 10 lines for `(default)` and 9 lines for the restore. The summary row now gives the same totals. `make verify` passed.
