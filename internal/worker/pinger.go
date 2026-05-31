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

func StartWorker(ctx context.Context, wg *sync.WaitGroup, app *api.Application) {
	defer wg.Done()
	slog.Info("worker started")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("worker shutdown gracefully")
			return
		case <-ticker.C:
			monitors, err := app.DB.GetMonitorsDueForCheck()
			if err != nil {
				slog.Error("failed to get all monitors due for check from database", "error", err)
				continue
			}
			var batchWg sync.WaitGroup

			for _, monitor := range monitors {
				batchWg.Add(1)
				go func(m models.Monitor) {
					defer batchWg.Done()
					status, duration, err := Ping(m.Url)
					if err != nil {
						slog.Warn("failed to ping monitor url", "url", m.Url, "error", err)
						status = 0
					}
					slog.Info("Url Pinged")
					err = app.DB.InsertCheck(m.ID, status, int(duration.Milliseconds()))
					if err != nil {
						slog.Error("failed to insert checks to monitor from database", "url", m.Url, "error", err)
						return
					}
					err = app.DB.UpdateLastCheckedMonitor(m.ID)
					if err != nil {
						slog.Error("failed to update lash checked monitor by monitor id from database", "url", m.Url, "error", err)
						return
					}
					slog.Info("Url Inserted Normally")
				}(monitor)
			}
			batchWg.Wait()
		}
	}
}
