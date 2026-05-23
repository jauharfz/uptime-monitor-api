package api

import "net/http"

func Routes(app *Application) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users/register", app.PostUserRegister)
	mux.HandleFunc("POST /users/login", app.PostUserLogin)
	mux.Handle("GET /secret", app.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("testing secret"))
	})))

	return mux
}
