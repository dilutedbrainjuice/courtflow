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
