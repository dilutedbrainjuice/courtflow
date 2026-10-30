CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE TABLE IF NOT EXISTS facilities (
 id integer PRIMARY KEY, name text NOT NULL, timezone text NOT NULL,
 opens integer NOT NULL CHECK(opens BETWEEN 0 AND 22),
 closes integer NOT NULL CHECK(closes > opens AND closes <= 24)
);
CREATE TABLE IF NOT EXISTS courts (
 id integer PRIMARY KEY, facility_id integer NOT NULL REFERENCES facilities(id),
 name text NOT NULL, price_sen integer NOT NULL CHECK(price_sen > 0)
);
CREATE TABLE IF NOT EXISTS bookings (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 court_id integer NOT NULL REFERENCES courts(id),
 start_at timestamptz NOT NULL, end_at timestamptz NOT NULL,
 amount_sen integer NOT NULL CHECK(amount_sen > 0),
 status text NOT NULL DEFAULT 'confirmed' CHECK(status IN ('confirmed','cancelled')),
 created_at timestamptz NOT NULL DEFAULT now(),
 CHECK(end_at = start_at + interval '1 hour'),
 EXCLUDE USING gist (court_id WITH =, tstzrange(start_at,end_at,'[)') WITH &&)
 WHERE (status = 'confirmed')
);
INSERT INTO facilities VALUES (1,'CourtFlow Demo Badminton','Asia/Kuala_Lumpur',8,23) ON CONFLICT DO NOTHING;
INSERT INTO courts VALUES (1,1,'Court A',2500),(2,1,'Court B',2500),(3,1,'Court C',2500),(4,1,'Court D',2500) ON CONFLICT DO NOTHING;
