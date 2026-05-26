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
	err := row.Scan(&monitor.ID, &monitor.UserID, &monitor.Url, &monitor.Interval, &monitor.CreatedAt, &monitor.UpdatedAt)
	if err != nil {
		return monitor, err
	}
	return monitor, nil
}

func (s *PostgresStore) InsertMonitor(monitor models.Monitor) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO monitors VALUES($1,$2,$3,$4,$5,$6)`
	_, err := s.DB.ExecContext(ctx, stmt, monitor.ID, monitor.UserID, monitor.Url, monitor.Interval, time.Now(), time.Now())
	if err != nil {
		return err
	}
	return nil
}
