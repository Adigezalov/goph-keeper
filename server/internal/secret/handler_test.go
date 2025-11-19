package secret

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/middleware"
	"github.com/gorilla/mux"
)

type MockService struct {
	secrets map[string]*Secret
}

func NewMockService() *MockService {
	return &MockService{
		secrets: make(map[string]*Secret),
	}
}

func (m *MockService) CreateSecret(userID int, req *CreateSecretRequest, excludeSessionID string) (*Secret, error) {
	if req.Login == "" {
		return nil, ErrLoginRequired
	}
	if req.Password == "" {
		return nil, ErrPasswordRequired
	}

	secret := &Secret{
		ID:         "test-uuid-" + time.Now().Format("20060102150405"),
		UserID:     userID,
		Login:      req.Login,
		Password:   req.Password,
		Metadata:   req.Metadata,
		BinaryData: req.BinaryData,
		Version:    1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	m.secrets[secret.ID] = secret
	return secret, nil
}

func (m *MockService) GetSecret(id string, userID int) (*Secret, error) {
	secret, ok := m.secrets[id]
	if !ok || secret.UserID != userID {
		return nil, ErrSecretNotFound
	}
	return secret, nil
}

func (m *MockService) GetAllSecrets(userID int) ([]*Secret, error) {
	var result []*Secret
	for _, secret := range m.secrets {
		if secret.UserID == userID {
			result = append(result, secret)
		}
	}
	return result, nil
}

func (m *MockService) UpdateSecret(id string, userID int, req *UpdateSecretRequest, excludeSessionID string) (*Secret, error) {
	if req.Login == "" {
		return nil, ErrLoginRequired
	}
	if req.Password == "" {
		return nil, ErrPasswordRequired
	}

	secret, ok := m.secrets[id]
	if !ok || secret.UserID != userID {
		return nil, ErrSecretNotFound
	}
	if secret.Version != req.Version {
		return nil, ErrVersionConflict
	}

	secret.Login = req.Login
	secret.Password = req.Password
	secret.Metadata = req.Metadata
	secret.BinaryData = req.BinaryData
	secret.Version++
	secret.UpdatedAt = time.Now()

	return secret, nil
}

func (m *MockService) DeleteSecret(id string, userID int, excludeSessionID string) error {
	secret, ok := m.secrets[id]
	if !ok || secret.UserID != userID {
		return ErrSecretNotFound
	}
	delete(m.secrets, id)
	return nil
}

func (m *MockService) GetSecretsForSync(userID int, since *time.Time) (*SyncResponse, error) {
	var result []*Secret
	for _, secret := range m.secrets {
		if secret.UserID == userID {
			if since == nil || secret.UpdatedAt.After(*since) {
				result = append(result, secret)
			}
		}
	}
	return &SyncResponse{
		Secrets:    result,
		ServerTime: time.Now(),
	}, nil
}

func addUserIDToContext(r *http.Request, userID int) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	return r.WithContext(ctx)
}

func TestHandler_Create(t *testing.T) {
	service := NewMockService()
	handler := NewHandler(service)

	reqBody := CreateSecretRequest{
		Login:    "encrypted_login",
		Password: "encrypted_password",
		Metadata: map[string]interface{}{"app": "github"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", bytes.NewReader(body))
	req = addUserIDToContext(req, 1)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response SecretResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Login != reqBody.Login {
		t.Errorf("Expected login %s, got %s", reqBody.Login, response.Login)
	}
}

func TestHandler_Create_Validation(t *testing.T) {
	service := NewMockService()
	handler := NewHandler(service)

	reqBody := CreateSecretRequest{
		Login:    "",
		Password: "password",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", bytes.NewReader(body))
	req = addUserIDToContext(req, 1)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandler_Get(t *testing.T) {
	service := NewMockService()
	handler := NewHandler(service)

	secret, _ := service.CreateSecret(1, &CreateSecretRequest{
		Login:    "login",
		Password: "password",
	}, "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets/"+secret.ID, nil)
	req = addUserIDToContext(req, 1)
	req = mux.SetURLVars(req, map[string]string{"id": secret.ID})
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response SecretResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != secret.ID {
		t.Errorf("Expected ID %s, got %s", secret.ID, response.ID)
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	service := NewMockService()
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets/nonexistent", nil)
	req = addUserIDToContext(req, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistent"})
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandler_GetAll(t *testing.T) {
	service := NewMockService()
	handler := NewHandler(service)

	secret, _ := service.CreateSecret(1, &CreateSecretRequest{Login: "login1", Password: "pass1"}, "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)
	req = addUserIDToContext(req, 1)
	w := httptest.NewRecorder()

	handler.GetAll(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []SecretResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	found := false
	for _, s := range response {
		if s.ID == secret.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Created secret not found in response")
	}
}

func TestHandler_Update(t *testing.T) {
	service := NewMockService()
	handler := NewHandler(service)

	secret, _ := service.CreateSecret(1, &CreateSecretRequest{
		Login:    "old_login",
		Password: "old_password",
	}, "")

	reqBody := UpdateSecretRequest{
		Login:    "new_login",
		Password: "new_password",
		Version:  secret.Version,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/secrets/"+secret.ID, bytes.NewReader(body))
	req = addUserIDToContext(req, 1)
	req = mux.SetURLVars(req, map[string]string{"id": secret.ID})
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response SecretResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Login != reqBody.Login {
		t.Errorf("Expected login %s, got %s", reqBody.Login, response.Login)
	}

	if response.Version != 2 {
		t.Errorf("Expected version 2, got %d", response.Version)
	}
}

func TestHandler_Update_VersionConflict(t *testing.T) {
	service := NewMockService()
	handler := NewHandler(service)

	secret, _ := service.CreateSecret(1, &CreateSecretRequest{
		Login:    "login",
		Password: "password",
	}, "")

	reqBody := UpdateSecretRequest{
		Login:    "new_login",
		Password: "new_password",
		Version:  999,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/secrets/"+secret.ID, bytes.NewReader(body))
	req = addUserIDToContext(req, 1)
	req = mux.SetURLVars(req, map[string]string{"id": secret.ID})
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestHandler_Delete(t *testing.T) {
	service := NewMockService()
	handler := NewHandler(service)

	secret, _ := service.CreateSecret(1, &CreateSecretRequest{
		Login:    "login",
		Password: "password",
	}, "")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/secrets/"+secret.ID, nil)
	req = addUserIDToContext(req, 1)
	req = mux.SetURLVars(req, map[string]string{"id": secret.ID})
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestHandler_Sync(t *testing.T) {
	service := NewMockService()
	handler := NewHandler(service)

	secret, _ := service.CreateSecret(1, &CreateSecretRequest{Login: "login1", Password: "pass1"}, "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets/sync", nil)
	req = addUserIDToContext(req, 1)
	w := httptest.NewRecorder()

	handler.Sync(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response struct {
		Secrets    []SecretResponse `json:"secrets"`
		ServerTime string           `json:"server_time"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	found := false
	for _, s := range response.Secrets {
		if s.ID == secret.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Created secret not found in sync response")
	}

	if response.ServerTime == "" {
		t.Error("Expected ServerTime to be set")
	}
}

func TestHandler_Sync_WithSince(t *testing.T) {
	service := NewMockService()
	handler := NewHandler(service)

	service.CreateSecret(1, &CreateSecretRequest{Login: "login1", Password: "pass1"}, "")

	futureTime := time.Now().Add(1 * time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets/sync", nil)
	q := req.URL.Query()
	q.Add("since", futureTime.Format(time.RFC3339))
	req.URL.RawQuery = q.Encode()
	req = addUserIDToContext(req, 1)
	w := httptest.NewRecorder()

	handler.Sync(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response struct {
		Secrets    []SecretResponse `json:"secrets"`
		ServerTime string           `json:"server_time"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Secrets) != 0 {
		t.Errorf("Expected 0 secrets, got %d", len(response.Secrets))
	}
}
