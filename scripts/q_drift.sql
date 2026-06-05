-- Single bare value: mean scheduling drift (ms) vs target interval. Used by the
-- orchestrator scripts with psql -tA. Returns 0 when there is not enough data.
SELECT COALESCE(round(avg(abs(extract(epoch FROM (created_at - prev)) - check_interval)) * 1000), 0) AS mean_drift_ms
FROM (
    SELECT ch.created_at,
           m.check_interval,
           LAG(ch.created_at) OVER (PARTITION BY ch.monitor_id ORDER BY ch.created_at) AS prev
    FROM checks ch
    JOIN monitors m ON m.id = ch.monitor_id
) t
WHERE prev IS NOT NULL;
