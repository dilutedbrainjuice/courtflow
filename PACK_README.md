# CourtFlow — Codex handoff pack

This pack is **not** your complete CourtFlow application and does not claim to have modified your Fedora machine. It contains handoff instructions, project-local Codex skills, local test scripts and product/deployment specs, and optional MCP configuration guidance.

Use: extract/copy the files into `/home/terasit/VSProjects/court` **after comparing with existing files** (especially `AGENTS.md`, `DEVELOPMENT.md`, `Makefile`, and Obsidian docs). Alternatively hand `CODEX_START_HERE.md` to Codex and ask it to integrate the pack. Do not overwrite pre-existing documents.

First: install the *official Chisle project-local skill* (not the global Codex installer):

```bash
cd /home/terasit/VSProjects/court
mkdir -p .agents/skills/chisle
curl -fsSL https://raw.githubusercontent.com/JayPokale/Chisle/main/skills/chisle/SKILL.md \
  -o .agents/skills/chisle/SKILL.md
curl -fsSL https://raw.githubusercontent.com/JayPokale/Chisle/main/LICENSE \
  -o .agents/skills/chisle/LICENSE
```

Review downloaded skill before activating. It is MIT licensed with attribution. The Chisle Codex global `npx chisle` installer inserts always-on instructions in `~/.codex/AGENTS.md`, so avoid it for this project's coding-only requirement.

For coding, invoke `$chisle` in an implementation chat. For analysis, open a separate documentation chat and invoke `$system-analysis` with no Chisle. Local tests: `$local-first`. Only approved releases: `$release-gate`.

**Owner must authorize external account actions, remote repository creation, production DB migrations, and deployment.**
