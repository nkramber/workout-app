#!/usr/bin/env python3
"""Bind one Claude Code session to one pull request branch (D-6, D-12).

A PreToolUse hook on the Bash tool. Claude Code gives the hook a JSON
object on stdin with `session_id`, `cwd`, and `tool_input.command`. The
hook reads the branch that a command starts work on, and binds the
session to the first such branch. A later command that starts work on
another branch exits 2, and Claude Code blocks it.

These commands start work on a branch:
- git switch -c, git checkout -b, git worktree add -b
- git push (the refspec, or the current branch)
- gh pr create (--head, or the current branch)
- gh pr checkout, comment, review, edit, ready, close, reopen

The binding lives in the git common dir of the repository that the
command reads, so every worktree of it reads the same binding:
<common-dir>/claude-session-bind/<session_id>.json. The hook names no
repository, so it works in any checkout. The owner removes that file to
release a binding. A session never does.

The hook fails open. When it can not read the command, the directory,
or the branch, it allows the call. The pr-contract check and the skill
carry the rest. The hook is a port of the Decktome hook.
"""
import json
import os
import re
import shlex
import subprocess
import sys
import time

BLOCKED = "Blocked: start a new clean session for this PR."
STORE = "claude-session-bind"
SESSION_ID = re.compile(r"^[A-Za-z0-9._-]{1,128}$")
ASSIGN = re.compile(r"^([A-Za-z_][A-Za-z0-9_]*)=(.*)$")
SPLIT = re.compile(r"&&|\|\||;|\||\n")
PR_VERBS = {"checkout", "comment", "review", "edit", "ready", "close", "reopen"}
IGNORED = {"", "HEAD", "main"}


def run(args, cwd, timeout):
    try:
        out = subprocess.run(args, cwd=cwd, capture_output=True, text=True, timeout=timeout)
    except (OSError, subprocess.SubprocessError):
        return None
    return out.stdout.strip() if out.returncode == 0 else None


def current_branch(directory):
    if not directory or not os.path.isdir(directory):
        return None
    return run(["git", "rev-parse", "--abbrev-ref", "HEAD"], directory, 5)


def pr_branch(selector, directory):
    if not directory or not os.path.isdir(directory):
        return None
    return run(["gh", "pr", "view", selector, "--json", "headRefName", "--jq", ".headRefName"], directory, 15)


def expand(text, env):
    if text.startswith("~"):
        text = os.path.expanduser(text)

    def sub(match):
        name = match.group(1) or match.group(2)
        return env.get(name, os.environ.get(name, "\0"))

    text = re.sub(r"\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)", sub, text)
    return None if "\0" in text or "$" in text or "`" in text else text


def option_value(args, names):
    for i, arg in enumerate(args):
        for name in names:
            if arg == name and i + 1 < len(args):
                return args[i + 1]
            if name.startswith("--") and arg.startswith(name + "="):
                return arg[len(name) + 1:]
    return None


def positional(args, with_value=()):
    out, skip = [], False
    for arg in args:
        if skip:
            skip = False
            continue
        if arg.startswith("-"):
            skip = arg in with_value
            continue
        out.append(arg)
    return out


def git_target(args, directory, branch_of):
    """The branch one git command starts work on, or None."""
    while args and args[0].startswith("-"):
        if args[0] == "-C" and len(args) > 1:
            directory = args[1] if os.path.isabs(args[1]) else os.path.join(directory or "", args[1])
            args = args[2:]
        elif args[0] == "-c" and len(args) > 1:
            args = args[2:]
        else:
            args = args[1:]
    if not args:
        return None
    verb, rest = args[0], args[1:]
    if verb == "switch":
        return option_value(rest, ["-c", "-C", "--create", "--force-create"])
    if verb == "checkout":
        return option_value(rest, ["-b", "-B"])
    if verb == "worktree" and rest[:1] == ["add"]:
        return option_value(rest[1:], ["-b", "-B"])
    if verb == "push":
        if any(a in ("-d", "--delete", "--tags") for a in rest):
            return None
        refs = positional(rest, with_value=("-o", "--push-option", "--repo"))
        if len(refs) >= 2:
            ref = refs[1].lstrip("+")
            if ref.startswith(":"):
                return None
            ref = ref.split(":", 1)[1] if ":" in ref else ref
            ref = ref[len("refs/heads/"):] if ref.startswith("refs/heads/") else ref
            return branch_of(directory) if ref == "HEAD" else ref
        return branch_of(directory)
    return None


def gh_target(args, directory, branch_of, branch_of_pr):
    """The branch one gh command starts work on, or None."""
    if len(args) < 2 or args[0] != "pr":
        return None
    verb, rest = args[1], args[2:]
    if verb == "create":
        return option_value(rest, ["--head", "-H"]) or branch_of(directory)
    if verb not in PR_VERBS:
        return None
    selectors = positional(rest, with_value=("-b", "--body", "-F", "--body-file", "-R", "--repo", "-t", "--title", "--add-label", "--remove-label", "--base", "-B"))
    if not selectors:
        return branch_of(directory)
    selector = selectors[0]
    if re.fullmatch(r"#?\d+", selector) or selector.startswith("http"):
        return branch_of_pr(selector.lstrip("#"), directory)
    return selector


def targets(command, cwd, branch_of=current_branch, branch_of_pr=pr_branch):
    """Every branch a Bash command starts work on, in order."""
    if "git" not in command and "gh" not in command:
        return []
    found, env, directory = [], {}, cwd
    for segment in SPLIT.split(command):
        try:
            words = shlex.split(segment, comments=True)
        except ValueError:
            continue
        while words and ASSIGN.match(words[0]):
            name, value = ASSIGN.match(words[0]).groups()
            env[name] = value
            words = words[1:]
        if not words:
            continue
        head, args = words[0], words[1:]
        if head == "cd":
            path = expand(args[0], env) if args else os.path.expanduser("~")
            if path is None:
                directory = None
            elif os.path.isabs(path):
                directory = path
            elif directory:
                directory = os.path.join(directory, path)
            continue
        if head == "git":
            branch = git_target(args, directory, branch_of)
        elif head == "gh":
            branch = gh_target(args, directory, branch_of, branch_of_pr)
        else:
            continue
        if branch and branch not in IGNORED:
            found.append(branch)
    return found


def decide(store, session_id, branches, command, now=None):
    """Return (allowed, message). Write the binding on the first branch."""
    if not branches:
        return True, ""
    path = os.path.join(store, session_id + ".json")
    bound = None
    if os.path.exists(path):
        try:
            with open(path, encoding="utf-8") as handle:
                bound = json.load(handle)
        except (OSError, ValueError):
            bound = None
    if bound is None:
        bound = {
            "branch": branches[0],
            "bound_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime(now)),
            "command": command[:200],
        }
        os.makedirs(store, exist_ok=True)
        with open(path, "w", encoding="utf-8") as handle:
            json.dump(bound, handle)
    for branch in branches:
        if branch != bound.get("branch"):
            return False, (
                f"{BLOCKED}\n"
                f"This session is bound to branch {bound.get('branch')} since {bound.get('bound_at')} (D-12).\n"
                f"The command starts work on branch {branch}. A session works on one pull request.\n"
                "Finish the bound pull request, end this session, and start a new clean session."
            )
    return True, ""


def main():
    try:
        event = json.load(sys.stdin)
        session_id = str(event.get("session_id") or "")
        command = str((event.get("tool_input") or {}).get("command") or "")
        cwd = event.get("cwd") or os.getcwd()
        if not SESSION_ID.match(session_id) or not command:
            return 0
        branches = targets(command, cwd)
        if not branches:
            return 0
        common = run(["git", "rev-parse", "--path-format=absolute", "--git-common-dir"], cwd, 5)
        if not common:
            return 0
        allowed, message = decide(os.path.join(common, STORE), session_id, branches, command)
    except Exception:  # noqa: BLE001 - the hook fails open, see the module docstring
        return 0
    if allowed:
        return 0
    print(message, file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main())
