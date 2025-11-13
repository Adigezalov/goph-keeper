package secret

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/middleware"
	"github.com/gorilla/mux"
)

// SecretService интерфейс для бизнес-логики секретов
type SecretService interface {
	CreateSecret(userID int, req *CreateSecretRequest) (*Secret, error)
	GetSecret(id string, userID int) (*Secret, error)
	GetAllSecrets(userID int) ([]*Secret, error)
	UpdateSecret(id string, userID int, req *UpdateSecretRequest) (*Secret, error)
	DeleteSecret(id string, userID int) error
	GetSecretsForSync(userID int, since *time.Time) (*SyncResponse, error)
}

// Handler обрабатывает HTTP запросы для секретов
type Handler struct {
	service SecretService
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service SecretService) *Handler {
	return &Handler{
		service: service,
	}
}

// Create обрабатывает POST /api/v1/secrets
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста (добавлен auth middleware)
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	var req CreateSecretRequest

	// Декодируем JSON запрос
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Создаем секрет
	secret, err := h.service.CreateSecret(userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrLoginRequired),
			errors.Is(err, ErrPasswordRequired),
			errors.Is(err, ErrRequestRequired):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			log.Printf("Ошибка создания секрета: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
	}

	// Возвращаем созданный секрет
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(secret.ToResponse()); err != nil {
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}

// Get обрабатывает GET /api/v1/secrets/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	// Получаем ID секрета из URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "ID секрета обязателен", http.StatusBadRequest)
		return
	}

	// Получаем секрет
	secret, err := h.service.GetSecret(id, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrSecretNotFound):
			http.Error(w, "Секрет не найден", http.StatusNotFound)
			return
		default:
			log.Printf("Ошибка получения секрета: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
	}

	// Возвращаем секрет
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(secret.ToResponse()); err != nil {
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}

// GetAll обрабатывает GET /api/v1/secrets
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	// Получаем все секреты пользователя
	secrets, err := h.service.GetAllSecrets(userID)
	if err != nil {
		log.Printf("Ошибка получения секретов: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Конвертируем в response
	response := make([]SecretResponse, 0, len(secrets))
	for _, secret := range secrets {
		response = append(response, secret.ToResponse())
	}

	// Возвращаем список секретов
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}

// Update обрабатывает PUT /api/v1/secrets/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	// Получаем ID секрета из URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "ID секрета обязателен", http.StatusBadRequest)
		return
	}

	var req UpdateSecretRequest

	// Декодируем JSON запрос
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Обновляем секрет
	secret, err := h.service.UpdateSecret(id, userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrSecretNotFound):
			http.Error(w, "Секрет не найден", http.StatusNotFound)
			return
		case errors.Is(err, ErrVersionConflict):
			http.Error(w, "Конфликт версий: секрет был изменен на другом устройстве", http.StatusConflict)
			return
		case errors.Is(err, ErrLoginRequired),
			errors.Is(err, ErrPasswordRequired),
			errors.Is(err, ErrRequestRequired):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			log.Printf("Ошибка обновления секрета: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
	}

	// Возвращаем обновленный секрет
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(secret.ToResponse()); err != nil {
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}

// Delete обрабатывает DELETE /api/v1/secrets/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	// Получаем ID секрета из URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "ID секрета обязателен", http.StatusBadRequest)
		return
	}

	// Удаляем секрет
	err := h.service.DeleteSecret(id, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrSecretNotFound):
			http.Error(w, "Секрет не найден", http.StatusNotFound)
			return
		default:
			log.Printf("Ошибка удаления секрета: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
	}

	// Возвращаем успешный ответ без тела
	w.WriteHeader(http.StatusNoContent)
}

// Sync обрабатывает GET /api/v1/secrets/sync?since=<timestamp>
// Возвращает все секреты для синхронизации
func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	// Получаем параметр since из query
	var since *time.Time
	sinceStr := r.URL.Query().Get("since")
	if sinceStr != "" {
		// Парсим timestamp в формате RFC3339
		parsedTime, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			http.Error(w, "Неверный формат параметра since (ожидается RFC3339)", http.StatusBadRequest)
			return
		}
		since = &parsedTime
	}

	// Получаем секреты для синхронизации
	response, err := h.service.GetSecretsForSync(userID, since)
	if err != nil {
		log.Printf("Ошибка синхронизации секретов: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Конвертируем секреты в response
	secretResponses := make([]SecretResponse, 0, len(response.Secrets))
	for _, secret := range response.Secrets {
		secretResponses = append(secretResponses, secret.ToResponse())
	}

	// Формируем ответ для синхронизации
	syncResponse := struct {
		Secrets    []SecretResponse `json:"secrets"`
		ServerTime string           `json:"server_time"` // RFC3339 формат
	}{
		Secrets:    secretResponses,
		ServerTime: response.ServerTime.Format(time.RFC3339),
	}

	// Возвращаем результат синхронизации
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(syncResponse); err != nil {
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}
