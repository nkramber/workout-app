"""The first draft of the policy rules table (D-23), for the Luna plan spike.

Luna proposes a plan. The policy checks every set and load of it. Each rule
has an id, a version, its sources, and the open questions that it touches.
The whole table is a draft: the owner approves no rule of it yet (D-39). A
rule that touches an open question (Q-92, Q-101, Q-102, Q-104, Q-105, or
Q-106) uses a draft value for it, and it does not answer the question. The
restart threshold of 3 months is such a draft value for Q-102.

The values come from `docs/research/exercise-safety.md` sections 5.2, 5.7,
5.8, and 5.12. Most of them are assumptions of that research, not owner
decisions.
"""
import re
from dataclasses import dataclass, field

BUNDLE_VERSION = "policy-draft-0.1.0"
RESTART_MONTHS = 3
RESTART_START_FRACTION = 0.8
UNKNOWN_ESTIMATE_STEPS = 2
JUMP_FRACTION = 0.10
JUMP_STEP_LB = 5
REPS_MIN, REPS_MAX = 6, 15
SETS_MAX, RESTART_SETS_MAX = 3, 2
RESTART_REGION_SETS_MAX = 8
REST_MIN, REST_MAX = 60, 180
SESSIONS_MIN, SESSIONS_MAX = 2, 3
CARDIO_MINUTES_MAX = 45
BLOCKED_CLAIMS = re.compile(
    r"\b(diagnos\w*|treat|treats|treated|treatment\w*|rehab\w*|therap\w*|cure|cures|cured|"
    r"heal|heals|healed|prescri\w*|medication\w*|physiotherap\w*)\b"
    r"|\b(lower|lowers|reduce|reduces|fix|fixes|control|controls)\s+(your\s+)?blood pressure\b",
    re.I,
)


@dataclass(frozen=True)
class Rule:
    rule_id: str
    version: int
    title: str
    check: str
    sources: tuple
    questions: tuple = field(default=())

    @property
    def status(self):
        if self.questions:
            return "draft, no answer: " + ", ".join(self.questions)
        return "draft"


RULES = (
    Rule("POL-001", 1, "Known machine",
         "Each exercise uses a strength machine of the inventory, and each cardio block uses a cardio machine of it. No plan uses an excluded machine.",
         ("D-48", "D-49", "D-54")),
    Rule("POL-002", 1, "Complete plan",
         "The plan has one or more sessions, each session one or more exercises, and each exercise one or more sets. Each day is 1 to 7 and occurs once.",
         ("D-23",)),
    Rule("POL-003", 1, "Reps in reserve range",
         "Each set targets 1 to 3 reps in reserve. No set targets failure.",
         ("D-37",)),
    Rule("POL-004", 1, "Restart reps in reserve",
         "For a novice, or after a break of 3 months or more, each set targets 3 reps in reserve.",
         ("D-37", "exercise-safety 5.2", "exercise-safety 5.7"), ("Q-102", "Q-106")),
    Rule("POL-005", 1, "5 lb rounding",
         "Each load is a multiple of 5 lb.",
         ("D-65",)),
    Rule("POL-006", 1, "Machine range",
         "Each load is between the lowest and the highest weight of the machine.",
         ("D-54",)),
    Rule("POL-007", 1, "Load exists on the machine",
         "Each load is a weight that the stack of the machine has.",
         ("D-54", "D-65"), ("Q-105",)),
    Rule("POL-008", 1, "Start load bound",
         "Each load is at most the user estimate, or 80% of it for a restart, rounded to the nearest 5 lb with a tie down. With no estimate, the bound is the lowest weight plus two steps.",
         ("D-41", "D-65", "exercise-safety 5.7"), ("Q-102", "Q-104")),
    Rule("POL-009", 1, "Load jump",
         "In one exercise, and from one session to a later one, a load rises by at most 10% or 5 lb, whichever is larger. A restart plan never raises a load from one session to a later one.",
         ("D-65", "exercise-safety 5.8"), ("Q-92", "Q-102")),
    Rule("POL-010", 1, "Rep range",
         "Each set has 6 to 15 reps.",
         ("exercise-safety 5.2",)),
    Rule("POL-011", 1, "Sets for each exercise",
         "Each exercise has at most 3 sets, and at most 2 for a restart.",
         ("exercise-safety 5.2",), ("Q-102",)),
    Rule("POL-012", 1, "Restart weekly volume",
         "For a restart, each region gets at most 8 direct sets in the week.",
         ("exercise-safety 5.2",), ("Q-102",)),
    Rule("POL-013", 1, "Rest",
         "Each exercise rests 60 to 180 seconds between sets.",
         ("D-59", "exercise-safety 5.2")),
    Rule("POL-014", 1, "Sessions each week",
         "The plan has 2 or 3 sessions in the week.",
         ("exercise-safety 5.2",)),
    Rule("POL-015", 1, "Fitness boundary",
         "No text holds a blocked claim: diagnosis, treatment, rehabilitation, therapy, cure, prescription, medication, or blood pressure control.",
         ("D-36", "D-93", "exercise-safety 5.10"), ("Q-101",)),
    Rule("POL-016", 1, "Cardio block",
         "A cardio block with a machine has 1 to 45 minutes and an intensity. A cardio block with no machine has 0 minutes and the intensity none.",
         ("D-44",)),
)
RULES_BY_ID = {rule.rule_id: rule for rule in RULES}


def round5(value):
    """Round to the nearest 5 lb (D-65). A tie rounds down (draft, Q-104)."""
    low = (value // 5) * 5
    return low if value - low <= 2.5 else low + 5


def on_stack(load, weights):
    if load < weights["min"] - 1e-9 or load > weights["max"] + 1e-9:
        return False
    steps = (load - weights["min"]) / weights["step"]
    return abs(steps - round(steps)) < 1e-6


def is_restart(profile):
    exp = profile["experience"]
    months = exp.get("months_since_last_training")
    return exp["level"] == "novice" or months is None or months >= RESTART_MONTHS


def start_bound(profile, weights, estimate):
    if estimate is None:
        return weights["min"] + UNKNOWN_ESTIMATE_STEPS * weights["step"]
    base = estimate * RESTART_START_FRACTION if is_restart(profile) else estimate
    return max(min(round5(base), weights["max"]), weights["min"])


def jump_limit(previous):
    return max(previous * JUMP_FRACTION, JUMP_STEP_LB)


class Context:
    def __init__(self, profile, machines):
        self.profile = profile
        self.catalog = {m["machine_id"]: m for m in machines}
        self.estimates = {i["machine_id"]: i["load_estimate_lb"] for i in profile["inventory"]}
        self.excluded = set(profile.get("excluded_machine_ids", []))
        self.restart = is_restart(profile)

    def machine(self, machine_id):
        if machine_id in self.estimates and machine_id in self.catalog:
            return self.catalog[machine_id]
        return None


def exercises(plan):
    for session in plan.get("sessions", []):
        for ex in session.get("exercises", []):
            yield session, ex


def weighted_sets(ctx, plan):
    for session, ex in exercises(plan):
        machine = ctx.machine(ex["machine_id"])
        if not machine or machine["kind"] != "strength":
            continue
        for n, s in enumerate(ex["sets"], 1):
            yield session, ex, machine, n, s


def where(session, ex=None, n=None):
    text = f"day {session.get('day')}"
    if ex is not None:
        text += f" {ex.get('machine_id')}"
    if n is not None:
        text += f" set {n}"
    return text


def pol_001(ctx, plan):
    for session, ex in exercises(plan):
        machine = ctx.machine(ex["machine_id"])
        if machine is None:
            yield where(session, ex), "machine not in the inventory"
        elif machine["kind"] != "strength":
            yield where(session, ex), "exercise on a cardio machine"
        if ex["machine_id"] in ctx.excluded:
            yield where(session, ex), "machine excluded by the user"
    for session in plan["sessions"]:
        mid = session["cardio"]["machine_id"]
        if not mid:
            continue
        machine = ctx.machine(mid)
        if machine is None:
            yield where(session), f"cardio machine {mid} not in the inventory"
        elif machine["kind"] != "cardio":
            yield where(session), f"cardio on strength machine {mid}"
        if mid in ctx.excluded:
            yield where(session), f"cardio machine {mid} excluded by the user"


def pol_002(ctx, plan):
    if not plan["sessions"]:
        yield "plan", "no session"
    days = [s["day"] for s in plan["sessions"]]
    for day in days:
        if not 1 <= day <= 7:
            yield f"day {day}", "day outside 1 to 7"
    if len(days) != len(set(days)):
        yield "plan", f"a day occurs more than once: {days}"
    for session in plan["sessions"]:
        if not session["exercises"]:
            yield where(session), "no exercise"
        for ex in session["exercises"]:
            if not ex["sets"]:
                yield where(session, ex), "no set"


def pol_003(ctx, plan):
    for session, ex in exercises(plan):
        for n, s in enumerate(ex["sets"], 1):
            if not 1 <= s["rir_target"] <= 3:
                yield where(session, ex, n), f"rir_target {s['rir_target']}"


def pol_004(ctx, plan):
    if not ctx.restart:
        return
    for session, ex in exercises(plan):
        for n, s in enumerate(ex["sets"], 1):
            if s["rir_target"] != 3:
                yield where(session, ex, n), f"restart set with rir_target {s['rir_target']}"


def pol_005(ctx, plan):
    for session, ex, _, n, s in weighted_sets(ctx, plan):
        if abs(s["load_lb"] / 5 - round(s["load_lb"] / 5)) > 1e-9:
            yield where(session, ex, n), f"load {s['load_lb']} lb is not a multiple of 5"


def pol_006(ctx, plan):
    for session, ex, machine, n, s in weighted_sets(ctx, plan):
        w = machine["weights_lb"]
        if not w["min"] <= s["load_lb"] <= w["max"]:
            yield where(session, ex, n), f"load {s['load_lb']} lb outside {w['min']} to {w['max']} lb"


def pol_007(ctx, plan):
    for session, ex, machine, n, s in weighted_sets(ctx, plan):
        w = machine["weights_lb"]
        if w["min"] <= s["load_lb"] <= w["max"] and not on_stack(s["load_lb"], w):
            yield where(session, ex, n), f"load {s['load_lb']} lb not on a stack of {w['min']} + n x {w['step']} lb"


def pol_008(ctx, plan):
    for session, ex, machine, n, s in weighted_sets(ctx, plan):
        estimate = ctx.estimates.get(machine["machine_id"])
        bound = start_bound(ctx.profile, machine["weights_lb"], estimate)
        if s["load_lb"] > bound + 1e-9:
            yield where(session, ex, n), f"load {s['load_lb']} lb above the start bound {bound:g} lb (estimate {estimate})"


def pol_009(ctx, plan):
    previous_max = {}
    for session in sorted(plan["sessions"], key=lambda s: s["day"]):
        session_max = {}
        for ex in session["exercises"]:
            machine = ctx.machine(ex["machine_id"])
            if not machine or machine["kind"] != "strength":
                continue
            mid = machine["machine_id"]
            loads = [s["load_lb"] for s in ex["sets"]]
            for n in range(1, len(loads)):
                if loads[n] - loads[n - 1] > jump_limit(loads[n - 1]) + 1e-9:
                    yield where(session, ex, n + 1), f"load rises {loads[n - 1]} to {loads[n]} lb in one exercise"
            if loads and mid in previous_max:
                prev, top = previous_max[mid], max(loads)
                if ctx.restart and top > prev + 1e-9:
                    yield where(session, ex), f"restart load rises {prev} to {top} lb from an earlier session"
                elif top - prev > jump_limit(prev) + 1e-9:
                    yield where(session, ex), f"load rises {prev} to {top} lb from an earlier session"
            if loads:
                session_max[mid] = max(session_max.get(mid, 0), max(loads))
        previous_max.update(session_max)


def pol_010(ctx, plan):
    for session, ex in exercises(plan):
        for n, s in enumerate(ex["sets"], 1):
            if not REPS_MIN <= s["reps"] <= REPS_MAX:
                yield where(session, ex, n), f"{s['reps']} reps"


def pol_011(ctx, plan):
    limit = RESTART_SETS_MAX if ctx.restart else SETS_MAX
    for session, ex in exercises(plan):
        if len(ex["sets"]) > limit:
            yield where(session, ex), f"{len(ex['sets'])} sets, limit {limit}"


def pol_012(ctx, plan):
    if not ctx.restart:
        return
    totals = {}
    for _, ex in exercises(plan):
        machine = ctx.machine(ex["machine_id"])
        if machine and machine["kind"] == "strength":
            totals[machine["region"]] = totals.get(machine["region"], 0) + len(ex["sets"])
    for region, total in sorted(totals.items()):
        if total > RESTART_REGION_SETS_MAX:
            yield f"region {region}", f"{total} sets in the week, limit {RESTART_REGION_SETS_MAX}"


def pol_013(ctx, plan):
    for session, ex in exercises(plan):
        if not REST_MIN <= ex["rest_seconds"] <= REST_MAX:
            yield where(session, ex), f"rest {ex['rest_seconds']} s"


def pol_014(ctx, plan):
    count = len(plan["sessions"])
    if not SESSIONS_MIN <= count <= SESSIONS_MAX:
        yield "plan", f"{count} sessions"


def texts(plan):
    yield "summary", plan["summary"]
    yield "guidance", plan["guidance"]
    for session in plan["sessions"]:
        for key in ("title", "warm_up", "cool_down"):
            yield f"{where(session)} {key}", session[key]
        for ex in session["exercises"]:
            yield f"{where(session, ex)} exercise_name", ex["exercise_name"]
            yield f"{where(session, ex)} reason", ex["reason"]


def pol_015(ctx, plan):
    for place, text in texts(plan):
        for match in BLOCKED_CLAIMS.finditer(text):
            start = max(0, match.start() - 40)
            yield place, f"blocked claim {match.group(0)!r} in: ...{text[start:match.end() + 40]}..."


def pol_016(ctx, plan):
    for session in plan["sessions"]:
        c = session["cardio"]
        if c["machine_id"]:
            if not 1 <= c["minutes"] <= CARDIO_MINUTES_MAX or c["intensity"] == "none":
                yield where(session), f"cardio {c['minutes']} min at {c['intensity']}"
        elif c["minutes"] != 0 or c["intensity"] != "none":
            yield where(session), f"cardio with no machine: {c['minutes']} min at {c['intensity']}"


CHECKS = {
    "POL-001": pol_001, "POL-002": pol_002, "POL-003": pol_003, "POL-004": pol_004,
    "POL-005": pol_005, "POL-006": pol_006, "POL-007": pol_007, "POL-008": pol_008,
    "POL-009": pol_009, "POL-010": pol_010, "POL-011": pol_011, "POL-012": pol_012,
    "POL-013": pol_013, "POL-014": pol_014, "POL-015": pol_015, "POL-016": pol_016,
}


def evaluate(plan, profile, machines):
    """Return each violation of a schema-valid plan, as a list of dicts."""
    ctx = Context(profile, machines)
    found = []
    for rule in RULES:
        for place, detail in CHECKS[rule.rule_id](ctx, plan):
            found.append({"rule_id": rule.rule_id, "rule_version": rule.version, "where": place, "detail": detail})
    return found


def rules_table():
    """Return the rules table as Markdown rows, for the report."""
    rows = ["| Id | Version | Title | Check | Sources | Status |", "|---|---|---|---|---|---|"]
    for r in RULES:
        rows.append(f"| {r.rule_id} | {r.version} | {r.title} | {r.check} | {', '.join(r.sources)} | {r.status} |")
    return "\n".join(rows)
