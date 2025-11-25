package realtime

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockTokenService struct {
	ValidateAccessTokenFunc func(tokenString string) (*tokens.Claims, error)
}

func (m *MockTokenService) ValidateAccessToken(tokenString string) (*tokens.Claims, error) {
	if m.ValidateAccessTokenFunc != nil {
		return m.ValidateAccessTokenFunc(tokenString)
	}
	return &tokens.Claims{
		UserID: 1,
		Email:  "test@example.com",
	}, nil
}

func (m *MockTokenService) GenerateTokenPair(userID int, email string) (*tokens.TokenPair, error) {
	return nil, nil
}

func (m *MockTokenService) GetRefreshToken(tokenString string) (*tokens.RefreshToken, error) {
	return nil, nil
}

func (m *MockTokenService) RefreshTokenPair(refreshTokenString string, userID int, email string) (*tokens.TokenPair, error) {
	return nil, nil
}

func (m *MockTokenService) Logout(refreshTokenString string) error {
	return nil
}

func (m *MockTokenService) LogoutAll(userID int) error {
	return nil
}

type mockTokenRepo struct{}

func (m *mockTokenRepo) SaveRefreshToken(token *tokens.RefreshToken) error { return nil }
func (m *mockTokenRepo) GetRefreshToken(token string) (*tokens.RefreshToken, error) {
	return &tokens.RefreshToken{Token: token, UserID: 1}, nil
}
func (m *mockTokenRepo) DeleteRefreshToken(token string) error    { return nil }
func (m *mockTokenRepo) DeleteUserRefreshTokens(userID int) error { return nil }

func createMockTokenService() *tokens.Service {
	repo := &mockTokenRepo{}
	return tokens.NewService(repo, "test-secret-key-for-testing-purposes-only", 10*time.Minute, 2*time.Hour)
}

func TestHandler_HandleWebSocket_NoToken(t *testing.T) {
	hub := NewHub()
	tokenService := createMockTokenService()
	handler := NewHandler(hub, tokenService)

	req := httptest.NewRequest("GET", "/api/v1/realtime", nil)
	w := httptest.NewRecorder()

	handler.HandleWebSocket(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleWebSocket_InvalidToken(t *testing.T) {
	hub := NewHub()
	tokenService := createMockTokenService()
	handler := NewHandler(hub, tokenService)

	req := httptest.NewRequest("GET", "/api/v1/realtime?token=invalid", nil)
	w := httptest.NewRecorder()

	handler.HandleWebSocket(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleWebSocket_ValidTokenInQuery(t *testing.T) {
	hub := NewHub()
	tokenService := createMockTokenService()
	handler := NewHandler(hub, tokenService)

	tokenPair, err := tokenService.GenerateTokenPair(1, "test@example.com")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/realtime?token="+tokenPair.AccessToken+"&session_id=test-session", nil)
	w := httptest.NewRecorder()

	handler.HandleWebSocket(w, req)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleWebSocket_ValidTokenInHeader(t *testing.T) {
	hub := NewHub()
	tokenService := createMockTokenService()
	handler := NewHandler(hub, tokenService)

	tokenPair, err := tokenService.GenerateTokenPair(2, "test2@example.com")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/realtime", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	w := httptest.NewRecorder()

	handler.HandleWebSocket(w, req)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}
