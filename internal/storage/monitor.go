package storage

import (
	"context"
	"time"
	"uptime-monitor/internal/models"
)

func (s *PostgresStore) GetAllMonitorByUserID(userID int) ([]models.Monitor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT * FROM monitors WHERE user_id = $1`
	var monitors []models.Monitor
	var monitor models.Monitor

	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err = rows.Scan(&monitor.ID, &monitor.UserID, &monitor.Url, &monitor.CheckInterval, &monitor.CreatedAt, &monitor.UpdatedAt)
		if err != nil {
			return nil, err
		}
		monitors = append(monitors, monitor)
	}
	if rows.Err() != nil {
		return nil, err
	}
	return monitors, nil
}

func (s *PostgresStore) GetMonitorByID(id, userID int) (models.Monitor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT * FROM monitors WHERE id = $1 and user_id = $2`
	var monitor models.Monitor
	row := s.DB.QueryRowContext(ctx, query, id, userID)
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

func (s *PostgresStore) UpdateMonitorByID(id, userID, checkInterval int, url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE monitors SET url = $1, check_interval = $2, updated_at = $3 WHERE id = $4 and user_id = $5`
	_, err := s.DB.ExecContext(ctx, stmt, url, checkInterval, time.Now(), id, userID)
	if err != nil {
		return err
	}
	return nil
}
