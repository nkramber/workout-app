"""The providers of the Luna plan spike: a fake provider and the OpenAI provider.

Tests and CI use the fake provider only, so CI makes no paid call. The
OpenAI provider reads its model id from the role configuration (D-24),
and the API key from the environment of the owner machine. No key goes
into a log, an error, or a result.
"""
import copy
import json
import time
import urllib.error
import urllib.request
from dataclasses import dataclass, field

import policy

RETRY_STATUS = {429, 500, 502, 503, 504}
CARDIO_BY_PREFERENCE = {"treadmill": "M14", "bike": "M15", "rower": "M16"}


@dataclass
class Result:
    status: str  # completed, incomplete, refusal, or error
    text: str = ""
    detail: str = ""
    usage: dict = field(default_factory=dict)
    seconds: float = 0.0
    retries: int = 0


def usage_dict(input_tokens=0, cached=0, output_tokens=0, reasoning=0):
    return {"input_tokens": input_tokens, "cached_input_tokens": cached,
            "output_tokens": output_tokens, "reasoning_tokens": reasoning}


def cost_usd(usage, prices):
    """Reasoning tokens are part of the output tokens, and bill as output."""
    cached = usage.get("cached_input_tokens", 0)
    fresh = usage.get("input_tokens", 0) - cached
    return (fresh * prices["input"] + cached * prices["cached_input"]
            + usage.get("output_tokens", 0) * prices["output"]) / 1_000_000


# The fake provider -------------------------------------------------------

def baseline_plan(profile, machines):
    """A plan that obeys every draft rule. The fake provider starts from it."""
    ctx = policy.Context(profile, machines)
    usable = [ctx.machine(i["machine_id"]) for i in profile["inventory"]
              if i["machine_id"] not in ctx.excluded]
    strength = [m for m in usable if m and m["kind"] == "strength"]
    cardio = [m for m in usable if m and m["kind"] == "cardio"]
    by_region = {}
    for m in strength:
        by_region.setdefault(m["region"], []).append(m)
    sets, rir = (2, 3) if ctx.restart else (3, 2)
    wanted = CARDIO_BY_PREFERENCE.get(profile["cardio_preference"])
    pick = next((m for m in cardio if m["machine_id"] == wanted), None)
    if pick is None and profile["cardio_preference"] == "any" and cardio:
        pick = cardio[0]
    sessions = []
    for index, day in enumerate((1, 3, 5)):
        chosen = [ms[index % len(ms)] for _, ms in sorted(by_region.items())]
        exercises = []
        for m in chosen:
            load = start_load(profile, m, ctx.estimates.get(m["machine_id"]))
            exercises.append({
                "machine_id": m["machine_id"], "exercise_name": m["name"], "rest_seconds": 90,
                "sets": [{"reps": 10, "load_lb": load, "rir_target": rir} for _ in range(sets)],
                "reason": "The load comes from the load estimate and the experience of the user.",
            })
        cardio_block = ({"machine_id": pick["machine_id"], "minutes": 10, "intensity": "easy"}
                        if pick else {"machine_id": "", "minutes": 0, "intensity": "none"})
        sessions.append({"day": day, "title": f"Full body {index + 1}",
                         "warm_up": "Five minutes of easy movement, then one light set on the first machine.",
                         "exercises": exercises, "cardio": cardio_block,
                         "cool_down": "Five minutes of easy walking and gentle stretching."})
    return {"summary": "Three full body sessions on the machines of the gym.",
            "sessions": sessions,
            "guidance": "Sleep well, drink water, and rest one day between sessions."}


def start_load(profile, machine, estimate):
    w = machine["weights_lb"]
    bound = policy.start_bound(profile, w, estimate)
    candidates = [w["min"] + k * w["step"] for k in range(int((w["max"] - w["min"]) / w["step"]) + 1)]
    good = [c for c in candidates if c <= bound + 1e-9 and abs(c / 5 - round(c / 5)) < 1e-9]
    return good[-1] if good else w["min"]


def first_weighted(plan, profile, machines):
    ctx = policy.Context(profile, machines)
    for session in plan["sessions"]:
        for ex in session["exercises"]:
            m = ctx.machine(ex["machine_id"])
            if m and m["kind"] == "strength":
                return ex, m
    raise ValueError("no weighted exercise")


def apply_fault(fault, plan, profile, machines):
    """Break a baseline plan in one known way. Return the output text."""
    plan = copy.deepcopy(plan)
    ex, machine = first_weighted(plan, profile, machines)
    if fault == "not_json":
        return "Here is your plan: " + json.dumps(plan)[:200]
    if fault == "missing_field":
        del plan["guidance"]
    elif fault == "rir_zero":
        ex["sets"][0]["rir_target"] = 0
    elif fault == "off_step":
        ex["sets"][0]["load_lb"] += 2.5
    elif fault == "over_bound":
        for s in ex["sets"]:
            s["load_lb"] = machine["weights_lb"]["max"]
    elif fault == "claim":
        plan["guidance"] += " This plan will treat your knee pain."
    elif fault == "excluded":
        target = (profile.get("excluded_machine_ids") or ["M99"])[0]
        extra = copy.deepcopy(ex)
        extra["machine_id"] = target
        plan["sessions"][0]["exercises"].append(extra)
    elif fault == "rest_short":
        ex["rest_seconds"] = 30
    elif fault == "reps_high":
        ex["sets"][0]["reps"] = 20
    elif fault == "load_jump":
        ex["sets"][-1]["load_lb"] = ex["sets"][0]["load_lb"] * 2
    elif fault == "extra_sessions":
        for day in (2, 4):
            extra = copy.deepcopy(plan["sessions"][0])
            extra["day"] = day
            plan["sessions"].append(extra)
    else:
        raise ValueError(f"unknown fault {fault}")
    return json.dumps(plan)


# A fixed map of faults, so that a fake run shows the report with each kind
# of failure. The key is (profile_id, attempt).
DEFAULT_FAULTS = {
    ("SP-02", 2): "excluded", ("SP-03", 3): "rir_zero", ("SP-05", 1): "not_json",
    ("SP-06", 2): "over_bound", ("SP-09", 1): "missing_field", ("SP-10", 3): "off_step",
    ("SP-12", 1): "load_jump", ("SP-13", 2): "claim", ("SP-15", 1): "rest_short",
    ("SP-18", 3): "reps_high", ("SP-20", 2): "extra_sessions", ("SP-11", 3): "refusal",
}


class FakeProvider:
    name = "fake"

    def __init__(self, machines, faults=None):
        self.machines = machines
        self.faults = DEFAULT_FAULTS if faults is None else faults

    def plan(self, instructions, user_text, schema, profile, attempt):
        started = time.monotonic()
        fault = self.faults.get((profile["profile_id"], attempt))
        usage = usage_dict(input_tokens=(len(instructions) + len(user_text)) // 4, reasoning=800)
        if fault == "refusal":
            usage["output_tokens"] = 20
            return Result("refusal", detail="fake refusal", usage=usage, seconds=time.monotonic() - started)
        plan = baseline_plan(profile, self.machines)
        text = apply_fault(fault, plan, profile, self.machines) if fault else json.dumps(plan)
        usage["output_tokens"] = len(text) // 4 + usage["reasoning_tokens"]
        return Result("completed", text=text, usage=usage, seconds=time.monotonic() - started)


# The OpenAI provider -----------------------------------------------------

def parse_response(body):
    """Read a Responses API body into a Result, with no seconds or retries."""
    text, refusal = "", None
    for item in body.get("output") or []:
        if item.get("type") != "message":
            continue
        for part in item.get("content") or []:
            if part.get("type") == "output_text":
                text += part.get("text", "")
            elif part.get("type") == "refusal":
                refusal = part.get("refusal", "")
    raw = body.get("usage") or {}
    usage = usage_dict(
        input_tokens=raw.get("input_tokens", 0),
        cached=(raw.get("input_tokens_details") or {}).get("cached_tokens", 0),
        output_tokens=raw.get("output_tokens", 0),
        reasoning=(raw.get("output_tokens_details") or {}).get("reasoning_tokens", 0),
    )
    if refusal is not None:
        return Result("refusal", detail=refusal[:300], usage=usage)
    status = body.get("status", "completed")
    if status != "completed":
        reason = (body.get("incomplete_details") or {}).get("reason", "")
        return Result("incomplete" if status == "incomplete" else "error",
                      text=text, detail=f"status {status} {reason}".strip(), usage=usage)
    return Result("completed", text=text, usage=usage)


class OpenAIProvider:
    name = "openai"

    def __init__(self, role, config, api_key, opener=urllib.request.urlopen, sleep=time.sleep,
                 timeout=600, attempts=3):
        if not api_key:
            raise ValueError(f"set {config['api_key_env']} in the environment")
        self.role, self.config, self._key = role, config, api_key
        self.opener, self.sleep, self.timeout, self.attempts = opener, sleep, timeout, attempts

    def request_body(self, instructions, user_text, schema):
        return {
            "model": self.role["model"],
            "reasoning": {"effort": self.role["reasoning_effort"]},
            "instructions": instructions,
            "input": user_text,
            "text": {"format": {"type": "json_schema", "name": "luna_plan", "schema": schema, "strict": True}},
            "max_output_tokens": self.role["max_output_tokens"],
            "store": False,
        }

    def plan(self, instructions, user_text, schema, profile, attempt):
        data = json.dumps(self.request_body(instructions, user_text, schema)).encode("utf-8")
        started = time.monotonic()
        detail, n = "", 0
        for n in range(self.attempts):
            request = urllib.request.Request(self.config["endpoint"], data=data, method="POST", headers={
                "Authorization": f"Bearer {self._key}", "Content-Type": "application/json"})
            try:
                with self.opener(request, timeout=self.timeout) as response:
                    body = json.loads(response.read().decode("utf-8"))
                result = parse_response(body)
                result.seconds, result.retries = time.monotonic() - started, n
                return result
            except urllib.error.HTTPError as exc:
                detail = f"HTTP {exc.code}: {read_error(exc)}"
                if exc.code not in RETRY_STATUS:
                    break
            except (urllib.error.URLError, TimeoutError, OSError, ValueError) as exc:
                detail = f"{exc.__class__.__name__}: {str(exc)[:200]}"
            if n + 1 < self.attempts:
                self.sleep(2 ** (n + 1))
        return Result("error", detail=detail, seconds=time.monotonic() - started, retries=n)


def read_error(exc):
    try:
        body = json.loads(exc.read().decode("utf-8"))
        return str((body.get("error") or {}).get("message", ""))[:300]
    except (ValueError, OSError, AttributeError):
        return ""
