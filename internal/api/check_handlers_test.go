package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"uptime-monitor/internal/models"
)

func TestApplication_CheckMonitor(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := models.User{
		Username: "testCheckMonitor",
		Password: "passwordCheckMonitor",
		Email:    "testCheckMonitor@example.com",
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
		t.Fatalf("failed to get all monitor by user id %v", err)
	}
	targetID := monitors[0].ID

	check1 := models.Check{
		MonitorID:    targetID,
		StatusCode:   200,
		ResponseTime: 15,
	}
	err = app.DB.InsertCheck(check1.MonitorID, check1.StatusCode, check1.ResponseTime)
	if err != nil {
		t.Fatalf("failed to insert check monitor %v", err)
	}
	check2 := models.Check{
		MonitorID:    targetID,
		StatusCode:   500,
		ResponseTime: 500,
	}
	err = app.DB.InsertCheck(check2.MonitorID, check2.StatusCode, check2.ResponseTime)
	if err != nil {
		t.Fatalf("failed to insert check monitor %v", err)
	}

	checks, err := app.DB.GetChecksByMonitorID(targetID)
	if err != nil {
		t.Fatalf("failed to get checks by monitor id %v", err)
	}

	url := fmt.Sprintf("/monitor/%d/checks", targetID)
	r := httptest.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	r.SetPathValue("id", fmt.Sprintf("%d", targetID))
	ctx = context.WithValue(r.Context(), contextKeyUserID, user.ID)
	r = r.WithContext(ctx)
	rr := httptest.NewRecorder()
	handlers := http.HandlerFunc(app.CheckMonitor)
	handlers.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("test failed, expected: %d, result: %d", http.StatusOK, rr.Code)
	}

	var response struct {
		Status  string         `json:"status"`
		Message string         `json:"message"`
		Data    []models.Check `json:"data"`
	}

	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode json response %v", err)
	}

	if response.Status != "success" {
		t.Errorf("test failed, expected: success, result: %s", response.Status)
	}
	if response.Message != "get all checks for this url" {
		t.Errorf("test failed, expected: get all checks for this url, result: %s", response.Message)
	}

	if len(response.Data) != len(checks) {
		t.Errorf("test failed, expected: %d, result: %d", len(checks), len(response.Data))
	}
	if response.Data[0].MonitorID != checks[0].MonitorID {
		t.Errorf("test failed, expected %d, result: %d", checks[0].MonitorID, response.Data[0].MonitorID)
	}
	if response.Data[0].StatusCode != checks[0].StatusCode {
		t.Errorf("test failed, expected: %d, result: %d", checks[0].StatusCode, response.Data[0].StatusCode)
	}
}

func TestApplication_ShowMonitorStats(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := models.User{
		Username: "testShowMonitorStats",
		Password: "passwordShowMonitorStats",
		Email:    "testShowMonitorStats@example.com",
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
		t.Fatalf("failed to get all monitor by user id %v", err)
	}
	targetID := monitors[0].ID

	check1 := models.Check{
		MonitorID:    targetID,
		StatusCode:   200,
		ResponseTime: 15,
	}
	err = app.DB.InsertCheck(check1.MonitorID, check1.StatusCode, check1.ResponseTime)
	if err != nil {
		t.Fatalf("failed to insert check monitor %v", err)
	}
	check2 := models.Check{
		MonitorID:    targetID,
		StatusCode:   500,
		ResponseTime: 500,
	}
	err = app.DB.InsertCheck(check2.MonitorID, check2.StatusCode, check2.ResponseTime)
	if err != nil {
		t.Fatalf("failed to insert check monitor %v", err)
	}

	stats, err := app.DB.GetMonitorStatsById(targetID)
	if err != nil {
		t.Fatalf("failed to get monitor stats by id %v", err)
	}

	url := fmt.Sprintf("/monitor/%d/stats", targetID)
	r := httptest.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	r.SetPathValue("id", fmt.Sprintf("%d", targetID))
	ctx = context.WithValue(r.Context(), contextKeyUserID, user.ID)
	r = r.WithContext(ctx)
	rr := httptest.NewRecorder()
	handlers := http.HandlerFunc(app.ShowMonitorStats)
	handlers.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("test failed, expected: %d, result: %d", http.StatusOK, rr.Code)
	}

	var response struct {
		Status  string              `json:"status"`
		Message string              `json:"message"`
		Data    models.MonitorStats `json:"data"`
	}

	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode json response, %v", err)
	}

	if response.Status != "success" {
		t.Errorf("test failed, expected: success, result: %s", response.Status)
	}
	if response.Message != "get monitor stats" {
		t.Errorf("test failed, expected: get monitor stats, result: %s", response.Message)
	}

	if response.Data.AvgResponseTime != stats.AvgResponseTime {
		t.Errorf("test failed, expected: %f, result: %f", stats.AvgResponseTime, response.Data.AvgResponseTime)
	}
	if response.Data.UptimePercentage != stats.UptimePercentage {
		t.Errorf("test failed, expected: %f, result: %f", stats.UptimePercentage, response.Data.UptimePercentage)
	}
}
