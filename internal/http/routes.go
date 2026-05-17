package api

import "net/http"

func Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", HealthTest)
	return mux
}
