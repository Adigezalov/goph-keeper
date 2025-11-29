package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Adigezalov/goph-keeper/internal/localization"
	"github.com/Adigezalov/goph-keeper/internal/logger"
	"github.com/Adigezalov/goph-keeper/internal/tokens"
)

type UserContextKey string

const (
	UserIDKey    UserContextKey = "user_id"
	UserEmailKey UserContextKey = "user_email"
	SessionIDKey UserContextKey = "session_id"
)

type AuthMiddleware struct {
	tokenService *tokens.Service
}

func NewAuthMiddleware(tokenService *tokens.Service) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Log.WithFields(map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
			}).Warn("[Auth] Отсутствует заголовок Authorization")
			localization.LocalizedError(w, r, http.StatusUnauthorized, "auth.missing_authorization_header", nil)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Log.WithFields(map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
			}).Warn("[Auth] Неверный формат токена")
			localization.LocalizedError(w, r, http.StatusUnauthorized, "auth.invalid_token_format", nil)
			return
		}

		tokenString := parts[1]

		claims, err := m.tokenService.ValidateAccessToken(tokenString)
		if err != nil {
			logger.Log.WithFields(map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
				"error":  err.Error(),
			}).Warn("[Auth] Ошибка валидации токена")
			localization.LocalizedError(w, r, http.StatusUnauthorized, "auth.invalid_token", map[string]interface{}{
				"Error": err.Error(),
			})
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

		if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
			ctx = context.WithValue(ctx, SessionIDKey, sessionID)
		}

		r = r.WithContext(ctx)

		next(w, r)
	}
}

func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}

func GetSessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(SessionIDKey).(string)
	return sessionID, ok
}
