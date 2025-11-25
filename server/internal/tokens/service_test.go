package tokens

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MockRepository struct {
	SaveRefreshTokenFunc        func(token *RefreshToken) error
	GetRefreshTokenFunc         func(token string) (*RefreshToken, error)
	DeleteRefreshTokenFunc      func(token string) error
	DeleteUserRefreshTokensFunc func(userID int) error
}

func (m *MockRepository) SaveRefreshToken(token *RefreshToken) error {
	if m.SaveRefreshTokenFunc != nil {
		return m.SaveRefreshTokenFunc(token)
	}
	token.ID = 1
	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()
	return nil
}

func (m *MockRepository) GetRefreshToken(tokenString string) (*RefreshToken, error) {
	if m.GetRefreshTokenFunc != nil {
		return m.GetRefreshTokenFunc(tokenString)
	}
	return &RefreshToken{
		ID:        1,
		Token:     tokenString,
		UserID:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *MockRepository) DeleteRefreshToken(tokenString string) error {
	if m.DeleteRefreshTokenFunc != nil {
		return m.DeleteRefreshTokenFunc(tokenString)
	}
	return nil
}

func (m *MockRepository) DeleteUserRefreshTokens(userID int) error {
	if m.DeleteUserRefreshTokensFunc != nil {
		return m.DeleteUserRefreshTokensFunc(userID)
	}
	return nil
}

func TestService_GenerateTokenPair_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	tokenPair, err := service.GenerateTokenPair(1, "test@example.com")

	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка: %v", err)
	}

	if tokenPair == nil {
		t.Fatal("tokenPair не должен быть nil")
	}

	if tokenPair.AccessToken == "" {
		t.Error("AccessToken не должен быть пустым")
	}

	if tokenPair.RefreshToken == "" {
		t.Error("RefreshToken не должен быть пустым")
	}

	token, err := jwt.Parse(tokenPair.AccessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})

	if err != nil {
		t.Errorf("access токен должен быть валидным JWT: %v", err)
	}

	if !token.Valid {
		t.Error("access токен должен быть валидным")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if claims["user_id"].(float64) != 1 {
			t.Errorf("ожидался user_id=1, получен %v", claims["user_id"])
		}
		if claims["email"].(string) != "test@example.com" {
			t.Errorf("ожидался email=test@example.com, получен %v", claims["email"])
		}
		if claims["type"].(string) != "access" {
			t.Errorf("ожидался type=access, получен %v", claims["type"])
		}
	}

	if len(tokenPair.RefreshToken) != 64 {
		t.Errorf("refresh токен должен быть 64 символа (hex), получен %d", len(tokenPair.RefreshToken))
	}
}

func TestService_GenerateTokenPair_SaveError(t *testing.T) {
	expectedErr := errors.New("database error")
	mockRepo := &MockRepository{
		SaveRefreshTokenFunc: func(token *RefreshToken) error {
			return expectedErr
		},
	}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	_, err := service.GenerateTokenPair(1, "test@example.com")

	if err == nil {
		t.Fatal("ожидалась ошибка")
	}

	if !strings.Contains(err.Error(), "не удалось сохранить refresh токен") {
		t.Errorf("ожидалось сообщение о сохранении токена, получено: %v", err)
	}
}

func TestService_ValidateAccessToken_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	tokenPair, _ := service.GenerateTokenPair(42, "test@example.com")

	claims, err := service.ValidateAccessToken(tokenPair.AccessToken)

	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка: %v", err)
	}

	if claims == nil {
		t.Fatal("claims не должен быть nil")
	}

	if claims.UserID != 42 {
		t.Errorf("ожидался UserID=42, получен %d", claims.UserID)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("ожидался Email=test@example.com, получен %s", claims.Email)
	}

	if claims.Type != "access" {
		t.Errorf("ожидался Type=access, получен %s", claims.Type)
	}
}

func TestService_ValidateAccessToken_InvalidToken(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	_, err := service.ValidateAccessToken("invalid-token")

	if err == nil {
		t.Fatal("ожидалась ошибка для невалидного токена")
	}
}

func TestService_ValidateAccessToken_WrongSecret(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	claims := jwt.MapClaims{
		"user_id": 1,
		"email":   "test@example.com",
		"type":    "access",
		"exp":     time.Now().Add(10 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("wrong-secret"))

	_, err := service.ValidateAccessToken(tokenString)

	if err == nil {
		t.Fatal("ожидалась ошибка для токена с неверным секретом")
	}
}

func TestService_ValidateAccessToken_ExpiredToken(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	claims := jwt.MapClaims{
		"user_id": 1,
		"email":   "test@example.com",
		"type":    "access",
		"exp":     time.Now().Add(-1 * time.Hour).Unix(),
		"iat":     time.Now().Add(-2 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	_, err := service.ValidateAccessToken(tokenString)

	if err == nil {
		t.Fatal("ожидалась ошибка для истекшего токена")
	}
}

func TestService_ValidateAccessToken_WrongType(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	claims := jwt.MapClaims{
		"user_id": 1,
		"email":   "test@example.com",
		"type":    "refresh",
		"exp":     time.Now().Add(10 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	_, err := service.ValidateAccessToken(tokenString)

	if err == nil {
		t.Fatal("ожидалась ошибка для токена с неверным типом")
	}

	if !strings.Contains(err.Error(), "Неверный тип токена") {
		t.Errorf("ожидалось сообщение о неверном типе токена, получено: %v", err)
	}
}

func TestService_ValidateAccessToken_WrongSigningMethod(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	claims := jwt.MapClaims{
		"user_id": 1,
		"email":   "test@example.com",
		"type":    "access",
		"exp":     time.Now().Add(10 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := service.ValidateAccessToken(tokenString)

	if err == nil {
		t.Fatal("ожидалась ошибка для токена с неверным методом подписи")
	}
}

func TestService_GetRefreshToken_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	token, err := service.GetRefreshToken("test-token")

	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка: %v", err)
	}

	if token == nil {
		t.Fatal("token не должен быть nil")
	}

	if token.Token != "test-token" {
		t.Errorf("ожидался Token=test-token, получен %s", token.Token)
	}
}

func TestService_GetRefreshToken_NotFound(t *testing.T) {
	expectedErr := errors.New("токен не найден")
	mockRepo := &MockRepository{
		GetRefreshTokenFunc: func(token string) (*RefreshToken, error) {
			return nil, expectedErr
		},
	}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	_, err := service.GetRefreshToken("nonexistent-token")

	if err == nil {
		t.Fatal("ожидалась ошибка")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("ожидалась ошибка %v, получена: %v", expectedErr, err)
	}
}

func TestService_RefreshTokenPair_Success(t *testing.T) {
	mockRepo := &MockRepository{
		GetRefreshTokenFunc: func(token string) (*RefreshToken, error) {
			return &RefreshToken{
				ID:     1,
				Token:  token,
				UserID: 1,
			}, nil
		},
	}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	tokenPair, err := service.RefreshTokenPair("old-token", 1, "test@example.com")

	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка: %v", err)
	}

	if tokenPair == nil {
		t.Fatal("tokenPair не должен быть nil")
	}

	if tokenPair.AccessToken == "" {
		t.Error("AccessToken не должен быть пустым")
	}

	if tokenPair.RefreshToken == "" {
		t.Error("RefreshToken не должен быть пустым")
	}

	if tokenPair.RefreshToken == "old-token" {
		t.Error("RefreshToken должен быть новым, не таким как старый")
	}
}

func TestService_RefreshTokenPair_TokenNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		GetRefreshTokenFunc: func(token string) (*RefreshToken, error) {
			return nil, errors.New("token not found")
		},
	}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	_, err := service.RefreshTokenPair("nonexistent-token", 1, "test@example.com")

	if err == nil {
		t.Fatal("ожидалась ошибка")
	}

	if !strings.Contains(err.Error(), "Недействительный refresh токен") {
		t.Errorf("ожидалось сообщение о недействительном токене, получено: %v", err)
	}
}

func TestService_RefreshTokenPair_WrongUserID(t *testing.T) {
	mockRepo := &MockRepository{
		GetRefreshTokenFunc: func(token string) (*RefreshToken, error) {
			return &RefreshToken{
				ID:     1,
				Token:  token,
				UserID: 1,
			}, nil
		},
	}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	_, err := service.RefreshTokenPair("token", 2, "test@example.com")

	if err == nil {
		t.Fatal("ожидалась ошибка")
	}

	if !strings.Contains(err.Error(), "Недействительный refresh токен") {
		t.Errorf("ожидалось сообщение о недействительном токене, получено: %v", err)
	}
}

func TestService_RefreshTokenPair_DeleteError(t *testing.T) {
	mockRepo := &MockRepository{
		GetRefreshTokenFunc: func(token string) (*RefreshToken, error) {
			return &RefreshToken{
				ID:     1,
				Token:  token,
				UserID: 1,
			}, nil
		},
		DeleteRefreshTokenFunc: func(token string) error {
			return errors.New("delete failed")
		},
	}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	_, err := service.RefreshTokenPair("token", 1, "test@example.com")

	if err == nil {
		t.Fatal("ожидалась ошибка")
	}

	if !strings.Contains(err.Error(), "не удалось удалить старый refresh токен") {
		t.Errorf("ожидалось сообщение об ошибке удаления, получено: %v", err)
	}
}

func TestService_Logout_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	err := service.Logout("valid-token")

	if err != nil {
		t.Errorf("ожидался успех, получена ошибка: %v", err)
	}
}

func TestService_Logout_EmptyToken(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	err := service.Logout("")

	if err == nil {
		t.Fatal("ожидалась ошибка для пустого токена")
	}

	if !strings.Contains(err.Error(), "Refresh токен отсутствует") {
		t.Errorf("ожидалось сообщение об отсутствии токена, получено: %v", err)
	}
}

func TestService_Logout_TokenNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		GetRefreshTokenFunc: func(token string) (*RefreshToken, error) {
			return nil, errors.New("token not found")
		},
	}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	err := service.Logout("nonexistent-token")

	if err == nil {
		t.Fatal("ожидалась ошибка")
	}

	if !strings.Contains(err.Error(), "Недействительный refresh токен") {
		t.Errorf("ожидалось сообщение о недействительном токене, получено: %v", err)
	}
}

func TestService_Logout_DeleteError(t *testing.T) {
	mockRepo := &MockRepository{
		DeleteRefreshTokenFunc: func(token string) error {
			return errors.New("delete failed")
		},
	}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	err := service.Logout("token")

	if err == nil {
		t.Fatal("ожидалась ошибка")
	}

	if !strings.Contains(err.Error(), "не удалось удалить refresh токен") {
		t.Errorf("ожидалось сообщение об ошибке удаления, получено: %v", err)
	}
}

func TestService_LogoutAll_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	err := service.LogoutAll(1)

	if err != nil {
		t.Errorf("ожидался успех, получена ошибка: %v", err)
	}
}

func TestService_LogoutAll_Error(t *testing.T) {
	expectedErr := errors.New("delete failed")
	mockRepo := &MockRepository{
		DeleteUserRefreshTokensFunc: func(userID int) error {
			return expectedErr
		},
	}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	err := service.LogoutAll(1)

	if err == nil {
		t.Fatal("ожидалась ошибка")
	}

	if !strings.Contains(err.Error(), "не удалось удалить refresh токены пользователя") {
		t.Errorf("ожидалось сообщение об ошибке удаления токенов, получено: %v", err)
	}
}

func TestService_GenerateRefreshToken_UniqueTokens(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, "test-secret", 10*time.Minute, 24*time.Hour)

	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := service.generateRefreshToken()
		if err != nil {
			t.Fatalf("не удалось сгенерировать токен: %v", err)
		}

		if tokens[token] {
			t.Errorf("токен %s сгенерирован дважды", token)
		}
		tokens[token] = true

		if len(token) != 64 {
			t.Errorf("токен должен быть 64 символа, получен %d", len(token))
		}
	}
}
