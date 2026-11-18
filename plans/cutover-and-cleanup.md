# Plan: Go API cutover, `backend/` removal, `backend-v2/` → `api/`, and open fixes

_Written 2026-11-18. Owner: you. Background and evidence: `FINDINGS_AND_DEVELOPMENT.md` (§11–12). The code side of the migration is done; everything below needs real infrastructure or a decision, so it was left for you._

Work through the phases in order. Each phase ends with a **gate**: don't start the next phase until it's met.

---

## Phase 1 — Database (staging first, then production)

- [ ] Apply the security and index patch to **staging**:
  ```bash
  make db-patch DB_URL="postgresql://postgres:<pw>@<staging-host>:5432/postgres"
  ```
  It's idempotent, so running it twice is safe.
- [ ] Create the API's least-privilege role on staging. Open `database/backend_v2_role.sql`, replace `REPLACE_ME_BEFORE_RUNNING` with a generated password (**don't commit it**), and run it with `psql`.
- [ ] Smoke-test web-admin against staging. The patch narrows participant policies, so check that admins can still edit About Me rows, assign groups, and manage calendar events and weekly videos.
- [ ] Repeat both steps on **production**.

**Gate:** web-admin works on production after the patch, and `SELECT * FROM pg_roles WHERE rolname='app_backend_v2'` returns one row.

---

## Phase 2 — Staging run of the Go API

- [ ] `cp backend-v2/.env.example backend-v2/.env`, then fill in the staging values: `DATABASE_URL` as `app_backend_v2` via the Supavisor pooler, plus `SUPABASE_*`, `R2_*`, and `CORS_ORIGINS`.
- [ ] Check which JWT signing mode the Supabase project uses (Dashboard → Settings → API → JWT Keys). Set `JWT_VERIFICATION_MODE=jwks`, or `hs256` together with `SUPABASE_JWT_SECRET`.
- [ ] `make seed`, which creates `seed.participant@mindflow.app` / `seed.admin@mindflow.app` on **staging only**.
- [ ] `make build-api && (cd backend-v2 && NODE_ENV=production ./bin/server)`
- [ ] Point a staging build of the clients at it: mobile `EXPO_PUBLIC_API_URL`, web-app `VITE_API_BASE_URL`. Then walk the whole flow by hand:
  - [ ] signup → OTP → About Me (both partial save and complete) → dashboard
  - [ ] Daily Sliders as an **ex** user (can resubmit the same day) and as a **cg** user (second submit gets 409, and the mindfulness fields are stored as NULL)
  - [ ] Weekly Whispers: record → upload → submit, then confirm the audio plays from the R2 URL
  - [ ] Thrive / Stress / Mindful submit, then a resubmit gets 409
  - [ ] calendar (a cg user doesn't see "Mindfulness Session" events), journey, profile, sign-out
  - [ ] Forgot password → reset code → new password → login
- [ ] Load tests against staging (from the repo root, with `BASE_URL` set to the staging API):
  ```bash
  make loadtest-concurrency   # MUST pass: exactly one 200, the rest 409, no 500s or duplicate rows
  make loadtest-stress
  make loadtest-spike
  make loadtest-soak
  ```
  Compare p95 and error rates with the old backend's numbers on the same staging data.

**Gate:** every manual flow passes, `loadtest-concurrency` passes, and the other scenarios meet their `ensure` thresholds (p95 < 1000 ms).

---

## Phase 3 — Production cutover

- [ ] Deploy the container: `make docker`, then push the `mindflow-api` image to your host (Render, Railway, Fly, a VM — anything that runs a container). Set the production env vars in the host's secret store, not in a file.
- [ ] Health check: `curl https://<api-host>/health`.
- [ ] Set `TRUST_PROXY_HOPS` to match the host's proxy chain. It's usually 1, the default. This value is what makes rate limiting use the real client IP.
- [ ] If you run **more than one** instance, switch the rate limiter to Redis first (`backend-v2/cmd/server/main.go`, `limitermemory` → ulule's Redis store). Otherwise each instance enforces its own separate limit.
- [ ] Switch the clients:
  - mobile: update `EXPO_PUBLIC_API_URL` in the EAS production env, then ship a build or OTA update
  - web-app: update `VITE_API_BASE_URL` on its host and redeploy
- [ ] Keep the Express deployment running, idle, as a rollback target. Rollback is just pointing the URLs back.
- [ ] Soak for an agreed period (suggested: 2 weeks, covering at least one 08:00/19:00 reminder cycle each day). Watch for:
  - [ ] no 5xx spikes in the API logs
  - [ ] **one** reminder per participant per run (the log line "skipped: another instance is running it" appears only if you have more than one instance)
  - [ ] new sign-ups complete onboarding

**Gate:** soak period clean, and the Express service receives no traffic.

---

## Phase 4 — Remove `backend/`, rename `backend-v2/` → `api/`

Do this as **one PR** on a branch from `development`.

- [ ] Decommission the Express deployment on the host.
- [ ] Remove and rename:
  ```bash
  git rm -r backend
  git mv backend-v2 api
  ```
- [ ] Change the Go module path:
  ```bash
  cd api
  go mod edit -module github.com/BrAINLabs-Inc/mindflow-platform/api
  grep -rl 'mindflow-platform/backend-v2' --include='*.go' . \
    | xargs sed -i 's#mindflow-platform/backend-v2#mindflow-platform/api#g'
  gofmt -l .   # should print nothing
  ```
- [ ] Root `Makefile`: `API_DIR ?= api`. The `legacy-dev` and `test-legacy` targets turn themselves off once `backend/` is gone; delete them and `LEGACY_DIR` for tidiness.
- [ ] `.github/workflows/ci.yml`: delete the Node `backend` job. Rename the `backend-v2` job to `api`, with `working-directory: ./api`, `go-version-file: api/go.mod`, and `cache-dependency-path: api/go.sum`.
- [ ] `git mv database/backend_v2_role.sql database/api_role.sql`. Keep the role name `app_backend_v2` unless you also want to rotate `DATABASE_URL`.
- [ ] Fix the remaining references:
  ```bash
  grep -rn "backend-v2\|backend/src\|backend/" --include='*.md' --include='*.go' --include='*.yml' --include='Makefile' . | grep -v node_modules
  ```
  This covers `CLAUDE.md`, `README.md`, `docs/architecture.md`, `docs/getting-started.md` (setup becomes Go + `make`, not npm), `api/README.md`, `api/loadtest/README.md`, and code comments that say "port of backend/src/...". Comments can point at the git history.
- [ ] Move the migration plan to an ADR: `mkdir -p docs/decisions && git mv plans/backend-go-migration.md docs/decisions/0001-go-backend.md`.
- [ ] Verify: `make check` passes locally (Docker needed for the API tests), CI is green, then merge.

**Gate:** CI is green on `development`, and a fresh clone runs `make install && make dev`.

---

## Phase 5 — Open fixes (any time; independent of phases 1–4)

Ranked by impact. Details are in `FINDINGS_AND_DEVELOPMENT.md` §7, §9 and §12.

1. [ ] **Web-app token refresh.** Web participants are logged out about every hour. Port `mobile/src/lib/apiClient.ts`'s refresh logic (single shared refresh, 401 → refresh → retry, session-expired event) into `web-app/src/lib/api.ts` and `web-app/src/lib/auth.ts`, and store the refresh token alongside the access token.
2. [ ] **Web-app About Me focus bug.** Move the components defined inside `AboutMePage`'s render (lines ~249–441) to module level.
3. [ ] **Web-app lint, 17 errors** (setState inside effects, `Date.now()` during render). Then make web-app lint blocking in `ci.yml` and in the Makefile's `lint-web` (remove the leading `-`).
4. [ ] **Dashboard streak timezone.** Ask the research team first: switching streaks to Sri Lanka time shifts some participants' numbers. Then make a one-line change in `api/internal/services/dashboard.go` (use `timeutil.SLT`), and fix the web-admin overview chart's UTC bucketing at the same time.
5. [ ] **After cutover:** generic auth error messages (no account enumeration) and 400 instead of 500 for an invalid push token.
6. [ ] Make `DB_MAX_CONNS` configurable, tuned from the phase-2 load-test results.
7. [ ] OpenAPI spec at `api/api/openapi.yaml` (migration plan §6).
8. [ ] Mobile 20 s request timeout (`AbortController`) on non-upload requests.
9. [ ] Code-split web-admin's dashboard and analytics tabs. The main chunk is over 500 kB.
