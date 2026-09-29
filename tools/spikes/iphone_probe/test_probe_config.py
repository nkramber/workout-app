"""The configuration tests of the iPhone probe (D-113). They read the
files of the probe only, so they need no Node and no network. The browser
tests of `make probe` test the behavior."""
import json
import os
import re
import subprocess
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
SEMVER = re.compile(r"^\d+\.\d+\.\d+$")
EMAIL = re.compile(r"[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}")
# The browser tests make each account on the Auth emulator with a
# reserved domain (RFC 2606). No other address goes into the repository.
ALLOWED_EMAIL_DOMAINS = {"example.com"}


def load(name):
    with open(os.path.join(HERE, name), encoding="utf-8") as f:
        return json.load(f)


def tracked_files():
    """The files of the probe that Git tracks or would track."""
    out = subprocess.run(
        ["git", "ls-files", "--cached", "--others", "--exclude-standard", "."],
        cwd=HERE, capture_output=True, text=True, check=True,
    ).stdout
    return [line for line in out.splitlines() if line]


class FirebaseServicesTest(unittest.TestCase):
    """The project serves Firebase Hosting and Firebase Authentication only (D-99)."""

    def test_firebase_json_names_hosting_auth_and_the_emulators_only(self):
        config = load("firebase.json")
        self.assertEqual(set(config), {"hosting", "auth", "emulators"})

    def test_auth_enables_email_and_password_alone(self):
        # D-75: sign-in by email and password. `firebase deploy --only auth`
        # reads this block.
        self.assertEqual(load("firebase.json")["auth"], {"providers": {"emailPassword": True}})

    def test_the_emulators_are_auth_only(self):
        emulators = load("firebase.json")["emulators"]
        services = set(emulators) - {"ui", "singleProjectMode"}
        self.assertEqual(services, {"auth"})
        self.assertFalse(emulators["ui"]["enabled"])

    def test_hosting_serves_the_build(self):
        hosting = load("firebase.json")["hosting"]
        self.assertEqual(hosting["public"], "dist")

    def test_the_test_script_starts_the_auth_emulator_alone(self):
        script = load("package.json")["scripts"]["test:e2e"]
        self.assertIn("--only auth", script)
        self.assertIn("--project demo-", script)

    def test_the_app_imports_no_other_firebase_service(self):
        allowed = {"firebase/app", "firebase/auth"}
        found = set()
        for rel in tracked_files():
            if rel.startswith("src/") and rel.endswith((".ts", ".tsx")):
                with open(os.path.join(HERE, rel), encoding="utf-8") as f:
                    found |= set(re.findall(r"[\"'](firebase/[a-z-]+)[\"']", f.read()))
        self.assertTrue(found)
        self.assertLessEqual(found, allowed)


class PinsTest(unittest.TestCase):
    """Each package has an exact version, as in Decktome."""

    def test_each_dependency_is_pinned(self):
        pkg = load("package.json")
        for section in ("dependencies", "devDependencies"):
            for name, version in pkg[section].items():
                with self.subTest(package=name):
                    self.assertRegex(version, SEMVER)

    def test_the_lock_file_agrees_with_the_package_file(self):
        pkg = load("package.json")
        root = load("package-lock.json")["packages"][""]
        for section in ("dependencies", "devDependencies"):
            self.assertEqual(root.get(section, {}), pkg[section])

    def test_the_node_version_is_22(self):
        with open(os.path.join(HERE, ".nvmrc"), encoding="utf-8") as f:
            self.assertRegex(f.read().strip(), r"^22\.\d+\.\d+$")


class PrivacyTest(unittest.TestCase):
    """No account email goes into the repository (AGENTS.md, PR-5)."""

    def test_no_file_holds_a_real_email_address(self):
        for rel in tracked_files():
            if rel.endswith((".png", "package-lock.json")):
                continue
            with open(os.path.join(HERE, rel), encoding="utf-8") as f:
                for address in EMAIL.findall(f.read()):
                    with self.subTest(file=rel):
                        self.assertIn(address.split("@", 1)[1], ALLOWED_EMAIL_DOMAINS)

    def test_the_probe_has_no_camera_page(self):
        # D-112: the probe tests only the device items of the app.
        for rel in tracked_files():
            if rel.startswith("src/"):
                with open(os.path.join(HERE, rel), encoding="utf-8") as f:
                    self.assertNotIn("getUserMedia", f.read(), rel)


if __name__ == "__main__":
    unittest.main()
