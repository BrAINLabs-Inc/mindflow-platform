# MindFlow — Findings & Development Plan

_Audit date: 2026-11-18 · Branch: `development` @ `17daadf` · Scope: `backend-v2/` (Go migration), `backend/` (Express, for parity reference), `database/`, `mobile/`, `web-admin/`, with notes on `web-app/`._

Ground rule for this pass: **no participant-visible behaviour changes.** Every fix below either restores parity with the Express backend, closes a security hole that only an attacker would hit, or makes the same result cheaper to compute. Anything that would change product behaviour (new features, UI/UX redesign, the dashboard timezone question) is listed under **Proposals** and needs a decision first.

Status legend: ✅ fixed in this pass · 📝 documented / needs decision · ⏭️ deferred (larger work, plan below)

---

## 1. Summary

| Area | Critical | High | Medium | Low |
|---|---|---|---|---|
| backend-v2 parity | 1 | 1 | 2 | 2 |
| Security (all apps + DB) | – | 4 | 5 | 3 |
| Performance / scalability | – | 2 | 4 | 1 |
| Code quality / CI | – | – | 3 | 2 |

The headline items:

1. **backend-v2 is missing `POST /api/profile/about-me`.** Onboarding cannot be completed against the Go server, and the onboarding gate sends every new participant back to the About Me screen. Cutting over in the current state would have blocked new sign-ups. ✅
2. **Participants can change their own study arm.** The `profiles` UPDATE RLS policy has no column restriction, so anyone holding the public anon key and their own JWT can `PATCH profiles.research_id` straight through PostgREST, for example from `.cg` to `.ex`. That breaks the randomisation. ✅ (DB trigger)
3. **Participants can rewrite or delete their own research responses** through direct PostgREST writes (`FOR ALL` policies on the questionnaire tables, plus INSERT/UPDATE on `daily_sliders`, `voice_recordings` and `about_me_profiles`). This bypasses every server-side lockout and the cg field-nulling rule. Neither participant client writes to Supabase directly, so these policies can be read-only without any functional change. ✅
4. **Rate-limit bypass in backend-v2.** With `TRUST_PROXY_HOPS ≥ 1` (the default), Gin trusted *every* proxy, so a spoofed `X-Forwarded-For` header got a fresh rate-limit bucket on every request, including the auth brute-force limiter. ✅

---

## 2. backend-v2 (Go) — migration correctness

### 2.1 Parity gaps vs. Express

| # | Sev | Finding | Status |
|---|---|---|---|
| P1 | **Critical** | `POST /api/profile/about-me` (`updateAboutMe`) was never ported. The `profile.sql` comment deferred it to "when the write endpoints land", but phase 4/5 never picked it up. Mobile `AboutMe-Questionnaire` and web-app `AboutMePage` both call it. | ✅ Ported with Express's exact semantics: Zod-equivalent validation (trim, max lengths, `age` int 1–150, required-field check when `is_completed=true`), **partial upsert** (only fields present in the body are written, the rest keep their stored values, unset columns get table defaults on insert), and the full row is returned. |
| P2 | High | Daily entry: Express normalises falsy optional values to `NULL` (`feelings \|\| null`, `practice_duration \|\| null`, …). Go stored `""` and `0` verbatim. web-app sends `feelings: ''` when no factor is chosen, so the two servers wrote different research data for the same request. | ✅ Same normalisation applied in `services/daily.go`. |
| P3 | Medium | R2 public URL: Express falls back to `PUBLIC_URL_BASE` when `R2_PUBLIC_URL` is unset. Go only read `R2_PUBLIC_URL`, so an environment configured the legacy way would produce `fileUrl: "/WeeklyVoice/..."`. | ✅ Fallback added (the `VITE_`/`EXPO_PUBLIC_` fallbacks were deliberately **not** ported, per CLAUDE.md). |
| P4 | Medium | Request body limit: Express caps JSON bodies at 1 MB and returns `413 {"error":"Request body too large."}`. Go had no cap, so any authenticated or anonymous client could stream an arbitrarily large body into `ShouldBindJSON`. | ✅ `middleware.BodyLimit(1MB)` on every route except the multipart upload, which keeps its own 10 MB `MaxBytesReader`. |
| P5 | Low | `GET /api/journey` ran its 5 queries sequentially; Express used `Promise.all`. Same for `GET /api/dashboard/summary` (8 queries). | ✅ Both now run concurrently via `errgroup`. Results and error behaviour are identical: any error still fails the request, as in Express. |
| P6 | Low | `RegisterToken` with a syntactically invalid Expo token → 500 (Express also 500s, because the service throws). | 📝 Kept for parity. Recommend changing both to 400 after cutover. |

Verified correct (no change needed): FFMQ-15 facet scoring, R2 key scheme and prefix check, cg lockout and field-nulling asymmetry, journey `allSettled` fallback shape, about-me default shape, unregister scoping, username-collision retry, the dashboard's deliberately server-local streak maths, the trailing-slash route registration, and the 409 messages.

### 2.2 Security (backend-v2)

| # | Sev | Finding | Status |
|---|---|---|---|
| S1 | **High** | **Spoofable client IP.** `main.go` only called `SetTrustedProxies(nil)` when hops ≤ 0. Otherwise Gin's default is *trust all proxies*, so `c.ClientIP()` returned the left-most, attacker-controlled `X-Forwarded-For` entry. The 20/15 min auth limiter and the 300/15 min API limiter could be bypassed by rotating the header. | ✅ New `middleware.ClientIPResolver(hops)` reproduces Express's `trust proxy = N` exactly: it walks N hops in from the right of `X-Forwarded-For`, otherwise uses the socket address. Gin's own proxy trust is disabled. Unit-tested. |
| S2 | High | **Stored XSS on the R2 public domain.** The upload allow-list passes if *either* the MIME type *or* the extension matches, but the client-supplied MIME type was stored as the object's `Content-Type`. `file=x.wav` with `Content-Type: text/html` was therefore served as HTML from the public bucket. Express has the same bug. | ✅ (both backends) The stored Content-Type is now the client MIME only if it is on the audio allow-list, otherwise it is derived from the extension. |
| S3 | High | **Arbitrary `file_url` stored for weekly recordings.** `file_key` was prefix-checked but `file_url` was not, so a participant could store `javascript:…` or a phishing link that researchers would later open from exports. Same bug in Express. | ✅ (both backends) `file_url` must equal the public URL the server itself derives for `file_key`. The legitimate client flow (echoing back `fileUrl` from `/upload`) is unaffected. |
| S4 | Medium | No HTTP server timeouts (`ReadHeaderTimeout`, `ReadTimeout`, `IdleTimeout`), which leaves the server open to slow-loris connection exhaustion. | ✅ Timeouts set. The write timeout leaves room for the 10 MB upload. |
| S5 | Medium | JWT verification didn't *require* `exp` (golang-jwt v5 only validates it if present), and a non-UUID `sub` produced a 500 in every handler rather than a 401. | ✅ `jwt.WithExpirationRequired()`. The `sub` claim must parse as a UUID, otherwise 401 `Invalid token`. |
| S6 | Medium | Local JWT verification cannot see server-side session revocation (sign-out, password change) until the access token expires (Supabase default 1 h). Express's `getUser` round trip did see it. | 📝 Inherent to the plan's design. Keep the Supabase JWT expiry at ≤ 1 h. If immediate revocation is ever required, add a `session_id` deny-list check. |
| S7 | Low | Auth proxy error messages are passed through verbatim (e.g. "User already registered"), which enables account enumeration. Express does the same. | 📝 Parity kept. Recommend generic messages for signup and reset after cutover. |

### 2.3 Scalability (backend-v2)

| # | Sev | Finding | Status |
|---|---|---|---|
| C1 | **High** | **Cron jobs double-send on horizontal scale.** Every instance runs the 08:00 and 19:00 jobs, so with N replicas each participant gets N copies. | ✅ Each job runs inside a transaction holding `pg_try_advisory_xact_lock`: transaction-scoped, so it stays correct behind Supavisor's transaction-mode pooler. Only one instance per tick sends; the others log and skip. Tested with a simulated second replica. |
| C2 | High | The 19:00 reminder job ran 5 status queries per user *sequentially* (≈ 5 × N round trips; ~1 min+ at 1 000 users over a pooler). It also silently swallowed per-user errors, which Express logs. | ✅ Bounded concurrency (4 workers, leaving most of the 25-connection pool for live traffic). Message order and skip-on-error behaviour are unchanged, and the error log line is restored. |
| C3 | Medium | `time.LoadLocation("Asia/Colombo")` fails at boot on minimal container images that don't ship tzdata (distroless/scratch/alpine). | ✅ `import _ "time/tzdata"` embeds the database. |
| C4 | Medium | The in-memory rate-limit store is per instance. Behind a load balancer with N replicas the effective limit is N×. | ⏭️ Switch `ulule/limiter` to its Redis store when you scale past one instance (config-only change, documented in `main.go`). |
| C5 | Low | `MaxConns=25` is hard-coded. | ⏭️ Make it env-configurable (`DB_MAX_CONNS`) once load tests (phase 7) give real numbers. |

### 2.4 Still open from the migration plan (unchanged by this pass)

- Phase 7: Artillery parity and load tests against backend-v2, plus the item-by-item Jest→Go coverage cross-check.
- §6: OpenAPI spec and the parity-check script.
- §9: cutover.
- ✅ §8: CI job, added in this pass (`backend-v2` job: `go vet`, `go build`, `go test` with testcontainers).

---

## 3. Database (`database/`)

| # | Sev | Finding | Status |
|---|---|---|---|
| D1 | **High** | `profiles` "Users can update own profile" has no column restriction, so a participant can rewrite `research_id` (switch arm, or claim another participant's unassigned ID) via PostgREST with the public anon key. | ✅ `BEFORE INSERT OR UPDATE` trigger `guard_profiles_research_id`: a change to `research_id` made under a participant JWT is rejected unless the caller is in `admins`. Service-role, backend, seed and admin-portal writes are unaffected. |
| D2 | **High** | The questionnaire tables have `FOR ALL USING (user_id = auth.uid())`, and `daily_sliders`/`voice_recordings` have participant INSERT/UPDATE policies. Participants can therefore insert unlimited PSS-10/FFMQ-15/WEMWBS-14 rows (bypassing the advisory-lock RPC), edit past scores, delete responses, or write mindfulness fields as a control-group participant. | ✅ Participant policies narrowed to `SELECT`. All participant writes already go through the backend, which uses a service-role/BYPASSRLS connection. Admin policies are unchanged. |
| D3 | Medium | `about_me_profiles` `FOR ALL` lets a participant flip `is_completed` without filling the fields. Low impact because the gate is UX-only, but it is still research data. | ✅ Participant policy narrowed to `SELECT`. |
| D4 | Medium | `backend_v2_role.sql` granted `DELETE` on every table and write access to admin-managed tables (`calendar_events`, `weekly_recordings`), which is broader than the Go code needs. | ✅ Narrowed to what the queries use: DELETE only on `push_tokens`, read-only on admin-managed tables, INSERT/SELECT on questionnaire tables. |
| D5 | Medium | **Missing composite indexes.** Every status, streak and journey query has the shape `WHERE user_id=$1 AND created_at>=$2 ORDER BY created_at DESC LIMIT n`, but only single-column indexes exist, so Postgres must merge or filter. | ✅ `(user_id, created_at DESC)` on `daily_sliders`, `voice_recordings` and the 3 questionnaire tables. The now-redundant single-column `user_id` indexes are left in place (cheap; can be dropped later). |
| D6 | Low | `project_db.sql` is not re-runnable on an existing database: `"Users can update own profile"`/`"Users can insert own profile"` are created without a preceding `DROP POLICY IF EXISTS`. | ✅ Guards added. |

**To apply to an existing database:** run `database/patches/2026-11-18-security-hardening.sql` (idempotent; contains exactly D1, D2, D3, D5, D6). `project_db.sql` carries the same changes for fresh bootstraps. Re-run `backend_v2_role.sql` for D4 (it now `REVOKE`s before granting).

---

## 4. Express backend (`backend/`)

Still the production server until cutover. Only the two shared security bugs were patched here; everything else is frozen per the migration plan.

- ✅ S2 (stored Content-Type) and S3 (`file_url` must match `file_key`): same fixes as Go, with Jest tests added.
- 📝 Uses `supabase.auth.getUser` per request (network hop). This is the reason for the migration, so it is not being optimised here.

---

## 5. Mobile (`mobile/`)

| # | Sev | Finding | Status |
|---|---|---|---|
| M1 | Medium | **Sign-out leaves the device registered for push**, so after a participant signs out the device keeps receiving "Hi \<name\>, …" reminders that name them (privacy on shared or handed-down devices). | ✅ `handleSignOut` makes a best-effort `POST /api/notifications/unregister-token` *before* clearing the session. Failure never blocks sign-out. |
| M2 | Low | 2 ESLint errors (`react/display-name` in `DashboardScreen`, unescaped `'` in `ForgotPasswordScreen`). | ✅ Fixed. The 34 pre-existing `exhaustive-deps` warnings on Reanimated shared values are false positives and were left alone. |
| M3 | Low | `apiFetch` has no request timeout, so on a dead connection a screen spinner can hang indefinitely. | 📝 Proposal: a 20 s `AbortController` timeout on non-upload requests. Not done here, because it changes what users see on slow networks. |

Verified good: tokens in SecureStore with a legacy migration, single-flight refresh, 401→refresh→retry, `onSessionExpired` pub/sub, and the onboarding gate defaulting to `MainTabs` on failure.

---

## 6. Web admin (`web-admin/`)

| # | Sev | Finding | Status |
|---|---|---|---|
| A1 | Medium | **CSV formula injection.** Participant free text (`feelings`, about-me fields, usernames) beginning with `=`, `+`, `-`, `@`, tab or CR is exported raw. Opening the export in Excel/Sheets executes it as a formula (data exfiltration via `HYPERLINK`/`WEBSERVICE`). | ✅ String cells starting with those characters get a leading `'` (OWASP recommendation). Numbers, booleans and UUIDs are untouched. |
| A2 | Medium | **Analytics silently truncates at 1 000 rows.** `loadAnalyticsData` issues unbounded `select('*')` calls, and PostgREST caps each at `max-rows` (1 000 by default). With 1 000+ participants, a 30-day window of `daily_sliders` was already truncated, so every chart under-reported without any error. | ✅ The six analytics queries are paged (1 000 per page), the same pattern `useUserMap` already uses. |
| A3 | Low | `npm run lint` crashed (`reactHooks.configs.flat` is undefined in eslint-plugin-react-hooks 5.x). CI marks web-admin lint non-blocking, so this went unnoticed. Once it ran, it reported 5 real errors. | ✅ Uses `configs['recommended-latest']`. The 5 errors are fixed (an `any`, an empty interface, and a shadcn `ui/` override for react-refresh). web-admin lint is now **blocking** in CI. |
| A4 | Low | The overview "check-ins last 7 days" chart buckets by UTC date, while the study day is Sri Lanka time (entries made 00:00–05:29 SLT land on the previous day). | 📝 Display-only. Fix together with the dashboard timezone decision (§9). |

Verified good: the admin gate is enforced by RLS (`admins` table), not just `ProtectedRoute`; `.or()` search input is escaped; profiles are paged.

---

## 7. Web app (`web-app/`): not changed in this pass, findings only

- 📝 **No token refresh.** The web app stores only the access token (localStorage) and never calls `/api/auth/refresh`, so web participants are effectively logged out after ~1 h with confusing errors. Recommend porting mobile's `apiFetch` refresh logic.
- 📝 Access token in `localStorage` is readable by any XSS. Acceptable while the app has no third-party scripts. Long term, use an httpOnly cookie via the backend.
- 📝 17 lint errors under react-hooks v7's new React Compiler rules (setState-in-effect, components created during render in `AboutMePage`, `Date.now()` during render). The components defined inside `AboutMePage`'s render remount on every keystroke and can drop input focus. **Worth fixing first.**

---

## 8. Verification performed

All of the following were run after the changes:

| Project | Command | Result |
|---|---|---|
| backend-v2 | `gofmt -l .`, `go vet ./...`, `go build ./...` | clean |
| backend-v2 | `go test -count=1 ./...` (testcontainers Postgres 17) | all packages pass |
| backend-v2 | `sqlc diff` (v1.31.1, same version that generated the existing code) | generated code in sync |
| backend | `npm run lint`, `npm run build`, `npx jest` | clean; **114/114** tests pass (9 new) |
| mobile | `npx tsc --noEmit`, `npm test`, `npx eslint .` | clean; 15/15 tests pass; **0** lint errors |
| web-admin | `npx eslint .`, `npm run build` | clean |

New tests, all against real Postgres where the database is involved:
- **DB security:** as a participant JWT, changing your own `research_id` fails (42501) while a username update succeeds, an admin can reassign, and a backend connection is unaffected. Direct participant INSERT/UPDATE on `daily_sliders`, `voice_recordings`, `questionnaire_pss10_responses` and `about_me_profiles` is blocked, while reading your own rows still works. The patch file applies twice without error. **These tests were also run against the original `project_db.sql`, and all five vulnerability checks failed there, confirming they reproduce the holes.**
- **About-me:** partial-upsert semantics (absent field kept, explicit null cleared, insert path uses defaults) and the 15-case validation table (Zod parity: unknown keys stripped, `is_completed:null` rejected, UTF-16 max lengths, required-when-completed).
- Daily `""`/`0` → `NULL` normalisation. Weekly `file_url` mismatch rejection (3 attack URLs) and stored Content-Type (Go and Express). Client-IP resolution (7 proxy-addr cases). 1 MB body limit, including chunked bodies with no Content-Length. JWT without `exp` / non-UUID `sub` → 401. Scheduler single-instance lock.

Not verified here (no credentials or infrastructure in this environment): a real Supabase project, R2, Expo push delivery, and Artillery load tests. Those are still the phase-7 gate.

## 9. Proposals that need a decision (not implemented)

You asked for scalability, new features, and UI/UX improvements. Scalability groundwork is done above (C1–C3, D5, A2). The items below change what participants or researchers see, so they need sign-off, and some need research-team approval because they could bias the study.

### 9.1 Decisions pending
1. **Dashboard streak timezone.** `GET /api/dashboard/summary` computes streaks in server-local time; everything else uses Sri Lanka time. On a UTC host, entries made 00:00–05:29 SLT count toward the previous day. Both backends replicate this deliberately. Fix (one line in each backend) once the research team agrees, since it will shift some participants' displayed streaks.
2. **Generic auth error messages** (S7) and **400 for invalid push tokens** (P6) after cutover.

### 9.2 Feature ideas (ranked by value / effort)
| Feature | Why | Effort | Study-design risk |
|---|---|---|---|
| Web-app token refresh (§7) | Fixes web participants being kicked out hourly | S | none |
| Offline queue for Daily Sliders (mobile) | Rural or university Wi-Fi drops currently lose a submission | M | none |
| Researcher "participant timeline" view in web-admin (one participant's daily, weekly and questionnaire history on one page) | Today researchers must cross-reference 5 tables | M | none |
| Compliance alerts for researchers (participants with no entry in N days) | Early drop-out detection | M | none |
| Configurable reminder time per participant | 19:00 doesn't suit everyone | M | low: changes the reminder protocol, needs sign-off |
| Personal trend charts for participants (stress/sleep over time) | Engagement | M | **high**: feedback can itself be an intervention. Needs a research decision, possibly ex-only. |
| OpenAPI spec + generated TS client for mobile/web-app | Removes hand-typed response shapes in 2 clients | M | none |

### 9.3 UI/UX ideas
- Web-app `AboutMePage`: lift the inner components out of render (fixes focus loss while typing). This is a bug fix as much as UX.
- Skeleton loaders instead of full-screen spinners on Dashboard/Journey (mobile and web-app).
- Explicit offline/error banners with a Retry button (today most failures only `console.log`).
- Accessibility pass: slider `accessibilityLabel`s, dynamic-type support, colour-contrast check on the roadmap node states.
- Web-admin: a sticky filter bar and URL-synced filters/tabs, so researchers can bookmark or share a filtered view.

---

## 10. Development roadmap

1. **Now (this pass):** everything marked ✅ above.
2. **Before cutover:** apply the DB patch to staging, then production. Run Artillery (`make loadtest-*`) against backend-v2 and compare with the baselines. Write the OpenAPI spec and parity-check script (plan §6). Do a staging side-by-side run.
3. **At cutover:** Redis-backed rate limiter if running more than one instance (C4). Env-configurable pool size (C5). Flip `EXPO_PUBLIC_API_URL` / `VITE_API_BASE_URL`.
4. **After cutover:** delete `backend/`, rename `backend-v2/` → `api/`, update CLAUDE.md and the docs (§11). Then the decided items from §9.

The step-by-step version of this roadmap, with commands and checkboxes, is **`plans/cutover-and-cleanup.md`**.

---

## 11. Removing the old backend and renaming `backend-v2/` → `api/`

### 11.1 Is the migration complete?

**Not yet.** The code side is now complete: backend-v2 serves every route the clients call, and it no longer depends on anything in `backend/`. The rollout side still has steps that need your infrastructure:

| Requirement | Status |
|---|---|
| Every `/api/*` route the clients call, with Express parity | ✅ (about-me POST added in this pass) |
| Seed script (`npm run seed`) | ✅ ported to `backend-v2/cmd/seed` → `make seed` |
| Load-test scenarios | ✅ copied to `backend-v2/loadtest` → `make loadtest-*` (not run) |
| `.env` loading like dotenv | ✅ fixed (it was missing, so `make dev-api` failed with a fresh `.env`) |
| Production container image | ✅ `backend-v2/Dockerfile` → `make docker` |
| CI job | ✅ |
| DB patch + least-privilege role applied to the real database | ⏳ **you**: `make db-patch DB_URL=...`, then run `database/backend_v2_role.sql` with a real password |
| Load tests run against a staging server | ⏳ **you**: needs a staging Supabase + R2 (`make seed`, `make loadtest-*`) |
| Go API deployed and clients switched | ⏳ **you**: deploy the image, set `EXPO_PUBLIC_API_URL` / `VITE_API_BASE_URL`, watch logs for a soak period |

Production currently runs `backend/` from a host configured outside this repo (there is no deploy config in the repo). Deleting `backend/` before the last three rows are done risks the next production redeploy failing.

### 11.2 Removal and rename checklist (run once 11.1 is all ✅)

1. Confirm all production traffic is on the Go API (its logs show requests; the Express service shows none). Then decommission the Express deployment.
2. `git rm -r backend/` and `git mv backend-v2 api`.
3. In `api/go.mod`, change the module path suffix `/backend-v2` → `/api` and update every import under `api/` (a single `sed` over `*.go`).
4. Root `Makefile`: `API_DIR ?= api`. The `legacy-*`/`test-legacy` targets switch off automatically once `backend/` is gone.
5. `ci.yml`: delete the Node `backend` job and rename the `backend-v2` job to `api` (`working-directory: ./api`, `go.mod`/`go.sum` paths).
6. Rename the Postgres role in `database/backend_v2_role.sql` → `database/api_role.sql` (role name can stay; renaming it means a new `DATABASE_URL`).
7. Update `CLAUDE.md`, `README.md`, `docs/architecture.md`, `docs/getting-started.md` (Go toolchain + `make` instead of npm for the API) and the `backend/...` references in code comments (`grep -rn "backend/src" api/`).
8. Move `plans/backend-go-migration.md` → `docs/decisions/0001-go-backend.md`.

Each step is mechanical. `git mv` keeps history, and CI catches any missed path.

### 11.3 Proposed repo structure after the rename

```
mindflow-platform/
├── api/                  # Go API (was backend-v2/)
├── mobile/               # Expo participant app
├── web-app/              # Vite participant web app
├── web-admin/            # Vite researcher portal
├── database/
│   ├── project_db.sql    # bootstrap source of truth
│   ├── api_role.sql      # least-privilege role (was backend_v2_role.sql)
│   └── patches/          # dated, idempotent, for live DBs
├── docs/
│   ├── architecture.md, database.md, features.md, getting-started.md, research-methodology.md
│   ├── decisions/        # ADRs (was plans/)
│   └── audits/           # this file, once actioned
├── .github/workflows/
├── Makefile              # single entry point: make help
└── CLAUDE.md, README.md, CONTRIBUTING.md, LICENSE.md
```

The app folders keep their current names, so the mobile EAS build workflows and the web hosting configs don't need to change. That keeps the rename PR small and low-risk.

---

## 12. What else still needs correcting (open items)

In priority order:

1. **Web-app token refresh** (§7). Web participants are logged out about every hour, and the app shows confusing errors when it happens. Port mobile's `apiFetch` refresh logic.
2. **Web-app `AboutMePage` input focus bug** (§7). Components defined during render remount on every keystroke.
3. **Web-app lint (17 errors).** Mostly real React issues, not style. Fixing them lets web-app lint become blocking in CI, as web-admin's now is.
4. **Dashboard streak timezone** (§9.1). Needs a research-team decision, then a one-line fix in each backend.
5. **Generic auth error messages + 400 for invalid push tokens** (S7/P6). Do this after cutover so both backends don't diverge.
6. **Redis-backed rate limiter** before running more than one API instance (C4). Make the pool size configurable (C5).
7. **OpenAPI spec** (plan §6). It is the contract between the API and both clients.
8. **Mobile request timeout** (M3).
9. **web-admin bundle size.** The Vite build warns that the main chunk is over 500 kB. Code-split the dashboard/analytics tabs.
10. **web-admin overview chart buckets by UTC** (A4). Fix together with item 4.
