#!/usr/bin/env bash
# Print where the checkout stands, before a commit, a push, or a merge (D-14).
# `make where` runs this file. It is a port of the Decktome script.
set -uo pipefail

git fetch --quiet origin 2>/dev/null || true

branch=$(git rev-parse --abbrev-ref HEAD)
dirty=$(git status --porcelain)
printf 'branch          %s\n' "$branch"
printf 'tree            %s\n' "$([ -z "$dirty" ] && echo clean || echo "DIRTY, $(echo "$dirty" | wc -l | tr -d ' ') files")"
printf 'head            %s\n' "$(git rev-parse --short HEAD 2>/dev/null || echo none)"

if git rev-parse --abbrev-ref '@{u}' >/dev/null 2>&1; then
  read -r behind ahead < <(git rev-list --left-right --count '@{u}...HEAD' | awk '{print $1, $2}')
  printf 'vs upstream     behind %s, ahead %s\n' "$behind" "$ahead"
else
  printf 'vs upstream     none, this branch is not pushed\n'
fi

if git rev-parse --verify --quiet origin/main >/dev/null; then
  if git rev-parse --verify --quiet main >/dev/null; then
    # Only the behind count is read. The ahead count goes to a name the
    # shell discards.
    read -r _ mbehind < <(git rev-list --left-right --count "main...origin/main" | awk '{print $1, $2}')
    printf 'main            %s\n' "$([ "$mbehind" = 0 ] && echo "current with origin" || echo "BEHIND origin by $mbehind")"
  fi
  printf 'origin/main     %s\n' "$(git rev-parse --short origin/main)"
  if [ "$branch" != "main" ]; then
    printf 'unmerged here   %s commits not in origin/main\n' "$(git rev-list --count origin/main..HEAD)"
  fi
else
  printf 'origin/main     none\n'
fi

if command -v gh >/dev/null 2>&1; then
  pr=$(gh pr list --head "$branch" --state all --json number,state --jq 'if length == 0 then "" else "#\(.[0].number) \(.[0].state)" end' 2>/dev/null || echo "")
  printf 'pull request    %s\n' "${pr:-none}"
  case "$pr" in
    *MERGED*) printf '\nWARNING: this branch is merged. A commit here never reaches main.\n' ;;
  esac
else
  printf 'pull request    unknown, gh is not installed\n'
fi

if [ "$(git config --get core.hooksPath)" != ".githooks" ]; then
  printf '\nWARNING: the git hooks are not installed. Run: make hooks\n'
fi

if [ "$branch" = "main" ]; then
  printf '\nWARNING: main takes no commit (D-14). Start a short branch first.\n'
fi
