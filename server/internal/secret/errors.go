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

// WrapError оборачивает ошибку с дополнительным контекстом
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}
