package models

import (
	"time"
)

type Check struct {
	ID           int       `json:"id"`
	MonitorID    int       `json:"monitor_id"`
	StatusCode   int       `json:"status_code"`
	ResponseTime int       `json:"response_time"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CheckWithMonitor struct {
	MonitorID     int       `json:"monitor_id"`
	Url           string    `json:"url"`
	CheckInterval int       `json:"check_interval"`
	CheckID       int       `json:"check_id"`
	StatusCode    int       `json:"status_code"`
	ResponseTime  int       `json:"response_time"`
	CheckedAt     time.Time `json:"checked_at"`
}
