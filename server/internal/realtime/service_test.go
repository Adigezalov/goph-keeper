package realtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_NotifySecretCreated(t *testing.T) {
	hub := NewHub()
	service := NewService(hub)

	userID := 1
	secretID := "test-secret-id"
	excludeSessionID := "exclude-session"

	err := service.NotifySecretCreated(userID, secretID, excludeSessionID)
	assert.NoError(t, err)
}

func TestService_NotifySecretCreated_WithExcludeSession(t *testing.T) {
	hub := NewHub()
	service := NewService(hub)

	userID := 1
	secretID := "test-secret-id"
	excludeSessionID := "exclude-session"

	err := service.NotifySecretCreated(userID, secretID, excludeSessionID)
	assert.NoError(t, err)
}

func TestService_NotifySecretUpdated(t *testing.T) {
	hub := NewHub()
	service := NewService(hub)

	userID := 1
	secretID := "test-secret-id"
	excludeSessionID := "exclude-session"

	err := service.NotifySecretUpdated(userID, secretID, excludeSessionID)
	assert.NoError(t, err)
}

func TestService_NotifySecretDeleted(t *testing.T) {
	hub := NewHub()
	service := NewService(hub)

	userID := 1
	secretID := "test-secret-id"
	excludeSessionID := "exclude-session"

	err := service.NotifySecretDeleted(userID, secretID, excludeSessionID)
	assert.NoError(t, err)
}

func TestService_NotifySecretCreated_WithActiveConnections(t *testing.T) {
	hub := NewHub()
	service := NewService(hub)

	userID := 1
	secretID := "test-secret-id"

	err := service.NotifySecretCreated(userID, secretID, "")
	require.NoError(t, err)
}
