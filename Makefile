.PHONY: stray-binaries src-hash build build-cli build-mcp run-cli run-mcp lint fmt test test-focused clean help generate docs parity-check smoke sync-github sync-github-dry sync-check

# Build output directory
BUILD_DIR := build/bin

# Default environment
ENV ?= dev

# Version info embedded at build time
VERSION ?= 0.1.0
GIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
# ⚠️ GIT_HASH is HEAD *at build time*. The normal loop is write -> build -> test -> commit,
# so a build made before the commit is stamped with the PREVIOUS sha while containing the new
# code. That is how 0.1.0-eb914c8 shipped with c42e1dd's --from-start in it (#341).
# -dirty says "this sha is a lower bound, do not trust it". Tracked files only, so an
# untracked scratch dir does not pin the tree to a permanent -dirty.
GIT_DIRTY := $(shell git diff --quiet HEAD 2>/dev/null || echo "-dirty")
# ⚑ SRC_HASH is the only part of the stamp that cannot be wrong (#343). GIT_HASH
# answers "what was HEAD when someone ran make", which is nobody's question; this
# answers "what code is in this binary" — same sources, same hash, committed or
# not. @Sten: `-dirty` means the sha no longer identifies reproducible code, and
# a CLEAN sha can be stale too (commit after building, and the binary keeps the
# previous sha until someone rebuilds). ~/.local/bin/pfw is a shared symlink, so
# that lands on the whole fleet, not just the bench who built it.
# Inputs only — sources plus the module graph. Recompute with `make src-hash`.
SRC_FILES := $(shell find cmd internal -name '*.go' -type f 2>/dev/null | sort)
SRC_HASH := $(shell cat $(SRC_FILES) go.mod go.sum 2>/dev/null | shasum | cut -c1-8)
LDFLAGS := -ldflags "-X main.Version=$(VERSION)-$(GIT_HASH)$(GIT_DIRTY)+src.$(SRC_HASH)"

# Code generation
PROSEFORGE_REPO ?= ../proseforge
SWAGGER_SPEC := $(PROSEFORGE_REPO)/src/backend/docs/swagger/swagger.json
OPENAPI_SPEC := build/openapi3.json
GEN_DIR := internal/api/gen
OAPI_CODEGEN := $(shell go env GOPATH)/bin/oapi-codegen
GOFMT ?= $(shell command -v gofmt 2>/dev/null)
GOIMPORTS ?= $(shell command -v goimports 2>/dev/null || printf '%s/bin/goimports' "$$(go env GOPATH)")

## help: Show this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'

## src-hash: Print the source content hash — does a binary match this tree?
# The verification half of #343. `pfw --version` ends in +src.XXXXXXXX; if that
# does not match this, the binary was built from different code — whatever sha it
# claims. Turns "was this the current code?" into a comparison instead of a guess.
src-hash:
	@printf 'src.%s\n' "$(SRC_HASH)"
	@printf '  build/bin/pfw: %s\n' "$$($(BUILD_DIR)/pfw --version 2>/dev/null || echo 'not built')"

## stray-binaries: Warn (never fail) about compiled binaries outside build/
##
## #342 caught a 12MB `mcp` at the repo root — 1,329 absolute /Users/... paths,
## invisible to the sync guardrail because that scan is `grep -I`, which skips
## binaries. The sync now aborts on it, but ⚠️ a sync is RARE: that catch can sit
## days behind the mistake.
##
## @Sten had the signal 90 minutes before I found it — `git status` printed
## `?? mcp` in his own output and he read past it. His tell is the useful part:
## *"I was running git status to answer 'is my pull clean', so I read only the
## lines relevant to that question."* Evidence on screen loses to the question
## you came with, so the fix is not "read more carefully" — it is to say the
## thing at a moment when it IS the question. Every build is that moment.
##
## Fires almost never, so it does not become noise people scroll past (the test
## the mtime gate failed in #282). Warns only — build and push stay untouched.
stray-binaries:
	@found=$$(find . -type f -perm -u+x \
		-not -path './.git/*' -not -path './build/*' -not -path './dist/*' \
		-exec file {} + 2>/dev/null \
		| grep -E 'Mach-O|ELF|PE32' | cut -d: -f1); \
	if [ -n "$$found" ]; then \
		echo ""; \
		echo "  ┌────────────────────────────────────────────────────────────────┐"; \
		echo "  │  Compiled binary OUTSIDE build/ — it would sync to PUBLIC.     │"; \
		echo "  │                                                                │"; \
		echo "  │  Go bakes absolute /Users/... paths into every build, and the   │"; \
		echo "  │  sync guardrail cannot see inside binaries (grep -I skips       │"; \
		echo "  │  them). make sync-github aborts, but only when you next sync.   │"; \
		echo "  └────────────────────────────────────────────────────────────────┘"; \
		echo "$$found" | sed 's/^/      /'; \
		echo ""; \
		echo "      Delete it, or build with 'make build' (writes to build/)."; \
		echo ""; \
	fi

## build: Build both CLI and MCP binaries
build: build-cli build-mcp client-freshness stray-binaries

## build-cli: Build the CLI binary
build-cli:
	go build $(LDFLAGS) -o $(BUILD_DIR)/pfw ./cmd/cli

## build-mcp: Build the MCP server binary
build-mcp:
	go build $(LDFLAGS) -o $(BUILD_DIR)/workbench-mcp ./cmd/mcp

## publish: Install the tested binary onto PATH — THIS is what reaches the fleet
##
## 🛑 Two-stage build (@Clayton, #383). `make build` compiles PRIVATELY into
## build/bin. `make publish` is the only thing that changes what other benches
## run. Before this split they were the same act: ~/.local/bin/pfw was a symlink
## into build/bin, so every `make build` — including a mid-development one —
## replaced the binary ~28 live watchers re-exec onto. @Tate and @Sten each
## shipped untested code to the fleet on 2026-08-22, four hours apart, both
## having read the warning against it. The operation that published was spelled
## identically to the operation that compiled.
PUBLISH_DIR ?= $(HOME)/.local/bin
PUBLISH_PATH := $(PUBLISH_DIR)/pfw

publish: build
	@test -d "$(PUBLISH_DIR)" || (echo "publish dir $(PUBLISH_DIR) does not exist" >&2; exit 2)
	@echo "  was: $$($(PUBLISH_PATH) --version 2>/dev/null | head -1 || echo 'nothing installed')"
	@echo "  new: $$($(BUILD_DIR)/pfw --version | head -1)"
	@# ⚑ ATOMIC. Write beside the target, then rename — rename(2) on one filesystem
	@# is atomic, so a watcher re-execing mid-publish sees either the old image or
	@# the new one and never a half-written file. `cp` straight onto the target
	@# would expose exactly that window, and the readers are 28 processes polling
	@# on their own schedule.
	@# ⚠️ ONE shell. Each recipe line is its own process, so $$$$ (the PID) differs
	@# between lines — splitting this made chmod look for a file cp created under a
	@# different name. It failed safely (the symlink was untouched), but a
	@# three-line "atomic" install is not atomic.
	@set -e; \
	  tmp="$(PUBLISH_PATH).tmp.$$$$"; \
	  cp "$(BUILD_DIR)/pfw" "$$tmp"; \
	  chmod 755 "$$tmp"; \
	  mv -f "$$tmp" "$(PUBLISH_PATH)"
	@echo ""
	@echo "  PUBLISHED to $(PUBLISH_PATH)"
	@echo "  Every 'room watch --loop' re-execs onto this at its next tick."
	@echo "  60s legs: within a minute. 30m legs: up to 30 minutes."

## run-cli: Build and run the CLI (pass ARGS for command arguments)
run-cli: build-cli
	$(BUILD_DIR)/pfw $(ARGS)

## run-mcp: Build and run the MCP server
run-mcp: build-mcp
	$(BUILD_DIR)/workbench-mcp

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format code
fmt:
	@test -x "$(GOFMT)" || (echo "gofmt not found; install Go or set GOFMT=/path/to/gofmt" >&2; exit 2)
	@test -x "$(GOIMPORTS)" || (echo "goimports not found; run 'go install golang.org/x/tools/cmd/goimports@latest' or set GOIMPORTS=/path/to/goimports" >&2; exit 2)
	$(GOFMT) -w .
	$(GOIMPORTS) -w .

## test: Run all tests
test:
	go test ./...

## test-focused: Run selected tests (PKGS=./path RUN=TestName)
test-focused:
	@test -n "$(PKGS)" || (echo "PKGS is required, e.g. PKGS=./internal/api" >&2; exit 2)
	@test -n "$(RUN)" || (echo "RUN is required, e.g. RUN=TestName" >&2; exit 2)
	go test $(PKGS) -run "$(RUN)"

## test-integration: Run integration tests (requires running API server)
test-integration:
	go test -tags integration -v -run TestIntegration ./...

## whisper-models: Fetch the Whisper STT models the narration harness needs (#443)
whisper-models:
	@mkdir -p build/whisper-models
	@for m in tiny.en base.en; do \
		f=build/whisper-models/ggml-$$m.bin; \
		if [ -s "$$f" ]; then echo "  have  $$f"; else \
			echo "  fetch $$f"; \
			curl -sL --fail -o "$$f" \
				"https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-$$m.bin" \
				|| { echo "  FAILED to fetch $$m"; rm -f "$$f"; exit 1; }; \
		fi; \
	done
	@echo "Whisper models ready in build/whisper-models/"

## narration-validate: Run narration rebuild test matrix against dev (#218)
narration-validate: build-cli whisper-models
	@./scripts/test/narration/reset-fixture.sh --narrate
	@for s in scripts/test/narration/scenarios/*.sh; do \
		echo "=== $$s ==="; \
		"$$s" || exit 1; \
	done

## narration-fixture-reset: Reset (or bootstrap) the narration test fixture story
narration-fixture-reset: build-cli
	@./scripts/test/narration/reset-fixture.sh

## test-verbose: Run all tests with verbose output
test-verbose:
	go test -v ./...

## smoke: Run smoke tests against live API (read-only, requires .env.$(ENV))
smoke: build-cli
	@if [ ! -f .env.$(ENV) ]; then echo "Missing .env.$(ENV) — copy .env.example and fill in values"; exit 1; fi
	@set -a && . ./.env.$(ENV) && set +a && \
	TOKEN=$$PROSEFORGE_TOKEN && \
	URL=$$PROSEFORGE_URL && \
	if [ -z "$$URL" ] || [ -z "$$TOKEN" ]; then echo "PROSEFORGE_URL and token required in .env.$(ENV)"; exit 1; fi && \
	PASS=0 && FAIL=0 && \
	run() { \
		DESC="$$1"; shift; \
		if PROSEFORGE_URL=$$URL PROSEFORGE_TOKEN=$$TOKEN $(BUILD_DIR)/pfw "$$@" >/dev/null 2>&1; then \
			echo "  ✓ $$DESC"; PASS=$$((PASS+1)); \
		else \
			echo "  ✗ $$DESC"; FAIL=$$((FAIL+1)); \
		fi; \
	} && \
	echo "Smoke tests (env: $(ENV), url: $$URL)" && \
	run "story list" story list && \
	run "story get (published)" story get 260570eb-5f13-46bc-9580-441760a2443a && \
	run "story section get" story section get 260570eb-5f13-46bc-9580-441760a2443a 48c0bfa8-6020-4a8b-ab94-c0620c425b4d && \
	run "story quality" story quality 260570eb-5f13-46bc-9580-441760a2443a && \
	run "story export" story export 260570eb-5f13-46bc-9580-441760a2443a --format markdown && \
	run "feedback list" feedback list 260570eb-5f13-46bc-9580-441760a2443a && \
	echo "" && \
	echo "Results: $$PASS passed, $$FAIL failed" && \
	if [ $$FAIL -gt 0 ]; then exit 1; fi

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)/*

## tidy: Run go mod tidy
tidy:
	go mod tidy

## docs: Generate CLI reference documentation
docs: build-cli
	$(BUILD_DIR)/pfw gendocs --dir docs/cli

## parity-check: Check CLI <-> MCP command-surface drift (internal tooling)
parity-check: build-cli
## 🛑 The `||` this replaces could not tell "the file is missing" from "the check
## ran and FAILED" — both printed "internal tooling not present" and exited 0. So
## a real red parity result reported itself as an absence, and the ~27-item
## backlog grew behind a green light (#345).
##
## A check whose failure is indistinguishable from its non-execution is worse
## than no check: it produces a green people reasonably trust. Absence stays
## benign (a public-mirror contributor has no scripts/ at all); failure must
## propagate.
	@if [ ! -f scripts/parity-check.py ]; then \
		echo "parity-check: internal tooling not present, skipped"; \
	else \
		python3 scripts/parity-check.py --check; \
	fi

## generate: Generate Go types and client from ProseForge Swagger spec
generate:
	@echo "Converting Swagger 2.0 → OpenAPI 3.0..."
	@mkdir -p build $(GEN_DIR)
	npx swagger2openapi $(SWAGGER_SPEC) -o $(OPENAPI_SPEC)
	@echo "Generating Go types and client..."
	$(OAPI_CODEGEN) --config oapi-codegen.yaml $(OPENAPI_SPEC)
	@echo "Running go mod tidy..."
	go mod tidy
	@shasum -a 256 $(SWAGGER_SPEC) | cut -d' ' -f1 > $(GEN_DIR)/.spec-sha256
	@echo "Done. Generated code in $(GEN_DIR)/"

## client-freshness: Warn (never fail) if the spec's CONTENT differs from what we generated from
##
## Compares a recorded sha256 of the spec, not file times. mtime cannot work here:
## the spec is no longer tracked in the proseforge repo and is regenerated at
## build time (proseforge a5c23cf7), so **every deploy makes it newer than our
## client** whether or not a single route changed. An mtime gate therefore fires
## constantly and correctly almost never — and a warning that is usually wrong is
## one people learn to scroll past, which is the failure #282 set out to fix.
##
## A hash costs one shasum and has no false positives: the spec carries no build
## timestamp (info is description/title/contact/version only), so identical APIs
## hash identically across deploys.
##
## Still only warns and still points at generate-stale for the authoritative
## answer — that part of #282's design was right. Nothing here blocks a build or
## a push. Missing hash file or missing spec: silent, since neither proves
## anything (a public-repo contributor has no sibling checkout at all).
client-freshness:
	@if [ -f $(SWAGGER_SPEC) ] && [ -f $(GEN_DIR)/.spec-sha256 ] && \
	   [ "$$(shasum -a 256 $(SWAGGER_SPEC) | cut -d' ' -f1)" != "$$(cat $(GEN_DIR)/.spec-sha256)" ]; then \
		echo ""; \
		echo "  ┌────────────────────────────────────────────────────────────────┐"; \
		echo "  │  The API spec CHANGED since the client was generated.          │"; \
		echo "  │                                                                │"; \
		echo "  │  A route renamed or removed upstream is invisible here until   │"; \
		echo "  │  a tool call 404s in front of a user. Regenerating turns that   │"; \
		echo "  │  into a compile error instead.                                 │"; \
		echo "  │                                                                │"; \
		echo "  │      make generate-stale    is it actually stale?              │"; \
		echo "  │      make generate          update it                          │"; \
		echo "  └────────────────────────────────────────────────────────────────┘"; \
		echo ""; \
	fi

## generate-stale: Fail if the committed client disagrees with the current spec
##
## A renamed or removed route is invisible here until an agent calls the tool and
## gets a 404 in front of a user. Regenerating turns that class into a compile
## error instead — the cheapest cross-repo guard we have, and the only cost is
## remembering to run it (#278).
##
## Cannot run in CI: the spec lives in the sibling ../proseforge checkout, which
## the public mirror's workflows never have. So this is a local gate — run it
## when an upstream slice lands, before trusting the client.
## It is a PURE CHECK — it restores the committed file before returning.
##
## An earlier version left the regenerated file in place, which made it
## self-healing and therefore useless as a gate: `git push origin` pushes to two
## URLs, git runs pre-push once per URL, and the first run "fixed" the tree so
## the second passed. The result was Gitea refused and Bitbucket accepted — the
## exact split-brain the dual remote exists to avoid. A check that changes what
## it is checking cannot be run twice and mean the same thing.
generate-stale:
	@test -f $(SWAGGER_SPEC) || { \
		echo "generate-stale: no spec at $(SWAGGER_SPEC) — needs the ../proseforge checkout"; exit 1; }
	@cp $(GEN_DIR)/proseforge.gen.go build/.gen-before.go
## 🛑 RESTORE THE HASH TOO, NOT JUST THE CLIENT. `generate` writes
## .spec-sha256, and that file is the ONLY input to the client-freshness
## warning. Restoring proseforge.gen.go alone left the hash matching the CURRENT
## spec while the client stayed OLD -- so running this diagnostic SILENCED the
## warning it exists to raise, and printed "Your tree is untouched" while having
## touched it. Measured 2026-09-03: hashes differed before, matched after, and
## the client was still missing GetAdminLatency both times.
	@cp $(GEN_DIR)/.spec-sha256 build/.spec-sha256-before 2>/dev/null || true
	@$(MAKE) --no-print-directory generate >/dev/null
	@cp $(GEN_DIR)/proseforge.gen.go build/.gen-after.go
	@cp build/.gen-before.go $(GEN_DIR)/proseforge.gen.go
	@[ -f build/.spec-sha256-before ] && cp build/.spec-sha256-before $(GEN_DIR)/.spec-sha256 || true
	@rm -f build/.spec-sha256-before
	@if diff -q build/.gen-before.go build/.gen-after.go >/dev/null; then \
		rm -f build/.gen-before.go build/.gen-after.go; \
		echo "generate-stale: client matches the spec"; \
	else \
		echo "generate-stale: THE COMMITTED CLIENT IS STALE."; \
		echo ""; \
		echo "Client methods that changed:"; \
		grep -o 'func (c \*Client) [A-Za-z]*' build/.gen-before.go | sort -u > build/.gen-m-before; \
		grep -o 'func (c \*Client) [A-Za-z]*' build/.gen-after.go | sort -u > build/.gen-m-after; \
		diff build/.gen-m-before build/.gen-m-after \
			| grep '^[<>]' | sed 's/^< /  REMOVED /; s/^> /  ADDED   /' || true; \
		echo "  (no method names above = a route MOVED, or fields/types changed)"; \
		echo ""; \
		echo "Routes that changed:"; \
		grep -o 'operationPath := fmt\.Sprintf("[^"]*"' build/.gen-before.go \
			| sed 's/.*Sprintf("//; s/"$$//' | sort -u > build/.gen-r-before; \
		grep -o 'operationPath := fmt\.Sprintf("[^"]*"' build/.gen-after.go \
			| sed 's/.*Sprintf("//; s/"$$//' | sort -u > build/.gen-r-after; \
		diff build/.gen-r-before build/.gen-r-after \
			| grep '^[<>]' | sed 's/^< /  GONE  /; s/^> /  NEW   /' || true; \
		echo "  (a path rename keeps its operationId, so it appears HERE and not above)"; \
		echo ""; \
		echo "Your tree is untouched. Run 'make generate', then 'go build ./...' to see"; \
		echo "what breaks, fix the call sites, and commit. A build error here is a 404"; \
		echo "you did not ship."; \
		rm -f build/.gen-before.go build/.gen-after.go build/.gen-m-before build/.gen-m-after \
			build/.gen-r-before build/.gen-r-after; \
		exit 1; \
	fi

## sync-github: Sync public content to GitHub (regenerate docs, stage, verify, push)
sync-github: docs
	scripts/sync-github.sh

## sync-github-dry: Dry run — regenerate docs, stage and verify, but don't push
sync-github-dry: docs
	scripts/sync-github.sh --dry-run

## sync-check: Run guardrail checks only (no staging)
sync-check:
	scripts/sync-github.sh --check
