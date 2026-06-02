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

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return monitors, err
	}
	defer rows.Close()
	for rows.Next() {
		var monitor models.Monitor
		err = rows.Scan(&monitor.ID, &monitor.UserID, &monitor.Url, &monitor.CheckInterval, &monitor.CreatedAt, &monitor.UpdatedAt, &monitor.LastCheckedAt)
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
	defer rows.Close()
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

func (s *PostgresStore) GetMonitorsDueForCheck() ([]models.Monitor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT * FROM monitors WHERE last_checked_at IS NULL OR
	last_checked_at + (interval '1 second' * check_interval) <= NOW()`

	var monitors []models.Monitor
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return monitors, err
	}
	defer rows.Close()

	for rows.Next() {
		var monitor models.Monitor
		err = rows.Scan(&monitor.ID, &monitor.UserID, &monitor.Url, &monitor.CheckInterval, &monitor.CreatedAt, &monitor.UpdatedAt, &monitor.LastCheckedAt)
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

func (s *PostgresStore) UpdateLastCheckedMonitor(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE monitors SET last_checked_at = NOW() WHERE id = $1`

	_, err := s.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}
	return nil
}

// total checks, uptime precentage,avg response time
func (s *PostgresStore) GetMonitorStatsById(monitorID int) (models.MonitorStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	query := `SELECT COUNT(id) AS total_checks, 
		COALESCE(AVG(response_time),0) AS avg_response_time,
		COALESCE ( COUNT(id) FILTER(WHERE status_code >= 200 AND status_code < 300)::FLOAT/NULLIF(COUNT(id),0)*100,0) as uptime_percentage
		FROM checks WHERE monitor_id = $1`

	var stats models.MonitorStats

	rows := s.DB.QueryRowContext(ctx, query, monitorID)
	err := rows.Scan(&stats.TotalChecks, &stats.AvgResponseTime, &stats.UptimePercentage)
	if err != nil {
		return stats, err
	}
	return stats, nil
}
