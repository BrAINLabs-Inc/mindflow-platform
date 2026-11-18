# Load & reliability tests (Artillery)

Copied from `backend/loadtest/` per `plans/backend-go-migration.md` §7. The
scenarios are plain HTTP against `BASE_URL`, so they run unchanged against
either backend. They keep the same `ensure` thresholds (p95 < 1000 ms, error
rates per scenario) as the acceptance bar for cutover. Run all targets from
the repo root.

**Never point these at production.** Use a local server backed by a
staging/dev Supabase project.

```bash
make seed                      # creates seed.participant@mindflow.app once
make build-api && (cd backend-v2 && NODE_ENV=production ./bin/server) &

export BASE_URL=http://localhost:3000
make loadtest-stress           # ramps traffic until it slows/fails
make loadtest-spike            # sudden burst, then recovery
make loadtest-soak             # sustained moderate load (10 min sanity run)
make loadtest-concurrency      # simultaneous logins + Daily Slider submits
```

Each target fetches `AUTH_TOKEN` once via `get-token.js` (unless you export
it yourself), because `/api/auth/login` sits behind the 20 / 15 min auth rate
limiter and per-user logins would only measure the limiter.

`concurrency.yml`'s concurrent daily submit (exactly one 200, the rest 409,
never a 500 or duplicate row) is the most important scenario: it validates
the advisory-lock race safety through the whole HTTP stack.

Requires Node 20+ (`npx` fetches Artillery 2.0.33, the version `backend/`
pinned).
