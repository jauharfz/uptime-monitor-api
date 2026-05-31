package api

import (
	"encoding/json"
	"log/slog"
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
		slog.Error("failed to encoding response json", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
