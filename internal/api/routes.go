package api

import "net/http"

func Routes(app *Application) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users/register", app.PostUserRegister)
	mux.HandleFunc("POST /users/login", app.PostUserLogin)

	return mux
}
