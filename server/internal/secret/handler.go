package secret

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/localization"
	"github.com/Adigezalov/goph-keeper/internal/middleware"
	"github.com/gorilla/mux"
)

type SecretService interface {
	CreateSecret(userID int, req *CreateSecretRequest, excludeSessionID string) (*Secret, error)
	GetSecret(id string, userID int) (*Secret, error)
	GetAllSecrets(userID int) ([]*Secret, error)
	UpdateSecret(id string, userID int, req *UpdateSecretRequest, excludeSessionID string) (*Secret, error)
	DeleteSecret(id string, userID int, excludeSessionID string) error
	GetSecretsForSync(userID int, since *time.Time) (*SyncResponse, error)
}

type Handler struct {
	service        SecretService
	chunkedService *ChunkedUploadService
}

func NewHandler(service SecretService) *Handler {
	return &Handler{
		service:        service,
		chunkedService: NewChunkedUploadService(),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		localization.LocalizedError(w, r, http.StatusUnauthorized, "common.authorization_error", nil)
		return
	}

	var excludeSessionID string
	if sessionID, ok := middleware.GetSessionIDFromContext(r.Context()); ok {
		excludeSessionID = sessionID
	}

	var req CreateSecretRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		localization.LocalizedError(w, r, http.StatusBadRequest, "common.invalid_request_format", nil)
		return
	}

	secret, err := h.service.CreateSecret(userID, &req, excludeSessionID)
	if err != nil {
		switch {
		case errors.Is(err, ErrLoginRequired),
			errors.Is(err, ErrPasswordRequired),
			errors.Is(err, ErrRequestRequired):
			localization.LocalizedError(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		default:
			log.Printf("Ошибка создания секрета: %v", err)
			localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(secret.ToResponse()); err != nil {
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		localization.LocalizedError(w, r, http.StatusUnauthorized, "common.authorization_error", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		localization.LocalizedError(w, r, http.StatusBadRequest, "secret.id_required", nil)
		return
	}

	secret, err := h.service.GetSecret(id, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrSecretNotFound):
			localization.LocalizedError(w, r, http.StatusNotFound, "secret.not_found", nil)
			return
		default:
			log.Printf("Ошибка получения секрета: %v", err)
			localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(secret.ToResponse()); err != nil {
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	secrets, err := h.service.GetAllSecrets(userID)
	if err != nil {
		log.Printf("Ошибка получения секретов: %v", err)
		localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
		return
	}

	response := make([]SecretResponse, 0, len(secrets))
	for _, secret := range secrets {
		response = append(response, secret.ToResponse())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	var excludeSessionID string
	if sessionID, ok := middleware.GetSessionIDFromContext(r.Context()); ok {
		excludeSessionID = sessionID
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		localization.LocalizedError(w, r, http.StatusBadRequest, "secret.id_required", nil)
		return
	}

	var req UpdateSecretRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		localization.LocalizedError(w, r, http.StatusBadRequest, "common.invalid_request_format", nil)
		return
	}

	secret, err := h.service.UpdateSecret(id, userID, &req, excludeSessionID)
	if err != nil {
		switch {
		case errors.Is(err, ErrSecretNotFound):
			localization.LocalizedError(w, r, http.StatusNotFound, "secret.not_found", nil)
			return
		case errors.Is(err, ErrVersionConflict):
			localization.LocalizedError(w, r, http.StatusConflict, "secret.version_conflict_detailed", nil)
			return
		case errors.Is(err, ErrLoginRequired),
			errors.Is(err, ErrPasswordRequired),
			errors.Is(err, ErrRequestRequired):
			localization.LocalizedError(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		default:
			log.Printf("Ошибка обновления секрета: %v", err)
			localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(secret.ToResponse()); err != nil {
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	var excludeSessionID string
	if sessionID, ok := middleware.GetSessionIDFromContext(r.Context()); ok {
		excludeSessionID = sessionID
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		localization.LocalizedError(w, r, http.StatusBadRequest, "secret.id_required", nil)
		return
	}

	err := h.service.DeleteSecret(id, userID, excludeSessionID)
	if err != nil {
		switch {
		case errors.Is(err, ErrSecretNotFound):
			localization.LocalizedError(w, r, http.StatusNotFound, "secret.not_found", nil)
			return
		default:
			log.Printf("Ошибка удаления секрета: %v", err)
			localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	var since *time.Time
	sinceStr := r.URL.Query().Get("since")
	if sinceStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			localization.LocalizedError(w, r, http.StatusBadRequest, "common.invalid_since_format", nil)
			return
		}
		since = &parsedTime
	}

	response, err := h.service.GetSecretsForSync(userID, since)
	if err != nil {
		log.Printf("Ошибка синхронизации секретов: %v", err)
		localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
		return
	}

	secretResponses := make([]SecretResponse, 0, len(response.Secrets))
	for _, secret := range response.Secrets {
		secretResponses = append(secretResponses, secret.ToResponseForSync())
	}

	syncResponse := struct {
		Secrets    []SecretResponse `json:"secrets"`
		ServerTime string           `json:"server_time"`
	}{
		Secrets:    secretResponses,
		ServerTime: response.ServerTime.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(syncResponse); err != nil {
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}

func (h *Handler) InitChunkedUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		localization.LocalizedError(w, r, http.StatusUnauthorized, "common.authorization_error", nil)
		return
	}

	var req InitChunkedUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		localization.LocalizedError(w, r, http.StatusBadRequest, "common.invalid_request_format", nil)
		return
	}

	session, err := h.chunkedService.InitUpload(fmt.Sprintf("%d", userID), req.TotalChunks, req.TotalSize)
	if err != nil {
		log.Printf("Ошибка инициализации chunked upload: %v", err)
		localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
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

func (h *Handler) UploadChunk(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		localization.LocalizedError(w, r, http.StatusUnauthorized, "common.authorization_error", nil)
		return
	}

	vars := mux.Vars(r)
	secretID := vars["id"]

	var req UploadChunkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		localization.LocalizedError(w, r, http.StatusBadRequest, "common.invalid_request_format", nil)
		return
	}

	session, err := h.chunkedService.GetSession(req.UploadID)
	if err != nil {
		localization.LocalizedError(w, r, http.StatusNotFound, "secret.session_not_found", nil)
		return
	}

	if session.UserID != fmt.Sprintf("%d", userID) {
		localization.LocalizedError(w, r, http.StatusForbidden, "secret.access_denied", nil)
		return
	}

	if session.SecretID != secretID {
		localization.LocalizedError(w, r, http.StatusBadRequest, "secret.invalid_secret_id", nil)
		return
	}

	if err := h.chunkedService.UploadChunk(req.UploadID, req.ChunkIndex, req.Data); err != nil {
		log.Printf("Ошибка загрузки чанка: %v", err)
		localization.LocalizedError(w, r, http.StatusInternalServerError, "secret.chunk_upload_error", nil)
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

func (h *Handler) FinalizeChunkedUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		localization.LocalizedError(w, r, http.StatusUnauthorized, "common.authorization_error", nil)
		return
	}

	vars := mux.Vars(r)
	secretID := vars["id"]

	var req FinalizeChunkedUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		localization.LocalizedError(w, r, http.StatusBadRequest, "common.invalid_request_format", nil)
		return
	}

	session, err := h.chunkedService.GetSession(req.UploadID)
	if err != nil {
		localization.LocalizedError(w, r, http.StatusNotFound, "secret.session_not_found_simple", nil)
		return
	}

	if session.UserID != fmt.Sprintf("%d", userID) {
		localization.LocalizedError(w, r, http.StatusForbidden, "secret.access_denied", nil)
		return
	}

	if session.SecretID != secretID {
		localization.LocalizedError(w, r, http.StatusBadRequest, "secret.invalid_secret_id", nil)
		return
	}

	binaryData, err := h.chunkedService.GetCompleteData(req.UploadID)
	if err != nil {
		log.Printf("Ошибка сборки чанков: %v", err)
		localization.LocalizedError(w, r, http.StatusBadRequest, "secret.chunks_not_complete", nil)
		return
	}

	metadata := make(map[string]interface{})
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	var excludeSessionID string
	if sessionID, ok := middleware.GetSessionIDFromContext(r.Context()); ok {
		excludeSessionID = sessionID
	}

	createReq := &CreateSecretRequest{
		Login:      req.Login,
		Password:   req.Password,
		Metadata:   metadata,
		BinaryData: binaryData,
	}

	var secret *Secret
	if req.Version != nil {
		updateReq := &UpdateSecretRequest{
			Login:      req.Login,
			Password:   req.Password,
			Metadata:   metadata,
			BinaryData: binaryData,
			Version:    *req.Version,
		}
		secret, err = h.service.UpdateSecret(secretID, userID, updateReq, excludeSessionID)
	} else {
		secret, err = h.service.CreateSecret(userID, createReq, excludeSessionID)
	}

	if err != nil {
		log.Printf("Ошибка создания/обновления секрета: %v", err)

		if errors.Is(err, ErrVersionConflict) {
			localization.LocalizedError(w, r, http.StatusConflict, "secret.version_conflict", nil)
			return
		}

		localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
		return
	}

	h.chunkedService.CleanupSession(req.UploadID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(secret.ToResponse())
}

func (h *Handler) DownloadChunk(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		localization.LocalizedError(w, r, http.StatusUnauthorized, "common.authorization_error", nil)
		return
	}

	vars := mux.Vars(r)
	secretID := vars["id"]
	chunkIndex := 0
	if _, err := fmt.Sscanf(vars["chunkIndex"], "%d", &chunkIndex); err != nil {
		localization.LocalizedError(w, r, http.StatusBadRequest, "secret.chunk_index_invalid", nil)
		return
	}

	secret, err := h.service.GetSecret(secretID, userID)
	if err != nil {
		if errors.Is(err, ErrSecretNotFound) {
			localization.LocalizedError(w, r, http.StatusNotFound, "secret.not_found", nil)
			return
		}
		log.Printf("Ошибка получения секрета: %v", err)
		localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
		return
	}

	const chunkSize = 100 * 1024
	chunks := SplitIntoChunks(secret.BinaryData, chunkSize)

	if chunkIndex < 0 || chunkIndex >= len(chunks) {
		localization.LocalizedError(w, r, http.StatusBadRequest, "secret.chunk_index_invalid", nil)
		return
	}

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
