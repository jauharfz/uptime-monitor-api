package api

import (
	"encoding/json"
	"log"
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
		log.Println("ERROR DECODE JSON: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(contextKeyUserID).(float64)
	if !ok {
		http.Error(w, "cannot get userID", http.StatusUnauthorized)
		return
	}
	monitor.UserID = int(userID)
	err = app.DB.InsertMonitor(monitor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := jsonResponse{
		Status:  "success",
		Message: "monitor creted",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		log.Println("Error encoding response", err)
		return
	}
}

func (app *Application) ShowMonitor(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	monitorID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid monitor id", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(contextKeyUserID).(float64)
	if !ok {
		http.Error(w, "cannot get userID", http.StatusUnauthorized)
		return
	}

	monitor, err := app.DB.GetMonitorByID(monitorID, int(userID))
	if err != nil {
		http.Error(w, "monitor not found", http.StatusInternalServerError)
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
		log.Println("error encoding response", err)
		return
	}
}
