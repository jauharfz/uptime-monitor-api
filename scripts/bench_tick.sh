#!/usr/bin/env bash
# Tick-granularity sensitivity sweep for the polling strategy.
#
# Fixes the strategy (polling) and the number of monitors (N), then varies the
# scheduling tick. For each tick it seeds N monitors, (re)starts the api with
# that tick, samples the api container memory + CPU for $DURATION seconds, then
# computes mean scheduling drift and the number of polling "due" queries. This
# isolates the tick as the only independent variable, so its effect on accuracy
# (drift) and database read load (queries/min) can be characterised directly.
#
# Run from the repo root, AFTER the stack images are built (or it builds them).
# Override via env vars, e.g.:
#   DURATION=120 N=1000 TICKS="1s 5s" ./scripts/bench_tick.sh
set -euo pipefail

TICKS=(${TICKS:-1s 2s 5s 10s})
N=${N:-1000}
DURATION=${DURATION:-600}
INTERVAL=${INTERVAL:-30}
SAMPLE=${SAMPLE:-5}
HEALTH=${HEALTH:-http://127.0.0.1:8080/health}

COMPOSE="docker compose -f docker-compose.yml -f docker-compose.bench.yml"

psql_exec() { $COMPOSE exec -T db sh -c "psql -U \"\$POSTGRES_USER\" -d \"\$POSTGRES_DB\" -q -c \"$1\"" >/dev/null; }
psql_stdin() { $COMPOSE exec -T db sh -c "psql -U \"\$POSTGRES_USER\" -d \"\$POSTGRES_DB\" $1"; }

wait_db() {
  for _ in $(seq 1 60); do
    if $COMPOSE exec -T db sh -c 'pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"' >/dev/null 2>&1; then return; fi
    sleep 2
  done
  echo "database not ready" >&2; exit 1
}

wait_api() {
  for _ in $(seq 1 60); do
    if curl -fsS "$HEALTH" >/dev/null 2>&1; then return; fi
    sleep 1
  done
  echo "api not healthy at $HEALTH" >&2; exit 1
}

to_mib() { awk -v s="$1" 'BEGIN{ n=s+0; if(s ~ /GiB/) print n*1024; else if(s ~ /MiB/) print n; else if(s ~ /KiB/) print n/1024; else print n/1048576 }'; }

PEAK_MEM=0; AVG_CPU=0; MAX_CPU=0
sample() {
  local cid="$1" dur="$2" every="$3" end peak=0 cpus="" line mem cpu mib
  end=$(( $(date +%s) + dur ))
  while [ "$(date +%s)" -lt "$end" ]; do
    line=$(docker stats --no-stream --format '{{.MemUsage}};{{.CPUPerc}}' "$cid" 2>/dev/null || true)
    if [ -n "$line" ]; then
      mem="${line%%;*}"; mem="${mem%% /*}"
      cpu="${line##*;}"; cpu="${cpu%\%}"
      mib=$(to_mib "$mem")
      peak=$(awk -v a="$peak" -v b="$mib" 'BEGIN{print (b>a)?b:a}')
      cpus="$cpus $cpu"
    fi
    sleep "$every"
  done
  PEAK_MEM=$(awk -v p="$peak" 'BEGIN{printf "%.1f", p}')
  AVG_CPU=$(echo "$cpus" | awk '{s=0;n=0;for(i=1;i<=NF;i++){s+=$i;n++} if(n>0) printf "%.1f", s/n; else print 0}')
  MAX_CPU=$(echo "$cpus" | awk '{m=0;for(i=1;i<=NF;i++) if($i>m)m=$i; printf "%.1f", m}')
}

echo "== Building images =="
$COMPOSE build
echo "== Starting db + echo =="
$COMPOSE up -d db echo
wait_db
psql_exec 'CREATE EXTENSION IF NOT EXISTS pg_stat_statements;'

CSV="scripts/bench_tick_results.csv"
echo "tick,N,drift_ms,peak_mem_mib,avg_cpu_pct,max_cpu_pct,due_query_calls,queries_per_min" > "$CSV"
printf "%-6s %-6s %-9s %-13s %-11s %-11s %-10s %-12s\n" tick N drift_ms peak_mem_MiB avg_cpu_% max_cpu_% due_calls queries/min

for tick in "${TICKS[@]}"; do
  echo "== tick=$tick N=$N strategy=polling (running ${DURATION}s) ==" >&2
  $COMPOSE stop api >/dev/null 2>&1 || true
  psql_exec 'SELECT pg_stat_statements_reset();'
  cat scripts/seed.sql | $COMPOSE exec -T db sh -c "psql -U \"\$POSTGRES_USER\" -d \"\$POSTGRES_DB\" -q -v n=$N -v interval=$INTERVAL" >/dev/null

  SCHEDULER="polling" TICK_INTERVAL="$tick" $COMPOSE up -d --force-recreate --no-deps api >/dev/null
  wait_api

  cid=$($COMPOSE ps -q api | tr -d '[:space:]')
  sample "$cid" "$DURATION" "$SAMPLE"

  drift=$(cat scripts/q_drift.sql | psql_stdin "-tA" | tr -d '[:space:]')
  due=$(cat scripts/q_duecalls.sql | psql_stdin "-tA" | tr -d '[:space:]')
  qpm=$(awk -v d="$due" -v dur="$DURATION" 'BEGIN{ if(dur>0) printf "%.1f", d/(dur/60); else print 0 }')

  printf "%-6s %-6s %-9s %-13s %-11s %-11s %-10s %-12s\n" "$tick" "$N" "$drift" "$PEAK_MEM" "$AVG_CPU" "$MAX_CPU" "$due" "$qpm"
  echo "$tick,$N,$drift,$PEAK_MEM,$AVG_CPU,$MAX_CPU,$due,$qpm" >> "$CSV"
done

$COMPOSE stop api >/dev/null 2>&1 || true
echo "== Saved $CSV =="
echo "Stop the stack with: $COMPOSE down"
