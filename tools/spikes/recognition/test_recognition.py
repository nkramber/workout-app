"""The unit tests of the recognition spike. They use the fake provider and
no network, so they make no paid call. They need no image cache."""
import contextlib
import io
import json
import os
import tempfile
import unittest
from unittest import mock

import photos
import providers
import recognize
import vision

CATALOG, MANIFEST = photos.load_set()
TYPE_IDS = [t["type_id"] for t in CATALOG["types"]]
ROLES = recognize.load_json("roles.json")
ROLE = ROLES["roles"]["recognize"]
SCHEMA = vision.answer_schema(CATALOG)
TEXT = vision.instructions(CATALOG)


def stub_loader(path, long_side, quality):
    return b"\xff\xd8 fake jpeg"


def fake_records(faults=None, cap=2.0, photo_list=None):
    provider = vision.FakeProvider(CATALOG, faults)
    budget = recognize.plan_harness.Budget(cap)
    chosen = photo_list or photos.photo_list(MANIFEST)
    return [recognize.run_one(provider, ROLE, budget, SCHEMA, TEXT, p, stub_loader) for p in chosen], budget


def answer(machine_type, confidence=0.9):
    return json.dumps({"observed_text": "", "evidence": "e", "quality_flags": [],
                       "machine_type": machine_type, "confidence": confidence})


class FakeResponse(io.BytesIO):
    def __enter__(self):
        return self

    def __exit__(self, *exc):
        return False


class SchemaAndPromptTest(unittest.TestCase):
    def test_enum_is_the_catalog_and_the_abstentions(self):
        enum = SCHEMA["properties"]["machine_type"]["enum"]
        self.assertEqual(enum, TYPE_IDS + ["none_of_these", "abstain"])

    def test_strict_shape(self):
        self.assertIs(SCHEMA["additionalProperties"], False)
        self.assertEqual(set(SCHEMA["required"]), set(SCHEMA["properties"]))
        # The evidence comes before the machine type (section 6.2).
        order = list(SCHEMA["properties"])
        self.assertLess(order.index("evidence"), order.index("machine_type"))

    def test_prompt_lists_each_type_and_has_a_stable_hash(self):
        for type_id in TYPE_IDS:
            self.assertIn(f"- {type_id}: ", TEXT)
        self.assertEqual(vision.prompt_hash(CATALOG), vision.prompt_hash(CATALOG))
        self.assertEqual(len(vision.prompt_hash(CATALOG)), 16)

    def test_no_model_id_at_a_call_site(self):
        # D-24: the model id is in roles.json alone.
        for name in ("vision.py", "recognize.py", "photos.py"):
            with open(os.path.join(photos.HERE, name), encoding="utf-8") as handle:
                self.assertNotIn(ROLE["model"], handle.read(), name)


class ScoreTest(unittest.TestCase):
    def test_catalog_photo(self):
        truth = TYPE_IDS[0]
        self.assertEqual(recognize.score(truth, truth), "correct")
        self.assertEqual(recognize.score(TYPE_IDS[1], truth), "wrong")
        self.assertEqual(recognize.score("abstain", truth), "abstained")
        self.assertEqual(recognize.score("none_of_these", truth), "abstained")

    def test_photo_outside_the_catalog(self):
        self.assertEqual(recognize.score("none_of_these", "none_of_these"), "correct")
        self.assertEqual(recognize.score("abstain", "none_of_these"), "correct")
        self.assertEqual(recognize.score(TYPE_IDS[0], "none_of_these"), "wrong")

    def test_check_answer(self):
        self.assertEqual(recognize.check_answer(answer(TYPE_IDS[0]), SCHEMA)[1], [])
        self.assertTrue(recognize.check_answer(answer(TYPE_IDS[0], 1.5), SCHEMA)[1])
        self.assertTrue(recognize.check_answer(answer("bench_press"), SCHEMA)[1])
        self.assertTrue(recognize.check_answer("not json", SCHEMA)[1])


class PhotoListTest(unittest.TestCase):
    def test_originals_and_degraded(self):
        items = photos.photo_list(MANIFEST, cache="/cache")
        self.assertEqual(len(items), len(MANIFEST["images"]) + len(MANIFEST["degraded"]))
        self.assertEqual(len({p["photo_id"] for p in items}), len(items))
        by_id = {i["image_id"]: i for i in MANIFEST["images"]}
        for p in items:
            if p["group"] == "degraded":
                self.assertEqual(p["machine_type"], by_id[p["source_id"]]["machine_type"])
                self.assertEqual(p["path"], os.path.join("/cache", "degraded", p["photo_id"] + ".jpg"))

    def test_prepare_rotates_downscales_and_drops_exif(self):
        try:
            from PIL import Image
        except ImportError:
            self.skipTest("Pillow is a local tool, and CI does not install it")
        with tempfile.TemporaryDirectory() as tmp:
            path = os.path.join(tmp, "p.jpg")
            exif = Image.Exif()
            exif[0x0112] = 6  # orientation: rotate 90 degrees
            exif[0x010F] = "Maker"
            Image.new("RGB", (3000, 2000), (200, 10, 10)).save(path, "JPEG", exif=exif.tobytes())
            data = photos.prepare(path, 1536, 80)
        self.assertNotIn(b"Exif\x00\x00", data)
        with Image.open(io.BytesIO(data)) as image:
            self.assertEqual(image.format, "JPEG")
            self.assertEqual(image.size, (1024, 1536))


class HarnessTest(unittest.TestCase):
    def test_fake_run_scores_each_fault(self):
        records, _ = fake_records()
        by_id = {r["photo_id"]: r for r in records}
        self.assertEqual(by_id["R002"]["outcome"], "wrong")
        self.assertTrue(by_id["R002"]["high_confidence"])
        self.assertFalse(by_id["R010"]["high_confidence"])
        for photo_id in ("R040", "R050", "R060"):
            self.assertEqual(by_id[photo_id]["outcome"], "invalid", photo_id)
        summary = recognize.summarize(records, {"cap_usd": 2.0})
        self.assertEqual(sum(summary["outcomes"].values()), len(records))
        self.assertEqual(summary["high_confidence_wrong"], len(summary["high_confidence_wrong_answers"]))
        self.assertIn("R002", [w["photo_id"] for w in summary["high_confidence_wrong_answers"]])
        self.assertNotIn("R010", [w["photo_id"] for w in summary["high_confidence_wrong_answers"]])
        self.assertEqual(sum(sum(c.values()) for c in summary["by_group"].values()), len(records))

    def test_go_bar(self):
        base = {"high_confidence_wrong_rate": 0.01, "correct_rate": 0.8, "cost_per_photo_usd": 0.001}
        self.assertEqual(recognize.go_decision(base)["result"], "go")
        for key, value in (("high_confidence_wrong_rate", 0.03), ("correct_rate", 0.6),
                           ("cost_per_photo_usd", 0.02), ("correct_rate", None)):
            result = recognize.go_decision(dict(base, **{key: value}))
            self.assertEqual(result["result"], "no-go", key)

    def test_cap_skip_spends_nothing(self):
        cheap = recognize.worst_case_usd(ROLE, TEXT)
        records, budget = fake_records(cap=cheap / 2)
        self.assertEqual({r["status"] for r in records}, {"cap_skip"})
        self.assertEqual({r["outcome"] for r in records}, {"invalid"})
        self.assertEqual(budget.spent, 0.0)

    def test_the_cap_covers_the_full_run_at_its_worst_case(self):
        count = len(photos.photo_list(MANIFEST))
        self.assertLess(recognize.worst_case_usd(ROLE, TEXT) * count, ROLE["run_cap_usd"])

    def test_main_writes_the_results(self):
        with tempfile.TemporaryDirectory() as cache, tempfile.TemporaryDirectory() as out:
            for p in photos.photo_list(MANIFEST, cache):
                os.makedirs(os.path.dirname(p["path"]), exist_ok=True)
                open(p["path"], "wb").close()
            with contextlib.redirect_stdout(io.StringIO()):
                code = recognize.main(["--out", out], loader=stub_loader, cache=cache)
            self.assertEqual(code, 0)
            with open(os.path.join(out, "summary.json"), encoding="utf-8") as handle:
                summary = json.load(handle)
            self.assertEqual(summary["requested"], len(MANIFEST["images"]) + len(MANIFEST["degraded"]))
            self.assertEqual(summary["model"], ROLE["model"])
            with open(os.path.join(out, "summary.md"), encoding="utf-8") as handle:
                self.assertIn("| R002 | original |", handle.read())

    def test_main_refuses_a_missing_cache(self):
        with tempfile.TemporaryDirectory() as cache, contextlib.redirect_stderr(io.StringIO()) as err:
            self.assertEqual(recognize.main([], loader=stub_loader, cache=cache), 2)
        self.assertIn("download.py", err.getvalue())

    def test_paid_run_needs_approval_a_key_and_the_cap(self):
        env = dict(os.environ)
        env.pop(ROLES["providers"]["openai"]["api_key_env"], None)
        with contextlib.redirect_stderr(io.StringIO()) as err:
            self.assertEqual(recognize.main(["--provider", "openai"]), 2)
            self.assertIn("--owner-approved", err.getvalue())
            self.assertEqual(recognize.main(["--cap-usd", "5"]), 2)
            self.assertIn("D-94", err.getvalue())
        with tempfile.TemporaryDirectory() as cache, contextlib.redirect_stderr(io.StringIO()) as err:
            for p in photos.photo_list(MANIFEST, cache):
                os.makedirs(os.path.dirname(p["path"]), exist_ok=True)
                open(p["path"], "wb").close()
            with mock.patch.dict(os.environ, env, clear=True):
                code = recognize.main(["--provider", "openai", "--owner-approved"], loader=stub_loader, cache=cache)
            self.assertEqual(code, 2)
            self.assertIn("OPENAI_API_KEY", err.getvalue())

    def test_select(self):
        items = photos.photo_list(MANIFEST)
        args = recognize.parse_args(["--photos", "R001,D001"])
        self.assertEqual([p["photo_id"] for p in recognize.select(items, args)], ["R001", "D001"])
        self.assertEqual(len(recognize.select(items, recognize.parse_args(["--limit", "3"]))), 3)


class OpenAIProviderTest(unittest.TestCase):
    def make(self, replies):
        calls = []

        def opener(request, timeout):
            calls.append(request)
            return FakeResponse(json.dumps(replies.pop(0)).encode("utf-8"))

        provider = vision.OpenAIProvider(ROLE, ROLES["providers"]["openai"], "sk-test",
                                         opener=opener, sleep=lambda s: None)
        return provider, calls

    def test_request_body(self):
        provider, _ = self.make([])
        body = provider.recognize_body(TEXT, b"\xff\xd8abc", SCHEMA)
        self.assertEqual(body["model"], ROLE["model"])
        self.assertIs(body["store"], False)
        self.assertIs(body["text"]["format"]["strict"], True)
        image = body["input"][0]["content"][1]
        self.assertEqual(image["detail"], ROLE["image_detail"])
        self.assertTrue(image["image_url"].startswith("data:image/jpeg;base64,/9hhYmM"))

    def test_the_request_holds_no_truth(self):
        body = {"output": [{"type": "message", "content": [{"type": "output_text", "text": answer("rower")}]}],
                "usage": {"input_tokens": 1500, "output_tokens": 300}, "status": "completed"}
        provider, calls = self.make([body])
        photo = photos.photo_list(MANIFEST, cache="/secret-cache")[0]
        gate = recognize.plan_harness.Gate(recognize.plan_harness.Budget(1.0), 0.01,
                                           ROLE["price_usd_per_million_tokens"])
        result = provider.recognize(TEXT, photo, b"img", SCHEMA, gate)
        self.assertEqual(result.status, "completed")
        self.assertAlmostEqual(gate.known_usd, providers.cost_usd(result.usage, ROLE["price_usd_per_million_tokens"]))
        sent = calls[0].data.decode("utf-8")
        self.assertNotIn(photo["photo_id"], sent)
        self.assertNotIn("/secret-cache", sent)


if __name__ == "__main__":
    unittest.main()
