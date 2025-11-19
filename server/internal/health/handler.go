package health

import (
	"encoding/json"
	"log"
	"net/http"
)

type HealthService interface {
	CheckHealth() bool
}

type Handler struct {
	service HealthService
}

func NewHandler(service HealthService) *Handler {
	return &Handler{
		service: service,
	}
}

type HealthResponse struct {
	Status string `json:"status"`
}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	isHealthy := h.service.CheckHealth()

	if !isHealthy {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(HealthResponse{
			Status: "unavailable",
		}); err != nil {
			log.Printf("Ошибка отправки JSON ответа: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(HealthResponse{
		Status: "ok",
	}); err != nil {
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}
