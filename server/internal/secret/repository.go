package secret

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// Repository интерфейс для работы с секретами в БД
type Repository interface {
	CreateSecret(secret *Secret) error
	GetSecretByID(id string, userID int) (*Secret, error)
	GetSecretsByUserID(userID int) ([]*Secret, error)
	GetSecretsModifiedSince(userID int, since time.Time) ([]*Secret, error)
	UpdateSecret(secret *Secret) error
	SoftDeleteSecret(id string, userID int) error
}

// DatabaseRepository реализация Repository для PostgreSQL
type DatabaseRepository struct {
	db *sql.DB
}

// NewDatabaseRepository создает новый экземпляр DatabaseRepository
func NewDatabaseRepository(db *sql.DB) *DatabaseRepository {
	return &DatabaseRepository{db: db}
}

// CreateSecret создает новый секрет в БД
func (r *DatabaseRepository) CreateSecret(secret *Secret) error {
	// Сериализуем metadata в JSONB
	var metadataJSON []byte
	var err error
	if secret.Metadata != nil {
		metadataJSON, err = json.Marshal(secret.Metadata)
		if err != nil {
			return WrapError(err, "не удалось сериализовать metadata")
		}
	}

	query := `
		INSERT INTO secrets (user_id, login, password, metadata, binary_data, version)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(
		query,
		secret.UserID,
		secret.Login,
		secret.Password,
		metadataJSON,
		secret.BinaryData,
		secret.Version,
	).Scan(&secret.ID, &secret.CreatedAt, &secret.UpdatedAt)

	if err != nil {
		return WrapError(err, "не удалось создать секрет")
	}

	return nil
}

// GetSecretByID получает секрет по ID (только если принадлежит пользователю)
func (r *DatabaseRepository) GetSecretByID(id string, userID int) (*Secret, error) {
	var secret Secret
	var metadataJSON []byte

	query := `
		SELECT id, user_id, login, password, metadata, binary_data, version,
		       created_at, updated_at, deleted_at
		FROM secrets
		WHERE id = $1 AND user_id = $2
	`

	err := r.db.QueryRow(query, id, userID).Scan(
		&secret.ID,
		&secret.UserID,
		&secret.Login,
		&secret.Password,
		&metadataJSON,
		&secret.BinaryData,
		&secret.Version,
		&secret.CreatedAt,
		&secret.UpdatedAt,
		&secret.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSecretNotFound
		}
		return nil, WrapError(err, "не удалось получить секрет")
	}

	// Десериализуем metadata
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &secret.Metadata); err != nil {
			return nil, WrapError(err, "не удалось десериализовать metadata")
		}
	}

	return &secret, nil
}

// GetSecretsByUserID получает все активные секреты пользователя (без удаленных)
func (r *DatabaseRepository) GetSecretsByUserID(userID int) ([]*Secret, error) {
	query := `
		SELECT id, user_id, login, password, metadata, binary_data, version,
		       created_at, updated_at, deleted_at
		FROM secrets
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, WrapError(err, "не удалось получить секреты")
	}
	defer rows.Close()

	var secrets []*Secret
	for rows.Next() {
		var secret Secret
		var metadataJSON []byte

		err := rows.Scan(
			&secret.ID,
			&secret.UserID,
			&secret.Login,
			&secret.Password,
			&metadataJSON,
			&secret.BinaryData,
			&secret.Version,
			&secret.CreatedAt,
			&secret.UpdatedAt,
			&secret.DeletedAt,
		)

		if err != nil {
			return nil, WrapError(err, "не удалось прочитать секрет")
		}

		// Десериализуем metadata
		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &secret.Metadata); err != nil {
				return nil, WrapError(err, "не удалось десериализовать metadata")
			}
		}

		secrets = append(secrets, &secret)
	}

	if err = rows.Err(); err != nil {
		return nil, WrapError(err, "ошибка при чтении секретов")
	}

	return secrets, nil
}

// GetSecretsModifiedSince получает все секреты пользователя, измененные после указанного времени
// Это ключевой метод для инкрементальной синхронизации
// Возвращает созданные, обновленные и удаленные (deleted_at IS NOT NULL) секреты
func (r *DatabaseRepository) GetSecretsModifiedSince(userID int, since time.Time) ([]*Secret, error) {
	query := `
		SELECT id, user_id, login, password, metadata, binary_data, version,
		       created_at, updated_at, deleted_at
		FROM secrets
		WHERE user_id = $1 
		  AND (
		      created_at >= $2 
		      OR updated_at >= $2 
		      OR (deleted_at IS NOT NULL AND deleted_at >= $2)
		  )
		ORDER BY updated_at ASC
	`

	rows, err := r.db.Query(query, userID, since)
	if err != nil {
		return nil, WrapError(err, "не удалось получить измененные секреты")
	}
	defer rows.Close()

	var secrets []*Secret
	for rows.Next() {
		var secret Secret
		var metadataJSON []byte

		err := rows.Scan(
			&secret.ID,
			&secret.UserID,
			&secret.Login,
			&secret.Password,
			&metadataJSON,
			&secret.BinaryData,
			&secret.Version,
			&secret.CreatedAt,
			&secret.UpdatedAt,
			&secret.DeletedAt,
		)

		if err != nil {
			return nil, WrapError(err, "не удалось прочитать секрет")
		}

		// Десериализуем metadata
		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &secret.Metadata); err != nil {
				return nil, WrapError(err, "не удалось десериализовать metadata")
			}
		}

		secrets = append(secrets, &secret)
	}

	if err = rows.Err(); err != nil {
		return nil, WrapError(err, "ошибка при чтении измененных секретов")
	}

	return secrets, nil
}

// UpdateSecret обновляет секрет в БД
// Проверяет version для оптимистической блокировки (предотвращение конфликтов)
func (r *DatabaseRepository) UpdateSecret(secret *Secret) error {
	// Сериализуем metadata в JSONB
	var metadataJSON []byte
	var err error
	if secret.Metadata != nil {
		metadataJSON, err = json.Marshal(secret.Metadata)
		if err != nil {
			return WrapError(err, "не удалось сериализовать metadata")
		}
	}

	query := `
		UPDATE secrets
		SET login = $1, 
		    password = $2, 
		    metadata = $3, 
		    binary_data = $4, 
		    version = $5
		WHERE id = $6 
		  AND user_id = $7 
		  AND version = $8 
		  AND deleted_at IS NULL
		RETURNING updated_at
	`

	err = r.db.QueryRow(
		query,
		secret.Login,
		secret.Password,
		metadataJSON,
		secret.BinaryData,
		secret.Version+1, // Инкрементируем версию
		secret.ID,
		secret.UserID,
		secret.Version, // Проверяем текущую версию (оптимистическая блокировка)
	).Scan(&secret.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrVersionConflict
		}
		return WrapError(err, "не удалось обновить секрет")
	}

	// Обновляем версию в объекте после успешного обновления
	secret.Version++

	return nil
}

// SoftDeleteSecret выполняет мягкое удаление секрета (устанавливает deleted_at)
func (r *DatabaseRepository) SoftDeleteSecret(id string, userID int) error {
	query := `
		UPDATE secrets
		SET deleted_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := r.db.QueryRow(query, id, userID).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSecretNotFound
		}
		return WrapError(err, "не удалось удалить секрет")
	}

	return nil
}
