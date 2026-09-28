#!/usr/bin/env python3
"""The harness of the Luna plan spike (work area 1.1, D-100).

It sends each synthetic profile to the plan role a fixed number of times,
checks each answer against the plan schema and the draft policy, and
writes the results and a summary.

Free run, with the fake provider:
    python3 tools/spikes/luna_plan/harness.py

Paid run. The owner approves it at run time (D-25), and the cap is D-98:
    python3 tools/spikes/luna_plan/harness.py --provider openai --owner-approved

The results go to .local/spikes/luna_plan/<run id>/, which Git ignores.
"""
import argparse
import datetime
import json
import os
import sys
import threading
from concurrent.futures import ThreadPoolExecutor

import policy
import prompt
import providers
import schema_check

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", ".."))
# The go bar of the Luna plan risk (D-101).
GO_BAR = {"schema_pass_rate_min": 0.95, "rejection_rate_max": 0.25, "cost_per_plan_max_usd": 0.01}
RETURNED = {"completed", "incomplete", "refusal"}


def load_json(name):
    with open(os.path.join(HERE, name), encoding="utf-8") as handle:
        return json.load(handle)


class Budget:
    """Reserve the worst-case cost of a call before the call, so that the
    spend never passes the cap, with any number of workers."""

    def __init__(self, cap_usd):
        self.cap = cap_usd
        self.spent = 0.0
        self.reserved = 0.0
        self.lock = threading.Lock()

    def reserve(self, amount):
        with self.lock:
            if self.spent + self.reserved + amount > self.cap + 1e-12:
                return False
            self.reserved += amount
            return True

    def settle(self, reserved, actual):
        with self.lock:
            self.reserved -= reserved
            self.spent += actual


def worst_case_usd(role, instructions, user_text):
    prices = role["price_usd_per_million_tokens"]
    input_tokens = (len(instructions) + len(user_text)) // 2  # a high estimate of the tokens
    return (input_tokens * prices["input"] + role["max_output_tokens"] * prices["output"]) / 1_000_000


def run_one(provider, role, budget, schema, profile, machines, attempt):
    instructions = prompt.instructions()
    user_text = prompt.user_input(profile, machines)
    record = {"profile_id": profile["profile_id"], "attempt": attempt}
    worst = worst_case_usd(role, instructions, user_text)
    if not budget.reserve(worst):
        record.update(status="cap_skip", detail="the next call can pass the cap", cost_usd=0.0)
        return record
    result = None
    try:
        result = provider.plan(instructions, user_text, schema, profile, attempt)
    finally:
        cost = providers.cost_usd(result.usage, role["price_usd_per_million_tokens"]) if result else 0.0
        budget.settle(worst, cost)
    record.update(status=result.status, detail=result.detail, usage=result.usage, cost_usd=cost,
                  seconds=round(result.seconds, 2), retries=result.retries, raw_text=result.text)
    if result.status == "completed":
        plan, errors = schema_check.check_text(result.text, schema)
        record["schema_errors"] = errors
        record["schema_pass"] = not errors
        if not errors:
            record["plan"] = plan
            record["violations"] = policy.evaluate(plan, profile, machines)
    else:
        record["schema_pass"] = False
        record["schema_errors"] = [f"no plan: {result.status}"]
    return record


def summarize(records, meta):
    requested = len(records)
    returned = [r for r in records if r["status"] in RETURNED]
    valid = [r for r in records if r.get("schema_pass")]
    rejected = [r for r in valid if r["violations"]]
    open_rules = {r.rule_id for r in policy.RULES if r.questions}
    rejected_closed = [r for r in valid if any(v["rule_id"] not in open_rules for v in r["violations"])]
    total_cost = sum(r.get("cost_usd", 0.0) for r in records)
    per_rule = {}
    for rule in policy.RULES:
        hits = [r for r in valid if any(v["rule_id"] == rule.rule_id for v in r["violations"])]
        per_rule[rule.rule_id] = {
            "plans": len(hits),
            "violations": sum(1 for r in valid for v in r["violations"] if v["rule_id"] == rule.rule_id),
        }
    usage_keys = ("input_tokens", "cached_input_tokens", "output_tokens", "reasoning_tokens")
    tokens = {k: sum((r.get("usage") or {}).get(k, 0) for r in returned) for k in usage_keys}
    seconds = [r["seconds"] for r in returned if "seconds" in r]
    status_counts = {}
    for r in records:
        status_counts[r["status"]] = status_counts.get(r["status"], 0) + 1
    summary = dict(meta)
    summary.update({
        "requested": requested,
        "status_counts": status_counts,
        "returned": len(returned),
        "schema_pass": len(valid),
        "schema_pass_rate": len(valid) / requested if requested else None,
        "rejected": len(rejected),
        "rejection_rate": len(rejected) / len(valid) if valid else None,
        "rejected_without_open_question_rules": len(rejected_closed),
        "rejection_rate_without_open_question_rules": len(rejected_closed) / len(valid) if valid else None,
        "per_rule": per_rule,
        "total_cost_usd": round(total_cost, 6),
        "cost_per_plan_usd": round(total_cost / len(returned), 6) if returned else None,
        "tokens": tokens,
        "mean_seconds": round(sum(seconds) / len(seconds), 1) if seconds else None,
        "max_seconds": round(max(seconds), 1) if seconds else None,
    })
    summary["go"] = go_decision(summary)
    return summary


def go_decision(s):
    reasons = []
    if s["schema_pass_rate"] is None or s["schema_pass_rate"] < GO_BAR["schema_pass_rate_min"]:
        reasons.append("schema pass rate under the bar")
    if s["rejection_rate"] is None or s["rejection_rate"] > GO_BAR["rejection_rate_max"]:
        reasons.append("policy rejection rate over the bar")
    if s["cost_per_plan_usd"] is None or s["cost_per_plan_usd"] > GO_BAR["cost_per_plan_max_usd"]:
        reasons.append("cost per plan over the bar")
    return {"result": "no-go" if reasons else "go", "reasons": reasons, "bar": GO_BAR}


def summary_markdown(summary, records):
    pct = lambda v: "n/a" if v is None else f"{100 * v:.1f}%"
    lines = [
        f"# Luna plan spike run {summary['run_id']}",
        "",
        f"- Provider: {summary['provider']}. Model: {summary['model']}. Effort: {summary['reasoning_effort']}.",
        f"- Prompt: {summary['prompt_version']} ({summary['prompt_hash']}). Schema: {summary['schema_version']}. Policy: {summary['policy_version']}.",
        f"- Plans requested: {summary['requested']}. Status counts: {summary['status_counts']}.",
        f"- Schema pass rate: {pct(summary['schema_pass_rate'])} ({summary['schema_pass']} of {summary['requested']}).",
        f"- Policy rejection rate: {pct(summary['rejection_rate'])} ({summary['rejected']} of {summary['schema_pass']}).",
        f"- Rejection rate without the rules of open questions: {pct(summary['rejection_rate_without_open_question_rules'])}.",
        f"- Total cost: {summary['total_cost_usd']:.4f} USD. Cost per plan: {summary['cost_per_plan_usd']} USD. Cap: {summary['cap_usd']} USD.",
        f"- Tokens: {summary['tokens']}. Mean seconds: {summary['mean_seconds']}. Max seconds: {summary['max_seconds']}.",
        f"- Go bar result: {summary['go']['result']} {summary['go']['reasons']}.",
        "",
        "| Rule | Plans | Violations |",
        "|---|---|---|",
    ]
    for rule_id, counts in summary["per_rule"].items():
        lines.append(f"| {rule_id} | {counts['plans']} | {counts['violations']} |")
    lines += ["", "| Profile | Attempt | Rule | Where | Detail |", "|---|---|---|---|---|"]
    for r in records:
        for v in r.get("violations", []):
            detail = v["detail"].replace("|", "/")
            lines.append(f"| {r['profile_id']} | {r['attempt']} | {v['rule_id']} | {v['where']} | {detail} |")
    lines += ["", "| Profile | Attempt | Status | Schema errors |", "|---|---|---|---|"]
    for r in records:
        if not r.get("schema_pass"):
            errors = "; ".join(r.get("schema_errors", []))[:300].replace("|", "/")
            lines.append(f"| {r['profile_id']} | {r['attempt']} | {r['status']} | {errors} {r.get('detail', '')} |")
    return "\n".join(lines) + "\n"


def make_provider(args, roles, machines):
    if args.provider == "fake":
        return providers.FakeProvider(machines)
    config = roles["providers"]["openai"]
    return providers.OpenAIProvider(roles["roles"]["plan"], config, os.environ.get(config["api_key_env"], ""))


def parse_args(argv):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--provider", choices=["fake", "openai"], default="fake")
    parser.add_argument("--attempts", type=int, default=3, help="plans for each profile")
    parser.add_argument("--limit", type=int, default=0, help="use the first N profiles only")
    parser.add_argument("--cap-usd", type=float, default=None, help="lower the cap of roles.json")
    parser.add_argument("--workers", type=int, default=4)
    parser.add_argument("--owner-approved", action="store_true",
                        help="the owner approved this paid run at run time (D-25)")
    parser.add_argument("--out", default=None)
    return parser.parse_args(argv)


def main(argv=None):
    args = parse_args(argv)
    roles = load_json("roles.json")
    role = roles["roles"]["plan"]
    machines = load_json("machines.json")["machines"]
    profiles = load_json("profiles.json")["profiles"]
    if args.limit:
        profiles = profiles[:args.limit]
    cap = role["run_cap_usd"] if args.cap_usd is None else args.cap_usd
    if cap > role["run_cap_usd"]:
        print(f"harness: the cap {cap} USD is over the cap of roles.json ({role['run_cap_usd']} USD, D-98)", file=sys.stderr)
        return 2
    if args.provider != "fake" and not args.owner_approved:
        print("harness: a paid run needs --owner-approved, after the owner approves it at run time (D-25)", file=sys.stderr)
        return 2
    try:
        provider = make_provider(args, roles, machines)
    except ValueError as exc:
        print(f"harness: {exc}", file=sys.stderr)
        return 2
    now = datetime.datetime.now(datetime.timezone.utc)
    run_id = now.strftime("%Y%m%dT%H%M%SZ") + f"-{args.provider}"
    out = args.out or os.path.join(ROOT, ".local", "spikes", "luna_plan", run_id)
    os.makedirs(out, exist_ok=True)
    schema = schema_check.load_schema()
    budget = Budget(cap)
    jobs = [(p, a) for p in profiles for a in range(1, args.attempts + 1)]
    workers = 1 if args.provider == "fake" else max(1, args.workers)
    with ThreadPoolExecutor(max_workers=workers) as pool:
        records = list(pool.map(lambda job: run_one(provider, role, budget, schema, job[0], machines, job[1]), jobs))
    meta = {
        "run_id": run_id, "date_utc": now.isoformat(timespec="seconds"), "provider": provider.name,
        "model": role["model"], "reasoning_effort": role["reasoning_effort"],
        "prompt_version": prompt.PROMPT_VERSION, "prompt_hash": prompt.prompt_hash(),
        "schema_version": schema_check.SCHEMA_VERSION, "policy_version": policy.BUNDLE_VERSION,
        "cap_usd": cap, "attempts": args.attempts, "profiles": len(profiles),
    }
    summary = summarize(records, meta)
    with open(os.path.join(out, "plans.jsonl"), "w", encoding="utf-8") as handle:
        for r in records:
            handle.write(json.dumps(r) + "\n")
    with open(os.path.join(out, "summary.json"), "w", encoding="utf-8") as handle:
        json.dump(summary, handle, indent=1)
    with open(os.path.join(out, "summary.md"), "w", encoding="utf-8") as handle:
        handle.write(summary_markdown(summary, records))
    pct = lambda v: "n/a" if v is None else f"{100 * v:.1f}%"
    print(f"run {run_id}: {summary['requested']} plans, schema pass {pct(summary['schema_pass_rate'])}, "
          f"rejection {pct(summary['rejection_rate'])}, cost {summary['total_cost_usd']:.4f} USD "
          f"({summary['cost_per_plan_usd']} USD per plan), {summary['go']['result']}")
    print(f"results: {os.path.relpath(out, ROOT)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
