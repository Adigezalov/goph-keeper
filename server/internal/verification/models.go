package verification

import "time"

type VerificationCode struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	Code      string    `db:"code"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
	Used      bool      `db:"used"`
}

type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type ResendCodeRequest struct {
	Email string `json:"email"`
}
