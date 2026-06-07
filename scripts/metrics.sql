-- Compute the benchmark metrics that come from the database, for the run that
-- just finished. Run AFTER stopping the measurement window.
--   psql ... -f scripts/metrics.sql

-- 1) Scheduling accuracy (Tabel 1): mean absolute drift of the actual interval
--    between consecutive checks vs the target interval, per monitor, in ms.
WITH ordered AS (
    SELECT c.monitor_id,
           c.created_at,
           m.check_interval,
           LAG(c.created_at) OVER (PARTITION BY c.monitor_id ORDER BY c.created_at) AS prev
    FROM checks c
    JOIN monitors m ON m.id = c.monitor_id
)
SELECT count(*)                                                                      AS intervals_measured,
       round(avg(abs(extract(epoch FROM (created_at - prev)) - check_interval)) * 1000) AS mean_drift_ms,
       round(avg(extract(epoch FROM (created_at - prev))) * 1000)                    AS mean_period_ms,
       count(DISTINCT monitor_id)                                                    AS monitors_with_checks
FROM ordered
WHERE prev IS NOT NULL;

-- 2) Database scheduling overhead (Tabel 3): how many times the polling
--    "due" query ran. For the in-memory strategy this returns no rows (0).
--    Requires the pg_stat_statements extension (enabled by the bench overlay).
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

SELECT calls                       AS due_query_calls,
       round(total_exec_time)::int AS total_exec_ms,
       left(query, 60)             AS query_sample
FROM pg_stat_statements
WHERE query ILIKE '%monitors WHERE last_checked_at IS NULL%'
ORDER BY calls DESC;

-- Helper for total throughput (sanity): number of checks recorded this run.
SELECT count(*) AS total_checks FROM checks;
