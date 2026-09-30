"""Skip a stale deploy (P2-1 of the review of PR 11).

Three Cloud Build triggers deploy the API, the web app, and the Firestore
rules (D-14, D-142). The builds run apart, so the build of an older merge
can reach its deploy step after the build of a newer merge. Each build runs
this script just before its deploy step. The script reads `main` and
answers one question: does a commit after the commit of this build change a
path that this trigger watches?

- No: this build holds the newest change of its part, and it deploys.
- Yes: the build of that newer commit deploys the part. This build writes
  the skip file, and its deploy step and its check step do nothing.

A newer commit that changes no watched path starts no build of this
trigger, so it never makes this build skip. The repository is public, so a
build reads the history of `main` with no credential.

Exit codes: 0 when this build deploys, 3 when a newer change exists, and 1
on an error. An error stops the build, and the live part stays as it is.

Run: python3 docs/tools/deploy_order.py --commit SHA --skip-file F -- PATH [PATH ...]
"""
import argparse
import os
import subprocess
import sys
import tempfile

REPO = "https://github.com/nkramber/workout-app.git"
NEWER = 3


def git(repo, *args):
    return subprocess.run(["git", "-C", repo, *args], capture_output=True, text=True, check=True).stdout


def newer_changes(repo, commit, main, paths):
    """Return the commits of `main` after `commit` that change one of the paths, newest first.

    `commit` must be on `main`, because the triggers read `main` alone.
    """
    if subprocess.run(["git", "-C", repo, "merge-base", "--is-ancestor", commit, main]).returncode != 0:
        raise ValueError(f"{commit} is not on {main}")
    out = git(repo, "rev-list", f"{commit}..{main}", "--", *paths)
    return out.split()


def clone(url, into):
    """Clone the history of `main` with no file content."""
    subprocess.run(["git", "clone", "--quiet", "--filter=blob:none", "--no-checkout",
                    "--single-branch", "--branch", "main", url, into], check=True)
    return "origin/main"


def main(argv):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--commit", required=True)
    parser.add_argument("--skip-file")
    parser.add_argument("--repo", default=REPO)
    parser.add_argument("paths", nargs="+")
    args = parser.parse_args(argv)
    try:
        with tempfile.TemporaryDirectory() as tmp:
            repo = os.path.join(tmp, "main")
            ref = clone(args.repo, repo)
            newer = newer_changes(repo, args.commit, ref, args.paths)
    except (subprocess.CalledProcessError, ValueError) as err:
        print(f"deploy_order: {err}")
        return 1
    if not newer:
        print(f"deploy_order: {args.commit} is the newest change of {' '.join(args.paths)} on main")
        return 0
    print(f"deploy_order: {newer[0]} on main changes {' '.join(args.paths)} after {args.commit}."
          " Its build deploys, so this build skips.")
    if args.skip_file:
        with open(args.skip_file, "w", encoding="utf-8") as handle:
            handle.write(newer[0] + "\n")
    return NEWER


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
