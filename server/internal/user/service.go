package user

import (
	"errors"
	"net/mail"

	"github.com/Adigezalov/goph-keeper/internal/tokens"
	"golang.org/x/crypto/bcrypt"
)

type TokenService interface {
	GenerateTokenPair(userID int, email string) (*tokens.TokenPair, error)
	GetRefreshToken(tokenString string) (*tokens.RefreshToken, error)
	RefreshTokenPair(refreshTokenString string, userID int, email string) (*tokens.TokenPair, error)
	Logout(refreshTokenString string) error
	LogoutAll(userID int) error
}

type Service struct {
	repo         Repository
	tokenService TokenService
}

func NewService(repo Repository, tokenService TokenService) *Service {
	return &Service{
		repo:         repo,
		tokenService: tokenService,
	}
}

func (s *Service) RegisterUser(req *RegisterRequest) (*tokens.TokenPair, error) {
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	existingUser, err := s.repo.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, WrapError(err, "не удалось захешировать пароль")
	}

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

	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, WrapError(err, "не удалось создать токены")
	}

	return tokenPair, nil
}

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

func (s *Service) LoginUser(req *LoginRequest) (*tokens.TokenPair, error) {
	if err := s.validateLoginRequest(req); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, WrapError(err, "не удалось создать токены")
	}

	return tokenPair, nil
}

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

func (s *Service) RefreshTokens(refreshTokenString string) (*tokens.TokenPair, error) {
	if refreshTokenString == "" {
		return nil, ErrRefreshTokenMissing
	}

	refreshToken, err := s.tokenService.GetRefreshToken(refreshTokenString)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.repo.GetUserByID(refreshToken.UserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, WrapError(err, "не удалось получить пользователя")
	}

	tokenPair, err := s.tokenService.RefreshTokenPair(refreshTokenString, user.ID, user.Email)
	if err != nil {
		return nil, WrapError(err, "не удалось обновить токены")
	}

	return tokenPair, nil
}

func (s *Service) Logout(refreshTokenString string) error {
	return s.tokenService.Logout(refreshTokenString)
}

func (s *Service) LogoutAll(userID int) error {
	return s.tokenService.LogoutAll(userID)
}
