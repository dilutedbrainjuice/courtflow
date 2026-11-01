#!/usr/bin/env python3
"""Exercise real API processes and PostgreSQL using an isolated local schema."""
import argparse
import concurrent.futures
import datetime
import json
import os
from pathlib import Path
import subprocess
import threading
import time
import urllib.error
import urllib.request


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--binary', required=True)
    parser.add_argument('--instances', type=int, choices=[1, 2], default=2)
    parser.add_argument('--rounds', type=int, default=100)
    parser.add_argument('--output', required=True)
    args = parser.parse_args()
    if args.rounds < 1:
        parser.error('rounds must be positive')
    schema = f'reliability_{os.getpid()}_{time.time_ns()}'
    root = Path(__file__).resolve().parents[1]
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
    rows = []
    processes = []

    def sql(statement):
        result = subprocess.run(
            ['podman', 'exec', '-i', 'courtflow-postgres', 'psql', '-X', '-qAt',
             '-v', 'ON_ERROR_STOP=1', '-U', 'courtflow', '-d', 'courtflow'],
            input=statement, text=True, capture_output=True, check=True)
        return result.stdout.strip()

    def query(statement):
        return sql(f'SET search_path TO {schema},public; ' + statement)

    tomorrow = datetime.datetime.now(datetime.timezone(datetime.timedelta(hours=8))).date() + datetime.timedelta(days=1)
    body = json.dumps({'court_id': 1, 'start_at': f'{tomorrow}T10:00:00+08:00'}).encode()
    log_path = Path(args.output).with_suffix('.server.log')
    log_path.parent.mkdir(parents=True, exist_ok=True)
    with log_path.open('w') as log:
        try:
            sql(f'CREATE SCHEMA {schema};')
            query((root / 'backend/migrations/001_initial.sql').read_text())
            for iteration in range(args.rounds):
                for instance in range(args.instances):
                    port = 18091 + instance
                    env = dict(os.environ, PORT=str(port), DATABASE_URL=
                               'postgres://courtflow:courtflow_dev_only@127.0.0.1:5433/courtflow'
                               f'?sslmode=disable&search_path={schema},public&application_name={schema}_{instance}')
                    process = subprocess.Popen([args.binary], env=env, stdout=log, stderr=log)
                    processes.append(process)
                    for attempt in range(100):
                        if process.poll() is not None:
                            raise RuntimeError('audit API exited at startup; check port ownership/log')
                        try:
                            with opener.open(f'http://127.0.0.1:{port}/healthz', timeout=1) as response:
                                response.read()
                            break
                        except OSError:
                            time.sleep(.02)
                    else:
                        raise RuntimeError('API startup timed out')
                barrier = threading.Barrier(101)

                def request(index):
                    barrier.wait(timeout=15)
                    started = time.monotonic()
                    req = urllib.request.Request(
                        f'http://127.0.0.1:{18091 + index % args.instances}/api/v1/bookings',
                        data=body, headers={'Content-Type': 'application/json'})
                    try:
                        with opener.open(req, timeout=12) as response:
                            payload = json.load(response)
                            return response.status, payload, time.monotonic() - started
                    except urllib.error.HTTPError as error:
                        return error.code, json.load(error), time.monotonic() - started

                with concurrent.futures.ThreadPoolExecutor(max_workers=100) as executor:
                    futures = [executor.submit(request, index) for index in range(100)]
                    started = time.monotonic()
                    barrier.wait(timeout=15)
                    responses = [future.result() for future in futures]
                    elapsed = time.monotonic() - started
                counts = {str(code): sum(item[0] == code for item in responses)
                          for code in {item[0] for item in responses}}
                persisted = query('SELECT count(*),min(amount_sen),min(id::text) FROM bookings;').split('|')
                created = [payload for code, payload, _ in responses if code == 201]
                valid = (counts == {'201': 1, '409': 99} and persisted[0:2] == ['1', '2500']
                         and created[0]['id'] == persisted[2] and created[0]['amount_sen'] == 2500)
                rows.append({'round': iteration + 1, 'codes': counts, 'persisted': persisted,
                             'seconds': elapsed, 'latencies_seconds': [item[2] for item in responses],
                             'passed': valid})
                for process in processes:
                    process.terminate()
                    process.wait(timeout=15)
                processes.clear()
                query('TRUNCATE bookings;')
                if not valid:
                    raise AssertionError(f'round {iteration + 1}: {counts}, persisted={persisted}')
                if (iteration + 1) % 20 == 0:
                    print(f'{iteration + 1}/{args.rounds} rounds passed', flush=True)
        finally:
            for process in processes:
                if process.poll() is None:
                    process.terminate()
                    try:
                        process.wait(timeout=15)
                    except subprocess.TimeoutExpired:
                        process.kill()
                        process.wait()
            sql(f'DROP SCHEMA IF EXISTS {schema} CASCADE;')
            Path(args.output).write_text(json.dumps({'instances': args.instances, 'rounds': rows}, indent=2))
    print(f'PASS: {args.instances} process(es), {args.rounds} rounds, 100 contenders each')


if __name__ == '__main__':
    main()
