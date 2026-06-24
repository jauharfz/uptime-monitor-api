package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"uptime-monitor/internal/auth"
)

func (app *Application) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			slog.Warn("invalid auth header format")
			http.Error(w, "invalid auth header", http.StatusUnauthorized)
			return
		}
		authToken := strings.TrimPrefix(authHeader, "Bearer ")
		payload, err := auth.Verify(authToken)
		if err != nil {
			slog.Warn("invalid token", "error", err)
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}
		userID, ok := payload["user_id"].(float64)
		if !ok {
			slog.Warn("invalid payload token", "token", payload)
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyUserID, int(userID))
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

	})
}
