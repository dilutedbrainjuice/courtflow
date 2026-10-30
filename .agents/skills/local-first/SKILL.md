---
name: local-first
description: Audit or run the CourtFlow local development, database, tests and build workflow while avoiding Vercel, Render and Neon usage. Use for tests, debugging, setup and build checks.
---

# CourtFlow local-first execution

Never run tests or schema experiments against Neon production. Use Podman PostgreSQL on localhost and the local Go/Next.js processes. Check existing scripts before installing anything. Run narrowly scoped checks while iterating and full local checks before marking complete. Go: `go test ./...`, `go test -race ./...`, `go vet ./...`, build server. Next.js: npm lint and build, TS checks. Use seeded fake data only. Include a race test for concurrent attempts to reserve same court/time and assert exactly one success (other requests conflict). Record which checks actually ran; report failures and skipped checks, not fictional passes. Do not deploy or add cloud infrastructure.
