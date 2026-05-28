package worker

import (
	"log"
	"net/http"
	"time"
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

func StartWorker() {
	ticker := time.NewTicker(1 * time.Second)
	log.Println("worker started")
	for {
		<-ticker.C
		log.Println("chekcing url....")
		statusCode, duration, err := Ping("https://google.com")
		log.Println(statusCode, duration, err)
	}
}
