package secret

import (
	"errors"
	"fmt"
)

// Ошибки для работы с секретами
var (
	ErrSecretNotFound  = errors.New("секрет не найден")
	ErrVersionConflict = errors.New("конфликт версий: секрет был изменен другим устройством")
)

// Ошибки валидации
var (
	ErrRequestRequired  = errors.New("запрос обязателен")
	ErrLoginRequired    = errors.New("логин обязателен")
	ErrPasswordRequired = errors.New("пароль обязателен")
)

// WrapError оборачивает ошибку с дополнительным контекстом
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}
