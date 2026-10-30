#!/usr/bin/env bash
set -euo pipefail
: "${TEST_DATABASE_URL:?Set a loopback TEST_DATABASE_URL; integration tests are mandatory} "
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

if [[ ! -f "$ROOT/backend/go.mod" ]]; then echo "ERROR: backend/go.mod missing" >&2; exit 1; fi
if [[ ! -f "$ROOT/frontend/package.json" ]]; then echo "ERROR: frontend/package.json missing" >&2; exit 1; fi

(
  cd "$ROOT/backend"
  echo '[Go] gofmt check'
  bad=$(gofmt -l .)
  [[ -z "$bad" ]] || { printf 'Unformatted Go files:\n%s\n' "$bad" >&2; exit 1; }
  echo '[Go] vet'
  go vet ./...
  echo '[Go] tests'
  go test ./...
  echo '[Go] race detector'
  go test -race ./...
  echo '[Go] compilation'
  go build ./...
)
(
  cd "$ROOT/frontend"
  echo '[Next.js] dependencies'
  npm ci --no-audit --no-fund
  echo '[Next.js] lint (if configured)'
  npm run lint
  echo '[Next.js] typecheck (if configured)'
  npm run typecheck
  echo '[Next.js] production build'
  npm run build
)
echo "Local application checks passed; see separate persistence and container checks."
