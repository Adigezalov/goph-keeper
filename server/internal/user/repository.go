package user

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

// Repository интерфейс для работы с пользователями в БД
type Repository interface {
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
}

// DatabaseRepository реализация Repository для Postgres
type DatabaseRepository struct {
	db *sql.DB
}

// NewDatabaseRepository создает новый экземпляр DatabaseRepository
func NewDatabaseRepository(db *sql.DB) *DatabaseRepository {
	return &DatabaseRepository{db: db}
}

// CreateUser создает нового пользователя
func (r *DatabaseRepository) CreateUser(user *User) error {
	query := `
		INSERT INTO users (email, password_hash) 
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query, user.Email, user.PasswordHash).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		// Проверяем, является ли ошибка нарушением уникального ограничения
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return ErrUserAlreadyExists
		}
		return WrapError(err, "не удалось создать пользователя")
	}

	return nil
}

// GetUserByEmail получает пользователя по email
func (r *DatabaseRepository) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, email, password_hash, created_at, updated_at 
		FROM users 
		WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, WrapError(err, "не удалось получить пользователя")
	}

	return user, nil
}

// GetUserByID получает пользователя по ID
func (r *DatabaseRepository) GetUserByID(id int) (*User, error) {
	user := &User{}
	query := `
		SELECT id, email, password_hash, created_at, updated_at 
		FROM users 
		WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, WrapError(err, "не удалось получить пользователя")
	}

	return user, nil
}
