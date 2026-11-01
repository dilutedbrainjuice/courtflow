# CourtFlow backend interview readiness audit

> Follow-up (2026-11-01): RELIABILITY-010 is implemented and locally verified.
> See [RELIABILITY_RESULTS.md](RELIABILITY_RESULTS.md). The original analysis below is preserved as historical evidence.

Date: 2026-10-31 (Asia/Kuala_Lumpur). Audited local HEAD `77fa3a9`. Audit only: application source unchanged, Chisle never enabled, no production requests, cloud actions, deployment, push or commit. Reports live in the existing Obsidian CourtFlow directory, as specified by DEVELOPMENT.md.

## Executive verdict

**64/100 — CONDITIONAL GO for a junior Go backend interview.** A small, credible booking demonstration with a substantive database invariant; insufficient evidence of breadth or operational maturity for an impressive mid-level Go portfolio by itself. Go supports concurrent HTTP/database work idiomatically, but the strongest engineering is PostgreSQL consistency, not advanced Go concurrency. Do not sell this as a complete facility-management platform or a high-throughput production service.

The standard 24-contender integration test passed, including persisted row verification. The stronger two-process, 100-contender exercise exposed unexpected 503s despite retaining overlap safety. Only 73/100 rounds met the expected response contract: 2,697/10,000 attempts returned 503. In 14 rounds all responses were 503 despite one committed booking. See PERFORMANCE_BASELINE.md for measurements and denominators. Correctness and successful service under contention are separate claims.

## Scorecard

| Category | Score | Evidence and reason points were withheld |
|---|---:|---|
| Business functionality | 9/15 | Five endpoints and a working one-hour simulated booking journey (`backend/internal/api/api.go:45–51`, `booking.go:37–86`, `frontend/app/page.tsx:92–115`). No cancellation, owned bookings, operations screen, multi-slot transactions, identity or payment integration. Intentional demo scope still limits portfolio depth. |
| Concurrency and consistency | 18/25 | GiST exclusion, FK/check constraints and server price in atomic INSERT (`backend/migrations/001_initial.sql:11–20`, `booking.go:67–82`). Race/integration checks pass; two-process audit persists at most one row per round. Unexpected contention failures, no idempotency or multi-slot behavior, and no recovery for ambiguous outcomes prevent full marks. |
| Go engineering | 10/15 | Cohesive small package, context-aware pgx, bounded pool, standard HTTP and explicit cleanup (`backend/cmd/server/main.go:29–40,53–74`, `api.go:69–73`). Shutdown is not joined; database errors are discarded from operator logs, weak response typing and concrete pool dependency limit diagnostics/test substitution. |
| Testing and performance | 9/15 | Go unit/integration/race/vet/build and frontend checks actually executed. Existing tests validate 24 competitors and stored rows (`api_test.go:88–114`); audit adds stronger external evidence. Only two Go test functions, mocked browser API, no CI, sustained write-load/recovery tests or pool-wait telemetry. 100-contender expectation fails in some rounds. |
| Security | 6/10 | Parameterized SQL, strict 1 KiB JSON, no PII and exact origin matching (`booking.go:33–47,67–72`, `api.go:55–67`). No protected identity/ownership flows to demonstrate; global process cap enables starvation and scales per instance. No dependency vulnerability scan or history-wide secret scan performed. |
| Deployment and operations | 6/10 | Non-root multistage image (`backend/Dockerfile:1–12`), manual deployment configuration (`render.yaml:1–16`, `frontend/vercel.json:1`), reproducible local tools. No joined drain, readiness/DB health, request/error telemetry, migration ledger, tested backup restore or CI. Current provider state not inspected. |
| Interview demonstration | 6/10 | Clear UI, real DB invariant and reproducible audit evidence. Good discussion of tradeoffs. Contention failures, superficial scope, stale documentation sections and lack of live real-backend browser test weaken confidence. |
| **Total** | **64/100** | Missing evidence is not credited. |

Paths above are relative to the repository root; line ranges identify the reviewed implementation rather than generated build artifacts.

## Blockers and severity

1. **High — contention reliability, showcase blocker:** concurrent valid reservations can return 503 instead of one 201 and 99 expected 409s. In 14 rounds a booking was persisted although all 100 responses were unavailable. Some failures occur around the five-second request deadline (`api.go:71`, `booking.go:80–82`). PostgreSQL logs contain canceled exclusion-constraint checks. The audit does not establish a definitive root cause or a retry strategy. Do not advertise flawless high-contention booking. DB integrity survives; this is not evidence of double-booking.
2. **High — shutdown drain is incomplete:** the signal goroutine calls `srv.Shutdown`, which closes listeners; `ListenAndServe` returns and `main` can then close the pool and exit without waiting for Shutdown (`main.go:62–74`). The ten-second context does not guarantee ten seconds of drain. Static control-flow finding; no controlled in-flight shutdown reproduction executed.
3. **Medium — one anonymous visitor can starve booking attempts:** all attempts, including malformed JSON, consume the same 120/minute counter (`booking.go:17–31`). CORS does not stop scripts omitting Origin. The local probe sends 120 invalid requests and then a legitimate one; see baseline for status results. Multiple processes each have independent limits.

**Critical security/integrity blockers:** no double-booking, cross-user data disclosure, unauthorized privileged mutation or data loss was demonstrated in the exercised demo surface. This is a bounded observation, not a security certification. There is no authentication or ownership model; adding private data or booking-management mutations without authorization would create a critical blocker. The missing features must not be presented as implemented security.

## Verified features and actual API flows

| Endpoint | Implemented flow | Evidence |
|---|---|---|
| GET /healthz | Static liveness JSON, no DB ping | `api.go:45–47` |
| GET /api/v1/facilities | Ordered public names/timezones | `catalog.go:9–35` |
| GET /api/v1/facilities/{id}/courts | Positive integer ID; parameterized catalog query; unknown venue returns [] | `catalog.go:40–72` |
| GET /api/v1/courts/{id}/availability?date=YYYY-MM-DD | Lookup court + venue rules; local-calendar horizon; confirmed occupancy; 15 hourly sessions for seeded 08:00–23:00 venue | `availability.go:14–83`, `slots.go:17–41` |
| POST /api/v1/bookings | Attempt cap → strict JSON → court lookup → whole-hour/future/opening/horizon validation → INSERT SELECT price → UUID/amount response; 23P01 becomes 409 | `booking.go:16–86` |

All routes share exact CORS, no-store/nosniff and five-second request context middleware. Public availability is a snapshot; it confers no reservation. Server calculations supply one-hour end time and integer-sen price. No user personal data is collected. Frontend uses bounded fetch, loading/error/pending states, MYT formatting and availability refresh (`page.tsx:11–38,56–115`). It hardcodes venue ID 1 and MYT; a generalized multi-facility UI is not implemented. Cancellation of fetches exists, but callbacks do not all independently guard stale fulfilled responses.

**Simulated/deferred:** payments are a response label, not a payment provider or payment state machine. Bookings are immediately confirmed; no hold, expiration worker, refund, cancellation endpoint, owner lookup, authentication, RBAC or reception list. Database status permits cancelled but no exposed transition implements cancellation. No mocks in the Go database path; browser tests intentionally mock HTTP responses.

## Where consistency is guaranteed

- `book` does not lock availability in Go. Its mutex protects the attempt counter only (`api.go:21–24`, `booking.go:20–27`). Multiple API processes do not share that mutex.
- The INSERT reads the current court price and creates one row in one PostgreSQL statement (`booking.go:67–72`). PostgreSQL statement atomicity/implicit transaction applies; there is no application BEGIN/COMMIT around multiple operations. The earlier court-rule read is a separate statement, so hypothetical concurrent schedule edits would not share the INSERT snapshot. No schedule mutation API currently exists.
- The database rejects confirmed intervals for the same court where `[start_at,end_at)` overlaps (`001_initial.sql:18–20`). Adjacent intervals and different courts remain valid. This applies to all connections and independent processes, including direct SQL. It also protects off-hour overlaps even though HTTP accepts only hour-aligned starts.
- FK, positive amount/price, one-hour duration, status and opening-schedule checks are DB-enforced. Whole-hour starts, future/horizon and venue opening eligibility are application rules, not DB constraints. A privileged direct SQL writer can bypass those business rules but not confirmed overlap.
- Audit direct SQL checks reject a 30-minute-offset overlap, accept adjacency, and prove an explicit test transaction rolls back its earlier court-2 insert when a later conflicting insert fails. This demonstrates PostgreSQL rollback, **not an implemented atomic multi-slot API**.
- No idempotency key, replayed original response, booking ownership or expiration exists. Retrying an already committed request returns conflict; if the first response is lost, the user cannot retrieve an owned result. UUID uniqueness is not request idempotency. A Scan/write/network failure after a commit can create an ambiguous outcome, not automatic rollback.
- Two processes share PostgreSQL integrity. They also double the configured pool ceiling (8 each) and the attempt allowance. Under stronger contention, safety held but the expected successful/conflict response contract was not reliable.

## Executed checks and evidence limits

Executed locally with the development-only loopback DB URL, not any secret environment file:

```bash
cd backend
TEST_DATABASE_URL='postgres://courtflow:courtflow_dev_only@127.0.0.1:5433/courtflow?sslmode=disable' go test -count=1 -v ./...
TEST_DATABASE_URL='postgres://courtflow:courtflow_dev_only@127.0.0.1:5433/courtflow?sslmode=disable' go test -count=1 -race ./...
go vet ./...
go build -o /tmp/courtflow-audit-api ./cmd/server
cd ../frontend
npm run lint
npm run typecheck
npm run build
WATCHPACK_POLLING=true npm test
```

- **PASS:** TestSlotValidation, TestIntegration (real PostgreSQL, temporary schema, migration twice, 24 competitors → 1 created/23 conflicts/1 row, adjacent/different courts, invalid requests, CORS rejection and rate cap). No integration skip; `-count=1` avoids cached results. Race passed; vet and compilation passed.
- **PASS:** frontend lint and standalone typecheck. Initial sandbox build failed parsing TypeScript --showConfig; identical host-access retry passed compilation, TypeScript and static generation. Reported build success refers to that retry.
- **PASS:** all five existing Playwright tests, 10.3 seconds using polling. Default watcher run encountered EMFILE and was stopped; polling retry passed. Tests mock backend responses and cannot prove end-to-end persistence. Generated next-env.d.ts changes were restored to the previously clean version.
- **FAIL:** strict all-round 100-contender expectation. Full counts, distributions, dataset, resource measurements and harness are in the baseline/evidence directory. Persisted state was queried after every round, not inferred from responses.
- **PASS:** direct SQL partial overlap, adjacency and explicit transaction rollback checks, completed and recorded in the final baseline.
- **NOT VERIFIED in this audit:** restart durability (existing script inspected but restarting shared local PostgreSQL during isolated tests was avoided), container image rebuild/smoke, actual provider settings, production health, authenticated/cross-user flows (absent), browser against real API, vulnerability/adversarial fuzz scan, backup recovery and graceful in-flight shutdown. Prior HANDOFF receipts are historical evidence, not fresh passes.

Existing automated scope is narrow: five fixed-clock slot cases and one broad integration function; catalog tests largely assert status, not full response contracts. TestIntegration skips without TEST_DATABASE_URL (`api_test.go:32–36`); `make test` supplies it, but bare go test may misleadingly look green. No .github CI workflow found. Integration schema cleanup is deferred; extension creation can live outside schema isolation depending on prior installation. A disposable local database is preferable for migration portability checks.

## Security review

- **Authentication/RBAC/ownership:** absent by design. All facilities/courts/occupancy and booking creation are public. No booking-detail listing or mutation route exists, so no demonstrated IDOR path into private bookings. Cross-facility read/create is public behavior; tenant isolation is not claimed or tested.
- **Validation:** positive IDs, strict unknown-field rejection, one bounded JSON object, hour/time window/opening checks, server-owned price. Content-Type comparison rejects otherwise common `application/json; charset=utf-8` (`booking.go:33`); robustness issue, not an exploit. GET response/catalog size lacks explicit pagination, acceptable only for this tiny seed.
- **SQL injection:** user values bound as `$1` etc.; no user-controlled SQL concatenation in runtime handlers. Test schema concatenation uses generated numeric suffixes, not client input.
- **CORS:** exact allowlist, no wildcard, no credentialed sessions, no-Origin clients accepted (`api.go:55–67`). It is browser policy, not API authentication or anti-abuse. OPTIONS is broad but does not grant privileged routes.
- **Abuse:** global process cap has collateral denial-of-service effects; GET endpoints have no rate cap. Request body/time/header limits reduce individual request cost but do not solve distributed abuse or booking inventory exhaustion.
- **Secrets:** tracked environment examples contain development placeholders/settings; gitignore excludes local env/key files. No actual production credential was needed/read for this audit. No exhaustive historical secret scan performed. TLS behavior depends on operator-supplied DATABASE_URL; server does not mandate TLS for remote DB hosts.
- **Observability:** errors are safely generic externally but causes are not logged internally. Startup uses slog only. No request ID, latency/status logs, audit trail or DB-pool metrics. Raw user data/DSNs must remain out of future logs.

## Go-specific assessment

Good: small modular monolith, net/http method patterns, practical direct SQL instead of ceremonial service/repository layers, errors.Is for no rows and errors.As for PostgreSQL SQLSTATE, context propagation, defer rows.Close plus rows.Err, integer money, bounded pgx pool, transport/database timeouts, embedded tzdata and mutex around shared mutable counters. No unnecessary goroutine-per-operation pattern. Letting PostgreSQL arbitrate distributed writes is a meaningful engineering choice.

Weak: shutdown goroutine not joined; returned errors lose useful operator context; JSON helper ignores write/encoding errors; maps/any make contracts less explicit than response structs; clock and pool are concrete dependencies, limiting isolated failure tests. A consumer interface is justified only if real fault substitution is needed. Health does not verify startup DB connectivity; pool construction is lazy. Pool close can interact with active requests during premature shutdown. There are no background workers to supervise; lack of extra goroutines is not a defect.

Go is used meaningfully for simultaneous I/O and bounded request execution, but most endpoints remain small catalog/booking handlers. The project does not demonstrate channels/pipelines, complex goroutine coordination or Go-specific performance optimization; none is required to make this design sound. Be ready to explain scheduling and race-detector limits without claiming PostgreSQL guarantees come from Go.

## Interview questions and answer readiness

| Question specific to this code | Existing evidence supports a convincing answer? / What to learn |
|---|---|
| 1. Why can two API processes not book the same court interval? | **Yes for safety:** exclusion constraint, independent process test and stored row checks. Explain GiST, predicate and MVCC/locking; do not credit the Go mutex. |
| 2. Why does 10–11 not block 11–12? What about 10:30–11:30? | **Yes:** half-open range and direct SQL overlap/adjacency tests. Distinguish API alignment from DB overlap. |
| 3. Why did stronger contention yield 503, and would you retry? | **Partial:** reproducible failure/deadline evidence; root cause/retry semantics unresolved. Explain waits, cancellation, deadlocks vs exclusion violation, and ambiguous commit before proposing a fix. |
| 4. Which operations are transactional and can I book three hours atomically? | **Yes if honest:** one INSERT SELECT is atomic; no multi-slot API. Explain explicit transaction rollback and why three HTTP calls are not atomic. |
| 5. What happens if INSERT commits but the client never gets the UUID? | **Partial:** known limitation documented in code. No idempotent replay/owner retrieval; explain keys, unique scope and transactionally stored outcomes. |
| 6. Why is pool max 8, and what happens with ten replicas? | **Partial:** bounded source configuration; observed connection samples in baseline. Learn pool acquisition waits, DB budgets, replica multiplication and cancellation. |
| 7. Does SIGTERM wait for in-flight requests? | **No reliable guarantee:** main exits without joining Shutdown. Explain listener closure vs handler completion and pool lifetime. |
| 8. How do users access only their own bookings? | **No implementation:** anonymous non-PII demo, no management endpoints. Explain authentication vs authorization, tenant/object checks before extending scope. |
| 9. How are MYT dates, money and the last booking day handled? | **Yes within fixed venue scope:** timestamptz, integer sen and local calendar vs exact instant horizon. Learn DST/general-zone limitations; frontend hardcodes MYT. |
| 10. What do the tests prove, and what would you monitor first? | **Partial:** actual race/integration/browser/benchmark evidence. Race detector cannot establish SQL correctness; mocked browser cannot prove DB flow. Add error causes, statuses, latency and pool wait metrics. |

Deep understanding required: SQLSTATE 23P01 vs other database errors; predicate range constraints; implicit transactions and uncertain outcomes; context deadline and connection/pool behavior; HTTP shutdown lifecycle; calendar/date vs timestamp semantics; rate limiting across processes; boundaries of public-demo security. Memorizing endpoint names is insufficient.

Strongest discussion topics: DB-enforced overlap across instances; atomic server pricing; half-open intervals; bounded contexts/pool; intentionally small monolith; reproducible isolated data. Questionable decisions: shared anonymous attempt budget; incomplete shutdown; low observability; weak response contracts; hardcoded facility/timezone; migration repeatability without migration tracking; historical append-only status text that leaves stale architecture and HANDOFF claims easy to misread. Current architecture doc begins “implementation has not started” despite appended V1 sections; use source and this audit as current evidence.

## Three-minute local demonstration

- **0:00–0:35:** introduce fictional/no-payment scope. Show local UI, four courts, MYT and server-owned RM25/hour price. Select tomorrow's free session; confirm and refresh.
- **0:35–1:10:** show booking handler INSERT SELECT and the exclusion constraint. Explain why UI availability and Go mutex cannot establish reservation correctness.
- **1:10–1:55:** run existing isolated integration test with the loopback TEST_DATABASE_URL and `-count=1 -v`. Show 24-contender assertion and stored count, then the saved 100-round two-process report. Explicitly disclose 503 failures rather than presenting it as a flawless live stress demo.
- **1:55–2:30:** show direct SQL adjacent/partial-overlap and rollback evidence. Explain one-row atomicity vs unsupported multi-slot transaction and lost-response idempotency limitation.
- **2:30–3:00:** show race/vet/build receipts and baseline caveats. Name the shutdown and contention improvements, and explain what would change before adding user data. Do not issue load tests against the public site.

Prepare local PostgreSQL/dependencies first; run the UI/API locally in separate terminals. Use the saved isolated harness for audit evidence, not a destructive reset of ordinary demo tables. A real UI confirmation is a suggested live step, not a claim that browser-to-real-API was reverified in this audit.

## Three resume bullets using verified achievements

- Built a Go net/http and PostgreSQL badminton booking demo with five API endpoints, server-calculated integer-sen pricing, venue-time validation and a Next.js customer interface.
- Enforced confirmed reservation overlap prevention with a PostgreSQL GiST exclusion constraint; verified at most one stored reservation per round in a 100-round, two-process contention audit with 100 simultaneous attempts per round, while identifying response-reliability failures under contention.
- Added isolated-schema PostgreSQL integration tests and exercised Go race/vet/build checks plus five mocked Playwright booking/error/mobile tests; documented local endpoint latency and the limits of those measurements.

Do not claim a throughput improvement, guaranteed 100-user success, production-scale reliability, secure multi-tenant access, real payments or performance parity with cloud hosting.

## Top five improvements by interview value versus effort

| Rank | Improvement | Interview value / estimated effort | Observable completion condition |
|---|---|---|---|
| 1 | Diagnose and resolve contention failures; retain a repeatable two-process regression with persisted counts | Very high / medium | 100 rounds each yield exactly 1 created, 99 expected conflicts, 1 row, no unexpected errors; log specific causes and verify timeouts. Do not hide failures by loosening assertions. |
| 2 | Join HTTP shutdown before pool close; test an in-flight request under SIGTERM | High / low | Process remains until handler drain or explicit deadline and reports shutdown failures. |
| 3 | Add request IDs, useful error causes and pool-wait/latency/status telemetry | High / low–medium | A failed reservation can be traced without DSNs/PII; benchmarks report pool acquisition behavior. |
| 4 | Implement narrowly scoped request idempotency with durable replay semantics | High / medium | Concurrent/retried identical keys replay the original result; mismatched payload rejected; response-loss test. Requires a separate authorized implementation. |
| 5 | Make checks a repeatable CI/local acceptance gate and add real-API UI + failure tests | High / medium | Integration cannot silently skip in gate; real browser-to-local-DB success/conflict; shutdown/deadline/constraint cases; more than status-only catalog assertions. |

For public demo abuse, revise the shared limiter before increasing exposure; protect all future management routes with identity and object authorization. These do not require inventing payment integrations, microservices or background jobs for interview appeal.

## Final recommendation and smallest credible improvement set

**CONDITIONAL GO.** Present now as a small junior-level demo only with explicit limits. Before describing it as an impressive Go backend showcase, fix the contention response failures and shutdown lifecycle, then rerun a reproducible isolated two-process gate with enough error/pool diagnostics to explain every result. Keep the architecture small. Idempotency is the next strongest feature, but it is not a substitute for fixing the observed reliability gaps. No application changes were made by this assessment.
