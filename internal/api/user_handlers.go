package api

import (
	"encoding/json"
	"log"
	"net/http"
	"uptime-monitor/internal/auth"
	"uptime-monitor/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func (app *Application) PostUserRegister(w http.ResponseWriter, r *http.Request) {
	var user models.User
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	decode := json.NewDecoder(r.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		http.Error(w, "Failed to Process Password", http.StatusInternalServerError)
		return
	}

	user.Password = string(hash)
	err = app.DB.InsertUser(user)
	if err != nil {
		http.Error(w, "Database Error(probably)", http.StatusInternalServerError)
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
		log.Println("Error encoding response: ", err)
		return
	}
}

func (app *Application) PostUserLogin(w http.ResponseWriter, r *http.Request) {
	var user models.User
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dbUser, err := app.DB.GetUserByEmail(user.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	token, err := auth.GenerateJWT(dbUser.ID)
	if err != nil {
		http.Error(w, "Failed to Generate Token", http.StatusInternalServerError)
		return
	}

	response := jsonResponse{
		Status:  "success",
		Message: "Login Success",
		Data: map[string]string{
			"token": token,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		log.Println("Error encoding response: ", err)
		return
	}
}
