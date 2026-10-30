#!/usr/bin/env bash
set -euo pipefail

NAME=courtflow-postgres
VOLUME=courtflow_pgdata
IMAGE=docker.io/library/postgres:17
PORT=5433
DB_USER=courtflow
DB_NAME=courtflow
DB_PASSWORD=${COURTFLOW_LOCAL_DB_PASSWORD:-courtflow_dev_only}

require_podman() { command -v podman >/dev/null || { echo "Podman not installed" >&2; exit 1; }; }
container_exists() { podman container exists "$NAME"; }
container_running() { [[ $(podman inspect -f '{{.State.Running}}' "$NAME" 2>/dev/null || echo false) == true ]]; }

case "${1:-}" in
  start)
    require_podman
    if ! container_exists; then
      podman volume create "$VOLUME" >/dev/null
      podman run -d --name "$NAME" -p "127.0.0.1:${PORT}:5432" \
        -e "POSTGRES_USER=$DB_USER" -e "POSTGRES_PASSWORD=$DB_PASSWORD" -e "POSTGRES_DB=$DB_NAME" \
        -v "$VOLUME:/var/lib/postgresql/data" "$IMAGE" >/dev/null
    elif ! container_running; then
      podman start "$NAME" >/dev/null
    fi
    echo "Local PostgreSQL started at 127.0.0.1:$PORT (container: $NAME)"
    for attempt in {1..30}; do
      if podman exec "$NAME" pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null; then exit 0; fi
      sleep 1
    done
    echo "Database readiness timed out" >&2; exit 1
    ;;
  stop)
    require_podman
    if container_exists && container_running; then podman stop "$NAME" >/dev/null; fi
    echo "Container stopped; named volume preserved"
    ;;
  status)
    require_podman
    if container_exists; then
      podman inspect -f 'State: {{.State.Status}}' "$NAME"
      if container_running; then podman exec "$NAME" pg_isready -U "$DB_USER" -d "$DB_NAME"; fi
    else
      echo "Container not created"
    fi
    ;;
  psql)
    require_podman
    if ! container_running; then echo "Start local DB first" >&2; exit 1; fi
    shift
    podman exec -it "$NAME" psql -U "$DB_USER" -d "$DB_NAME" "$@"
    ;;
  *)
    echo "Usage: $0 {start|stop|status|psql}" >&2
    exit 2
    ;;
esac
