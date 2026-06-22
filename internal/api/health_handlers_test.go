package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var app = &Application{}

func TestApplication_HealthTest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	// r = r.WithContext(getCtx(r))
	// r.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.HealthTest)
	handler.ServeHTTP(rr, r)

	if rr.Code != http.StatusOK {
		t.Errorf("test failed, expected: %d, result: %d", http.StatusOK, rr.Code)
	}

	var response jsonResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode request json: %v", err)
	}
	if response.Status != "success" {
		t.Errorf("test failed, expected: success, result: %s", response.Status)
	}
	if response.Message != "health tested" {
		t.Errorf("test failed. expected: health tested, result: %s", response.Message)
	}
}

func getCtx(r *http.Request) context.Context {
	return context.WithValue(r.Context(), contextKeyUserID, r.Header.Get("X-Session"))
}
