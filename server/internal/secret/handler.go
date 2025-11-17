package secret

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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
	service        SecretService
	chunkedService *ChunkedUploadService
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service SecretService) *Handler {
	return &Handler{
		service:        service,
		chunkedService: NewChunkedUploadService(),
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

	// Конвертируем секреты в response для синхронизации
	// Для больших файлов (>1MB) не отправляем binary_data, клиент скачает чанками
	secretResponses := make([]SecretResponse, 0, len(response.Secrets))
	for _, secret := range response.Secrets {
		secretResponses = append(secretResponses, secret.ToResponseForSync())
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

// InitChunkedUpload обрабатывает POST /api/v1/secrets/chunks/init
func (h *Handler) InitChunkedUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	var req InitChunkedUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Инициализируем сессию
	session, err := h.chunkedService.InitUpload(fmt.Sprintf("%d", userID), req.TotalChunks, req.TotalSize)
	if err != nil {
		log.Printf("Ошибка инициализации chunked upload: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	response := InitChunkedUploadResponse{
		UploadID: session.UploadID,
		SecretID: session.SecretID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UploadChunk обрабатывает POST /api/v1/secrets/:id/chunks
func (h *Handler) UploadChunk(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	secretID := vars["id"]

	var req UploadChunkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Проверяем, что сессия принадлежит этому пользователю
	session, err := h.chunkedService.GetSession(req.UploadID)
	if err != nil {
		http.Error(w, "Сессия не найдена или истекла", http.StatusNotFound)
		return
	}

	if session.UserID != fmt.Sprintf("%d", userID) {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	if session.SecretID != secretID {
		http.Error(w, "Неверный ID секрета", http.StatusBadRequest)
		return
	}

	// Загружаем чанк
	if err := h.chunkedService.UploadChunk(req.UploadID, req.ChunkIndex, req.Data); err != nil {
		log.Printf("Ошибка загрузки чанка: %v", err)
		http.Error(w, "Ошибка загрузки чанка", http.StatusInternalServerError)
		return
	}

	response := UploadChunkResponse{
		ChunkIndex: req.ChunkIndex,
		Received:   true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// FinalizeChunkedUpload обрабатывает POST /api/v1/secrets/:id/chunks/finalize
func (h *Handler) FinalizeChunkedUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	secretID := vars["id"]

	var req FinalizeChunkedUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Проверяем сессию
	session, err := h.chunkedService.GetSession(req.UploadID)
	if err != nil {
		http.Error(w, "Сессия не найдена", http.StatusNotFound)
		return
	}

	if session.UserID != fmt.Sprintf("%d", userID) {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	if session.SecretID != secretID {
		http.Error(w, "Неверный ID секрета", http.StatusBadRequest)
		return
	}

	// Собираем все чанки
	binaryData, err := h.chunkedService.GetCompleteData(req.UploadID)
	if err != nil {
		log.Printf("Ошибка сборки чанков: %v", err)
		http.Error(w, "Не все чанки загружены", http.StatusBadRequest)
		return
	}

	// Конвертируем metadata из map[string]string в map[string]interface{}
	metadata := make(map[string]interface{})
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	// Создаем или обновляем секрет
	createReq := &CreateSecretRequest{
		Login:      req.Login,
		Password:   req.Password,
		Metadata:   metadata,
		BinaryData: binaryData,
	}

	var secret *Secret
	if req.Version != nil {
		// Обновление существующего секрета
		updateReq := &UpdateSecretRequest{
			Login:      req.Login,
			Password:   req.Password,
			Metadata:   metadata,
			BinaryData: binaryData,
			Version:    *req.Version,
		}
		secret, err = h.service.UpdateSecret(secretID, userID, updateReq)
	} else {
		// Создание нового секрета
		secret, err = h.service.CreateSecret(userID, createReq)
	}

	if err != nil {
		log.Printf("Ошибка создания/обновления секрета: %v", err)

		// Проверяем на конфликт версий
		if errors.Is(err, ErrVersionConflict) {
			http.Error(w, "Конфликт версий", http.StatusConflict)
			return
		}

		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Очищаем сессию
	h.chunkedService.CleanupSession(req.UploadID)

	// Возвращаем созданный секрет
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(secret.ToResponse())
}

// DownloadChunk обрабатывает GET /api/v1/secrets/:id/chunks/:chunkIndex
func (h *Handler) DownloadChunk(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	secretID := vars["id"]
	chunkIndex := 0
	if _, err := fmt.Sscanf(vars["chunkIndex"], "%d", &chunkIndex); err != nil {
		http.Error(w, "Неверный индекс чанка", http.StatusBadRequest)
		return
	}

	// Получаем секрет
	secret, err := h.service.GetSecret(secretID, userID)
	if err != nil {
		if errors.Is(err, ErrSecretNotFound) {
			http.Error(w, "Секрет не найден", http.StatusNotFound)
			return
		}
		log.Printf("Ошибка получения секрета: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Разбиваем binary data на чанки
	const chunkSize = 100 * 1024 // 100 KB
	chunks := SplitIntoChunks(secret.BinaryData, chunkSize)

	if chunkIndex < 0 || chunkIndex >= len(chunks) {
		http.Error(w, "Неверный индекс чанка", http.StatusBadRequest)
		return
	}

	// Кодируем чанк в base64
	chunkData := chunks[chunkIndex]
	base64Data := base64.StdEncoding.EncodeToString(chunkData)

	response := DownloadChunkResponse{
		ChunkIndex:  chunkIndex,
		Data:        base64Data,
		TotalChunks: len(chunks),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
