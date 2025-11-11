package user

import (
	"errors"
	"fmt"
)

// Определяем типизированные ошибки для пользовательского домена
var (
	// ErrUserAlreadyExists ошибка при попытке создать пользователя с существующим email
	ErrUserAlreadyExists = errors.New("Пользователь уже существует")

	// ErrUserNotFound ошибка когда пользователь не найден
	ErrUserNotFound = errors.New("Пользователь не найден")

	// ErrInvalidEmail ошибка при неверном формате email
	ErrInvalidEmail = errors.New("Неверный формат email")

	// ErrEmailRequired ошибка когда email не указан
	ErrEmailRequired = errors.New("Email обязателен")

	// ErrPasswordRequired ошибка когда пароль не указан
	ErrPasswordRequired = errors.New("Пароль обязателен")

	// ErrPasswordTooShort ошибка когда пароль слишком короткий
	ErrPasswordTooShort = errors.New("Пароль должен содержать минимум 6 символов")

	// ErrInvalidCredentials ошибка при неверной паре email/пароль
	ErrInvalidCredentials = errors.New("Неверная пара email/пароль")

	// ErrRefreshTokenMissing ошибка когда refresh токен отсутствует
	ErrRefreshTokenMissing = errors.New("Refresh токен отсутствует")

	// ErrInvalidRefreshToken ошибка когда refresh токен недействителен
	ErrInvalidRefreshToken = errors.New("Недействительный refresh токен")

	// ErrRequestRequired ошибка когда запрос не указан
	ErrRequestRequired = errors.New("Запрос обязателен")
)

// HTTPError представляет ошибку с HTTP статус кодом
type HTTPError struct {
	Err        error
	StatusCode int
	Message    string
}

// Error реализует интерфейс error
func (e *HTTPError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

// Unwrap возвращает исходную ошибку
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError создает новую HTTP ошибку
func NewHTTPError(err error, statusCode int, message string) *HTTPError {
	return &HTTPError{
		Err:        err,
		StatusCode: statusCode,
		Message:    message,
	}
}

// WrapError оборачивает ошибку с дополнительным контекстом
func WrapError(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}
