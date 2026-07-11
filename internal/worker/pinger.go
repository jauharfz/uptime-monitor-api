package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	params := models.CheckWithMonitor{
		MonitorID:     monitor.ID,
		Url:           monitor.Url,
		CheckInterval: monitor.CheckInterval,
		CheckID:       check.ID,
		StatusCode:    check.StatusCode,
		ResponseTime:  check.ResponseTime,
		CheckedAt:     check.CreatedAt,
	}

	status := "DOWN"
	if params.StatusCode >= 200 && params.StatusCode < 300 {
		status = "UP"
	}
	header := fmt.Sprintf("%s <- this URL is %s right now, here's the detail:", params.Url, status)

	content := fmt.Sprintf("**%s**\nMonitor ID: %d\nURL: %s\nCheck Interval: %d\nCheck ID: %d\nStatus Code: %d\nResponse Time: %d\nChecked At: %v", header, params.MonitorID, params.Url, params.CheckInterval, params.CheckID, params.StatusCode, params.ResponseTime, params.CheckedAt)

	payload := struct {
		Content string `json:"content"`
	}{
		Content: content,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal payload json", "error", err)
		return
	}
	bytesReader := bytes.NewBuffer(jsonData)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, monitor.WebhookUrl, bytesReader)

	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Warn("failed to send request to client", "error", err)
		return
	}
	defer res.Body.Close()
	slog.Info("payload sended to webhook", "url", monitor.Url, "status_code", res.StatusCode)
}
