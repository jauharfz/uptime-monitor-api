package api

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"uptime-monitor/internal/storage"
	"uptime-monitor/internal/worker"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var app *Application

// TODO: MockDB w/ postgres & docker
func TestMain(m *testing.M) {
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://postgres:1234567k@localhost:5454/uptime_monitor_test?sslmode=disable")
	}
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-secret-not-for-prod")
	}
	dbUrl := os.Getenv("DATABASE_URL")

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

	migrationFile := filepath.Join("..", "..", "migrations", "001_init.sql")
	sqlBytes, err := os.ReadFile(migrationFile)
	if err != nil {
		slog.Error("failed to get migration file", "error", err)
	}
	_, err = conn.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		slog.Error("failed to clean database test", "error", err)
	}
	_, err = conn.Exec(string(sqlBytes))
	if err != nil {
		slog.Error("failed to migrate sql database", "error", err)
	}

	db := storage.NewPostgresStore(conn)
	app = NewApplication(db)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	worker := worker.NewWorker(db)
	wg.Add(1)
	go worker.StartWorker(ctx, &wg)

	code := m.Run()

	cancel()
	wg.Wait()

	os.Exit(code)
}
