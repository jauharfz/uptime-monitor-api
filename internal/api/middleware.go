package api

import (
	"fmt"
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

		fmt.Printf("DEBUG TOKEN: '%s'\n", authToken)

		_, err := auth.Verify(authToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
