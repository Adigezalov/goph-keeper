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
	ID         string                 `json:"id"` // UUID
	Login      string                 `json:"login"`
	Password   string                 `json:"password"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	BinaryData []byte                 `json:"binary_data,omitempty"`
	Version    int                    `json:"version"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	DeletedAt  *time.Time             `json:"deleted_at,omitempty"`
}

// ToResponse конвертирует Secret в SecretResponse
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
