# Traps of a Gitar review

Part of the `gitar-review` skill. This file holds the cases that cost a session a wrong answer, from Decktome and what-you-carry. Read it before the first Gitar pass of a pull request, and again when a wait does not end.

- A green Gitar check does not prove that no finding is open. Read the threads.
- A completed Gitar check on the head does not prove the end of the review. On what-you-carry PR #103, the dashboard came 53 seconds after the check completed.
- A green Gitar check on the head does not prove that the review is current. A paused Gitar attaches a check with the pause note.
- A manual review can edit the dashboard comment and attach no Gitar check to the new head. Apply the rule in "Prove that a review is current".
- The pause note can come beside a full review. Open the collapsed `Code Review` block before you comment `Gitar review`.
- A request before a push gets a review of the old head. Push first, then ask.
- A request during the push wait can start a second review beside the automatic review. It can also use the request limit of Gitar. Let `make gitar-wait` do the request.
- Gitar limits requests. When Gitar replies "You've sent several Gitar comments in a short window", wait ten minutes. Then comment `Gitar review` one time.
- Do not send a second `Gitar review` comment while the first review runs.
- Gitar refuses a request in a reply, and the dashboard comment does not change. A wait that watches the dashboard comment alone then never ends. Read the reply first.
- Gitar can replace the dashboard comment during a review. A saved comment id then returns HTTP 404, or it shows an old edit time. Read the newest id in each check.
- The REST API names the author `gitar-bot[bot]`, and the GraphQL API names it `gitar-bot`.
- The issue comments API returns 30 comments on each page. Use `--paginate`, or you can read an old dashboard comment.
- Gitar does not review a pull request of a bot, such as Dependabot. An admin must add the bot under "Allowed bots" in the Gitar settings (what-you-carry:D-773). Tell the owner, because no author is exempt from the review (D-90).
- The owner can merge a pull request before a finding gets its answer. A commit on that branch then never gets to `main`. Carry the fix to a new branch from `main`. Reply on the old thread with the new pull request.
- Gitar can confirm a fix in a reply and resolve its own thread. Read the thread before you resolve it yourself.
