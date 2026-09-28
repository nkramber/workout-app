#!/usr/bin/env python3
"""Tell a session when its context passes the checkpoint limit (D-6, D-12).

A PostToolUse hook on every tool. Claude Code gives the hook a JSON
object on stdin with `session_id`, `cwd`, and `transcript_path`. The hook
reads the usage of the last API call of the main thread from the
transcript. The context of that call is its input tokens, its cache read
tokens, and its cache write tokens.

A session can not see the size of its own context. So this hook tells
the session at 300,000 tokens,
and again at each further 100,000. The session then follows section 4 of
the `one-pr-one-session` skill: it updates the hand-off, tells the owner
that it is ready for a context compaction, and continues the same pull
request until the merge (D-85).

The hook writes the last level it told to the git common dir of the
repository, so each level speaks one time:
<common-dir>/claude-context-checkpoint/<session_id>.json. The hook names
no repository, so it works in any checkout. It is a port of the Decktome hook.

The hook fails open. When it can not read the event, the transcript, or
the directory, it says nothing. A subagent gets no message, because the
main thread holds the context that grows.
"""
import json
import os
import re
import subprocess
import sys

LIMIT = 300_000
STEP = 100_000
TAIL = 512 * 1024
STORE = "claude-context-checkpoint"
SESSION_ID = re.compile(r"^[A-Za-z0-9._-]{1,128}$")


def context_of(usage):
    return sum(int(usage.get(k) or 0) for k in ("input_tokens", "cache_read_input_tokens", "cache_creation_input_tokens"))


def last_context(lines):
    """Return the context of the last main-thread API call in the lines, or None."""
    for line in reversed(lines):
        try:
            record = json.loads(line)
        except ValueError:
            continue
        if record.get("type") != "assistant" or record.get("isSidechain"):
            continue
        usage = (record.get("message") or {}).get("usage")
        if usage:
            return context_of(usage)
    return None


def read_tail(path, size=TAIL):
    with open(path, "rb") as handle:
        handle.seek(0, os.SEEK_END)
        end = handle.tell()
        handle.seek(max(0, end - size))
        data = handle.read().decode("utf-8", errors="replace")
    lines = data.splitlines()
    return lines[1:] if end > size else lines


def level(context):
    """Return the checkpoint level that the context reached, or 0 below the limit."""
    if context < LIMIT:
        return 0
    return LIMIT + (context - LIMIT) // STEP * STEP


def message(context):
    return (
        f"Context checkpoint (D-12): the last call of this session held {context:,} tokens of context, "
        f"past the limit of {LIMIT:,}. Do section 4 of the one-pr-one-session skill at the next safe point. "
        "Update the resume section of docs/session-handoff.md, and tell the owner that the session "
        "is ready for a context compaction. Do not end the session. "
        "Continue the same pull request until it merges (D-85)."
    )


def decide(store, session_id, context):
    """Return the message for this call, or an empty string. Record each level that speaks."""
    reached = level(context)
    if not reached:
        return ""
    path = os.path.join(store, session_id + ".json")
    told = 0
    try:
        with open(path, encoding="utf-8") as handle:
            told = int(json.load(handle).get("level") or 0)
    except (OSError, ValueError, AttributeError):
        told = 0
    if reached <= told:
        return ""
    os.makedirs(store, exist_ok=True)
    with open(path, "w", encoding="utf-8") as handle:
        json.dump({"level": reached, "context": context}, handle)
    return message(context)


def common_dir(cwd):
    try:
        out = subprocess.run(["git", "rev-parse", "--path-format=absolute", "--git-common-dir"],
                             cwd=cwd, capture_output=True, text=True, timeout=5)
    except (OSError, subprocess.SubprocessError):
        return None
    return out.stdout.strip() if out.returncode == 0 else None


def main():
    try:
        event = json.load(sys.stdin)
        session_id = str(event.get("session_id") or "")
        path = str(event.get("transcript_path") or "")
        cwd = event.get("cwd") or os.getcwd()
        if event.get("agent_id") or not SESSION_ID.match(session_id) or not os.path.isfile(path):
            return 0
        context = last_context(read_tail(path))
        if context is None or context < LIMIT:
            return 0
        common = common_dir(cwd)
        if not common:
            return 0
        text = decide(os.path.join(common, STORE), session_id, context)
    except Exception:  # noqa: BLE001 - the hook fails open, see the module docstring
        return 0
    if text:
        print(json.dumps({"hookSpecificOutput": {"hookEventName": "PostToolUse", "additionalContext": text}}))
    return 0


if __name__ == "__main__":
    sys.exit(main())
