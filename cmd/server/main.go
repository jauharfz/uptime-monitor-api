package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
	err := config.LoadEnv(".env")
	if err != nil {
		log.Println(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	conn, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect database %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected to database")

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
		log.Println("running server locally")
		err = srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP Server Error", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		log.Println("HTTP Server Shutdown Error", err)
	}

	wg.Wait()
}
