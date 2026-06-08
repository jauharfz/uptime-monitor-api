package worker

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
	"uptime-monitor/internal/api"
	"uptime-monitor/internal/models"
)

func Ping(targetURL string) (int, time.Duration, error) {
	start := time.Now()
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(targetURL)
	if err != nil {
		slog.Warn("failed to get url target", "error", err)
		return 0, 0, err
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)
	return resp.StatusCode, elapsed, nil
}

// maxCheckConcurrency bounds how many monitors are probed at the same time, so
// the polling worker's peak memory and connection use stay flat regardless of
// the number of monitors. A real system never fans out one goroutine per
// monitor without limit.
const maxCheckConcurrency = 50

// checkAndStore probes one monitor and records the result. Shared by both
// schedulers so each performs identical work — only the scheduling layer differs.
func checkAndStore(app *api.Application, m models.Monitor) {
	status, duration, err := Ping(m.Url)
	if err != nil {
		status = 0
	}
	if err := app.DB.InsertCheck(m.ID, status, int(duration.Milliseconds())); err != nil {
		slog.Error("failed to insert check", "url", m.Url, "error", err)
		return
	}
	if err := app.DB.UpdateLastCheckedMonitor(m.ID); err != nil {
		slog.Error("failed to update last_checked_at", "url", m.Url, "error", err)
	}
}

// StartPollingWorker runs the database-polling (stateless) scheduling strategy:
// on every tick it asks the database which monitors are due and checks them,
// with the number of concurrent checks bounded by maxCheckConcurrency. All
// scheduling state lives in the monitors.last_checked_at column, never in
// process memory. The tick duration is the scheduling granularity knob.
func StartPollingWorker(ctx context.Context, wg *sync.WaitGroup, app *api.Application, tick time.Duration) {
	defer wg.Done()
	slog.Info("worker started", "strategy", "polling", "tick", tick.String(), "maxConcurrency", maxCheckConcurrency)
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	sem := make(chan struct{}, maxCheckConcurrency)
	for {
		select {
		case <-ctx.Done():
			slog.Info("worker shutdown gracefully")
			return
		case <-ticker.C:
			monitors, err := app.DB.GetMonitorsDueForCheck()
			if err != nil {
				slog.Error("failed to get monitors due for check", "error", err)
				continue
			}
			var batchWg sync.WaitGroup
		dispatch:
			for _, monitor := range monitors {
				select {
				case <-ctx.Done():
					break dispatch
				case sem <- struct{}{}:
					batchWg.Add(1)
					go func(m models.Monitor) {
						defer batchWg.Done()
						defer func() { <-sem }()
						checkAndStore(app, m)
					}(monitor)
				}
			}
			batchWg.Wait()
		}
	}
}
