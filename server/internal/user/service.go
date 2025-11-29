package user

import (
	"errors"
	"net/mail"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/tokens"
	"github.com/Adigezalov/goph-keeper/internal/verification"
	"golang.org/x/crypto/bcrypt"
)

type TokenService interface {
	GenerateTokenPair(userID int, email string) (*tokens.TokenPair, error)
	GetRefreshToken(tokenString string) (*tokens.RefreshToken, error)
	RefreshTokenPair(refreshTokenString string, userID int, email string) (*tokens.TokenPair, error)
	Logout(refreshTokenString string) error
	LogoutAll(userID int) error
}

type EmailService interface {
	GenerateVerificationCode() string
	SendEmail(toEmail, code string)
}

type VerificationRepository interface {
	CreateVerificationCode(code *verification.VerificationCode) error
	GetActiveVerificationCode(userID int, code string) (*verification.VerificationCode, error)
	MarkCodeAsUsed(id int) error
	DeleteUserCodes(userID int) error
}

type Service struct {
	repo                Repository
	tokenService        TokenService
	emailService        EmailService
	verificationRepo    VerificationRepository
	verificationCodeTTL time.Duration
}

func NewService(
	repo Repository,
	tokenService TokenService,
	emailService EmailService,
	verificationRepo VerificationRepository,
	verificationCodeTTL time.Duration,
) *Service {
	return &Service{
		repo:                repo,
		tokenService:        tokenService,
		emailService:        emailService,
		verificationRepo:    verificationRepo,
		verificationCodeTTL: verificationCodeTTL,
	}
}

func (s *Service) RegisterUser(req *RegisterRequest) error {
	if err := s.validateRegisterRequest(req); err != nil {
		return err
	}

	existingUser, err := s.repo.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		return ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return WrapError(err, "не удалось захешировать пароль")
	}

	user := &User{
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		EmailVerified: false,
	}

	if err := s.repo.CreateUser(user); err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return ErrUserAlreadyExists
		}
		return WrapError(err, "не удалось создать пользователя")
	}

	// Генерируем и отправляем код верификации
	code := s.emailService.GenerateVerificationCode()
	verificationCode := &verification.VerificationCode{
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: time.Now().Add(s.verificationCodeTTL),
	}

	if err := s.verificationRepo.CreateVerificationCode(verificationCode); err != nil {
		return WrapError(err, "не удалось создать код верификации")
	}

	// Асинхронная отправка через worker
	s.emailService.SendEmail(user.Email, code)

	return nil
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

func (s *Service) VerifyEmail(req *verification.VerifyEmailRequest) (*tokens.TokenPair, error) {
	if req == nil {
		return nil, verification.ErrRequestRequired
	}
	if req.Email == "" {
		return nil, verification.ErrEmailRequired
	}
	if req.Code == "" {
		return nil, verification.ErrCodeRequired
	}

	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, verification.ErrInvalidCode
	}

	if user.EmailVerified {
		// Если email уже верифицирован, просто возвращаем токены
		tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Email)
		if err != nil {
			return nil, WrapError(err, "не удалось создать токены")
		}
		return tokenPair, nil
	}

	verificationCode, err := s.verificationRepo.GetActiveVerificationCode(user.ID, req.Code)
	if err != nil {
		if errors.Is(err, verification.ErrCodeNotFound) {
			return nil, verification.ErrInvalidCode
		}
		return nil, err
	}

	if verificationCode.ExpiresAt.Before(time.Now()) {
		return nil, verification.ErrCodeExpired
	}

	if verificationCode.Used {
		return nil, verification.ErrCodeAlreadyUsed
	}

	// Помечаем код как использованный
	if err := s.verificationRepo.MarkCodeAsUsed(verificationCode.ID); err != nil {
		return nil, WrapError(err, "не удалось пометить код как использованный")
	}

	// Верифицируем email пользователя
	if err := s.repo.VerifyUserEmail(user.ID); err != nil {
		return nil, WrapError(err, "не удалось верифицировать email")
	}

	// Генерируем токены
	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, WrapError(err, "не удалось создать токены")
	}

	return tokenPair, nil
}

func (s *Service) ResendVerificationCode(req *verification.ResendCodeRequest) error {
	if req == nil {
		return verification.ErrRequestRequired
	}
	if req.Email == "" {
		return verification.ErrEmailRequired
	}

	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		// Не раскрываем, существует ли пользователь
		return nil
	}

	if user.EmailVerified {
		// Email уже верифицирован, ничего не делаем
		return nil
	}

	// Удаляем старые коды пользователя
	if err := s.verificationRepo.DeleteUserCodes(user.ID); err != nil {
		return WrapError(err, "не удалось удалить старые коды")
	}

	// Генерируем новый код
	code := s.emailService.GenerateVerificationCode()
	verificationCode := &verification.VerificationCode{
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: time.Now().Add(s.verificationCodeTTL),
	}

	if err := s.verificationRepo.CreateVerificationCode(verificationCode); err != nil {
		return WrapError(err, "не удалось создать код верификации")
	}

	// Асинхронная отправка через worker
	s.emailService.SendEmail(user.Email, code)

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

	if !user.EmailVerified {
		return nil, ErrEmailNotVerified
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
