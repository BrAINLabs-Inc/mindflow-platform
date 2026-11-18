# MindFlow monorepo — single entry point for every project.
#
#   make help          list all targets
#   make install       install dependencies for everything
#   make dev           run API + web admin + web app together
#   make check         what CI runs: lint, typecheck, build, test
#
# Each project keeps its own toolchain and lockfile; this file only
# orchestrates them. API_DIR is the Go API (backend-v2/). After cutover it
# is renamed (see FINDINGS_AND_DEVELOPMENT.md §11), and only this line
# changes.

API_DIR    ?= backend-v2
LEGACY_DIR ?= backend
MOBILE_DIR ?= mobile
ADMIN_DIR  ?= web-admin
WEB_DIR    ?= web-app

BASE_URL          ?= http://localhost:3000
IMAGE             ?= mindflow-api
ARTILLERY_VERSION ?= 2.0.33
PATCH             ?= database/patches/2026-11-18-security-hardening.sql

# The Express backend exists only until cutover; its targets disappear
# automatically once the directory is removed.
HAS_LEGACY := $(wildcard $(LEGACY_DIR)/package.json)

.DEFAULT_GOAL := help
.PHONY: help install dev dev-api dev-admin dev-web dev-mobile \
        build build-api build-admin build-web \
        lint lint-api lint-admin lint-web lint-mobile typecheck \
        test test-api test-api-short test-mobile test-legacy \
        check fmt sqlc sqlc-check seed docker db-patch \
        loadtest-token loadtest-stress loadtest-spike loadtest-soak loadtest-concurrency \
        legacy-dev clean

help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ── Setup ────────────────────────────────────────────────────────────────────

install: ## Install all dependencies (npm ci per app + go mod download)
	cd $(API_DIR) && go mod download
	cd $(MOBILE_DIR) && npm ci
	cd $(ADMIN_DIR) && npm ci
	cd $(WEB_DIR) && npm ci
ifneq ($(HAS_LEGACY),)
	cd $(LEGACY_DIR) && npm ci
endif

# ── Run ──────────────────────────────────────────────────────────────────────

dev: ## Run API + web admin (:5173) + web app (:5174) in parallel
	@$(MAKE) -j3 --no-print-directory dev-api dev-admin dev-web

dev-api: ## Run the Go API on :3000 (reads the API .env)
	cd $(API_DIR) && go run ./cmd/server

dev-admin: ## Run the web admin portal
	cd $(ADMIN_DIR) && npm run dev

dev-web: ## Run the participant web app
	cd $(WEB_DIR) && npm run dev

dev-mobile: ## Start the Expo dev server
	cd $(MOBILE_DIR) && npm start

# ── Build ────────────────────────────────────────────────────────────────────

build: build-api build-admin build-web ## Build API binary + both web apps

build-api: ## Build the API binary (bin/server)
	cd $(API_DIR) && go build -o bin/server ./cmd/server

build-admin:
	cd $(ADMIN_DIR) && npm run build

build-web:
	cd $(WEB_DIR) && npm run build

# ── Quality ──────────────────────────────────────────────────────────────────

lint: lint-api lint-admin lint-web lint-mobile ## Lint everything

lint-api: ## gofmt check + go vet
	@cd $(API_DIR) && out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	cd $(API_DIR) && go vet ./...

lint-admin:
	cd $(ADMIN_DIR) && npm run lint

# web-app and mobile lint are non-blocking, matching CI: web-app has
# known react-hooks v7 errors (FINDINGS_AND_DEVELOPMENT.md §7), mobile
# only warnings.
lint-web:
	-cd $(WEB_DIR) && npm run lint

lint-mobile:
	-cd $(MOBILE_DIR) && npm run lint

typecheck: ## Type-check mobile (web apps type-check during build)
	cd $(MOBILE_DIR) && npx tsc --noEmit

fmt: ## Format Go code
	cd $(API_DIR) && gofmt -w .

test: test-api test-mobile test-legacy ## Run all test suites (API tests need Docker)

test-api: ## All Go tests (starts throwaway Postgres containers via Docker)
	cd $(API_DIR) && go test ./...

test-api-short: ## Go tests that need no Docker
	cd $(API_DIR) && go test ./internal/auth/... ./internal/config ./internal/notify ./internal/r2 ./internal/researchgroup ./internal/timeutil ./cmd/seed
	cd $(API_DIR) && go test -run 'StoredContentType|ParseAboutMe' ./internal/handlers
	cd $(API_DIR) && go test -run 'ClientIP|BodyLimit|RequireAuth_|RateLimit' ./internal/middleware

test-mobile:
	cd $(MOBILE_DIR) && npm test

test-legacy:
ifneq ($(HAS_LEGACY),)
	cd $(LEGACY_DIR) && npm run lint && npm test
else
	@echo "legacy backend removed; skipping"
endif

check: lint typecheck build test ## Everything CI runs

# ── API tooling ──────────────────────────────────────────────────────────────

sqlc: ## Regenerate Go query code from internal/db/sql (needs sqlc v1.31.1 on PATH)
	cd $(API_DIR)/internal/db && sqlc generate -f sqlc.yaml

sqlc-check: ## Fail if generated sqlc code is stale
	cd $(API_DIR)/internal/db && sqlc diff -f sqlc.yaml

seed: ## Seed a DEV/STAGING Supabase project — never production
	cd $(API_DIR) && go run ./cmd/seed

docker: ## Build the API production image
	docker build -t $(IMAGE) $(API_DIR)

# ── Database ─────────────────────────────────────────────────────────────────

db-patch: ## Apply the DB security patch (idempotent): make db-patch DB_URL=postgres://...
	@test -n "$(DB_URL)" || (echo "DB_URL is required, e.g. make db-patch DB_URL=postgres://..."; exit 1)
	psql "$(DB_URL)" -v ON_ERROR_STOP=1 -f $(PATCH)

# ── Load tests (against a dev/staging server only) ───────────────────────────
# Each run logs in once to get AUTH_TOKEN (unless already exported), because
# /api/auth/login is rate-limited to 20 per 15 min. Artillery is fetched by
# npx on first use.

define artillery
	@token="$${AUTH_TOKEN:-$$(BASE_URL=$(BASE_URL) node $(API_DIR)/loadtest/get-token.js)}" && \
	BASE_URL=$(BASE_URL) AUTH_TOKEN="$$token" npx -y artillery@$(ARTILLERY_VERSION) run $(API_DIR)/loadtest/$(1).yml
endef

loadtest-token: ## Print a JWT for the seeded participant
	@BASE_URL=$(BASE_URL) node $(API_DIR)/loadtest/get-token.js

loadtest-stress: ## Ramp traffic to breaking point
	$(call artillery,stress)

loadtest-spike: ## Sudden burst, then recovery
	$(call artillery,spike)

loadtest-soak: ## Sustained moderate load
	$(call artillery,soak)

loadtest-concurrency: ## Concurrent logins + daily submits (race safety)
	$(call artillery,concurrency)

# ── Legacy Express backend (until cutover) ───────────────────────────────────

legacy-dev: ## Run the old Express backend on :3000
ifneq ($(HAS_LEGACY),)
	cd $(LEGACY_DIR) && npm run dev
else
	@echo "legacy backend removed"
endif

# ── Housekeeping ─────────────────────────────────────────────────────────────

clean: ## Remove build output
	rm -rf $(API_DIR)/bin $(ADMIN_DIR)/dist $(WEB_DIR)/dist
ifneq ($(HAS_LEGACY),)
	rm -rf $(LEGACY_DIR)/dist
endif
