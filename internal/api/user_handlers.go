package api

import (
	"encoding/json"
	"net/http"
	"uptime-monitor/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func (app *Application) PostUserRegister(w http.ResponseWriter, r *http.Request) {
	var user models.User
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	decode := json.NewDecoder(r.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.Password = string(hash)
	err = app.DB.InsertUser(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := jsonResponse{
		Status:  "success",
		Message: "user created",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
