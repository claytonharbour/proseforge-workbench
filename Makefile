.PHONY: build build-cli build-mcp run-cli run-mcp lint fmt test clean help generate docs smoke sync-github sync-github-dry sync-check

# Build output directory
BUILD_DIR := build/bin

# Default environment
ENV ?= dev

# Version info embedded at build time
VERSION ?= 0.1.0
GIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.Version=$(VERSION)-$(GIT_HASH)"

# Code generation
PROSEFORGE_REPO ?= ../proseforge
SWAGGER_SPEC := $(PROSEFORGE_REPO)/src/backend/docs/swagger/swagger.json
OPENAPI_SPEC := build/openapi3.json
GEN_DIR := internal/api/gen
OAPI_CODEGEN := $(shell go env GOPATH)/bin/oapi-codegen

## help: Show this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'

## build: Build both CLI and MCP binaries
build: build-cli build-mcp

## build-cli: Build the CLI binary
build-cli:
	go build $(LDFLAGS) -o $(BUILD_DIR)/pfw ./cmd/cli

## build-mcp: Build the MCP server binary
build-mcp:
	go build $(LDFLAGS) -o $(BUILD_DIR)/workbench-mcp ./cmd/mcp

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
	gofmt -w .
	goimports -w .

## test: Run all tests
test:
	go test ./...

## test-integration: Run integration tests (requires running API server)
test-integration:
	go test -tags integration -v -run TestIntegration ./...

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

## generate: Generate Go types and client from ProseForge Swagger spec
generate:
	@echo "Converting Swagger 2.0 → OpenAPI 3.0..."
	@mkdir -p build $(GEN_DIR)
	npx swagger2openapi $(SWAGGER_SPEC) -o $(OPENAPI_SPEC)
	@echo "Generating Go types and client..."
	$(OAPI_CODEGEN) --config oapi-codegen.yaml $(OPENAPI_SPEC)
	@echo "Running go mod tidy..."
	go mod tidy
	@echo "Done. Generated code in $(GEN_DIR)/"

## sync-github: Sync public content to GitHub (stage, verify, push)
sync-github:
	scripts/sync-github.sh

## sync-github-dry: Dry run — stage and verify, but don't push
sync-github-dry:
	scripts/sync-github.sh --dry-run

## sync-check: Run guardrail checks only (no staging)
sync-check:
	scripts/sync-github.sh --check
