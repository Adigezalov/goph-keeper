package health

import (
	"encoding/json"
	"log"
	"net/http"
)

// HealthService интерфейс для бизнес-логики health check
type HealthService interface {
	CheckHealth() bool
}

// Handler обрабатывает HTTP запросы для health check
type Handler struct {
	service HealthService
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service HealthService) *Handler {
	return &Handler{
		service: service,
	}
}

// HealthResponse представляет ответ health check
type HealthResponse struct {
	Status string `json:"status"`
}

// Check обрабатывает GET /api/health
// Возвращает 200 OK если сервер доступен
func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	// Проверяем состояние сервера
	isHealthy := h.service.CheckHealth()

	if !isHealthy {
		// Если сервер не здоров, возвращаем 503 Service Unavailable
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(HealthResponse{
			Status: "unavailable",
		}); err != nil {
			log.Printf("Ошибка отправки JSON ответа: %v", err)
		}
		return
	}

	// Сервер здоров, возвращаем 200 OK
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(HealthResponse{
		Status: "ok",
	}); err != nil {
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}
