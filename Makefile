# Workout App - the single human entry point.
# Each target prints what it does. `make help` lists each target. Each
# target is free, except each target whose help text says CAUTION.

# Every recipe runs under bash. GNU Make 3.82 added .SHELLFLAGS, and GNU
# Make 3.81 (the make of macOS) ignores it. So each recipe line that pipes
# sets pipefail itself as well.
SHELL := bash
.SHELLFLAGS := -o pipefail -c

.PHONY: help doctor lint ste-check ref-check lifecycle-check context-budget test verify probe proto contract go-test emulator-test web where hooks pr-check gitar-wait codex-review claude-review ruleset-check

help: ## Show this help
	@set -o pipefail; grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

doctor: ## Check the local tools: python3, git, and gh are required, and codex is optional, free
	@status=0; \
	for tool in python3 git gh; do \
	  if command -v $$tool >/dev/null 2>&1; then printf '  ok       %-8s %s\n' "$$tool" "$$($$tool --version 2>&1 | head -1)"; \
	  else printf '  MISSING  %s\n' "$$tool"; status=1; fi; \
	done; \
	if command -v codex >/dev/null 2>&1; then printf '  ok       %-8s %s\n' codex "$$(codex --version 2>&1 | head -1)"; \
	else printf '  warn     codex    absent. make codex-review installs the newest npm release.\n'; fi; \
	exit $$status

lint: ste-check ref-check lifecycle-check context-budget ## Check the documents, the references, the skills and the wiring, and the start-read budget, free

# Every .md file follows ASD-STE100 (D-83). The script reads the prose and
# skips tables and code blocks. Tracked and untracked files are both
# checked, so a new document gets the check before its first commit. Test
# fixtures under testdata are data.
STE_FILES := $(shell (git ls-files '*.md'; git ls-files --others --exclude-standard '*.md') | sort -u | grep -vE '(^|/)testdata/|(^|/)node_modules/')

ste-check: ## Check every hand-written .md file against the STE rules, free (D-83)
	@echo "==> ste-check"
	@python3 docs/tools/ste-check.py $(STE_FILES)

# Every D- and Q- id and every repository path of a document resolves (D-6).
ref-check: ## Check that every cited id and every repository path resolves, free (D-6)
	@echo "==> ref-check"
	@python3 docs/tools/ref_check.py $(STE_FILES)

# The skill frontmatter, and the wiring of the skill, the hooks, the rule
# files, and the template (D-6, D-7).
lifecycle-check: ## Check every skill and the one-pr-one-session wiring, free (D-6)
	@echo "==> lifecycle-check"
	@python3 docs/tools/pr_check.py skills

# The files of the start read stay under a byte budget, and the paid
# targets of this file and of AGENTS.md agree (D-6).
context-budget: ## Check the byte budget of AGENTS.md, CLAUDE.md, the hand-off, and the skills, free (D-6)
	@echo "==> context-budget"
	@python3 docs/tools/context_budget.py

# Each spike of tools/spikes has its own folder and its own tests. The
# tests use a fake provider only, so they make no paid call.
test: ## Run the unit tests of docs/tools, of the hooks, and of each spike of tools/spikes, free
	@echo "==> test"
	@python3 -m unittest discover -s docs/tools -p 'test_*.py'
	@for dir in tools/spikes/*/; do \
	  [ -d "$$dir" ] || continue; \
	  echo "==> test $$dir"; \
	  python3 -m unittest discover -s "$$dir" -p 'test_*.py' || exit 1; \
	done

verify: lint test ## Run every check that CI runs, on this machine, free
	@echo "verify: every check passed."

# The iPhone probe of work area 1.3 (D-113). It needs Node 22 and the
# Playwright browsers, so `make verify` does not run it. The CI job
# verify:probe runs it. The browser tests use the local Auth emulator and
# make no call to the real project (D-115).
PROBE := tools/spikes/iphone_probe

probe: ## Build the iPhone probe and run its browser tests in WebKit and Chromium, free, needs Node 22 (D-113)
	@echo "==> probe"
	@node --version 2>/dev/null | grep -q '^v22\.' || { echo "probe: needs Node 22, see $(PROBE)/.nvmrc"; exit 2; }
	@cd $(PROBE) && npm ci --no-audit --no-fund
	@cd $(PROBE) && npx playwright install chromium webkit
	@cd $(PROBE) && npm run build
	@cd $(PROBE) && npm run test:e2e

# The product checks of work areas 2.1 and 2.2 (D-126). Each target has
# its own CI job, and each job is a required check of main (D-127).
# `make verify` stays Python only (D-113), so it runs none of them. Go
# comes from go/go.mod, and buf and the Go plugins come from its tool
# directives (D-130). The TypeScript plugin comes from
# web/package-lock.json.
GO := go -C go
BUF := .bin/buf
PROTOC_GEN_ES := web/node_modules/.bin/protoc-gen-es

# The ref of the breaking check. A CI checkout of a pull request has no
# local main branch, only origin/main.
PROTO_BASE = $(shell git rev-parse --verify --quiet origin/main >/dev/null && echo origin/main || echo main)

$(BUF): go/go.mod go/go.sum
	@mkdir -p .bin && $(GO) build -o ../$(BUF) github.com/bufbuild/buf/cmd/buf

$(PROTOC_GEN_ES): web/package-lock.json
	@cd web && npm ci --no-audit --no-fund --loglevel=error
	@touch $@

proto: $(BUF) $(PROTOC_GEN_ES) ## Lint the contract in proto/ and write the generated code of go/gen and web/src/gen, free, needs Node 22
	@echo "==> buf lint"
	@$(BUF) lint
	@echo "==> buf generate"
	@$(BUF) generate

# The breaking check reads the contract of the base from a copy of its
# tree, so it works in a worktree and in a shallow CI checkout. A base
# with no proto/ folder has no contract to break.
contract: proto ## Check the contract: buf lint, the generated code in Git, buf breaking against main, and the package move probe, free
	@echo "==> generated code in Git"
	@git diff --exit-code --stat -- go/gen web/src/gen && test -z "$$(git ls-files --others --exclude-standard -- go/gen web/src/gen)" \
	  || { echo "contract: the generated code is stale. Run make proto, then commit go/gen and web/src/gen."; exit 1; }
	@echo "==> buf breaking against $(PROTO_BASE)"
	@if git cat-file -e "$(PROTO_BASE):proto" 2>/dev/null; then \
	  base=$$(mktemp -d) && trap 'rm -rf "$$base"' EXIT && \
	  git archive "$(PROTO_BASE)" proto buf.yaml | tar -x -C "$$base" && \
	  $(BUF) breaking --against "$$base"; \
	else echo "contract: $(PROTO_BASE) has no proto/ folder, so no change can break it"; fi
	@echo "==> the package rules refuse a package move (D-138)"
	@scripts/package_move_probe.sh $(BUF)

go-test: ## Check the Go code: gofmt, go mod tidy, go vet, and the unit tests, free
	@echo "==> gofmt"
	@out=$$(gofmt -l go); [ -z "$$out" ] || { echo "gofmt: format these files:"; echo "$$out"; exit 1; }
	@echo "==> go mod tidy"
	@$(GO) mod tidy -diff
	@echo "==> go vet"
	@$(GO) vet ./...
	@$(GO) vet -tags emulator ./...
	@echo "==> go test"
	@$(GO) test -count=1 ./...

# The emulator tests start the Auth and Firestore emulators of
# firebase.json with the pinned firebase-tools of emulators/. The ports
# do not collide with Decktome or the probe. The project id starts with
# demo-, so no call reaches a real project (D-115). The Firestore
# emulator needs Java 21, and scripts/java21.sh finds it.
EMULATOR_PROJECT := demo-workout-app
EMULATORS := env -u FIREBASE_AUTH_EMULATOR_HOST -u FIRESTORE_EMULATOR_HOST GOOGLE_CLOUD_PROJECT=$(EMULATOR_PROJECT) \
  emulators/node_modules/.bin/firebase emulators:exec --config firebase.json --only auth,firestore --project $(EMULATOR_PROJECT)

emulator-test: ## Run the Go emulator tests over the local Auth and Firestore emulators, free, needs Node 20 or later and Java 21
	@echo "==> emulator-test"
	@cd emulators && npm ci --no-audit --no-fund --loglevel=error
	@java21=$$(scripts/java21.sh) || exit 2; export PATH="$$java21:$$PATH"; unset JAVA_HOME; \
	  $(EMULATORS) '$(GO) test -tags emulator -count=1 ./...'

# The web client of work area 2.2 (D-84). The browser tests run in WebKit
# and in Chromium with phone emulation. They start the API of go/ and a
# build of the web client that talks to the local emulators, so no call
# reaches a real project (D-115).
web: ## Check the web client: types, unit tests, the build, and the browser tests over the emulators and the API, free, needs Node 22, Go, and Java 21
	@echo "==> web"
	@node --version 2>/dev/null | grep -q '^v22\.' || { echo "web: needs Node 22, see web/.nvmrc"; exit 2; }
	@cd web && npm ci --no-audit --no-fund --loglevel=error
	@cd emulators && npm ci --no-audit --no-fund --loglevel=error
	@cd web && npx playwright install chromium webkit
	@echo "==> web typecheck"
	@cd web && npm run typecheck
	@echo "==> web unit tests"
	@cd web && npm test
	@echo "==> web build"
	@cd web && npm run build
	@echo "==> web browser tests"
	@mkdir -p .bin && $(GO) build -o ../.bin/api ./cmd/api
	@java21=$$(scripts/java21.sh) || exit 2; export PATH="$$java21:$$PATH"; unset JAVA_HOME; \
	  $(EMULATORS) 'cd web && npx playwright test'

where: ## Print the branch, the tree, main, and the pull request state, free (D-14)
	@./scripts/where.sh

hooks: ## Install the git hooks: no commit on main, STE on staged documents, and the commit message rules, free (D-14)
	@git config core.hooksPath .githooks
	@echo "hooks installed from .githooks"

# The one-pr-one-session contract (D-10, D-12, D-14). pr-check reads the
# pull request of this branch through gh. For a draft body before the pull
# request exists, pass PR_BODY_FILE and PR_TITLE. The pr-contract workflow
# runs the same check on each push and each body edit.
pr-check: ## Check the pull request of this branch against the one-pr-one-session contract, free (D-12)
	@git fetch --quiet origin main 2>/dev/null || true
	@if [ -n "$(PR_BODY_FILE)" ]; then python3 docs/tools/pr_check.py pr --body-file "$(PR_BODY_FILE)" --title "$(PR_TITLE)"; \
	else python3 docs/tools/pr_check.py pr --gh; fi

# The wait for the Gitar review after a push (D-338). Run it in the
# background at once after each push. It posts at most one `Gitar review`
# comment, and it exits 1 at its limit of 15 minutes.
gitar-wait: ## Wait for a current Gitar review of the head of one pull request: make gitar-wait PR=<n>, free, needs the network (D-338)
	@[ -n "$(PR)" ] || { echo "gitar-wait: set PR to the number of the pull request. Usage: make gitar-wait PR=12"; exit 2; }
	@python3 docs/tools/gitar_wait.py --pr "$(PR)"

# The Codex review of one pull request (D-8). docs/tools/codex_review.py
# holds the refusals, the run, and the read. Its own exit code names the
# outcome, and make turns each code that is not 0 into 2, so read the last
# line: `outcome: <name> (exit <n>)`. Each review target refuses to start
# before the Gitar pass is complete (D-336). SKIP_GITAR=1 reads no Gitar
# pass, and only the owner states the pause that permits it (D-337).
# Only the exact value 1 skips, so SKIP_GITAR=0 or false reads the pass.
GITAR_FLAG = $(if $(filter 1,$(SKIP_GITAR)),--skip-gitar-review)

codex-review: ## Start the Codex review of one pull request and read its record: make codex-review PR=<n>. CAUTION: it spends the Codex plan of the owner, never the API (D-8)
	@[ -n "$(PR)" ] || { echo "codex-review: set PR to the number of the pull request. Usage: make codex-review PR=12"; exit 2; }
	@python3 docs/tools/codex_review.py --pr "$(PR)" $(GITAR_FLAG)

claude-review: ## Start the Claude Code review of one pull request that Codex writes: make claude-review PR=<n>. CAUTION: it spends the Claude plan of the owner, never the API (D-88)
	@[ -n "$(PR)" ] || { echo "claude-review: set PR to the number of the pull request. Usage: make claude-review PR=12"; exit 2; }
	@python3 docs/tools/codex_review.py --pr "$(PR)" --reviewer claude $(GITAR_FLAG)

# The live ruleset does not exist until the owner applies it after the
# merge, so CI does not run this target.
ruleset-check: ## Compare the live ruleset and merge settings of main with .github/rulesets, free, needs the network (D-13, D-14)
	@python3 docs/tools/ruleset_check.py
