"""The configuration tests of the deploy of work area 2.3 (D-14, D-18).

They read the files of the repository only, so they need no network and no
cloud account. docs/deploy-and-rollback.md describes the deploy.
"""
import importlib.util
import json
import os
import re
import unittest

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
PROJECT = "nk-workout-app-prod"


def read(path):
    with open(os.path.join(ROOT, path), encoding="utf-8") as f:
        return f.read()


def build_images(text):
    return re.findall(r"^\s*(?:- )?name:\s*(\S+)", text, re.M)


class PinnedImagesTest(unittest.TestCase):
    """Each image names its digest, so a new release of an image changes no build."""

    def test_each_cloud_build_image_names_a_digest(self):
        for name in ("api", "web", "rules"):
            images = build_images(read(f"cloudbuild/{name}.yaml"))
            self.assertTrue(images, name)
            for image in images:
                self.assertRegex(image, r"@sha256:[0-9a-f]{64}$", f"cloudbuild/{name}.yaml")

    def test_each_dockerfile_image_names_a_digest(self):
        froms = re.findall(r"^FROM\s+(\S+)", read("docker/api.Dockerfile"), re.M)
        self.assertEqual(len(froms), 2)
        for image in froms:
            self.assertRegex(image, r"@sha256:[0-9a-f]{64}$")

    def test_the_go_image_is_the_version_of_go_mod(self):
        go = re.search(r"^go (\S+)$", read("go/go.mod"), re.M).group(1)
        self.assertIn(f"FROM golang:{go}@", read("docker/api.Dockerfile"))

    def test_the_node_image_is_the_version_of_web_nvmrc(self):
        node = read("web/.nvmrc").strip()
        for name in ("web", "rules"):
            for image in build_images(read(f"cloudbuild/{name}.yaml")):
                if image.startswith("node:"):
                    self.assertTrue(image.startswith(f"node:{node}@"), image)

    def test_firebase_tools_is_the_version_of_emulators(self):
        pinned = json.loads(read("emulators/package.json"))["devDependencies"]["firebase-tools"]
        for name in ("web", "rules"):
            used = re.findall(r"firebase-tools@(\S+)", read(f"cloudbuild/{name}.yaml"))
            self.assertEqual(set(used), {pinned}, name)


class SeparateAccountsTest(unittest.TestCase):
    """Each build deploys one part, so each trigger account holds one deploy role."""

    def test_the_web_build_releases_hosting_alone(self):
        text = read("cloudbuild/web.yaml")
        self.assertIn("deploy --only hosting ", text)
        self.assertNotIn("firestore", text.split("steps:", 1)[1])
        self.assertNotIn("run deploy", text)

    def test_the_rules_build_releases_the_rules_alone(self):
        text = read("cloudbuild/rules.yaml")
        self.assertIn("deploy --only firestore:rules ", text)
        self.assertNotIn("npm ci", text)
        self.assertNotIn("--prefix web", text)

    def test_the_api_build_deploys_the_api_alone(self):
        text = read("cloudbuild/api.yaml")
        self.assertIn("      - gcloud\n      - run\n      - deploy\n      - api\n", text)
        self.assertEqual(text.count("      - deploy\n"), 2)
        self.assertNotIn("firebase", text)
        # The deploy names the image alone, so the account, the
        # environment, and the scale of the service carry over.
        for flag in ("--service-account", "--set-env-vars", "--update-env-vars", "--allow-unauthenticated"):
            self.assertNotIn(flag, text)

    def test_the_api_build_gives_the_commit_to_the_version_route(self):
        self.assertIn("--build-arg=COMMIT=$COMMIT_SHA", read("cloudbuild/api.yaml"))
        self.assertIn('-X main.commit=${COMMIT}', read("docker/api.Dockerfile"))
        self.assertIn('var commit = "unknown"', read("go/cmd/api/main.go"))


class ProjectTest(unittest.TestCase):
    """The web build calls the API of the project, and Hosting serves web/dist."""

    def test_the_api_url_is_the_service_api_of_the_project(self):
        number = re.search(r"appId: \"1:(\d+):web:", read("web/src/lib/firebase-config.ts")).group(1)
        self.assertIn(f"_API_BASE_URL: https://api-{number}.us-central1.run.app\n", read("cloudbuild/web.yaml"))
        self.assertIn(f'projectId: "{PROJECT}"', read("web/src/lib/firebase-config.ts"))

    def test_hosting_serves_the_web_build_with_no_cache_on_the_shell(self):
        hosting = json.loads(read("firebase.json"))["hosting"]
        self.assertEqual(hosting["public"], "web/dist")
        self.assertEqual(hosting["rewrites"], [{"source": "**", "destination": "/index.html"}])
        # D-133: sw.js, index.html, and the manifest are not cached.
        everything = [h for h in hosting["headers"] if h["source"] == "**"][0]["headers"]
        self.assertIn({"key": "Cache-Control", "value": "no-cache"}, everything)

    def test_firestore_names_the_deny_all_rules(self):
        firestore = json.loads(read("firebase.json"))["firestore"]
        self.assertEqual(firestore, {"database": "(default)", "rules": "firestore.rules"})
        rules = read("firestore.rules")
        # D-77: one match of every document, and it refuses each call.
        self.assertEqual(re.findall(r"allow [^;]*;", rules), ["allow read, write: if false;"])
        self.assertIn("match /{document=**}", rules)


WATCHED = {
    "api": ["go", "docker", "cloudbuild/api.yaml"],
    "web": ["web", "firebase.json", "cloudbuild/web.yaml"],
    "rules": ["firestore.rules", "cloudbuild/rules.yaml"],
}


def step_ids(text):
    return re.findall(r"^  - id: (\S+)$", text, re.M)


class DeployOrderTest(unittest.TestCase):
    """Each build deploys under the lock of its part (P2-1 of the review of PR 11)."""

    DEPLOY_STEP = {"api": "deploy", "web": "release", "rules": "release"}

    def deploy_call(self, name):
        text = read(f"cloudbuild/{name}.yaml")
        body = text.split(f"- id: {self.DEPLOY_STEP[name]}\n", 1)[1].split("\n  - id:", 1)[0]
        # The API step lists its arguments. The two npm steps give one
        # command line. Both forms give the same words.
        if "entrypoint: python3" in body:
            words = re.findall(r"^      - (\S+)$", body, re.M)
        else:
            words = body.replace("\\\n", " ").split()
        return body, words

    def test_each_deploy_runs_under_the_lock_of_its_part(self):
        for name in WATCHED:
            body, words = self.deploy_call(name)
            self.assertIn("docs/tools/deploy_order.py", words, name)
            at = words.index("docs/tools/deploy_order.py")
            self.assertEqual(words[at + 1], "deploy", name)
            self.assertIn(f"--part={name}", words, name)
            self.assertIn("--bucket=$PROJECT_ID-deploy-lock", words, name)
            self.assertIn("--skip-file=/workspace/.deploy-skip", words, name)
            # The deploy command comes after --, so the script runs it
            # inside the lock, and no other step deploys.
            command = " ".join(words[words.index("--") + 1:])
            deploy = {"api": "gcloud run deploy api", "web": "deploy --only hosting",
                      "rules": "deploy --only firestore:rules"}[name]
            self.assertIn(deploy, command, name)
            self.assertEqual(" ".join(words).count(deploy), 1, name)
            self.assertIn("timeout: 2400s", body, name)

    def test_the_lock_watches_the_paths_of_its_trigger(self):
        # docs/deploy-and-rollback.md gives the paths of each trigger. A
        # lock with fewer paths can let an old build deploy, and a lock
        # with more paths can skip a deploy that no newer build makes.
        table = read("docs/deploy-and-rollback.md")
        for name, paths in WATCHED.items():
            _, words = self.deploy_call(name)
            listed = words[words.index("--paths") + 1:words.index("--")]
            self.assertEqual(listed, paths, name)
            row = [line for line in table.splitlines() if line.startswith(f"| `deploy-{name}`")][0]
            globs = re.findall(r"`([^`]+)`", row.split("|")[2])
            self.assertEqual([g[:-3] if g.endswith("/**") else g for g in globs], paths, name)

    def test_each_check_reads_the_skip_file_and_main(self):
        for name in ("api", "web"):
            text = read(f"cloudbuild/{name}.yaml")
            body = text.split("- id: check\n", 1)[1]
            self.assertIn("if [ -f /workspace/.deploy-skip ]", body, name)
            self.assertIn(f"deploy_order.py newer --commit=$COMMIT_SHA {' '.join(WATCHED[name])} ", body, name)

    def test_the_lock_waits_less_than_the_step_timeout(self):
        spec = importlib.util.spec_from_file_location("deploy_order", os.path.join(ROOT, "docs/tools/deploy_order.py"))
        order = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(order)
        # A live build holds the lock for HOLD at most, and HOLD is less
        # than STALE, so no build removes the lock of a live build. The
        # wait and the hold end before the step timeout of 2400 s.
        self.assertLess(order.HOLD, order.STALE)
        self.assertLess(order.WAIT + order.HOLD + 120, 2400)


if __name__ == "__main__":
    unittest.main()
