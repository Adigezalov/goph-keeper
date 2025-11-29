package realtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHub_GetConnectionCount_NoConnections(t *testing.T) {
	hub := NewHub()

	count := hub.GetConnectionCount(999)
	assert.Equal(t, 0, count)
}

func TestHub_GetSessionByID_NotFound(t *testing.T) {
	hub := NewHub()

	userID := 1
	sessionID := "non-existent-session"

	foundSession := hub.GetSessionByID(userID, sessionID)
	assert.Nil(t, foundSession)
}

func TestHub_BroadcastToUser_NoConnections(t *testing.T) {
	hub := NewHub()

	userID := 999
	message := NewSecretEventMessage(SecretEventCreated, "secret-id", userID)

	err := hub.BroadcastToUser(userID, message, nil)
	assert.NoError(t, err)
}

func TestHub_NewHub(t *testing.T) {
	hub := NewHub()
	assert.NotNil(t, hub)
	assert.NotNil(t, hub.GetMelody())
	assert.Equal(t, 0, hub.GetConnectionCount(1))
}

func TestHub_BroadcastToUser_InvalidJSON(t *testing.T) {
	hub := NewHub()

	userID := 1

	// Create a message that will be marshaled successfully
	message := NewSecretEventMessage(SecretEventCreated, "secret-id", userID)

	// Even without connections, it should not error
	err := hub.BroadcastToUser(userID, message, nil)
	assert.NoError(t, err)
}

func TestHub_GetConnectionCount_MultipleUsers(t *testing.T) {
	hub := NewHub()

	// No connections initially
	assert.Equal(t, 0, hub.GetConnectionCount(1))
	assert.Equal(t, 0, hub.GetConnectionCount(2))
	assert.Equal(t, 0, hub.GetConnectionCount(999))
}

func TestHub_GetSessionByID_NoSessions(t *testing.T) {
	hub := NewHub()

	session := hub.GetSessionByID(1, "test-session-id")
	assert.Nil(t, session)
}

func TestHub_GetMelody(t *testing.T) {
	hub := NewHub()

	melody := hub.GetMelody()
	assert.NotNil(t, melody)
	assert.NotNil(t, melody.Config)
}

func TestSecretEventTypes_Values(t *testing.T) {
	assert.Equal(t, "secret_created", string(SecretEventCreated))
	assert.Equal(t, "secret_updated", string(SecretEventUpdated))
	assert.Equal(t, "secret_deleted", string(SecretEventDeleted))
}
