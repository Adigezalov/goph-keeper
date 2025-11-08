package utils

import (
	"net/http"
	"time"
)

// SetRefreshTokenCookie устанавливает refresh token в HTTP cookie
func SetRefreshTokenCookie(w http.ResponseWriter, refreshToken string, refreshTokenTTL time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   int(refreshTokenTTL.Seconds()),
		HttpOnly: true,
		Secure:   false, // Установите true при использовании HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}

// DeleteRefreshTokenCookie удаляет refresh token cookie
func DeleteRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Удаляет cookie
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}
