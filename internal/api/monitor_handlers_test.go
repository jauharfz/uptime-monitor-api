package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"uptime-monitor/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// TODO:call insert user, get user by email, pass context with id
func TestApplication_CreateMonitor(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload := models.Monitor{
		Url:           "google.com",
		CheckInterval: 10,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal json data %v", err)
	}

	bodyReader := bytes.NewBuffer([]byte(jsonData))

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, "/monitor", bodyReader)
	if err != nil {
		t.Fatalf("failed to get new req w/ ctx %v", err)
	}

	user := models.User{
		Username: "testCreateMonitor",
		Password: "passwordCreateMonitor",
		Email:    "testCreateMonitor@example.com",
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		t.Fatalf("failed to hashing user password")
	}
	user.Password = string(hashedPwd)

	err = app.DB.InsertUser(user)
	if err != nil {
		t.Fatalf("failed to insert new user")
	}

	user, err = app.DB.GetUserByEmail(user.Email)
	if err != nil {
		t.Fatalf("failed to get user by email")
	}

	// r = r.WithContext(getCtx(r))
	r.Header.Set("Content-Type", "application/json")
	ctx = context.WithValue(r.Context(), contextKeyUserID, float64(user.ID))
	r = r.WithContext(ctx)
	rr := httptest.NewRecorder()
	handlers := http.HandlerFunc(app.CreateMonitor)
	handlers.ServeHTTP(rr, r)
	if rr.Code != http.StatusCreated {
		t.Errorf("test failed, expected: %d, result: %d", http.StatusCreated, rr.Code)
	}

	var response jsonResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Errorf("failed to decode response json %v", err)
	}

	if response.Status != "success" {
		t.Errorf("test failed, expected: success, result: %s", response.Status)
	}
	if response.Message != "monitor created" {
		t.Errorf("test failed, expected: monitor created, result: %s", response.Message)
	}
}
