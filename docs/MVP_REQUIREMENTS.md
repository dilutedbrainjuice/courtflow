# CourtFlow V1 — product and acceptance specification

## Target operator
One demo badminton facility with 4 named courts. People can see availability, make one-hour reservations and receive a simulated confirmation. Reception can view current bookings, but write/admin actions must be protected.

## Why this product
Small Malaysian sports facilities frequently receive bookings through phone/WhatsApp and need a consistent single view of availability that prevents conflicts between customers and front desk staff. The V1 focuses on this concrete problem rather than a large marketplace.

## User paths
1. Browse venue → choose court → choose local date and time → review server-calculated price → confirm a **demo booking** → show confirmation.
2. Choose the same court and overlapping time from another tab → backend rejects conflict with HTTP 409, refreshes availability, no double booking.
3. View bookings through a deliberately limited demo operations screen; no unsafe unauthenticated delete/change controls.

## Data assumptions
- Venue time zone: `Asia/Kuala_Lumpur`.
- Initial venue: fictional; courts A–D; display RM25 per hour (configurable in database).
- Duration for V1: one hour, on the hour, operating hours configurable (e.g. 08:00–23:00 local time), finite booking horizon.
- Payment: simulated only; do not store real payment/identity data.
- Bookings initially `confirmed` or `cancelled` only. No pending payment or expiring seat holds until a future feature.
- Model time intervals as half-open `[start,end)` and store UTC instants (`timestamptz`). Consecutive 18:00–19:00 and 19:00–20:00 bookings do not overlap.

## Must-have API
- GET `/healthz` → 200 service health; no secrets.
- GET `/api/v1/facilities` → array of demo facilities.
- GET `/api/v1/facilities/{facility_id}/courts` → court list + attributes.
- GET `/api/v1/courts/{court_id}/availability?date=YYYY-MM-DD` → available and occupied time slots calculated from authoritative DB.
- POST `/api/v1/bookings` → create confirmed demo booking atomically; reject invalid/out-of-hours/occupied requests.

## Correctness tests
- 20 simultaneous attempts for the same open court/time ⇒ exactly one success, all others conflict, database contains exactly one confirmed booking.
- Two adjacent slots both succeed.
- Same time different courts both succeed.
- Malformed timestamps, incorrect prices, closed-hours bookings, out-of-horizon dates rejected.
- Data persists across a restart of Go process and Postgres container (named volume).
- Frontend renders pending, success, invalid, conflict, and network-unavailable states.

## Non-goals V1
No real payments, paid SaaS multi-tenancy, login/registration, SMS, WebSockets, Redis, Kafka, Kubernetes, promotions, QR access control or advanced recurring bookings. Add later only with an operational reason.

## Public demo safeguards
Use fictional non-sensitive data, public demo label, minimal data retention, server-side validation, request/body limits, timeouts, strict origin policy, and protection from unlimited anonymous booking spam. Do not expose production admin mutation APIs unauthenticated.
