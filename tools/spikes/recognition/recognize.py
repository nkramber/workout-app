#!/usr/bin/env python3
"""The harness of the recognition spike (work area 1.2, D-100).

It sends each photo of the recognition test set with the catalog shortlist
to the recognize role, one photo in each call. It checks each answer against
the answer schema, scores it against the true machine type, and writes the
results and a summary with the go result of D-107.

Run download.py and degrade.py of tools/spikes/recognition_set first. The
harness prepares each photo with Pillow, a local tool.

Free run, with the fake provider:
    python3 tools/spikes/recognition/recognize.py

Paid run. The owner approves it at run time (D-25), and the cap is D-94:
    python3 tools/spikes/recognition/recognize.py --provider openai --owner-approved

The results go to .local/spikes/recognition/<run id>/, which Git ignores.
"""
import argparse
import datetime
import json
import os
import sys
from concurrent.futures import ThreadPoolExecutor

import photos
import vision
import harness as plan_harness  # tools/spikes/luna_plan: the Budget and the Gate
import providers  # tools/spikes/luna_plan
import schema_check  # tools/spikes/luna_plan

ROOT = os.path.abspath(os.path.join(photos.HERE, "..", "..", ".."))
# The go bar of the recognition risk (D-107). Each rate is out of all the
# photos of the run.
GO_BAR = {"high_confidence_min": 0.8, "high_confidence_wrong_rate_max": 0.02,
          "correct_rate_min": 0.70, "cost_per_photo_max_usd": 0.01}
RETURNED = {"completed", "incomplete", "refusal"}
OUTCOMES = ("correct", "wrong", "abstained", "invalid")
CONFIDENCE_BINS = ((0.0, 0.5), (0.5, 0.8), (0.8, 0.9), (0.9, 1.01))


def load_json(name):
    with open(os.path.join(photos.HERE, name), encoding="utf-8") as handle:
        return json.load(handle)


def check_answer(text, schema):
    """Return (answer, errors). The schema check of the plan spike reads the
    keywords of the schema. This function adds the range of the confidence."""
    answer, errors = schema_check.check_text(text, schema)
    if not errors and not 0 <= answer["confidence"] <= 1:
        errors = [f"$.confidence: {answer['confidence']} is not from 0 to 1"]
    return answer, errors


def score(machine_type, truth):
    """Score one valid answer against the true machine type.

    A photo of a catalog type: the true type is correct, an abstention is
    abstained, and another type is wrong. A photo of a machine outside the
    catalog: an abstention is correct, and any catalog type is wrong."""
    if truth in vision.ABSTAIN:
        return "correct" if machine_type in vision.ABSTAIN else "wrong"
    if machine_type == truth:
        return "correct"
    return "abstained" if machine_type in vision.ABSTAIN else "wrong"


def worst_case_usd(role, instructions_text):
    prices = role["price_usd_per_million_tokens"]
    input_tokens = (len(instructions_text) + 200) // 2 + role["image_tokens_max"]
    return (input_tokens * prices["input"] + role["max_output_tokens"] * prices["output"]) / 1_000_000


def run_one(provider, role, budget, schema, instructions_text, photo, loader):
    record = {k: photo.get(k) for k in ("photo_id", "group", "degrade", "source_id", "split", "gym",
                                         "hard_negative")}
    record["truth"] = photo["machine_type"]
    image = loader(photo["path"], role["image_long_side_px"], role["image_jpeg_quality"])
    record["image_bytes"] = len(image)
    gate = plan_harness.Gate(budget, worst_case_usd(role, instructions_text), role["price_usd_per_million_tokens"])
    result = provider.recognize(instructions_text, photo, image, schema, gate)
    record.update(status=result.status, detail=result.detail, usage=result.usage, cost_usd=gate.known_usd,
                  unknown_charges=gate.unknown_charges, seconds=round(result.seconds, 2),
                  retries=result.retries, raw_text=result.text)
    answer, errors = check_answer(result.text, schema) if result.status == "completed" else (
        None, [f"no answer: {result.status}"])
    record["errors"] = errors
    if errors:
        record.update(outcome="invalid", answer=None, confidence=None, high_confidence=False)
    else:
        record.update(outcome=score(answer["machine_type"], photo["machine_type"]),
                      answer=answer["machine_type"], confidence=answer["confidence"],
                      high_confidence=answer["confidence"] >= GO_BAR["high_confidence_min"],
                      quality_flags=answer["quality_flags"], observed_text=answer["observed_text"],
                      evidence=answer["evidence"])
    return record


def outcome_counts(records):
    counts = {k: 0 for k in OUTCOMES}
    for r in records:
        counts[r["outcome"]] += 1
    return counts


def calibration(records):
    """For each confidence bin: the answers that name a catalog type, and the
    share of them that is correct."""
    rows = []
    named = [r for r in records if r["outcome"] != "invalid" and r["answer"] not in vision.ABSTAIN]
    for low, high in CONFIDENCE_BINS:
        hits = [r for r in named if low <= r["confidence"] < high]
        correct = sum(1 for r in hits if r["outcome"] == "correct")
        rows.append({"bin": f"{low:.1f}-{min(high, 1.0):.1f}", "answers": len(hits), "correct": correct,
                     "precision": round(correct / len(hits), 3) if hits else None})
    return rows


def summarize(records, meta):
    requested = len(records)
    returned = [r for r in records if r["status"] in RETURNED]
    counts = outcome_counts(records)
    high_wrong = [r for r in records if r["outcome"] == "wrong" and r["high_confidence"]]
    total_cost = sum(r.get("cost_usd", 0.0) for r in records)
    usage_keys = ("input_tokens", "cached_input_tokens", "output_tokens", "reasoning_tokens")
    tokens = {k: sum((r.get("usage") or {}).get(k, 0) for r in returned) for k in usage_keys}
    inputs = [r["usage"]["input_tokens"] for r in returned if (r.get("usage") or {}).get("input_tokens")]
    seconds = [r["seconds"] for r in returned]
    status_counts = {}
    for r in records:
        status_counts[r["status"]] = status_counts.get(r["status"], 0) + 1
    rate = lambda n: n / requested if requested else None
    summary = dict(meta)
    summary.update({
        "requested": requested,
        "status_counts": status_counts,
        "returned": len(returned),
        "outcomes": counts,
        "correct_rate": rate(counts["correct"]),
        "wrong_rate": rate(counts["wrong"]),
        "abstained_rate": rate(counts["abstained"]),
        "high_confidence_wrong": len(high_wrong),
        "high_confidence_wrong_rate": rate(len(high_wrong)),
        "by_group": {g: outcome_counts([r for r in records if r["group"] == g]) for g in photos.GROUPS},
        "by_truth": {
            "catalog": outcome_counts([r for r in records if r["truth"] not in vision.ABSTAIN]),
            "outside": outcome_counts([r for r in records if r["truth"] in vision.ABSTAIN]),
        },
        "by_split": {s: outcome_counts([r for r in records if r["split"] == s])
                     for s in sorted({r["split"] for r in records})},
        "hard_negatives": outcome_counts([r for r in records if r["hard_negative"]]),
        "calibration": calibration(records),
        "high_confidence_wrong_answers": [
            {k: r.get(k) for k in ("photo_id", "group", "degrade", "truth", "answer", "confidence", "hard_negative")}
            for r in high_wrong],
        "total_cost_usd": round(total_cost, 6),
        "unknown_charges": sum(r.get("unknown_charges", 0) for r in records),
        "cost_per_photo_usd": round(total_cost / len(returned), 6) if returned else None,
        "tokens": tokens,
        "input_tokens_per_photo": ({"min": min(inputs), "mean": round(sum(inputs) / len(inputs)), "max": max(inputs)}
                                   if inputs else None),
        "mean_seconds": round(sum(seconds) / len(seconds), 1) if seconds else None,
        "max_seconds": round(max(seconds), 1) if seconds else None,
    })
    summary["go"] = go_decision(summary)
    return summary


def go_decision(s):
    reasons = []
    if s["high_confidence_wrong_rate"] is None or s["high_confidence_wrong_rate"] > GO_BAR["high_confidence_wrong_rate_max"]:
        reasons.append("high-confidence wrong rate over the bar")
    if s["correct_rate"] is None or s["correct_rate"] < GO_BAR["correct_rate_min"]:
        reasons.append("correct rate under the bar")
    if s["cost_per_photo_usd"] is None or s["cost_per_photo_usd"] > GO_BAR["cost_per_photo_max_usd"]:
        reasons.append("cost per photo over the bar")
    return {"result": "no-go" if reasons else "go", "reasons": reasons, "bar": GO_BAR}


def summary_markdown(summary, records):
    pct = lambda v: "n/a" if v is None else f"{100 * v:.1f}%"
    o = summary["outcomes"]
    lines = [
        f"# Recognition spike run {summary['run_id']}",
        "",
        f"- Provider: {summary['provider']}. Model: {summary['model']}. Effort: {summary['reasoning_effort']}. Image detail: {summary['image_detail']}.",
        f"- Prompt: {summary['prompt_version']} ({summary['prompt_hash']}). Catalog: version {summary['catalog_version']}, {summary['catalog_types']} types.",
        f"- Photos: {summary['requested']}. Status counts: {summary['status_counts']}.",
        f"- Correct: {o['correct']} ({pct(summary['correct_rate'])}). Wrong: {o['wrong']} ({pct(summary['wrong_rate'])}). "
        f"Abstained: {o['abstained']} ({pct(summary['abstained_rate'])}). Invalid: {o['invalid']}.",
        f"- Wrong with a stated confidence of {GO_BAR['high_confidence_min']} or more: {summary['high_confidence_wrong']} ({pct(summary['high_confidence_wrong_rate'])}).",
        f"- Total cost: {summary['total_cost_usd']:.4f} USD. Cost per photo: {summary['cost_per_photo_usd']} USD. Cap: {summary['cap_usd']} USD.",
        f"- Input tokens per photo: {summary['input_tokens_per_photo']}. Tokens: {summary['tokens']}.",
        f"- Attempts with an unknown charge: {summary['unknown_charges']}. Spend bound with each at its worst case: {summary['spend_bound_usd']:.4f} USD.",
        f"- Mean seconds: {summary['mean_seconds']}. Max seconds: {summary['max_seconds']}.",
        f"- Go bar result: {summary['go']['result']} {summary['go']['reasons']}.",
        "",
        "| Photos | Correct | Wrong | Abstained | Invalid |",
        "|---|---|---|---|---|",
    ]
    rows = [(f"group {k}", v) for k, v in summary["by_group"].items()]
    rows += [(f"truth {k}", v) for k, v in summary["by_truth"].items()]
    rows += [(f"split {k}", v) for k, v in summary["by_split"].items()]
    rows.append(("hard negatives", summary["hard_negatives"]))
    for name, c in rows:
        lines.append(f"| {name} | {c['correct']} | {c['wrong']} | {c['abstained']} | {c['invalid']} |")
    lines += ["", "| Stated confidence | Answers with a catalog type | Correct | Precision |", "|---|---|---|---|"]
    for row in summary["calibration"]:
        lines.append(f"| {row['bin']} | {row['answers']} | {row['correct']} | {row['precision']} |")
    lines += ["", "| Photo | Group | Truth | Answer | Confidence | Outcome |", "|---|---|---|---|---|---|"]
    for r in records:
        if r["outcome"] in ("wrong", "invalid"):
            group = r["group"] if not r["degrade"] else f"{r['group']} {r['degrade']}"
            detail = r["answer"] or "; ".join(r["errors"])[:200].replace("|", "/")
            lines.append(f"| {r['photo_id']} | {group} | {r['truth']} | {detail} | {r['confidence']} | {r['outcome']} |")
    return "\n".join(lines) + "\n"


def make_provider(args, roles, catalog):
    if args.provider == "fake":
        return vision.FakeProvider(catalog)
    config = roles["providers"]["openai"]
    return vision.OpenAIProvider(roles["roles"]["recognize"], config, os.environ.get(config["api_key_env"], ""))


def select(all_photos, args):
    chosen = all_photos
    if args.photos:
        wanted = set(args.photos.split(","))
        chosen = [p for p in chosen if p["photo_id"] in wanted]
    if args.limit:
        chosen = chosen[:args.limit]
    return chosen


def parse_args(argv):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--provider", choices=["fake", "openai"], default="fake")
    parser.add_argument("--photos", default="", help="a comma list of photo ids, for a smoke call")
    parser.add_argument("--limit", type=int, default=0, help="use the first N photos only")
    parser.add_argument("--cap-usd", type=float, default=None, help="lower the cap of roles.json")
    parser.add_argument("--workers", type=int, default=4)
    parser.add_argument("--owner-approved", action="store_true",
                        help="the owner approved this paid run at run time (D-25)")
    parser.add_argument("--out", default=None)
    return parser.parse_args(argv)


def main(argv=None, loader=photos.prepare, cache=photos.download.CACHE):
    args = parse_args(argv)
    roles = load_json("roles.json")
    role = roles["roles"]["recognize"]
    catalog, manifest = photos.load_set()
    cap = role["run_cap_usd"] if args.cap_usd is None else args.cap_usd
    if cap > role["run_cap_usd"]:
        print(f"recognize: the cap {cap} USD is over the cap of roles.json ({role['run_cap_usd']} USD, D-94)", file=sys.stderr)
        return 2
    if args.provider != "fake" and not args.owner_approved:
        print("recognize: a paid run needs --owner-approved, after the owner approves it at run time (D-25)", file=sys.stderr)
        return 2
    chosen = select(photos.photo_list(manifest, cache), args)
    missing = [p["photo_id"] for p in chosen if not os.path.exists(p["path"])]
    if missing:
        print(f"recognize: {len(missing)} photos are not in the cache, for example {missing[0]}. "
              "Run download.py and degrade.py of tools/spikes/recognition_set.", file=sys.stderr)
        return 2
    try:
        provider = make_provider(args, roles, catalog)
    except ValueError as exc:
        print(f"recognize: {exc}", file=sys.stderr)
        return 2
    now = datetime.datetime.now(datetime.timezone.utc)
    run_id = now.strftime("%Y%m%dT%H%M%SZ") + f"-{args.provider}"
    out = args.out or os.path.join(ROOT, ".local", "spikes", "recognition", run_id)
    os.makedirs(out, exist_ok=True)
    schema = vision.answer_schema(catalog)
    text = vision.instructions(catalog)
    budget = plan_harness.Budget(cap)
    workers = 1 if args.provider == "fake" else max(1, args.workers)
    with ThreadPoolExecutor(max_workers=workers) as pool:
        records = list(pool.map(lambda p: run_one(provider, role, budget, schema, text, p, loader), chosen))
    meta = {
        "run_id": run_id, "date_utc": now.isoformat(timespec="seconds"), "provider": provider.name,
        "model": role["model"], "reasoning_effort": role["reasoning_effort"], "image_detail": role["image_detail"],
        "image_long_side_px": role["image_long_side_px"], "prompt_version": vision.PROMPT_VERSION,
        "prompt_hash": vision.prompt_hash(catalog), "catalog_version": catalog["version"],
        "catalog_types": len(catalog["types"]), "cap_usd": cap,
    }
    summary = summarize(records, meta)
    summary["spend_bound_usd"] = round(budget.spent, 6)
    with open(os.path.join(out, "answers.jsonl"), "w", encoding="utf-8") as handle:
        for r in records:
            handle.write(json.dumps(r) + "\n")
    with open(os.path.join(out, "summary.json"), "w", encoding="utf-8") as handle:
        json.dump(summary, handle, indent=1)
    with open(os.path.join(out, "summary.md"), "w", encoding="utf-8") as handle:
        handle.write(summary_markdown(summary, records))
    pct = lambda v: "n/a" if v is None else f"{100 * v:.1f}%"
    o = summary["outcomes"]
    print(f"run {run_id}: {summary['requested']} photos, correct {o['correct']} ({pct(summary['correct_rate'])}), "
          f"wrong {o['wrong']}, abstained {o['abstained']}, invalid {o['invalid']}, "
          f"high-confidence wrong {summary['high_confidence_wrong']} ({pct(summary['high_confidence_wrong_rate'])}), "
          f"cost {summary['total_cost_usd']:.4f} USD ({summary['cost_per_photo_usd']} USD per photo), {summary['go']['result']}")
    print(f"results: {os.path.relpath(out, ROOT)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
