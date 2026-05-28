package models

import "time"

type Check struct {
	ID           int       `json:"id"`
	MonitorID    int       `json:"monitor_id"`
	StatusCode   int       `json:"status_code"`
	ResponseTime int       `json:"response_time"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
