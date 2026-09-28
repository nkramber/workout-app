"""The prompt of the Luna plan spike, version luna-plan-prompt-v1.

The instructions state the fitness boundary of D-36 and the rules of the
draft policy. The policy checks each rule again after Luna answers (D-23).
"""
import hashlib
import json

import policy

PROMPT_VERSION = "luna-plan-prompt-v1"

TEMPLATE = """You plan machine strength training for one user of a fitness app.

Boundary:
- Give fitness guidance only. Do not diagnose, treat, or prescribe rehabilitation. Give no medical or emergency advice.
- When the user asks for medical help, keep the plan conservative. Say in one short sentence that a qualified professional can answer medical questions.
- Do not use these words in any text: {blocked}.

Task:
- Write the first week of a continuous plan for the user in the input JSON.
- Use only the machines of the inventory. Loads are in lb. A null load estimate means that the user does not know the load.
- A restart user is a novice, or a user with 3 or more months since the last training.
- Give each exercise a reason of one short sentence that names the input that it uses.
- Give a short warm-up and cool-down for each session. The guidance field holds short recovery and mobility guidance.
- For a session with no cardio, set the cardio machine_id to "", minutes to 0, and intensity to "none".

A deterministic policy checks each rule below and rejects a plan that breaks one:
{rules}
"""

BLOCKED_WORDS = "diagnose, diagnosis, treat, treatment, rehab, rehabilitation, therapy, cure, heal, prescribe, prescription, medication"


def instructions():
    rules = "\n".join(f"- {r.rule_id}: {r.check}" for r in policy.RULES)
    return TEMPLATE.format(blocked=BLOCKED_WORDS, rules=rules)


def prompt_hash():
    return hashlib.sha256(instructions().encode("utf-8")).hexdigest()[:16]


def user_input(profile, machines):
    catalog = {m["machine_id"]: m for m in machines}
    inventory = []
    for item in profile["inventory"]:
        m = catalog[item["machine_id"]]
        inventory.append({
            "machine_id": m["machine_id"],
            "name": m["name"],
            "kind": m["kind"],
            "region": m["region"],
            "available_weights_lb": m["weights_lb"],
            "load_estimate_lb": item["load_estimate_lb"],
        })
    body = {k: v for k, v in profile.items() if k not in ("inventory", "profile_id")}
    body["inventory"] = inventory
    return json.dumps(body, indent=1)
