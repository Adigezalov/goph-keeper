package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetRefreshTokenCookie(t *testing.T) {
	tests := []struct {
		name            string
		refreshToken    string
		refreshTokenTTL time.Duration
		expectedMaxAge  int
	}{
		{
			name:            "Set cookie with 1 hour TTL",
			refreshToken:    "test-refresh-token-123",
			refreshTokenTTL: 1 * time.Hour,
			expectedMaxAge:  3600,
		},
		{
			name:            "Set cookie with 2 hours TTL",
			refreshToken:    "another-token-456",
			refreshTokenTTL: 2 * time.Hour,
			expectedMaxAge:  7200,
		},
		{
			name:            "Set cookie with 30 minutes TTL",
			refreshToken:    "short-lived-token",
			refreshTokenTTL: 30 * time.Minute,
			expectedMaxAge:  1800,
		},
		{
			name:            "Set cookie with empty token",
			refreshToken:    "",
			refreshTokenTTL: 1 * time.Hour,
			expectedMaxAge:  3600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			SetRefreshTokenCookie(w, tt.refreshToken, tt.refreshTokenTTL)

			result := w.Result()
			defer result.Body.Close()

			cookies := result.Cookies()
			require.Len(t, cookies, 1, "Expected exactly one cookie to be set")

			cookie := cookies[0]
			assert.Equal(t, "refresh_token", cookie.Name)
			assert.Equal(t, tt.refreshToken, cookie.Value)
			assert.Equal(t, "/", cookie.Path)
			assert.Equal(t, tt.expectedMaxAge, cookie.MaxAge)
			assert.True(t, cookie.HttpOnly)
			assert.False(t, cookie.Secure)
			assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
		})
	}
}

func TestSetRefreshTokenCookie_Properties(t *testing.T) {
	w := httptest.NewRecorder()
	refreshToken := "test-token"
	refreshTokenTTL := 1 * time.Hour

	SetRefreshTokenCookie(w, refreshToken, refreshTokenTTL)

	result := w.Result()
	defer result.Body.Close()

	cookies := result.Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]

	t.Run("Cookie is HttpOnly", func(t *testing.T) {
		assert.True(t, cookie.HttpOnly, "Cookie should be HttpOnly to prevent XSS attacks")
	})

	t.Run("Cookie path is root", func(t *testing.T) {
		assert.Equal(t, "/", cookie.Path, "Cookie should be available on all paths")
	})

	t.Run("Cookie has SameSite Lax", func(t *testing.T) {
		assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite, "Cookie should have SameSite=Lax")
	})

	t.Run("Cookie is not Secure in development", func(t *testing.T) {
		assert.False(t, cookie.Secure, "Cookie Secure flag should be false for development")
	})
}

func TestDeleteRefreshTokenCookie(t *testing.T) {
	w := httptest.NewRecorder()

	DeleteRefreshTokenCookie(w)

	result := w.Result()
	defer result.Body.Close()

	cookies := result.Cookies()
	require.Len(t, cookies, 1, "Expected exactly one cookie to be set")

	cookie := cookies[0]
	assert.Equal(t, "refresh_token", cookie.Name)
	assert.Empty(t, cookie.Value, "Cookie value should be empty when deleting")
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, -1, cookie.MaxAge, "MaxAge should be -1 to delete the cookie")
	assert.True(t, cookie.HttpOnly)
	assert.False(t, cookie.Secure)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestDeleteRefreshTokenCookie_Multiple(t *testing.T) {
	t.Run("Delete cookie multiple times", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			w := httptest.NewRecorder()
			DeleteRefreshTokenCookie(w)

			result := w.Result()
			cookies := result.Cookies()
			result.Body.Close()

			require.Len(t, cookies, 1)
			assert.Equal(t, -1, cookies[0].MaxAge)
		}
	})
}

func TestSetAndDeleteRefreshTokenCookie_Sequence(t *testing.T) {
	// Test setting a cookie
	w1 := httptest.NewRecorder()
	refreshToken := "test-token-sequence"
	refreshTokenTTL := 2 * time.Hour

	SetRefreshTokenCookie(w1, refreshToken, refreshTokenTTL)

	result1 := w1.Result()
	defer result1.Body.Close()
	cookies1 := result1.Cookies()
	require.Len(t, cookies1, 1)
	assert.Equal(t, refreshToken, cookies1[0].Value)
	assert.Equal(t, 7200, cookies1[0].MaxAge)

	// Test deleting the cookie
	w2 := httptest.NewRecorder()
	DeleteRefreshTokenCookie(w2)

	result2 := w2.Result()
	defer result2.Body.Close()
	cookies2 := result2.Cookies()
	require.Len(t, cookies2, 1)
	assert.Empty(t, cookies2[0].Value)
	assert.Equal(t, -1, cookies2[0].MaxAge)
}

func TestSetRefreshTokenCookie_ZeroDuration(t *testing.T) {
	w := httptest.NewRecorder()
	refreshToken := "zero-duration-token"
	refreshTokenTTL := 0 * time.Second

	SetRefreshTokenCookie(w, refreshToken, refreshTokenTTL)

	result := w.Result()
	defer result.Body.Close()

	cookies := result.Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, "refresh_token", cookie.Name)
	assert.Equal(t, refreshToken, cookie.Value)
	assert.Equal(t, 0, cookie.MaxAge, "MaxAge should be 0 for zero duration")
}

func TestSetRefreshTokenCookie_LongToken(t *testing.T) {
	w := httptest.NewRecorder()
	// Create a long token string
	longToken := ""
	for i := 0; i < 100; i++ {
		longToken += "abcdefghij"
	}
	refreshTokenTTL := 1 * time.Hour

	SetRefreshTokenCookie(w, longToken, refreshTokenTTL)

	result := w.Result()
	defer result.Body.Close()

	cookies := result.Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, longToken, cookie.Value, "Should handle long token values")
	assert.Equal(t, 3600, cookie.MaxAge)
}
