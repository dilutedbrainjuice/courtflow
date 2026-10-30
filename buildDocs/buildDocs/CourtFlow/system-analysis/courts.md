# Courts flow

## Business purpose
Show the venue courts and their configured hourly prices.

## API endpoint or triggering event
`GET /api/v1/facilities/{id}/courts`. Public fictional demo; no authentication or private user records.

## Request and response
Positive integer facility path ID. Returns 200 array of `{id,name,price_sen}`. Unknown numeric facility returns an empty array.

```bash
curl http://localhost:8080/api/v1/facilities/1/courts
```

## Complete ordered function call flow
1. `cmd/server/main.go:main` creates a bounded eight-connection pgx pool and passes it to `API`.
2. `internal/api/api.go:Handler` applies no-store/nosniff headers, validates a supplied Origin against the configured allowlist, answers preflight, and attaches a five-second context deadline.
3. Route-specific chain: `Handler → courts → strconv.Atoi → pgxpool.Query → rows.Scan → JSON`.
4. `JSON` writes the response, or `fail` wraps a stable public error string as `{"error":"..."}`. Database details are not returned.

## Functions, parameters and returns
courts validates the path identifier, reads matching courts, checks scan/iteration errors and writes the array. Prices are integer sen, avoiding floating-point money calculations.
All database operations receive the request context. Handler functions return through HTTP rather than Go return values. `fail(writer,status,message)` and `JSON(writer,status,value)` are shared response helpers.

## Business rules and validation
The seeded facility is CourtFlow Demo Badminton with courts A–D at RM25/hour, opening 08:00–23:00 MYT. Time is stored as timestamptz. One-hour intervals are half-open. Booking starts must be strictly in the future and no more than 30 rolling days ahead. These constraints are checked for writes, and availability uses the same slot predicate.

## Database queries and transactions
SELECT id,name,price_sen FROM courts WHERE facility_id=$1 ORDER BY id. Parameter binding prevents SQL injection. One read statement.

## Error handling and edge cases
Malformed/nonpositive ID returns 400. Query/iteration failure returns 503; scan failure returns 500.
Origin policy is browser-origin filtering, not authentication: clients without Origin remain allowed. Unknown routes/methods use Go ServeMux defaults, which are plain-text errors. A request deadline bounds SQL work; server read/write timeouts bound transport work.

## Concurrency and consistency considerations
Reads do not reserve slots. A customer must submit a booking to establish a reservation; the database exclusion constraint remains authoritative. Concurrent writes may change the next read.

## Mermaid sequence diagram
```mermaid
sequenceDiagram
    participant C as Client
    participant M as Handler middleware
    participant H as courts handler
    participant D as PostgreSQL
    C->>M: GET /api/v1/facilities/{id}/courts
    M->>H: Valid origin + bounded context
    H->>D: Parameterized SQL
    D-->>H: Rows or error
    H-->>C: JSON response or public error
```

## Relevant source-file references
- `backend/cmd/server/main.go`
- `backend/internal/api/api.go`
- `backend/migrations/001_initial.sql`
- `backend/internal/api/api_test.go`

## Tests validating the flow
`make test` runs Go race-enabled tests against loopback PostgreSQL. `TestIntegration` creates a unique schema, applies migration twice, checks all five routes, submits 24 concurrent identical bookings (one 201, 23 conflicts), verifies row count, adjacent/different-court success, invalid input rejection, CORS and rate limiting. `TestSlotValidation` checks temporal boundaries. Schema cleanup occurs after the test. These passed locally on 2026-10-31. No production requests were made.

## Limits and follow-up
The demo rate limit is global per process, resets on restart and can deny legitimate traffic during abuse. It is a basic bound, not distributed abuse prevention. There is no cancellation, identity, payment, private booking lookup or operations mutation API. No holds or expiry are advertised.
