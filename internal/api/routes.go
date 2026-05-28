package api

import "net/http"

func Routes(app *Application) http.Handler {

	// Public Routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users/register", app.PostUserRegister)
	mux.HandleFunc("POST /users/login", app.PostUserLogin)

	// Protected Routes
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /secret", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Testing Secret Success"))
	})
	protectedMux.HandleFunc("POST /monitor", app.CreateMonitor)
	protectedMux.HandleFunc("GET /monitor", app.ListMonitors)
	protectedMux.HandleFunc("GET /monitor/{id}", app.ShowMonitor)
	protectedMux.HandleFunc("PATCH /monitor/{id}", app.UpdateMonitor)

	mux.Handle("/", app.RequireAuth(protectedMux))

	return mux
}

// TODO: CREATE A GET,PUT, DELETE MONITOR ENDPOINT and the "STUFF"
