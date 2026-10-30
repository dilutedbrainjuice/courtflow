# CourtFlow — workflow addendum to merge into DEVELOPMENT.md and AGENTS.md

## Project and scope
- Repo: `/home/terasit/VSProjects/court`.
- Obsidian vault: `/home/terasit/VSProjects/court/buildDocs/buildDocs`.
- Docs: `buildDocs/buildDocs/CourtFlow/` and tracked in Git, excluding only `.obsidian/` personal settings and secrets.
- Preserve all existing project instructions and files; reconcile rather than overwrite.

## Required cycle for *each* logical feature/fix
1. Inspect relevant source, API and docs; state scope.
2. Coding task: invoke `$chisle` locally **only for implementation**; implement minimal complete solution with essential tests. Do not globally install Chisle for Codex.
3. Test locally: Go unit/integration/race, local PostgreSQL, frontend lint/typecheck/build as relevant. Tests skipped are not passes.
4. For docs: stop Chisle explicitly (`stop chisle`), preferably use a **new Codex documentation chat** invoking `$system-analysis`; do not load Chisle there. Analyze verified source and code flow comprehensively.
5. Update flow-specific Obsidian Markdown; update BUILD_LOG.md (append), HANDOFF.md (replace current snapshot).
6. Review `git diff --check`, `git status`, staged files and secrets; commit a logically cohesive changeset. Do not automatically deploy or push without user permission.

## Costs and environments
- Local development/testing ONLY. Production hosting and database for public recruiter demo, manual approved deploys only.
- Never use production for testing concurrency, stress, migrations experimentation, data cleanup experiments, or debugging loops.
- Local PostgreSQL container named volume; Podman is installed; no Docker required.
- Local env files `.env`, `.env.local`, `.env.*.local`, secrets excluded from Git. Only `.env.example` files committed.
- Production requires approved build, DB migration, health and one minimal happy-path smoke check, performed or explicitly authorized by owner.

## Definition of done
No feature is done without working code, relevant local checks, flow documentation, log entry, and updated handoff. No claimed deployment without actual provider verification. No public claims of benchmark numbers until measured.
