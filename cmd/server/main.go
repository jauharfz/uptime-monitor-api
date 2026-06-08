package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"uptime-monitor/internal/api"
	"uptime-monitor/internal/auth"
	"uptime-monitor/internal/config"
	"uptime-monitor/internal/storage"
	"uptime-monitor/internal/worker"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	err := config.LoadEnv(".env")
	if err != nil {
		slog.Error(".env not found", "error", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "postgres://postgres:1234567k@localhost:5454/uptime_monitor?sslmode=disable"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("failed to get jwt secret from environment", "error", err)
		os.Exit(1)
	}
	auth.SetSecret(jwtSecret)

	conn, err := sql.Open("pgx", dbUrl)
	if err != nil {
		slog.Error("unable to connect database", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		slog.Error("failed to ping database connection", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to database")

	// Bound the connection pool so a burst of concurrent checks cannot exhaust
	// PostgreSQL connections (default max_connections is 100).
	conn.SetMaxOpenConns(60)
	conn.SetMaxIdleConns(20)
	conn.SetConnMaxIdleTime(5 * time.Minute)

	db := storage.NewPostgresStore(conn)
	repo := api.NewApplication(db)
	handler := api.Routes(repo)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	// Scheduling granularity for the polling strategy (the accuracy/DB-load
	// knob). Default 5s preserves the original behaviour.
	tick := 5 * time.Second
	if v := os.Getenv("TICK_INTERVAL"); v != "" {
		if d, parseErr := time.ParseDuration(v); parseErr == nil && d > 0 {
			tick = d
		} else {
			slog.Warn("invalid TICK_INTERVAL, using default", "value", v, "default", tick.String())
		}
	}

	// SCHEDULER selects which scheduling strategy the background worker runs.
	// This is the single switch point for the comparative experiment; the rest
	// of the application (API, storage, schema) is identical for both.
	wg.Add(1)
	switch strings.ToLower(os.Getenv("SCHEDULER")) {
	case "inmemory", "in-memory":
		go worker.StartInMemoryWorker(ctx, &wg, repo)
	default: // "polling" (default); "query"/"query-driven" accepted as aliases
		go worker.StartPollingWorker(ctx, &wg, repo, tick)
	}

	wg.Add(1)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	go func() {
		defer wg.Done()
		slog.Info("running server locally")
		err = srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP Server Error", "error", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("HTTP Shutdown Error", "error", err)
	}

	wg.Wait()
}
