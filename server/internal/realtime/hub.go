package realtime

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/olahol/melody"
)

// Hub управляет WebSocket соединениями пользователей
type Hub struct {
	// connections хранит активные соединения по userID
	// map[userID][]*melody.Session
	connections map[int][]*melody.Session
	mu          sync.RWMutex
	melody      *melody.Melody
}

// NewHub создает новый Hub
func NewHub() *Hub {
	m := melody.New()
	m.Config.MaxMessageSize = 1024 // Ограничиваем размер сообщений (у нас только легковесные события)

	hub := &Hub{
		connections: make(map[int][]*melody.Session),
		melody:      m,
	}

	// Обработчик подключения клиента
	m.HandleConnect(func(s *melody.Session) {
		// Получаем userID и sessionID из keys, переданных при подключении
		userIDValue, exists := s.Keys["user_id"]
		if !exists {
			log.Printf("[Realtime] Ошибка: user_id не найден в keys сессии")
			return
		}

		userID, ok := userIDValue.(int)
		if !ok {
			log.Printf("[Realtime] Ошибка: неверный тип user_id в keys сессии")
			return
		}

		hub.RegisterSession(userID, s)
	})

	// Обработчик отключения клиента
	m.HandleDisconnect(func(s *melody.Session) {
		hub.unregisterSession(s)
	})

	return hub
}

// GetMelody возвращает экземпляр Melody для использования в handler
func (h *Hub) GetMelody() *melody.Melody {
	return h.melody
}

// RegisterSession регистрирует новое соединение для пользователя
func (h *Hub) RegisterSession(userID int, session *melody.Session) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Генерируем уникальный sessionID для этой сессии
	sessionID := session.Request.URL.Query().Get("session_id")
	if sessionID == "" {
		// Если sessionID не передан, генерируем новый UUID
		sessionID = uuid.New().String()
	}

	// Добавляем userID и sessionID в ключи сессии для быстрого поиска
	session.Set("user_id", userID)
	session.Set("session_id", sessionID)

	// Добавляем сессию в список соединений пользователя
	h.connections[userID] = append(h.connections[userID], session)

	log.Printf("[Realtime] Зарегистрировано соединение для userID=%d, sessionID=%s, всего соединений: %d", userID, sessionID, len(h.connections[userID]))
}

// unregisterSession удаляет соединение из списка
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

	// Удаляем сессию из списка
	sessions := h.connections[userID]
	for i, s := range sessions {
		if s == session {
			// Удаляем элемент из slice
			h.connections[userID] = append(sessions[:i], sessions[i+1:]...)
			break
		}
	}

	// Если соединений не осталось, удаляем запись
	if len(h.connections[userID]) == 0 {
		delete(h.connections, userID)
	}

	log.Printf("[Realtime] Отключено соединение для userID=%d, осталось соединений: %d", userID, len(h.connections[userID]))
}

// BroadcastToUser отправляет сообщение всем соединениям пользователя, кроме исключенной сессии
func (h *Hub) BroadcastToUser(userID int, message *SecretEventMessage, excludeSession *melody.Session) error {
	h.mu.RLock()
	sessions, exists := h.connections[userID]
	h.mu.RUnlock()

	if !exists || len(sessions) == 0 {
		// Нет активных соединений - это нормально, просто не отправляем
		return nil
	}

	// Сериализуем сообщение в JSON
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Отправляем всем соединениям, кроме исключенной сессии
	sentCount := 0
	for _, session := range sessions {
		if excludeSession != nil && session == excludeSession {
			continue // Пропускаем сессию отправителя
		}

		if err := session.Write(messageBytes); err != nil {
			log.Printf("[Realtime] Ошибка отправки сообщения userID=%d: %v", userID, err)
			continue
		}
		sentCount++
	}

	if sentCount > 0 {
		log.Printf("[Realtime] Отправлено сообщение userID=%d, тип=%s, secretID=%s, получателей=%d", userID, message.Type, message.SecretID, sentCount)
	}

	return nil
}

// GetConnectionCount возвращает количество активных соединений пользователя
func (h *Hub) GetConnectionCount(userID int) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.connections[userID])
}

// GetSessionByID находит сессию по userID и sessionID
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
