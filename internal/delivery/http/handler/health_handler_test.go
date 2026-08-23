package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	h := NewHealthHandler()
	r.GET("/api/v1/health", h.Health)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}

	if !body.Success {
		t.Errorf("expected success=true, got %v", body.Success)
	}
	if body.Message != "System operational" {
		t.Errorf("expected message 'System operational', got %q", body.Message)
	}
	if body.Data.Status != "UP" {
		t.Errorf("expected data.status='UP', got %q", body.Data.Status)
	}
}
