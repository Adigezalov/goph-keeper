package user

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/middleware"
	"github.com/Adigezalov/goph-keeper/internal/utils"
)

// Handler обрабатывает HTTP запросы для пользователей
type Handler struct {
	service         *Service
	refreshTokenTTL time.Duration
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service *Service, refreshTokenTTL time.Duration) *Handler {
	return &Handler{
		service:         service,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// TokenResponse представляет ответ с access token
type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

// sendTokenResponse отправляет access token в JSON ответе
func (h *Handler) sendTokenResponse(w http.ResponseWriter, accessToken string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(TokenResponse{
		AccessToken: accessToken,
	}); err != nil {
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}

// Register обрабатывает POST /api/v1/user/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	// Декодируем JSON запрос
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Регистрируем пользователя
	tokenPair, err := h.service.RegisterUser(&req)
	if err != nil {
		if strings.Contains(err.Error(), "пользователь уже существует") {
			http.Error(w, "Email уже занят", http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "обязателен") ||
			strings.Contains(err.Error(), "слишком длинный") ||
			strings.Contains(err.Error(), "минимум") ||
			strings.Contains(err.Error(), "неверный формат email") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Устанавливаем refresh token в cookie
	utils.SetRefreshTokenCookie(w, tokenPair.RefreshToken, h.refreshTokenTTL)

	// Возвращаем access token в теле ответа
	h.sendTokenResponse(w, tokenPair.AccessToken)
}

// Login обрабатывает POST /api/v1/user/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	// Декодируем JSON запрос
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Авторизуем пользователя
	tokenPair, err := h.service.LoginUser(&req)
	if err != nil {
		if strings.Contains(err.Error(), "неверная пара email/пароль") {
			http.Error(w, "неверная пара email/пароль", http.StatusBadRequest)
			return
		}
		if strings.Contains(err.Error(), "обязателен") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Устанавливаем refresh token в cookie
	utils.SetRefreshTokenCookie(w, tokenPair.RefreshToken, h.refreshTokenTTL)

	// Возвращаем access token в теле ответа
	h.sendTokenResponse(w, tokenPair.AccessToken)
}

// Refresh обрабатывает POST /api/v1/user/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Получаем refresh token из cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "refresh токен отсутствует", http.StatusUnauthorized)
		return
	}

	refreshTokenString := cookie.Value
	if refreshTokenString == "" {
		http.Error(w, "refresh токен отсутствует", http.StatusUnauthorized)
		return
	}

	// Обновляем токены
	tokenPair, err := h.service.RefreshTokens(refreshTokenString)
	if err != nil {
		if strings.Contains(err.Error(), "недействительный refresh токен") ||
			strings.Contains(err.Error(), "refresh токен отсутствует") ||
			strings.Contains(err.Error(), "пользователь не найден") {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Устанавливаем новый refresh token в cookie
	utils.SetRefreshTokenCookie(w, tokenPair.RefreshToken, h.refreshTokenTTL)

	// Возвращаем access token в теле ответа
	h.sendTokenResponse(w, tokenPair.AccessToken)
}

// Logout обрабатывает POST /api/v1/user/logout
// Удаляет refresh токен текущего устройства из cookie и БД
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Получаем refresh token из cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		// Если cookie нет, считаем что logout уже выполнен
		w.WriteHeader(http.StatusOK)
		return
	}

	refreshTokenString := cookie.Value
	if refreshTokenString == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Удаляем refresh токен из БД
	if err := h.service.Logout(refreshTokenString); err != nil {
		// Если токен не найден или уже удален, все равно считаем успешным logout
		// (токен мог быть удален ранее или истечь)
	}

	// Удаляем cookie
	utils.DeleteRefreshTokenCookie(w)

	w.WriteHeader(http.StatusOK)
}

// LogoutAll обрабатывает POST /api/v1/user/logout-all
func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Токен авторизации отсутствует", http.StatusUnauthorized)
		return
	}

	// Удаляем все refresh токены пользователя
	if err := h.service.LogoutAll(userID); err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Удаляем cookie текущего устройства
	utils.DeleteRefreshTokenCookie(w)

	w.WriteHeader(http.StatusOK)
}
