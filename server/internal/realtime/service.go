package realtime

import (
	"log"

	"github.com/olahol/melody"
)

type Service struct {
	hub *Hub
}

func NewService(hub *Hub) *Service {
	return &Service{
		hub: hub,
	}
}

func (s *Service) NotifySecretCreated(userID int, secretID string, excludeSessionID string) error {
	log.Printf("[Realtime] NotifySecretCreated: userID=%d, secretID=%s, excludeSessionID=%s", userID, secretID, excludeSessionID)

	var excludeSession *melody.Session
	if excludeSessionID != "" {
		excludeSession = s.hub.GetSessionByID(userID, excludeSessionID)
		if excludeSession != nil {
			log.Printf("[Realtime] Найдена сессия для исключения: userID=%d, sessionID=%s", userID, excludeSessionID)
		}
	}

	message := NewSecretEventMessage(SecretEventCreated, secretID, userID)
	err := s.hub.BroadcastToUser(userID, message, excludeSession)
	if err != nil {
		log.Printf("[Realtime] Ошибка отправки события NotifySecretCreated: %v", err)
	}
	return err
}

func (s *Service) NotifySecretUpdated(userID int, secretID string, excludeSessionID string) error {
	var excludeSession *melody.Session
	if excludeSessionID != "" {
		excludeSession = s.hub.GetSessionByID(userID, excludeSessionID)
	}

	message := NewSecretEventMessage(SecretEventUpdated, secretID, userID)
	return s.hub.BroadcastToUser(userID, message, excludeSession)
}

func (s *Service) NotifySecretDeleted(userID int, secretID string, excludeSessionID string) error {
	var excludeSession *melody.Session
	if excludeSessionID != "" {
		excludeSession = s.hub.GetSessionByID(userID, excludeSessionID)
	}

	message := NewSecretEventMessage(SecretEventDeleted, secretID, userID)
	return s.hub.BroadcastToUser(userID, message, excludeSession)
}
