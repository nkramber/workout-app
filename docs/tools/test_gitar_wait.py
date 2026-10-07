"""Tests of gitar_wait.py (D-338). A port of the GitarWaitTests of what-you-carry."""
import contextlib
import importlib.util
import io
import json
import os
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
spec = importlib.util.spec_from_file_location("gitar_wait", os.path.join(HERE, "gitar_wait.py"))
gw = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gw)
cr = gw.cr

N = 44
HEAD = "a" * 40
PUSHED = "2026-10-06T10:00:00Z"
DASH = f"<details><summary>{cr.DASHBOARD}</summary></details>"
SPIN = '<kbd><img src="https://x/gitar-spin.svg"> Responding to your feedback</kbd>\n' + DASH


def comment(login, body, created, updated=None):
    return {"user": {"login": login}, "body": body, "created_at": created, "updated_at": updated or created}


def dashboard(updated, body=DASH):
    return comment(cr.GITAR, body, "2026-10-06T09:00:00Z", updated)


class GitHub:
    """A fake `gh`. Each poll reads the next state of a list, and the last state stays."""

    def __init__(self, states, fail=None):
        self.states = states
        self.fail = fail
        self.polls = 0
        self.calls = []
        self.comments = []

    def state(self):
        return self.states[min(self.polls, len(self.states) - 1)]

    def __call__(self, cmd, cwd=None, timeout=None, stdout=None, stderr=None, env=None):
        self.calls.append(cmd)
        if self.fail and self.fail in " ".join(cmd):
            return 1, "", "HTTP 502"
        if cmd[:3] == ["gh", "repo", "view"]:
            return 0, "o/r\n", ""
        if cmd[:3] == ["gh", "pr", "view"]:
            return 0, HEAD + "\n", ""
        if cmd[:3] == ["gh", "pr", "comment"]:
            self.comments.append(cmd[-1])
            return 0, "", ""
        if cmd[2].endswith("/check-suites"):
            return 0, PUSHED + "\n", ""
        if "/check-runs" in cmd[2]:
            runs = [{"status": s, "app": {"slug": cr.GITAR_SLUG}} for s in self.state()["runs"]]
            runs.append({"status": "in_progress", "app": {"slug": "github-actions"}})
            return 0, json.dumps([{"check_runs": runs}]), ""
        if cmd[-1].endswith(f"/issues/{N}/comments"):
            state = self.state()
            self.polls += 1
            return 0, json.dumps([state["comments"]]), ""
        return 1, "", f"unexpected: {cmd}"


class Clock:
    def __init__(self):
        self.now = 0.0
        self.sleeps = []

    def sleep(self, seconds):
        self.sleeps.append(seconds)
        self.now += seconds

    def __call__(self):
        return self.now


def wait(github):
    clock = Clock()
    lines = []
    code = gw.wait(github, N, sleep=clock.sleep, clock=clock, out=lines.append)
    return code, clock, lines


class Times(unittest.TestCase):
    def test_the_times_are_those_of_d_338(self):
        self.assertEqual((gw.FIRST, gw.POLL, gw.REQUEST, gw.LIMIT), (60, 30, 360, 900))


class Wait(unittest.TestCase):
    def test_a_completed_run_and_a_new_dashboard_end_the_wait_with_no_request(self):
        code, clock, lines = wait(GitHub([{"runs": ["completed"], "comments": [dashboard("2026-10-06T10:02:00Z")]}]))
        self.assertEqual(code, gw.EXIT_DONE)
        self.assertEqual(clock.sleeps, [gw.FIRST])
        self.assertIn("is current", lines[-1])

    def test_a_running_run_keeps_the_wait(self):
        states = [{"runs": ["in_progress"], "comments": [dashboard("2026-10-06T10:02:00Z")]}] * 3
        states.append({"runs": ["completed"], "comments": [dashboard("2026-10-06T10:04:00Z")]})
        github = GitHub(states)
        code, clock, _ = wait(github)
        self.assertEqual(code, gw.EXIT_DONE)
        self.assertEqual(clock.sleeps, [gw.FIRST, gw.POLL, gw.POLL, gw.POLL])
        self.assertEqual(github.comments, [])

    def test_a_completed_run_with_an_old_dashboard_is_no_review(self):
        # On PR #103 of what-you-carry, the check run completed 53 seconds before the dashboard came.
        github = GitHub([{"runs": ["completed"], "comments": [dashboard("2026-10-06T09:30:00Z")]}])
        code, clock, lines = wait(github)
        self.assertEqual(code, gw.EXIT_STOP)
        self.assertGreaterEqual(clock.now, gw.LIMIT)
        self.assertEqual(github.comments, [])
        self.assertIn("tell the owner (D-338)", "\n".join(lines))

    def test_a_dashboard_with_the_spinner_keeps_the_wait(self):
        states = [{"runs": ["completed"], "comments": [dashboard("2026-10-06T10:02:00Z", SPIN)]}] * 2
        states.append({"runs": ["completed"], "comments": [dashboard("2026-10-06T10:03:00Z")]})
        code, clock, _ = wait(GitHub(states))
        self.assertEqual(code, gw.EXIT_DONE)
        self.assertEqual(len(clock.sleeps), 3)

    def test_no_run_gives_one_request_then_a_stop_at_the_limit(self):
        github = GitHub([{"runs": [], "comments": [dashboard("2026-10-06T09:30:00Z")]}])
        code, clock, lines = wait(github)
        self.assertEqual(code, gw.EXIT_STOP)
        self.assertEqual(github.comments, ["Gitar review"])
        self.assertTrue(any("Posted one Gitar review comment" in line for line in lines))
        self.assertGreaterEqual(clock.now, gw.LIMIT)

    def test_no_request_before_the_request_time(self):
        github = GitHub([{"runs": [], "comments": []}])
        clock = Clock()
        lines = []
        gw.wait(github, N, sleep=clock.sleep, clock=clock, out=lines.append)
        request = [i for i, cmd in enumerate(github.calls) if cmd[:3] == ["gh", "pr", "comment"]]
        self.assertEqual(len(request), 1)
        self.assertTrue(any(line.startswith(f"gitar-wait: no Gitar check run on {HEAD[:7]} after {gw.REQUEST} s") for line in lines), lines)

    def test_a_manual_review_with_no_check_run_ends_the_wait(self):
        ask = comment("nkramber", "Gitar review", "2026-10-06T10:06:00Z")
        reply = comment(cr.GITAR, "> Gitar review\n\nOn it", "2026-10-06T10:06:10Z")
        states = [{"runs": [], "comments": [dashboard("2026-10-06T09:30:00Z")]}] * 11
        states.append({"runs": [], "comments": [dashboard("2026-10-06T10:08:00Z"), ask, reply]})
        github = GitHub(states)
        code, _, _ = wait(github)
        self.assertEqual(code, gw.EXIT_DONE)
        self.assertEqual(github.comments, ["Gitar review"])

    def test_a_refused_request_does_not_end_the_wait(self):
        ask = comment("nkramber", "Gitar review", "2026-10-06T10:06:00Z")
        refusal = comment(cr.GITAR, "> Gitar review\n\nYou've sent several Gitar comments in a short window", "2026-10-06T10:06:10Z")
        github = GitHub([{"runs": [], "comments": [dashboard("2026-10-06T10:08:00Z"), ask, refusal]}])
        code, _, lines = wait(github)
        self.assertEqual(code, gw.EXIT_STOP)
        self.assertTrue(any("refused" in line for line in lines), lines)


class Main(unittest.TestCase):
    def test_a_failed_read_exits_1_and_names_the_read(self):
        clock = Clock()
        with contextlib.redirect_stdout(io.StringIO()) as out:
            code = gw.main(["--pr", str(N)], run=GitHub([{"runs": [], "comments": []}], fail="check-runs"),
                           sleep=clock.sleep, clock=clock)
        self.assertEqual(code, gw.EXIT_STOP)
        self.assertIn("check-runs", out.getvalue())
        self.assertIn("HTTP 502", out.getvalue())

    def test_no_pr_number_is_a_usage_error(self):
        with contextlib.redirect_stderr(io.StringIO()):
            self.assertEqual(gw.main([]), gw.EXIT_USAGE)


if __name__ == "__main__":
    unittest.main()
