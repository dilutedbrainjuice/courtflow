# CourtFlow

Sports facility booking and management platform planned with Go, Next.js and
PostgreSQL. This repository currently contains development configuration and
documentation only.

## Layout

- `DEVELOPMENT.md`: permanent development requirements; read before making changes.
- `scripts/`: local environment checks.
- `buildDocs/buildDocs/CourtFlow/`: Obsidian architecture, flow analysis, build log and handoff.
- Future application directories: `backend/` (Go) and `frontend/` (Next.js).

## Local workflow

Run `make check-env` to inspect tools and `make check` for repository checks.
No application dependencies, build commands or application tests exist yet.
Do not interpret repository checks as application test coverage.

At the first implementation, add Go modules and npm manifests with a lockfile,
then document exact install, development, build, lint and test commands here.
Use `gofmt` for Go; configure one frontend formatter and linter at that time.
Run Next.js and Go locally and use local PostgreSQL with mock payments.

Podman is installed. The intended database setup is a local PostgreSQL container
with a named volume and a port bound to `127.0.0.1` only. Select and pin its image
version and provide placeholder environment configuration when provisioning it.
No image has been downloaded and no database has been created. A native local
PostgreSQL installation is an alternative if Podman cannot run.

Run relevant local tests before each completed implementation changeset, update
flow documentation, append the build log, update handoff, and inspect the Git diff.
Never use production data for tests. Deployment requires passing local tests and
explicit approval. No remote repository or deployment is configured; the eventual
GitHub repository must be public. Keep secrets and personal vault settings out of Git.

## V1 implementation (supersedes the initial setup status above)

The Go API and Next.js customer booking interface are implemented. Backend uses
`net/http`, pgx and PostgreSQL exclusion constraints. Frontend uses Next.js App
Router, TypeScript and npm. Four seeded courts support immediate simulated guest
bookings, with MYT display, conflict feedback and server-calculated prices.

### Reproduce locally

```bash
make dev-db
make migrate-local
(cd backend && go mod download)
(cd frontend && npm ci && npx playwright install chromium)
make dev-api   # terminal 1: http://localhost:8080
make dev-web   # terminal 2: http://127.0.0.1:3000
```

`make dev-api` uses the documented development-only password. If you override
`COURTFLOW_LOCAL_DB_PASSWORD`, also override `LOCAL_DB_URL` when invoking Make.
Never use that password outside localhost. `.env.example` files describe settings;
Go does not automatically load `.env` files. Supply environment variables explicitly.

```bash
make verify-local       # Go vet/tests/race/build; npm ci/lint/typecheck/build
make test               # integration tests in isolated schemas, including 24 contenders
(cd frontend && npm test) # mocked browser states and responsive layout
bash scripts/persistence-local.sh # restarts local DB/API; removes its test booking
podman build -t localhost/courtflow-api:v1 backend
```

The persistence check requires its future court D test slot to be free. Database
migrations run in a transaction with ON_ERROR_STOP. The initial migration is
repeatable; future schema changes require new numbered migrations, never edits to
an already deployed migration. No production migrations run automatically at startup.

Formatting: Go `gofmt`; frontend Prettier. Lint: `go vet`, ESLint. Tests use Go's
standard testing package and Playwright. No cloud test services are used.

### Local production-image smoke

```bash
podman run -d --name courtflow-api-local --network host -e PORT=18081 \
  -e DATABASE_URL='postgres://courtflow:courtflow_dev_only@127.0.0.1:5433/courtflow?sslmode=disable' \
  localhost/courtflow-api:v1
bash scripts/container-smoke-local.sh
podman stop courtflow-api-local
podman rm courtflow-api-local
```

The API binds all interfaces as required by Render; use local firewall protections
when testing with host networking. PostgreSQL binds only 127.0.0.1:5433.

### Boundaries

No user identities, real payments, cancellation or privileged operations endpoints.
Availability is a public occupancy view without booking details. The optional
separate reception list is deferred. Basic booking abuse control caps total
attempts at 120/minute per process; it is deliberately simple and resets on restart.
Health checks report process liveness, not database readiness.

See `docs/PRODUCTION_RUNBOOK.md` for the approval-gated release. No remote exists,
no hosting resources have been created, and no deployment has occurred.
