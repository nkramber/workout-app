#!/usr/bin/env python3
"""Check the one-pr-one-session contract (D-10, D-12, D-14).

Two modes:

  pr_check.py pr --event FILE [--base REF]
  pr_check.py pr --gh [--base REF]
  pr_check.py pr --body-file FILE --title TEXT [--head BRANCH] [--base REF]
      Check a pull request title, body, and diff against the contract:
      the Conventional Commits title, the session block, the milestone,
      the documentation-impact matrix, the hand-off change, deferred
      documentation, AI attribution, and a merge-record pull request.

  pr_check.py skills
      Check the frontmatter of every skill, and the wiring of the
      one-pr-one-session skill, the hooks, the rule files, and the template.

The script reads what CI can know. It can not read the conversation of a
session, so the skill and the hook carry the session rules. The tool is a
port of the Decktome check. It drops the roadmap merge mark, because the
roadmap holds no permanent pull request numbers (D-9).
"""
import argparse
import json
import os
import re
import subprocess
import sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
SKILL = ".claude/skills/one-pr-one-session/SKILL.md"
TEMPLATE = ".github/pull_request_template.md"
HANDOFF = "docs/session-handoff.md"
AGENTS = "AGENTS.md"
CLAUDE = "CLAUDE.md"
GIT_HOOKS = (".githooks/pre-commit", ".githooks/commit-msg")
EXEMPT_AUTHORS = {"dependabot[bot]", "app/dependabot"}
BLOCKED = "Blocked: start a new clean session for this PR."
COMPLETE = "This session is bound to PR #N and is complete. End this session. Start a new clean session before beginning another PR."

# The canonical document categories, in match order. A changed file joins
# the first category with a matching path. A path that ends in "/" is a folder.
CATEGORIES = [
    ("session hand-off", [HANDOFF]),
    ("agent rules", [AGENTS, CLAUDE]),
    ("readme", ["README.md"]),
    ("design", ["docs/design.md"]),
    ("decisions", ["docs/decisions.md"]),
    ("questions", ["docs/questions.md"]),
    ("research", ["docs/research/"]),
    ("roadmaps", ["docs/roadmaps/"]),
    ("skills and hooks", [".claude/"]),
    ("tools and ci", ["docs/tools/", ".github/", ".githooks/", "scripts/", "Makefile", ".gitignore"]),
]
PREFIXES = {prefix for _, prefixes in CATEGORIES for prefix in prefixes}
CHANGED = "Changed:"
REVIEWED = "Reviewed; no change needed:"
NOT_APPLICABLE = "Not applicable:"
STATUSES = (CHANGED, REVIEWED, NOT_APPLICABLE)
ROLES = {"author", "reviewer", "correction author"}
MIN_REASON_WORDS = 6
MIN_STORY_WORDS = 6
MILESTONE_FIELDS = ("Concerns", "Acceptance story", "Owner approval")

# D-14: the squash merge takes the title as its commit subject.
CONVENTIONAL = re.compile(
    r"^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-z0-9][a-z0-9._/-]*\))?!?: \S"
)
# D-86: every pull request after #1 names its roadmap id, as in
# the-thing-below and what-you-carry. Title: "<type>: <summary> (PR-<n>)".
# Branch: "<type>/pr-<n>-<slug>". The two carry the same type and id.
# The id comes from a focused roadmap, and it is not the GitHub number.
TITLE_ID = re.compile(r"^([a-z]+)(\([a-z0-9][a-z0-9._/-]*\))?!?: .+ \(PR-([1-9][0-9]*)\)$")
BRANCH_ID = re.compile(r"^([a-z]+)/pr-([1-9][0-9]*)-[a-z0-9]+(-[a-z0-9]+)*$")
# PR #1 started before D-86, so its branch keeps its old name.
FIRST_BRANCH = "docs/foundation-roadmap"
# D-14: no AI attribution. The commit-msg hook holds the same rule.
AI_NAME = r"(claude|anthropic|codex|openai|chatgpt|gpt)"
ATTRIBUTION = re.compile(
    rf"^\s*co-authored-by:.*\b{AI_NAME}|^\W*generated (with|by)\b|\bgenerated (with|by)\b.*\b{AI_NAME}",
    re.I | re.M,
)
PROVIDER_LINE = re.compile(r"^\s*[-*]?\s*Author provider:\s*(Claude Code|Codex)\b", re.M)

GENERIC = re.compile(
    r"^(none|n/?a|unchanged|not needed|no impact|no changes?( needed)?|nothing to (update|change)|"
    r"no (doc|docs|documentation|document|documents) (impact|change|changes)( needed)?)\.?$"
)
PLACEHOLDER = re.compile(r"<[a-z][a-z ,/-]*>|\b(TODO|TBD|FIXME|XXX)\b|^Pending\b", re.I)
DOC_NOUN = re.compile(r"\b(doc|docs|document|documents|documentation|hand-off|handoff|roadmap|decision|decisions|readme|claude\.md|agents\.md|record|records)\b", re.I)
DEFER = re.compile(
    r"\b(after (the )?merge|once (it|this|the pull request|the pr) (merges|is merged)|follow-up|follow up|"
    r"(later|separate|next|another|second|new) (pr|pull request)|will (update|follow|come)|to follow|"
    r"at a later|in a later|post-merge)\b",
    re.I,
)
NEGATION = re.compile(r"\b(no|not|never|none|nor|without|forbid|forbids|reject|rejects|refuse|refuses|cannot|stop|stops)\b", re.I)
MERGE_RECORD_TITLE = re.compile(r"\b(read|reads|record|records)\b[^.]*\b(merge|merged|merges|deploy|deployed)\b|documents read the", re.I)


def category_of(path):
    for name, prefixes in CATEGORIES:
        for prefix in prefixes:
            if path == prefix or (prefix.endswith("/") and path.startswith(prefix)):
                return name
    return None


def is_document(path):
    return path.endswith(".md")


def strip_code(text):
    text = re.sub(r"```.*?```", " ", text, flags=re.S)
    return re.sub(r"`[^`\n]*`", " ", text)


def section(body, title):
    match = re.search(rf"^##\s+{re.escape(title)}\s*$(.*?)(?=^##\s|\Z)", body, flags=re.M | re.S | re.I)
    return match.group(1) if match else None


def table_rows(text):
    rows = []
    for line in text.splitlines():
        line = line.strip()
        if not line.startswith("|"):
            continue
        cells = [c.strip() for c in line.strip("|").split("|")]
        if all(re.fullmatch(r":?-{3,}:?", c) for c in cells if c):
            continue
        rows.append(cells)
    return rows[1:] if rows else rows


def clean_name(cell):
    return re.sub(r"[`*_]", "", cell).strip().lower()


def named_paths(entry):
    return [p for p in re.findall(r"`([^`\s]+)`", entry) if "/" in p or "." in p or p in PREFIXES]


def reason_words(entry, status):
    reason = entry[len(status):]
    reason = re.sub(r"`[^`]*`", " ", reason)
    return reason.strip(), len(re.findall(r"[A-Za-z0-9][A-Za-z0-9'-]*", reason))


def check_title(title):
    if title and not CONVENTIONAL.match(title):
        return [f"the title '{title}' is not a Conventional Commits subject, such as 'docs: add the roadmap' (D-14)"]
    return []


def check_session(body, head_ref, is_ancestor):
    errors = []
    block = section(body, "Session")
    if block is None:
        return ["the body holds no '## Session' section (role, branch, and base)"]
    role = re.search(r"^\s*[-*]?\s*Role:\s*(.+?)\s*$", block, flags=re.M)
    branch = re.search(r"^\s*[-*]?\s*Branch:\s*`?([^`\s]+)`?\s*$", block, flags=re.M)
    base = re.search(r"^\s*[-*]?\s*Base:\s*`?([0-9a-f]{7,40})`?\s*$", block, flags=re.M)
    if not role or role.group(1).strip().lower() not in ROLES:
        errors.append("the session block names no role: author, reviewer, or correction author")
    if not branch:
        errors.append("the session block names no branch")
    elif head_ref and branch.group(1) != head_ref:
        errors.append(f"the session block names branch {branch.group(1)}, and the pull request head is {head_ref}")
    if not base:
        errors.append("the session block records no base commit")
    elif is_ancestor is not None and not is_ancestor(base.group(1)):
        errors.append(f"the base commit {base.group(1)} is not an ancestor of the head")
    return errors


def check_milestone(body):
    """One milestone, its concerns, one acceptance story, and the approval of the owner (D-10, D-12)."""
    block = section(body, "Milestone")
    if block is None:
        return ["the body holds no '## Milestone' section (concerns, acceptance story, and owner approval)"]
    errors = []
    for field in MILESTONE_FIELDS:
        found = re.findall(rf"^\s*[-*]?\s*{re.escape(field)}:\s*(.*?)\s*$", block, flags=re.M)
        if not found or not found[0]:
            errors.append(f"the milestone names no '{field}:' line")
            continue
        if len(found) > 1:
            errors.append(f"the milestone holds {len(found)} '{field}:' lines. A pull request holds one milestone (D-10)")
        value = found[0]
        if PLACEHOLDER.search(value):
            errors.append(f"the milestone '{field}:' line holds a placeholder")
        elif field == "Acceptance story" and len(value.split()) < MIN_STORY_WORDS:
            errors.append(f"the acceptance story holds {len(value.split())} words, and a story needs {MIN_STORY_WORDS} or more")
    return errors


def check_matrix(body, changed, exists):
    errors = []
    block = section(body, "Documentation impact")
    if block is None:
        return ["the body holds no '## Documentation impact' section"]
    seen = {}
    for cells in table_rows(block):
        if len(cells) < 2:
            errors.append(f"the matrix row '{' | '.join(cells)}' holds no entry")
            continue
        name, entry = clean_name(cells[0]), cells[1]
        if name in seen:
            errors.append(f"the matrix names the category '{name}' two times")
        seen[name] = entry
    by_category = {}
    for path in changed:
        cat = category_of(path)
        if cat:
            by_category.setdefault(cat, []).append(path)
    for name, _ in CATEGORIES:
        if name not in seen:
            errors.append(f"the matrix holds no row for the category '{name}'")
    for name, entry in seen.items():
        errors.extend(check_entry(name, entry, by_category.get(name, []), exists))
    return errors


def check_entry(name, entry, touched, exists):
    status = next((s for s in STATUSES if entry.startswith(s)), None)
    if status is None:
        return [f"'{name}': the entry starts with none of: {', '.join(STATUSES)}"]
    errors = []
    if PLACEHOLDER.search(entry):
        errors.append(f"'{name}': the entry holds a placeholder")
    reason, words = reason_words(entry, status)
    if GENERIC.match(reason.lower()) or "no documentation impact" in reason.lower():
        errors.append(f"'{name}': the reason is generic. Say why this document stays correct")
    elif words < MIN_REASON_WORDS:
        errors.append(f"'{name}': the reason holds {words} words, and a specific reason needs {MIN_REASON_WORDS} or more")
    known = next((prefixes for cat, prefixes in CATEGORIES if cat == name), None)
    paths = named_paths(entry)
    if name == "session hand-off" and status != CHANGED:
        errors.append(f"'session hand-off': every pull request changes {HANDOFF}")
    if name == "session hand-off" and HANDOFF not in touched:
        errors.append(f"'session hand-off': the diff does not change {HANDOFF}")
    if status == CHANGED:
        if not touched and known is not None:
            errors.append(f"'{name}': the entry says Changed, and the diff changes no file of this category")
        elif paths and touched and not any(p.rstrip("/") in touched or any(t.startswith(p) for t in touched) for p in paths):
            errors.append(f"'{name}': no path the entry names is in the diff")
        elif not paths:
            errors.append(f"'{name}': the entry names no path in backticks")
    else:
        if touched:
            errors.append(f"'{name}': the entry says '{status}', and the diff changes {', '.join(sorted(touched))}")
        if status == REVIEWED:
            if not paths:
                errors.append(f"'{name}': the entry names no reviewed path in backticks")
            elif not all(exists(p) for p in paths):
                errors.append(f"'{name}': a reviewed path does not exist: {', '.join(p for p in paths if not exists(p))}")
    return errors


def deferred_sentences(text):
    found = []
    for sentence in re.split(r"(?<=[.!?])\s+|\n+", strip_code(text)):
        if DEFER.search(sentence) and DOC_NOUN.search(sentence) and not NEGATION.search(sentence):
            found.append(sentence.strip())
    return found


def attribution_lines(text):
    return [m.group(0).strip() for m in ATTRIBUTION.finditer(text or "")]


def check_naming(title, head_ref):
    """Check the roadmap id of the title and of the branch (D-86)."""
    if head_ref == FIRST_BRANCH:
        return []
    errors = []
    branch = BRANCH_ID.match(head_ref or "")
    if not branch:
        errors.append(f"the branch '{head_ref}' is not '<type>/pr-<n>-<slug>', such as 'feat/pr-7-rest-timer' (D-86)")
    if not title:
        return errors
    named = TITLE_ID.match(title)
    if not named:
        errors.append(f"the title '{title}' does not end with the roadmap id, such as 'feat: the rest timer (PR-7)' (D-86)")
    elif branch and (named.group(1), named.group(3)) != (branch.group(1), branch.group(2)):
        errors.append(f"the title '{title}' and the branch '{head_ref}' name a different type or PR id (D-86)")
    return errors


def check_pr(title, body, author, head_ref, changed, exists, is_ancestor=None, handoff_added="", handoff_text=None):
    """Return every contract error of one pull request. Empty means it passes.

    handoff_text is the hand-off at the head, or None when the caller did
    not read it.
    """
    if author in EXEMPT_AUTHORS:
        return []
    body = body or ""
    errors = []
    errors.extend(check_title(title))
    errors.extend(check_naming(title, head_ref))
    errors.extend(check_session(body, head_ref, is_ancestor))
    errors.extend(check_milestone(body))
    errors.extend(check_matrix(body, changed, exists))
    for sentence in deferred_sentences(body):
        errors.append(f"the body defers documentation: \"{sentence}\"")
    for sentence in deferred_sentences(handoff_added):
        errors.append(f"the hand-off defers documentation: \"{sentence}\"")
    for line in attribution_lines(f"{title}\n{body}"):
        errors.append(f"the pull request holds AI attribution: \"{line}\" (D-14)")
    if handoff_text is not None and not PROVIDER_LINE.search(handoff_text):
        errors.append(f"{HANDOFF} holds no 'Author provider: Claude Code' or 'Author provider: Codex' line. The review reads it (D-15)")
    if changed and all(is_document(p) for p in changed) and MERGE_RECORD_TITLE.search(title or ""):
        errors.append("a pull request of documents alone records an earlier merge or deploy. Git and GitHub hold those facts")
    return errors


def git(*args):
    out = subprocess.run(["git", *args], cwd=ROOT, capture_output=True, text=True)
    if out.returncode != 0:
        raise SystemExit(f"pr_check: git {' '.join(args)} failed: {out.stderr.strip()}")
    return out.stdout


def run_pr(args):
    if args.event:
        with open(args.event, encoding="utf-8") as handle:
            pr = json.load(handle).get("pull_request") or {}
        title, body = pr.get("title", ""), pr.get("body") or ""
        author, head = (pr.get("user") or {}).get("login", ""), (pr.get("head") or {}).get("ref", "")
    elif args.gh:
        out = subprocess.run(["gh", "pr", "view", "--json", "title,body,author,headRefName"], cwd=ROOT, capture_output=True, text=True)
        if out.returncode != 0:
            raise SystemExit("pr_check: gh reads no pull request for this branch. Pass PR_BODY_FILE and PR_TITLE for a draft.")
        pr = json.loads(out.stdout)
        title, body = pr["title"], pr["body"]
        author, head = pr["author"]["login"], pr["headRefName"]
    else:
        with open(args.body_file, encoding="utf-8") as handle:
            body = handle.read()
        title, author, head = args.title or "", "", args.head or git("rev-parse", "--abbrev-ref", "HEAD").strip()
    if author in EXEMPT_AUTHORS:
        print(f"pr_check: {author} is exempt. The owner reads a dependency pull request.")
        return 0
    changed = [p for p in git("diff", "--name-only", f"{args.base}...HEAD").splitlines() if p]
    handoff_added = "\n".join(
        line[1:] for line in git("diff", "--unified=0", f"{args.base}...HEAD", "--", HANDOFF).splitlines()
        if line.startswith("+") and not line.startswith("+++")
    )
    try:
        with open(os.path.join(ROOT, HANDOFF), encoding="utf-8") as handle:
            handoff_text = handle.read()
    except OSError:
        handoff_text = ""

    def exists(path):
        return os.path.exists(os.path.join(ROOT, path.rstrip("/")))

    def is_ancestor(sha):
        return subprocess.run(["git", "merge-base", "--is-ancestor", sha, "HEAD"], cwd=ROOT, capture_output=True).returncode == 0

    errors = check_pr(title, body, author, head, changed, exists, is_ancestor, handoff_added, handoff_text)
    for error in errors:
        print(f"pr_check: {error}")
    print(f"pr_check: {len(changed)} changed files, {len(errors)} contract errors")
    return 1 if errors else 0


def frontmatter(text):
    match = re.match(r"^---\n(.*?)\n---\n", text, flags=re.S)
    if not match:
        return None
    fields = {}
    for line in match.group(1).splitlines():
        if ":" in line and not line.startswith((" ", "\t")):
            key, value = line.split(":", 1)
            fields[key.strip()] = value.strip()
    return fields


def check_skills(read, listdir):
    """Check the skills and the wiring. read gives a text or None, and listdir gives the folders of a path."""
    errors = []
    allowed = {"name", "description", "license", "allowed-tools", "metadata", "compatibility"}
    for name in sorted(listdir(".claude/skills")):
        path = f".claude/skills/{name}/SKILL.md"
        text = read(path)
        if text is None:
            errors.append(f"{path} does not exist")
            continue
        fields = frontmatter(text)
        if fields is None:
            errors.append(f"{path} holds no frontmatter")
            continue
        if fields.get("name") != name:
            errors.append(f"{path}: name '{fields.get('name')}' is not the directory name '{name}'")
        if not re.fullmatch(r"[a-z0-9]+(-[a-z0-9]+)*", name) or len(name) > 64:
            errors.append(f"{path}: the name is not kebab case of 64 characters or fewer")
        description = fields.get("description", "")
        if not description or len(description) > 1024 or "<" in description or ">" in description:
            errors.append(f"{path}: the description is empty, over 1024 characters, or holds angle brackets")
        for key in fields:
            if key not in allowed:
                errors.append(f"{path}: unknown frontmatter key '{key}'")
        if re.search(r"gitar", text, flags=re.I):
            errors.append(f"{path} names Gitar. Gitar stays out until the owner approves it (D-3)")
    skill = read(SKILL) or ""
    for phrase in (BLOCKED, COMPLETE):
        if phrase not in skill:
            errors.append(f"{SKILL} does not hold the exact text: {phrase}")
    # D-7: AGENTS.md holds the rules, and CLAUDE.md points to it alone.
    if SKILL not in (read(AGENTS) or ""):
        errors.append(f"{AGENTS} does not require {SKILL}")
    if AGENTS not in (read(CLAUDE) or ""):
        errors.append(f"{CLAUDE} does not point to {AGENTS} (D-7)")
    template = read(TEMPLATE) or ""
    for heading in ("## Session", "## Milestone", "## Documentation impact"):
        if heading not in template:
            errors.append(f"{TEMPLATE} holds no '{heading}' section")
    rows = {clean_name(c[0]) for c in table_rows(section(template, "Documentation impact") or "")}
    for name, _ in CATEGORIES:
        if name not in rows:
            errors.append(f"{TEMPLATE} holds no matrix row for '{name}'")
    try:
        settings = json.loads(read(".claude/settings.json") or "{}")
    except ValueError:
        settings = {}
        errors.append(".claude/settings.json is not valid JSON")
    for event, hook in (("PreToolUse", "session_bind.py"), ("PostToolUse", "context_checkpoint.py")):
        if hook not in json.dumps(settings.get("hooks", {}).get(event, [])):
            errors.append(f".claude/settings.json does not run the {hook} hook on {event}")
    for hook in GIT_HOOKS:
        if read(hook) is None:
            errors.append(f"{hook} does not exist")
    return errors


def run_skills():
    def read(path):
        try:
            with open(os.path.join(ROOT, path), encoding="utf-8") as handle:
                return handle.read()
        except OSError:
            return None

    def listdir(path):
        full = os.path.join(ROOT, path)
        if not os.path.isdir(full):
            return []
        return [d for d in os.listdir(full) if os.path.isdir(os.path.join(full, d))]

    errors = check_skills(read, listdir)
    for error in errors:
        print(f"skill-check: {error}")
    print(f"skill-check: {len(errors)} errors")
    return 1 if errors else 0


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = parser.add_subparsers(dest="mode", required=True)
    pr = sub.add_parser("pr")
    source = pr.add_mutually_exclusive_group(required=True)
    source.add_argument("--event")
    source.add_argument("--gh", action="store_true")
    source.add_argument("--body-file")
    pr.add_argument("--title")
    pr.add_argument("--head")
    pr.add_argument("--base", default="origin/main")
    sub.add_parser("skills")
    args = parser.parse_args()
    return run_pr(args) if args.mode == "pr" else run_skills()


if __name__ == "__main__":
    sys.exit(main())
