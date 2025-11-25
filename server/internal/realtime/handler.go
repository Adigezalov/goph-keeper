package realtime

import (
	"context"
	"net/http"
	"strings"

	"github.com/Adigezalov/goph-keeper/internal/localization"
	"github.com/Adigezalov/goph-keeper/internal/logger"
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
	logger.Log.WithFields(map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.String(),
		"remote": r.RemoteAddr,
	}).Info("[Realtime] Получен запрос на WebSocket подключение")

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
		logger.Error("[Realtime] Ошибка: токен не предоставлен")
		localization.LocalizedError(w, r, http.StatusUnauthorized, "realtime.token_not_provided", nil)
		return
	}

	claims, err := h.tokenService.ValidateAccessToken(tokenString)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("[Realtime] Ошибка валидации токена")
		localization.LocalizedError(w, r, http.StatusUnauthorized, "realtime.invalid_token", map[string]interface{}{
			"Error": err.Error(),
		})
		return
	}

	userID := claims.UserID
	sessionID := r.URL.Query().Get("session_id")
	logger.Log.WithFields(map[string]interface{}{
		"user_id":    userID,
		"session_id": sessionID,
	}).Info("[Realtime] Валидация токена успешна")

	keys := make(map[string]interface{})
	keys["user_id"] = userID
	keys["session_id"] = sessionID

	m := h.hub.GetMelody()

	logger.Log.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("[Realtime] Выполняем WebSocket upgrade")
	if err := m.HandleRequestWithKeys(w, r, keys); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("[Realtime] Ошибка обработки WebSocket запроса")
	} else {
		logger.Log.WithFields(map[string]interface{}{
			"user_id": userID,
		}).Info("[Realtime] WebSocket upgrade успешен")
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
