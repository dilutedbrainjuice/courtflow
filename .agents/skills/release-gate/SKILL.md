---
name: release-gate
description: Prepare or review CourtFlow manual production deployment to Vercel, Render and Neon after local tests pass. Use only when user asks for a release or production readiness; never deploy without explicit approval.
---

# Manual release only

1. Verify clean Git state, commits, no secrets, local integration/race tests, frontend production build, Go production build and database migration compatibility.
2. Confirm Neon Free project/env credentials exist through owner action; never fabricate URLs or run tests on production DB. Back up schema/data as appropriate before migrations.
3. Render Free Go API: Docker image, start on `$PORT`, `/healthz`, manually deploy; auto-deploy OFF. Confirm Render will sleep after idle.
4. Vercel Hobby Next.js frontend with root directory `frontend`; `git.deploymentEnabled=false` to avoid Git-triggered deployments. Trigger manual deployment only with owner approval; correct public API origin.
5. Verify minimal health and one controlled end-to-end demo booking *after user approval*, not load tests. Check errors/logs briefly and document cold-start limitation.
6. Append actual deployment URL, commit SHA, dates, settings and observed checks to BUILD_LOG.md and HANDOFF.md. If access is blocked, write a step-by-step owner action list and stop, without claiming deployment.
