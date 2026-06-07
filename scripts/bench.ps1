<#
.SYNOPSIS
  Run the scheduling-strategy benchmark end to end on Docker Desktop (Windows).

.DESCRIPTION
  For each (strategy x N) it: seeds N monitors, (re)starts the api with that
  strategy so it loads them, samples the api container's peak memory and CPU via
  `docker stats` for -DurationSec, then computes mean scheduling drift from the
  checks table. Results feed Tabel 1 (drift), Tabel 2 (memory), Tabel 3 (DB
  overhead) and Tabel 5 (CPU) of the journal.

.EXAMPLE
  ./scripts/bench.ps1 -Ns 10,100 -DurationSec 120        # quick smoke
  ./scripts/bench.ps1 -Ns 10,100,1000,5000 -DurationSec 600   # full run

.NOTES
  Run from the repo root. Requires Docker Desktop running and a .env file
  (POSTGRES_USER/PASSWORD/DB, JWT_SECRET, PORT). It TRUNCATEs monitors/checks
  between runs, so point it at a dedicated benchmark database.
#>
[CmdletBinding()]
param(
    [string[]] $Strategies = @('polling', 'inmemory'),
    [int[]]    $Ns = @(10, 100, 1000, 5000),
    [int]      $DurationSec = 600,
    [int]      $IntervalSec = 30,
    [string]   $Tick = '5s',
    [int]      $SampleSec = 5,
    [string]   $HealthUrl = 'http://127.0.0.1:8080/health'
)

# Deliberately NOT $ErrorActionPreference='Stop': docker writes progress to
# stderr and Windows PowerShell would turn that into a terminating error. We
# check $LASTEXITCODE explicitly. We also keep POSTGRES_USER/DB UNQUOTED inside
# sh (they have no spaces) and always feed SQL via stdin, to avoid the PS 5.1
# native-argument double-quote mangling bug.
$compose = @('compose', '-f', 'docker-compose.yml', '-f', 'docker-compose.bench.yml')

function Compose { docker @compose @args }

function Assert-Ok([string]$what) {
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: $what failed (exit $LASTEXITCODE)" -ForegroundColor Red
        exit 1
    }
}

function S([object]$x) { if ($null -eq $x) { return '' } return ([string]$x).Trim() }

# Run SQL (from stdin) and return its output.
function Psql([string]$sqlText, [string]$extraArgs = '') {
    $sqlText | docker @compose exec -T db sh -c "psql -U `$POSTGRES_USER -d `$POSTGRES_DB $extraArgs"
}

# Run SQL (from stdin) discarding output.
function PsqlQuiet([string]$sqlText) {
    $sqlText | docker @compose exec -T db sh -c 'psql -U $POSTGRES_USER -d $POSTGRES_DB -q' | Out-Null
}

function Wait-DbHealthy {
    for ($i = 0; $i -lt 60; $i++) {
        docker @compose exec -T db sh -c 'pg_isready -U $POSTGRES_USER -d $POSTGRES_DB' *> $null
        if ($LASTEXITCODE -eq 0) { return }
        Start-Sleep -Seconds 2
    }
    Write-Host 'ERROR: database did not become ready in time' -ForegroundColor Red; exit 1
}

function Wait-ApiHealthy {
    for ($i = 0; $i -lt 60; $i++) {
        try {
            $r = Invoke-WebRequest -UseBasicParsing -TimeoutSec 2 -Uri $HealthUrl -ErrorAction Stop
            if ($r.StatusCode -eq 200) { return }
        } catch { }
        Start-Sleep -Seconds 1
    }
    Write-Host "ERROR: api did not become healthy at $HealthUrl" -ForegroundColor Red; exit 1
}

function ConvertTo-MiB([string]$s) {
    if ($s -match '([\d\.]+)\s*([KMGT]?i?B)') {
        $v = [double]$matches[1]
        switch -regex ($matches[2]) {
            'GiB' { return $v * 1024 }
            'MiB' { return $v }
            'KiB' { return $v / 1024 }
            default { return $v / 1048576 }
        }
    }
    return 0.0
}

function Sample-Stats([string]$cid, [int]$durationSec, [int]$everySec) {
    $peakMem = 0.0; $cpu = @()
    $deadline = (Get-Date).AddSeconds($durationSec)
    while ((Get-Date) -lt $deadline) {
        $line = docker stats --no-stream --format '{{.MemUsage}};{{.CPUPerc}}' $cid 2>$null
        if ($line) {
            $parts = $line.Split(';')
            $mem = ConvertTo-MiB ($parts[0].Split('/')[0].Trim())
            if ($mem -gt $peakMem) { $peakMem = $mem }
            $cpu += [double]($parts[1].TrimEnd('%').Trim())
        }
        Start-Sleep -Seconds $everySec
    }
    $avgCpu = if ($cpu.Count) { ($cpu | Measure-Object -Average).Average } else { 0 }
    $maxCpu = if ($cpu.Count) { ($cpu | Measure-Object -Maximum).Maximum } else { 0 }
    return [pscustomobject]@{
        PeakMemMiB = [math]::Round($peakMem, 1)
        AvgCpuPct  = [math]::Round($avgCpu, 1)
        MaxCpuPct  = [math]::Round($maxCpu, 1)
    }
}

Write-Host '== Building images (cached after first run) ==' -ForegroundColor Cyan
Compose build *> $null
Assert-Ok 'compose build'

Write-Host '== Starting db + echo ==' -ForegroundColor Cyan
Compose up -d db echo *> $null
Assert-Ok 'compose up db echo'
Wait-DbHealthy
PsqlQuiet 'CREATE EXTENSION IF NOT EXISTS pg_stat_statements;'

$driftSql = Get-Content -Raw scripts/q_drift.sql
$dueSql = Get-Content -Raw scripts/q_duecalls.sql
$seedSql = Get-Content -Raw scripts/seed.sql
$results = @()

foreach ($strategy in $Strategies) {
    foreach ($n in $Ns) {
        Write-Host ("== strategy={0} N={1} (running {2}s) ==" -f $strategy, $n, $DurationSec) -ForegroundColor Yellow

        Compose stop api *> $null
        PsqlQuiet 'SELECT pg_stat_statements_reset();'
        $seedCmd = "psql -U `$POSTGRES_USER -d `$POSTGRES_DB -q -v n=$n -v interval=$IntervalSec"
        $seedSql | docker @compose exec -T db sh -c $seedCmd | Out-Null
        Assert-Ok "seed N=$n"

        $env:SCHEDULER = $strategy
        $env:TICK_INTERVAL = $Tick
        Compose up -d --force-recreate --no-deps api *> $null
        Assert-Ok 'recreate api'
        Wait-ApiHealthy

        $cid = S (Compose ps -q api | Select-Object -First 1)
        $stats = Sample-Stats -cid $cid -durationSec $DurationSec -everySec $SampleSec

        $drift = S (Psql $driftSql '-tA')
        $due = S (Psql $dueSql '-tA')

        $row = [pscustomobject]@{
            Strategy    = $strategy
            N           = $n
            DriftMs     = $drift
            PeakMemMiB  = $stats.PeakMemMiB
            AvgCpuPct   = $stats.AvgCpuPct
            MaxCpuPct   = $stats.MaxCpuPct
            DueQueryCnt = $due
        }
        $results += $row
        $row | Format-Table -AutoSize | Out-Host
    }
}

Compose stop api *> $null

Write-Host '== RESULTS ==' -ForegroundColor Green
$results | Format-Table -AutoSize | Out-Host
$csv = 'scripts/bench_results.csv'
$results | Export-Csv -NoTypeInformation -Path $csv
Write-Host "Saved $csv" -ForegroundColor Green
Write-Host 'Stop the stack with:  docker compose -f docker-compose.yml -f docker-compose.bench.yml down' -ForegroundColor DarkGray
