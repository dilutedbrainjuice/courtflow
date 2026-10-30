---
name: system-analysis
description: Write or update comprehensive CourtFlow Obsidian Markdown system-analysis documents for a verified API, controller or business workflow. Use only for documentation tasks. Never activate Chisle during this task.
---

# CourtFlow System Analysis — explanatory mode (Chisle OFF)

1. Read `AGENTS.md`, `DEVELOPMENT.md`, the flow's source, relevant caller/callee files, SQL migration/schema and tests. Stop Chisle if activated. Prefer a **new Codex conversation** that has not invoked Chisle.
2. Create/update one Markdown per API/business workflow in `buildDocs/buildDocs/CourtFlow/system-analysis/`. Match existing `TEMPLATE.md` without deleting sections.
3. Include: objective and who uses it; HTTP method/path; auth and actors; request/response and status codes; numbered *actual* route → middleware → handler → service → repository → SQL/transaction → response call chain with file/function references; all important parameters, returns and validations; DB reads/writes and transaction boundaries; errors and recovery; concurrency and race guarantees; practical example; Mermaid sequence diagram; tests (commands and assertions); known limits and follow-ups.
4. For significant helper functions, document each within its caller flow. Keep trivial helpers in the call trace rather than making redundant Markdown documents.
5. Explain **why** the implementation works, especially booking conflicts and what the database guarantees. Treat `[start,end)` slots and SQLSTATE `23P01` carefully. Do not claim exactly-once semantics.
6. Use full explanatory prose: never minimize completeness because of a token-saver. Do not invent call graphs, tests or timings. If code is incomplete, label gaps clearly.
7. Update architecture records if the flow introduces a new invariant or decision. Append to `BUILD_LOG.md` and refresh `HANDOFF.md`.
