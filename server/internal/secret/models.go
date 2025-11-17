package secret

import (
	"database/sql"
	"time"
)

// Secret представляет секретные данные пользователя
// Все чувствительные поля (login, password, binaryData) приходят с клиента уже зашифрованными
type Secret struct {
	ID         string                 `json:"id" db:"id"` // UUID
	UserID     int                    `json:"user_id" db:"user_id"`
	Login      string                 `json:"login" db:"login"`                       // Зашифрованный логин (base64)
	Password   string                 `json:"password" db:"password"`                 // Зашифрованный пароль (base64)
	Metadata   map[string]interface{} `json:"metadata,omitempty" db:"metadata"`       // Незашифрованные метаданные (fileName, fileExtension, app и т.д.)
	BinaryData []byte                 `json:"binary_data,omitempty" db:"binary_data"` // Зашифрованные бинарные данные (файлы)
	Version    int                    `json:"version" db:"version"`                   // Версия для разрешения конфликтов
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt  sql.NullTime           `json:"deleted_at,omitempty" db:"deleted_at"` // Soft delete
}

// CreateSecretRequest представляет запрос на создание секрета
type CreateSecretRequest struct {
	Login      string                 `json:"login"`                 // Зашифрованный
	Password   string                 `json:"password"`              // Зашифрованный
	Metadata   map[string]interface{} `json:"metadata,omitempty"`    // Незашифрованный
	BinaryData []byte                 `json:"binary_data,omitempty"` // Зашифрованный
}

// UpdateSecretRequest представляет запрос на обновление секрета
type UpdateSecretRequest struct {
	Login      string                 `json:"login"`                 // Зашифрованный
	Password   string                 `json:"password"`              // Зашифрованный
	Metadata   map[string]interface{} `json:"metadata,omitempty"`    // Незашифрованный
	BinaryData []byte                 `json:"binary_data,omitempty"` // Зашифрованный
	Version    int                    `json:"version"`               // Для разрешения конфликтов
}

// SecretResponse представляет ответ с секретом (для клиента)
type SecretResponse struct {
	ID             string                 `json:"id"` // UUID
	Login          string                 `json:"login"`
	Password       string                 `json:"password"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	BinaryData     []byte                 `json:"binary_data,omitempty"`
	BinaryDataSize *int64                 `json:"binary_data_size,omitempty"` // Размер binary_data (если данные не включены)
	Version        int                    `json:"version"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	DeletedAt      *time.Time             `json:"deleted_at,omitempty"`
}

const (
	// MinSizeForChunks - минимальный размер для использования chunked transfer (1 MB)
	MinSizeForChunks = 1 * 1024 * 1024
)

// ToResponse конвертирует Secret в SecretResponse (для создания/обновления)
func (s *Secret) ToResponse() SecretResponse {
	resp := SecretResponse{
		ID:         s.ID,
		Login:      s.Login,
		Password:   s.Password,
		Metadata:   s.Metadata,
		BinaryData: s.BinaryData,
		Version:    s.Version,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}

	if s.DeletedAt.Valid {
		resp.DeletedAt = &s.DeletedAt.Time
	}

	return resp
}

// ToResponseForSync конвертирует Secret в SecretResponse для синхронизации
// Для больших файлов (>1MB) не отправляет binary_data, а только размер
func (s *Secret) ToResponseForSync() SecretResponse {
	resp := SecretResponse{
		ID:        s.ID,
		Login:     s.Login,
		Password:  s.Password,
		Metadata:  s.Metadata,
		Version:   s.Version,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}

	if s.DeletedAt.Valid {
		resp.DeletedAt = &s.DeletedAt.Time
	}

	// Для больших файлов отправляем только размер, клиент скачает чанками
	if len(s.BinaryData) > MinSizeForChunks {
		size := int64(len(s.BinaryData))
		resp.BinaryDataSize = &size
	} else {
		// Для маленьких файлов отправляем данные целиком
		resp.BinaryData = s.BinaryData
	}

	return resp
}
