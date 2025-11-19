package secret

import (
	"database/sql"
	"time"
)

type Secret struct {
	ID         string                 `json:"id" db:"id"`
	UserID     int                    `json:"user_id" db:"user_id"`
	Login      string                 `json:"login" db:"login"`
	Password   string                 `json:"password" db:"password"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	BinaryData []byte                 `json:"binary_data,omitempty" db:"binary_data"`
	Version    int                    `json:"version" db:"version"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt  sql.NullTime           `json:"deleted_at,omitempty" db:"deleted_at"`
}

type CreateSecretRequest struct {
	Login      string                 `json:"login"`
	Password   string                 `json:"password"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	BinaryData []byte                 `json:"binary_data,omitempty"`
}

type UpdateSecretRequest struct {
	Login      string                 `json:"login"`
	Password   string                 `json:"password"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	BinaryData []byte                 `json:"binary_data,omitempty"`
	Version    int                    `json:"version"`
}

type SecretResponse struct {
	ID             string                 `json:"id"`
	Login          string                 `json:"login"`
	Password       string                 `json:"password"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	BinaryData     []byte                 `json:"binary_data,omitempty"`
	BinaryDataSize *int64                 `json:"binary_data_size,omitempty"`
	Version        int                    `json:"version"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	DeletedAt      *time.Time             `json:"deleted_at,omitempty"`
}

const (
	MinSizeForChunks = 1 * 1024 * 1024
)

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

	if len(s.BinaryData) > MinSizeForChunks {
		size := int64(len(s.BinaryData))
		resp.BinaryDataSize = &size
	} else {
		resp.BinaryData = s.BinaryData
	}

	return resp
}
