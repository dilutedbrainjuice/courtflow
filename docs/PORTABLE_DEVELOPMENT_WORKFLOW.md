# Portable development workflow — setup and operating instructions

This file transfers the workflow developed for CourtFlow to another repository,
machine or coding model. It is self-contained: the receiving model can follow it
without access to the original chat or any plugin. It describes a process, not the
state, credentials, providers or release authorization of the original project.

## How to use this file

1. Copy this file into the destination repository, preferably under `docs/`.
2. Give the receiving model the setup prompt below and the path to this file.
3. Let it inspect the destination repository and merge these rules into its existing
   contributor instructions. Preserve existing content and user changes.
4. Review the generated setup summary, then give it a normal implementation task.

### Setup prompt to give the receiving model

> Read this entire workflow file and adopt it for this repository. First inspect
> the existing project instructions, source, manifests, tests, documentation, Git
> state and installed local tools. Preserve existing files and user changes. Find
> the actual repository root and documentation root; do not copy paths or hosting
> assumptions from another project. Merge the workflow into the instruction files
> your environment actually loads and create any missing project documentation
> templates. Adapt checks to the existing language and stack. Do not implement new
> product features, create cloud resources, push or deploy during workflow setup.
> Report the files changed, verified local commands and any setup gaps. Keep the
> implementation/comment/review/testing/documentation order exactly as specified.

## 1. Establish project context before changing files

Inspect first; do not scaffold over an existing application. Read the project's
AGENTS.md or equivalent, DEVELOPMENT.md, README, workflow addenda, manifests,
lockfiles, Makefile/task runner, current handoff and build log where present.
Inspect Git status and relevant history. Preserve existing work, Git history,
documentation vaults and personal settings. Never reinitialize a working repository.

Resolve these values from the repository and owner context:

| Setting | What to establish |
| --- | --- |
| Repository root | Existing checkout's absolute path |
| Documentation root | Existing tracked docs or Obsidian project directory |
| Application components | Frontend, backend, database, workers and integrations actually present |
| Local commands | Install, format, lint, typecheck, test, build and run commands |
| Local dependencies | Database/container setup, fixtures and required services |
| Configuration | Required variable names; examples contain placeholders only |
| Cloud policy | Provider choices, allowed budget and manual release requirements |
| Commit policy | Whether local commits are authorized; remote writes remain separate |

Do not invent architecture, installed tools, successful tests or prior approvals.
Ask only for missing decisions that materially block useful work; continue work
that does not depend on the answer. User authorization applies to its stated scope.

## 2. Install the workflow into repository instructions

Reuse existing instruction files. If the model loads AGENTS.md, add a short pointer
there to DEVELOPMENT.md and this workflow. If it loads another filename, use its
supported mechanism instead. A Markdown file has no automatic enforcement unless
the model actually receives or reads it.

Merge permanent rules into DEVELOPMENT.md and keep a concise continuation entry
point if useful. Do not create contradictory copies of the workflow. Designate one
canonical specification and keep pointers or summaries consistent with it.

Suggested instruction text:

> Follow docs/PORTABLE_DEVELOPMENT_WORKFLOW.md. Finish implementation in the local
> Chisle/implementation mode, explicitly stop that mode, then complete the explanatory
> comment pass before maintainability/code review, final local validation and
> system-analysis writing. Keep implementation brevity constraints off throughout
> comments, review and documentation. Update the append-only build log and current
> handoff. Production work and remote publication require owner authorization.

If this file is placed elsewhere, update the pointer to its real path.

## 3. Keep implementation mode local and explicitly bounded

In the original project, Chisle is a project-local skill used only for implementation.
Its useful role is to favor the smallest complete solution: reuse existing code,
prefer standard facilities, avoid speculative abstractions and dependencies, and
implement only the required behavior with essential tests.

If the destination model supports local skills, it may use the existing trusted
Chisle skill. To transfer the exact original skill, copy the entire
`.agents/skills/chisle/` directory, including LICENSE and UPSTREAM_COMMIT, and adapt
its discovery path to the destination tool. Do not globally install an always-on
ruleset. Do not assume `$chisle` is a universal model command.

If the exact skill is unavailable, follow the implementation principles above as an
explicit temporary mode. Do not claim the upstream skill was installed or invoked.
This fallback is sufficient to execute the workflow; no plugin is required.

Implementation efficiency never permits compressed unreadable code, missing error
handling, omitted tests, weak validation or incomplete product behavior. Critical
correctness and clarity take priority over line count or token savings.

## 4. Required cycle for every feature, fix or refactor

### Step 1 — inspect and define the change

Read the relevant source, callers, dependencies, tests and existing SA. State the
observable behavior being changed and the boundaries affected. Understand the
current design before choosing a new abstraction. Keep the task cohesive.

### Step 2 — finish implementation

Use local implementation mode for application changes. Finish the behavior and
its essential tests before the dedicated explanatory comment pass. Preserve
validation, resource cleanup, accessibility and compatibility. Use the established
formatter and language conventions. Testing during implementation is encouraged;
the final validation gate still follows comments and review.

### Step 3 — explicitly stop Chisle, then write comments

State `stop chisle` or the destination model's equivalent. Do this **before** the
comment-writing pass, not only when documentation begins.

Re-read the finished code, its callers and relevant database rules in normal
explanatory mode. Write or revise comments that explain:

- Intent and the reason for a non-obvious decision.
- Public contracts, parameter meanings, units and important return/error behavior.
- Ownership and lifetime of resources and shared mutable state.
- Invariants, concurrency guarantees and where those guarantees are enforced.
- Boundary behavior such as timezones, intervals, cancellation and retries.

Remove stale, misleading or redundant comments. Do not narrate obvious statements
or imply guarantees that the code does not provide. Prefer clear names and control
flow over comments that attempt to excuse confusing code. Complete this pass before
proceeding to review. Existing necessary comments need not be deleted during coding.

### Step 4 — maintainability review and code review

Keep Chisle off. Review the whole changed flow and its callers/tests, not just lines
that were edited. Check correctness, validation, errors, resource ownership, races,
compatibility, security, accessibility and test coverage.

Use small cohesive functions and responsibility-based files. Apply SOLID principles
pragmatically. For Go, prefer standard conventions, explicit errors, composition
and small interfaces at real consumer boundaries. Do not manufacture layers,
constructors or interfaces solely to look architectural. For React, review state
transitions, effect dependencies/cancellation, stale responses, loading and error
states. Keep SQL parameterized and readable.

Passing tests and formatting do not replace this review. Fix meaningful findings.
If fixes change implementation, repeat the explicit mode stop, comment pass and
review for the affected code before final validation. Do not let implementation
brevity constraints leak into review or explanation.

### Step 5 — validate locally and fix failures

Use reproducible commands from the project's manifest/task runner. Install from
its lockfile and run relevant checks against local or explicitly isolated test
services. Run dependent installation/build operations sequentially.

Typical checks, adapted to the repository:

| Component | Relevant verification |
| --- | --- |
| Go | gofmt, go vet, unit/integration tests, race detector, compilation |
| TypeScript/frontend | Existing formatter/linter, typecheck, production build, relevant browser tests |
| Database | Migration compatibility, constraints, transaction/error behavior, concurrency assertions |
| Deployment image | Local image build and minimal local HTTP smoke when image/runtime changes |
| Documentation only | Source verification, references, Markdown structure/diagrams and diff checks |

Test observable behavior and meaningful edge cases. Avoid tests that simply mirror
implementation. A concurrency invariant should be checked against the actual local
database, not inferred from a mock. Distinguish mocked browser checks from real API
and database integration. Skipped tests are not passes. Record failures, blockers
and checks not run. After fixes, repeat the comment/review gates and affected checks.

Do not rerun the entire application suite for comments-only or documentation-only
changes without a concrete reason. Verify those edits did not change behavior.

### Step 6 — write or update system analysis

Keep Chisle off. Re-read the verified implementation and write complete explanatory
prose for a developer new to the project. Maintain one document per meaningful
API/business workflow; document significant helpers in the flow that uses them.

For every new or modified flow, cover all of these:

1. Business purpose, actors and triggering action.
2. Endpoint/event, authentication/authorization and boundary assumptions.
3. Request/response schemas, units, status codes and a practical example.
4. Numbered actual call sequence from caller through middleware, handlers, helpers,
   services/repositories **if present**, database and response/UI updates.
5. Significant functions: source path, responsibility, inputs, outputs and errors.
6. Business rules and validation, with reasons and concrete boundary examples.
7. Actual SQL/queries, parameter meanings, writes and transaction boundaries.
8. Error paths, recovery, retries, cancellation and uncertain outcomes.
9. Concurrency and consistency: what is guaranteed, by which component, and what
   can become stale. Do not claim exactly-once semantics without an implementation.
10. A Mermaid sequence diagram matching the real participants and calls.
11. Source references, test names/commands, assertions and coverage gaps.
12. Known limitations and follow-up work, clearly distinguished from implemented behavior.

Explain how and why the code works. Do not replace explanations with arrow chains,
generic copied paragraphs or diagrams alone. Do not invent architectural layers.
Read test assertions before describing coverage. Separate historical test receipts
from checks actually run for the current change. Update architecture documents when
real boundaries, invariants or design decisions change.

### Step 7 — update history and handoff

Append a dated entry to BUILD_LOG.md; never replace existing history. Include the
objective/change identifier, files, implementation summary, reasons for decisions,
review findings, exact checks/results, known limitations, commit when available
and next action. Never include secrets or fabricate a result.

Refresh HANDOFF.md as a current snapshot: architecture, completed/pending work,
important paths/commands, local state, known bugs, test receipt, deployment state
and precise continuation instructions. Distinguish local commits, pushed commits
and deployed commits. Historical notes must not obscure the current state.

### Step 8 — inspect and commit the cohesive change

Review Git status, the complete diff, staged files, whitespace and accidental
secrets/generated files. Use `git diff --check` when Git is available. Preserve
unrelated user work. Commit locally if authorized, with a concise imperative subject.
Push, cloud provisioning, migration and deployment are separate authorized actions.
Do not deploy merely because code was committed or tests passed.

## 5. Local-first and manual-release policy

Develop, debug and test locally. Use local database containers or an equivalent
isolated database with persistent storage where needed. Keep credentials out of
Git; provide environment examples with placeholders. Preserve personal editor or
Obsidian settings outside tracked project documents.

Production is not a test environment for concurrency, stress, migration experiments,
data cleanup experiments or repeated debugging. Use simulated payments unless the
owner has separately authorized and specified a real integration.

Keep Git-triggered deployments disabled. Where infrastructure synchronization can
change production separately, control that too. Do not assume one provider switch
disables every deployment mechanism: environment updates may initiate deployments.

A release requires successful applicable local checks and explicit owner approval.
Confirm provider access, target project/environment, plan/budget and commit. Review
migrations and backup/rollback needs before applying them. Deploy manually in an
order compatible with the API and schema. Then perform only the essential approved
health and happy-path smoke checks against fictional/dedicated demo data.

Verify provider state before reporting success. Record real URLs, deployment IDs,
commit SHA, migration version, timestamps, observed smoke results and limitations.
A public-access/security change may require separate explicit authorization. If a
tool blocks an action, explain the actual blocker and finish unaffected work.

## 6. Suggested documentation layout

Reuse an existing documentation root. Otherwise this is a starting point:

```text
AGENTS.md                     # Or the receiving tool's instruction entry point
DEVELOPMENT.md                # Permanent project rules and workflow pointer
README.md                     # Current product, layout and runnable commands
docs/
  PORTABLE_DEVELOPMENT_WORKFLOW.md
  architecture/
  system-analysis/
    TEMPLATE.md
    <one document per implemented flow>.md
  BUILD_LOG.md                # Append-only change history
  HANDOFF.md                  # Current continuation snapshot
  PRODUCTION_RUNBOOK.md       # Manual release and rollback steps
```

Do not create duplicate roots if a project already uses an Obsidian vault or another
layout. Optional local skills can encode implementation, local-first verification,
system-analysis and release gates when supported, but plain instructions suffice.

## 7. Setup completion and definition of done

Workflow setup is complete when the receiving model has inspected the project,
merged the rules into loaded instructions, established actual paths/commands,
provided missing templates and reported remaining environment gaps. It must not
claim that unexecuted commands were verified. Missing runtime dependencies can be
reported without blocking the installation of documentation rules.

A normal implementation change is complete when working behavior, the Chisle-off
comment pass, maintainability/code review, applicable local checks, verified SA,
build log, handoff and diff review are complete. A release is complete only after
its separately approved deployment and smoke checks are actually verified.

### Continuation prompt for later sessions

> Read the repository instructions, DEVELOPMENT.md, current HANDOFF.md and the
> relevant system analysis. Continue the existing project without overwriting
> unrelated work. Follow implementation → stop Chisle → explanatory comments →
> maintainability/code review → local validation → verified SA → build log/handoff
> → reviewed commit. Keep Chisle off after implementation. Report what changed,
> checks actually run and remaining limitations. Use existing authorization only
> within its stated scope; do not infer permission to push or deploy.
