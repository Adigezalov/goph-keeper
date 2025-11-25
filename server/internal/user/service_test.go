package user

import (
	"errors"
	"testing"

	"github.com/Adigezalov/goph-keeper/internal/tokens"
	"golang.org/x/crypto/bcrypt"
)

type MockRepository struct {
	GetUserByEmailFunc func(email string) (*User, error)
	CreateUserFunc     func(user *User) error
	GetUserByIDFunc    func(id int) (*User, error)
}

func (m *MockRepository) GetUserByEmail(email string) (*User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(email)
	}
	return nil, ErrUserNotFound
}

func (m *MockRepository) CreateUser(user *User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(user)
	}
	user.ID = 1
	return nil
}

func (m *MockRepository) GetUserByID(id int) (*User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(id)
	}
	return nil, ErrUserNotFound
}

type MockTokenService struct {
	GenerateTokenPairFunc func(userID int, email string) (*tokens.TokenPair, error)
	GetRefreshTokenFunc   func(tokenString string) (*tokens.RefreshToken, error)
	RefreshTokenPairFunc  func(refreshTokenString string, userID int, email string) (*tokens.TokenPair, error)
	LogoutFunc            func(refreshTokenString string) error
	LogoutAllFunc         func(userID int) error
}

func (m *MockTokenService) GenerateTokenPair(userID int, email string) (*tokens.TokenPair, error) {
	if m.GenerateTokenPairFunc != nil {
		return m.GenerateTokenPairFunc(userID, email)
	}
	return &tokens.TokenPair{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
	}, nil
}

func (m *MockTokenService) GetRefreshToken(tokenString string) (*tokens.RefreshToken, error) {
	if m.GetRefreshTokenFunc != nil {
		return m.GetRefreshTokenFunc(tokenString)
	}
	return &tokens.RefreshToken{
		Token:  tokenString,
		UserID: 1,
	}, nil
}

func (m *MockTokenService) RefreshTokenPair(refreshTokenString string, userID int, email string) (*tokens.TokenPair, error) {
	if m.RefreshTokenPairFunc != nil {
		return m.RefreshTokenPairFunc(refreshTokenString, userID, email)
	}
	return &tokens.TokenPair{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}, nil
}

func (m *MockTokenService) Logout(refreshTokenString string) error {
	if m.LogoutFunc != nil {
		return m.LogoutFunc(refreshTokenString)
	}
	return nil
}

func (m *MockTokenService) LogoutAll(userID int) error {
	if m.LogoutAllFunc != nil {
		return m.LogoutAllFunc(userID)
	}
	return nil
}

func (m *MockTokenService) ValidateAccessToken() (*tokens.Claims, error) {
	return nil, nil
}

func TestService_RegisterUser_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	tokenPair, err := service.RegisterUser(req)

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
}

func TestService_RegisterUser_EmailRequired(t *testing.T) {
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	req := &RegisterRequest{
		Email:    "",
		Password: "password123",
	}

	_, err := service.RegisterUser(req)

	if !errors.Is(err, ErrEmailRequired) {
		t.Errorf("ожидалась ошибка ErrEmailRequired, получена: %v", err)
	}
}

func TestService_RegisterUser_PasswordRequired(t *testing.T) {
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: "",
	}

	_, err := service.RegisterUser(req)

	if !errors.Is(err, ErrPasswordRequired) {
		t.Errorf("ожидалась ошибка ErrPasswordRequired, получена: %v", err)
	}
}

func TestService_RegisterUser_InvalidEmail(t *testing.T) {
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	req := &RegisterRequest{
		Email:    "invalid-email",
		Password: "password123",
	}

	_, err := service.RegisterUser(req)

	if !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("ожидалась ошибка ErrInvalidEmail, получена: %v", err)
	}
}

func TestService_RegisterUser_PasswordTooShort(t *testing.T) {
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: "123",
	}

	_, err := service.RegisterUser(req)

	if !errors.Is(err, ErrPasswordTooShort) {
		t.Errorf("ожидалась ошибка ErrPasswordTooShort, получена: %v", err)
	}
}

func TestService_RegisterUser_UserAlreadyExists(t *testing.T) {
	mockRepo := &MockRepository{
		GetUserByEmailFunc: func(email string) (*User, error) {
			return &User{
				ID:    1,
				Email: email,
			}, nil
		},
	}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	req := &RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	_, err := service.RegisterUser(req)

	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Errorf("ожидалась ошибка ErrUserAlreadyExists, получена: %v", err)
	}
}

func TestService_RegisterUser_NilRequest(t *testing.T) {
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	_, err := service.RegisterUser(nil)

	if !errors.Is(err, ErrRequestRequired) {
		t.Errorf("ожидалась ошибка ErrRequestRequired, получена: %v", err)
	}
}

func TestService_LoginUser_Success(t *testing.T) {
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockRepo := &MockRepository{
		GetUserByEmailFunc: func(email string) (*User, error) {
			return &User{
				ID:           1,
				Email:        email,
				PasswordHash: string(hashedPassword),
			}, nil
		},
	}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	req := &LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	tokenPair, err := service.LoginUser(req)

	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка: %v", err)
	}

	if tokenPair == nil {
		t.Fatal("tokenPair не должен быть nil")
	}
}

func TestService_LoginUser_InvalidCredentials_WrongPassword(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	mockRepo := &MockRepository{
		GetUserByEmailFunc: func(email string) (*User, error) {
			return &User{
				ID:           1,
				Email:        email,
				PasswordHash: string(hashedPassword),
			}, nil
		},
	}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	_, err := service.LoginUser(req)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("ожидалась ошибка ErrInvalidCredentials, получена: %v", err)
	}
}

func TestService_LoginUser_InvalidCredentials_UserNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		GetUserByEmailFunc: func(email string) (*User, error) {
			return nil, ErrUserNotFound
		},
	}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	req := &LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	_, err := service.LoginUser(req)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("ожидалась ошибка ErrInvalidCredentials, получена: %v", err)
	}
}

func TestService_LoginUser_EmailRequired(t *testing.T) {
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	req := &LoginRequest{
		Email:    "",
		Password: "password123",
	}

	_, err := service.LoginUser(req)

	if !errors.Is(err, ErrEmailRequired) {
		t.Errorf("ожидалась ошибка ErrEmailRequired, получена: %v", err)
	}
}

func TestService_LoginUser_PasswordRequired(t *testing.T) {
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "",
	}

	_, err := service.LoginUser(req)

	if !errors.Is(err, ErrPasswordRequired) {
		t.Errorf("ожидалась ошибка ErrPasswordRequired, получена: %v", err)
	}
}

func TestService_LoginUser_NilRequest(t *testing.T) {
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	_, err := service.LoginUser(nil)

	if !errors.Is(err, ErrRequestRequired) {
		t.Errorf("ожидалась ошибка ErrRequestRequired, получена: %v", err)
	}
}

func TestService_RefreshTokens_Success(t *testing.T) {
	mockRepo := &MockRepository{
		GetUserByIDFunc: func(id int) (*User, error) {
			return &User{
				ID:    id,
				Email: "test@example.com",
			}, nil
		},
	}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	tokenPair, err := service.RefreshTokens("valid-refresh-token")

	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка: %v", err)
	}

	if tokenPair == nil {
		t.Fatal("tokenPair не должен быть nil")
	}
}

func TestService_RefreshTokens_MissingToken(t *testing.T) {
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	_, err := service.RefreshTokens("")

	if !errors.Is(err, ErrRefreshTokenMissing) {
		t.Errorf("ожидалась ошибка ErrRefreshTokenMissing, получена: %v", err)
	}
}

func TestService_RefreshTokens_InvalidToken(t *testing.T) {
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{
		GetRefreshTokenFunc: func(tokenString string) (*tokens.RefreshToken, error) {
			return nil, errors.New("token not found")
		},
	}
	service := NewService(mockRepo, mockTokenService)

	_, err := service.RefreshTokens("invalid-token")

	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("ожидалась ошибка ErrInvalidRefreshToken, получена: %v", err)
	}
}

func TestService_RefreshTokens_UserNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		GetUserByIDFunc: func(id int) (*User, error) {
			return nil, ErrUserNotFound
		},
	}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	_, err := service.RefreshTokens("valid-refresh-token")

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestService_Logout_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	err := service.Logout("valid-refresh-token")

	if err != nil {
		t.Errorf("ожидался успех, получена ошибка: %v", err)
	}
}

func TestService_Logout_Error(t *testing.T) {
	expectedErr := errors.New("logout failed")
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{
		LogoutFunc: func(refreshTokenString string) error {
			return expectedErr
		},
	}
	service := NewService(mockRepo, mockTokenService)

	err := service.Logout("some-token")

	if !errors.Is(err, expectedErr) {
		t.Errorf("ожидалась ошибка %v, получена: %v", expectedErr, err)
	}
}

func TestService_LogoutAll_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{}
	service := NewService(mockRepo, mockTokenService)

	err := service.LogoutAll(1)

	if err != nil {
		t.Errorf("ожидался успех, получена ошибка: %v", err)
	}
}

func TestService_LogoutAll_Error(t *testing.T) {
	expectedErr := errors.New("logoutAll failed")
	mockRepo := &MockRepository{}
	mockTokenService := &MockTokenService{
		LogoutAllFunc: func(userID int) error {
			return expectedErr
		},
	}
	service := NewService(mockRepo, mockTokenService)

	err := service.LogoutAll(1)

	if !errors.Is(err, expectedErr) {
		t.Errorf("ожидалась ошибка %v, получена: %v", expectedErr, err)
	}
}
