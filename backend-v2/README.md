# backend-v2

Go (Gin) rewrite of `backend/`, built in parallel per
`plans/backend-go-migration.md`. Not yet a replacement for `backend/` — see
that plan for phased scope, cutover criteria, and architectural decisions.

## Status

Feature-complete against `backend/`'s route table (every `/api/*` endpoint
mobile and web-app call), with the cron reminder jobs. Not yet cut over:
load testing, the OpenAPI spec and the staging run are still open, per
the plan's "Progress" section and `FINDINGS_AND_DEVELOPMENT.md` at the
repo root.

## Development

Run everything from the **repo root** Makefile (`make help` lists targets).
Requires Go 1.25+. Tests that touch the database need Docker
(testcontainers starts a throwaway Postgres and applies
`database/project_db.sql`).

```bash
cp backend-v2/.env.example backend-v2/.env   # fill in DATABASE_URL, SUPABASE_*, R2_*
make dev-api           # go run ./cmd/server (reads backend-v2/.env)
curl http://localhost:3000/health
```

Common targets:

- `make dev-api` / `make build-api`: run / build `bin/server`.
- `make test-api`: all Go tests (Docker). `make test-api-short` runs the ones that need no Docker.
- `make lint-api`: gofmt check + `go vet`.
- `make sqlc`: regenerate `internal/db/queries` after editing `internal/db/sql/*.sql` (needs sqlc v1.31.1 on PATH; CI fails if generated code is stale).
- `make seed`: seed a **dev/staging** Supabase project (port of `backend/src/seed.ts`).
- `make docker`: build the production image (`Dockerfile`, distroless, non-root).
- `make loadtest-*`: Artillery scenarios against `BASE_URL` (see `loadtest/README.md`).
