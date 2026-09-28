#!/usr/bin/env python3
"""Check Markdown files against the ASD-STE100 rules that a script can test.

The checker applies these rules (D-83). It is a port of the Decktome checker,
and the rules are the same. The word lists name the technical names of this repo.
- 3.2/3.4: no modal verbs and no perfect tenses ("should", "has been").
- 3.5: no "-ing" verb form at the start of a sentence, after a helper word,
  or after a preposition.
- 3.6: no passive voice ("is stored", "was written").
- 4.2: no contractions.
- 5.1: max 20 words in a sentence of a numbered step.
- 6.3: max 25 words in a descriptive sentence.
- 6.6: max six sentences in a paragraph.
- 8.1: no semicolons.

The checker skips tables, code blocks, headings, front matter, inline code,
and URLs. A comma list of technical names (rules 4.3 and 8.6) is exempt
from the length rules. The words "can", "must", and "will" are allowed.
A past participle in the ALLOW_STATE list counts as an adjective (rule 3.3).

Usage: python3 docs/tools/ste-check.py [--rules] FILE [FILE ...]
"""
import re
import sys

RULES = {
    "3.2/3.4": "modal verb or perfect tense",
    "3.5": "-ing verb form",
    "3.6": "passive voice",
    "4.2": "contraction",
    "5.1": "over 20 words in a numbered step",
    "6.3": "over 25 words in a sentence",
    "6.6": "over six sentences in a paragraph",
    "8.1": "semicolon",
}

CONTRACTIONS = re.compile(r"\b(\w+n't|\w+'(re|ve|ll|d|m)|it's|let's|that's|there's|what's|here's)\b", re.I)

IRREGULAR = {
    "written", "read", "built", "made", "set", "sent", "kept", "held", "run", "done",
    "given", "taken", "found", "seen", "known", "shown", "chosen", "put", "cut", "left",
    "lost", "met", "paid", "said", "told", "thought", "brought", "bought", "caught",
    "taught", "fought", "sought", "won", "begun", "sung", "drawn", "grown", "thrown",
    "broken", "spoken", "frozen", "stolen", "driven", "hidden", "ridden", "forgotten",
    "gotten", "bitten", "eaten", "fallen", "risen", "beaten", "blown", "flown", "torn",
    "worn", "born", "sworn", "understood", "withheld", "upheld", "split", "spread",
    "shut", "hit", "let", "bet", "cost", "hurt", "quit", "fed", "led", "bred", "sped",
    "lit", "slid", "struck", "stuck", "swung", "hung", "dug", "spun", "wound", "bound",
    "ground", "meant", "dealt", "felt", "dreamt", "learnt", "burnt", "leant", "spelt",
    "smelt", "spilt", "spoilt", "laid", "rebuilt", "reset", "rerun", "overwritten",
    "undone", "redone", "unset", "reread", "resent", "withdrawn", "overridden", "sold",
    "misspelt", "lent", "bent", "spent", "sat", "stood", "become", "come", "gone",
}
# Past participles that name a state in this repo. They are adjectives (rule 3.3).
ALLOW_STATE = {
    "done", "gone", "over", "ready", "broken", "frozen", "retired", "merged", "pinned",
    "owned", "locked", "banned", "legal", "known", "unknown", "verified", "unverified",
    "scored", "dated", "exempt", "open", "closed", "empty", "full", "set", "green",
    "red", "left", "right", "flat", "clean", "dirty", "stale", "stuck", "wrong", "complete",
    "incomplete", "present", "absent", "missing", "unavailable", "available", "idle",
    "unattended", "identical", "silent", "aligned", "interested", "dead", "alive",
    "correct", "incorrect", "wired", "unwired", "worth", "sure", "unsure", "related",
    "unrelated", "sound", "bound", "supported", "unsupported", "authenticated",
    "expected", "unexpected", "used", "unused", "installed", "uninstalled", "cached",
    "outdated", "limited", "unlimited", "finished", "unfinished", "logged", "welcome",
    "detailed", "advanced", "fixed", "tied", "united", "rooted", "based", "sized",
    "colored", "named", "numbered", "dated", "signed", "enabled", "disabled", "blocked",
    "defined", "undefined", "defined", "unchanged", "unread", "untouched", "unresolved",
    "resolved", "committed", "uncommitted", "tracked", "untracked", "unmerged", "unscored",
}
# Words that end in "ed" and are never a past participle.
NOT_PARTICIPLE = {
    "need", "seed", "feed", "speed", "breed", "bleed", "proceed", "succeed", "exceed",
    "indeed", "agreed", "freed", "red", "bed", "shed", "wed", "sled", "fled", "hundred",
    "naked", "wicked", "sacred", "unlimited", "rugged", "wretched", "crooked", "jagged",
    "beloved", "biased", "coed", "med", "greed", "reed", "deed", "creed", "steed", "weed",
    "tweed", "ahead", "instead", "dead", "lead", "read", "bread", "thread", "spread",
}
BE = r"(is|are|was|were|be|been|being|am)"
PASSIVE = re.compile(
    r"\b" + BE + r"\s+(?:(?:not|also|then|now|never|always|still|only|often|already|both|all|each|first|last|later|again|usually|fully|partly|either|neither|just|rarely|soon|thus|so|well|hence)\s+){0,2}([a-z]+)\b",
    re.I,
)
PERFECT = re.compile(
    r"\b(has|have|had)\s+(?:(?:not|also|never|always|already|since|just|only|both|all|each|often|ever|now|thus|so)\s+){0,2}([a-z]+)\b",
    re.I,
)
MODAL = re.compile(r"\b(should|would|could|might|may|shall|ought)\b", re.I)

ING_AFTER = re.compile(
    r"\b(is|are|was|were|be|been|being|am|by|of|for|from|before|after|while|when|without|on|in|at|to|than|worth|start|starts|started|keep|keeps|kept|stop|stops|stopped|avoid|avoids|allow|allows|about|through|until|via)\s+(\w+ing)\b",
    re.I,
)
ING_START = re.compile(r"^(\w+ing)\b")
# Technical names and nouns that end in "ing". They are not verb forms.
ING_ALLOW = {
    "thing", "nothing", "something", "anything", "everything", "during", "string",
    "ring", "king", "bring", "spring", "ping", "wing", "sing", "morning", "evening",
    "sibling", "sling", "swing", "fling", "cling", "sting", "wring", "ceiling", "sterling",
    "meaning", "warning", "setting", "settings", "ranking", "rankings", "logging",
    "caching", "streaming", "routing", "rendering", "scoring", "pricing", "scaling",
    "engineering", "tagging", "listing", "listings", "ordering", "spelling", "heading",
    "headings", "sizing", "mapping", "mappings", "binding", "bindings", "batching",
    "polling", "tuning", "linting", "tooling", "styling", "formatting", "monitoring",
    "alerting", "encoding", "encodings", "embedding", "embeddings", "chunking", "wiring",
    "wording", "timing", "timings", "sampling", "stemming", "finding", "findings",
    "grouping", "groupings", "pairing", "pairings", "nesting", "spacing", "docstring",
    "docstrings", "lightning", "hosting", "messaging", "onboarding", "tracking",
    # Fitness nouns and adjectives of this repo. They are technical names (rule 1.5).
    "training", "rounding", "loading", "deloading", "rowing", "running", "stretching",
    "programming", "screening", "rating", "ratings", "cycling", "swimming", "walking",
    "hiking", "jogging", "sprinting", "lifting", "weightlifting", "powerlifting",
    "bodybuilding", "conditioning", "pacing", "breathing", "spotting",
    "missing", "existing", "following", "remaining", "underlying", "leading", "trailing",
    "pending", "outstanding", "according", "including", "excluding", "regarding",
}
MAX_WORDS = 25
MAX_STEP_WORDS = 20


def strip_md(line: str) -> str:
    # Curly quotes count like straight ones (rule 8.6), and a curly
    # apostrophe hides a contraction from the check.
    line = line.replace("’", "'").replace("‘", "'").replace("“", '"').replace("”", '"')
    line = re.sub(r"`[^`]*`", "X", line)
    line = re.sub(r"\[([^\]]*)\]\([^)]*\)", r"\1", line)
    line = re.sub(r"https?://\S+", "URL", line)
    line = re.sub(r'"[^"]*"', "QUOTE", line)  # rule 8.6: quoted text counts as one word
    line = re.sub(r"\([^)]*\)", "(X)", line)  # rule 8.5: parentheses count as one word
    line = re.sub(r"[*_>#]+", "", line)
    return line


def is_name_list(s: str) -> bool:
    """A run of short comma-separated items is a list of names, not a sentence."""
    items = [i.strip() for i in s.split(",")]
    return len(items) >= 6 and sum(len(i.split()) for i in items) / len(items) <= 3


def sentences(text: str):
    parts = re.split(r"(?<=[.!?])\s+(?=[A-Z0-9\"'(])", text)
    return [p.strip() for p in parts if p.strip()]


def is_participle(word: str) -> bool:
    w = word.lower()
    if w in NOT_PARTICIPLE:
        return False
    if w in IRREGULAR:
        return True
    return len(w) > 3 and w.endswith("ed")


def grammar_findings(n: int, txt: str, findings):
    """Add the verb-form findings of one line (rules 3.2/3.4, 3.5, 3.6)."""
    for m in MODAL.finditer(txt):
        findings.append((n, "3.2/3.4", f"modal verb '{m.group(0)}'"))
    for m in PERFECT.finditer(txt):
        w = m.group(2).lower()
        if w == "been" or (is_participle(w) and w not in ALLOW_STATE):
            findings.append((n, "3.2/3.4", f"perfect tense '{m.group(0)}'"))
    for m in PASSIVE.finditer(txt):
        w = m.group(2).lower()
        if re.match(r"\s+\d", txt[m.end():]):
            continue  # "is run 2" names a run, and "is set 3" names a set
        if is_participle(w) and w not in ALLOW_STATE:
            findings.append((n, "3.6", f"passive voice '{m.group(0)}'"))
    for m in ING_AFTER.finditer(txt):
        w = m.group(2).lower()
        if w not in ING_ALLOW:
            findings.append((n, "3.5", f"-ing form '{m.group(0)}'"))
    for s in sentences(txt):
        m = ING_START.match(s)
        if m and m.group(1).lower() not in ING_ALLOW:
            findings.append((n, "3.5", f"-ing form starts a sentence '{m.group(1)}'"))


def length_findings(n: int, s: str, limit: int, rule: str, findings):
    words = len(s.split())
    if words > limit and not is_name_list(s):
        findings.append((n, rule, f"{words} words: {s[:70]}..."))


def check(path: str):
    findings = []
    in_code = in_front = False
    para = []
    held = []  # (line number, text) of the paragraph so far, for wrapped sentences
    n = 0

    def flush_held():
        # A sentence may span hard-wrapped lines. Join the held lines and
        # split into sentences once, so a wrap is not a sentence end.
        if not held:
            return
        first = held[0][0]
        joined = " ".join(t for _, t in held)
        for s in sentences(joined):
            length_findings(first, s, MAX_WORDS, "6.3", findings)
            para.append(s)
        held.clear()
    with open(path, encoding="utf-8") as fh:
        for n, raw in enumerate(fh, 1):
            line = raw.rstrip("\n")
            if n == 1 and line.strip() == "---":
                in_front = True
                continue
            if in_front:
                if line.strip() == "---":
                    in_front = False
                continue
            if line.strip().startswith("```"):
                in_code = not in_code
                continue
            if in_code or line.strip().startswith("|") or line.strip().startswith("#"):
                continue
            if not line.strip() or line.strip() == ">":
                flush_held()
                if len(para) > 6:
                    findings.append((n, "6.6", f"paragraph has {len(para)} sentences"))
                para = []
                continue
            if line.startswith(">") or line.startswith("**"):
                # a block quote or a bold entry title starts a new paragraph
                flush_held()
                if len(para) > 6:
                    findings.append((n, "6.6", f"paragraph has {len(para)} sentences"))
                para = []
            txt = strip_md(line)
            if ";" in txt:
                findings.append((n, "8.1", "semicolon"))
            for m in CONTRACTIONS.finditer(txt):
                findings.append((n, "4.2", f"contraction '{m.group(0)}'"))
            grammar_findings(n, txt, findings)
            if re.match(r"^\s*[-*\d]", line):
                # A list item is one unit. It is not part of the paragraph count.
                flush_held()
                step = re.match(r"^\s*\d+\.", line) is not None
                limit, rule = (MAX_STEP_WORDS, "5.1") if step else (MAX_WORDS, "6.3")
                for s in sentences(txt):
                    length_findings(n, s, limit, rule, findings)
                continue
            held.append((n, txt))
    # The last paragraph of a file ends with no blank line after it.
    flush_held()
    if len(para) > 6:
        findings.append((n, "6.6", f"paragraph has {len(para)} sentences"))
    return findings


if __name__ == "__main__":
    args = sys.argv[1:]
    if "--rules" in args:
        for rule, desc in RULES.items():
            print(f"rule {rule}: {desc}")
        args = [a for a in args if a != "--rules"]
        if not args:
            sys.exit(0)
    total = 0
    for p in args:
        for n, rule, msg in check(p):
            print(f"{p}:{n}: rule {rule}: {msg}")
            total += 1
    print(f"{total} finding(s)")
    sys.exit(1 if total else 0)
