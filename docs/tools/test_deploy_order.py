"""Tests of the deploy order guard (P2-1 of the review of PR 11).

They make a local repository, so they need no network.
Run: python3 -m unittest discover -s docs/tools -p 'test_*.py'
"""
import importlib.util
import os
import subprocess
import tempfile
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


class Main(unittest.TestCase):
    def test_the_exit_code_and_the_skip_file(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Repo(os.path.join(tmp, "src"))
            first = repo.merge("go/main.go", "1")
            skip = os.path.join(tmp, "skip")
            args = ["--repo", repo.root, "--commit", first, "--skip-file", skip, *API]
            self.assertEqual(do.main(args), 0)
            self.assertFalse(os.path.exists(skip))
            second = repo.merge("go/main.go", "2")
            self.assertEqual(do.main(args), do.NEWER)
            with open(skip, encoding="utf-8") as handle:
                self.assertEqual(handle.read().strip(), second)

    def test_an_unknown_commit_stops_the_build(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Repo(os.path.join(tmp, "src"))
            repo.merge("go/main.go", "1")
            self.assertEqual(do.main(["--repo", repo.root, "--commit", "0" * 40, *API]), 1)


if __name__ == "__main__":
    unittest.main()
