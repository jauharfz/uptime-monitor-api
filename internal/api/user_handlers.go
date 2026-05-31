package api

import (
	"encoding/json"
	"log/slog"
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
		slog.Warn("failed to decode request body", "error", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		slog.Warn("failed to hashing user password", "error", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	user.Password = string(hash)
	err = app.DB.InsertUser(user)
	if err != nil {
		slog.Error("failed to insert user to database", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
		slog.Error("failed to encoding json response to user", "error", err)
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
		slog.Warn("failed to decode request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dbUser, err := app.DB.GetUserByEmail(user.Email)
	if err != nil {
		slog.Warn("failed to get user by email to database", "error", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		slog.Warn("failed to hashing user password", "error", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	token, err := auth.GenerateJWT(dbUser.ID)
	if err != nil {
		slog.Error("failed to generate jwt token from user", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
		slog.Error("failed to encoding json response to user", "error", err)
		return
	}
}
