# Confirming a guest demo booking

Verified against the RELIABILITY-010 working tree on 2026-11-01 MYT. This supersedes
the earlier implicit-transaction implementation. See RELIABILITY_RESULTS.md for receipts.

## Business purpose

A guest reserves one court for one hour, with immediate simulated confirmation and
no personal data/payment provider. The change makes cooperating writers wait before
inserting, avoiding the deadlocks observed under same-slot contention while retaining
the database exclusion constraint for every writer.

## API endpoint or triggering event

Public POST /api/v1/bookings, called by Home.book in frontend/app/page.tsx.
No identity, ownership, cancellation or idempotency endpoint is introduced.

## Request and response

Exact Content-Type application/json; one object at most 1024 bytes, for example
`{"court_id":1,"start_at":"2026-11-03T02:00:00Z"}`. Choose a future eligible time.
Success remains 201 with UUID id, court_id, UTC start_at/end_at, integer amount_sen,
status confirmed and payment simulated. Error shape remains `{"error":"message"}`.
All responses receive a generated X-Request-ID; it is correlation, not an access token.
A commit error returns 503 with “booking outcome uncertain; refresh availability”.

## Ordered function call flow

1. API.Handler installs observe around the origin/header/deadline wrapper and ServeMux.
   observe generates a request ID, records status and logs completion without raw URLs.
2. API.book applies the per-process 120/minute attempt cap, exact content type,
   bounded strict JSON decoding, trailing-data rejection and positive court ID.
3. book calls reserve(request,courtID,start), which owns transaction lifetime.
4. reserve begins READ COMMITTED using the request context. It installs a deferred
   rollback with an independent one-second cleanup context.
5. A joined court/facility query uses FOR UPDATE OF c. Only the court row is locked.
   Missing court becomes 404; other errors are classified/logged and become 503.
6. validSlot(start,time.Now(),court) checks eligibility after acquiring the lock.
7. A separate SELECT EXISTS checks confirmed overlap using start_at < end and
   end_at > start. At READ COMMITTED it sees a fresh snapshot after the lock wait.
   Occupancy becomes 409 without issuing another conflicting INSERT.
8. INSERT SELECT obtains server price and UUID inside the transaction. The exclusion
   constraint still rejects conflicting writers that did not take the court lock;
   SQLSTATE 23P01 maps to 409. Other failures become 503.
9. reserve commits explicitly. Only acknowledged commit returns a success result.
   Deferred rollback handles every early return; ErrTxClosed after commit is normal.
10. After reserve has finished cleanup, book writes fail/JSON. Slow response writes
    therefore do not hold a court lock or transaction connection.

## Functions and responsibilities

book owns HTTP parsing and response output. reserve returns a result map, HTTP status
and safe error message; it owns begin, lock, eligibility, overlap, insert, commit and
rollback. Its start parameter is an instant, courtID an integer; prices are sen.
validSlot in slots.go returns eligibility only. getCourt remains used by availability;
booking loads configuration within its transaction instead. databaseError records
request ID, failed phase, elapsed time since reserve began, safe SQLSTATE/error class
and cumulative pool statistics. It does not log raw driver messages or body values.
observe/responseStatus capture ordinary response status and duration; Unwrap supports
http.ResponseController. JSON/fail retain the existing public encoding contract.

## Business rules and validation

Strictly future whole-hour start, no later than now.AddDate(0,0,30), venue opening
hours and timezone; exactly one hour, server-owned price/status/duration. Unknown
fields, malformed/trailing input and nonpositive ID return 400. Invalid stored zone
fails eligibility. The mutex protects only attempt counting, never reservations.
The limiter remains shared by all visitors to one process and resets on restart.

## Database queries and transactions

```sql
BEGIN ISOLATION LEVEL READ COMMITTED;
SELECT c.price_sen,f.opens,f.closes,f.timezone
FROM courts c JOIN facilities f ON f.id=c.facility_id
WHERE c.id=$1 FOR UPDATE OF c;
-- Separate statement: fresh snapshot after the lock wait.
SELECT EXISTS(SELECT 1 FROM bookings
 WHERE court_id=$1 AND status='confirmed' AND start_at < $3 AND end_at > $2);
-- Only if unoccupied:
INSERT INTO bookings(court_id,start_at,end_at,amount_sen)
SELECT id,$2,$3,price_sen FROM courts WHERE id=$1 RETURNING id,amount_sen;
COMMIT;
```

$1 court ID, $2 start, $3 start plus one hour. Price is read in INSERT SELECT.
Early outcomes rollback before HTTP response. No migration changed. The GiST
exclusion over court_id and `[start,end)` remains the final overlap invariant,
with existing FK, duration, status and positive-price/amount checks.

## Error handling and edge cases

403 origin, 415 content type, 400 validation, 404 court, 409 occupancy/23P01,
429 attempt cap with Retry-After:60, 503 database/cancellation/uncertain commit.
No automatic retry is introduced. A failed commit acknowledgment or lost HTTP
response can still conceal a committed booking; this is not exactly-once delivery.
Rollback uses a fresh bounded context even if the request has expired. Failed
rollback is logged. Existing pgx cleanup discards unusable connections as needed.

Structured failure fields include phase, class, SQLSTATE, pool total/acquired/idle,
acquire count/time, empty acquisition and canceled acquisition counts. Counters are
cumulative per process, not per-request wait percentiles; elapsed_ms is cumulative
within reserve, not a separate duration for each phase. Expected conflicts do not
produce database error logs. Request completion logs include generated ID, method,
matched route, status and elapsed milliseconds; no query strings or client IDs.

## Concurrency and consistency

Court-row locks coordinate independent API processes. Same-court requests for
unrelated hours serialize briefly; separate courts remain independent. Facility
schedule rows are not locked, so future schedule-write APIs need their own policy.
The separate overlap query matters: a combined lock/check statement can retain a
snapshot from before a competing commit. Noncooperating SQL writers are protected
by the exclusion constraint, not by the application protocol.

Diagnosis reproduced 40P01 deadlocks and transactionid waits in the old insert path,
with pooled requests timing out. Explicit transactions also keep an INSERT from
implicitly committing before its Scan succeeds; failures before commit roll back.
Unknown outcomes at commit or response delivery remain possible. The current local
100-round gates demonstrate reliable healthy contention, not outage immunity.

## Sequence diagram

```mermaid
sequenceDiagram
    participant U as Guest
    participant A as book / reserve
    participant D as PostgreSQL
    U->>A: POST court and start
    A->>A: Attempt cap and JSON validation
    A->>D: Begin; lock court and load rules
    A->>A: Validate time
    A->>D: Separate overlap query
    alt Occupied
        A->>D: Rollback and release
        A-->>U: 409 conflict
    else Available
        A->>D: INSERT SELECT price
        A->>D: Commit
        D-->>A: Acknowledged
        A-->>U: 201 confirmation
    end
```

## Source-file references

backend/internal/api/{booking,api,diagnostics,slots}.go;
backend/migrations/001_initial.sql; frontend/app/page.tsx.
Tests: api_test.go, reliability_test.go, diagnostics_test.go;
scripts/concurrency-local.py.

## Tests validating the flow

Existing TestIntegration still verifies 24 contenders, one row, adjacency, different
courts, malformed requests, CORS and rate limiting with real isolated PostgreSQL.
TestBookingLockCancellationAndIndependentCourt waits for an observed DB lock wait,
proves a second court succeeds, cancels the first request, checks no write and then
checks the pool/lock recover. TestDirectOverlapAndRollback verifies partial overlap,
adjacency and rollback of an earlier insert in an aborted transaction.
TestRequestDiagnostics checks request correlation, matched route/status and omission
of untrusted request IDs/query values. Go race tests execute these assertions.

Run `make test` with local PostgreSQL. For process-level stress, build a local binary
and run `python3 scripts/concurrency-local.py --binary /absolute/path/to/binary
--instances 2 --output /tmp/two.json` (one shell line). Repeat with --instances 1.
Defaults are 100 rounds and 100 contenders. Each round checks exact statuses, row
count, returned UUID and amount against PostgreSQL. Fixtures use unique schemas;
process restarts keep the real limiter from masking contention. No cloud target.

## Limitations and follow-up

No durable idempotency, fairness improvement, identity or multi-slot transaction.
No sustained capacity claim; retain original baseline and see RELIABILITY_RESULTS.md.
Connection budget still eight per process. Response-write errors remain ignored by
JSON after headers; completion logs record HTTP status, not proof of client delivery.
