package realtime

import (
	"github.com/olahol/melody"
)

// Service предоставляет интерфейс для отправки событий в реальном времени
type Service struct {
	hub *Hub
}

// NewService создает новый экземпляр Service
func NewService(hub *Hub) *Service {
	return &Service{
		hub: hub,
	}
}

// NotifySecretCreated отправляет уведомление о создании секрета
// excludeSessionID - опциональный ID сессии, которую нужно исключить из рассылки
func (s *Service) NotifySecretCreated(userID int, secretID string, excludeSessionID string) error {
	var excludeSession *melody.Session
	if excludeSessionID != "" {
		excludeSession = s.hub.GetSessionByID(userID, excludeSessionID)
	}

	message := NewSecretEventMessage(SecretEventCreated, secretID, userID)
	return s.hub.BroadcastToUser(userID, message, excludeSession)
}

// NotifySecretUpdated отправляет уведомление об обновлении секрета
// excludeSessionID - опциональный ID сессии, которую нужно исключить из рассылки
func (s *Service) NotifySecretUpdated(userID int, secretID string, excludeSessionID string) error {
	var excludeSession *melody.Session
	if excludeSessionID != "" {
		excludeSession = s.hub.GetSessionByID(userID, excludeSessionID)
	}

	message := NewSecretEventMessage(SecretEventUpdated, secretID, userID)
	return s.hub.BroadcastToUser(userID, message, excludeSession)
}

// NotifySecretDeleted отправляет уведомление об удалении секрета
// excludeSessionID - опциональный ID сессии, которую нужно исключить из рассылки
func (s *Service) NotifySecretDeleted(userID int, secretID string, excludeSessionID string) error {
	var excludeSession *melody.Session
	if excludeSessionID != "" {
		excludeSession = s.hub.GetSessionByID(userID, excludeSessionID)
	}

	message := NewSecretEventMessage(SecretEventDeleted, secretID, userID)
	return s.hub.BroadcastToUser(userID, message, excludeSession)
}
