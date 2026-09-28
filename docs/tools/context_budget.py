#!/usr/bin/env python3
"""Check the size of the files a session reads at start (D-6).

Every session reads AGENTS.md and the hand-off first, and Claude Code
loads CLAUDE.md into every call. Each byte of them sits in the context of
every later call. A skill loads whole into the context of the session that
reads it. The check is a port of the Decktome check.

The check fails when:
- AGENTS.md, CLAUDE.md, or the hand-off passes its byte limit, or is absent,
- the `## Resume here` section of the hand-off passes its byte limit,
- the hand-off holds more session records than its limit. Each record
  holds one `Author provider:` line, so the check counts those lines,
- a SKILL.md file of .claude/skills passes its byte limit,
- a Makefile target whose help text says CAUTION is absent from the
  paid-target sentence of AGENTS.md, or that sentence names a target that
  is not a paid target of the Makefile.

The paid-target sentence is the first sentence of AGENTS.md that holds the
words "paid target" (any case). It names each paid target in backticks, for
example: The one paid target is `make codex-review`.

Run: python3 docs/tools/context_budget.py
"""
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

AGENTS = "AGENTS.md"
CLAUDE = "CLAUDE.md"
HANDOFF = "docs/session-handoff.md"
MAKEFILE = "Makefile"

FILE_LIMITS = {AGENTS: 11000, CLAUDE: 1000, HANDOFF: 24000}

SKILLS = ".claude/skills"
SKILL_LIMIT = 36864
RESUME_HEADING = "## Resume here"
RESUME_LIMIT = 6000
SESSIONS_LIMIT = 3
PROVIDER = re.compile(r"^\s*[-*]?\s*Author provider:", re.M)

PAID_MARKER = re.compile(r"paid target", re.I)
PAID_NAME = re.compile(r"`make ([a-z0-9-]+)[^`]*`")
CAUTION_TARGET = re.compile(r"^([a-zA-Z0-9_-]+):[^\n#]*##[^\n]*\bCAUTION\b", re.M)
TARGET = re.compile(r"^([a-zA-Z0-9_-]+):", re.M)


def section(text, heading):
    """Return the text from the '## ' heading that starts with `heading` to the next '## ' heading."""
    out, inside = [], False
    for line in text.splitlines(keepends=True):
        if line.startswith("## "):
            if inside:
                break
            inside = line.lower().startswith(heading.lower())
        if inside:
            out.append(line)
    return "".join(out) if out else None


def paid_names(text):
    """Return the make targets that the first paid-target sentence names, or None."""
    clean = re.sub(r"```.*?```", " ", text, flags=re.S)
    for block in re.split(r"\n\s*\n|\n(?=\s*[-*] )", clean):
        for sentence in re.split(r"(?<=[.!?])\s+", " ".join(block.split())):
            if PAID_MARKER.search(sentence):
                return set(PAID_NAME.findall(sentence))
    return None


def caution_targets(makefile):
    return set(CAUTION_TARGET.findall(makefile))


def check(read, skill_files=()):
    """Return (report lines, errors). `read` maps a path to its text or None.

    `skill_files` names every SKILL.md file of the skills folder.
    """
    report, errors = [], []
    texts = {path: read(path) for path in (AGENTS, CLAUDE, HANDOFF, MAKEFILE)}
    for path, limit in FILE_LIMITS.items():
        text = texts[path]
        if text is None:
            errors.append(f"{path} does not exist")
            continue
        size = len(text.encode("utf-8"))
        report.append(f"{path}: {size} of {limit} bytes")
        if size > limit:
            errors.append(f"{path} holds {size} bytes, over its limit of {limit}. Move the detail to a document that the start read does not hold")
    handoff = texts[HANDOFF]
    if handoff is not None:
        resume = section(handoff, RESUME_HEADING)
        if resume is None:
            errors.append(f"{HANDOFF} holds no '{RESUME_HEADING}' section")
        else:
            size = len(resume.encode("utf-8"))
            report.append(f"{HANDOFF} resume section: {size} of {RESUME_LIMIT} bytes")
            if size > RESUME_LIMIT:
                errors.append(f"the resume section holds {size} bytes, over its limit of {RESUME_LIMIT}. Keep the state and the next step, and link the detail")
        count = len(PROVIDER.findall(handoff))
        report.append(f"{HANDOFF} session records: {count} of {SESSIONS_LIMIT}")
        if count > SESSIONS_LIMIT:
            errors.append(f"the hand-off holds {count} session records, over its limit of {SESSIONS_LIMIT}. Remove the oldest record. Git keeps its history")
    makefile = texts[MAKEFILE] or ""
    paid = caution_targets(makefile)
    if texts[AGENTS] is not None:
        named = paid_names(texts[AGENTS])
        if named is None:
            errors.append(f"{AGENTS} holds no sentence with the words 'paid target' that names the paid targets")
        else:
            for name in sorted(paid - named):
                errors.append(f"the Makefile help of `make {name}` says CAUTION, and the paid-target sentence of {AGENTS} does not name it")
            targets = set(TARGET.findall(makefile))
            for name in sorted(named - paid):
                if name not in targets:
                    errors.append(f"the paid target `make {name}` of {AGENTS} is no Makefile target")
                else:
                    errors.append(f"{AGENTS} names `make {name}` as a paid target, and its Makefile help says no CAUTION")
            report.append(f"paid targets: {len(paid)} in the Makefile, {len(named)} in {AGENTS}")
    over = []
    for path in sorted(skill_files):
        size = len((read(path) or "").encode("utf-8"))
        if size > SKILL_LIMIT:
            over.append(path)
            errors.append(f"{path} holds {size} bytes, over its limit of {SKILL_LIMIT}. Move the detail to a file of the references folder")
    report.append(f"skill files: {len(skill_files)} read, {len(over)} over {SKILL_LIMIT} bytes")
    return report, errors


def main():
    def read(path):
        full = os.path.join(ROOT, path)
        if not os.path.exists(full):
            return None
        with open(full, encoding="utf-8") as handle:
            return handle.read()

    skills = []
    for base, _, files in os.walk(os.path.join(ROOT, SKILLS)):
        if "SKILL.md" in files:
            skills.append(os.path.relpath(os.path.join(base, "SKILL.md"), ROOT))

    report, errors = check(read, skills)
    for line in report:
        print(f"context_budget: {line}")
    for error in errors:
        print(f"context_budget: {error}")
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
