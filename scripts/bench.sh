#!/usr/bin/env bash
# Run the scheduling-strategy benchmark end to end (Linux / Azure VM / WSL).
#
# For each (strategy x N): seed N monitors, (re)start the api with that strategy
# so it loads them, sample the api container peak memory + CPU via `docker stats`
# for $DURATION seconds, then compute mean scheduling drift from the checks table.
#
# Run from the repo root. Requires Docker, curl, awk, and a .env file.
# Override via env vars, e.g.:
#   DURATION=120 NS="10 100" ./scripts/bench.sh
set -euo pipefail

STRATEGIES=(${STRATEGIES:-query inmemory})
NS=(${NS:-10 100 1000 5000})
DURATION=${DURATION:-600}
INTERVAL=${INTERVAL:-30}
TICK=${TICK:-5s}
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

CSV="scripts/bench_results.csv"
echo "strategy,N,drift_ms,peak_mem_mib,avg_cpu_pct,max_cpu_pct,due_query_calls" > "$CSV"
printf "%-9s %-6s %-9s %-13s %-11s %-11s %-9s\n" strategy N drift_ms peak_mem_MiB avg_cpu_% max_cpu_% due_calls

for strategy in "${STRATEGIES[@]}"; do
  for n in "${NS[@]}"; do
    echo "== strategy=$strategy N=$n (running ${DURATION}s) ==" >&2
    $COMPOSE stop api >/dev/null 2>&1 || true
    psql_exec 'SELECT pg_stat_statements_reset();'
    cat scripts/seed.sql | $COMPOSE exec -T db sh -c "psql -U \"\$POSTGRES_USER\" -d \"\$POSTGRES_DB\" -q -v n=$n -v interval=$INTERVAL" >/dev/null

    SCHEDULER="$strategy" TICK_INTERVAL="$TICK" $COMPOSE up -d --force-recreate --no-deps api >/dev/null
    wait_api

    cid=$($COMPOSE ps -q api | tr -d '[:space:]')
    sample "$cid" "$DURATION" "$SAMPLE"

    drift=$(cat scripts/q_drift.sql | psql_stdin "-tA" | tr -d '[:space:]')
    due=$(cat scripts/q_duecalls.sql | psql_stdin "-tA" | tr -d '[:space:]')

    printf "%-9s %-6s %-9s %-13s %-11s %-11s %-9s\n" "$strategy" "$n" "$drift" "$PEAK_MEM" "$AVG_CPU" "$MAX_CPU" "$due"
    echo "$strategy,$n,$drift,$PEAK_MEM,$AVG_CPU,$MAX_CPU,$due" >> "$CSV"
  done
done

$COMPOSE stop api >/dev/null 2>&1 || true
echo "== Saved $CSV =="
echo "Stop the stack with: $COMPOSE down"
