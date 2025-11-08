package user

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

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
		return nil, fmt.Errorf("пользователь уже существует")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("не удалось захешировать пароль: %w", err)
	}

	// Создаем пользователя
	user := &User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.repo.CreateUser(user); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, fmt.Errorf("пользователь уже существует")
		}
		return nil, fmt.Errorf("не удалось создать пользователя: %w", err)
	}

	// Генерируем токены
	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать токены: %w", err)
	}

	return tokenPair, nil
}

// validateRegisterRequest валидирует запрос на регистрацию
func (s *Service) validateRegisterRequest(req *RegisterRequest) error {
	if req == nil {
		return errors.New("запрос обязателен")
	}

	if strings.TrimSpace(req.Email) == "" {
		return errors.New("email обязателен")
	}

	if strings.TrimSpace(req.Password) == "" {
		return errors.New("пароль обязателен")
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		return errors.New("неверный формат email")
	}

	if len(req.Password) < 6 {
		return errors.New("пароль должен содержать минимум 6 символов")
	}

	return nil
}

// LoginUser авторизует пользователя
func (s *Service) LoginUser(req *LoginRequest) (*tokens.TokenPair, error) {
	// Валидация входных данных
	if err := s.validateLoginRequest(req); err != nil {
		return nil, err
	}

	// Получаем пользователя по логину
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("неверная пара email/пароль")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("неверная пара логин/пароль")
	}

	// Генерируем новые токены (старые refresh токены автоматически удалятся)
	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать токены: %w", err)
	}

	return tokenPair, nil
}

// validateLoginRequest валидирует запрос на авторизацию
func (s *Service) validateLoginRequest(req *LoginRequest) error {
	if req == nil {
		return errors.New("запрос обязателен")
	}

	if strings.TrimSpace(req.Email) == "" {
		return errors.New("email обязателен")
	}

	if strings.TrimSpace(req.Password) == "" {
		return errors.New("пароль обязателен")
	}

	return nil
}

// RefreshTokens обновляет пару токенов на основе валидного refresh токена
func (s *Service) RefreshTokens(refreshTokenString string) (*tokens.TokenPair, error) {
	if refreshTokenString == "" {
		return nil, fmt.Errorf("refresh токен отсутствует")
	}

	// Получаем refresh токен из БД
	refreshToken, err := s.tokenService.GetRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("недействительный refresh токен")
	}

	// Получаем пользователя по ID из refresh токена
	user, err := s.repo.GetUserByID(refreshToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден")
	}

	// Обновляем пару токенов
	tokenPair, err := s.tokenService.RefreshTokenPair(refreshTokenString, user.ID, user.Email)
	if err != nil {
		return nil, err
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
