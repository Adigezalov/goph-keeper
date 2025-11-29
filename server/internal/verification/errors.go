package verification

import "errors"

var (
	ErrInvalidCode     = errors.New("verification.invalid_code")
	ErrCodeExpired     = errors.New("verification.code_expired")
	ErrEmailRequired   = errors.New("verification.email_required")
	ErrCodeRequired    = errors.New("verification.code_required")
	ErrRequestRequired = errors.New("verification.request_required")
	ErrCodeNotFound    = errors.New("verification.code_not_found")
	ErrCodeAlreadyUsed = errors.New("verification.code_already_used")
)
