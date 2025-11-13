package secret

import (
	"testing"
	"time"
)

// MockRepository для тестирования
type MockRepository struct {
	secrets map[string]*Secret
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		secrets: make(map[string]*Secret),
	}
}

func (m *MockRepository) CreateSecret(secret *Secret) error {
	// Имитируем генерацию ID и timestamps
	secret.ID = "test-uuid-" + time.Now().Format("20060102150405")
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

// Тесты

func TestService_CreateSecret(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	req := &CreateSecretRequest{
		Login:    "encrypted_login",
		Password: "encrypted_password",
		Metadata: map[string]interface{}{"app": "github"},
	}

	secret, err := service.CreateSecret(1, req)
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
			_, err := service.CreateSecret(1, tt.req)
			if err != tt.wantErr {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestService_GetSecretsForSync(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	// Создаем тестовый секрет
	req := &CreateSecretRequest{
		Login:    "encrypted_login",
		Password: "encrypted_password",
	}
	secret, _ := service.CreateSecret(1, req)

	// Первая синхронизация (без since)
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

	// Инкрементальная синхронизация
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

	// Создаем секрет
	createReq := &CreateSecretRequest{
		Login:    "login",
		Password: "password",
	}
	secret, _ := service.CreateSecret(1, createReq)

	// Обновляем с неправильной версией
	updateReq := &UpdateSecretRequest{
		Login:    "new_login",
		Password: "new_password",
		Version:  999, // Неправильная версия
	}

	_, err := service.UpdateSecret(secret.ID, 1, updateReq)
	if err != ErrVersionConflict {
		t.Errorf("Expected ErrVersionConflict, got %v", err)
	}
}
