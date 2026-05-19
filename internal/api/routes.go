package api

import "net/http"

func Routes(repo *Application) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users/register", repo.PostUserRegister)

	return mux
}
