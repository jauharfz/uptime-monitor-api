package worker

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
	"uptime-monitor/internal/api"
)

func Ping(targetURL string) (int, time.Duration, error) {
	start := time.Now()
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(targetURL)
	if err != nil {
		log.Println(err)
		return 0, 0, err
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)
	return resp.StatusCode, elapsed, nil
}

func StartWorker(ctx context.Context, wg *sync.WaitGroup, app *api.Application) {
	defer wg.Done()
	log.Println("worker started")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("worker shutdown gracefully")
			return
		case <-ticker.C:
			monitors, err := app.DB.GetMonitorsDueForCheck()
			if err != nil {
				log.Println(err)
				continue
			}

			for _, monitor := range monitors {
				status, duration, err := Ping(monitor.Url)
				if err != nil {
					log.Println("cannot ping a url", monitor.Url, err)
					status = 0
				}
				log.Println("Url Pinged")
				err = app.DB.InsertCheck(monitor.ID, status, int(duration.Milliseconds()))
				if err != nil {
					log.Println("internal server error", monitor.Url, err)
					continue
				}
				err = app.DB.UpdateLastCheckedMonitor(monitor.ID)
				if err != nil {
					log.Println("internal server error", monitor.Url, err)
					continue
				}
				log.Println("Url Inserted Normally")
			}
		}
	}
}
