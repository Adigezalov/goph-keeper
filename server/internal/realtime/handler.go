package realtime

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/Adigezalov/goph-keeper/internal/middleware"
	"github.com/Adigezalov/goph-keeper/internal/tokens"
	"github.com/olahol/melody"
)

// Handler обрабатывает WebSocket соединения
type Handler struct {
	hub          *Hub
	tokenService *tokens.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(hub *Hub, tokenService *tokens.Service) *Handler {
	return &Handler{
		hub:          hub,
		tokenService: tokenService,
	}
}

// HandleWebSocket обрабатывает WebSocket подключение
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Realtime] Получен запрос на WebSocket подключение: %s %s", r.Method, r.URL.String())

	// Получаем токен из query параметра или заголовка
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		// Пробуем получить из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}
	}

	if tokenString == "" {
		log.Printf("[Realtime] Ошибка: токен не предоставлен")
		http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
		return
	}

	// Валидируем токен
	claims, err := h.tokenService.ValidateAccessToken(tokenString)
	if err != nil {
		log.Printf("[Realtime] Ошибка валидации токена: %v", err)
		http.Error(w, "Неверный токен: "+err.Error(), http.StatusUnauthorized)
		return
	}

	userID := claims.UserID
	sessionID := r.URL.Query().Get("session_id")
	log.Printf("[Realtime] Валидация токена успешна, userID=%d, sessionID=%s", userID, sessionID)

	// Сохраняем userID и sessionID в контексте запроса для использования в обработчиках
	keys := make(map[string]interface{})
	keys["user_id"] = userID
	keys["session_id"] = sessionID

	// Получаем Melody instance
	m := h.hub.GetMelody()

	// Обновляем WebSocket соединение с ключами
	log.Printf("[Realtime] Выполняем WebSocket upgrade для userID=%d", userID)
	if err := m.HandleRequestWithKeys(w, r, keys); err != nil {
		log.Printf("[Realtime] Ошибка обработки WebSocket запроса: %v", err)
	} else {
		log.Printf("[Realtime] WebSocket upgrade успешен для userID=%d", userID)
	}
}

// GetSessionFromContext извлекает сессию WebSocket из контекста запроса
// Использует sessionID из контекста для поиска сессии в Hub
func (h *Handler) GetSessionFromContext(ctx context.Context) *melody.Session {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return nil
	}

	sessionID, ok := middleware.GetSessionIDFromContext(ctx)
	if !ok {
		return nil
	}

	return h.hub.GetSessionByID(userID, sessionID)
}
