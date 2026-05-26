package models

import "time"

type Monitor struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Url       string    `json:"url"`
	Interval  int       `json:"interval"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
