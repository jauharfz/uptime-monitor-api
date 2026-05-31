package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

func (app *Application) CheckMonitor(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("failed to parsing path value to int", "error", err)
		http.Error(w, "invalid path value", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(contextKeyUserID).(float64)
	if !ok {
		slog.Warn("failed to get user id from context value")
		http.Error(w, "invalid request", http.StatusUnauthorized)
		return
	}

	_, err = app.DB.GetMonitorByID(id, int(userID))
	if err != nil {
		slog.Warn("user failed to get monitor by user id", "error", err)
		http.Error(w, "monitor not found", http.StatusNotFound)
		return
	}

	checks, err := app.DB.GetChecksByMonitorID(id)
	if err != nil {
		slog.Error("failed to check monitor by id from database", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := jsonResponse{
		Status:  "success",
		Message: "get all checks for this url",
		Data:    checks,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error("failed to encoding json response to user", "error", err)
		return
	}
}
