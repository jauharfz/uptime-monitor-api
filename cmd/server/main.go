package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"uptime-monitor/internal/api"
	"uptime-monitor/internal/config"
	"uptime-monitor/internal/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const addr = ":8080"

func main() {
	err := config.LoadEnv(".env")
	if err != nil {
		log.Println(err)
	}

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
	repo := api.NewApplication(*db)
	handler := api.Routes(repo)

	log.Println("running server locally")
	// go worker.StartWorker(repo)
	err = http.ListenAndServe(addr, handler)
	if err != nil {
		log.Fatal(err)
	}
}
