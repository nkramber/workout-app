"""Tests of the context checkpoint hook (D-6, D-12).

Run: python3 -m unittest discover -s docs/tools -p 'test_*.py'
"""
import importlib.util
import io
import json
import os
import subprocess
import sys
import tempfile
import unittest
from unittest import mock

HERE = os.path.dirname(os.path.abspath(__file__))
HOOK = os.path.join(HERE, "..", "..", ".claude", "hooks", "context_checkpoint.py")
spec = importlib.util.spec_from_file_location("context_checkpoint", HOOK)
cc = importlib.util.module_from_spec(spec)
spec.loader.exec_module(cc)


def call(context, sidechain=False):
    usage = {"input_tokens": 3, "cache_read_input_tokens": context - 1003, "cache_creation_input_tokens": 1000}
    return json.dumps({"type": "assistant", "isSidechain": sidechain, "message": {"usage": usage}})


class LastContext(unittest.TestCase):
    def test_reads_the_last_main_thread_call(self):
        lines = [call(100_000), call(310_000), json.dumps({"type": "user"}), call(900_000, sidechain=True)]
        self.assertEqual(cc.last_context(lines), 310_000)

    def test_skips_a_broken_line_and_a_call_with_no_usage(self):
        lines = [call(120_000), "{broken", json.dumps({"type": "assistant", "message": {}})]
        self.assertEqual(cc.last_context(lines), 120_000)

    def test_no_call_gives_none(self):
        self.assertIsNone(cc.last_context([json.dumps({"type": "user"})]))

    def test_the_tail_drops_a_cut_first_line(self):
        with tempfile.NamedTemporaryFile("w", suffix=".jsonl", delete=False) as handle:
            handle.write(call(301_000) + "\n" + call(302_000) + "\n")
        try:
            lines = cc.read_tail(handle.name, size=len(call(302_000)) + 10)
            self.assertEqual(cc.last_context(lines), 302_000)
            self.assertEqual(len(lines), 1)
        finally:
            os.unlink(handle.name)


class Decide(unittest.TestCase):
    def setUp(self):
        self.store = tempfile.mkdtemp()

    def test_levels(self):
        self.assertEqual(cc.level(299_999), 0)
        self.assertEqual(cc.level(300_000), 300_000)
        self.assertEqual(cc.level(399_999), 300_000)
        self.assertEqual(cc.level(401_000), 400_000)

    def test_below_the_limit_says_nothing(self):
        self.assertEqual(cc.decide(self.store, "s1", 299_999), "")
        self.assertFalse(os.listdir(self.store))

    def test_each_level_speaks_one_time(self):
        first = cc.decide(self.store, "s1", 305_123)
        self.assertIn("305,123", first)
        self.assertIn("D-12", first)
        self.assertIn("one-pr-one-session", first)
        self.assertIn("docs/session-handoff.md", first)
        self.assertNotIn("decktome", cc.STORE)
        self.assertEqual(cc.decide(self.store, "s1", 350_000), "")
        self.assertIn("402,000", cc.decide(self.store, "s1", 402_000))
        self.assertEqual(cc.decide(self.store, "s1", 390_000), "")

    def test_the_session_continues_until_the_merge(self):
        # D-85: the checkpoint asks for a compaction, never for the end of the session.
        text = cc.message(305_000)
        self.assertIn("D-85", text)
        self.assertIn("context compaction", text)
        self.assertIn("Do not end the session", text)
        self.assertNotIn("end this session", text)
        self.assertNotIn("new clean session", text)

    def test_sessions_are_apart(self):
        self.assertTrue(cc.decide(self.store, "s1", 310_000))
        self.assertTrue(cc.decide(self.store, "s2", 310_000))

    def test_a_broken_record_speaks_again(self):
        with open(os.path.join(self.store, "s1.json"), "w", encoding="utf-8") as handle:
            handle.write("{broken")
        self.assertTrue(cc.decide(self.store, "s1", 310_000))


class Main(unittest.TestCase):
    def run_hook(self, event):
        out = io.StringIO()
        with mock.patch.object(sys, "stdin", io.StringIO(json.dumps(event))), mock.patch.object(sys, "stdout", out):
            code = cc.main()
        return code, out.getvalue()

    def transcript(self, context):
        handle = tempfile.NamedTemporaryFile("w", suffix=".jsonl", delete=False)
        handle.write(call(context) + "\n")
        handle.close()
        self.addCleanup(os.unlink, handle.name)
        return handle.name

    def repo(self):
        directory = tempfile.mkdtemp()
        subprocess.run(["git", "init", "-q", directory], check=True)
        return directory

    def test_past_the_limit_gives_additional_context(self):
        repo = self.repo()
        code, out = self.run_hook({"session_id": "abc-1", "cwd": repo, "transcript_path": self.transcript(333_000)})
        self.assertEqual(code, 0)
        body = json.loads(out)["hookSpecificOutput"]
        self.assertEqual(body["hookEventName"], "PostToolUse")
        self.assertIn("333,000", body["additionalContext"])
        self.assertTrue(os.path.isfile(os.path.join(repo, ".git", cc.STORE, "abc-1.json")))
        self.assertEqual(self.run_hook({"session_id": "abc-1", "cwd": repo,
                                        "transcript_path": self.transcript(340_000)}), (0, ""))

    def test_silent_cases(self):
        repo = self.repo()
        path = self.transcript(333_000)
        for event in (
            {"session_id": "abc-2", "cwd": repo, "transcript_path": self.transcript(200_000)},
            {"session_id": "abc-3", "cwd": repo, "transcript_path": path, "agent_id": "sub-1"},
            {"session_id": "../x", "cwd": repo, "transcript_path": path},
            {"session_id": "abc-4", "cwd": repo, "transcript_path": "/no/such/file.jsonl"},
            {"session_id": "abc-5", "cwd": tempfile.mkdtemp(), "transcript_path": path},
        ):
            self.assertEqual(self.run_hook(event), (0, ""), event)

    def test_bad_stdin_fails_open(self):
        with mock.patch.object(sys, "stdin", io.StringIO("{broken")):
            self.assertEqual(cc.main(), 0)


if __name__ == "__main__":
    unittest.main()
