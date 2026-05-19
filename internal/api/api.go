package api

import (
	"uptime-monitor/internal/storage"
)

type Application struct {
	DB storage.PostgresStore
}

func NewApplication(db storage.PostgresStore) *Application {
	return &Application{
		DB: db,
	}
}
