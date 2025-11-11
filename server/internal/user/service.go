package user

import (
	"errors"
	"net/mail"

	"github.com/Adigezalov/goph-keeper/internal/tokens"
	"golang.org/x/crypto/bcrypt"
)

// Service содержит бизнес-логику для пользователей
type Service struct {
	repo         Repository
	tokenService *tokens.Service
}

// NewService создает новый экземпляр Service
func NewService(repo Repository, tokenService *tokens.Service) *Service {
	return &Service{
		repo:         repo,
		tokenService: tokenService,
	}
}

// RegisterUser регистрирует нового пользователя
func (s *Service) RegisterUser(req *RegisterRequest) (*tokens.TokenPair, error) {
	// Валидация входных данных
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Проверяем, не существует ли уже пользователь с таким email
	existingUser, err := s.repo.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, WrapError(err, "не удалось захешировать пароль")
	}

	// Создаем пользователя
	user := &User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.repo.CreateUser(user); err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		return nil, WrapError(err, "не удалось создать пользователя")
	}

	// Генерируем токены
	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, WrapError(err, "не удалось создать токены")
	}

	return tokenPair, nil
}

// validateRegisterRequest валидирует запрос на регистрацию
func (s *Service) validateRegisterRequest(req *RegisterRequest) error {
	if req == nil {
		return ErrRequestRequired
	}

	if req.Email == "" {
		return ErrEmailRequired
	}

	if req.Password == "" {
		return ErrPasswordRequired
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		return ErrInvalidEmail
	}

	if len(req.Password) < 6 {
		return ErrPasswordTooShort
	}

	return nil
}

// LoginUser авторизует пользователя
func (s *Service) LoginUser(req *LoginRequest) (*tokens.TokenPair, error) {
	// Валидация входных данных
	if err := s.validateLoginRequest(req); err != nil {
		return nil, err
	}

	// Получаем пользователя по email
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		// Не раскрываем, существует ли пользователь (безопасность)
		return nil, ErrInvalidCredentials
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Генерируем новые токены (старые refresh токены автоматически удалятся)
	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, WrapError(err, "не удалось создать токены")
	}

	return tokenPair, nil
}

// validateLoginRequest валидирует запрос на авторизацию
func (s *Service) validateLoginRequest(req *LoginRequest) error {
	if req == nil {
		return ErrRequestRequired
	}

	if req.Email == "" {
		return ErrEmailRequired
	}

	if req.Password == "" {
		return ErrPasswordRequired
	}

	return nil
}

// RefreshTokens обновляет пару токенов на основе валидного refresh токена
func (s *Service) RefreshTokens(refreshTokenString string) (*tokens.TokenPair, error) {
	if refreshTokenString == "" {
		return nil, ErrRefreshTokenMissing
	}

	// Получаем refresh токен из БД
	refreshToken, err := s.tokenService.GetRefreshToken(refreshTokenString)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Получаем пользователя по ID из refresh токена
	user, err := s.repo.GetUserByID(refreshToken.UserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, WrapError(err, "не удалось получить пользователя")
	}

	// Обновляем пару токенов
	tokenPair, err := s.tokenService.RefreshTokenPair(refreshTokenString, user.ID, user.Email)
	if err != nil {
		return nil, WrapError(err, "не удалось обновить токены")
	}

	return tokenPair, nil
}

// Logout удаляет refresh токен текущего устройства
func (s *Service) Logout(refreshTokenString string) error {
	return s.tokenService.Logout(refreshTokenString)
}

// LogoutAll удаляет все refresh токены пользователя
func (s *Service) LogoutAll(userID int) error {
	return s.tokenService.LogoutAll(userID)
}
