"""Tests of ruleset_check.py (D-6, D-13, D-14)."""
import copy
import importlib.util
import json
import os
import re
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
spec = importlib.util.spec_from_file_location("ruleset_check", os.path.join(HERE, "ruleset_check.py"))
rc = importlib.util.module_from_spec(spec)
spec.loader.exec_module(rc)

ROOT = os.path.abspath(os.path.join(HERE, "..", ".."))
with open(os.path.join(ROOT, rc.RULESET), encoding="utf-8") as handle:
    FILE = json.load(handle)
with open(os.path.join(ROOT, rc.SETTINGS), encoding="utf-8") as handle:
    SETTINGS = json.load(handle)


def live():
    """The file as GitHub gives it: other key order, extra fields, and new defaults."""
    body = copy.deepcopy(FILE)
    body.update({"id": 42, "source": "owner/gym-route", "_links": {}})
    for rule in body["rules"]:
        if rule["type"] == "required_status_checks":
            rule["parameters"]["required_status_checks"].reverse()
        if rule["type"] == "pull_request":
            rule["parameters"]["required_reviewers"] = []
    return body


def rule(body, kind):
    return next(r for r in body["rules"] if r["type"] == kind)


def job_names(workflow):
    with open(os.path.join(ROOT, ".github/workflows", workflow), encoding="utf-8") as handle:
        return set(re.findall(r'^\s+name:\s*"?([a-z:-]+)"?\s*$', handle.read(), flags=re.M))


class TheFile(unittest.TestCase):
    def contexts(self):
        return {c["context"] for c in rule(FILE, "required_status_checks")["parameters"]["required_status_checks"]}

    # The job of the iPhone probe runs on each pull request, but it is not
    # a required check (D-114).
    OPTIONAL = {"verify:probe"}

    def test_the_file_requires_the_gate_the_contract_and_each_verify_job(self):
        verify = {n for n in job_names("verify.yml") if n.startswith("verify:")} - self.OPTIONAL
        # The product jobs of work area 2.1 are required checks too (D-127).
        self.assertEqual(verify, {"verify:lint", "verify:test", "verify:contract", "verify:go", "verify:emulator"})
        self.assertIn("pr-contract", job_names("pr-contract.yml"))
        self.assertIn("review-gate", job_names("review-gate.yml"))
        self.assertEqual(self.contexts(), verify | {"review-gate", "pr-contract"})

    def test_the_probe_job_runs_and_is_no_required_check(self):
        self.assertIn("verify:probe", job_names("verify.yml"))
        self.assertTrue(self.OPTIONAL.isdisjoint(self.contexts()))

    def test_the_ruleset_check_itself_is_no_required_check(self):
        # The live ruleset does not exist until the owner applies it, so CI can not run the check.
        self.assertNotIn("ruleset-check", self.contexts())

    def test_every_check_comes_from_the_actions_app(self):
        checks = rule(FILE, "required_status_checks")["parameters"]["required_status_checks"]
        self.assertEqual({c["integration_id"] for c in checks}, {15368})

    def test_the_checks_are_strict(self):
        self.assertTrue(rule(FILE, "required_status_checks")["parameters"]["strict_required_status_checks_policy"])

    def test_no_bypass_no_human_approval_and_squash_alone(self):
        self.assertEqual(FILE["bypass_actors"], [])
        params = rule(FILE, "pull_request")["parameters"]
        self.assertEqual(params["required_approving_review_count"], 0)
        self.assertEqual(params["allowed_merge_methods"], ["squash"])
        self.assertTrue(params["required_review_thread_resolution"])
        self.assertEqual(SETTINGS, {"allow_auto_merge": True, "allow_squash_merge": True, "allow_rebase_merge": False,
                                    "allow_merge_commit": False, "delete_branch_on_merge": True})

    def test_main_refuses_a_force_push_and_a_deletion(self):
        # D-14: no direct push to main.
        self.assertTrue(rule(FILE, "non_fast_forward"))
        self.assertTrue(rule(FILE, "deletion"))
        self.assertEqual(FILE["conditions"]["ref_name"]["include"], ["~DEFAULT_BRANCH"])


class Compare(unittest.TestCase):
    def test_the_live_form_of_the_file_matches(self):
        self.assertEqual(rc.compare(FILE, live()), [])

    def test_a_missing_check_is_a_difference(self):
        body = live()
        rule(body, "required_status_checks")["parameters"]["required_status_checks"].pop()
        self.assertEqual(len(rc.compare(FILE, body)), 1)

    def test_an_extra_live_check_is_a_difference(self):
        body = live()
        rule(body, "required_status_checks")["parameters"]["required_status_checks"].append({"context": "x", "integration_id": 1})
        self.assertIn("the file does not", rc.compare(FILE, body)[0])

    def test_a_bypass_actor_is_a_difference(self):
        body = live()
        body["bypass_actors"] = [{"actor_id": 5, "actor_type": "RepositoryRole", "bypass_mode": "always"}]
        self.assertTrue(rc.compare(FILE, body))

    def test_an_extra_live_rule_is_a_difference(self):
        body = live()
        body["rules"].append({"type": "required_signatures"})
        self.assertIn("does not name", rc.compare(FILE, body)[0])

    def test_a_missing_live_rule_is_a_difference(self):
        body = live()
        body["rules"] = [r for r in body["rules"] if r["type"] != "non_fast_forward"]
        self.assertIn("GitHub has no `non_fast_forward` rule", rc.compare(FILE, body)[0])

    def test_a_disabled_ruleset_is_a_difference(self):
        body = live()
        body["enforcement"] = "disabled"
        self.assertTrue(rc.compare(FILE, body))

    def test_a_rebase_merge_is_a_difference(self):
        repo = dict(SETTINGS, allow_rebase_merge=True, private=False)
        self.assertEqual(len(rc.compare(SETTINGS, repo, "repository")), 1)


class FindRuleset(unittest.TestCase):
    def test_the_ruleset_is_found_by_name(self):
        listing = [{"id": 7, "name": "other", "target": "branch"}, {"id": 9, "name": "review-gate", "target": "branch"}]
        self.assertEqual(rc.find_ruleset(listing, "review-gate"), 9)

    def test_no_ruleset_gives_none(self):
        self.assertIsNone(rc.find_ruleset([], "review-gate"))
        self.assertIsNone(rc.find_ruleset([{"id": 3, "name": "review-gate", "target": "tag"}], "review-gate"))

    def test_two_rulesets_of_one_name_stop(self):
        listing = [{"id": 1, "name": "review-gate", "target": "branch"}, {"id": 2, "name": "review-gate", "target": "branch"}]
        with self.assertRaises(SystemExit):
            rc.find_ruleset(listing, "review-gate")


if __name__ == "__main__":
    unittest.main()
