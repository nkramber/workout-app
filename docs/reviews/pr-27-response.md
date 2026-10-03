# Pull request 27 - author response

Date: 2026-10-03. Review round 1 recorded effective head `07acb3c`, with the verdict "Changes required".

## P2-1: The Go README says that the API does not call Luna

**Result: full merit.**

The trigger reproduces at `07acb3c`. Line 37 of `go/README.md` said that the API does not call Luna and reads neither cap variable. That sentence came from a pull request before this one. This pull request makes `go/cmd/api` read both caps and the key at its start. The plan service calls Luna through the cap hook of `go/internal/capstore`.

**Correction.**

- `go/README.md`: the sentence now says that the API reads both variables at its start. The cap hook applies them to each planner call, and a plan request over a cap gives an error at once (D-25, D-230).

**Regression checks.**

- A search of `go/README.md`, `docs/design.md`, `AGENTS.md`, `docs/setup-gcp.md`, and `docs/deploy-and-rollback.md` for "does not call Luna", "reads no OpenAI", and "not use the role layer" finds nothing.
- `make ste-check`, `make ref-check`, and `make verify` pass.
