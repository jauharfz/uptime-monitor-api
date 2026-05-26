package storage

import (
	"context"
	"time"
	"uptime-monitor/internal/models"
)

func (s *PostgresStore) GetMonitorByUserID(userID int) (models.Monitor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT * FROM monitors WHERE user_ID = $1`
	var monitor models.Monitor
	row := s.DB.QueryRowContext(ctx, query, monitor.UserID)
	err := row.Scan(&monitor.ID, &monitor.UserID, &monitor.Url, &monitor.CheckInterval, &monitor.CreatedAt, &monitor.UpdatedAt)
	if err != nil {
		return monitor, err
	}
	return monitor, nil
}

func (s *PostgresStore) InsertMonitor(monitor models.Monitor) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO monitors(user_id,url,check_interval,created_at,updated_at) VALUES($1,$2,$3,$4,$5)`
	_, err := s.DB.ExecContext(ctx, stmt, monitor.UserID, monitor.Url, monitor.CheckInterval, time.Now(), time.Now())
	if err != nil {
		return err
	}
	return nil
}
