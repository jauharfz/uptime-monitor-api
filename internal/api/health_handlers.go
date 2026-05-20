package api

import (
	"encoding/json"
	"net/http"
)

func (app *Application) HealthTest(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		Status:  "success",
		Message: "health tested",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
