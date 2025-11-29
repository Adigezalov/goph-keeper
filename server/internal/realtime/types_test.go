package realtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSecretEventMessage(t *testing.T) {
	userID := 123
	secretID := "test-secret-id"
	eventType := SecretEventCreated

	message := NewSecretEventMessage(eventType, secretID, userID)

	assert.NotNil(t, message)
	assert.Equal(t, eventType, message.Type)
	assert.Equal(t, secretID, message.SecretID)
	assert.Equal(t, userID, message.UserID)
	assert.NotEmpty(t, message.Timestamp)

	// Verify timestamp is in RFC3339 format
	_, err := time.Parse(time.RFC3339, message.Timestamp)
	assert.NoError(t, err, "Timestamp should be in RFC3339 format")
}

func TestSecretEventTypes(t *testing.T) {
	tests := []struct {
		name      string
		eventType SecretEventType
		expected  string
	}{
		{
			name:      "Secret Created Event",
			eventType: SecretEventCreated,
			expected:  "secret_created",
		},
		{
			name:      "Secret Updated Event",
			eventType: SecretEventUpdated,
			expected:  "secret_updated",
		},
		{
			name:      "Secret Deleted Event",
			eventType: SecretEventDeleted,
			expected:  "secret_deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.eventType))
		})
	}
}

func TestSecretEventMessage_AllEventTypes(t *testing.T) {
	userID := 1
	secretID := "secret-123"

	tests := []struct {
		name      string
		eventType SecretEventType
	}{
		{"Created event", SecretEventCreated},
		{"Updated event", SecretEventUpdated},
		{"Deleted event", SecretEventDeleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := NewSecretEventMessage(tt.eventType, secretID, userID)

			assert.Equal(t, tt.eventType, message.Type)
			assert.Equal(t, secretID, message.SecretID)
			assert.Equal(t, userID, message.UserID)
		})
	}
}

func TestSecretEventMessage_TimestampUniqueness(t *testing.T) {
	msg1 := NewSecretEventMessage(SecretEventCreated, "secret-1", 1)
	time.Sleep(1 * time.Millisecond)
	msg2 := NewSecretEventMessage(SecretEventCreated, "secret-2", 2)

	// Timestamps should be close but might be different if there's enough time between calls
	ts1, _ := time.Parse(time.RFC3339, msg1.Timestamp)
	ts2, _ := time.Parse(time.RFC3339, msg2.Timestamp)

	assert.True(t, ts2.After(ts1) || ts2.Equal(ts1), "Second message timestamp should be equal or after the first")
}
