---
name: one-pr-one-session
description: Bind a session to one repository, one branch, one pull request, one milestone, and one role, and make the pull request the complete unit with its code, tests, decisions, documents, and hand-off. Load before any work for a pull request - a start, a revision, a review, a review answer, a merge message, or the hand-off. Stops a second pull request in the same session, stops work that the owner did not approve, and stops a pull request that only records an earlier merge.
---

# One pull request, one clean session

The owner decisions are D-10, D-12, D-13, and D-14, and D-4, D-8, and D-15 for the review. `AGENTS.md` holds the other rules of this repo, and this skill does not repeat them (D-7). The skill is a port of the Decktome skill.

## The rule

A session works on one pull request. A session can make many turns and commits for it. More than one clean session can work on the same pull request, for example a review or a correction. The pull request carries all its work: code, tests, decisions, documents, the review answers, and the hand-off.

A pull request holds one cohesive milestone with one acceptance story (D-10). The milestone can hold two, three, or more concerns when the scope needs them (D-12). The owner approves the work of a pull request before the work starts (D-12).

## 1. The start gate

Do these steps before the first edit.

1. Read this conversation, not the checkout. Look for work on another pull request, another repository, or a finished pull request.
2. When you find such work, stop. Answer only `Blocked: start a new clean session for this PR.`
3. A fork, a subagent, a compaction, or a summary of such a session is not clean. Stop for these too.
4. Run `make where`. Record the branch, the pull request number or the intent, and the base commit.
5. When `make where` says that the git hooks are not installed, run `make hooks`.
6. Keep changes in the checkout that are not yours. Make a worktree for your branch.
7. Name your role: author, reviewer, or correction author.
8. Name the milestone, each concern of it, and its one acceptance story.
9. As the author, ask the owner to approve the milestone. Do not edit a file before the approval (D-12).
10. Read `AGENTS.md`, `docs/session-handoff.md`, and the decisions that the change touches.
11. Write the draft body of section 2 before the code.

A merge message for the pull request of this session is the one exception to step 2. Section 5 gives the answer for it.

A reviewer and a correction author work under the approval of the milestone of the author. When a correction changes the milestone, the correction author asks the owner again.

Start a short branch with the name `<type>/pr-<n>-<slug>`, for example "feat/pr-7-rest-timer" (D-14, D-86). The type is a Conventional Commits type. The focused roadmap of the phase gives the PR-<n> id. The title of the pull request is `<type>: <summary> (PR-<n>)`, with the same type and id.

When you can not do one step, do not start the work. The hook `.claude/hooks/session_bind.py` binds the session to its first branch. When it blocks a command, end the session. Only the owner removes a binding.

## 2. The documentation gate

The pull request body holds the sections of `.github/pull_request_template.md`: `## Session`, `## Milestone`, `## Documentation impact`, `## Checks`, and `## Review`. The `## Milestone` section names the concerns, the one acceptance story, and the approval of the owner.

The `## Documentation impact` table holds one row for each category. Each entry starts with one of these statuses:

- `Changed: <reason and path>`
- `Reviewed; no change needed: <specific reason and path>`
- `Not applicable: <specific reason>`

Obey these rules:

- Change `docs/session-handoff.md` in every pull request. Record the finished state, the checks, the review state, and "pending the owner merge".
- Keep the `## Resume here` section of the hand-off current. It holds the state and the next action.
- Write the line `Author provider: Claude Code` or `Author provider: Codex` in the session record of the hand-off. The Codex review reads it (D-15).
- Keep three session records or fewer in the hand-off. Remove the oldest record, because Git keeps its history.
- Change every document whose facts or contracts the pull request changes. Name the path in backticks.
- A reason says why the document stays correct. "No documentation impact" is not a reason.
- Never defer a document to "after the merge" or to another pull request.
- The hand-off describes only the work in this pull request and the state of its base.
- The roadmap names work areas, and never a pull request number (D-9).
- Never write a merge commit, a merge time, or a deploy result that you did not read. Git and GitHub hold those facts.

The title of the pull request and each commit subject are Conventional Commits subjects, such as `docs(roadmap): add the first phase` (D-14). The squash merge uses the title as the commit subject on `main`. No commit, body, branch, or comment holds AI attribution (D-14). The `commit-msg` hook of `make hooks` refuses a `Co-Authored-By` line that names an AI, and a "Generated with" line.

Run `make pr-check` before you start the review. For a draft body, set `PR_BODY_FILE` and `PR_TITLE`. CI runs the same check on each body edit and push.

## 3. The author loop and the completion gate

Do these steps for each round of changes (D-8):

1. Commit the round, and push it one time.
2. Wait until each CI check of the tip is green. Correct a red check first.
3. Answer each review thread, and resolve it.
4. Run the review of the other provider in the background, and wait for the notice of its end. A Claude Code author runs `make codex-review PR=<number>`. A Codex author runs `make claude-review PR=<number>` (D-15, D-88).
5. Run `git pull --ff-only`, and read the outcome line.
6. For `changes`, answer each finding with `references/answer-review.md` of the `pr-review` skill.
7. After the answer, go to step 1.
8. For `three-strike stop`, stop the loop, and ask the owner.
9. For `approve`, go to the completion gate below.

The author session starts each round itself, with no approval of the owner for each round (D-8, D-88). The target spends the Codex plan or the Claude plan of the owner, and never the API. Each target reads the `Author provider` lines of the branch, and it refuses a review by the provider of the author.

The Codex review applies to every pull request, and to a pull request of documents alone too (D-4). The owner states when the roadmap period of D-4 ends. After that period, D-15 needs a review of the other provider only for a change of code or of safety behavior. `OVERRIDE_ENABLED` in `docs/tools/review_gate.py` holds that state, and it stays `False` until the owner ends the period.

The pull request is complete only when all of these are true:

- The code and its regression tests are in the pull request.
- The decisions and questions are in `docs/decisions.md` and `docs/questions.md`.
- The roadmap and the hand-off read the state of this pull request.
- `make pr-check` passes, and every category has its row.
- `make verify` passes.
- The `review-gate` check passes. A Codex record approves the effective head (D-8).
- The acceptance story of the milestone holds, with evidence.
- No work waits for a second pull request.

### The merge

The owner confirms every merge (D-13). Turn on the auto-merge only when each of these conditions is true:

- The last metadata commit is on origin. It holds the record and the hand-off.
- Each review thread is resolved.
- The record says `Ready for owner merge` for the effective head.
- The owner confirmed the merge after the summary below.

Write the summary in four sections, with a few sentences in each section:

- **What:** the milestone, and the problem that it solves.
- **How:** the method of the change, the evidence, and each risk that stays open.
- **CI:** green or not. Name each check that is not green.
- **Codex review:** the verdict of the record, `Ready for owner merge`, `Blocked`, or `Changes required`. Name the commit that last changed the record, from rule RG 6 of `review-gate` or from `git log -1 -- docs/reviews/pr-<n>.md`. Name the author provider of the hand-off and the provider of the reviewer. The owner reads both before the merge (D-87, D-89).

Write the summary in the question text of `AskUserQuestion`, and ask the owner for the confirmation of the merge in the same text. The owner can see the question alone, so a summary outside it does not reach the owner. Without the confirmation, do not turn on the auto-merge.

After the confirmation, run these commands, in this order:

```bash
gh pr merge <number> --auto --squash
gh pr checks <number> --watch
gh pr view <number> --json state,mergedAt,mergeCommit
```

The ruleset of `main` merges the pull request when each required check passes. When the state is `MERGED`, write the transitional prompt of section 5 at once. When the merge does not come, name the check that blocks it, and ask the owner.

After the prompt, write this line with the number:

`This session is bound to PR #N and is complete. End this session. Start a new clean session before beginning another PR.`

Do not offer the next pull request.

## 4. While the pull request waits

The session stays bound to the pull request while it waits for CI, for the Codex review, or for the owner. It answers each finding on the same pull request (D-12).

- Tell the owner that the session is ready for a context compaction while the pull request waits.
- Read the resume section of `docs/session-handoff.md` again after a context compaction.
- A context compaction of this session keeps its binding. It starts no new pull request.

### The context checkpoint

The session continues until its pull request merges (D-85). A large context never ends the session, and no new session takes over an open pull request.

A session can not see the size of its context. So the hook `.claude/hooks/context_checkpoint.py` tells it at 300K tokens, and again at each further 100K. The hook runs in Claude Code alone. A Codex session does the same step past 300K.

Do these steps after the message of the hook:

1. Finish the current step. A paid run, a push, and a background job end first.
2. Update the resume section of `docs/session-handoff.md` with the finished work, the open work, and the next action.
3. Tell the owner that the session is ready for a context compaction.
4. Continue the work of the same pull request. After a compaction, read the resume section again.

Never write a prompt that hands an open pull request to a new session. The only prompt at the end of a session is the transitional prompt of section 5, after the merge.

## 5. The transitional prompt

### The trigger

Two events start this section. The session reads the state `MERGED` after the auto-merge of section 3. Or the owner says that the pull request merged. The owner asks for no prompt, and the session waits for no other word. Each of these messages is a trigger, and any other variant that names the merge of this pull request:

- `Merged`
- `Merged PR #N`
- `PR #N is merged`
- `#N merged`
- `merged it`

The trigger is the one exception to step 2 of section 1. A merge message for another pull request is not an exception, and it gets the blocked answer of step 2.

Two cases stop the prompt. Ask the owner, and write no prompt until the answer arrives:

- The message names no pull request, and this session holds no binding. Ask which pull request it names.
- The owner merged the pull request before section 3 called it ready. Name each part that did not land, then ask the owner for the next step.

### The procedure

1. Read the merge commit: `git fetch origin && git log --oneline -1 origin/main`.
2. Confirm that the commit names this pull request.
3. Read the next step of `docs/session-handoff.md`.
4. Name the next work area of `docs/roadmaps/high-level-roadmap.md`, and its PR-<n> id in the focused roadmap.
5. Read `docs/questions.md`, and name each open question of that work area.
6. Name each check that needs `main` or the deploy of this merge.
7. Write the block below in the last message, and stop.

The pick of step 4 is provisional, and the owner can name a different work area. The next session asks the owner to approve its milestone before it starts (D-12).

The prompt is one fenced block, and the owner pastes it into the next clean session:

```
Start <work area>: <the proposed milestone>

PR #<x> merged to `main` as <sha>. Read `AGENTS.md` and `docs/session-handoff.md` first.
Branch: `<type>/pr-<n>-<slug>`. Base: `<sha>`. Role: author.
Title: `<type>: <summary> (PR-<n>)` (D-86).
Load the `one-pr-one-session` skill and the skills of the task before any change.
Proposed concerns: <each concern>. Proposed acceptance story: <one story>.
Ask the owner to approve the milestone before the first edit (D-12).
<Each check that needs `main` or the deploy of this merge. Run it before the work.>
Open questions for this work area: <each Q- id with its subject, or `none`>.
First action: <the first concrete action>.
```

### The rules of the prompt

- The prompt carries one milestone. A second milestone needs a second session and a second prompt.
- The prompt never asks the next session to record this merge. Git holds the merge.
- A check that needs `main` comes first. The branch of the next work area gives no such result.
- Remove that line of the block when this merge needs no such check.
- The next step of the hand-off holds the same first action. The two agree, or the hand-off wins.

The session ends with this prompt. It makes no branch and no change for the next pull request.

## Enforcement

| Rule | Enforced by |
|---|---|
| The title, the session block, the milestone, the table, the hand-off change, and a deferred document | `make pr-check` and the `pr-contract` workflow (D-12) |
| The `Author provider:` line of the hand-off | `make pr-check` and the `pr-contract` workflow (D-15) |
| AI attribution in a pull request | `make pr-check` and the `pr-contract` workflow (D-14) |
| A commit on `main`, a commit on a merged branch, and a staged document that fails STE | The `pre-commit` hook of `make hooks` (D-14, D-83) |
| A commit subject that is not Conventional Commits, and AI attribution in a commit | The `commit-msg` hook of `make hooks` (D-14) |
| A merge with no approved record of the other provider | The `review-gate` workflow and the ruleset of `main` (D-4, D-8) |
| A merge with a red check, or an open review thread | The ruleset of `main`, and `make ruleset-check` for its content (D-14) |
| A Codex review on a dirty tree, or with an open review thread | `make codex-review` (D-8) |
| The third open round of one finding | `make codex-review`, exit 4 |
| An API key in a Codex process | `make codex-review` (D-8) |
| A second branch in one session | `.claude/hooks/session_bind.py` (D-12) |
| The context checkpoint at 300K tokens | `.claude/hooks/context_checkpoint.py`, and `make lifecycle-check` reads its wiring (D-6) |
| The skill frontmatter and the wiring | `make lifecycle-check` (D-6) |
| The byte budget of the start read, and the paid targets | `make context-budget` (D-6) |
| Each cited id and each repository path | `make ref-check` (D-6) |
| One pull request in each session, and a clean session for each one | The agent. No check reads the conversation |
| The approval of the owner before the work | The agent records it in the `## Milestone` section (D-12) |
| The truth of each reason, and the cohesion of the milestone | The agent, then the owner (D-10) |
| The merge | The confirmation of the owner, then the auto-merge under the ruleset (D-13) |
| The deploy | Cloud Build, from `main` alone (D-14) |

## Rules of this repo that win over other skills

- This session answers each review finding of its pull request, and an answer never needs a new session.
- Never call the pull request ready before a current Codex record approves its effective head (D-4, D-8).
- Never turn on the auto-merge before the owner confirms the merge (D-13).
- A merge or a deploy of an earlier pull request never gets its own pull request. The next work area reads the base when its own milestone needs it.
