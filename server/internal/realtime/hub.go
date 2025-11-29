package realtime

import (
	"encoding/json"
	"sync"

	"github.com/Adigezalov/goph-keeper/internal/logger"
	"github.com/google/uuid"
	"github.com/olahol/melody"
)

type Hub struct {
	connections map[int][]*melody.Session
	mu          sync.RWMutex
	melody      *melody.Melody
}

func NewHub() *Hub {
	m := melody.New()
	m.Config.MaxMessageSize = 1024

	hub := &Hub{
		connections: make(map[int][]*melody.Session),
		melody:      m,
	}

	m.HandleConnect(func(s *melody.Session) {
		userIDValue, exists := s.Keys["user_id"]
		if !exists {
			logger.Error("[Realtime] Ошибка: user_id не найден в keys сессии")
			return
		}

		userID, ok := userIDValue.(int)
		if !ok {
			logger.Error("[Realtime] Ошибка: неверный тип user_id в keys сессии")
			return
		}

		hub.RegisterSession(userID, s)
	})

	m.HandleDisconnect(func(s *melody.Session) {
		hub.unregisterSession(s)
	})

	return hub
}

func (h *Hub) GetMelody() *melody.Melody {
	return h.melody
}

func (h *Hub) RegisterSession(userID int, session *melody.Session) {
	h.mu.Lock()
	defer h.mu.Unlock()

	sessionID := session.Request.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	session.Set("user_id", userID)
	session.Set("session_id", sessionID)

	h.connections[userID] = append(h.connections[userID], session)

	logger.Log.WithFields(map[string]interface{}{
		"user_id":     userID,
		"session_id":  sessionID,
		"connections": len(h.connections[userID]),
	}).Info("[Realtime] WebSocket подключение установлено")
}

func (h *Hub) unregisterSession(session *melody.Session) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userIDValue, exists := session.Get("user_id")
	if !exists {
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		return
	}

	sessions := h.connections[userID]
	for i, s := range sessions {
		if s == session {
			h.connections[userID] = append(sessions[:i], sessions[i+1:]...)
			break
		}
	}

	if len(h.connections[userID]) == 0 {
		delete(h.connections, userID)
	}

	logger.Log.WithFields(map[string]interface{}{
		"user_id":     userID,
		"connections": len(h.connections[userID]),
	}).Info("[Realtime] WebSocket подключение закрыто")
}

func (h *Hub) BroadcastToUser(userID int, message *SecretEventMessage, excludeSession *melody.Session) error {
	h.mu.RLock()
	sessions, exists := h.connections[userID]
	h.mu.RUnlock()

	if !exists || len(sessions) == 0 {
		return nil
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	sentCount := 0
	for _, session := range sessions {
		if excludeSession != nil && session == excludeSession {
			continue
		}

		if err := session.Write(messageBytes); err != nil {
			logger.Log.WithFields(map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			}).Error("[Realtime] Ошибка отправки сообщения")
			continue
		}
		sentCount++
	}

	if sentCount > 0 {
		logger.Log.WithFields(map[string]interface{}{
			"user_id":    userID,
			"type":       message.Type,
			"secret_id":  message.SecretID,
			"recipients": sentCount,
		}).Info("[Realtime] Сообщение отправлено")
	}

	return nil
}

func (h *Hub) GetConnectionCount(userID int) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.connections[userID])
}

func (h *Hub) GetSessionByID(userID int, sessionID string) *melody.Session {
	h.mu.RLock()
	defer h.mu.RUnlock()

	sessions, exists := h.connections[userID]
	if !exists {
		return nil
	}

	for _, session := range sessions {
		sessionIDValue, exists := session.Get("session_id")
		if !exists {
			continue
		}

		if sid, ok := sessionIDValue.(string); ok && sid == sessionID {
			return session
		}
	}

	return nil
}
