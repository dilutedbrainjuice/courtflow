#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
podman exec -i courtflow-postgres psql -v ON_ERROR_STOP=1 -U courtflow -d courtflow --single-transaction < backend/migrations/001_initial.sql
