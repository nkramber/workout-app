#!/usr/bin/env python3
"""Compare the live merge rules of `main` with the files of the repo (D-6, D-13, D-14).

  ruleset_check.py [--repo DIR]

`.github/rulesets/review-gate.json` is the body of the ruleset, and
`.github/rulesets/merge-settings.json` holds the merge settings of the
repository. The owner applies both after the merge that adds them:

  gh api -X POST repos/OWNER/REPO/rulesets --input .github/rulesets/review-gate.json
  gh api -X PATCH repos/OWNER/REPO --input .github/rulesets/merge-settings.json

A later change of the ruleset uses PUT on repos/OWNER/REPO/rulesets/ID.

The check finds the live ruleset by its name, so the file holds no id. It
reads both through `gh api`, so it needs the network, and CI does not run
it. Each key of a file must equal its live value. A list of status checks
compares as a set. A live parameter that the file does not name passes,
because GitHub adds new parameters with their defaults. A live rule that
the file does not name fails.

Exit 0 when each value matches, 1 on a difference or no live ruleset, and
2 on an error. The tool is a port of the Decktome check.
"""
import argparse
import json
import os
import subprocess
import sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", ".."))
RULESET = ".github/rulesets/review-gate.json"
SETTINGS = ".github/rulesets/merge-settings.json"


def checks_key(check):
    return check.get("context"), check.get("integration_id")


def compare(want, live, where="ruleset"):
    """Each difference between a file value and its live value, as text."""
    diffs = []
    if isinstance(want, dict):
        if not isinstance(live, dict):
            return [f"{where}: the file gives an object, and GitHub gives {json.dumps(live)}."]
        for key, value in want.items():
            if key not in live:
                diffs.append(f"{where}.{key}: GitHub gives no value, and the file gives {json.dumps(value)}.")
            elif key == "required_status_checks" and isinstance(value, list):
                have = {checks_key(c) for c in live[key]}
                need = {checks_key(c) for c in value}
                for context, app in sorted(need - have, key=str):
                    diffs.append(f"{where}.{key}: GitHub does not require `{context}` from app {app}.")
                for context, app in sorted(have - need, key=str):
                    diffs.append(f"{where}.{key}: GitHub requires `{context}` from app {app}, and the file does not.")
            elif key == "rules":
                diffs += compare_rules(value, live[key], f"{where}.rules")
            else:
                diffs += compare(value, live[key], f"{where}.{key}")
        return diffs
    if want != live:
        return [f"{where}: the file gives {json.dumps(want)}, and GitHub gives {json.dumps(live)}."]
    return []


def compare_rules(want, live, where):
    diffs = []
    live_by_type = {r["type"]: r for r in live}
    want_by_type = {r["type"]: r for r in want}
    for kind, rule in want_by_type.items():
        if kind not in live_by_type:
            diffs.append(f"{where}: GitHub has no `{kind}` rule.")
        else:
            diffs += compare(rule.get("parameters", {}), live_by_type[kind].get("parameters", {}), f"{where}.{kind}")
    for kind in live_by_type:
        if kind not in want_by_type:
            diffs.append(f"{where}: GitHub has a `{kind}` rule that the file does not name.")
    return diffs


def find_ruleset(listing, name):
    """The id of the one branch ruleset with this name, or None. Two with the name is an error."""
    found = [r for r in listing if r.get("name") == name and r.get("target", "branch") == "branch"]
    if len(found) > 1:
        raise SystemExit(f"ruleset_check: GitHub holds {len(found)} rulesets named `{name}`. Keep one.")
    return found[0]["id"] if found else None


def gh(*args):
    out = subprocess.run(["gh", *args], capture_output=True, text=True)
    if out.returncode != 0:
        raise SystemExit(f"ruleset_check: gh {' '.join(args)} failed: {out.stderr.strip()}")
    return json.loads(out.stdout)


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--repo", default=ROOT)
    args = parser.parse_args()
    with open(os.path.join(args.repo, RULESET), encoding="utf-8") as handle:
        ruleset = json.load(handle)
    with open(os.path.join(args.repo, SETTINGS), encoding="utf-8") as handle:
        settings = json.load(handle)
    slug = subprocess.run(["gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner"],
                          cwd=args.repo, capture_output=True, text=True).stdout.strip()
    if not slug:
        print("ruleset_check: gh can not name the repository.", file=sys.stderr)
        return 2
    listing = gh("api", "--paginate", "--slurp", f"repos/{slug}/rulesets")
    ruleset_id = find_ruleset([r for page in listing for r in page], ruleset["name"])
    if ruleset_id is None:
        print(f"ruleset-check: GitHub holds no ruleset named `{ruleset['name']}`. The owner applies {RULESET}.")
        return 1
    diffs = compare(ruleset, gh("api", f"repos/{slug}/rulesets/{ruleset_id}"), "ruleset")
    diffs += compare(settings, gh("api", f"repos/{slug}"), "repository")
    for line in diffs:
        print(line)
    print("ruleset-check: the live rules match the files." if not diffs else f"ruleset-check: {len(diffs)} difference(s).")
    return 1 if diffs else 0


if __name__ == "__main__":
    sys.exit(main())
