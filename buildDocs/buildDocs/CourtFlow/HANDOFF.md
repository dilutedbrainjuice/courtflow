# CourtFlow handoff

## Current development state — 2026-10-31
Local V1 implemented. Git branch main; no remote or deployment. See BUILD_LOG for
append-only results and local Git history for commits. Original vault content and
paths preserved. Read AGENTS, DEVELOPMENT, CODEX_START_HERE and WORKFLOW_ADDENDUM.

## Architecture and implemented features
Next.js/TypeScript customer page → Go net/http API → PostgreSQL 17. pgx pool max 8.
Five APIs: health, facilities, courts, availability, create booking. Four fictional
badminton courts, MYT display, one-hour sessions, 30-day horizon, server pricing,
simulated immediate confirmation, conflict handling. PostgreSQL GiST exclusion
constraint prevents confirmed overlap; 23P01 maps to 409. No PII collected.
Timeouts, JSON size/field validation, exact CORS and process-wide 120/minute booking
attempt cap. No authentication/admin writes, cancellation, holds or real payments.
Optional reception booking list not implemented; availability shows occupancy.

## Paths and commands
Root /home/terasit/VSProjects/court; vault buildDocs/buildDocs; project docs beneath
CourtFlow. `make dev-db`, `make migrate-local`, `make dev-api`, `make dev-web`.
`make verify-local` installs frontend from lockfile and runs Go vet/tests/race/build,
frontend lint/typecheck/build. `make test` uses local isolated-schema integration.
`cd frontend && npm test` runs Playwright; first install Chromium with
`npx playwright install chromium`. `bash scripts/persistence-local.sh` checks API
and DB restart durability. README documents production-image local smoke commands.
PostgreSQL named volume courtflow_pgdata persists; container courtflow-postgres.

## Verified results
Go unit/integration/race/vet/build passed. 24 concurrent same-slot requests: one
confirmation, 23 conflicts, one stored row. Adjacent/different courts succeed.
Invalid requests, CORS and rate cap verified. Database migration repeatability and
API/Postgres restart persistence passed. Podman image build and real HTTP container
health/booking smoke passed. Frontend lint/typecheck/build and five mocked browser
state/mobile tests passed. Mobile screenshot reviewed; no overflow.

## Known limitations / outstanding work
No known failing application checks at the time of snapshot. Browser tests mock API
responses; backend integration and container smoke exercise real PostgreSQL.
Global rate cap can affect other demo visitors and resets on restart. No automated
retention job; runbook requires approved monthly cleanup of bookings ended >30 days
ago. Health does not test DB readiness. Future schema changes need new migrations.

## Production status and owner actions
Prepared backend/Dockerfile, render.yaml (autoDeployTrigger off), frontend/vercel.json
(git.deploymentEnabled false), environment examples and release runbook. No cloud
resources, remote push, Neon migration or deployment performed. Owner must approve
publication/release, authenticate accounts, create public GitHub/Neon/Render/Vercel
resources, provide secrets through provider settings, approve migration and manual
deployment, then verify one essential production booking. Disable Blueprint Auto
Sync separately if using Render Blueprints. See docs/PRODUCTION_RUNBOOK.md.

## Continuation instructions
Use Chisle only for implementation, stop explicitly for documentation. Project-local
skills include upstream Chisle with MIT license and pinned provenance, local-first,
system-analysis and release-gate. Do not globally install Chisle. Every feature
needs relevant local tests, verified flow docs, append-only log and updated handoff.
Never point local test scripts at production or infer release approval from this file.

## Final verification receipt
Implementation commit: `1da4803`. Final lockfile-based `make verify-local` passed.
Five browser tests passed separately. No remote configured. PostgreSQL remains
running locally; API smoke containers were removed and test records cleaned up.
ESLint 9 reports an upstream support deprecation warning, with no lint failures.
