package user

import (
	"errors"
	"fmt"
)

var (
	ErrUserAlreadyExists = errors.New("user.already_exists")

	ErrUserNotFound = errors.New("user.not_found")

	ErrInvalidEmail = errors.New("user.invalid_email")

	ErrEmailRequired = errors.New("user.email_required")

	ErrPasswordRequired = errors.New("user.password_required")

	ErrPasswordTooShort = errors.New("user.password_too_short")

	ErrInvalidCredentials = errors.New("user.invalid_credentials")

	ErrRefreshTokenMissing = errors.New("user.refresh_token_missing")

	ErrInvalidRefreshToken = errors.New("user.invalid_refresh_token")

	ErrRequestRequired = errors.New("user.request_required")

	ErrEmailNotVerified = errors.New("user.email_not_verified")
)

type HTTPError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func NewHTTPError(err error, statusCode int, message string) *HTTPError {
	return &HTTPError{
		Err:        err,
		StatusCode: statusCode,
		Message:    message,
	}
}

func WrapError(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}
