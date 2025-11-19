package realtime

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/Adigezalov/goph-keeper/internal/localization"
	"github.com/Adigezalov/goph-keeper/internal/middleware"
	"github.com/Adigezalov/goph-keeper/internal/tokens"
	"github.com/olahol/melody"
)

type Handler struct {
	hub          *Hub
	tokenService *tokens.Service
}

func NewHandler(hub *Hub, tokenService *tokens.Service) *Handler {
	return &Handler{
		hub:          hub,
		tokenService: tokenService,
	}
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Realtime] Получен запрос на WebSocket подключение: %s %s", r.Method, r.URL.String())

	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
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
		localization.LocalizedError(w, r, http.StatusUnauthorized, "realtime.token_not_provided", nil)
		return
	}

	claims, err := h.tokenService.ValidateAccessToken(tokenString)
	if err != nil {
		log.Printf("[Realtime] Ошибка валидации токена: %v", err)
		localization.LocalizedError(w, r, http.StatusUnauthorized, "realtime.invalid_token", map[string]interface{}{
			"Error": err.Error(),
		})
		return
	}

	userID := claims.UserID
	sessionID := r.URL.Query().Get("session_id")
	log.Printf("[Realtime] Валидация токена успешна, userID=%d, sessionID=%s", userID, sessionID)

	keys := make(map[string]interface{})
	keys["user_id"] = userID
	keys["session_id"] = sessionID

	m := h.hub.GetMelody()

	log.Printf("[Realtime] Выполняем WebSocket upgrade для userID=%d", userID)
	if err := m.HandleRequestWithKeys(w, r, keys); err != nil {
		log.Printf("[Realtime] Ошибка обработки WebSocket запроса: %v", err)
	} else {
		log.Printf("[Realtime] WebSocket upgrade успешен для userID=%d", userID)
	}
}

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
