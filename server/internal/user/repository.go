package user

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Repository interface {
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	VerifyUserEmail(userID int) error
}

type DatabaseRepository struct {
	db *sql.DB
}

func NewDatabaseRepository(db *sql.DB) *DatabaseRepository {
	return &DatabaseRepository{db: db}
}

func (r *DatabaseRepository) CreateUser(user *User) error {
	query := `
		INSERT INTO users (email, password_hash, email_verified) 
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query, user.Email, user.PasswordHash, user.EmailVerified).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrUserAlreadyExists
		}
		return WrapError(err, "не удалось создать пользователя")
	}

	return nil
}

func (r *DatabaseRepository) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, email, password_hash, email_verified, created_at, updated_at 
		FROM users 
		WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, WrapError(err, "не удалось получить пользователя")
	}

	return user, nil
}

func (r *DatabaseRepository) GetUserByID(id int) (*User, error) {
	user := &User{}
	query := `
		SELECT id, email, password_hash, email_verified, created_at, updated_at 
		FROM users 
		WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, WrapError(err, "не удалось получить пользователя")
	}

	return user, nil
}

func (r *DatabaseRepository) VerifyUserEmail(userID int) error {
	query := `UPDATE users SET email_verified = true WHERE id = $1`
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return WrapError(err, "не удалось верифицировать email")
	}
	return nil
}
