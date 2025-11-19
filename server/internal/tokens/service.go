package tokens

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	repo       Repository
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewService(repo Repository, jwtSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		repo:       repo,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *Service) GenerateTokenPair(userID int, email string) (*TokenPair, error) {
	accessToken, err := s.generateJWT(userID, email, "access", s.accessTTL)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать access токен: %w", err)
	}

	refreshTokenString, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("не удалось создать refresh токен: %w", err)
	}

	refreshToken := &RefreshToken{
		Token:  refreshTokenString,
		UserID: userID,
	}

	if err := s.repo.SaveRefreshToken(refreshToken); err != nil {
		return nil, fmt.Errorf("не удалось сохранить refresh токен: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	}, nil
}

func (s *Service) generateJWT(userID int, email, tokenType string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"type":    tokenType,
		"exp":     time.Now().Add(ttl).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга токена: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenType, ok := claims["type"].(string)
		if !ok || tokenType != "access" {
			return nil, fmt.Errorf("Неверный тип токена")
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			return nil, fmt.Errorf("отсутствует user_id в токене")
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, fmt.Errorf("отсутствует email в токене")
		}

		return &Claims{
			UserID: int(userID),
			Email:  email,
			Type:   tokenType,
		}, nil
	}

	return nil, fmt.Errorf("Недействительный токен")
}

func (s *Service) GetRefreshToken(tokenString string) (*RefreshToken, error) {
	return s.repo.GetRefreshToken(tokenString)
}

func (s *Service) RefreshTokenPair(refreshTokenString string, userID int, email string) (*TokenPair, error) {
	refreshToken, err := s.repo.GetRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("Недействительный refresh токен")
	}

	if refreshToken.UserID != userID {
		return nil, fmt.Errorf("Недействительный refresh токен")
	}

	if err := s.repo.DeleteRefreshToken(refreshTokenString); err != nil {
		return nil, fmt.Errorf("не удалось удалить старый refresh токен: %w", err)
	}

	return s.GenerateTokenPair(userID, email)
}

func (s *Service) Logout(refreshTokenString string) error {
	if refreshTokenString == "" {
		return fmt.Errorf("Refresh токен отсутствует")
	}

	_, err := s.repo.GetRefreshToken(refreshTokenString)
	if err != nil {
		return fmt.Errorf("Недействительный refresh токен")
	}

	if err := s.repo.DeleteRefreshToken(refreshTokenString); err != nil {
		return fmt.Errorf("не удалось удалить refresh токен: %w", err)
	}

	return nil
}

func (s *Service) LogoutAll(userID int) error {
	if err := s.repo.DeleteUserRefreshTokens(userID); err != nil {
		return fmt.Errorf("не удалось удалить refresh токены пользователя: %w", err)
	}

	return nil
}
