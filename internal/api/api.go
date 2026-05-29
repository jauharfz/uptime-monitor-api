package api

import (
	"uptime-monitor/internal/storage"
)

const maxBytes = 1048576

type Application struct {
	DB storage.PostgresStore
}

func NewApplication(db *storage.PostgresStore) *Application {
	return &Application{
		DB: *db,
	}
}
