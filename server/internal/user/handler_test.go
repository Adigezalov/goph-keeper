package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/middleware"
	"github.com/Adigezalov/goph-keeper/internal/tokens"
)

type MockService struct {
	RegisterUserFunc  func(req *RegisterRequest) (*tokens.TokenPair, error)
	LoginUserFunc     func(req *LoginRequest) (*tokens.TokenPair, error)
	RefreshTokensFunc func(refreshTokenString string) (*tokens.TokenPair, error)
	LogoutFunc        func(refreshTokenString string) error
	LogoutAllFunc     func(userID int) error
}

func (m *MockService) RegisterUser(req *RegisterRequest) (*tokens.TokenPair, error) {
	if m.RegisterUserFunc != nil {
		return m.RegisterUserFunc(req)
	}
	return &tokens.TokenPair{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
	}, nil
}

func (m *MockService) LoginUser(req *LoginRequest) (*tokens.TokenPair, error) {
	if m.LoginUserFunc != nil {
		return m.LoginUserFunc(req)
	}
	return &tokens.TokenPair{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
	}, nil
}

func (m *MockService) RefreshTokens(refreshTokenString string) (*tokens.TokenPair, error) {
	if m.RefreshTokensFunc != nil {
		return m.RefreshTokensFunc(refreshTokenString)
	}
	return &tokens.TokenPair{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}, nil
}

func (m *MockService) Logout(refreshTokenString string) error {
	if m.LogoutFunc != nil {
		return m.LogoutFunc(refreshTokenString)
	}
	return nil
}

func (m *MockService) LogoutAll(userID int) error {
	if m.LogoutAllFunc != nil {
		return m.LogoutAllFunc(userID)
	}
	return nil
}

func TestHandler_Register_Success(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService, 5*time.Minute)

	reqBody := RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус 200, получен %d", w.Code)
	}

	var response TokenResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("не удалось декодировать ответ: %v", err)
	}

	if response.AccessToken == "" {
		t.Error("AccessToken не должен быть пустым")
	}

	cookies := w.Result().Cookies()
	foundRefreshToken := false
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			foundRefreshToken = true
			if cookie.Value == "" {
				t.Error("refresh_token cookie не должен быть пустым")
			}
		}
	}
	if !foundRefreshToken {
		t.Error("refresh_token cookie не найден")
	}
}

func TestHandler_Register_InvalidJSON(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ожидался статус 400, получен %d", w.Code)
	}
}

func TestHandler_Register_EmailRequired(t *testing.T) {
	mockService := &MockService{
		RegisterUserFunc: func(req *RegisterRequest) (*tokens.TokenPair, error) {
			return nil, ErrEmailRequired
		},
	}
	handler := NewHandler(mockService, 5*time.Minute)

	reqBody := RegisterRequest{
		Email:    "",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ожидался статус 400, получен %d", w.Code)
	}
}

func TestHandler_Register_UserAlreadyExists(t *testing.T) {
	mockService := &MockService{
		RegisterUserFunc: func(req *RegisterRequest) (*tokens.TokenPair, error) {
			return nil, ErrUserAlreadyExists
		},
	}
	handler := NewHandler(mockService, 5*time.Minute)

	reqBody := RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("ожидался статус 409, получен %d", w.Code)
	}
}

func TestHandler_Register_InternalError(t *testing.T) {
	mockService := &MockService{
		RegisterUserFunc: func(req *RegisterRequest) (*tokens.TokenPair, error) {
			return nil, errors.New("database connection failed")
		},
	}
	handler := NewHandler(mockService, 5*time.Minute)

	reqBody := RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("ожидался статус 500, получен %d", w.Code)
	}
}

func TestHandler_Login_Success(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService, 5*time.Minute)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус 200, получен %d", w.Code)
	}

	var response TokenResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("не удалось декодировать ответ: %v", err)
	}

	if response.AccessToken == "" {
		t.Error("AccessToken не должен быть пустым")
	}
}

func TestHandler_Login_InvalidJSON(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ожидался статус 400, получен %d", w.Code)
	}
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	mockService := &MockService{
		LoginUserFunc: func(req *LoginRequest) (*tokens.TokenPair, error) {
			return nil, ErrInvalidCredentials
		},
	}
	handler := NewHandler(mockService, 5*time.Minute)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ожидался статус 400, получен %d", w.Code)
	}
}

func TestHandler_Login_EmailRequired(t *testing.T) {
	mockService := &MockService{
		LoginUserFunc: func(req *LoginRequest) (*tokens.TokenPair, error) {
			return nil, ErrEmailRequired
		},
	}
	handler := NewHandler(mockService, 5*time.Minute)

	reqBody := LoginRequest{
		Email:    "",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ожидался статус 400, получен %d", w.Code)
	}
}

func TestHandler_Refresh_Success(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "valid-refresh-token",
	})
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус 200, получен %d", w.Code)
	}

	var response TokenResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("не удалось декодировать ответ: %v", err)
	}

	if response.AccessToken == "" {
		t.Error("AccessToken не должен быть пустым")
	}
}

func TestHandler_Refresh_MissingCookie(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/refresh", nil)
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ожидался статус 401, получен %d", w.Code)
	}
}

func TestHandler_Refresh_EmptyToken(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "",
	})
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ожидался статус 401, получен %d", w.Code)
	}
}

func TestHandler_Refresh_InvalidToken(t *testing.T) {
	mockService := &MockService{
		RefreshTokensFunc: func(refreshTokenString string) (*tokens.TokenPair, error) {
			return nil, ErrInvalidRefreshToken
		},
	}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "invalid-token",
	})
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ожидался статус 401, получен %d", w.Code)
	}
}

func TestHandler_Refresh_UserNotFound(t *testing.T) {
	mockService := &MockService{
		RefreshTokensFunc: func(refreshTokenString string) (*tokens.TokenPair, error) {
			return nil, ErrUserNotFound
		},
	}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "valid-token-but-user-deleted",
	})
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ожидался статус 401, получен %d", w.Code)
	}
}

func TestHandler_Refresh_InternalError(t *testing.T) {
	mockService := &MockService{
		RefreshTokensFunc: func(refreshTokenString string) (*tokens.TokenPair, error) {
			return nil, errors.New("database error")
		},
	}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "valid-token",
	})
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("ожидался статус 500, получен %d", w.Code)
	}
}

func TestHandler_Logout_Success(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "valid-refresh-token",
	})
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус 200, получен %d", w.Code)
	}

	cookies := w.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			if cookie.MaxAge != -1 {
				t.Error("refresh_token cookie должен быть удален (MaxAge=-1)")
			}
		}
	}
}

func TestHandler_Logout_NoCookie(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/logout", nil)
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус 200, получен %d (logout без cookie считается успешным)", w.Code)
	}
}

func TestHandler_Logout_EmptyToken(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "",
	})
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус 200, получен %d", w.Code)
	}
}

func TestHandler_Logout_ServiceError(t *testing.T) {
	mockService := &MockService{
		LogoutFunc: func(refreshTokenString string) error {
			return errors.New("token not found")
		},
	}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "some-token",
	})
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус 200, получен %d", w.Code)
	}
}

func TestHandler_LogoutAll_Success(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/logout-all", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.LogoutAll(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус 200, получен %d", w.Code)
	}
}

func TestHandler_LogoutAll_NoUserID(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/logout-all", nil)
	w := httptest.NewRecorder()

	handler.LogoutAll(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ожидался статус 401, получен %d", w.Code)
	}
}

func TestHandler_LogoutAll_ServiceError(t *testing.T) {
	mockService := &MockService{
		LogoutAllFunc: func(userID int) error {
			return errors.New("database error")
		},
	}
	handler := NewHandler(mockService, 5*time.Minute)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/logout-all", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.LogoutAll(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("ожидался статус 500, получен %d", w.Code)
	}
}
