# Booking and server reliability — proposed system analysis

> Follow-up (2026-11-01): RELIABILITY-010 is implemented and locally verified.
> See [RELIABILITY_RESULTS.md](../RELIABILITY_RESULTS.md). The original analysis below is preserved as historical evidence.

Status: DESIGN ONLY, 2026-10-31 MYT. No proposed behavior is implemented or tested.
Chisle is off. This document supplements, rather than replaces, the verified booking
and health analyses. User explicitly requested brainstorming → SA → review before
implementation. Source reviewed: local HEAD 77fa3a9 and AUDIT-008 evidence.

## Business purpose

Make valid booking contention produce understandable outcomes while preserving
PostgreSQL overlap protection, and allow accepted requests to finish during shutdown.
The immediate objective is a reliable small guest demo, not new payments or identities.
A committed booking must not be described as definitely absent merely because its
response failed. Complete recovery from lost responses requires a separate durable
idempotency design; serialization alone cannot promise that recovery.

## Brainstorming and decision

| Option | Benefit | Cost or weakness | Decision |
|---|---|---|---|
| Raise request timeout/pool size | May reduce observed timeouts | Can amplify waits; hides cause; no bounded correctness argument | Reject as first fix |
| Go mutex around booking | Simple in one process | Does not coordinate two processes | Reject |
| Retry every SQL failure | Sometimes recovers transient errors | Unknown commit outcome can create misleading results; extra contention | Reject |
| Retry only proven transaction aborts | Can recover deadlock/serialization aborts | Must classify actual SQLSTATE, bound retries, preserve total deadline | Conditional only after diagnosis |
| Transaction-scoped advisory lock per court/hour | Narrow scheduling scope | Key convention/collision concerns; all writers must cooperate; future interval bookings complicate keys | Defer |
| Lock court row, then check occupancy and insert | Existing stable lock target; process-independent ordering; simple reasoning | Serializes different hours on one court; transaction holds a pool connection | Preferred candidate, conditional on evidence |
| Keep exclusion constraint alone and correct another diagnosed fault | Smallest possible change | Requires diagnostic evidence explaining failures | Prefer if diagnosis supports it |
| Durable idempotency | Recover original result after response loss | New storage/API/client contract, key scope and retention policy | Separate follow-up design |

The audit showed 73/100 satisfactory rounds, 2,697 HTTP 503s, and 14 all-503 rounds
with a committed row. It did not prove deadlocks or a specific driver defect. Before
changing synchronization, reproduce with error classification, phase timing, database
wait snapshots and correlated client outcomes. If instrumentation identifies another
cause, revise this proposal rather than mechanically adding locks.

## API endpoint or triggering event

Booking: POST /api/v1/bookings. Lifecycle: SIGINT/SIGTERM or ListenAndServe failure.
No new public endpoint is required for the first remediation. Diagnostic collection
should use structured logs/local test observations, not an unauthenticated debug API.

## Request and response

Existing request remains `{ "court_id": 1, "start_at": "2026-11-01T10:00:00+08:00" }`.
The timestamp is illustrative; tests must choose a future valid venue slot.
Existing successful response retains UUID, court/start/end, amount_sen, confirmed
status and simulated payment. Preserve 400/404/415 validation, 403 origin policy,
429 limiter and 409 known occupancy behavior. A 503 is an unavailable/possibly
uncertain result, never proof of rollback. Send 201 only after confirmed commit.
Do not remap cancellation or unknown database errors to 409.

Proposed response header: server-generated X-Request-ID, correlated to structured
logs. Browser exposure needs Access-Control-Expose-Headers if the UI consumes it;
initial diagnostic scope does not require frontend consumption. Do not trust arbitrary
client-provided identifiers or log bodies, credentials, connection strings or raw SQL
error detail. A request ID provides traceability, not idempotency.

## Ordered function call flow

### Verified current flow

1. `main` creates a pool (max 8), API, HTTP server and signal goroutine.
2. `API.Handler` routes through origin filtering and a five-second context.
3. `API.book` counts attempts, decodes bounded JSON, calls `getCourt`, then `validSlot`.
4. `book` runs INSERT SELECT price via pool.QueryRow; one statement commits implicitly.
5. `book` maps 23P01 to 409, other errors to 503, successful Scan to JSON 201.
6. On signal a goroutine calls Shutdown. ListenAndServe can return before that
   goroutine finishes; main does not join it before pool.Close/process exit.

### Proposed booking flow, if diagnostics justify serialization

1. Preserve origin/body/rate-limit checks and cheap input validation.
2. Begin a short READ COMMITTED transaction using the request context.
3. Select court plus venue configuration, locking **only the court row** with
   `FOR UPDATE OF c`. Missing court returns 404. Validate slot against the loaded
   configuration and current time after obtaining the lock.
4. Issue a **separate statement** to check confirmed overlap. At READ COMMITTED this
   statement obtains a new snapshot after the preceding lock wait. Do not merge
   lock acquisition and overlap check into one CTE/statement with an older snapshot.
5. If overlapping, release transaction and respond 409 without attempting an insert.
6. Otherwise INSERT SELECT the server price in that transaction. Retain the exclusion
   constraint as the final authority, including against noncooperating SQL writers.
7. Commit explicitly. Only a successful commit permits 201. Record transaction phase
   and error class; distinguish definite rejection from unknown commit outcome.
8. Ensure rollback/connection release on every error/early return. Cleanup should use
   a separate short bounded context if the request context is already canceled.

### Proposed server lifecycle

1. Give one coordinator responsibility for serving, signal handling and cleanup.
2. On termination stop accepting requests; invoke Shutdown with a fresh ten-second
   context and wait for its return. Preserve active request contexts while draining.
3. If drain expires, log the outcome, force-close HTTP connections, cancel outstanding
   request work and join tracked handlers before pool cleanup. Pool.Close is not
   itself a deadline mechanism. A bounded forced-exit fallback must be explicit if
   handlers cannot finish; do not claim unlimited cleanup is a bounded shutdown.
4. Close pool only after drain/cancellation coordination. Unexpected serve failure
   must trigger the same cleanup path and a nonzero exit after cleanup, rather than
   calling os.Exit while resources are still owned by that scope.

## Functions and responsibilities

Current `API.Handler` (api.go) owns routing/origin/deadline context; `API.book`
(booking.go) owns decode/validation/write/result; `getCourt(ctx,id)` (slots.go)
returns price in sen, opening hours and venue timezone or an error; `validSlot`
returns eligibility only, never occupancy. `JSON` writes the response and currently
ignores encoding/write errors. `main` owns pool and server lifetimes.

Proposed extraction only if needed: a transaction-aware booking operation accepting
context, court ID and start instant, returning a typed confirmed result or classified
error. Keep it in the existing api package; no speculative repository/service layers.
A transaction-aware configuration query must use the same transaction, not the pool.
A lifecycle helper returning an error may make cleanup testable. Names and signatures
are proposals to settle during implementation, not existing functions.

## Business rules and validation

Preserve whole-hour, future, 30-day, opening-hours rules, integer money and one-hour
sessions. No holds or multi-slot API. Existing shared attempt cap remains a known
fairness limitation; never disable it in product code to pass tests. The repeatable
100-round harness can restart fixtures/processes between rounds as before. A database
lock is per court, so separate hours on the same court wait briefly; separate courts
should proceed independently. No network calls or response writes inside a transaction.
Future administrative schedule edits require their own locking/snapshot contract;
locking court rows alone does not lock facility schedule rows.

## Database queries and transactions

Candidate statements, schematic rather than applied migrations:

```sql
BEGIN ISOLATION LEVEL READ COMMITTED;
SELECT c.price_sen, f.opens, f.closes, f.timezone
FROM courts c JOIN facilities f ON f.id = c.facility_id
WHERE c.id = $1 FOR UPDATE OF c;
-- Separate statement, after lock acquisition:
SELECT EXISTS (
  SELECT 1 FROM bookings
  WHERE court_id = $1 AND status = 'confirmed'
    AND start_at < $3 AND end_at > $2
);
-- Only when no overlap; application chooses one-hour $3:
INSERT INTO bookings(court_id, start_at, end_at, amount_sen)
SELECT id, $2, $3, price_sen FROM courts WHERE id = $1
RETURNING id, amount_sen;
COMMIT;
```

No migration is required for this candidate. Exclusion, FK, status, positive amount
and duration constraints remain unchanged. Conflicts/errors must rollback and release
the row lock. Do not retry unknown commit errors. A future policy for known aborted
transactions (e.g. observed 40P01) must cap attempts and share the original deadline.

## Error handling and edge cases

Measure pool acquisition, lock wait, overlap query, insert and commit separately where
practical. Record request ID, route pattern, status, duration, phase, context outcome
and sanitized SQLSTATE/class. Record pgx pool snapshots/deltas: total/acquired/idle,
acquire duration/count, empty and canceled acquisition counts. Aggregate pool counters
cannot be presented as per-request wait percentiles.

Reproduction must observe pg_stat_activity/pg_locks while failures occur and preserve
SQLSTATE alongside context cancellation. Check process ownership/ports, proxy bypass,
fixture schema and host resource pressure; the original audit used a shared host.
Inspect pgx behavior locally before attributing committed-but-503 results to a driver.
A client disconnect may follow a successful commit even after this repair. Request
IDs can assist diagnosis but cannot authorize a booking lookup. No private lookup is
added without identity or a carefully scoped guest capability design.

## Concurrency and consistency

The database exclusion constraint remains the invariant across every writer. Proposed
court-row serialization organizes cooperating writers; it is not a replacement for
that constraint. Separate post-lock overlap snapshots prevent waiting requests from
seeing only the pre-winner state. Safety is unconditional on process count while the
constraint exists; low latency and one-created/all-conflict results are acceptance
criteria in a defined healthy local environment, not guarantees during outages.

## Sequence diagram

Proposed successful/cooperating-conflict behavior only:

```mermaid
sequenceDiagram
    participant C as Client
    participant A as API process A
    participant B as API process B
    participant D as PostgreSQL
    C->>A: POST booking
    A->>D: Begin; lock court row
    B->>D: Begin; wait for same court lock
    A->>D: Fresh overlap query; insert
    A->>D: Commit (release lock)
    D-->>A: Commit acknowledged
    A-->>C: 201 confirmed
    D-->>B: Lock acquired
    B->>D: Separate fresh overlap query
    D-->>B: Occupied
    B->>D: Rollback/release
    B-->>C: 409 conflict
```

## Source-file references

- backend/internal/api/booking.go:16–86 — actual current booking behavior.
- backend/internal/api/api.go:42–74 — middleware/deadlines.
- backend/internal/api/slots.go:17–41 — configuration and eligibility.
- backend/cmd/server/main.go:21–74 — pool/server and unjoined shutdown.
- backend/migrations/001_initial.sql:11–20 — database invariant.
- backend/internal/api/api_test.go:30–149 — existing integration assertions.
- ../PERFORMANCE_BASELINE.md and ../audit-evidence/ — observed workload and raw results.

## Tests validating the flow

No proposed fix tests have run. Historical audit tests remain historical receipts.
Future implementation gate:

1. Reproduce baseline with diagnostics: one and two API processes, isolated schemas,
   simultaneous same-slot requests, persisted rows and correlated outcomes. Preserve
   failing observations before changing booking logic.
2. Run 100 rounds × 100 requests for both one-process and two-process configurations.
   Each healthy round must have exactly one 201, 99 409s, one row, correct amount and
   no unexpected statuses. Restart per-round processes or allow the real limiter's
   window to expire; do not alter application limits to manufacture a pass.
3. Assert direct-SQL overlap rejection, adjacency, separate-court progress, rollback,
   canceled lock wait, pool recovery, missing court and unchanged validation/statuses.
   Block one court deliberately and show another remains serviceable.
4. Hold a request at a controlled DB barrier, signal server, release barrier and assert
   response completion before process exit. Separately force drain expiry; verify
   bounded termination, cancellation, no leaked processes/connections and logged error.
5. Fault-test commit/response ambiguity; do not assert a failed response means zero
   rows. If idempotency is later implemented, add durable result replay tests then.
6. Run `make verify-local`, fresh uncached Go race/integration tests with loopback
   TEST_DATABASE_URL, and `WATCHPACK_POLLING=true npm test`. Reuse baseline read
   benchmarks plus contention timing/pool metrics after correctness passes.

Tests should observe behavior and persisted state, not merely mock a proposed call
sequence. A deterministic lifecycle barrier is preferable to sleep-based timing.

## Limitations and follow-up

Design review is recorded in ../RELIABILITY_REMEDIATION_PLAN.md. Implementation is
not authorized by completion of this document. Root cause remains an open diagnostic
question; the transaction candidate may change. Idempotency, per-client abuse fairness,
authentication and CI are separate follow-ups. Existing unknown-commit outcomes remain
possible without durable replay semantics. No production changes are proposed here.
