# RELIABILITY-010 — implementation verification

2026-11-01 MYT. Local implementation following DESIGN-009; no cloud calls, push or deployment.
Original AUDIT-008 score and baseline are historical and have not been retrospectively changed.

## Diagnosis and fix

An instrumented repeat of the old write path produced 13/20 satisfactory rounds and
7 all-503 rounds. Logs recorded SQLSTATE **40P01 (deadlock detected)** in insert_scan,
then deadline failures and canceled pool acquisitions. A live pg_stat_activity sample
showed 16 active connections waiting on transactionid locks. This confirms competing
inserts can deadlock and exhaust the request/pool budget; it does not prove the exact
wire-level ordering of every old committed-but-503 outcome.

Implemented READ COMMITTED transaction: lock court row, validate, issue a separate
fresh-snapshot overlap query, insert server price, commit. Retained exclusion constraint.
The reserve helper cleans up before HTTP writes, so a slow response cannot retain its
court lock. There are no blind retries. Explicit transaction scope rolls back errors
before commit; failed commit acknowledgments remain uncertain and are described as such.

Added generated request IDs, route/status/duration logs and sanitized failure phase,
SQLSTATE and cumulative pool counters. No raw driver details, bodies or query strings.
Shutdown now joins serving/draining, cancels/closes on expiry, waits up to one second
for handler cleanup and closes the pool afterward. A permanently stuck handler takes
an explicit nonzero exit rather than an unbounded pool close.

## Validation

- Fresh `go test -count=1 -race ./...` with loopback TEST_DATABASE_URL passed;
  `go vet ./...` and build passed. Unit/integration tests also passed before final race.
- Existing TestIntegration plus new lock-cancellation/independent-court, direct-SQL
  overlap/rollback, diagnostic privacy/correlation, graceful/forced lifecycle and
  failed-listener tests passed. All DB fixtures used temporary schemas.
- Frontend npm lint/typecheck/build passed; WATCHPACK_POLLING=true npm test: 5 passed.
  Generated next-env.d.ts was restored; frontend behavior unchanged.
- Actual SIGTERM while booking waited on court lock: releasing lock produced 201,
  one persisted booking and process exit 0. Probe preserved in audit-evidence.
- Final real-process contention gate (below) verifies count, UUID and price, not
  response codes alone. No 503, transport errors, or additional rows in either run.

| API processes | Rounds | Attempts | Created | Expected conflicts | Unexpected | Mean request ms | p95 ms | p99 ms |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 100 | 10,000 | 100 | 9,900 | 0 | 53.10 | 74.36 | 88.99 |
| 2 | 100 | 10,000 | 100 | 9,900 | 0 | 63.69 | 84.75 | 92.92 |

These are healthy loopback bursts, not production capacity. Each round restarts API
processes to respect the real rate cap, uses a barrier for 100 contenders and resets
only isolated schema rows. Same host class as original baseline; machine load/date
and instrumentation differ. See scripts/concurrency-local.py for strict assertions.

```bash
(cd backend && go build -o /tmp/courtflow-fixed-api ./cmd/server)
python3 scripts/concurrency-local.py --binary /tmp/courtflow-fixed-api --instances 1 --output /tmp/one.json
python3 scripts/concurrency-local.py --binary /tmp/courtflow-fixed-api --instances 2 --output /tmp/two.json
```

## Comparable exploratory read measurements

Same old harness, three-second windows, concurrency 1/10/50/100, one facility/four
courts/two occupied slots, Python urllib on the same host and no TLS. Date advances
with the run. Raw post-read JSON and exact adapted script are retained. CPU is API
process percentage of one logical core; RSS is sampled KiB. DB sessions sampled
at most eight. All runs had zero unexpected responses. No claim of statistically
significant performance improvement from these short runs.

| Endpoint | Concurrency | req/s | Mean ms | p50 ms | p95 ms | p99 ms | API CPU % | RSS KiB |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| facilities | 1 | 1892.9 | 0.52 | 0.50 | 0.65 | 0.78 | 43.0 | 18340 |
| facilities | 10 | 3022.7 | 3.30 | 3.17 | 5.24 | 6.28 | 76.3 | 19656 |
| facilities | 50 | 3024.6 | 16.15 | 15.31 | 27.57 | 34.26 | 73.4 | 20448 |
| facilities | 100 | 2906.5 | 32.15 | 30.63 | 56.54 | 70.22 | 73.8 | 20692 |
| availability | 1 | 1159.6 | 0.86 | 0.82 | 1.11 | 1.32 | 59.3 | 21088 |
| availability | 10 | 2683.4 | 3.71 | 3.57 | 5.65 | 6.69 | 184.1 | 20724 |
| availability | 50 | 2749.3 | 17.81 | 16.92 | 30.85 | 37.57 | 181.6 | 21120 |
| availability | 100 | 2667.8 | 34.52 | 32.87 | 62.08 | 77.63 | 175.2 | 21244 |

## Review and remaining limits

Chisle stopped before comment pass, maintainability review and final validation.
Reviewed transaction cleanup/snapshot ordering, error classes, caller compatibility,
request-log privacy and handler admission vs WaitGroup joining. Fixed the cleanup
boundary so response writes occur after transaction release. No schema/dependency
change and no process-local reservation lock.

No durable idempotency, per-client limiter, private booking access or readiness
endpoint added. Different hours on one court serialize. Future facility schedule
writes need their own concurrency contract. Unknown commit/response delivery remains
possible. Fault-injected network loss at commit and permanently stuck-handler forced
exit are not directly tested; graceful and cancellation-expiry branches are tested.
The local successful contention result does not guarantee identical outcomes during
outages. Pool statistics are cumulative failure snapshots, not acquired-wait percentiles.
New logs increase log volume; use hosting retention controls rather than logging
request bodies. Production still runs its previously deployed version.

Temporary test schemas/processes were cleaned up. A local database restart/image
rebuild was not required or performed. Final diff/format checks accompany this receipt.
