package tokens

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Service содержит бизнес-логику для работы с токенами
type Service struct {
	repo       Repository
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewService создает новый экземпляр Service
func NewService(repo Repository, jwtSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		repo:       repo,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// GenerateTokenPair создает пару токенов для пользователя
func (s *Service) GenerateTokenPair(userID int, email string) (*TokenPair, error) {
	// Генерируем access токен
	accessToken, err := s.generateJWT(userID, email, "access", s.accessTTL)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать access токен: %w", err)
	}

	// Генерируем refresh токен
	refreshTokenString, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("не удалось создать refresh токен: %w", err)
	}

	// Сохраняем refresh токен в БД
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

// generateJWT создает JWT токен
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

// generateRefreshToken создает случайный refresh токен
func (s *Service) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateAccessToken проверяет access токен
func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["type"] != "access" {
			return nil, fmt.Errorf("Неверный тип токена")
		}

		return &Claims{
			UserID: int(claims["user_id"].(float64)),
			Email:  claims["email"].(string),
			Type:   claims["type"].(string),
		}, nil
	}

	return nil, fmt.Errorf("Недействительный токен")
}

// GetRefreshToken получает refresh токен из БД
func (s *Service) GetRefreshToken(tokenString string) (*RefreshToken, error) {
	return s.repo.GetRefreshToken(tokenString)
}

// RefreshTokenPair обновляет пару токенов на основе валидного refresh токена
// Принимает refreshTokenString для проверки и userID с email для генерации новой пары
func (s *Service) RefreshTokenPair(refreshTokenString string, userID int, email string) (*TokenPair, error) {
	// Получаем refresh токен из БД
	refreshToken, err := s.repo.GetRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("Недействительный refresh токен")
	}

	// Проверяем, что токен принадлежит указанному пользователю
	if refreshToken.UserID != userID {
		return nil, fmt.Errorf("Недействительный refresh токен")
	}

	// Удаляем старый refresh токен
	if err := s.repo.DeleteRefreshToken(refreshTokenString); err != nil {
		return nil, fmt.Errorf("не удалось удалить старый refresh токен: %w", err)
	}

	// Генерируем новую пару токенов
	return s.GenerateTokenPair(userID, email)
}

// Logout удаляет конкретный refresh токен
func (s *Service) Logout(refreshTokenString string) error {
	if refreshTokenString == "" {
		return fmt.Errorf("Refresh токен отсутствует")
	}

	// Проверяем, что токен существует
	_, err := s.repo.GetRefreshToken(refreshTokenString)
	if err != nil {
		return fmt.Errorf("Недействительный refresh токен")
	}

	// Удаляем токен
	if err := s.repo.DeleteRefreshToken(refreshTokenString); err != nil {
		return fmt.Errorf("не удалось удалить refresh токен: %w", err)
	}

	return nil
}

// LogoutAll удаляет все refresh токены пользователя
func (s *Service) LogoutAll(userID int) error {
	if err := s.repo.DeleteUserRefreshTokens(userID); err != nil {
		return fmt.Errorf("не удалось удалить refresh токены пользователя: %w", err)
	}

	return nil
}
