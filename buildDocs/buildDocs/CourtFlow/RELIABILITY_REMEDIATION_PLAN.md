# Reliability remediation plan and design review

> Follow-up (2026-11-01): RELIABILITY-010 is implemented and locally verified.
> See [RELIABILITY_RESULTS.md](RELIABILITY_RESULTS.md). The original analysis below is preserved as historical evidence.

2026-10-31 MYT — DESIGN REVIEW ONLY. No implementation, tests of a fix or deployment.
Read [the proposed SA](system-analysis/reliability-remediation-proposal.md) first.

## Ordered plan

1. **Preserve and diagnose:** retain AUDIT-008 evidence; prepare isolated local
   reproduction with phase/error/pool/wait diagnostics. Confirm fixture isolation,
   direct loopback transport and process ownership. Identify why committed rows can
   coexist with all-503 responses. Root-cause evidence is a gate, not an assumption.
2. **Choose the booking fix:** prefer the smallest fix supported by diagnosis. The
   reviewed candidate is READ COMMITTED court-row lock → separate overlap read →
   insert → explicit commit, retaining the exclusion constraint. Do not increase
   timeouts, blanket-retry or substitute a process-local mutex as the initial answer.
3. **Repair lifecycle:** coordinate serve exit and signal-driven drain, wait for
   Shutdown, cancel/close on expiry, then release the pool. Include startup failure
   cleanup and an explicit bounded termination policy.
4. **Verify:** deterministic single/two-process 100×100 persisted-state tests; partial
   overlaps/adjacency, cancellation and independent courts; controlled SIGTERM tests;
   existing Go/frontend checks and comparable local benchmarks. Keep 409 separate
   from unexpected failures. No production testing.
5. **Document verified behavior:** after any later authorized implementation, stop
   Chisle, complete comments/review, run final checks, update actual booking/health
   SA, audit status, HANDOFF and BUILD_LOG. Preserve the original baseline.

## Review against code and evidence

Review performed after drafting the SA, against booking.go, api.go, slots.go,
main.go, migration 001, existing tests and the audit results. This is an author
review, not an independent second reviewer or a runtime validation.

| Review question | Finding / disposition |
|---|---|
| Is root cause proven? | No. Exclusion waits canceled near deadline are evidence, not proof of deadlock. Diagnosis gate mandatory. |
| Does locking work across API instances? | Database court-row locks coordinate cooperating writers; Go locks would not. Retain exclusion for other writers. |
| Could a lock/check CTE still read stale occupancy? | Yes, single-statement snapshot is a design hazard. Require separate READ COMMITTED statements after lock acquisition. |
| Can one court serialize unrelated hours? | Yes. Acceptable provisional tradeoff for four-court demo; measure before choosing finer advisory keys. |
| Does this solve lost responses? | No. Explicit commit handling improves classification, not exactly-once delivery. Durable idempotency remains separate. |
| Does shutdown wait just because Shutdown has a timeout? | No. Main must coordinate completion and cancellation; pool.Close is not a bounded drain. |
| Can retries safely cover every 503? | No. Unknown commit outcomes forbid blind retry; known aborts require bounded, diagnosed policy. |
| Does the limiter invalidate repeated load tests? | Yes unless fixtures/processes/windows account for it. Keep product limit intact and disclose harness resets. |
| Are new services or migrations necessary? | Not for the preferred candidate or lifecycle fix. No Redis, queues or new framework required. |
| Can existing docs be rewritten as though fixed? | No. Proposal is explicitly separate from verified flow SA and historical audit. |

**Review outcome:** suitable as a conditional implementation plan; not approved as a
proven remedy. No known design-level integrity regression in the candidate, but
latency, fairness and root cause must be resolved experimentally. A successful review
does not authorize implementation in this request.

## Smallest initial implementation scope, when requested

Diagnostic error/request/pool visibility, evidence-supported booking correction,
joined bounded shutdown, and regression tests. Defer idempotency and limiter redesign
until their API/ownership/fairness requirements have separate designs. If diagnostics
show the proposed row locking is unnecessary, omit it and record the evidence.
