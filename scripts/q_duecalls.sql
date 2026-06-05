-- Single bare value: total calls of the query-driven "due" query since the last
-- pg_stat_statements_reset(). Returns 0 for the in-memory strategy (no such
-- query). Requires the pg_stat_statements extension (created once at setup).
SELECT COALESCE(sum(calls), 0) AS due_query_calls
FROM pg_stat_statements
WHERE query ILIKE '%monitors WHERE last_checked_at IS NULL%';
