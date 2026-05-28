package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func (app *Application) CheckMonitor(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid path value", http.StatusBadRequest)
		return
	}

	_, ok := r.Context().Value(contextKeyUserID).(float64)
	if !ok {
		http.Error(w, "cannot get user id", http.StatusUnauthorized)
		return
	}

	checks, err := app.DB.GetChecksByMonitorID(id)
	if err != nil {
		http.Error(w, "cannot get checks for this monitor", http.StatusInternalServerError)
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
		log.Println("error encoding response", err)
		return
	}
}
