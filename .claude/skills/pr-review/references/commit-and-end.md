# PR review: the commit and the session end

Part of the `pr-review` skill. Load this file before you commit a review record or a response file, and before the session ends.

## Commit the record

Commit and push the record in the session that writes it.

| After | Commit these files | Who commits |
|---|---|---|
| A review or a repeat review | `docs/reviews/pr-<number>.md` and `docs/session-handoff.md` | The reviewer |
| An answer to a review | `docs/reviews/pr-<number>-response.md`, each corrected file, and `docs/session-handoff.md` | The author |

The reviewer records the review state in the resume section of the hand-off: the effective head, the verdict, and each open finding id. It changes no other part of the hand-off. Both paths sit in the metadata set, so the commit keeps the effective head.

The check reads the head of the pull request. So the check sees the record only after the push. A review is complete only when the branch on GitHub holds the record.

Write the commit message as a Conventional Commits subject in an impersonal voice, for example `docs(review): add the review record of #12` (D-14). Name no provider, agent, harness, or model, and add no co-author line (D-14).

## The end gate

Run these commands after the commit, in this order:

```bash
git push origin <branch>
git fetch origin
git status --short --branch
gh pr view <number> --json headRefOid --jq .headRefOid
```

The status line must show no `[ahead N]`. The hash from `gh pr view` must be the same as `git rev-parse HEAD`. Write the push line in the `## Verification` section of the record.

`make codex-review` starts the reviewer in a worktree with a detached HEAD. There, push with `git push origin HEAD:<branch>`. The status line then names no branch, so the hash comparison is the proof. The target also refuses a push that changes a path outside the metadata set.

When the remote refuses the push, the review is not complete. Tell the owner that the record has a commit and no push. A sandbox with no network can refuse the push with no message from git, so read the status line.

At the start of a review, run `git fetch` and `git status --short --branch` too.
