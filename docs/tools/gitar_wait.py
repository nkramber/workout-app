#!/usr/bin/env python3
"""Wait for the Gitar review of the head of one pull request after a push (D-338).

  gitar_wait.py --pr NUMBER

`make gitar-wait PR=<n>` runs this file. The author session runs it in the
background at once after each push. It is a port of the wait script of
what-you-carry (what-you-carry:D-575), and it reads the Gitar pass with the
same rule as `make codex-review` (D-336):

  1. It waits FIRST seconds, the push wait of the `gitar-review` skill.
  2. Every POLL seconds, it reads the Gitar check runs of the head and the
     issue comments of the pull request.
  3. It exits 0 when `codex_review.gitar_problems` gives no problem for the
     push of the head: each Gitar check run completed, the newest dashboard
     changed after the push and shows no review in progress, and each
     `Gitar review` request has its accept reply and a later dashboard.
  4. When no Gitar check run and no current dashboard exist at REQUEST
     seconds, it posts one `Gitar review` comment.
  5. At LIMIT seconds, it exits 1, and the session stops and tells the owner.

The wait does not read the review threads. The `gitar-review` skill answers
them, and `make codex-review` refuses an open thread.

Each outcome has its own exit code: 0 a current review, 1 a stop at the
limit or a failed read, 2 a usage error.
"""
import argparse
import json
import os
import sys
import time

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import codex_review as cr  # noqa: E402

FIRST, POLL, REQUEST, LIMIT = 60, 30, 360, 900
EXIT_DONE, EXIT_STOP, EXIT_USAGE = 0, 1, 2


def gitar_runs(run, slug, head):
    """The status of each Gitar check run of the head. A page holds 100 runs, so the read follows each page."""
    # The path comes first, so the message of a failed read names it.
    pages = json.loads(cr.must(run, ["gh", "api", f"repos/{slug}/commits/{head}/check-runs?per_page=100",
                                     "--paginate", "--slurp"], cr.refuse))
    return [{"status": r.get("status")} for page in pages for r in page.get("check_runs", [])
            if (r.get("app") or {}).get("slug") == cr.GITAR_SLUG]


def wait(run, number, sleep=time.sleep, clock=time.monotonic, out=print):
    """Give the exit code of one wait. Each read that fails raises cr.Stop."""
    start = clock()
    slug = cr.must(run, ["gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner"], cr.refuse).strip()
    head = cr.must(run, ["gh", "pr", "view", str(number), "--json", "headRefOid", "--jq", ".headRefOid"], cr.refuse).strip()
    requested = False
    sleep(FIRST)
    while True:
        pushed = cr.push_time(run, slug, [head])
        runs = gitar_runs(run, slug, head)
        problems = cr.gitar_problems(pushed, cr.issue_comments(run, slug, number), runs, [])
        elapsed = int(clock() - start)
        if not problems:
            out(f"gitar-wait: the Gitar review of {head[:7]} on PR #{number} is current after {elapsed} s. "
                "Read the dashboard and the threads with the `gitar-review` skill.")
            return EXIT_DONE
        stale = any("before the push" in p or "no Gitar dashboard" in p for p in problems)
        if not runs and stale and not requested and elapsed >= REQUEST:
            cr.must(run, ["gh", "pr", "comment", str(number), "--body", "Gitar review"], cr.refuse)
            requested = True
            out(f"gitar-wait: no Gitar check run on {head[:7]} after {elapsed} s. Posted one Gitar review comment on PR #{number}.")
        if elapsed >= LIMIT:
            out(f"gitar-wait: no current Gitar review of {head[:7]} on PR #{number} after {elapsed} s. "
                "Stop, and tell the owner (D-338).")
            for problem in problems:
                out(f"  {problem}")
            return EXIT_STOP
        sleep(POLL)


def main(argv=None, run=cr.sh, sleep=time.sleep, clock=time.monotonic):
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--pr", type=int, required=True, help="the number of the pull request")
    try:
        args = parser.parse_args(argv)
    except SystemExit:
        return EXIT_USAGE
    try:
        return wait(run, args.pr, sleep, clock)
    except cr.Stop as stop:
        print(f"gitar-wait: {stop}")
        return EXIT_STOP


if __name__ == "__main__":
    sys.exit(main())
