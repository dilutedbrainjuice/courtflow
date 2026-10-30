# CourtFlow — Permanent Development Instructions

We are building CourtFlow, a sports facility booking and management platform.

## 1. Project paths

- Project root: `/home/terasit/VSProjects/court`
- Obsidian vault: `/home/terasit/VSProjects/court/buildDocs/buildDocs`
- Do not move, rename, recreate, or relocate these directories.
- Treat the current project directory as the repository root.

## 2. Git initialization

Git has not been initialized successfully yet.

1. Inspect the existing `.git` directory. A previous check reported it as invalid.
2. Determine whether it contains recoverable history or configuration.
3. Preserve any existing Git data. If invalid, back it up outside the repository before creating a fresh Git repository.
4. Initialize Git with `main` as the default branch.
5. Create a suitable `.gitignore`, including secrets, dependencies, build outputs and personal Obsidian settings.
6. Do not create the remote GitHub repository until the local project is ready.

The eventual GitHub repository must be public.

## 3. Local-first development policy — MANDATORY

All development and testing must happen locally first.

- Frontend: Next.js development server.
- Backend: local Go HTTP server.
- Database: local PostgreSQL.
- Use npm for frontend dependency management.
- Check whether Podman is installed; if not, recommend a suitable local PostgreSQL setup.
- Use mock or simulated payments.
- Never use the production database for development tests.
- Never use production as the primary debugging environment.
- Avoid unnecessary calls to Vercel, Render and Neon.
- Do not automatically deploy on every commit.
- Production deployment requires successful local tests and explicit user approval.

After deployment, run only essential smoke tests against a dedicated demo/test account.

## 4. Coding and documentation workflow

Use Chisle rules ONLY during application code implementation.

Disable Chisle-style constraints during system-analysis documentation.

For every new or modified API/controller/business flow:

1. Inspect the existing architecture and relevant code.
2. Implement the smallest appropriate change.
3. Run relevant local tests.
4. Fix failing tests.
5. Write or update its corresponding system-analysis document.
6. Update the append-only build log.
7. Update the model handoff file.
8. Verify the Git diff before committing.

Do not install a globally always-on Chisle ruleset.

## 5. Obsidian system analysis

Create and maintain documentation directly under:

`buildDocs/buildDocs/CourtFlow/`

Use this structure:

- `system-analysis/` — one Markdown document per API/business flow
- `architecture/` — system architecture and design decisions
- `BUILD_LOG.md` — chronological, append-only change history
- `HANDOFF.md` — current project status and continuation instructions

Each system-analysis document must explain:

- Business purpose
- API endpoint or triggering event
- Request and response
- Complete ordered function call flow
- Functions involved and their responsibilities
- Parameters and return values
- Business rules and validation
- Database queries and transactions
- Error handling and edge cases
- Concurrency and consistency considerations
- Mermaid sequence diagram
- Relevant source-file references
- Tests that validate the flow

Document all significant functions involved in each flow. Do not create redundant documentation for trivial helpers.

The documentation must be understandable to someone learning the codebase, not just an AI agent.

## 6. Build log and handoff

For every completed logical changeset, append to `BUILD_LOG.md`:

- Date and change identifier
- Objective
- Files added/modified
- Implementation summary
- Design decisions and reasoning
- Tests executed and results
- Known issues
- Git commit hash when available
- Next recommended action

Maintain `HANDOFF.md` with:

- Current architecture
- Implemented and pending features
- Important paths and commands
- Current development state
- Outstanding bugs
- Deployment status
- Instructions for another AI model to continue

Never fabricate test results or claim a deployment succeeded without verification.

## 7. Token and cost efficiency

- Prefer targeted edits over rewriting entire files.
- Reuse existing code and patterns.
- Avoid repeatedly reading unrelated files.
- Run local tests through scripts.
- Avoid unnecessary dependencies and services.
- Do not generate redundant documentation.
- Keep concise implementation summaries, but comprehensive system-analysis documents.
- Group related edits into logical changesets.
- Do not call cloud MCP tools when local testing is sufficient.

## 8. Initial task

First inspect the current repository, directory structure, existing instructions, `.git` state, installed local tooling and Obsidian vault.

Then initialize Git safely, create the required project configuration and documentation templates, and establish the local development/testing workflow.

Do not implement booking features, install additional external tools, or deploy yet.

Report what was completed, the exact files changed, and any remaining setup requirements.
## V1 continuation authorization — 2026-10-31

The initial setup-only restriction above is historical. The owner now authorizes
local V1 implementation, dependency installation, Podman PostgreSQL, tests and local
commits. Follow `CODEX_START_HERE.md`, `WORKFLOW_ADDENDUM.md` and `docs/` requirements.
Git is initialized; do not reinitialize it. Cloud creation, remote pushes, Neon
migrations and deployment still require explicit approval.
