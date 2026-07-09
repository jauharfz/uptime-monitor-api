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

// TODO: create body(reader), change newreq to newreqctx
func TestApplication_PostUserRegister(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	payload := models.User{
		Username: "test",
		Password: "password",
		Email:    "test@example.com",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload json %v", err)
	}

	bodyReader := bytes.NewReader(jsonData)

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/users/register", bodyReader)

	r.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handlers := http.HandlerFunc(app.PostUserRegister)
	handlers.ServeHTTP(rr, r)
	if rr.Code != http.StatusCreated {
		t.Errorf("test failed, expected %d, results %d", http.StatusCreated, rr.Code)
	}

	var response jsonResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response json %v", err)
	}

	if response.Status != "success" {
		t.Errorf("test failed, expected: success, result: %s", response.Status)
	}
	if response.Message != "user created" {
		t.Errorf("test failed, expected: user created, result: %s", response.Message)
	}
}

func TestApplication_PostUserLogin(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload := models.User{
		Username: "test",
		Email:    "testLogin@example.com",
		Password: "passwordLogin",
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		t.Fatalf("failed to hashing test password %v", err)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload json %v", err)
	}

	bodyReader := bytes.NewBuffer(jsonData)

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/users/login", bodyReader)

	payload.Password = string(hashedPwd)
	err = app.DB.InsertUser(payload)
	if err != nil {
		t.Fatalf("failed to insert user login test %v", err)
	}

	r.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handlers := http.HandlerFunc(app.PostUserLogin)
	handlers.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("test failed, expected: %d, result: %d", http.StatusOK, rr.Code)
	}

	var response jsonResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response json %v", err)
	}

	if response.Status != "success" {
		t.Errorf("test failed, expected: success, result: %s", response.Status)
	}
	if response.Message != "Login Success" {
		t.Errorf("test failed, expected: Login Success, result: %s", response.Message)
	}
	if response.Data == nil {
		t.Error("test failed, expected: data filled, result: nil")
	}
}
