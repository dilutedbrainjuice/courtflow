# CourtFlow production runbook — MANUAL RELEASE ONLY

**Status: first release in progress; see HANDOFF RELEASE-002 for provider IDs and blockers.** This is a public **non-commercial portfolio demo** using free plans. Local tests are the source of truth. Owner manually verifies the production build and smoke test.

## Stack
- GitHub: public monorepo (`/home/terasit/VSProjects/court`)
- Vercel Hobby: Next.js + TS app, project root `frontend/`, no Git-triggered deployment (`git.deploymentEnabled = false`)
- Render Free: Go API in `backend/`, multi-stage Dockerfile, `PORT` env, `/healthz`, `Auto-Deploy: Off`
- Neon Free PostgreSQL: schema migrated by approved one-time release step, credentials only in Render environment, TLS required

## Free plan caveats
- Vercel Hobby is personal/noncommercial only. A real paid facility service is not allowed on this plan.
- Render Free goes to sleep after ~15 min without traffic; first response may take ~1 min. No persistent local filesystem.
- Neon Free scales down after inactivity; respect monthly compute, storage and network allowances.
- Free services are not a promise of 24/7 responsiveness or production SLA. The public demo is noncommercial.

## Environment
Local (ignored): `frontend/.env.local` with `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080` and backend local `DATABASE_URL=postgres://courtflow:courtflow_dev_only@127.0.0.1:5433/courtflow?sslmode=disable`.

Production secrets/settings (never commit):
- Render `DATABASE_URL` = Neon PostgreSQL production URL (TLS required).
- Render `ALLOWED_ORIGINS` = exact Vercel site origin (no `*` for credentialed requests).
- Render `PORT` provided by Render; the Go server should read it and bind `0.0.0.0`.
- Vercel `NEXT_PUBLIC_API_BASE_URL` = exact Render HTTPS API origin. Public frontend env vars are not secrets.

## Preparing first release locally
1. Inspect `git status`, verify no secrets, commit reviewed code and all system-analysis docs/build log/handoff. Create repository remote only after user authorization.
2. Start local Postgres. Apply versioned SQL migrations. Seed only fictional venue/courts.
3. `go test ./...`, race and concurrent integration tests, `go vet ./...`, compile Go server. Lint/typecheck/build Next.js locally; fix every failure.
4. Build the same Go Docker image locally with Podman; run it locally with DB env and test `/healthz` and one booking. The Docker image must build from `backend/` with correct `go.mod` and source layout.
5. Validate local release config (Render and Vercel). No token-consuming production build yet.

## Manual first deployment (owner approval required)
1. Create Neon free project (once). Use a unique database/role for demo; keep connection string secret. Apply migration to Neon **once** after reviewing SQL (not an experimental test); seed fictional facility records. Never stress test Neon.
2. Create Render Free Web Service connected to the GitHub monorepo; select Docker runtime, `backend` root/directory; disable auto-deploy, set `DATABASE_URL` and `ALLOWED_ORIGINS`. Manually deploy the approved commit.
3. Verify one `GET /healthz` on the Render URL and inspect one log summary. Document cold start if observed.
4. Create Vercel Hobby project linked to same GitHub repo, root `frontend`. Ensure `frontend/vercel.json` disables Git-triggered deployments; configure `NEXT_PUBLIC_API_BASE_URL`; manually initiate first approved deployment (Vercel UI/CLI).
5. Owner checks landing page, loads availability and makes at most one test/demo booking. Verify server logs and DB only if necessary. No concurrent/load testing in production.
6. Append real URLs, commit hashes, migration version, UTC release time, smoke checks and known issues to BUILD_LOG/HANDOFF.

## Release policy afterward
- All development and tests on localhost. Push to GitHub without starting automatic deployments.
- On explicit release approval, review migrations and locally repeat tests, then manually deploy **backend before frontend** for backwards-compatible API/schema changes.
- Record and verify each manual release. No surprise provider MCP writes.
- Have a rollback plan for application code; DB schema rollback requires explicit operator review and backup strategy.

## V1 release artifacts and exact owner actions

Prepared locally: `backend/Dockerfile`, `render.yaml`, `frontend/vercel.json`,
backend/frontend `.env.example`, SQL migration and local verification scripts.
Render service auto-deploy is off; also turn Blueprint **Auto Sync to No** if using
Blueprints, because Blueprint synchronization is a separate control. Review before
initial import: resource creation itself can trigger a first deployment.
References: https://render.com/docs/blueprint-spec and
https://render.com/docs/infrastructure-as-code and
https://vercel.com/docs/project-configuration/git-configuration.

After approving publication/release:
1. Authenticate GitHub and create a public repository, then add its remote and push
   the reviewed local commits. No tokens belong in files or remote URLs.
2. Create the dedicated Neon demo database. Supply a TLS connection URL using
   `sslmode=verify-full` (or the provider's verified TLS configuration) to Render.
   Review `backend/migrations/001_initial.sql`, confirm `btree_gist` is supported,
   and execute the migration once using the SQL editor or a trusted migration
   client with transaction/error-stop behavior. Never point local test scripts at Neon.
3. Create Render service from `backend/Dockerfile`, configure DATABASE_URL,
   ALLOWED_ORIGINS and Auto-Deploy Off, then manually deploy only with approval.
4. Create Vercel project rooted at `frontend/`, confirm git.deploymentEnabled=false,
   set NEXT_PUBLIC_API_BASE_URL to the Render HTTPS origin and deploy manually.
   Set Render ALLOWED_ORIGINS to the exact final frontend origin (no trailing slash).
5. Owner checks health, availability and one demo booking, then records real URLs,
   approved commit and observed results. Do not run the concurrency suite remotely.

### Retention and rollback

No personal data is collected. Retain demo bookings for at most 30 days after their
end time. During an approved maintenance action, review and execute:
`DELETE FROM bookings WHERE end_at < now() - interval '30 days';`
There is no automatic cleanup worker or unauthenticated deletion API. Schedule
owner review monthly; do not run routine cloud experimentation. Future approved
cleanup automation may replace manual maintenance.

Rollback application code to the previous known-good image/commit with compatible
schema. Before any later schema migration, review backup/restore options and
backward compatibility. Do not drop tables or undo migrations automatically.
