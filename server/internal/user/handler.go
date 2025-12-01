package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/localization"
	"github.com/Adigezalov/goph-keeper/internal/logger"
	"github.com/Adigezalov/goph-keeper/internal/middleware"
	"github.com/Adigezalov/goph-keeper/internal/tokens"
	"github.com/Adigezalov/goph-keeper/internal/utils"
	"github.com/Adigezalov/goph-keeper/internal/verification"
)

type UserService interface {
	RegisterUser(req *RegisterRequest) error
	VerifyEmail(req *verification.VerifyEmailRequest) (*tokens.TokenPair, error)
	ResendVerificationCode(req *verification.ResendCodeRequest) error
	LoginUser(req *LoginRequest) (*tokens.TokenPair, error)
	RefreshTokens(refreshTokenString string) (*tokens.TokenPair, error)
	Logout(refreshTokenString string) error
	LogoutAll(userID int) error
}

type Handler struct {
	service         UserService
	refreshTokenTTL time.Duration
}

func NewHandler(service UserService, refreshTokenTTL time.Duration) *Handler {
	return &Handler{
		service:         service,
		refreshTokenTTL: refreshTokenTTL,
	}
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *Handler) sendTokenResponse(w http.ResponseWriter, accessToken string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(TokenResponse{
		AccessToken: accessToken,
	}); err != nil {
		logger.Errorf("[User] Ошибка отправки JSON ответа: %v", err)
	}
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя и отправляет код верификации на email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Данные для регистрации"
// @Success 200 {object} map[string]string "Код верификации отправлен"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 409 {object} map[string]string "Пользователь уже существует"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /user/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		localization.LocalizedError(w, r, http.StatusBadRequest, "common.invalid_request_format", nil)
		return
	}

	err := h.service.RegisterUser(&req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserAlreadyExists):
			localization.LocalizedError(w, r, http.StatusConflict, "user.email_already_taken", nil)
			return
		case errors.Is(err, ErrEmailRequired),
			errors.Is(err, ErrPasswordRequired),
			errors.Is(err, ErrInvalidEmail),
			errors.Is(err, ErrPasswordTooShort),
			errors.Is(err, ErrRequestRequired):
			localization.LocalizedError(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		default:
			logger.Log.WithFields(map[string]interface{}{
				"error": err.Error(),
			}).Error("[User] Ошибка регистрации пользователя")
			localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "verification_code_sent",
	})
}

// VerifyEmail godoc
// @Summary Верификация email
// @Description Проверяет код верификации и возвращает токены доступа
// @Tags auth
// @Accept json
// @Produce json
// @Param request body verification.VerifyEmailRequest true "Email и код верификации"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 401 {object} map[string]string "Неверный код"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /user/verify-email [post]
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verification.VerifyEmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		localization.LocalizedError(w, r, http.StatusBadRequest, "common.invalid_request_format", nil)
		return
	}

	tokenPair, err := h.service.VerifyEmail(&req)
	if err != nil {
		switch {
		case errors.Is(err, verification.ErrInvalidCode),
			errors.Is(err, verification.ErrCodeExpired),
			errors.Is(err, verification.ErrCodeAlreadyUsed):
			localization.LocalizedError(w, r, http.StatusUnauthorized, err.Error(), nil)
			return
		case errors.Is(err, verification.ErrEmailRequired),
			errors.Is(err, verification.ErrCodeRequired),
			errors.Is(err, verification.ErrRequestRequired):
			localization.LocalizedError(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		default:
			logger.Log.WithFields(map[string]interface{}{
				"error": err.Error(),
			}).Error("[User] Ошибка верификации email")
			localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
			return
		}
	}

	utils.SetRefreshTokenCookie(w, tokenPair.RefreshToken, h.refreshTokenTTL)
	h.sendTokenResponse(w, tokenPair.AccessToken)
}

// ResendCode godoc
// @Summary Повторная отправка кода верификации
// @Description Отправляет новый код верификации на email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body verification.ResendCodeRequest true "Email"
// @Success 200 {object} map[string]string "Код отправлен"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /user/resend-code [post]
func (h *Handler) ResendCode(w http.ResponseWriter, r *http.Request) {
	var req verification.ResendCodeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		localization.LocalizedError(w, r, http.StatusBadRequest, "common.invalid_request_format", nil)
		return
	}

	err := h.service.ResendVerificationCode(&req)
	if err != nil {
		switch {
		case errors.Is(err, verification.ErrEmailRequired),
			errors.Is(err, verification.ErrRequestRequired):
			localization.LocalizedError(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		default:
			logger.Log.WithFields(map[string]interface{}{
				"error": err.Error(),
			}).Error("[User] Ошибка повторной отправки кода")
			localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "verification_code_sent",
	})
}

// Login godoc
// @Summary Вход пользователя
// @Description Аутентифицирует пользователя и возвращает access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Данные для входа"
// @Success 200 {object} map[string]string "access_token"
// @Failure 400 {object} map[string]string "Неверные учетные данные"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /user/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		localization.LocalizedError(w, r, http.StatusBadRequest, "common.invalid_request_format", nil)
		return
	}

	tokenPair, err := h.service.LoginUser(&req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			localization.LocalizedError(w, r, http.StatusBadRequest, "user.invalid_credentials", nil)
			return
		case errors.Is(err, ErrEmailNotVerified):
			localization.LocalizedError(w, r, http.StatusBadRequest, "user.email_not_verified", nil)
			return
		case errors.Is(err, ErrEmailRequired),
			errors.Is(err, ErrPasswordRequired),
			errors.Is(err, ErrRequestRequired):
			localization.LocalizedError(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		default:
			logger.Log.WithFields(map[string]interface{}{
				"error": err.Error(),
			}).Error("[User] Ошибка авторизации пользователя")
			localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
			return
		}
	}

	utils.SetRefreshTokenCookie(w, tokenPair.RefreshToken, h.refreshTokenTTL)

	h.sendTokenResponse(w, tokenPair.AccessToken)
}

// Refresh godoc
// @Summary Обновление токена доступа
// @Description Обновляет access token используя refresh token из cookie
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string "access_token"
// @Failure 401 {object} map[string]string "Ошибка авторизации"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /user/refresh [get]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		localization.LocalizedError(w, r, http.StatusUnauthorized, "common.authorization_error", nil)
		return
	}

	refreshTokenString := cookie.Value
	if refreshTokenString == "" {
		localization.LocalizedError(w, r, http.StatusUnauthorized, "common.authorization_error", nil)
		return
	}

	tokenPair, err := h.service.RefreshTokens(refreshTokenString)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRefreshToken),
			errors.Is(err, ErrRefreshTokenMissing),
			errors.Is(err, ErrUserNotFound):
			localization.LocalizedError(w, r, http.StatusUnauthorized, "common.authorization_error", nil)
			return
		default:
			logger.Log.WithFields(map[string]interface{}{
				"error": err.Error(),
			}).Error("[User] Ошибка обновления токенов")
			localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
			return
		}
	}

	utils.SetRefreshTokenCookie(w, tokenPair.RefreshToken, h.refreshTokenTTL)

	h.sendTokenResponse(w, tokenPair.AccessToken)
}

// Logout godoc
// @Summary Выход пользователя
// @Description Выход из текущей сессии (удаляет refresh token)
// @Tags auth
// @Produce json
// @Success 200 "Успешный выход"
// @Router /user/logout [get]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	refreshTokenString := cookie.Value
	if refreshTokenString == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.service.Logout(refreshTokenString); err != nil {
	}

	utils.DeleteRefreshTokenCookie(w)

	w.WriteHeader(http.StatusOK)
}

// LogoutAll godoc
// @Summary Выход из всех сессий
// @Description Выход из всех сессий пользователя (удаляет все refresh tokens)
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 "Успешный выход из всех сессий"
// @Failure 401 {object} map[string]string "Ошибка авторизации"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /user/logout-all [get]
func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		localization.LocalizedError(w, r, http.StatusUnauthorized, "common.authorization_error", nil)
		return
	}

	if err := h.service.LogoutAll(userID); err != nil {
		localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
		return
	}

	utils.DeleteRefreshTokenCookie(w)

	w.WriteHeader(http.StatusOK)
}
