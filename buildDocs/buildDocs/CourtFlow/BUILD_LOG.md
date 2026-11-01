# CourtFlow build log

## 2026-10-31 — SETUP-001

- Objective: establish safe local repository and documentation workflow.
- Added: `.gitignore`, `.editorconfig`, `DEVELOPMENT.md`, `Makefile`, `README.md`,
  `scripts/check-env.sh`, and under `buildDocs/buildDocs/CourtFlow/`:
  `architecture/OVERVIEW.md`, `system-analysis/TEMPLATE.md`, `HANDOFF.md`, `BUILD_LOG.md`.
- Modified: `AGENTS.md` to reference permanent development instructions.
- Implementation: initialized Git on main; recorded permanent instructions,
  local environment checks, architecture and comprehensive flow-document template.
- Decisions: preserved all existing vault content; ignored personal Obsidian settings;
  deferred application scaffolding, tool installation and database provisioning.
- Git inspection: sandbox showed empty read-only metadata; actual host inspection
  found no `.git`, hence no history/configuration required backup.
- Validation: `make check` passed (shell syntax and tracked diff whitespace);
  `make check-env` passed; `git status` confirmed unborn main;
  `git check-ignore` confirmed personal Obsidian settings excluded.
  New files were untracked during the diff check. No application tests exist.
- Known issues: PostgreSQL not configured; Podman version verified but runtime untested.
- Commit: none created.
- Next: define first implementation scope and establish application toolchains and local database.

## 2026-10-31 — V1-001: local application and infrastructure

- Objective: implement the local CourtFlow V1 without cloud resources.
- Files: new backend/ Go server, API, SQL migration, go.mod/go.sum, tests,
  Dockerfile and environment example; frontend/ Next.js app, npm lockfile,
  lint/TypeScript/Playwright configuration, browser tests and Vercel configuration;
  render.yaml; migration, persistence and container-smoke scripts. Updated Makefile,
  local database and verification scripts, README, .gitignore, AGENTS and DEVELOPMENT.
  Installed upstream Chisle skill/LICENSE/commit marker under .agents only; preserved
  existing local-first, system-analysis and release-gate skills and user handoff pack.
- Implementation: fictional four-court venue; five public APIs; guest demo booking;
  server pricing; hourly MYT availability; PostgreSQL confirmed-range exclusion;
  bounded pool/body/timeouts; global booking attempt cap; exact CORS allowlist.
- Decisions: database enforces race safety, immediate simulated confirmation,
  no PII/payment integrations, no privileged admin API. Optional reception list
  deferred; public availability exposes occupancy only. Manual cloud release only.
- Validation: migration and repeat application passed; Go unit/integration/race/vet
  and build passed. 24 concurrent attempts yielded one 201 and 23 HTTP 409, one row.
  Adjacent/different-court and invalid request checks passed. API/Postgres restart
  persistence passed. Podman production image build, health and booking smoke passed.
  Frontend lint/typecheck/build passed; five Playwright tests passed (success,
  conflict, invalid, network, mobile layout). Mobile screenshot visually reviewed.
- Corrections during checks: concurrent npm operations disrupted generated
  node_modules; clean sequential installation repaired it. Fixed Next.js link/date
  lint findings and ambiguous Playwright selectors, then reran relevant checks.
- Documentation: Chisle explicitly stopped for system-analysis writing. Added health,
  facilities, courts, availability, booking, customer-booking and local-database
  flow documents; expanded architecture and manual production runbook.
- Known limits: no deployment, no production credentials, no automatic retention
  job; manual 30-day retention procedure documented. Global rate cap resets with
  process. Health is liveness only. Browser-state tests mock API responses.
- Commit: recorded in local Git history after this entry is staged.
- Next: finish final lockfile verification and commit review; owner approves public
  GitHub publication and provider/account setup before any production action.

## 2026-10-31 — V1-002: verification receipt

- Objective: record the completed local verification and implementation commit.
- Files modified: BUILD_LOG.md and HANDOFF.md only.
- Result: final `make verify-local` passed from npm lockfile, including Go vet,
  unit/integration/race checks, Go build, frontend lint, typecheck and production
  build. Five Playwright browser tests passed separately. Staged whitespace check
  passed; reviewed tracked files exclude dependencies, secrets and personal Obsidian
  state. Targeted credential-pattern search found no production credentials.
- Implementation commit: `1da4803` (Implement local CourtFlow booking demo and verification workflow).
- Known issues: ESLint 9 emits an upstream support deprecation warning; lint itself
  passes. Browser-state tests use mocks; real database behavior is covered by Go
  integration and production-image HTTP smoke tests. No cloud verification performed.
- Next action: owner review and explicit publication/release authorization, then
  account setup and minimal manual production smoke checks per runbook.

## 2026-10-31 — RELEASE-001: approved publication, access checks

- Owner approved GitHub publication and demo production setup/deployment, then
  supplied https://github.com/dilutedbrainjuice/courtflow.git.
- Added that repository as local origin; remote inspection showed no main branch.
- Neon, Render and Vercel configured MCP servers report OAuth authentication.
  Provider tools are not exposed to this running chat; no provider resources changed.
- Push did not complete: HTTPS Git credentials are unavailable (noninteractive
  check failed reading GitHub username). No remote branch or deployment claimed.
- Next: owner authenticates Git push locally and pushes main; reopen/reload Codex
  to expose newly authenticated provider tools, then continue the approved release.

## 2026-10-31 — RELEASE-002: approved free-tier provider provisioning

- Verified public GitHub HEAD 60c498a and direct Neon/Render MCP access. Vercel
  list-projects works; target team grant is pending (403, consent link supplied).
- Created dedicated Neon Free Singapore PostgreSQL 17 CourtFlow project and applied
  versioned 001 migration transaction once after confirming an empty schema.
  Verified facility/court/booking counts 1/4/0 and exclusion constraint presence.
- Created Render Free Singapore Docker service with auto-deploy explicitly off.
  Initial manual deployment builds 60c498a; database secret transferred directly
  to Render env with verified TLS. Temporary CORS origin blocks browser access
  until Vercel origin is available. No production booking or stress tests run.
- Files: updated HANDOFF and production runbook; appended this log. No application
  code changed. Prior local verification receipt retained; git diff check required.
- Pending: Render build/health verification, owner dashboard health path setting,
  Vercel team consent/Hobby verification/frontend deployment, exact CORS origin,
  one controlled browser booking smoke. See HANDOFF for stable provider IDs.

- Deployment result: Render manual deploy LIVE, finished 2026-10-30T18:32:38Z;
  single essential GET /healthz returned HTTP 200 and {"status":"ok"}. Brief logs
  confirmed live service. Frontend and end-to-end booking remain pending owner
  Vercel grant. git diff --check passed. No code changes or local test reruns.

## 2026-10-31 — RELEASE-003: frontend deployment and essential booking smoke

- Verified owner-completed Render health path and Vercel team authorization.
- Created CourtFlow Vercel project rooted at frontend, Node 22, production API URL;
  manually deployed 60c498a. Deployment READY; build completed in 21 seconds.
- Owner explicitly approved public access after automatic review blocked removal
  of Vercel Authentication. Public setting applied successfully.
- Set exact CORS origin https://courtflow-demo.vercel.app, verified HTTP 200 and
  matching CORS header. Render env updates trigger deployment; an extra manual API
  trigger queued a redundant build. Final live deploy dep-dbie7m203lvs73akrn10.
- One browser booking confirmed, reference 76af2022-7a95-4dc1-bc74-b7244847007f.
  Court A, 31 October 2026 09:00–10:00 MYT, RM25 simulated. UI confirmation and
  unavailable slot observed; targeted database read confirmed persistence.
- Vercel runtime error scan returned none. Render auto-deploy off and /healthz
  verified. Source Vercel git.deploymentEnabled false verified. No load tests,
  extra bookings, paid upgrades or code changes. Vercel billing plan not exposed by
  team summary; no claim of independent Hobby billing verification.
- Updated HANDOFF/runbook/log; previous local verification receipt still applies.
  Next: owner can share public demo; follow manual release and retention policy.

## 2026-10-31 — QUALITY-001: readability and review workflow

- Reviewed backend startup/routes/handlers/SQL/tests, frontend page/layout/styles/
  browser tests, local scripts and deployment configuration for maintainability.
- Split API code by responsibility into catalog.go, slots.go, availability.go and
  booking.go; retained one api package and explicit pgx dependency. Expanded SQL,
  server settings, state declarations and validation predicates; documented pricing,
  overlap, limiter and timezone invariants. HTTP contracts remain unchanged.
- Added mandatory post-implementation maintainability/code review gate to DEVELOPMENT
  and WORKFLOW_ADDENDUM. Chisle stopped before explanatory documentation.
- Updated four SA source maps, corrected stale README setup claims and wrote
  docs/READING_GUIDE.md covering every runtime/support component and SA reading order.
- Verification: make verify-local passed (format, vet, unit/integration/race, Go
  compile, npm ci/lint/typecheck/build). After SQL formatting, make test passed again.
  Initial sandbox run could not write Go cache; authorized local rerun passed.
- Review limitations: this is a focused readability refactor, not a new layered
  architecture. Response maps remain dynamic and frontend request JSON is unvalidated
  at runtime; typed contracts can be a separate improvement as the API expands.
- No production resources changed and no new deployment performed.

Final browser regression: all five Playwright tests passed; git diff --check passed.

## 2026-10-31 — WORKFLOW-004: comment pass after stopping Chisle

- Owner requested finishing implementation before disabling Chisle and writing clear
  comments, with that pass completed before proceeding.
- Aligned DEVELOPMENT, WORKFLOW_ADDENDUM, AGENTS, CODEX_START_HERE and reading guide:
  implementation → stop Chisle → comments → review → tests → SA → log/handoff → commit.
- Comments must be checked against completed code. Implementation fixes repeat the
  comment/review passes before final validation. Chisle remains off through review
  and documentation. No application behavior changed; application tests not rerun.
- Validation: reviewed instruction order and git diff --check.

## 2026-10-31 — COMMENTS-005: Go explanatory comment pass

- Re-read all seven Go source/test files and the SQL migration with Chisle off.
- Rewrote comments across startup, routing, catalog, slot rules, availability,
  booking and tests. Explained pool ownership/lazy connectivity, exact origins,
  deadline scope, venue-date versus instant validation, half-open overlap,
  server pricing, constraint conflicts and committed-but-lost response behavior.
- Corrected API ownership wording: main creates/closes the pool; API holds it.
  Test comments explain schema isolation, migration repeats and concurrency intent.
- Validation: gofmt, git diff --check, and comparison against HEAD with comments
  and whitespace removed confirmed no executable-code changes in all seven files.
  Application tests were not rerun for this comments-only change; prior behavior
  verification remains in QUALITY-001. No cloud actions or deployment.

## 2026-10-31 — SA-006: comprehensive explanatory SA rewrite

- Applied system-analysis skill with Chisle off; rechecked Go, React, SQL, local
  database scripts, smoke scripts and test assertions against source fecc983.
- Rewrote all seven flow documents and expanded TEMPLATE guidance. Documents now
  explain purpose, contracts, actual numbered calls, significant helpers, rules,
  SQL/transactions, failure recovery, concurrency, source references and diagrams.
- Replaced generic copied claims with endpoint-specific coverage and explicit gaps.
  Clarified lazy pool readiness, shutdown drain limitation, empty catalog behavior,
  date-versus-instant horizon, invalid timezone differences and lost-response risk.
  Frontend cancellation claims now distinguish abort wiring from response guards.
- Validation: reviewed source references, Markdown fences, section coverage and
  git diff --check. Documentation only; application tests not rerun. No deployment.

## 2026-10-31 — WORKFLOW-007: portable workflow setup guide

- Created docs/PORTABLE_DEVELOPMENT_WORKFLOW.md as a self-contained transfer file
  for another machine/model/repository, including setup and continuation prompts.
- Preserved the required implementation → stop Chisle → comments → review → local
  checks → SA → log/handoff → commit sequence and manual-release authorization.
- Added destination discovery, instruction merging, exact Chisle directory/license
  transfer option and a clearly labeled no-plugin implementation-mode fallback.
- Validation: reviewed against current workflow, Markdown fences and whitespace.
  Documentation only; application tests not rerun. No deployment or remote push.

## 2026-10-31 — AUDIT-008: backend interview readiness and local baseline

- Objective: critical evidence-based Go portfolio audit only, source HEAD 77fa3a9.
- Added PORTFOLIO_READINESS_AUDIT.md, PERFORMANCE_BASELINE.md and audit-evidence/
  local-audit.py/local-audit-results.json; appended current findings to HANDOFF.md.
  No application source edits, feature changes, Chisle, production access or commit.
- Reviewed Go/React/SQL/tests/scripts, Git history, architecture and release config.
  Scored 64/100, CONDITIONAL GO for junior interviews. Kept historical production
  receipts separate from current audit verification; no provider access.
- Fresh PASS: Go TestSlotValidation/TestIntegration with loopback PostgreSQL and
  unique test schema, uncached race, vet, compilation; frontend lint/typecheck and
  host-access production build; five mocked Playwright tests (polling, 10.3s).
  Initial sandbox Next build failed parsing TS config; host retry passed. Default
  browser watcher hit EMFILE; stopped audit processes and polling retry passed.
- FAIL: strict 100-attempt contention contract over 100 two-process rounds. 73
  rounds met 1 created/99 conflicts/1 row; 75 total created, 7,228 expected 409s,
  2,697 unexpected 503s. Never >1 stored row. 14 all-503 rounds committed one row,
  11 committed none. Database logs show canceled exclusion-constraint checks;
  root cause not claimed. DB integrity and API reliability distinguished.
- PASS: direct SQL partial overlap rejection, adjacent insert and explicit rollback
  with persisted count; limiter starvation probe 120 invalid then legitimate 429.
  Measured GET facilities/availability at concurrency 1/10/50/100 for ~3 seconds,
  RPS/mean/p50/p95/p99, API CPU/RSS and sampled DB sessions. Raw evidence and exact
  reproduction commands retained; no production extrapolation or invented metrics.
- Review findings: shutdown not joined, shared anonymous cap, no idempotency or
  identity/ownership surface, weak runtime error/pool visibility. No demonstrated
  double-booking/private-data authorization defect; bounded scope clearly stated.
- NOT VERIFIED fresh: restart durability, image build/smoke, provider settings,
  real-backend browser journey, backup restore, vulnerability scans, pool wait stats,
  contention CPU/RSS, sustained successful-write throughput and in-flight shutdown.
- Cleanup: removed audit schemas/stopped API processes; existing local Postgres
  remains running; public bookings count stayed 0. Generated next-env restored.
- Validation: reviewed report against raw results and source; git diff --check.
- Commit hash: none (audit documentation left for owner review).
- Next action: separately implement/verify contention reliability and joined
  shutdown with diagnostic telemetry; then repeat the isolated multi-process gate.

## 2026-10-31 — DESIGN-009: reliability brainstorm, proposed SA and review

- Objective: plan remediation after AUDIT-008 without implementing first.
- Added system-analysis/reliability-remediation-proposal.md and
  RELIABILITY_REMEDIATION_PLAN.md; updated HANDOFF with design-only continuation.
- Compared timeout/pool tuning, local mutex, selective retries, advisory locks,
  court-row transactions and idempotency. Chose diagnosis first and conditional
  court-row serialization; retained exclusion constraint and unknown-commit caveat.
- Reviewed source contracts, transaction snapshot boundaries, cross-process locking,
  limiter effects and shutdown ownership. Author review; no independent review claimed.
- Validation: documentation/source review and git diff --check only. No application
  changes, new tests executed, Chisle, production access, commit or deployment.
- Open issue: exact contention/commit-acknowledgment failure cause is unproven.
- Next action: user reviews plan; implement only on a subsequent request.

## 2026-11-01 — RELIABILITY-010: booking and shutdown reliability

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

Files: backend booking/routing/main/tests; new diagnostics and lifecycle helpers/tests;
scripts/concurrency-local.py; current booking/health SA, architecture, results/evidence,
audit/proposal follow-up pointers, README, HANDOFF and this append-only log. Existing
AUDIT-008/DESIGN-009 artifacts preserved. Initial test-helper extraction had a compile
error, corrected before final validation; SIGTERM probe filename initially shadowed
Python signal, renamed and passed. No omitted run is counted as a pass.

## RELEASE-011 — 2026-11-01 MYT: reliability release live

Owner explicitly requested push/deploy and confirmed existing Render My Workspace.
Pushed a08960ae86d979dcdd74c74086c70403f870aac6 to origin/main. Prior local
race/integration/vet/build/frontend/browser/20,000-attempt receipts apply to that
unchanged commit. No schema changes or production migration.

Vercel production dpl_14AMKqG6Kf1f2dShxPjNrDMgrqyD READY on a08960a;
https://courtflow-demo.vercel.app returns HTTP 200. Render manual deployment
dep-dbjlai2jnfac738bavc0 LIVE on the same full SHA, finished 2026-11-01
15:06:34 UTC (23:06 MYT). API https://courtflow-api-gqes.onrender.com.
Render remains Free, autoDeploy no / trigger off, health path /healthz. Source
frontend/vercel.json continues disabling Git-triggered deployments.

Minimal production API smoke: health 200; exact frontend Origin CORS and generated
request IDs verified. One POST returned 201 confirmed, Court D, 2026-11-02
08:00–09:00 MYT, 2500 sen. Booking reference 52e51912-99f9-4ece-9221-fff7863a7f34.
Subsequent availability read marked that exact interval unavailable. No second
booking, stress test, DB experiment or production cleanup. This was an HTTP/API
integration smoke, not an automated browser-to-API journey. Public frontend HTML
also returned 200. Free-tier cold starts/retention and idempotency limitations remain.

Release receipt is documentation-only and may be committed/pushed separately;
deployed application SHA stays a08960a. No automatic redeployment required.
