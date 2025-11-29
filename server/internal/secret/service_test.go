package secret

import (
	"fmt"
	"testing"
	"time"
)

type MockRepository struct {
	secrets   map[string]*Secret
	idCounter int
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		secrets:   make(map[string]*Secret),
		idCounter: 0,
	}
}

func (m *MockRepository) CreateSecret(secret *Secret) error {
	m.idCounter++
	secret.ID = fmt.Sprintf("test-uuid-%d", m.idCounter)
	secret.CreatedAt = time.Now()
	secret.UpdatedAt = time.Now()
	m.secrets[secret.ID] = secret
	return nil
}

func (m *MockRepository) GetSecretByID(id string, userID int) (*Secret, error) {
	secret, ok := m.secrets[id]
	if !ok || secret.UserID != userID {
		return nil, ErrSecretNotFound
	}
	return secret, nil
}

func (m *MockRepository) GetSecretsByUserID(userID int) ([]*Secret, error) {
	var result []*Secret
	for _, secret := range m.secrets {
		if secret.UserID == userID && !secret.DeletedAt.Valid {
			result = append(result, secret)
		}
	}
	return result, nil
}

func (m *MockRepository) GetSecretsModifiedSince(userID int, since time.Time) ([]*Secret, error) {
	var result []*Secret
	for _, secret := range m.secrets {
		if secret.UserID == userID && secret.UpdatedAt.After(since) {
			result = append(result, secret)
		}
	}
	return result, nil
}

func (m *MockRepository) UpdateSecret(secret *Secret) error {
	existing, ok := m.secrets[secret.ID]
	if !ok || existing.UserID != secret.UserID {
		return ErrSecretNotFound
	}
	if existing.Version != secret.Version {
		return ErrVersionConflict
	}
	secret.Version++
	secret.UpdatedAt = time.Now()
	m.secrets[secret.ID] = secret
	return nil
}

func (m *MockRepository) SoftDeleteSecret(id string, userID int) error {
	secret, ok := m.secrets[id]
	if !ok || secret.UserID != userID {
		return ErrSecretNotFound
	}
	secret.DeletedAt.Valid = true
	secret.DeletedAt.Time = time.Now()
	return nil
}

func TestService_CreateSecret(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	req := &CreateSecretRequest{
		Login:    "encrypted_login",
		Password: "encrypted_password",
		Metadata: map[string]interface{}{"app": "github"},
	}

	secret, err := service.CreateSecret(1, req, "")
	if err != nil {
		t.Fatalf("CreateSecret failed: %v", err)
	}

	if secret.ID == "" {
		t.Error("Expected secret ID to be set")
	}

	if secret.Version != 1 {
		t.Errorf("Expected version 1, got %d", secret.Version)
	}

	if secret.Login != req.Login {
		t.Errorf("Expected login %s, got %s", req.Login, secret.Login)
	}
}

func TestService_CreateSecret_Validation(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	tests := []struct {
		name    string
		req     *CreateSecretRequest
		wantErr error
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: ErrRequestRequired,
		},
		{
			name: "empty login",
			req: &CreateSecretRequest{
				Login:    "",
				Password: "password",
			},
			wantErr: ErrLoginRequired,
		},
		{
			name: "empty password",
			req: &CreateSecretRequest{
				Login:    "login",
				Password: "",
			},
			wantErr: ErrPasswordRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateSecret(1, tt.req, "")
			if err != tt.wantErr {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestService_GetSecretsForSync(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	req := &CreateSecretRequest{
		Login:    "encrypted_login",
		Password: "encrypted_password",
	}
	secret, _ := service.CreateSecret(1, req, "")

	resp, err := service.GetSecretsForSync(1, nil)
	if err != nil {
		t.Fatalf("GetSecretsForSync failed: %v", err)
	}

	if len(resp.Secrets) != 1 {
		t.Errorf("Expected 1 secret, got %d", len(resp.Secrets))
	}

	if resp.ServerTime.IsZero() {
		t.Error("Expected ServerTime to be set")
	}

	since := secret.CreatedAt.Add(-1 * time.Second)
	resp2, err := service.GetSecretsForSync(1, &since)
	if err != nil {
		t.Fatalf("GetSecretsForSync failed: %v", err)
	}

	if len(resp2.Secrets) != 1 {
		t.Errorf("Expected 1 modified secret, got %d", len(resp2.Secrets))
	}
}

func TestService_UpdateSecret_VersionConflict(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	createReq := &CreateSecretRequest{
		Login:    "login",
		Password: "password",
	}
	secret, _ := service.CreateSecret(1, createReq, "")

	updateReq := &UpdateSecretRequest{
		Login:    "new_login",
		Password: "new_password",
		Version:  999,
	}

	_, err := service.UpdateSecret(secret.ID, 1, updateReq, "")
	if err != ErrVersionConflict {
		t.Errorf("Expected ErrVersionConflict, got %v", err)
	}
}

func TestService_GetSecret(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	createReq := &CreateSecretRequest{
		Login:    "test_login",
		Password: "test_password",
		Metadata: map[string]interface{}{"key": "value"},
	}
	created, _ := service.CreateSecret(1, createReq, "")

	secret, err := service.GetSecret(created.ID, 1)
	if err != nil {
		t.Fatalf("GetSecret failed: %v", err)
	}

	if secret.ID != created.ID {
		t.Errorf("Expected secret ID %s, got %s", created.ID, secret.ID)
	}
	if secret.Login != createReq.Login {
		t.Errorf("Expected login %s, got %s", createReq.Login, secret.Login)
	}
}

func TestService_GetSecret_NotFound(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	_, err := service.GetSecret("nonexistent", 1)
	if err != ErrSecretNotFound {
		t.Errorf("Expected ErrSecretNotFound, got %v", err)
	}
}

func TestService_GetAllSecrets(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	service.CreateSecret(1, &CreateSecretRequest{Login: "login1", Password: "pass1"}, "")
	service.CreateSecret(1, &CreateSecretRequest{Login: "login2", Password: "pass2"}, "")
	service.CreateSecret(2, &CreateSecretRequest{Login: "login3", Password: "pass3"}, "")

	secrets, err := service.GetAllSecrets(1)
	if err != nil {
		t.Fatalf("GetAllSecrets failed: %v", err)
	}

	if len(secrets) != 2 {
		t.Errorf("Expected 2 secrets for user 1, got %d", len(secrets))
	}
}

func TestService_GetAllSecrets_Empty(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	secrets, err := service.GetAllSecrets(999)
	if err != nil {
		t.Fatalf("GetAllSecrets failed: %v", err)
	}

	if len(secrets) != 0 {
		t.Errorf("Expected 0 secrets, got %d", len(secrets))
	}
}

func TestService_UpdateSecret(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	createReq := &CreateSecretRequest{
		Login:    "old_login",
		Password: "old_password",
	}
	secret, _ := service.CreateSecret(1, createReq, "")

	updateReq := &UpdateSecretRequest{
		Login:    "new_login",
		Password: "new_password",
		Metadata: map[string]interface{}{"updated": true},
		Version:  secret.Version,
	}

	updated, err := service.UpdateSecret(secret.ID, 1, updateReq, "")
	if err != nil {
		t.Fatalf("UpdateSecret failed: %v", err)
	}

	if updated.Login != updateReq.Login {
		t.Errorf("Expected login %s, got %s", updateReq.Login, updated.Login)
	}
	if updated.Version != 2 {
		t.Errorf("Expected version 2, got %d", updated.Version)
	}
}

func TestService_UpdateSecret_NotFound(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	updateReq := &UpdateSecretRequest{
		Login:    "login",
		Password: "password",
		Version:  1,
	}

	_, err := service.UpdateSecret("nonexistent", 1, updateReq, "")
	if err != ErrSecretNotFound {
		t.Errorf("Expected ErrSecretNotFound, got %v", err)
	}
}

func TestService_UpdateSecret_Validation(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	tests := []struct {
		name    string
		req     *UpdateSecretRequest
		wantErr error
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: ErrRequestRequired,
		},
		{
			name: "empty login",
			req: &UpdateSecretRequest{
				Login:    "",
				Password: "password",
				Version:  1,
			},
			wantErr: ErrLoginRequired,
		},
		{
			name: "empty password",
			req: &UpdateSecretRequest{
				Login:    "login",
				Password: "",
				Version:  1,
			},
			wantErr: ErrPasswordRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.UpdateSecret("some-id", 1, tt.req, "")
			if err != tt.wantErr {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestService_DeleteSecret(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	createReq := &CreateSecretRequest{
		Login:    "login",
		Password: "password",
	}
	secret, _ := service.CreateSecret(1, createReq, "")

	err := service.DeleteSecret(secret.ID, 1, "")
	if err != nil {
		t.Fatalf("DeleteSecret failed: %v", err)
	}

	secrets, _ := service.GetAllSecrets(1)
	if len(secrets) != 0 {
		t.Errorf("Expected 0 secrets after deletion, got %d", len(secrets))
	}
}

func TestService_DeleteSecret_NotFound(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	err := service.DeleteSecret("nonexistent", 1, "")
	if err != ErrSecretNotFound {
		t.Errorf("Expected ErrSecretNotFound, got %v", err)
	}
}

func TestService_WithRealtimeService(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	mockRealtime := &MockRealtimeService{}
	service.SetRealtimeService(mockRealtime)

	createReq := &CreateSecretRequest{
		Login:    "login",
		Password: "password",
	}
	secret, err := service.CreateSecret(1, createReq, "session-123")
	if err != nil {
		t.Fatalf("CreateSecret failed: %v", err)
	}

	if !mockRealtime.CreatedCalled {
		t.Error("Expected NotifySecretCreated to be called")
	}
	if mockRealtime.LastExcludeSessionID != "session-123" {
		t.Errorf("Expected exclude session ID 'session-123', got '%s'", mockRealtime.LastExcludeSessionID)
	}

	updateReq := &UpdateSecretRequest{
		Login:    "new_login",
		Password: "new_password",
		Version:  secret.Version,
	}
	_, err = service.UpdateSecret(secret.ID, 1, updateReq, "session-456")
	if err != nil {
		t.Fatalf("UpdateSecret failed: %v", err)
	}
	if !mockRealtime.UpdatedCalled {
		t.Error("Expected NotifySecretUpdated to be called")
	}

	err = service.DeleteSecret(secret.ID, 1, "session-789")
	if err != nil {
		t.Fatalf("DeleteSecret failed: %v", err)
	}
	if !mockRealtime.DeletedCalled {
		t.Error("Expected NotifySecretDeleted to be called")
	}
}

type MockRealtimeService struct {
	CreatedCalled        bool
	UpdatedCalled        bool
	DeletedCalled        bool
	LastExcludeSessionID string
}

func (m *MockRealtimeService) NotifySecretCreated(userID int, secretID string, excludeSessionID string) error {
	m.CreatedCalled = true
	m.LastExcludeSessionID = excludeSessionID
	return nil
}

func (m *MockRealtimeService) NotifySecretUpdated(userID int, secretID string, excludeSessionID string) error {
	m.UpdatedCalled = true
	m.LastExcludeSessionID = excludeSessionID
	return nil
}

func (m *MockRealtimeService) NotifySecretDeleted(userID int, secretID string, excludeSessionID string) error {
	m.DeletedCalled = true
	m.LastExcludeSessionID = excludeSessionID
	return nil
}
