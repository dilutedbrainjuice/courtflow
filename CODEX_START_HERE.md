# CourtFlow — Codex continuation brief

**Purpose:** Continue the existing CourtFlow repository, not a fresh project. This document is a task specification to hand to Codex; inspect local files before changes.

## Immutable local configuration

- Repository root: `/home/terasit/VSProjects/court`
- Obsidian vault root (contains `.obsidian`): `/home/terasit/VSProjects/court/buildDocs/buildDocs`
- Project documents: `buildDocs/buildDocs/CourtFlow/`
- Git: `main` initialized locally; no remote, no commit yet per last reported status.
- Installed: Go 1.26, Node 22, npm 10, Git 2.56, Podman. No local psql or Docker required.
- Cost budget: RM0/month; use only free tiers.
- Product: CourtFlow, a **badminton court booking and facility operations app**.

## Direct instruction to Codex

You are working in the existing repository at the exact path above. **First inspect** `AGENTS.md`, `DEVELOPMENT.md`, `Makefile`, the existing README, docs, build log, handoff, `.gitignore`, and the repository status. Preserve all existing user content, especially the Obsidian vault. Merge or append this plan; do not blindly replace existing files. If this handoff pack is extracted into the repo, inspect and integrate its contents with existing files.

**Working style:** implement code leanly using the project-local official Chisle *skill* only in coding tasks, never install its global Codex rules. After implementation, explicitly stop Chisle and complete a clear explanatory comment pass before code review, final validation and documentation. Keep Chisle off for these steps. For system-analysis writing, stop Chisle explicitly and ideally use a separate documentation conversation with the `system-analysis` skill. Documentation cannot be sacrificed for token efficiency. See `WORKFLOW_ADDENDUM.md` and `docs/TOOLS_AND_MCPS.md`.

**Local-first:** develop, run PostgreSQL, test backend, test frontend, run integration/concurrency tests locally. No Neon queries, production API calls, Render/Vercel deployments, or remote schema changes while implementing. No real payment gateway. Never auto-deploy on Git pushes. External account setup and releases need the owner's authorization.

**Goal:** Finish a locally working, demo-ready V1 in this existing repo and document it thoroughly. Stop before actions requiring user account credentials/authorization (OAuth, GitHub remote creation, hosting resource creation, paid plans, first production deployment). Report what to approve and exact manual steps.

## Production stack — plan, not deployed yet

- Frontend: Next.js + TypeScript on **Vercel Hobby**, project root `frontend/`.
- Backend: single Go HTTP API on **Render Free** (native Go server built with Docker), `backend/`, listen on `$PORT` at `0.0.0.0`.
- DB: **Neon Free PostgreSQL** using `DATABASE_URL` with TLS and a bounded pgx connection pool.
- Version control: public GitHub monorepo.
- All secrets in host environment variables (never git).
- Avoid Redis, Kafka, Kubernetes, WebSockets, paid SMS/email and real payment processing in V1.
- Render `Auto-Deploy = Off`; Vercel `git.deploymentEnabled = false` in `frontend/vercel.json`. Manual production deployments only with explicit owner approval.
- Render cold starts after idle; Neon can scale to zero. Document degraded first-request experience.
- Vercel Hobby is **personal/noncommercial**; production here means a public portfolio demo, not commercial facility operations.

## V1 product acceptance criteria

- Fictional facility `CourtFlow Demo Badminton` with four courts, display in `Asia/Kuala_Lumpur` time zone.
- Public page showing courts, availability, one-hour booking slots, and prices (example RM25/hour, adjustable via seed data).
- Make a guest **demo booking** with no actual money. No real names, emails, phone numbers, payment information or external integrations needed. Show success or conflict clearly.
- HTTP API: `GET /healthz`, `GET /api/v1/facilities`, `GET /api/v1/facilities/{id}/courts`, `GET /api/v1/courts/{id}/availability?date=YYYY-MM-DD`, `POST /api/v1/bookings`; add other endpoints only when justified.
- Overlapping confirmed bookings for the same court **must never exist**. Enforce this in PostgreSQL with a range/exclusion constraint, in addition to application validation. Map contention to HTTP 409.
- Run concurrent local tests: 20+ simultaneous requests for the same court/slot must produce exactly one confirmation and no duplicates. Repeat deterministically using isolated test DB transactions/schema, not production.
- Bookings must be durable across backend restarts. Amount and slot availability calculated on the server, not trusted from browser.
- Simple read-only operations view displaying current demo bookings is fine. Never expose destructive admin mutation without authentication.
- Practical UX: responsive layout, clear loading/error states, no unnecessary real-time implementation (a refresh is acceptable for V1).
- Demo safeguards before public launch: bounded booking window, request validation and timeouts, basic abuse control, no real user PII, demo data cleaning strategy and sensible CORS.

## Technical decisions

- Modular Go monolith with `net/http` / Go standard library or minimal Chi router, `pgx`, PostgreSQL. Avoid microservices.
- SQL migrations versioned in repo, no manual production-only schema modifications. Local PostgreSQL first.
- Model times as `timestamptz` UTC instants; convert/format in the venue time zone for UI. Define `[start,end)` half-open intervals so adjacent slots don't conflict.
- For V1 use immediate **confirmed** simulated bookings; temporary payment holds/expiry are future work. Never advertise five-minute holds until implemented and tested.
- Use an appropriate PostgreSQL GiST exclusion constraint over `(court_id, tstzrange(start_at,end_at,'[)'))` restricted to confirmed bookings, with `btree_gist`. Handle exclusion violation (SQLSTATE `23P01`) as a booking conflict. Concurrent Go handlers alone do not provide correctness.
- For public demo, do not expose private booking details by guessable IDs; omit sensitive data and secure booking-management endpoints before adding them.
- APIs should be documented with clear request/response/error schema and `curl` examples. Health endpoint should be low-cost and not disclose secrets.

## Local setup & tests

1. Check `podman info`. Use a named-volume Postgres 17 container bound only to localhost. The included `scripts/dev-db.sh` is a candidate to review/merge, not something to overwrite existing scripts blindly. No local `psql` installation required: `podman exec courtflow-postgres psql ...`.
2. Generate / adjust `backend/go.mod`, `cmd/server`, Go packages, SQL migrations, tests, `.env.example`; `frontend/` Next.js TypeScript app with npm lockfile, `.env.example`.
3. Update Makefile with `dev-db`, `dev-api`, `dev-web`, `test`, `build`, `check`, `verify-local` targets. Commands must be repeatable and use local-only services by default.
4. Implement migration tests and concurrent booking integration tests before treating API as release-ready.
5. Test: `go test ./...`, `go test -race ./...`, `go vet ./...`, build API, npm lint/typecheck/build, and HTTP happy/conflict path tests. Avoid making lint output giant; display failures and concise summaries.
6. Track any incomplete tests explicitly — don't report 'passed' when skipped.

## Documentation invariants

For each API/business flow, write one Markdown file in `buildDocs/buildDocs/CourtFlow/system-analysis/` following the existing `TEMPLATE.md`. Include exact call chain from route through handlers/services/repositories/DB and back, input/output validation, Go functions with paths, transaction boundaries, SQL, errors, race safety, Mermaid sequence diagram, and tests. For significant helpers, document their role in the parent flow. **Do not invent behavior not verified in code.**

Append a dated section to `BUILD_LOG.md` for each logical changeset; update `HANDOFF.md` to match actual code. Do not overwrite the log or replace the handoff with generic commentary. Include exact local commands and their results. Prefer concise implementation chat, extensive SA docs. Add useful architecture decision records when choices matter.

## Work stages — continue autonomously locally

**A. Reconcile workspace:** inspect files, Git tree and existing docs; integrate workflow addendum; install Chisle as local skill (copy upstream skill + MIT license, not global `npx chisle` installer); add `system-analysis`, `local-first` and `release-gate` project skills, and local DB script if needed. Verify `AGENTS.md` and existing permanent instructions consistent. Record in log.

**B. Local infrastructure:** make Podman PostgreSQL repeatable, minimal dependencies, migrations and seeded fictional facility/courts. Validate container restarts preserve data and local SQL works. Record in log/SA.

**C. Go MVP:** API, storage, tests, concurrency invariants, graceful shutdown, health check, structured basic logs (never secrets), request timeouts, stable request validation. Document each flow.

**D. Next.js MVP:** pages for facility, availability and demo booking, simple management read-only; local Go API integration, clear error states, mobile responsiveness. Document flows. Use Chisle only for coding task; documentation separately.

**E. Quality and build:** full local acceptance checks, race test, concurrency test, restart persistence test, clean Git diff, README and local quickstart. Document results. Create initial local commits in logical sets only after checking what is staged; do not automatically publish GitHub repo.

**F. Production readiness (do not deploy without approval):** create Render Dockerfile + optional render.yaml with autoDeployTrigger `off`, Vercel manual-deployment config, `.env.example`, release runbook, migration plan, minimal smoke checklist, and ensure no secrets are tracked. Local build must pass in deployment-like config. Then stop and report precisely which account steps and manual approvals are needed.

A final summary should clearly separate **implemented**, **verified locally**, **prepared but not deployed**, and **blocked on owner authorization**.
