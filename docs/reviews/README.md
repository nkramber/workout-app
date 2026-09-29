# Review records

This folder holds one review record for each pull request. The record is the file `docs/reviews/pr-<n>.md`, and the number is the GitHub number of the pull request.

A Codex session writes the record with the `pr-review` skill after the author session runs `make codex-review PR=<n>` (D-8). The record names the effective head of the pull request, the findings, and one verdict. The `review-gate` check reads the head field and the verdict, and the ruleset of `main` requires that check.

The owner ended the roadmap period of D-4 on 2026-09-29 (D-125). A pull request of documents alone can now use the `review-override` label in place of a record (D-15). The owner confirms each merge after the review (D-13).

An author answer to a review goes in the file `docs/reviews/pr-<n>-response.md`. The check does not read that file.

Do not edit the record of a merged pull request. It is the history of that review. The skeleton and the rules of the record are in `.claude/skills/pr-review/references/review-record.md`.
