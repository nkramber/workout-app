"""Tests of the deploy order guard (P2-1 of the review of PR 11).

They make a local repository, so they need no network.
Run: python3 -m unittest discover -s docs/tools -p 'test_*.py'
"""
import importlib.util
import os
import subprocess
import tempfile
import threading
import time
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
spec = importlib.util.spec_from_file_location("deploy_order", os.path.join(HERE, "deploy_order.py"))
do = importlib.util.module_from_spec(spec)
spec.loader.exec_module(do)

API = ["go", "docker", "cloudbuild/api.yaml"]


class Repo:
    """A repository with a branch main, one commit for each merge."""

    def __init__(self, root):
        self.root = root
        os.makedirs(root, exist_ok=True)
        self.git("init", "--quiet", "--initial-branch=main")
        self.git("config", "user.email", "test@example.com")
        self.git("config", "user.name", "test")

    def git(self, *args):
        return subprocess.run(["git", "-C", self.root, *args], capture_output=True, text=True, check=True).stdout.strip()

    def merge(self, path, text):
        full = os.path.join(self.root, path)
        os.makedirs(os.path.dirname(full), exist_ok=True)
        with open(full, "w", encoding="utf-8") as handle:
            handle.write(text)
        self.git("add", "-A")
        self.git("commit", "--quiet", "-m", f"change {path}")
        return self.git("rev-parse", "HEAD")


class NewerChanges(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.repo = Repo(self.tmp.name)
        self.first = self.repo.merge("go/main.go", "1")

    def tearDown(self):
        self.tmp.cleanup()

    def newer(self, commit):
        return do.newer_changes(self.repo.root, commit, "main", API)

    def test_the_newest_change_deploys(self):
        self.assertEqual(self.newer(self.first), [])

    def test_an_older_build_skips_after_a_newer_merge_of_its_paths(self):
        # Two merges change go/. The build of the newer merge deploys
        # first. The older build then reaches its deploy step, and it must
        # not move the live API back.
        second = self.repo.merge("go/main.go", "2")
        self.assertEqual(self.newer(self.first), [second])
        self.assertEqual(self.newer(second), [])

    def test_a_newer_merge_of_other_paths_does_not_stop_the_deploy(self):
        # A newer merge of documents alone starts no API build. So the
        # older API build must still deploy.
        self.repo.merge("docs/note.md", "text")
        self.repo.merge("web/src/app.ts", "web")
        self.assertEqual(self.newer(self.first), [])

    def test_each_watched_path_counts(self):
        dockerfile = self.repo.merge("docker/api.Dockerfile", "FROM x")
        build_file = self.repo.merge("cloudbuild/api.yaml", "steps: []")
        self.assertEqual(self.newer(self.first), [build_file, dockerfile])

    def test_a_commit_outside_main_is_an_error(self):
        self.repo.git("switch", "--quiet", "-c", "feature")
        branch = self.repo.merge("go/main.go", "branch")
        self.repo.git("switch", "--quiet", "main")
        with self.assertRaises(ValueError):
            self.newer(branch)


class FakeBucket:
    """The lock objects of Cloud Storage in memory, with the same preconditions."""

    def __init__(self):
        self.objects = {}
        self.next = 1
        self.guard = threading.Lock()
        self.now = 0.0

    def create(self, name, body):
        with self.guard:
            if name in self.objects:
                return None
            self.next += 1
            self.objects[name] = (self.next, self.now, body)
            return self.next

    def get(self, name):
        with self.guard:
            if name not in self.objects:
                return None
            generation, created, body = self.objects[name]
            return generation, self.now - created, body

    def delete(self, name, generation):
        with self.guard:
            if name in self.objects and self.objects[name][0] == generation:
                del self.objects[name]


class Lock(unittest.TestCase):
    def test_a_free_lock_is_taken_and_given_back(self):
        store = FakeBucket()
        ran = []
        code = do.deploy(store, "api", "c1", API, lambda: ran.append("c1") or 0, newer=lambda: [], log=lambda _: None)
        self.assertEqual((code, ran, store.objects), (0, ["c1"], {}))

    def test_the_lock_goes_back_when_the_deploy_fails(self):
        store = FakeBucket()
        code = do.deploy(store, "api", "c1", API, lambda: 7, newer=lambda: [], log=lambda _: None)
        self.assertEqual((code, store.objects), (7, {}))

        def boom():
            raise ValueError("not on main")
        with self.assertRaises(ValueError):
            do.deploy(store, "api", "c1", API, lambda: 0, newer=boom, log=lambda _: None)
        self.assertEqual(store.objects, {})

    def test_a_held_lock_stops_the_deploy_after_the_wait(self):
        store = FakeBucket()
        store.create("web.lock", "c9")
        clock = iter(range(0, 10_000, 100))
        ran = []
        with self.assertRaises(do.LockTimeout):
            do.deploy(store, "web", "c1", ["web"], lambda: ran.append(1) or 0, newer=lambda: [],
                      log=lambda _: None, clock=lambda: next(clock), sleep=lambda _: None, wait=500)
        self.assertEqual(ran, [])
        self.assertIn("web.lock", store.objects)

    def test_a_stale_lock_is_removed(self):
        store = FakeBucket()
        store.create("rules.lock", "c9")
        store.now = do.STALE + 1
        code = do.deploy(store, "rules", "c1", ["firestore.rules"], lambda: 0, newer=lambda: [], log=lambda _: None)
        self.assertEqual((code, store.objects), (0, {}))

    def test_each_part_has_its_own_lock(self):
        store = FakeBucket()
        store.create("api.lock", "c9")
        code = do.deploy(store, "web", "c1", ["web"], lambda: 0, newer=lambda: [], log=lambda _: None)
        self.assertEqual(code, 0)
        self.assertIn("api.lock", store.objects)


class TwoBuilds(unittest.TestCase):
    """The regression check of P2-1: two builds of one trigger out of order."""

    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.repo = Repo(self.tmp.name)
        self.store = FakeBucket()
        self.live = []

    def tearDown(self):
        self.tmp.cleanup()

    def build(self, commit, command=None, skip=None):
        def deploy_now():
            self.live.append(commit)
            return 0
        return do.deploy(self.store, "api", commit, API, command or deploy_now,
                         newer=lambda: do.newer_changes(self.repo.root, commit, "main", API),
                         skip_file=skip, log=lambda _: None, sleep=lambda _: time.sleep(0.01))

    def test_an_older_build_paused_after_its_guard_can_not_move_the_part_back(self):
        # Codex: pause an older build after its guard. Merge and deploy a
        # newer change to the same part, then resume the older build.
        older = self.repo.merge("go/main.go", "1")
        paused = threading.Event()
        resume = threading.Event()
        results = {}

        def older_deploy():
            paused.set()
            resume.wait(5)
            self.live.append(older)
            return 0

        first = threading.Thread(target=lambda: results.update(older=self.build(older, older_deploy)))
        first.start()
        self.assertTrue(paused.wait(5))
        newer = self.repo.merge("go/main.go", "2")
        second = threading.Thread(target=lambda: results.update(newer=self.build(newer)))
        second.start()
        time.sleep(0.1)
        # The newer build waits for the lock, so it deployed nothing yet.
        self.assertEqual(self.live, [])
        resume.set()
        first.join(5)
        second.join(5)
        self.assertEqual(results, {"older": 0, "newer": 0})
        self.assertEqual(self.live, [older, newer])
        self.assertEqual(self.live[-1], newer)
        self.assertEqual(self.store.objects, {})

    def test_an_older_build_after_a_newer_deploy_skips(self):
        older = self.repo.merge("go/main.go", "1")
        newer = self.repo.merge("go/main.go", "2")
        self.assertEqual(self.build(newer), 0)
        skip = os.path.join(self.tmp.name, "skip")
        self.assertEqual(self.build(older, skip=skip), 0)
        self.assertEqual(self.live, [newer])
        with open(skip, encoding="utf-8") as handle:
            self.assertEqual(handle.read().strip(), newer)


class Main(unittest.TestCase):
    def test_newer_gives_the_exit_code(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Repo(os.path.join(tmp, "src"))
            first = repo.merge("go/main.go", "1")
            args = ["newer", "--repo", repo.root, "--commit", first, *API]
            self.assertEqual(do.main(args), 0)
            repo.merge("go/main.go", "2")
            self.assertEqual(do.main(args), do.NEWER)
            self.assertEqual(do.main(["newer", "--repo", repo.root, "--commit", "0" * 40, *API]), 1)

    def test_deploy_needs_a_command(self):
        with self.assertRaises(SystemExit):
            do.main(["deploy", "--part", "api", "--commit", "c", "--bucket", "b", "--skip-file", "f", "--paths", "go"])


if __name__ == "__main__":
    unittest.main()
