#!/usr/bin/env bash
set -euo pipefail
python3 - <<'PY'
import datetime,json,urllib.request,subprocess
base='http://127.0.0.1:18081'
assert json.load(urllib.request.urlopen(base+'/healthz'))['status']=='ok'
start=(datetime.datetime.now(datetime.timezone.utc)+datetime.timedelta(days=21)).replace(hour=4,minute=0,second=0,microsecond=0)
request=urllib.request.Request(base+'/api/v1/bookings',data=json.dumps({'court_id':4,'start_at':start.isoformat()}).encode(),headers={'Content-Type':'application/json'})
booking=json.load(urllib.request.urlopen(request))
assert booking['status']=='confirmed' and booking['amount_sen']==2500
subprocess.run(['podman','exec','courtflow-postgres','psql','-U','courtflow','-d','courtflow','-c',"DELETE FROM bookings WHERE id='"+booking['id']+"'"],check=True,capture_output=True)
print('Container health and booking smoke passed; test booking removed.')
PY
