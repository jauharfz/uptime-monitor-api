package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"uptime-monitor/internal/models"
)

func (app *Application) CreateMonitor(w http.ResponseWriter, r *http.Request) {
	var monitor models.Monitor
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&monitor)
	if err != nil {
		slog.Warn("failed to decode user request", "error", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(contextKeyUserID).(float64)
	if !ok {
		slog.Warn("failed to get user id from context")
		http.Error(w, "invalid request", http.StatusUnauthorized)
		return
	}
	monitor.UserID = int(userID)
	err = app.DB.InsertMonitor(monitor)
	if err != nil {
		slog.Error("user failed to insert monitor to database", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := jsonResponse{
		Status:  "success",
		Message: "monitor created",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		slog.Error("failed to encoding json response to user", "error", err)
		return
	}
}

func (app *Application) ListMonitors(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(contextKeyUserID).(float64)
	if !ok {
		slog.Warn("failed to get user id from context")
		http.Error(w, "invalid request", http.StatusUnauthorized)
		return
	}
	monitors, err := app.DB.GetAllMonitorByUserID(int(userID))
	if err != nil {
		slog.Error("user failed to get all monitor by user id", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := jsonResponse{
		Status:  "success",
		Message: "get all user monitors",
		Data:    monitors,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error("failed to encoding json response to user", "error", err)
		return
	}

}

func (app *Application) ShowMonitor(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	monitorID, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("failed parsing path value to int")
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(contextKeyUserID).(float64)
	if !ok {
		slog.Warn("failed to get user id from context")
		http.Error(w, "invalid request", http.StatusUnauthorized)
		return
	}

	monitor, err := app.DB.GetMonitorByID(monitorID, int(userID))
	if err != nil {
		slog.Warn("user failed to get monitor by user id", "error", err)
		http.Error(w, "monitor not found", http.StatusNotFound)
		return
	}

	response := jsonResponse{
		Status:  "success",
		Message: "get monitor",
		Data:    monitor,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error("failed to encoding json response to user", "error", err)
		return
	}
}

func (app *Application) UpdateMonitor(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	idStr := r.PathValue("id")
	monitorID, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("failed parsing path value to int")
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(contextKeyUserID).(float64)
	if !ok {
		slog.Warn("failed to get user id from context")
		http.Error(w, "invalid request", http.StatusUnauthorized)
		return
	}

	var monitor models.Monitor
	decode := json.NewDecoder(r.Body)
	decode.DisallowUnknownFields()
	err = decode.Decode(&monitor)
	if err != nil {
		slog.Warn("failed to decode user request", "error", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	monitor.ID = monitorID
	monitor.UserID = int(userID)

	_, err = app.DB.GetMonitorByID(monitorID, int(userID))
	if err != nil {
		slog.Warn("user failed to get monitor by user id", "error", err)
		http.Error(w, "monitor not found", http.StatusNotFound)
		return
	}

	err = app.DB.UpdateMonitorByID(monitor.ID, monitor.UserID, monitor.CheckInterval, monitor.Url)
	if err != nil {
		slog.Error("user cannot update monitor by id from database", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := jsonResponse{
		Status:  "success",
		Message: "monitor updated",
		Data:    monitor,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error("failed to encoding json response to user", "error", err)
		return
	}
}

func (app *Application) DeleteMonitor(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("failed parsing path value to int")
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(contextKeyUserID).(float64)
	if !ok {
		slog.Warn("failed to get user id from context")
		http.Error(w, "invalid request", http.StatusUnauthorized)
		return
	}
	err = app.DB.DeleteMonitorByID(id, int(userID))
	if err != nil {
		slog.Error("failed to delete monitor by id from database", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := jsonResponse{
		Status:  "success",
		Message: "monitor has been deleted",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error("failed to encoding json response to user", "error", err)
		return
	}
}
