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
