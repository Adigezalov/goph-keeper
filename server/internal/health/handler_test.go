package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockHealthService struct {
	isHealthy bool
}

func (m *MockHealthService) CheckHealth() bool {
	return m.isHealthy
}

func TestHandler_Check_Healthy(t *testing.T) {
	mockService := &MockHealthService{isHealthy: true}
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	handler.Check(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ожидали статус %d, получили %d", http.StatusOK, w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("не удалось декодировать ответ: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("ожидали статус 'ok', получили '%s'", response.Status)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("ожидали Content-Type 'application/json', получили '%s'", contentType)
	}
}

func TestHandler_Check_Unhealthy(t *testing.T) {
	mockService := &MockHealthService{isHealthy: false}
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	handler.Check(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ожидали статус %d, получили %d", http.StatusServiceUnavailable, w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("не удалось декодировать ответ: %v", err)
	}

	if response.Status != "unavailable" {
		t.Errorf("ожидали статус 'unavailable', получили '%s'", response.Status)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("ожидали Content-Type 'application/json', получили '%s'", contentType)
	}
}
