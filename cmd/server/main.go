package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"uptime-monitor/internal/api"
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

	db := storage.NewPostgresStore(conn)
	repo := api.NewApplication(db)
	handler := api.Routes(repo)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go worker.StartWorker(ctx, &wg, repo)

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
