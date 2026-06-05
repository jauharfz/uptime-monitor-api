-- Seed the database with N identical monitors for the scheduling benchmark.
-- All monitors point at the in-cluster echo target and share one interval, so
-- the only independent variables are the strategy and N.
--
-- Usage (variables optional; defaults n=100, interval=30):
--   psql ... -v n=1000 -v interval=30 -f scripts/seed.sql
--
-- IMPORTANT: this TRUNCATEs monitors and checks. Run it only against a
-- dedicated benchmark database, never production data.

\if :{?n}
\else
  \set n 100
\endif
\if :{?interval}
\else
  \set interval 30
\endif

-- A single benchmark-owner user (monitors require a user_id FK).
INSERT INTO users (id, username, password, email)
VALUES (1, 'bench', 'x', 'bench@example.test')
ON CONFLICT (id) DO NOTHING;

-- Clean slate: reset monitors + their checks and restart the id sequences.
TRUNCATE checks, monitors RESTART IDENTITY CASCADE;

-- Create N monitors in one statement.
INSERT INTO monitors (user_id, url, check_interval)
SELECT 1, 'http://echo:9000/', :interval
FROM generate_series(1, :n);

SELECT count(*) AS seeded_monitors,
       min(check_interval) AS interval_seconds
FROM monitors;
