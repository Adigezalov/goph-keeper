package realtime

import "time"

// SecretEventType тип события секрета
type SecretEventType string

const (
	SecretEventCreated SecretEventType = "secret_created"
	SecretEventUpdated SecretEventType = "secret_updated"
	SecretEventDeleted SecretEventType = "secret_deleted"
)

// SecretEventMessage сообщение о событии секрета
type SecretEventMessage struct {
	Type      SecretEventType `json:"type"`      // "secret_created", "secret_updated", "secret_deleted"
	SecretID  string          `json:"secret_id"` // ID секрета
	UserID    int             `json:"user_id"`   // ID пользователя (для валидации на клиенте)
	Timestamp string          `json:"timestamp"` // RFC3339 формат
}

// NewSecretEventMessage создает новое сообщение о событии секрета
func NewSecretEventMessage(eventType SecretEventType, secretID string, userID int) *SecretEventMessage {
	return &SecretEventMessage{
		Type:      eventType,
		SecretID:  secretID,
		UserID:    userID,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
