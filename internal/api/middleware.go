package api

import (
	"context"
	"net/http"
	"strings"
	"uptime-monitor/internal/auth"
)

func (app *Application) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "invalid auth header", http.StatusUnauthorized)
			return
		}
		authToken := strings.TrimPrefix(authHeader, "Bearer ")
		payload, err := auth.Verify(authToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyUserID, payload["user_id"])
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

	})
}
