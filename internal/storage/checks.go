package storage

import (
	"context"
	"time"
	"uptime-monitor/internal/models"
)

func (s *PostgresStore) GetAllMonitors() ([]models.Monitor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT * FROM monitors`
	var monitors []models.Monitor
	var monitor models.Monitor

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return monitors, err
	}
	for rows.Next() {
		err = rows.Scan(&monitor.ID, &monitor.UserID, &monitor.Url, &monitor.CheckInterval, &monitor.CreatedAt, &monitor.UpdatedAt)
		if err != nil {
			return monitors, err
		}
		monitors = append(monitors, monitor)
	}
	if rows.Err() != nil {
		return monitors, err
	}
	return monitors, nil
}

func (s *PostgresStore) InsertCheck(monitorID, status, duration int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO checks(monitor_id,status_code,response_time, created_at,updated_at) VALUES ($1,$2,$3,$4,$5)`

	_, err := s.DB.ExecContext(ctx, stmt, monitorID, status, duration, time.Now(), time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) GetChecksByMonitorID(id int) ([]models.Check, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT * FROM checks WHERE monitor_id = $1 ORDER BY created_at DESC LIMIT 50`
	var checks []models.Check
	var check models.Check

	rows, err := s.DB.QueryContext(ctx, query, id)
	if err != nil {
		return checks, err
	}
	for rows.Next() {
		err = rows.Scan(&check.ID, &check.MonitorID, &check.StatusCode, &check.ResponseTime, &check.CreatedAt, &check.UpdatedAt)
		if err != nil {
			return checks, err
		}
		checks = append(checks, check)
	}
	if rows.Err() != nil {
		return checks, err
	}
	return checks, nil
}
