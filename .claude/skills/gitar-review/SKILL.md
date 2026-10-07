---
name: gitar-review
description: Get a Gitar review of the head of a pull request after each push, prove that the review is current, and answer every finding. Verify each finding as a claim, then fix and reply, or refute, reply, and resolve. Load after each push to a pull request, documents alone included (D-335).
---

# Gitar review skill

The GitHub app `gitar-bot` reviews each pull request of this repository (D-335). This skill gets a Gitar review of the head of a pull request, and then answers each finding. A pull request of documents alone waits for the review too.

The skill is a port of the Decktome skill (`decktome:.claude/skills/gitar-review/SKILL.md`), with the wait of what-you-carry. A rule of `AGENTS.md` wins over this skill. Load `references/traps.md` before the first Gitar pass of a pull request, and again when a wait does not end.

## Terms

- **Head**: the newest commit of the pull request branch on GitHub.
- **Dashboard comment**: the Gitar comment on the pull request that holds the collapsed `Code Review` block. Gitar edits this comment for each review. Gitar can also delete it and post a new one with a new id.
- **Pause note**: the note at the top of the dashboard comment that starts "Automatic reviews are paused".
- **Manual review**: the review that a `Gitar review` comment starts.
- **Accept reply**: the Gitar reply that starts a manual review: "On it", "Running the review now", or "Running a review on this PR now".
- **Metadata set**: the paths of `metadata_paths` in `docs/tools/review_gate.py`: the review record, the response file, and `docs/session-handoff.md`.
- **Effective head**: the newest commit that changes a path outside the metadata set.
- **Current review**: a review of the effective head.
- **Stale review**: a review of a commit older than the effective head.
- **Gitar pass**: a current review, with an answer to each finding and no open review thread.

## Why a review goes stale

Gitar reviews each push until it pauses automatic reviews. After the pause, a push starts no review, and the dashboard comment keeps the review of an older commit. The dashboard comment names no commit. Thus a stale review looks the same as a current review.

A manual review also goes stale when a push comes after the `Gitar review` comment. On 2026-09-16, four pull requests in Decktome and what-you-carry had a review older than the head. Each one had a push after the last request, or a push and no request.

## Procedure

Do these steps after each push.

1. Push all the commits of this round of changes. Push one time, not one time for each fix.
2. Run command A. Continue only when the local head and the pull request head are the same commit.
3. Run `make gitar-wait PR=<number>` in the background at once (D-338). Wait for the notice of its end.
4. Do not comment `Gitar review` while the wait runs. The wait posts one request when no review started.
5. When the wait exits 1, read its last lines. Stop, and tell the owner.
6. When the wait exits 0, run command B. Apply the rule in "Prove that a review is current".
7. When the review is stale, comment `Gitar review` on the pull request. Then go to step 3.
8. Do not push while a manual review runs. A push at this time makes the review stale.
9. Open the collapsed `Code Review` block of the dashboard comment. Read the summary.
10. List the review threads with command C. Read each open thread.
11. Read each top-level Gitar comment with command B. A comment that names a specific item is a finding.
12. Read each finding as a claim, not a fact. Reproduce its trigger. Read the rule or the decision it names.
13. Decide the merit of the finding: full, partial, or none.
14. For full merit, make the smallest change that fixes the finding. Commit it.
15. For no merit, reply on the thread with the reason and the evidence. Then resolve the thread.
16. For partial merit, fix the part with merit. Refute the rest in the same reply.
17. When you have commits, go to step 1. After the push, reply on each thread with the commit that fixes it.
18. Resolve each thread after its reply. The ruleset of `main` refuses a merge with an open thread.
19. Stop when a current review approves, or when a current review adds no finding and each finding has its answer.
20. Wait until each CI check of the tip is green. Then go to the review step of the `one-pr-one-session` skill.

The review of the other provider comes after the Gitar pass and after green CI (D-335). The `answer-review.md` file of the `pr-review` skill gives its steps.

## The check of the review targets

`make codex-review` and `make claude-review` refuse to start until the Gitar pass is complete (D-336). They read these facts:

- The Gitar check on the tip completed.
- The newest dashboard comment changed after the push of the effective head. The first check suite of that push gives its time.
- A `Gitar review` comment after that push has an accept reply and a later change of the dashboard. "Prove that a review is current" gives the order.
- The dashboard shows no spinner of a review in progress.
- No review thread of the pull request is open.

A top-level Gitar comment is not a review thread, so no check reads its answer. Read each one with command B before the target runs. Read them again before the auto-merge.

The owner can state a pause of Gitar. Only then, run the target with `SKIP_GITAR=1` (D-337). The target then reads no Gitar pass. It still refuses an open review thread and an issue on the dashboard.

A pull request of documents alone can carry the `review-override` label in place of the review of the other provider (D-15, D-125). No target runs, so no check reads the Gitar pass. Complete the Gitar pass before you apply the label.

## Prove that a review is current

A review is current only when each of these conditions is true:

- The head from command B is the head of command A. A later commit of the metadata set also passes this condition.
- The dashboard comment has an edit time later than the push of the effective head.
- After a `Gitar review` comment, Gitar replied "On it", and the dashboard comment has an edit time later than that reply.
- Gitar can send an accept reply after the review ends. Then the dashboard edit time must be later than the request.
- The dashboard shows no spinner and no "Responding to your feedback" line. These mark a review in progress.
- You read the newest dashboard comment. Gitar can delete the dashboard comment and post a new one with a new id.

The summary is not a condition. A review that adds no finding can keep the summary of the older review, word for word. Do not ask for a review again only because the summary did not change.

When one condition is false, the review is stale. When you can not check one condition, treat the review as stale. A request for a manual review costs little. A merge on a stale review costs more.

A commit of the metadata set does not make a pass stale. The review record and the hand-off are in the metadata set. A rule that reads the branch tip alone makes each pass stale at the moment of its record, and the gate then never passes. Prove the effective head with `git diff --stat <reviewed head>..<tip>`.

After the approval of the other provider, a commit of documents alone can keep `review-gate` green. So the Gitar pass is the one review of that commit. Push no commit after the auto-merge turns on, because GitHub can then merge it before Gitar reads it.

## Rules for each reply

- State the evidence: the command, the test, the decision id, or the commit.
- Write no AI attribution (D-14). A reply names no provider, harness, or model.
- Never accept a finding only to close the review faster. A wrong fix costs more than a written disagreement.
- Never make a fix larger than the rule that the finding names.
- When a finding conflicts with an owner decision, quote both and ask the owner.
- A Gitar comment that names no specific item is not a finding. Do not reply to a plan notice, a pause note, or a dashboard with no issue.
- Only the author replies to Gitar. The reviewer of the other provider never replies to Gitar (D-335).

## Commands

Set the four variables once in each shell. The commands read the repository from the working directory.

```bash
repo=$(gh repo view --json nameWithOwner --jq .nameWithOwner)
owner=${repo%/*}
name=${repo#*/}
n=<number>
```

### A. The head check

```bash
# The two commit ids must be the same.
git rev-parse HEAD
gh pr view "$n" --json headRefOid --jq .headRefOid
```

### B. The freshness check

```bash
# The head now. Compare it with the head of command A.
echo "head:      $(gh pr view "$n" --json headRefOid --jq .headRefOid)"

# The time of the newest "Gitar review" comment, if any.
echo "requested: $(gh api --paginate "repos/$repo/issues/$n/comments" \
  --jq '.[] | select(.body | test("^\\s*gitar review\\s*$"; "i")) | .created_at' | tail -1)"

# The time and the first line of the newest Gitar reply to a request: an accept reply or a refusal.
echo "reply:     $(gh api --paginate "repos/$repo/issues/$n/comments" \
  --jq '.[] | select(.user.login == "gitar-bot[bot]")
        | select(.body | test("^> gitar review"; "i"))
        | "\(.created_at) \(.body | split("\n")[2])"' | tail -1)"

# The id and the last edit time of the newest dashboard comment.
echo "dashboard: $(gh api --paginate "repos/$repo/issues/$n/comments" \
  --jq '.[] | select(.user.login == "gitar-bot[bot]")
        | select(.body | test("<b>Code Review</b>"))
        | "\(.id) \(.updated_at)"' | tail -1)"

# The dashboard text. Replace <id> with the id above.
gh api "repos/$repo/issues/comments/<id>" --jq .body

# Each top-level Gitar comment: the time and the first line.
gh api --paginate "repos/$repo/issues/$n/comments" \
  --jq '.[] | select(.user.login == "gitar-bot[bot]") | "\(.created_at) \(.body | split("\n")[0])"'
```

### C. The review threads

```bash
gh api graphql -F owner="$owner" -F name="$name" -F number="$n" -f query='
  query($owner: String!, $name: String!, $number: Int!) {
    repository(owner: $owner, name: $name) {
      pullRequest(number: $number) {
        reviewThreads(first: 100) {
          nodes {
            id
            isResolved
            path
            line
            comments(first: 20) { nodes { databaseId author { login } body } }
          }
        }
      }
    }
  }'
```

### D. Replies and requests

```bash
# The checks of the pull request, the Gitar check included
gh pr checks "$n"

# Reply on a thread. The comment id is the databaseId of the first comment.
gh api "repos/$repo/pulls/$n/comments/<comment-id>/replies" -f body='<reply>'

# Resolve a thread. The thread id comes from command C.
gh api graphql -f id=<thread-id> -f query='
  mutation($id: ID!) {
    resolveReviewThread(input: {threadId: $id}) { thread { isResolved } }
  }'

# Ask for a manual review
gh pr comment "$n" --body "Gitar review"
```
