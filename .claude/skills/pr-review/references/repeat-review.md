# PR review: the repeat review

Part of the `pr-review` skill. Load this file when you review a pull request again after a correction. The scope rules of `SKILL.md` apply to each new finding.

## The procedure

Do these steps in order after the author changes the pull request.

1. Read the response file, when one exists.
2. Apply the provider gate again. A fix by the reviewer changes the result.
3. Read the new effective head, the new base, and the diff since the reviewed head.
4. Verify each fix against its trigger and its regression check.
5. Set the `Status` line of each earlier finding. Keep each id and each piece of evidence.
6. Add the new effective head to the `Open at:` line of each finding that stays open.
7. Examine the new diff for new defects and for the consumers that it touches.
8. Give each new finding the next index of its severity.
9. Put the new effective head in the `## Identity` list.
10. Put the commands that ran on the new head in the `## Verification` section.
11. Move the old verdict under `## Earlier verdicts`, and write the new verdict.
12. Commit the record, then apply the end gate of `references/commit-and-end.md`.

Edit the existing record. Do not make a second record for the same pull request.

The reason of the new verdict names no other verdict. Write "the earlier findings are fixed", and not "the changes required are done". The check counts each bold span of the section, and the prose stays plain.

## When a finding closes

A finding closes when the correction makes its trigger pass, and its regression check passes. Then set its status to `fixed in <sha>`.

A new trigger of the same class of defect is a new finding with a new id.

Stop at the third assessment of one id. Write the pattern in the record, and ask the owner whether this pull request or a later item holds the whole surface.

The `Open at:` line of `references/review-record.md` holds the count of this rule, and `make codex-review` reads it. A reviewer that `make codex-review` starts can not ask the owner. So it gives the verdict `Changes required`, it names the pattern under `## Open questions and accepted risks`, and the target stops the loop.
