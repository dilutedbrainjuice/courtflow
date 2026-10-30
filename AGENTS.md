# Repository Guidelines

## Project Structure & Module Organization

This repository is currently an initial workspace without application source code, tests, assets, or package manifests. No framework or module layout has been established.

When adding the first implementation, organize files by responsibility and document the layout in `README.md`. Suggested directories are `src/` for application code, `tests/` for automated tests, and `assets/` for static resources; introduce them only as needed. Keep configuration files at the repository root where tooling expects them.

## Build, Test, and Development Commands

No build, test, or local development commands are configured yet. When introducing a toolchain, provide reproducible commands in `README.md` for installing dependencies, running locally, building, and testing. Use the selected ecosystem’s package manifest or task runner as the authoritative command definition, and commit its dependency lockfile when applicable.

## Coding Style & Naming Conventions

No language, formatter, or linter is configured. Follow the chosen language’s standard conventions and add shared formatting and linting configuration with the first implementation. Use consistent indentation within each file, descriptive names, and small modules with clear responsibilities. Avoid adding competing tools for the same formatting task.

## Testing Guidelines

No testing framework or coverage threshold exists. Add tests alongside new behavior using the selected ecosystem’s standard framework and file naming convention. Document the exact test command and any required setup. Prioritize observable behavior, edge cases, and regression tests for bug fixes.

## Commit & Pull Request Guidelines

Git history is unavailable in this workspace, so no existing commit convention can be confirmed. Use concise, imperative commit subjects, such as `Add booking validation`, and keep changes focused.

Pull requests should explain the purpose, summarize behavior changes, link relevant issues, and report validation performed. Include screenshots for visual changes and identify any required configuration or migration steps.

## Security & Configuration

Keep credentials and local secrets out of committed files. When configuration is introduced, document required variables and provide examples containing placeholders rather than real values.

## CourtFlow Permanent Instructions

Read and follow `DEVELOPMENT.md` before working in this repository. It contains
project-specific local-first development, documentation, testing, and deployment
requirements. The confirmed root is `/home/terasit/VSProjects/court`.

## V1 workflow

Follow `WORKFLOW_ADDENDUM.md` and `CODEX_START_HERE.md`. Use repository Chisle only
for implementation; finish implementation, explicitly stop Chisle, and complete a
clear explanatory comment pass before code review, final validation and
system-analysis writing. Use
`.agents/skills/local-first` for checks and `.agents/skills/system-analysis` for
verified flow documentation. The setup-only statements above describe initial
state; current architecture and commands belong in README and HANDOFF.
