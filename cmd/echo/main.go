// Command echo is a minimal, fast HTTP target used as the probe destination
// during the scheduling benchmark. It always replies 200 OK so that the metric
// being measured is the scheduler's behaviour, not network/target variability.
//
//	GET /        -> 200 "ok" (counts the request; optional fixed latency)
//	GET /count   -> 200 plain-text total number of probes received
//	GET /healthz -> 200 "ok"
//
// Environment:
//
//	ECHO_PORT     listen port (default "9000")
//	ECHO_LATENCY  fixed artificial latency per request, e.g. "5ms" (default "0s")
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

func main() {
	port := os.Getenv("ECHO_PORT")
	if port == "" {
		port = "9000"
	}

	latency := time.Duration(0)
	if v := os.Getenv("ECHO_LATENCY"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			log.Fatalf("echo: invalid ECHO_LATENCY %q: %v", v, err)
		}
		latency = d
	}

	var count int64

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /count", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "%d\n", atomic.LoadInt64(&count))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&count, 1)
		if latency > 0 {
			time.Sleep(latency)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":" + port
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Printf("echo target listening on %s (latency=%s)", addr, latency)
	log.Fatal(srv.ListenAndServe())
}
