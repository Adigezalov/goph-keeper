package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Adigezalov/goph-keeper/internal/tokens"
)

// UserContextKey ключ для хранения данных пользователя в контексте
type UserContextKey string

const (
	UserIDKey    UserContextKey = "user_id"
	UserEmailKey UserContextKey = "user_email"
	SessionIDKey UserContextKey = "session_id" // ID WebSocket сессии (опционально)
)

// AuthMiddleware middleware для проверки авторизации
type AuthMiddleware struct {
	tokenService *tokens.Service
}

// NewAuthMiddleware создает новый экземпляр AuthMiddleware
func NewAuthMiddleware(tokenService *tokens.Service) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// RequireAuth проверяет наличие и валидность access токена
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Ошибка авторизации: отсутствует заголовок Authorization", http.StatusUnauthorized)
			return
		}

		// Проверяем формат Bearer токена
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Ошибка авторизации: неверный формат токена", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Валидируем токен
		claims, err := m.tokenService.ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, "Ошибка авторизации: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Добавляем данные пользователя в контекст
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

		// Получаем sessionID из заголовка (если есть) - опционально для WebSocket идентификации
		if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
			ctx = context.WithValue(ctx, SessionIDKey, sessionID)
		}

		r = r.WithContext(ctx)

		// Передаем управление следующему обработчику
		next(w, r)
	}
}

// GetUserIDFromContext извлекает ID пользователя из контекста
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}

// GetSessionIDFromContext извлекает ID WebSocket сессии из контекста
func GetSessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(SessionIDKey).(string)
	return sessionID, ok
}
