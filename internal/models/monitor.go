package models

import (
	"database/sql"
	"time"
)

type Monitor struct {
	ID             int          `json:"id"`
	UserID         int          `json:"user_id"`
	Url            string       `json:"url"`
	CheckInterval  int          `json:"check_interval"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	LastCheckedAt  sql.NullTime `json:"last_checked_at"`
	WebhookUrl     string       `json:"webhook_url,omitempty"`
	LastStatusCode int          `json:"last_status_code,omitempty"`
}

type MonitorStats struct {
	TotalChecks      int     `json:"total_checks"`
	AvgResponseTime  float64 `json:"avg_response_time"`
	UptimePercentage float64 `json:"uptime_percentage"`
}
