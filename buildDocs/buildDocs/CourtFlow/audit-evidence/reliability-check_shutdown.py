import subprocess,os,time,threading,json,urllib.request,datetime
schema='signal_'+str(os.getpid())
base=['podman','exec','-i','courtflow-postgres','psql','-X','-qAt','-v','ON_ERROR_STOP=1','-U','courtflow','-d','courtflow']
def sql(s):return subprocess.run(base,input=s,text=True,capture_output=True,check=True).stdout.strip()
p=None;lock=None
try:
 sql('CREATE SCHEMA '+schema)
 sql('SET search_path TO '+schema+',public;'+open('backend/migrations/001_initial.sql').read())
 lock=subprocess.Popen(base,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
 lock.stdin.write('SET search_path TO '+schema+',public; BEGIN; SELECT id FROM courts WHERE id=1 FOR UPDATE;\n');lock.stdin.flush()
 assert lock.stdout.readline().strip()=='1'
 env=dict(os.environ,PORT='18093',DATABASE_URL='postgres://courtflow:courtflow_dev_only@127.0.0.1:5433/courtflow?sslmode=disable&search_path='+schema+',public&application_name='+schema)
 p=subprocess.Popen(['/tmp/courtflow-fixed-api'],env=env,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
 opener=urllib.request.build_opener(urllib.request.ProxyHandler({}))
 for _ in range(100):
  try:opener.open('http://127.0.0.1:18093/healthz').close();break
  except OSError:time.sleep(.02)
 result=[]
 def post():
  day=datetime.datetime.now(datetime.timezone(datetime.timedelta(hours=8))).date()+datetime.timedelta(days=1)
  req=urllib.request.Request('http://127.0.0.1:18093/api/v1/bookings',data=json.dumps({'court_id':1,'start_at':str(day)+'T10:00:00+08:00'}).encode(),headers={'Content-Type':'application/json'})
  with opener.open(req,timeout=8) as response: result.append((response.status,json.load(response)))
 t=threading.Thread(target=post);t.start()
 for _ in range(50):
  if sql("SELECT count(*) FROM pg_stat_activity WHERE application_name='"+schema+"' AND wait_event_type='Lock'")=='1':break
  time.sleep(.02)
 else:raise AssertionError('no blocked request')
 p.terminate()
 assert p.poll() is None
 lock.stdin.write('COMMIT;\n');lock.stdin.close();lock.wait(timeout=3)
 t.join(timeout=8);p.wait(timeout=12)
 assert result[0][0]==201 and p.returncode==0,(result,p.returncode)
 assert sql('SET search_path TO '+schema+',public; SELECT count(*) FROM bookings;')=='1'
 print('PASS: SIGTERM during blocked booking drained to HTTP 201, persisted one row, exit 0')
finally:
 if lock and lock.poll() is None:lock.terminate();lock.wait()
 if p and p.poll() is None:p.kill();p.wait()
 sql('DROP SCHEMA '+schema+' CASCADE')
