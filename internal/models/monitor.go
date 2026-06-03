package models

import (
	"database/sql"
	"time"
)

type Monitor struct {
	ID            int          `json:"id"`
	UserID        int          `json:"user_id"`
	Url           string       `json:"url"`
	CheckInterval int          `json:"check_interval"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	LastCheckedAt sql.NullTime `json:"last_checked_at"`
}

type MonitorStats struct {
	TotalChecks      string `json:"total_checks"`
	AvgResponseTime  string `json:"avg_response_time"`
	UptimePercentage string `json:"uptime_percentage"`
}
