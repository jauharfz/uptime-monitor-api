package worker

import "uptime-monitor/internal/storage"

type Worker struct {
	DB storage.PostgresStore
}

func NewWorker(db *storage.PostgresStore) *Worker {
	return &Worker{
		DB: *db,
	}
}
