package realtime

import "time"

type SecretEventType string

const (
	SecretEventCreated SecretEventType = "secret_created"
	SecretEventUpdated SecretEventType = "secret_updated"
	SecretEventDeleted SecretEventType = "secret_deleted"
)

type SecretEventMessage struct {
	Type      SecretEventType `json:"type"`
	SecretID  string          `json:"secret_id"`
	UserID    int             `json:"user_id"`
	Timestamp string          `json:"timestamp"`
}

func NewSecretEventMessage(eventType SecretEventType, secretID string, userID int) *SecretEventMessage {
	return &SecretEventMessage{
		Type:      eventType,
		SecretID:  secretID,
		UserID:    userID,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
