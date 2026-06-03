package api

import "net/http"

func Routes(app *Application) http.Handler {

	// Public Routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", app.HealthTest)
	mux.HandleFunc("POST /users/register", app.PostUserRegister)
	mux.HandleFunc("POST /users/login", app.PostUserLogin)

	// Protected Routes
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("POST /monitor", app.CreateMonitor)
	protectedMux.HandleFunc("GET /monitor", app.ListMonitors)
	protectedMux.HandleFunc("GET /monitor/{id}", app.ShowMonitor)
	protectedMux.HandleFunc("PATCH /monitor/{id}", app.UpdateMonitor)
	protectedMux.HandleFunc("DELETE /monitor/{id}", app.DeleteMonitor)
	protectedMux.HandleFunc("GET /monitor/{id}/checks", app.CheckMonitor)
	protectedMux.HandleFunc("GET /monitor/{id}/stats", app.ShowMonitorStats)

	mux.Handle("/", app.RequireAuth(protectedMux))

	return mux
}
