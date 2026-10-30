# CourtFlow — tool, skill and MCP policy

## Local mode — default
- Codex + Go + Node.js/npm + Git + Podman + local PostgreSQL.
- Official Chisle installed **only as a repository skill** at `.agents/skills/chisle/SKILL.md`; copy original MIT LICENSE alongside it. Do not run `npx chisle` globally because Codex then receives persistent global rules, conflicting with documentation mode.
- Additional repo skills: `.agents/skills/local-first`, `.agents/skills/system-analysis`, `.agents/skills/release-gate`.
- Coding: separate implementation chat invoke `$chisle`; docs: a new documentation chat invoke `$system-analysis` (no `$chisle`). This is a strong practical separation, not a magical global off-switch.
- Avoid external MCP calls during coding/tests, both for token/cost and to eliminate accidental production changes.

## Deployment mode — only after user authorizes cloud access
There are official remote MCPs for all selected providers. Connect through Codex CLI and follow browser OAuth prompts:

```bash
codex mcp add vercel --url https://mcp.vercel.com
codex mcp add neon --url https://mcp.neon.tech/mcp
codex mcp add render --url https://mcp.render.com/mcp --oauth-client-id codex
codex mcp list
```

Use `codex mcp login <name>` if the client asks for authentication rather than launching OAuth automatically. Prefer **Render's official Codex plugin** through Plugins in the Codex desktop app if available; it bundles Render skills and MCP OAuth setup, so don't install both duplicated ways unnecessarily.

These connections can potentially change live services, databases, and deployments. Explicitly forbid any MCP *write* without user approval. Prefer read-only inspection for verifying an approved release. Connection is optional until the owner is ready to deploy.

Official docs:
- Vercel MCP: https://vercel.com/docs/agent-resources/vercel-mcp
- Neon MCP: https://neon.com/docs/ai/neon-mcp-server (verify current endpoint before configuration)
- Render MCP: https://render.com/docs/mcp-server

## GitHub
Git CLI is sufficient for local work. At the approved publication milestone, install/check `gh` if helpful, authenticate explicitly (`gh auth login`), and create a **public** GitHub remote only if no corresponding remote already exists. Check `git remote -v` first. No GitHub MCP needed for MVP; avoid extra tool output and permissions.

## Hosting providers and approvals
- Vercel Git-triggered deployments: disable via `frontend/vercel.json` `git.deploymentEnabled=false` (when Vercel root is `frontend`).
- Render automatic deploy: disable (Auto-Deploy Off, or Blueprint `autoDeployTrigger: off`).
- Neon: no schema edits, tests, or load tests during local development.
- First public deployment: execute only after owner reviews runbook and grants explicit release authorization; owner performs minimal production smoke test.
