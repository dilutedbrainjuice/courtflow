import subprocess,os,time,json,urllib.request,urllib.error,concurrent.futures,threading,datetime,statistics
ROOT='/home/terasit/VSProjects/court'; schema='audit_'+str(os.getpid()); results={'schema':schema,'benchmarks':[],'rounds':[]}
def sql(q,check=True):
 p=subprocess.run(['podman','exec','-i','courtflow-postgres','psql','-X','-qAt','-v','ON_ERROR_STOP=1','-U','courtflow','-d','courtflow'],input=q,text=True,capture_output=True)
 if check and p.returncode: raise RuntimeError(p.stderr)
 return p
prefix='SET search_path TO '+schema+',public; '
def query(q): return sql(prefix+q).stdout.strip()
def start(port):
 env=dict(os.environ,DATABASE_URL='postgres://courtflow:courtflow_dev_only@127.0.0.1:5433/courtflow?sslmode=disable&search_path='+schema+',public',PORT=str(port))
 p=subprocess.Popen(['/tmp/courtflow-fixed-api'],env=env,stdout=subprocess.DEVNULL,stderr=open('/tmp/courtflow-audit/server.log','a'))
 for _ in range(100):
  try:
   urllib.request.urlopen('http://127.0.0.1:'+str(port)+'/healthz',timeout=1).read();return p
  except Exception: time.sleep(.02)
 raise RuntimeError('startup')
def stop(p):
 p.terminate();p.wait(timeout=15)
def req(port,path,body=None):
 t=time.perf_counter()
 try:
  r=urllib.request.urlopen(urllib.request.Request('http://127.0.0.1:'+str(port)+path,data=body,headers={'Content-Type':'application/json'}),timeout=12);r.read();code=r.status
 except urllib.error.HTTPError as e:
  code=e.code; detail=e.read()
  if code==503: pass
 except Exception: code=0
 return code,(time.perf_counter()-t)*1000
now=datetime.datetime.now(datetime.timezone(datetime.timedelta(hours=8))); day=now.date()+datetime.timedelta(days=1); slot=str(day)+'T10:00:00+08:00'; body=json.dumps({'court_id':1,'start_at':slot}).encode()
procs=[]
try:
 sql('CREATE SCHEMA '+schema+';');sql(prefix+open(ROOT+'/backend/migrations/001_initial.sql').read())
 results['postgres']=query('SELECT version();')
 for n in range(0):
  procs=[start(18091),start(18092)]; gate=threading.Barrier(101)
  def contend(i): gate.wait(); return req(18091+i%2,'/api/v1/bookings',body)
  with concurrent.futures.ThreadPoolExecutor(max_workers=100) as ex:
   fs=[ex.submit(contend,i) for i in range(100)];t=time.perf_counter();gate.wait();r=[f.result() for f in fs];elapsed=time.perf_counter()-t
  counts={str(c):sum(x[0]==c for x in r) for c in set(x[0] for x in r)}
  persisted=int(query('SELECT count(*) FROM bookings;'))
  assert persisted<=1,(n,counts,persisted)
  results['rounds'].append({'round':n+1,'codes':counts,'rows':persisted,'seconds':elapsed,'latencies_ms':[x[1] for x in r]})
  for p in procs:stop(p)
  procs=[];query('TRUNCATE bookings;')
  if n%20==0:print('round',n+1,flush=True)
 # Direct SQL tests: off-hour overlap must fail; adjacent legal; failed transaction rolls back prior insert.
 query("INSERT INTO bookings(court_id,start_at,end_at,amount_sen) VALUES(1,'"+slot+"','"+slot+"'::timestamptz+interval '1 hour',2500);")
 overlap=sql(prefix+"INSERT INTO bookings(court_id,start_at,end_at,amount_sen) VALUES(1,'"+slot+"'::timestamptz+interval '30 minutes','"+slot+"'::timestamptz+interval '90 minutes',2500);",False)
 assert overlap.returncode and 'exclusion constraint' in overlap.stderr
 rollback=sql(prefix+"BEGIN; INSERT INTO bookings(court_id,start_at,end_at,amount_sen) VALUES(2,'"+slot+"','"+slot+"'::timestamptz+interval '1 hour',2500); INSERT INTO bookings(court_id,start_at,end_at,amount_sen) VALUES(1,'"+slot+"','"+slot+"'::timestamptz+interval '1 hour',2500); COMMIT;",False)
 assert rollback.returncode and query('SELECT count(*) FROM bookings WHERE court_id=2;')=='0'
 query("INSERT INTO bookings(court_id,start_at,end_at,amount_sen) VALUES(1,'"+slot+"'::timestamptz+interval '1 hour','"+slot+"'::timestamptz+interval '2 hour',2500);")
 results['sql_checks']={'partial_overlap_rejected':True,'transaction_rollback':True,'adjacent_accepted':True}
 procs=[start(18091)];p=procs[0]
 def cpu():
  fields=open('/proc/'+str(p.pid)+'/stat').read().split();return (int(fields[13])+int(fields[14]))/os.sysconf('SC_CLK_TCK')
 for path in ['/api/v1/facilities','/api/v1/courts/1/availability?date='+str(day)]:
  for concurrency in [1,10,50,100]:
   samples=[];rss=[];connections=[];finish=threading.Event()
   def monitor():
    while not finish.is_set():
     try:
      for line in open('/proc/'+str(p.pid)+'/status'):
       if line.startswith('VmRSS:'):rss.append(int(line.split()[1]))
      connections.append(query("SELECT count(*)||','||count(*) FILTER(WHERE state='active') FROM pg_stat_activity WHERE pid<>pg_backend_pid() AND datname='courtflow';"))
     except Exception:pass
     finish.wait(.2)
   m=threading.Thread(target=monitor);m.start();c0=cpu();t=time.perf_counter();deadline=t+3
   def worker():
    out=[]
    while time.perf_counter()<deadline:out.append(req(18091,path))
    return out
   with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency) as ex:
    for r in ex.map(lambda _:worker(),range(concurrency)): samples.extend(r)
   elapsed=time.perf_counter()-t;consumed=cpu()-c0;finish.set();m.join();lat=sorted(x[1] for x in samples)
   quant=lambda q:lat[min(len(lat)-1,int((len(lat)-1)*q))]
   row=dict(path=path,concurrency=concurrency,seconds=elapsed,requests=len(samples),rps=len(samples)/elapsed,average_ms=statistics.mean(lat),p50_ms=quant(.5),p95_ms=quant(.95),p99_ms=quant(.99),unexpected=sum(c!=200 for c,_ in samples),cpu_percent_one_core=100*consumed/elapsed,max_rss_kib=max(rss),db_total_active_samples=connections)
   results['benchmarks'].append(row);print(json.dumps(row),flush=True)
 # Demonstrate limiter starvation independently of slot occupancy.
 codes=[req(18091,'/api/v1/bookings',b'{}')[0] for _ in range(120)]
 results['limiter_probe']={'invalid_counts':{str(c):codes.count(c) for c in set(codes)},'legitimate_after_invalid':req(18091,'/api/v1/bookings',body)[0]}
finally:
 for p in procs:
  if p.poll() is None:stop(p)
 sql('DROP SCHEMA IF EXISTS '+schema+' CASCADE;',False)
 open('/tmp/courtflow-audit/post-read.json','w').write(json.dumps(results,indent=2))
