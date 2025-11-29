package realtime

import (
	"github.com/Adigezalov/goph-keeper/internal/logger"
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
	logger.Log.WithFields(map[string]interface{}{
		"user_id":         userID,
		"secret_id":       secretID,
		"exclude_session": excludeSessionID,
	}).Info("[Realtime] NotifySecretCreated")

	var excludeSession *melody.Session
	if excludeSessionID != "" {
		excludeSession = s.hub.GetSessionByID(userID, excludeSessionID)
		if excludeSession != nil {
			logger.Log.WithFields(map[string]interface{}{
				"user_id":    userID,
				"session_id": excludeSessionID,
			}).Info("[Realtime] Найдена сессия для исключения")
		}
	}

	message := NewSecretEventMessage(SecretEventCreated, secretID, userID)
	err := s.hub.BroadcastToUser(userID, message, excludeSession)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"user_id":   userID,
			"secret_id": secretID,
			"error":     err.Error(),
		}).Error("[Realtime] Ошибка отправки события NotifySecretCreated")
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
