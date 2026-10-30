# Architecture

Status: planned; application implementation has not started.

Next.js frontend → Go HTTP API → PostgreSQL.

Develop all three locally, using npm for frontend dependencies and simulated
payments. Podman is available for a future local PostgreSQL container. Database
schema, API contracts, authentication and deployment configuration remain undecided.

Keep backend and frontend in separate responsibility-based directories when they
are introduced. Do not create empty application scaffolds prematurely.

Apply Chisle rules only to application implementation, never to system-analysis
documentation. No Chisle ruleset has been supplied or installed.

## Implemented V1 — 2026-10-31

Go net/http monolith in backend/cmd/server and backend/internal/api; Next.js App
Router in frontend/app; PostgreSQL 17 in a rootless Podman named volume. The pool
is bounded to eight connections. API requests have five-second database contexts,
transport deadlines and bounded JSON bodies. Server shutdown allows ten seconds.

The confirmed-booking GiST exclusion constraint is the correctness boundary.
Application checks improve feedback, but do not serialize booking attempts.
Prices are read in the insert statement. All instants use timestamptz and display
in Asia/Kuala_Lumpur. No user identity or real payment data exists.

Current initial migration is repeatable and transactionally applied locally.
New deployed schema changes must use subsequent numbered files; never silently
rewrite a migration already applied in production. Health is liveness only.

The frontend is statically rendered then reads the API in the browser. No proxy,
SSR data credentials, realtime socket or privileged operations API is present.
Manual deploy configuration exists, with no cloud resources created.
