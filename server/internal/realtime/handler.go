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
		http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
		return
	}

	// Валидируем токен
	claims, err := h.tokenService.ValidateAccessToken(tokenString)
	if err != nil {
		http.Error(w, "Неверный токен: "+err.Error(), http.StatusUnauthorized)
		return
	}

	userID := claims.UserID

	// Обработчик подключения - регистрируем сессию
	handleConnect := func(s *melody.Session) {
		h.hub.RegisterSession(userID, s)
		log.Printf("[Realtime] Подключен WebSocket для userID=%d", userID)
	}

	// Обработчик сообщений от клиента (если нужно будет двустороннее общение)
	handleMessage := func(s *melody.Session, msg []byte) {
		// Пока клиент только получает события, но можно добавить обработку команд
		log.Printf("[Realtime] Получено сообщение от userID=%d: %s", userID, string(msg))
	}

	// Настраиваем обработчики
	m := h.hub.GetMelody()
	m.HandleConnect(handleConnect)
	m.HandleMessage(handleMessage)

	// Обновляем WebSocket соединение
	if err := m.HandleRequestWithKeys(w, r, nil); err != nil {
		log.Printf("[Realtime] Ошибка обработки WebSocket запроса: %v", err)
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
