package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"
	"uptime-monitor/internal/api"
	"uptime-monitor/internal/models"
)

// StartInMemoryWorker runs the in-memory scheduling strategy: at startup it
// loads every monitor from the database and spawns one goroutine holding one
// time.Timer per monitor. Each timer fires independently at the monitor's
// check_interval, runs the check, then resets itself.
//
// Contrast with StartPollingWorker (database polling): here the scheduling state (the
// per-target timers) lives entirely in process memory. Memory therefore grows
// O(N) with the number of monitors, and ALL scheduling state is lost when the
// process stops — on restart the timers must be rebuilt from the database.
// This function is the experimental counterpart used to compare the two
// strategies; everything else (prober, storage, schema) is identical.
func StartInMemoryWorker(ctx context.Context, wg *sync.WaitGroup, app *api.Application) {
	defer wg.Done()

	monitors, err := app.DB.GetAllMonitors()
	if err != nil {
		slog.Error("in-memory worker: failed to load monitors from database", "error", err)
		return
	}
	slog.Info("worker started", "strategy", "in-memory", "monitors", len(monitors))

	// One goroutine + one timer per monitor: this is the O(N) cost on purpose.
	var timers sync.WaitGroup
	for _, monitor := range monitors {
		timers.Add(1)
		go func(m models.Monitor) {
			defer timers.Done()
			runMonitorTimer(ctx, app, m)
		}(monitor)
	}

	<-ctx.Done()
	slog.Info("in-memory worker shutdown gracefully, draining timers")
	timers.Wait()
}

// runMonitorTimer owns the lifecycle of a single monitor's in-memory timer.
func runMonitorTimer(ctx context.Context, app *api.Application, m models.Monitor) {
	interval := time.Duration(m.CheckInterval) * time.Second
	if interval <= 0 {
		slog.Warn("in-memory worker: non-positive check_interval, skipping monitor", "id", m.ID, "url", m.Url)
		return
	}

	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			checkAndStore(app, m)
			timer.Reset(interval)
		}
	}
}
