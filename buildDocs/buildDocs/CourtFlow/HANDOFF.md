# CourtFlow handoff

## Current development state — 2026-10-31
Local V1 implemented. Git branch main; public GitHub published. First manual release verified. See BUILD_LOG for
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

## Release continuation — owner authorization received
Owner approved public publication and demo setup/deployment. Origin is now
https://github.com/dilutedbrainjuice/courtflow.git. Push remains blocked by missing
HTTPS GitHub authentication. Neon/Render/Vercel MCPs report OAuth, but their tools
are not exposed in this running chat. No cloud resources, migration or deployment
performed. Continue with existing authorization after Git authentication and tool
availability are resolved; do not ask for the same release approval again.

## RELEASE-002 — 2026-10-31 MYT: provider setup in progress

- Published GitHub HEAD verified as `60c498a45525033b527e0490c8569f06e2d9dce5`.
  Existing local verification receipt remains applicable; application code unchanged.
- Direct Neon and Render MCPs work. Codex-app wrappers return Unknown tool; use
  `mcp__neon__*` and `mcp__render__*` instead.
- Neon organization `org-small-haze-66334799` reports Free. Dedicated Singapore
  PostgreSQL 17 project `aged-water-08232396` (`courtflow-demo`), production branch
  `br-lively-block-azo53da1`, database `courtflow`, role `courtflow_owner`.
  Migration `backend/migrations/001_initial.sql` applied once as a transaction to
  the verified empty database. Verified 1 facility, 4 courts, 0 bookings, exclusion
  constraint present. No production load/concurrency tests performed.
- Compute bounded to 0.25 CU. Explicit suspend-time configuration was rejected by
  Free plan; creation with provider defaults succeeded. API reports timeout 0;
  official Free-plan documentation says idle suspension is fixed at 5 minutes.
  Do not upgrade plans or add keepalive jobs; confirm idle behavior if necessary.
- Render Free Singapore service `srv-dbie50uk1f9s73ej3ln0`, name courtflow-api,
  URL https://courtflow-api-gqes.onrender.com. Docker context `backend`, Dockerfile
  `backend/Dockerfile`, branch main, autoDeploy no / autoDeployTrigger off.
  Initial manual deploy `dep-dbie51uk1f9s73ej3okg` is LIVE on approved GitHub HEAD.
  Finished 2026-10-30T18:32:38Z (2026-10-31 02:32 MYT). Essential GET /healthz
  returned HTTP 200 with {"status":"ok"}; brief deployment logs show live URL.
  DATABASE_URL transferred directly to provider env with sslmode=verify-full;
  no credentials written into repository or chat. ALLOWED_ORIGINS temporarily
  `https://courtflow.invalid`; replace with exact final frontend origin and deploy.
- Render MCP cannot set healthCheckPath; dashboard browser requires owner login.
  Owner must set `/healthz` at the service settings page. No Blueprint was created,
  so Blueprint Auto Sync is not applicable.
- Vercel list-projects works, but team `team_CIAncTjfDuRgAoEHPUGPo0xa`
  (dilutedbrainjuices-projects) requires an additional team access grant (403).
  A single-use 10-minute authorization link was presented in chat. Retry get_team
  after owner approval; generate a fresh link if expired, without reconnecting.
  Verify Hobby plan, then create CourtFlow with root frontend and manual deploy.
  Existing frontend/vercel.json disables Git deployments. No CourtFlow Vercel
  project or deployment exists yet. No end-to-end booking smoke performed yet.
- Continue using existing owner release approval. Do not recreate Neon resources,
  rerun migration unnecessarily, or deploy automatically on Git pushes.

## RELEASE-003 — 2026-10-31 MYT: first release complete

This snapshot supersedes pending deployment statements above.
- Public frontend: https://courtflow-demo.vercel.app
- API: https://courtflow-api-gqes.onrender.com
- Vercel project `prj_vprkX5t78NS3vomgUaStzLpSzDTp`, team
  `team_CIAncTjfDuRgAoEHPUGPo0xa`. Production deployment
  `dpl_omFhfNnoMuKe9UALXrMJ5C32bwDR` READY; approved commit `60c498a`.
  Root frontend, Node 22, public API URL configured for production. Build completed
  in 21 seconds; public site opened successfully. No plan upgrades purchased;
  team API summary does not expose billing-plan details, so Hobby was not separately
  verified through that response.
- Owner completed Vercel team grant and Render health path `/healthz`; verified.
  Owner additionally approved disabling Vercel Authentication for public demo access
  after automatic review initially rejected that setting. Update succeeded and
  provider reports ssoProtection disabled.
- Render ALLOWED_ORIGINS is exactly https://courtflow-demo.vercel.app. Verified
  facilities GET returned 200 with matching Access-Control-Allow-Origin header.
  Environment writes trigger API deployments even with Git auto-deploy disabled.
  An extra explicit trigger queued one redundant build; latest live deploy is
  `dep-dbie7m203lvs73akrn10`, finished 2026-10-30T18:38:06Z.
- Browser loaded all four courts/availability and confirmed exactly one demo booking.
  Reference `76af2022-7a95-4dc1-bc74-b7244847007f`: Court A, 2026-10-31
  09:00–10:00 MYT, 2500 sen, confirmed. UI showed confirmation and disabled the
  occupied slot; targeted DB read confirmed the persisted row. No second booking
  or production concurrency/load test. Screenshot inspected in browser tool.
- Vercel runtime error scan for the past hour returned none. Render remains Free,
  autoDeploy no / autoDeployTrigger off, health path /healthz. Approved source
  frontend/vercel.json has git.deploymentEnabled=false. No Blueprint exists.
- Local application code unchanged; previous local tests remain the verification
  receipt. Release documentation committed locally; deployed application commit
  stays 60c498a. Retention/cold-start limitations remain in the runbook.

## QUALITY-001 — readability review (local changes, not deployed)

API files now separate routing/middleware (api.go), catalog reads (catalog.go),
shared slot rules (slots.go), availability and booking. Same package and behavior.
Frontend state declarations and backend SQL/configuration are expanded for reading.
Required workflow now includes maintainability/code review after implementation.
Start with docs/READING_GUIDE.md for the full repository and SA walkthrough.
make verify-local and final backend race/integration tests passed. Production still
runs 60c498a; this refactor needs a separately approved manual release.

Final browser regression: all five Playwright tests passed; git diff --check passed.

## WORKFLOW-004 — comment-writing order

Finish implementation with local Chisle, explicitly stop Chisle, then re-read the
finished code and write clear, accurate explanatory comments before proceeding to
code review, final local validation and SA. Keep Chisle off throughout those steps.
If implementation fixes follow, repeat the comment/review passes. This supersedes
older wording that stops Chisle only when SA writing begins.

## COMMENTS-005 — Go comments reviewed with Chisle off

Re-read every Go source/test file against the migration and rewrote explanatory
comments. Pool ownership, deadlines, time rules, booking consistency and test
isolation are now described at their relevant code boundaries. Changes contain
only comments/gofmt alignment; comparison against HEAD and diff checks passed.
Application tests not rerun; behavior and SA call chains are unchanged.

## SA-006 — onboarding-oriented system analysis

All seven flow documents in system-analysis were rewritten against fecc983 with
Chisle off, and TEMPLATE now requires explanatory prose and precise coverage.
Read customer-booking for the user journey, then facilities/courts, availability,
booking and the health/local-database documents for runtime and local operations.
Each names real source functions and explains errors/transactions/consistency.
Coverage gaps and implementation limitations are explicit; no application tests
were rerun for this documentation-only change. Deployment state is unchanged.

## WORKFLOW-007 — portable setup artifact

Use docs/PORTABLE_DEVELOPMENT_WORKFLOW.md to establish this process on another
machine/model. It contains its own setup prompt and does not require original
chat context, provider credentials or plugins. Destination paths and approvals
must be established independently.

## AUDIT-008 — 2026-10-31 MYT: backend interview readiness

Audit only at source HEAD 77fa3a9; no application changes, Chisle, cloud access,
deployment, push or commit. Current assessment: 64/100, CONDITIONAL GO for junior
interviews; read PORTFOLIO_READINESS_AUDIT.md and PERFORMANCE_BASELINE.md before
making portfolio claims. This audit supersedes earlier “no known failing checks”
statements for stronger contention behavior, not historical release receipts.

Fresh checks: Go unit/real-Postgres integration/race/vet/build, frontend lint,
typecheck and host-access production build passed. Five mocked browser tests passed
with WATCHPACK_POLLING=true (default watcher hit EMFILE). Initial sandbox Next build
failed TypeScript config parsing; host retry passed. No dependency installation.

Two independent local API processes, 100 simultaneous reservations × 100 rounds:
73 rounds exactly 1 created/99 conflicts/1 stored row. Total 75 created, 7,228
expected conflicts, 2,697 unexpected 503s. At most one persisted row in every
round; 14 all-503 rounds nevertheless committed a booking, 11 stored none.
PostgreSQL logs show cancellation during exclusion checks. Root cause unresolved.
Direct SQL partial-overlap rejection, adjacency and transaction rollback passed.
Read endpoint benchmarks and raw measurements are preserved in audit-evidence/.

Other findings: main does not join HTTP Shutdown before pool close/exit; shared
anonymous limiter permits starvation (120 invalid → next legitimate 429). No
critical unauthorized/private-data path demonstrated in current public demo.
Smallest next work: diagnose contention/ambiguous responses, fix shutdown lifecycle,
add error/pool diagnostics, rerun deterministic persisted-state multi-process gate.
Idempotency is the next valuable feature, not implemented by this audit.

Local audit schemas removed and API processes stopped; public bookings remained 0.
No fresh restart-durability/image/provider/security-scanner checks. Application
source retained unchanged, including restoration of generated next-env.d.ts.

## DESIGN-009 — reliability remediation proposal (2026-10-31 MYT)

User requested brainstorming, SA first, then review; no implementation. Added
system-analysis/reliability-remediation-proposal.md and RELIABILITY_REMEDIATION_PLAN.md.
Author review complete. Preferred conditional candidate: database court-row lock,
separate READ COMMITTED occupancy query, insert/commit with exclusion retained.
Diagnostic reproduction is mandatory before selecting that fix; root cause unresolved.
Shutdown requires joined drain, bounded cancellation and ordered pool cleanup.
No code changes or fix tests. Wait for a subsequent implementation request; use the
proposal's acceptance gates and preserve original audit measurements.

## RELIABILITY-010 — 2026-11-01 MYT: local remediation complete

Implemented DESIGN-009 after diagnostic reproduction showed 40P01 insert deadlocks,
transactionid waits and exhausted request budgets. Booking now owns a READ COMMITTED
court-row transaction with separate overlap snapshot and explicit commit. Constraint
unchanged. Cleanup completes before response writes. Added request IDs, sanitized
error phase/SQLSTATE/pool counters and joined drain/cancellation before pool closure.
No idempotency/fairness/authentication expansion or migration.

Final PASS: uncached Go race/unit/integration, vet/build; frontend lint/typecheck/build
and all five mocked browser tests. Both 100-round × 100-contender process gates passed:
20,000 total attempts, 200 created, 19,800 conflicts, zero unexpected outcomes, every
round persisted exactly one matching UUID/2500-sen row. Real SIGTERM during blocked
booking drained to 201/one row/exit 0. Local read baseline rerun and evidence preserved.
See RELIABILITY_RESULTS.md and current booking/health SA. Old audit score is historical;
no new score assigned. Unknown commit/response outcomes and shared limiter remain.

Chisle stopped; explanatory comments and code review completed before final checks.
No cloud, push or deployment. Local processes/schemas cleaned up. Next action: owner
can review local changes; production release requires its own explicit authorization.
