package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	payload := models.Monitor{
		Url:           "https://google.com",
		CheckInterval: 10,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal json data %v", err)
	}

	bodyReader := bytes.NewBuffer([]byte(jsonData))

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/monitors", bodyReader)

	r.Header.Set("Content-Type", "application/json")
	ctx = context.WithValue(r.Context(), contextKeyUserID, user.ID)
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
		t.Fatalf("failed to decode response json %v", err)
	}

	if response.Status != "success" {
		t.Errorf("test failed, expected: success, result: %s", response.Status)
	}
	if response.Message != "monitor created" {
		t.Errorf("test failed, expected: monitor created, result: %s", response.Message)
	}
}

// TODO:refactor context value type assertion to int
// TODO:write the rest of monitor handlers

func TestApplication_ListMonitors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := models.User{
		Username: "testListMonitors",
		Password: "passwordListMonitors",
		Email:    "testListMonitors@example.com",
	}

	err := app.DB.InsertUser(user)
	if err != nil {
		t.Fatalf("failed to insert user %v", err)
	}
	user, err = app.DB.GetUserByEmail(user.Email)
	if err != nil {
		t.Fatalf("failed to get user by email %v", err)
	}
	monitor1 := models.Monitor{
		UserID:        user.ID,
		Url:           "https://google.com",
		CheckInterval: 5,
	}
	monitor2 := models.Monitor{
		UserID:        user.ID,
		Url:           "https://yahoo.com",
		CheckInterval: 10,
	}

	err = app.DB.InsertMonitor(monitor1)
	if err != nil {
		t.Fatalf("failed to insert monitor %v", err)
	}
	err = app.DB.InsertMonitor(monitor2)
	if err != nil {
		t.Fatalf("failed to insert monitor %v", err)
	}

	r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/monitors", nil)

	ctx = context.WithValue(r.Context(), contextKeyUserID, user.ID)
	r = r.WithContext(ctx)
	rr := httptest.NewRecorder()
	handlers := http.HandlerFunc(app.ListMonitors)
	handlers.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("test failed, expected: %d, result: %d", http.StatusOK, rr.Code)
	}

	var response struct {
		Status  string           `json:"status"`
		Message string           `json:"message"`
		Data    []models.Monitor `json:"data"`
	}
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response json %v", err)
	}

	if response.Status != "success" {
		t.Errorf("test failed, expected: success, result: %s", response.Status)
	}
	if response.Message != "get all user monitors" {
		t.Errorf("test failed, expected: get all user monitors, result: %s", response.Message)
	}

	if len(response.Data) != 2 {
		t.Errorf("test failed, expected: %d, result: %d", 2, len(response.Data))
	}
	if response.Data[0].Url != monitor1.Url {
		t.Errorf("test failed, expected: %s, result: %s", monitor1.Url, response.Data[0].Url)
	}
}

func TestApplication_ShowMonitor(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := models.User{
		Username: "testShowMonitor",
		Password: "passwordShowMonitor",
		Email:    "testShowMonitor@example.com",
	}
	err := app.DB.InsertUser(user)
	if err != nil {
		t.Fatalf("failed to insert user %v", err)
	}
	user, err = app.DB.GetUserByEmail(user.Email)
	if err != nil {
		t.Fatalf("failed to get user by email %v", err)
	}

	monitor1 := models.Monitor{
		UserID:        user.ID,
		Url:           "https://google.com",
		CheckInterval: 5,
	}
	monitor2 := models.Monitor{
		UserID:        user.ID,
		Url:           "https://yahoo.com",
		CheckInterval: 10,
	}
	err = app.DB.InsertMonitor(monitor1)
	if err != nil {
		t.Fatalf("failed to insert monitor %v", err)
	}
	err = app.DB.InsertMonitor(monitor2)
	if err != nil {
		t.Fatalf("failed to insert monitor %v", err)
	}
	monitors, err := app.DB.GetAllMonitorByUserID(user.ID)
	if err != nil {
		t.Fatalf("failed to get all monitor by user id")
	}

	targetID := monitors[0].ID
	url := fmt.Sprintf("/api/v1/monitors/%d", targetID)
	r := httptest.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	ctx = context.WithValue(r.Context(), contextKeyUserID, user.ID)
	r = r.WithContext(ctx)
	r.SetPathValue("id", fmt.Sprintf("%d", targetID))
	rr := httptest.NewRecorder()
	handlers := http.HandlerFunc(app.ShowMonitor)
	handlers.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("test failed, expected: %d, result: %d", http.StatusOK, rr.Code)
	}

	var response struct {
		Status  string         `json:"status"`
		Message string         `json:"message"`
		Data    models.Monitor `json:"data"`
	}
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response json %v", err)
	}

	if response.Status != "success" {
		t.Errorf("test failed, expected: success, result: %s", response.Status)
	}
	if response.Message != "get monitor" {
		t.Errorf("test failed, expected: get monitor, result: %s", response.Message)
	}

	if response.Data.ID != targetID {
		t.Errorf("test failed, expected: %d, result: %d", targetID, response.Data.ID)
	}

	if response.Data.Url != monitor1.Url {
		t.Errorf("test failed, expected: %s, result: %s", monitor1.Url, response.Data.Url)
	}
}

func TestApplication_UpdateMonitor(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := models.User{
		Username: "testUpdateMonitor",
		Password: "passwordUpdateMonitor",
		Email:    "testUpdateMonitor@example.com",
	}
	err := app.DB.InsertUser(user)
	if err != nil {
		t.Fatalf("failed to insert user %v", err)
	}
	user, err = app.DB.GetUserByEmail(user.Email)
	if err != nil {
		t.Fatalf("failed to get user by email %v", err)
	}

	monitor1 := models.Monitor{
		UserID:        user.ID,
		Url:           "https://google.com",
		CheckInterval: 5,
	}
	monitor2 := models.Monitor{
		UserID:        user.ID,
		Url:           "https://yahoo.com",
		CheckInterval: 10,
	}
	err = app.DB.InsertMonitor(monitor1)
	if err != nil {
		t.Fatalf("failed to insert monitor %v", err)
	}
	err = app.DB.InsertMonitor(monitor2)
	if err != nil {
		t.Fatalf("failed to insert monitor %v", err)
	}
	monitors, err := app.DB.GetAllMonitorByUserID(user.ID)
	if err != nil {
		t.Fatalf("failed to get all monitor by user id")
	}
	targetID := monitors[0].ID

	payload := models.Monitor{
		Url:           "https://hoyolab.com",
		CheckInterval: 5,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload json")
	}

	bodyReader := bytes.NewBuffer(jsonData)

	url := fmt.Sprintf("/api/v1/monitors/%d", targetID)
	r := httptest.NewRequestWithContext(ctx, http.MethodPatch, url, bodyReader)

	r.Header.Set("Content-Type", "application/json")
	ctx = context.WithValue(r.Context(), contextKeyUserID, user.ID)
	r = r.WithContext(ctx)
	r.SetPathValue("id", fmt.Sprintf("%d", targetID))
	rr := httptest.NewRecorder()
	handlers := http.HandlerFunc(app.UpdateMonitor)
	handlers.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("test failed, expected: %d, result: %d", http.StatusOK, rr.Code)
	}

	var response struct {
		Status  string         `json:"status"`
		Message string         `json:"message"`
		Data    models.Monitor `json:"data"`
	}
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response json %v", err)
	}

	if response.Status != "success" {
		t.Errorf("test failed, expected: success, result: %s", response.Status)
	}
	if response.Message != "monitor updated" {
		t.Errorf("test failed, expected: monitor updated, result: %s", response.Message)
	}

	if response.Data.ID != targetID {
		t.Errorf("test failed, expected: %d, result: %d", targetID, response.Data.ID)
	}

	if response.Data.Url != payload.Url {
		t.Errorf("test failed, expected: %s, result: %s", payload.Url, response.Data.Url)
	}
}

func TestApplication_DeleteMonitor(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := models.User{
		Username: "testDeleteMonitor",
		Password: "passwordDeleteMonitor",
		Email:    "testDeleteMonitor@example.com",
	}
	err := app.DB.InsertUser(user)
	if err != nil {
		t.Fatalf("failed to insert user %v", err)
	}
	user, err = app.DB.GetUserByEmail(user.Email)
	if err != nil {
		t.Fatalf("failed to get user by email %v", err)
	}

	monitor1 := models.Monitor{
		UserID:        user.ID,
		Url:           "https://google.com",
		CheckInterval: 5,
	}
	monitor2 := models.Monitor{
		UserID:        user.ID,
		Url:           "https://yahoo.com",
		CheckInterval: 10,
	}
	err = app.DB.InsertMonitor(monitor1)
	if err != nil {
		t.Fatalf("failed to insert monitor %v", err)
	}
	err = app.DB.InsertMonitor(monitor2)
	if err != nil {
		t.Fatalf("failed to insert monitor %v", err)
	}
	monitors, err := app.DB.GetAllMonitorByUserID(user.ID)
	if err != nil {
		t.Fatalf("failed to get all monitor by user id")
	}
	targetID := monitors[0].ID

	url := fmt.Sprintf("/api/v1/monitors/%d", targetID)
	r := httptest.NewRequestWithContext(ctx, http.MethodDelete, url, nil)

	ctx = context.WithValue(r.Context(), contextKeyUserID, user.ID)
	r = r.WithContext(ctx)
	r.SetPathValue("id", fmt.Sprintf("%d", targetID))
	rr := httptest.NewRecorder()
	handlers := http.HandlerFunc(app.DeleteMonitor)
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
	if response.Message != "monitor has been deleted" {
		t.Errorf("test failed, expected: monitor updated, result: %s", response.Message)
	}
}
