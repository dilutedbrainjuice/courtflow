#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
export DATABASE_URL='postgres://courtflow:courtflow_dev_only@127.0.0.1:5433/courtflow?sslmode=disable'
export PORT=18080
(cd backend && go build -o /tmp/courtflow-persistence-api ./cmd/server)
/tmp/courtflow-persistence-api >/tmp/courtflow-persistence.log 2>&1 &
api_pid=$!
booking_id=''
cleanup(){ kill "$api_pid" 2>/dev/null || true; if [[ -n "$booking_id" ]]; then podman exec courtflow-postgres psql -U courtflow -d courtflow -c "DELETE FROM bookings WHERE id='$booking_id'" >/dev/null; fi; }
trap cleanup EXIT
for attempt in {1..30}; do curl -fsS http://127.0.0.1:18080/healthz >/dev/null 2>&1 && break; sleep 1; done
response=$(python3 - <<'PY'
import datetime,json,urllib.request
start=(datetime.datetime.now(datetime.timezone.utc)+datetime.timedelta(days=20)).replace(hour=3,minute=0,second=0,microsecond=0)
r=urllib.request.Request('http://127.0.0.1:18080/api/v1/bookings',data=json.dumps({'court_id':4,'start_at':start.isoformat()}).encode(),headers={'Content-Type':'application/json'})
print(json.load(urllib.request.urlopen(r))['id'])
PY
)
booking_id=$response
kill "$api_pid"
wait "$api_pid"
podman restart courtflow-postgres >/dev/null
for attempt in {1..30}; do podman exec courtflow-postgres pg_isready -U courtflow >/dev/null && break; sleep 1; done
/tmp/courtflow-persistence-api >/tmp/courtflow-persistence.log 2>&1 &
api_pid=$!
count=$(podman exec courtflow-postgres psql -U courtflow -d courtflow -Atc "SELECT count(*) FROM bookings WHERE id='$booking_id'")
[[ "$count" == 1 ]]
curl --retry 5 --retry-connrefused --retry-delay 1 -fsS http://127.0.0.1:18080/healthz
printf '\nBooking survived API and PostgreSQL restart.\n'
