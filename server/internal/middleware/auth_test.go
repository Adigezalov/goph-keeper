package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTokenRepo struct{}

func (m *mockTokenRepo) SaveRefreshToken(token *tokens.RefreshToken) error { return nil }
func (m *mockTokenRepo) GetRefreshToken(token string) (*tokens.RefreshToken, error) {
	return &tokens.RefreshToken{Token: token, UserID: 1}, nil
}
func (m *mockTokenRepo) DeleteRefreshToken(token string) error    { return nil }
func (m *mockTokenRepo) DeleteUserRefreshTokens(userID int) error { return nil }

func TestAuthMiddleware_RequireAuth_NoAuthHeader(t *testing.T) {
	mockRepo := &mockTokenRepo{}
	tokenService := tokens.NewService(mockRepo, "test-secret", 10*time.Minute, 2*time.Hour)
	middleware := NewAuthMiddleware(tokenService)

	req := httptest.NewRequest("GET", "/api/v1/secrets", nil)
	w := httptest.NewRecorder()

	handler := middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireAuth_InvalidFormat(t *testing.T) {
	mockRepo := &mockTokenRepo{}
	tokenService := tokens.NewService(mockRepo, "test-secret", 10*time.Minute, 2*time.Hour)
	middleware := NewAuthMiddleware(tokenService)

	req := httptest.NewRequest("GET", "/api/v1/secrets", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	w := httptest.NewRecorder()

	handler := middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireAuth_InvalidToken(t *testing.T) {
	mockRepo := &mockTokenRepo{}
	tokenService := tokens.NewService(mockRepo, "test-secret", 10*time.Minute, 2*time.Hour)
	middleware := NewAuthMiddleware(tokenService)

	req := httptest.NewRequest("GET", "/api/v1/secrets", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handler := middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireAuth_ValidToken(t *testing.T) {
	mockRepo := &mockTokenRepo{}
	tokenService := tokens.NewService(mockRepo, "test-secret-key-for-testing", 10*time.Minute, 2*time.Hour)
	middleware := NewAuthMiddleware(tokenService)

	tokenPair, err := tokenService.GenerateTokenPair(1, "test@example.com")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/secrets", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	w := httptest.NewRecorder()

	handlerCalled := false
	handler := middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		userID, ok := GetUserIDFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, 1, userID)
		w.WriteHeader(http.StatusOK)
	})

	handler(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_RequireAuth_WithSessionID(t *testing.T) {
	mockRepo := &mockTokenRepo{}
	tokenService := tokens.NewService(mockRepo, "test-secret-key-for-testing", 10*time.Minute, 2*time.Hour)
	middleware := NewAuthMiddleware(tokenService)

	tokenPair, err := tokenService.GenerateTokenPair(1, "test@example.com")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/secrets", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	req.Header.Set("X-Session-ID", "test-session-id")
	w := httptest.NewRecorder()

	handler := middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		sessionID, ok := GetSessionIDFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, "test-session-id", sessionID)
		w.WriteHeader(http.StatusOK)
	})

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserIDFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, 123)

	userID, ok := GetUserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, 123, userID)
}

func TestGetUserIDFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	userID, ok := GetUserIDFromContext(ctx)
	assert.False(t, ok)
	assert.Equal(t, 0, userID)
}

func TestGetSessionIDFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), SessionIDKey, "test-session")

	sessionID, ok := GetSessionIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "test-session", sessionID)
}

func TestGetSessionIDFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	sessionID, ok := GetSessionIDFromContext(ctx)
	assert.False(t, ok)
	assert.Empty(t, sessionID)
}
