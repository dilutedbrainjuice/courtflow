# CourtFlow — workflow addendum to merge into DEVELOPMENT.md and AGENTS.md

## Project and scope
- Repo: `/home/terasit/VSProjects/court`.
- Obsidian vault: `/home/terasit/VSProjects/court/buildDocs/buildDocs`.
- Docs: `buildDocs/buildDocs/CourtFlow/` and tracked in Git, excluding only `.obsidian/` personal settings and secrets.
- Preserve all existing project instructions and files; reconcile rather than overwrite.

## Required cycle for *each* logical feature/fix
1. Inspect relevant source, API and docs; state scope.
2. Invoke `$chisle` locally **only for implementation**; finish the minimal complete solution and essential tests. Do not globally install Chisle.
3. Explicitly `stop chisle`. Re-read the completed code and callers, then write or revise explanatory comments in normal, clear prose. Check intent, contracts, invariants and non-obvious decisions against the actual code; remove stale or misleading comments. Finish this pass before the next step.
4. With Chisle off, perform maintainability review and code review of the complete changed flow.
5. Test locally: Go unit/integration/race, local PostgreSQL, frontend lint/typecheck/build as relevant. Fix failures; if implementation changes, repeat the stop-Chisle/comment/review passes before final validation. Tests skipped are not passes.
6. With Chisle still off, use `$system-analysis` to analyze verified source and update flow-specific Obsidian Markdown. A separate documentation chat is optional.
7. Append BUILD_LOG.md and update HANDOFF.md.
8. Review `git diff --check`, `git status`, staged files and secrets; commit a cohesive changeset. Do not automatically deploy or push without user permission.

## Costs and environments
- Local development/testing ONLY. Production hosting and database for public recruiter demo, manual approved deploys only.
- Never use production for testing concurrency, stress, migrations experimentation, data cleanup experiments, or debugging loops.
- Local PostgreSQL container named volume; Podman is installed; no Docker required.
- Local env files `.env`, `.env.local`, `.env.*.local`, secrets excluded from Git. Only `.env.example` files committed.
- Production requires approved build, DB migration, health and one minimal happy-path smoke check, performed or explicitly authorized by owner.

## Definition of done
No feature is done without working code, relevant local checks, flow documentation, log entry, and updated handoff. No claimed deployment without actual provider verification. No public claims of benchmark numbers until measured.

## Required maintainability and code review gate

After implementation and the completed Chisle-off comment pass, before final
validation and documentation, review the complete
changed flow and its callers/tests for readability, correctness and maintainability.
This gate applies to Go, TypeScript/React, SQL, scripts and configuration.

- Prefer clear names, explicit control flow, small cohesive functions and files grouped
  by responsibility. Do not compress unrelated declarations or hide errors to save lines.
- Go: use gofmt, standard-library conventions, explicit errors and composition. Apply
  SOLID's responsibility/dependency principles pragmatically; introduce small interfaces
  at actual consumer boundaries only when substitution is needed. Do not manufacture
  service/repository layers, inheritance analogues or one-method interfaces for ceremony.
- Comments explain why, invariants, concurrency, units, non-obvious decisions and public
  contracts. Avoid narrating obvious statements. Keep SQL readable and parameterized.
- React: keep state transitions understandable; review effect dependencies/cancellation,
  loading/error states, accessibility and API boundaries. Extract components/helpers when
  they have a distinct responsibility, not solely to reduce file length.
- Review behavior compatibility, validation, error paths, races, security, resource
  cleanup and test coverage. Formatting and passing tests do not replace code review.
- Fix findings, rerun affected local checks, and update SA call chains/source references.
  Record review scope, results and remaining limitations in the build log/handoff.
- Chisle never overrides readability, comments, this review gate or complete documentation.
