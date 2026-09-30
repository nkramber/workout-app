"""Deploy one part of the app in order (P2-1 of the review of PR 11).

Three Cloud Build triggers deploy the API, the web app, and the Firestore
rules (D-14, D-142). The builds run apart, so the build of an older merge
can reach its deploy after the build of a newer merge. So each build
deploys its part with the command `deploy`, which makes one critical
section of three steps:

1. Get the lock of the part: the object `<part>.lock` in the lock bucket.
   Cloud Storage creates an object with `ifGenerationMatch=0` for one
   caller alone, so two builds can not hold the lock together.
2. Read `main`. When a commit after the commit of this build changes a
   path of the trigger, the build of that commit deploys the part. This
   build then writes the skip file and deploys nothing.
3. Run the deploy command, then remove the lock.

A newer commit that changes no watched path starts no build of the
trigger, so it never makes a build skip. The repository is public, so a
build reads the history of `main` with no credential.

A build that stops while it holds the lock leaves the object. A lock older
than STALE is stale, and the next build removes it. So a live build must
end its critical section before STALE (P2-2 of the review of PR 11):

- The read of `main` stops after READ_TIMEOUT, from the clone to the last
  git call. A read that ends later makes the build stop with no deploy.
- The deploy command stops after DEPLOY_TIMEOUT.
- So a live build holds the lock for HOLD at most, and HOLD is less than
  STALE. The difference is time for a deploy on the server side to end
  after its command stops.

The bucket also deletes each object after one day.

The command `newer` reads `main` alone, with no lock. The check step of a
build uses it after its wait.

Exit codes: 0 when the build deployed, or when `newer` finds no newer
change. 3 when a newer change exists. 1 on an error, and the live part
then stays as it is. `deploy` returns the exit code of a failed deploy
command.

Run: python3 docs/tools/deploy_order.py deploy --part api --commit SHA \\
       --bucket B --skip-file F --paths go docker -- gcloud run deploy ...
     python3 docs/tools/deploy_order.py newer --commit SHA PATH [PATH ...]
"""
import argparse
import datetime
import json
import os
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.parse
import urllib.request

REPO = "https://github.com/nkramber/workout-app.git"
NEWER = 3
WAIT = 15 * 60
POLL = 10
READ_TIMEOUT = 5 * 60
DEPLOY_TIMEOUT = 10 * 60
HOLD = READ_TIMEOUT + DEPLOY_TIMEOUT
STALE = 20 * 60


class LockTimeout(Exception):
    pass


def left(deadline):
    """Return the seconds before the deadline, or None for no deadline."""
    if deadline is None:
        return None
    rest = deadline - time.monotonic()
    if rest <= 0:
        raise subprocess.TimeoutExpired("git", 0)
    return rest


def git(repo, *args, deadline=None):
    return subprocess.run(["git", "-C", repo, *args], capture_output=True, text=True, check=True,
                          timeout=left(deadline)).stdout


def newer_changes(repo, commit, main, paths, deadline=None):
    """Return the commits of `main` after `commit` that change one of the paths, newest first.

    `commit` must be on `main`, because the triggers read `main` alone.
    """
    ancestor = subprocess.run(["git", "-C", repo, "merge-base", "--is-ancestor", commit, main],
                              timeout=left(deadline))
    if ancestor.returncode != 0:
        raise ValueError(f"{commit} is not on {main}")
    return git(repo, "rev-list", f"{commit}..{main}", "--", *paths, deadline=deadline).split()


def read_main(url, commit, paths, timeout=None):
    """Clone the history of `main` with no file content, and return the newer changes.

    With a timeout, the clone and each git call end before it, or the read
    raises TimeoutExpired.
    """
    deadline = None if timeout is None else time.monotonic() + timeout
    with tempfile.TemporaryDirectory() as tmp:
        repo = os.path.join(tmp, "main")
        subprocess.run(["git", "clone", "--quiet", "--filter=blob:none", "--no-checkout",
                        "--single-branch", "--branch", "main", url, repo], check=True,
                       timeout=left(deadline))
        return newer_changes(repo, commit, "origin/main", paths, deadline=deadline)


class Bucket:
    """The lock objects in Cloud Storage, through the JSON API.

    The token comes from the metadata server of the build, so no key of a
    person or of an account is in the repository.
    """

    API = "https://storage.googleapis.com"
    TOKEN = "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token"

    def __init__(self, name):
        self.name = name

    def _token(self):
        request = urllib.request.Request(self.TOKEN, headers={"Metadata-Flavor": "Google"})
        with urllib.request.urlopen(request, timeout=10) as answer:
            return json.load(answer)["access_token"]

    def _call(self, method, url, body=None):
        request = urllib.request.Request(url, data=body, method=method,
                                         headers={"Authorization": f"Bearer {self._token()}"})
        try:
            with urllib.request.urlopen(request, timeout=30) as answer:
                return answer.status, answer.read()
        except urllib.error.HTTPError as err:
            return err.code, b""

    def _object(self, name):
        return f"{self.API}/storage/v1/b/{self.name}/o/{urllib.parse.quote(name, safe='')}"

    def create(self, name, body):
        """Create the object when no object of that name exists. Return its generation, or None."""
        query = urllib.parse.urlencode({"uploadType": "media", "name": name, "ifGenerationMatch": "0"})
        status, raw = self._call("POST", f"{self.API}/upload/storage/v1/b/{self.name}/o?{query}", body.encode())
        if status == 412:
            return None
        if status != 200:
            raise RuntimeError(f"lock create: HTTP {status}")
        return json.loads(raw)["generation"]

    def get(self, name):
        """Return (generation, age in seconds, holder), or None when no object exists."""
        status, raw = self._call("GET", self._object(name))
        if status == 404:
            return None
        if status != 200:
            raise RuntimeError(f"lock read: HTTP {status}")
        meta = json.loads(raw)
        created = datetime.datetime.fromisoformat(meta["timeCreated"].replace("Z", "+00:00"))
        age = (datetime.datetime.now(datetime.timezone.utc) - created).total_seconds()
        status, body = self._call("GET", f"{self._object(name)}?alt=media")
        holder = body.decode(errors="replace").strip() if status == 200 else ""
        return meta["generation"], age, holder

    def delete(self, name, generation):
        """Remove the object of this generation alone. Another generation stays."""
        status, _ = self._call("DELETE", f"{self._object(name)}?ifGenerationMatch={generation}")
        if status not in (204, 404, 412):
            raise RuntimeError(f"lock delete: HTTP {status}")


def acquire(store, name, holder, clock=time.monotonic, sleep=time.sleep, wait=WAIT, stale=STALE, log=print):
    """Get the lock, and return its generation. Wait while a live build holds it."""
    end = clock() + wait
    while True:
        generation = store.create(name, holder)
        if generation is not None:
            return generation
        current = store.get(name)
        if current is None:
            continue
        held, age, by = current
        if age >= stale:
            log(f"deploy_order: remove the stale lock {name} of {by or 'an unknown build'}, {int(age)} s old")
            store.delete(name, held)
            continue
        if clock() >= end:
            raise LockTimeout(f"{name} is held by {by or 'an unknown build'} for {int(age)} s")
        log(f"deploy_order: {name} is held by {by or 'an unknown build'}. Wait {POLL} s.")
        sleep(POLL)


def deploy(store, part, commit, paths, command, newer=None, skip_file=None, log=print,
           clock=time.monotonic, **lock):
    """Hold the lock of the part, read main, and run the deploy command. Return the exit code.

    `newer(timeout)` reads main, and `command(timeout)` deploys. Each one
    gets its limit, so the lock stays for HOLD at most.
    """
    newer = newer or (lambda timeout: read_main(REPO, commit, paths, timeout))
    name = f"{part}.lock"
    generation = acquire(store, name, commit, log=log, clock=clock, **lock)
    start = clock()
    try:
        found = newer(READ_TIMEOUT)
        if clock() - start > READ_TIMEOUT:
            raise LockTimeout(f"the read of main took more than {READ_TIMEOUT} s, so this build does not deploy")
        if found:
            log(f"deploy_order: {found[0]} on main changes {' '.join(paths)} after {commit}."
                " Its build deploys, so this build skips.")
            if skip_file:
                with open(skip_file, "w", encoding="utf-8") as handle:
                    handle.write(found[0] + "\n")
            return 0
        log(f"deploy_order: {commit} is the newest change of {' '.join(paths)}. Deploy it.")
        return command(DEPLOY_TIMEOUT)
    finally:
        store.delete(name, generation)


def run_command(argv):
    def command(timeout):
        return subprocess.run(argv, timeout=timeout).returncode
    return command


def main(argv):
    command = []
    if "--" in argv:
        at = argv.index("--")
        argv, command = argv[:at], argv[at + 1:]
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    sub = parser.add_subparsers(dest="action", required=True)
    one = sub.add_parser("newer")
    one.add_argument("--commit", required=True)
    one.add_argument("--repo", default=REPO)
    one.add_argument("paths", nargs="+")
    two = sub.add_parser("deploy")
    two.add_argument("--part", required=True, choices=["api", "web", "rules"])
    two.add_argument("--commit", required=True)
    two.add_argument("--bucket", required=True)
    two.add_argument("--skip-file", required=True)
    two.add_argument("--paths", nargs="+", required=True)
    args = parser.parse_args(argv)
    try:
        if args.action == "newer":
            found = read_main(args.repo, args.commit, args.paths)
            if found:
                print(f"deploy_order: {found[0]} on main changes {' '.join(args.paths)} after {args.commit}")
                return NEWER
            print(f"deploy_order: {args.commit} is the newest change of {' '.join(args.paths)} on main")
            return 0
        if not command:
            parser.error("deploy needs the deploy command after --")
        return deploy(Bucket(args.bucket), args.part, args.commit, args.paths, run_command(command),
                      skip_file=args.skip_file)
    except (subprocess.CalledProcessError, subprocess.TimeoutExpired, ValueError,
            RuntimeError, LockTimeout, OSError) as err:
        print(f"deploy_order: {err}")
        return 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
