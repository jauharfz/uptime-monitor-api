package storage

import (
	"context"
	"time"
	"uptime-monitor/internal/models"
)

func (s *PostgresStore) InsertUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO users(username,password,email,created_at,updated_at) values($1,$2,$3,$4,$5)`
	_, err := s.DB.ExecContext(ctx, stmt, user.Username, user.Password, user.Email, time.Now(), time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) GetUserByEmail(email string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT * FROM users WHERE email=$1`
	var user models.User
	row := s.DB.QueryRowContext(ctx, query, email)
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return user, err
	}
	return user, nil
}
