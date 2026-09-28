"""Unit tests of the Luna plan spike. They use the fake provider and fake
HTTP only, so they make no paid call."""
import contextlib
import copy
import io
import json
import os
import tempfile
import unittest
import urllib.error
from unittest import mock

import harness
import policy
import prompt
import providers
import schema_check

HERE = os.path.dirname(os.path.abspath(__file__))
MACHINES = harness.load_json("machines.json")["machines"]
PROFILES = harness.load_json("profiles.json")["profiles"]
ROLES = harness.load_json("roles.json")
BY_ID = {p["profile_id"]: p for p in PROFILES}
D41_FIELDS = {"experience", "goals", "injuries_restrictions", "inventory", "age", "height_in", "weight_lb", "cardio_preference"}


def rule_ids(plan, profile):
    return {v["rule_id"] for v in policy.evaluate(plan, profile, MACHINES)}


class FixtureTest(unittest.TestCase):
    def test_twenty_profiles_with_the_fields_of_d41(self):
        self.assertEqual(len(PROFILES), 20)
        self.assertEqual(len(BY_ID), 20)
        for p in PROFILES:
            self.assertLessEqual(D41_FIELDS, set(p), p["profile_id"])

    def test_inventory_uses_the_catalog(self):
        catalog = {m["machine_id"]: m for m in MACHINES}
        for p in PROFILES:
            ids = [i["machine_id"] for i in p["inventory"]]
            self.assertEqual(len(ids), len(set(ids)), p["profile_id"])
            for i in p["inventory"]:
                self.assertIn(i["machine_id"], catalog)
                if catalog[i["machine_id"]]["kind"] == "cardio":
                    self.assertIsNone(i["load_estimate_lb"])
            self.assertLessEqual(set(p["excluded_machine_ids"]), set(ids), p["profile_id"])


class SchemaTest(unittest.TestCase):
    def setUp(self):
        self.schema = schema_check.load_schema()
        self.plan = providers.baseline_plan(BY_ID["SP-01"], MACHINES)

    def test_baseline_plans_pass(self):
        for p in PROFILES:
            plan, errors = schema_check.check_text(json.dumps(providers.baseline_plan(p, MACHINES)))
            self.assertEqual(errors, [], p["profile_id"])

    def test_errors(self):
        bad = copy.deepcopy(self.plan)
        del bad["guidance"]
        bad["extra"] = 1
        bad["sessions"][0]["day"] = "one"
        bad["sessions"][0]["cardio"]["intensity"] = "hard"
        bad["sessions"][0]["exercises"][0]["sets"][0]["reps"] = True
        errors = schema_check.validate(bad, self.schema)
        self.assertIn("$: missing guidance", errors)
        self.assertIn("$: extra property extra", errors)
        self.assertTrue(any("day: expected integer" in e for e in errors))
        self.assertTrue(any("'hard' is not one of" in e for e in errors))
        self.assertTrue(any("reps: expected integer" in e for e in errors))

    def test_integer_load_is_a_number(self):
        self.assertEqual(schema_check.validate(5, {"type": "number"}), [])

    def test_not_json(self):
        plan, errors = schema_check.check_text("Here is your plan")
        self.assertIsNone(plan)
        self.assertIn("not JSON", errors[0])

    def test_unknown_keyword_is_an_error(self):
        with self.assertRaises(ValueError):
            schema_check.validate(1, {"type": "integer", "minimum": 0})

    def test_strict_form(self):
        def walk(node):
            if node.get("type") == "object":
                self.assertIs(node["additionalProperties"], False)
                self.assertEqual(set(node["required"]), set(node["properties"]))
                for sub in node["properties"].values():
                    walk(sub)
            if node.get("type") == "array":
                walk(node["items"])
        walk(self.schema)


class PolicyTest(unittest.TestCase):
    def test_table_ids_versions_and_checks(self):
        ids = [r.rule_id for r in policy.RULES]
        self.assertEqual(len(ids), len(set(ids)))
        self.assertEqual(set(ids), set(policy.CHECKS))
        for r in policy.RULES:
            self.assertGreaterEqual(r.version, 1)
            self.assertTrue(r.sources)

    def test_open_question_rules_are_marked(self):
        marked = {r.rule_id: r.questions for r in policy.RULES if r.questions}
        self.assertEqual(marked, {"POL-004": ("Q-102", "Q-106"), "POL-007": ("Q-105",),
                                  "POL-008": ("Q-102", "Q-104"), "POL-009": ("Q-92", "Q-102"),
                                  "POL-011": ("Q-102",), "POL-012": ("Q-102",), "POL-015": ("Q-101",)})
        self.assertEqual(policy.RULES_BY_ID["POL-007"].status, "draft, no answer: Q-105")
        self.assertIn("| POL-009 | 1 |", policy.rules_table())

    def test_baseline_plans_obey_every_rule(self):
        for p in PROFILES:
            self.assertEqual(policy.evaluate(providers.baseline_plan(p, MACHINES), p, MACHINES), [], p["profile_id"])

    def test_each_fault_is_caught(self):
        cases = {
            ("SP-02", "excluded"): {"POL-001"},
            ("SP-03", "rir_zero"): {"POL-003"},
            ("SP-06", "over_bound"): {"POL-008"},
            ("SP-10", "off_step"): {"POL-005", "POL-007"},
            ("SP-12", "load_jump"): {"POL-009"},
            ("SP-13", "claim"): {"POL-015"},
            ("SP-15", "rest_short"): {"POL-013"},
            ("SP-18", "reps_high"): {"POL-010"},
            ("SP-20", "extra_sessions"): {"POL-012", "POL-014"},
        }
        for (pid, fault), expected in cases.items():
            p = BY_ID[pid]
            text = providers.apply_fault(fault, providers.baseline_plan(p, MACHINES), p, MACHINES)
            self.assertLessEqual(expected, rule_ids(json.loads(text), p), (pid, fault))

    def test_round5_tie_rounds_down(self):
        self.assertEqual(policy.round5(22.5), 20)
        self.assertEqual(policy.round5(22.6), 25)
        self.assertEqual(policy.round5(27), 25)
        self.assertEqual(policy.round5(28), 30)

    def test_on_stack(self):
        w = {"min": 25, "max": 200, "step": 12.5}
        self.assertTrue(policy.on_stack(37.5, w))
        self.assertTrue(policy.on_stack(50, w))
        self.assertFalse(policy.on_stack(45, w))
        self.assertFalse(policy.on_stack(212.5, w))

    def test_restart(self):
        self.assertTrue(policy.is_restart(BY_ID["SP-02"]))   # novice
        self.assertTrue(policy.is_restart(BY_ID["SP-10"]))   # 4 months
        self.assertFalse(policy.is_restart(BY_ID["SP-18"]))  # 2 months

    def test_start_bound(self):
        w = {"min": 10, "max": 250, "step": 10}
        self.assertEqual(policy.start_bound(BY_ID["SP-01"], w, 150), 120)   # restart, 80%
        self.assertEqual(policy.start_bound(BY_ID["SP-03"], w, 190), 190)   # no restart
        self.assertEqual(policy.start_bound(BY_ID["SP-03"], w, 300), 250)   # machine max
        self.assertEqual(policy.start_bound(BY_ID["SP-03"], w, 5), 10)      # machine min
        self.assertEqual(policy.start_bound(BY_ID["SP-03"], w, None), 30)   # min + 2 steps

    def plan_with_loads(self, pid, days_loads):
        p = BY_ID[pid]
        plan = providers.baseline_plan(p, MACHINES)
        plan["sessions"] = plan["sessions"][:len(days_loads)]
        for session, loads in zip(plan["sessions"], days_loads):
            ex = session["exercises"][0]
            ex["machine_id"] = "M04"
            ex["sets"] = [{"reps": 10, "load_lb": load, "rir_target": 3} for load in loads]
            session["exercises"] = [ex]
        return plan, p

    def test_load_jump(self):
        plan, p = self.plan_with_loads("SP-03", [[100, 110], [110, 110]])
        self.assertNotIn("POL-009", rule_ids(plan, p))          # +10% and +0
        plan, p = self.plan_with_loads("SP-03", [[20, 30]])
        self.assertIn("POL-009", rule_ids(plan, p))             # +10 lb on 20 lb
        plan, p = self.plan_with_loads("SP-03", [[100], [120]])
        self.assertIn("POL-009", rule_ids(plan, p))             # +20% across sessions
        plan, p = self.plan_with_loads("SP-01", [[40], [40]])
        self.assertNotIn("POL-009", rule_ids(plan, p))
        plan, p = self.plan_with_loads("SP-01", [[40], [50]])
        self.assertIn("POL-009", rule_ids(plan, p))             # restart: no rise

    def test_restart_rir(self):
        p = BY_ID["SP-01"]
        plan = providers.baseline_plan(p, MACHINES)
        plan["sessions"][0]["exercises"][0]["sets"][0]["rir_target"] = 2
        self.assertEqual(rule_ids(plan, p), {"POL-004"})

    def test_blocked_claims(self):
        for text in ("This helps treat pain.", "A rehab plan.", "It can lower your blood pressure.",
                     "No diagnosis here.", "Physical therapy work."):
            self.assertTrue(policy.BLOCKED_CLAIMS.search(text), text)
        for text in ("Walk on the treadmill.", "A healthy habit.", "Ask a qualified professional."):
            self.assertIsNone(policy.BLOCKED_CLAIMS.search(text), text)

    def test_complete_plan(self):
        p = BY_ID["SP-03"]
        plan = providers.baseline_plan(p, MACHINES)
        plan["sessions"][1]["day"] = plan["sessions"][0]["day"]
        plan["sessions"][2]["exercises"][0]["sets"] = []
        self.assertIn("POL-002", rule_ids(plan, p))
        plan["sessions"] = []
        self.assertIn("POL-002", rule_ids(plan, p))

    def test_sets_for_each_exercise(self):
        for pid, count, caught in (("SP-03", 3, False), ("SP-03", 4, True), ("SP-01", 3, True)):
            p = BY_ID[pid]
            plan = providers.baseline_plan(p, MACHINES)
            ex = plan["sessions"][0]["exercises"][0]
            ex["sets"] = [dict(ex["sets"][0]) for _ in range(count)]
            self.assertEqual("POL-011" in rule_ids(plan, p), caught, (pid, count))

    def test_cardio_block(self):
        p = BY_ID["SP-01"]
        plan = providers.baseline_plan(p, MACHINES)
        plan["sessions"][0]["cardio"] = {"machine_id": "", "minutes": 10, "intensity": "easy"}
        plan["sessions"][1]["cardio"] = {"machine_id": "M04", "minutes": 10, "intensity": "easy"}
        ids = rule_ids(plan, p)
        self.assertIn("POL-016", ids)
        self.assertIn("POL-001", ids)


class PromptTest(unittest.TestCase):
    def test_prompt_names_every_rule_and_the_boundary(self):
        text = prompt.instructions()
        for r in policy.RULES:
            self.assertIn(r.rule_id, text)
        self.assertIn("fitness guidance only", text)
        self.assertEqual(len(prompt.prompt_hash()), 16)

    def test_user_input_holds_no_profile_id(self):
        body = json.loads(prompt.user_input(BY_ID["SP-01"], MACHINES))
        self.assertNotIn("profile_id", body)
        self.assertEqual(body["inventory"][0]["available_weights_lb"]["step"], 10)

    def test_no_model_id_at_a_call_site(self):
        model = ROLES["roles"]["plan"]["model"]
        for name in os.listdir(HERE):
            if name.endswith(".py") and not name.startswith("test_"):
                with open(os.path.join(HERE, name), encoding="utf-8") as handle:
                    text = handle.read()
                self.assertFalse(model in text or "gpt-" in text, name)


class FakeResponse(io.BytesIO):
    def __enter__(self):
        return self

    def __exit__(self, *args):
        return False


def response_body(text="{}", status="completed", refusal=None):
    content = [{"type": "refusal", "refusal": refusal}] if refusal else [{"type": "output_text", "text": text}]
    return {"status": status, "output": [{"type": "reasoning"}, {"type": "message", "content": content}],
            "usage": {"input_tokens": 3000, "input_tokens_details": {"cached_tokens": 1000},
                      "output_tokens": 5000, "output_tokens_details": {"reasoning_tokens": 4000}}}


class OpenAIProviderTest(unittest.TestCase):
    KEY = "sk-test-secret"

    def make(self, replies):
        calls = []

        def opener(request, timeout):
            calls.append(request)
            reply = replies.pop(0)
            if isinstance(reply, Exception):
                raise reply
            return FakeResponse(json.dumps(reply).encode("utf-8"))

        provider = providers.OpenAIProvider(ROLES["roles"]["plan"], ROLES["providers"]["openai"], self.KEY,
                                            opener=opener, sleep=lambda s: None)
        return provider, calls

    def http_error(self, code):
        return urllib.error.HTTPError("u", code, "x", {}, io.BytesIO(b'{"error": {"message": "slow down"}}'))

    def test_needs_a_key(self):
        with self.assertRaises(ValueError):
            providers.OpenAIProvider(ROLES["roles"]["plan"], ROLES["providers"]["openai"], "")

    def test_request_body(self):
        provider, calls = self.make([response_body()])
        body = provider.request_body("i", "u", {"type": "object"})
        self.assertEqual(body["model"], ROLES["roles"]["plan"]["model"])
        self.assertEqual(body["reasoning"], {"effort": "medium"})
        self.assertIs(body["text"]["format"]["strict"], True)
        self.assertIs(body["store"], False)

    def test_completed(self):
        provider, calls = self.make([response_body('{"a": 1}')])
        result = provider.plan("i", "u", {}, BY_ID["SP-01"], 1)
        self.assertEqual((result.status, result.text, result.retries), ("completed", '{"a": 1}', 0))
        self.assertEqual(result.usage["cached_input_tokens"], 1000)
        self.assertEqual(calls[0].get_header("Authorization"), f"Bearer {self.KEY}")

    def test_refusal_and_incomplete(self):
        self.assertEqual(providers.parse_response(response_body(refusal="no")).status, "refusal")
        body = response_body(status="incomplete")
        body["incomplete_details"] = {"reason": "max_output_tokens"}
        result = providers.parse_response(body)
        self.assertEqual(result.status, "incomplete")
        self.assertIn("max_output_tokens", result.detail)

    def test_retry_then_success(self):
        provider, calls = self.make([self.http_error(429), self.http_error(503), response_body()])
        result = provider.plan("i", "u", {}, BY_ID["SP-01"], 1)
        self.assertEqual((result.status, result.retries, len(calls)), ("completed", 2, 3))

    def test_no_retry_on_400_and_no_key_in_detail(self):
        provider, calls = self.make([self.http_error(400)])
        result = provider.plan("i", "u", {}, BY_ID["SP-01"], 1)
        self.assertEqual((result.status, len(calls), result.retries), ("error", 1, 0))
        self.assertIn("HTTP 400", result.detail)
        self.assertNotIn(self.KEY, json.dumps(result.__dict__))

    def test_cost(self):
        prices = ROLES["roles"]["plan"]["price_usd_per_million_tokens"]
        usage = providers.parse_response(response_body()).usage
        # 2000 fresh x 0.10 + 1000 cached x 0.01 + 5000 output x 0.50, per million.
        self.assertAlmostEqual(providers.cost_usd(usage, prices), 0.00271)


def quiet_main(argv):
    with contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(io.StringIO()):
        return harness.main(argv)


class HarnessTest(unittest.TestCase):
    def test_budget(self):
        budget = harness.Budget(1.0)
        self.assertTrue(budget.reserve(0.6))
        self.assertFalse(budget.reserve(0.6))
        budget.settle(0.6, 0.1)
        self.assertTrue(budget.reserve(0.6))
        self.assertAlmostEqual(budget.spent, 0.1)

    def test_cap_skip(self):
        role = ROLES["roles"]["plan"]
        record = harness.run_one(providers.FakeProvider(MACHINES, {}), role, harness.Budget(0.0001),
                                 schema_check.load_schema(), BY_ID["SP-01"], MACHINES, 1)
        self.assertEqual(record["status"], "cap_skip")

    def test_worst_case_under_the_cap(self):
        role = ROLES["roles"]["plan"]
        worst = harness.worst_case_usd(role, prompt.instructions(), prompt.user_input(BY_ID["SP-03"], MACHINES))
        self.assertLess(worst * 60, 2 * role["run_cap_usd"])

    def test_fake_run(self):
        with tempfile.TemporaryDirectory() as out:
            self.assertEqual(quiet_main(["--out", out]), 0)
            with open(os.path.join(out, "summary.json"), encoding="utf-8") as handle:
                summary = json.load(handle)
            with open(os.path.join(out, "plans.jsonl"), encoding="utf-8") as handle:
                self.assertEqual(len(handle.readlines()), 60)
        self.assertEqual(summary["requested"], 60)
        self.assertEqual(summary["schema_pass"], 57)
        self.assertEqual(summary["rejected"], 9)
        self.assertEqual(summary["status_counts"], {"completed": 59, "refusal": 1})
        self.assertEqual(summary["per_rule"]["POL-015"]["plans"], 1)
        self.assertEqual(summary["provider"], "fake")

    def test_paid_run_needs_approval_a_key_and_the_cap(self):
        with tempfile.TemporaryDirectory() as out, mock.patch.dict(os.environ, {}, clear=True):
            self.assertEqual(quiet_main(["--provider", "openai", "--out", out]), 2)
            self.assertEqual(quiet_main(["--provider", "openai", "--owner-approved", "--out", out]), 2)
            self.assertEqual(quiet_main(["--cap-usd", "2.5", "--out", out]), 2)
            self.assertEqual(os.listdir(out), [])

    def test_go_decision(self):
        base = {"schema_pass_rate": 0.95, "rejection_rate": 0.25, "cost_per_plan_usd": 0.01}
        self.assertEqual(harness.go_decision(base)["result"], "go")
        for key, value in (("schema_pass_rate", 0.94), ("rejection_rate", 0.26), ("cost_per_plan_usd", 0.011)):
            self.assertEqual(harness.go_decision(dict(base, **{key: value}))["result"], "no-go", key)
        self.assertEqual(harness.go_decision(dict(base, rejection_rate=None))["result"], "no-go")


class PaidRunResultsTest(unittest.TestCase):
    """The report cites the committed results of the paid run. The current
    schema and policy must give the same results from the same plans."""

    def setUp(self):
        results = os.path.join(HERE, "results")
        with open(os.path.join(results, "paid-run-plans.jsonl"), encoding="utf-8") as handle:
            self.records = [json.loads(line) for line in handle]
        with open(os.path.join(results, "paid-run-summary.json"), encoding="utf-8") as handle:
            self.summary = json.load(handle)

    def test_plans_give_the_same_violations(self):
        schema = schema_check.load_schema()
        self.assertEqual(len(self.records), 60)
        for r in self.records:
            self.assertEqual(schema_check.validate(r["plan"], schema), [])
            found = policy.evaluate(r["plan"], BY_ID[r["profile_id"]], MACHINES)
            self.assertEqual(found, r["violations"], (r["profile_id"], r["attempt"]))

    def test_summary_agrees(self):
        again = harness.summarize(self.records, {})
        for key in ("requested", "schema_pass", "rejected", "per_rule", "total_cost_usd", "cost_per_plan_usd"):
            self.assertEqual(again[key], self.summary[key], key)
        self.assertEqual(self.summary["go"]["result"], "go")
        self.assertLessEqual(self.summary["total_cost_usd"], self.summary["cap_usd"])


if __name__ == "__main__":
    unittest.main()
