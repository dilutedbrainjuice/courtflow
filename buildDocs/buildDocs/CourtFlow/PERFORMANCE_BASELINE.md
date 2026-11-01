# CourtFlow local performance baseline

> Follow-up (2026-11-01): RELIABILITY-010 is implemented and locally verified.
> See [RELIABILITY_RESULTS.md](RELIABILITY_RESULTS.md). The original analysis below is preserved as historical evidence.

2026-10-31 MYT; local source HEAD `77fa3a9`. These are short exploratory local measurements, not production capacity or an SLA. No production service was contacted.

## Environment and reproducibility

- Intel Core i5-10400F @ 2.90 GHz; 6 physical cores / 12 logical CPUs; Linux x86_64. Host memory 15,911 MiB; snapshot available 4,641 MiB, swap in use 6,489 MiB. Shared desktop machine, no CPU pinning, no controlled idle host. Browser checks overlapped part of contention testing, not the read benchmarks. Do not infer isolated hardware limits.
- Go `go1.26.10-X:nodwarf5 linux/amd64`; Node v22.23.1. PostgreSQL 17.11 in existing rootless Podman container `courtflow-postgres`, localhost:5433. Native Go server, no reverse proxy or TLS on loopback.
- One temporary schema with migration 001 and its 1 fictional facility/4 courts, RM25/hour. API DATABASE_URL sets search_path to that schema,public. Contention uses court 1 at 2026-11-01 10:00 MYT, one hour. Each round begins empty. No ordinary public tables are reset. Two adjacent court-1 bookings remain for read benchmarks (15 offered slots, 2 occupied).
- Python standard-library urllib threaded client on the **same machine**, 12-second client timeout, HTTP requests without a reusable connection pool. TCP/client overhead and Python scheduling are included. This is closed-loop load, not a fixed arrival-rate test; tails can be understated by coordinated omission. No warmup or repeated read runs; warm OS/database caches and pool growth affect comparisons.
- Read benchmarks use one API process (pool max 8), concurrency 1/10/50/100, three-second request-start windows, wait for all responses. The process/pool persists across successive read runs. Latencies are wall-clock request completion, percentiles use sorted nearest-lower index `int((n-1)*p)`, not histogram interpolation.
- Contention uses two independent native API processes (ports 18091/18092, pools max 8 each), 100 Python threads released from one barrier, 50 requests directed to each process. They are simultaneous attempts, not a guarantee of identical packet arrival times. Processes restart between rounds so the real 120/minute per-process limiter does not mask DB contention; this does not bypass code in either process. Rounds run sequentially in one isolated schema, truncated between rounds. Audit cleanup drops that schema after stopping both processes.

Exact harness and raw measurements: [local-audit.py](audit-evidence/local-audit.py), [local-audit-results.json](audit-evidence/local-audit-results.json). The harness uses the documented local development-only credentials and fixed project path; review before executing. It never reads cloud environment files. Requires Python 3, Go, existing local Podman/PostgreSQL, and free local ports. To reproduce from repository root:

```bash
(cd backend && go build -o /tmp/courtflow-audit-api ./cmd/server)
mkdir -p /tmp/courtflow-audit
cp buildDocs/buildDocs/CourtFlow/audit-evidence/local-audit.py /tmp/courtflow-audit/run.py
python3 /tmp/courtflow-audit/run.py > /tmp/courtflow-audit/run.log 2>&1
```

Network/container access may need host permissions. Start with `make dev-db` if local PostgreSQL is stopped. The harness writes final JSON to `/tmp/courtflow-audit/results.json`. Expect workload-dependent failures: it records statuses and asserts no more than one persisted row, rather than aborting at the first violation of the stricter 1-created/99-conflict expectation. The report evaluates that stricter expectation separately. First diagnostic runs aborted early at unexpected 503s; their measurements are excluded from the full 100-round dataset.

## Read endpoint measurements

Latency in milliseconds. CPU is API process CPU time divided by elapsed wall time, **100% = one logical core**, so values above 100% are possible. RSS is sampled peak resident KiB, not Go heap size. Unexpected is any status other than 200 or a client transport error; no expected conflicts occur on GET.

| Endpoint | Concurrency | Requests | Seconds | req/s | Avg | p50 | p95 | p99 | Unexpected | API CPU % | Peak RSS KiB | Max sampled DB sessions / active |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| facilities | 1 | 5578 | 3.001 | 1859.0 | 0.53 | 0.51 | 0.68 | 0.82 | 0 | 43.7 | 18948 | 1 / 1 |
| facilities | 10 | 8845 | 3.003 | 2945.6 | 3.38 | 3.22 | 5.48 | 6.84 | 0 | 73.3 | 19620 | 8 / 0 |
| facilities | 50 | 8902 | 3.011 | 2956.4 | 16.59 | 15.71 | 28.59 | 35.50 | 0 | 71.1 | 20536 | 8 / 1 |
| facilities | 100 | 8973 | 3.019 | 2971.7 | 31.66 | 30.38 | 55.78 | 69.52 | 0 | 71.5 | 20500 | 8 / 0 |
| availability | 1 | 3605 | 3.001 | 1201.4 | 0.83 | 0.78 | 1.12 | 1.46 | 0 | 58.0 | 20908 | 8 / 0 |
| availability | 10 | 7605 | 3.004 | 2531.7 | 3.94 | 3.78 | 6.11 | 7.47 | 0 | 165.1 | 20736 | 8 / 2 |
| availability | 50 | 7774 | 3.013 | 2579.8 | 18.99 | 18.06 | 32.98 | 40.60 | 0 | 164.3 | 20856 | 8 / 2 |
| availability | 100 | 7771 | 3.023 | 2570.9 | 35.93 | 34.19 | 65.48 | 81.72 | 0 | 164.1 | 20936 | 8 / 1 |

Endpoint paths: `/api/v1/facilities` and `/api/v1/courts/1/availability?date=2026-11-01`. All read runs had 0% unexpected failures. Increasing concurrency past 10 did not provide proportional throughput; this setup cannot distinguish Python client, connection setup, API and database bottlenecks. Do not call the highest observed RPS the server's maximum.

## Contention results — strict expectation FAILED

100 rounds × 100 attempts = 10,000 attempts. 73/100 rounds met exactly one 201, 99 conflicts and one persisted row. **No round persisted more than one row.**

- HTTP 201: **75** (0.75%).
- Expected HTTP 409 conflicts: **7228** (72.28%).
- Unexpected HTTP 503: **2697** (26.97%). No other statuses/transport failures in the recorded full run.
- 73 rounds: 1 created, 99 conflicts, 1 stored row. One round: 1 created, 1 conflict, 98 unavailable, 1 row. One round: 1 created, 99 unavailable, 1 row.
- 25 rounds returned 100 unavailable responses: **14 nevertheless persisted one booking**, while 11 persisted none. Thus 89 total committed rows existed across rounds before fixture truncation, but only 75 created responses were received. This is verified response/commit ambiguity; no specific client can identify its result from those 503s.

Aggregate across all response classes: 141.099 summed round request-wall seconds, 70.87 attempts/s; average 1366.99 ms; p50 56.81, p95 5033.96, p99 5039.30 ms. These sums exclude process restarts, fixture setup, row verification and cleanup; they are not continuous throughput or a successful-booking capacity metric. Contention runs were bounded bursts, not a duration-based arrival-rate benchmark.

PostgreSQL logs captured `ERROR: canceling statement due to user request` with `CONTEXT: while checking exclusion constraint ... in relation "bookings"`. Source imposes five-second contexts; SQL errors other than 23P01 become 503 without operator causes. The cancellation/wait evidence supports a contention/deadline problem; it does not prove a particular deadlock SQLSTATE or why a winning commit was not acknowledged. Diagnose before choosing retries, serialization or a different write strategy. Do not weaken the exclusion invariant.

Database safety passed; **the requested deterministic one-success/all-others-conflict contract did not**. Race-detector success does not contradict this database/service-level failure.

## Resource and database-pool observations

API CPU measured from `/proc/<pid>/stat` user+system tick deltas (SC_CLK_TCK); RSS from `/proc/<pid>/status` sampled while requests ran. PostgreSQL sessions/active were sampled via `pg_stat_activity` through Podman roughly every 200 ms plus query/process overhead; monitoring connection excluded. Sampling is coarse and can miss transient peaks. Other audit API processes were stopped for read runs. Session count grew from 1 at initial concurrency 1 to the configured ceiling of 8 at higher concurrency and remained warm; active queries were often missed because individual queries were short. This is evidence of connection count, not pgx acquisition latency.

**NOT VERIFIED:** pgx Stat AcquireDuration/EmptyAcquireCount/canceled-acquire counts, pool wait percentiles, contention-phase CPU/RSS, PostgreSQL CPU/RSS, host-wide CPU under the full workload, long-run memory growth, garbage collection profile, saturation point, successful-write throughput over a broad dataset, multi-instance read scaling, long-run p99 and restart recovery performance. No metrics endpoint or benchmark instrumentation exists in the unmodified application. API CPU/RSS and database session samples above fulfill only the observed portions of resource/pool behavior.

## Additional local probes

- Direct SQL partial overlap (10:30 against 10:00), same court: rejected by exclusion constraint. Adjacent 11:00 start: accepted. Different-court insert followed by conflicting insert in explicit BEGIN: error; earlier insert rolled back, verified zero court-2 rows. This tests DB semantics, not a product multi-slot endpoint.
- 120 malformed booking bodies returned 400; the next legitimate body returned 429. Demonstrates anonymous global-cap starvation before slot validation. It does not demonstrate a per-user fairness policy.
- Temporary schema cleanup confirmed after completion; local `public.bookings` remained at its initial count 0. Both audit API processes stopped. Existing PostgreSQL service kept running; no shared DB restart.
- Existing Go unit/integration/race/vet/build, frontend lint/typecheck/production build and five mocked Playwright tests passed as detailed in PORTFOLIO_READINESS_AUDIT.md. No image rebuild, cloud checks, production load test or new product code.

## Interpretation

Read latencies are useful as a reproducible tiny-data local reference. They do not overcome the failed booking contention contract. Fix and instrument that behavior, rerun the same persisted-state test, then collect longer warmed measurements with a client that can independently establish the load generator is not the bottleneck. Compare only like-for-like local runs. No local-to-production performance comparison is justified.
