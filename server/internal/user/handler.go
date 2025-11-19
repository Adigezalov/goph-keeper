package user

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/localization"
	"github.com/Adigezalov/goph-keeper/internal/middleware"
	"github.com/Adigezalov/goph-keeper/internal/tokens"
	"github.com/Adigezalov/goph-keeper/internal/utils"
)

type UserService interface {
	RegisterUser(req *RegisterRequest) (*tokens.TokenPair, error)
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
		log.Printf("Ошибка отправки JSON ответа: %v", err)
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		localization.LocalizedError(w, r, http.StatusBadRequest, "common.invalid_request_format", nil)
		return
	}

	tokenPair, err := h.service.RegisterUser(&req)
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
			log.Printf("Ошибка регистрации пользователя: %v", err)
			localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
			return
		}
	}

	utils.SetRefreshTokenCookie(w, tokenPair.RefreshToken, h.refreshTokenTTL)

	h.sendTokenResponse(w, tokenPair.AccessToken)
}

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
		case errors.Is(err, ErrEmailRequired),
			errors.Is(err, ErrPasswordRequired),
			errors.Is(err, ErrRequestRequired):
			localization.LocalizedError(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		default:
			log.Printf("Ошибка авторизации пользователя: %v", err)
			localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
			return
		}
	}

	utils.SetRefreshTokenCookie(w, tokenPair.RefreshToken, h.refreshTokenTTL)

	h.sendTokenResponse(w, tokenPair.AccessToken)
}

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
			log.Printf("Ошибка обновления токенов: %v", err)
			localization.LocalizedError(w, r, http.StatusInternalServerError, "common.internal_error", nil)
			return
		}
	}

	utils.SetRefreshTokenCookie(w, tokenPair.RefreshToken, h.refreshTokenTTL)

	h.sendTokenResponse(w, tokenPair.AccessToken)
}

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
