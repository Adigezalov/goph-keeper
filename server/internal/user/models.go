package user

import (
	"time"
)

// User представляет пользователя системы
type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// RegisterRequest представляет запрос на регистрацию
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest представляет запрос на аутентификацию
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
