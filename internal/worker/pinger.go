package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"
	"uptime-monitor/internal/models"
)

func (w *Worker) Ping(targetURL string) (int, time.Duration, error) {
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

func (w *Worker) StartWorker(ctx context.Context, wg *sync.WaitGroup) {
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
			monitors, err := w.DB.GetMonitorsDueForCheck()
			if err != nil {
				slog.Error("failed to get all monitors due for check from database", "error", err)
				continue
			}
			var batchWg sync.WaitGroup

			for _, monitor := range monitors {
				batchWg.Add(1)
				go func(m models.Monitor) {
					defer batchWg.Done()
					status, duration, err := w.Ping(m.Url)
					if err != nil {
						slog.Warn("failed to ping monitor url", "url", m.Url, "error", err)
						status = 0
					}
					slog.Info("Url Pinged")
					check, err := w.DB.InsertCheck(m.ID, status, int(duration.Milliseconds()))
					if err != nil {
						slog.Error("failed to insert checks to monitor from database", "url", m.Url, "error", err)
						return
					}

					if status != m.LastStatusCode && m.WebhookUrl != "" {
						batchWg.Add(1)
						go func(m models.Monitor, c models.Check) {
							w.SendWebhook(&m, &c, &batchWg)
						}(monitor, check)
					}

					err = w.DB.UpdateLastCheckedAndStatusCode(status, m.ID)
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

func (w *Worker) SendWebhook(monitor *models.Monitor, check *models.Check, wg *sync.WaitGroup) {
	defer wg.Done()
	var responseStatus int
	payload := models.CheckWithMonitor{
		MonitorID:     monitor.ID,
		Url:           monitor.Url,
		CheckInterval: monitor.CheckInterval,
		CheckID:       check.ID,
		StatusCode:    check.StatusCode,
		ResponseTime:  check.ResponseTime,
		CheckedAt:     check.CreatedAt,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal payload json", "error", err)
		return
	}
	bytesReader := bytes.NewBuffer(jsonData)

	req, err := http.NewRequest(http.MethodPost, monitor.WebhookUrl, bytesReader)

	req.Header.Add("Content-Type", "application/json")
	for responseStatus != 204 {
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Warn("failed to send request to client", "error", err)
			return
		}
		responseStatus = res.StatusCode
	}
}
