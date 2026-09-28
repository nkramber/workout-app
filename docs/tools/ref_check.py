#!/usr/bin/env python3
"""Check that every id and every repository path in a document resolves (D-6).

The documents of this repo cite an id of a register and a path in backticks.
A rename or a split leaves a citation that points at nothing, and a reader
then trusts a dead reference. The checker applies two rules:

- REF 1: a cited D- or Q- id that no register defines.
- REF 2: a path of this repo in backticks that no file and no folder holds.

The registers: docs/decisions.md defines each D- id with a table row, and
docs/questions.md defines each Q- id with a table row. The first cell of
the row starts with the id, as in "| D-10 (amended by D-12) |".

A token with the prefix "decktome:" names an id or a path of the Decktome
repository, for example decktome:D-811 or decktome:docs/decisions.md. The
checker skips it, because this checkout does not hold that repository.

REF 2 reads a path with a slash and a first part that names a top-level
entry of the checkout. A token right after the word "branch", as in
"the branch `docs/foundation-roadmap`" or "Branch: `feat/x`", is a branch
name and not a path, so the rule skips it. A bare file name is ambiguous, so the rule skips it.
The last part of the path holds a file type, or the whole path holds
lowercase letters alone. So a placeholder such as `docs/reviews/pr-<n>.md`
or `go/X` takes no rule. The first part can start with one dot, for
`.claude` and `.github`. A path resolves from the root, from the folder of
the document, from the folder above it, or as the one path of the checkout
that ends with it.

A dated record is history, and a rewrite of it falsifies the record. So the
checker reads no rule in a file that ends with a date, and none in the
hand-off archive.

Usage: python3 docs/tools/ref_check.py FILE [FILE ...]
"""
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

DECISIONS = "docs/decisions.md"
QUESTIONS = "docs/questions.md"

# A path under `.local` is the scratch of a local run, and `.gitignore` holds it.
SCRATCH = "/.local/"

# A dated record is history. A rewrite of it falsifies the record.
DATED = re.compile(r"-\d{4}-\d{2}-\d{2}\.md$|session-handoff-archive\.md$")

# A reference to the other repository. It takes no rule.
OTHER_REPO = re.compile(r"\bdecktome:[^\s`),;]*", re.I)

ID = re.compile(r"\b((?:D|Q)-\d+)\b")
ROW_ID = re.compile(r"^\|\s*((?:D|Q)-\d+)\b", re.M)
CODE = re.compile(r"`([^`\n]+)`")
FENCE = re.compile(r"^```.*?^```", re.M | re.S)
PATH = re.compile(r"^\.?[A-Za-z0-9_][\w./-]*$")
BRANCH_LABEL = re.compile(r"\bbranch(es)?:?\s*$", re.I)


def read(path):
    full = os.path.join(ROOT, path)
    if not os.path.exists(full):
        return ""
    with open(full, encoding="utf-8") as handle:
        return handle.read()


def registers(decisions, questions):
    """Return the set of ids that the two registers define."""
    known = {i for i in ROW_ID.findall(decisions) if i.startswith("D-")}
    known |= {i for i in ROW_ID.findall(questions) if i.startswith("Q-")}
    return known


def repo_paths():
    """Return the paths of the checkout, and the folders above each one."""
    paths, folders = set(), set()
    for base, names, files in os.walk(ROOT):
        rel = os.path.relpath(base, ROOT)
        if rel == ".":
            rel = ""
        names[:] = [n for n in names if n not in {".git", "node_modules", "dist", "build", "__pycache__"}]
        for name in names:
            folders.add(os.path.join(rel, name) if rel else name)
        for name in files:
            # A git worktree holds `.git` as a file. It is no path of the repo.
            if not rel and name == ".git":
                continue
            paths.add(os.path.join(rel, name) if rel else name)
    return paths, folders


def resolves(token, doc, paths, folders):
    here = os.path.dirname(doc)
    above = os.path.dirname(here)
    for base in (".", here, above):
        candidate = os.path.normpath(os.path.join(base, token))
        if candidate in paths or candidate in folders:
            return True
    tail = "/" + token
    return any(p.endswith(tail) for p in paths) or any(f.endswith(tail) for f in folders)


def check(doc, text, known, top, paths, folders):
    """Return the findings of one document as (line, rule, message)."""
    findings = []
    clean = FENCE.sub(lambda m: "\n" * m.group(0).count("\n"), text)
    for number, line in enumerate(clean.splitlines(), start=1):
        line = OTHER_REPO.sub(" ", line)
        for match in ID.finditer(line):
            name = match.group(1)
            if name in known:
                continue
            findings.append((number, "REF 1", f"no register defines {name}"))
        for match in CODE.finditer(line):
            token = match.group(1).strip().rstrip("/")
            if "/" not in token or not PATH.match(token):
                continue
            if token.split("/")[0] not in top:
                continue
            if BRANCH_LABEL.search(line[:match.start()]):
                continue
            if SCRATCH in "/" + token:
                continue
            if "." not in token.rsplit("/", 1)[1] and token != token.lower():
                continue
            if resolves(token, doc, paths, folders):
                continue
            findings.append((number, "REF 2", f"no file and no folder holds `{token}`"))
    return findings


def main(argv):
    known = registers(read(DECISIONS), read(QUESTIONS))
    if not known:
        print("ref_check: the registers hold no id. Run this from the checkout")
        return 1
    paths, folders = repo_paths()
    top = {name.split("/")[0] for name in paths | folders}
    total = 0
    for doc in argv:
        if DATED.search(doc):
            continue
        for number, rule, message in check(doc, read(doc), known, top, paths, folders):
            print(f"{doc}:{number}: rule {rule}: {message}")
            total += 1
    print(f"{total} finding(s)")
    return 1 if total else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
