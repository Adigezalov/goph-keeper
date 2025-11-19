package tokens

import (
	"time"
)

type RefreshToken struct {
	ID        int       `json:"id" db:"id"`
	Token     string    `json:"token" db:"token"`
	UserID    int       `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Type   string `json:"type"`
}
