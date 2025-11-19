package secret

import (
	"errors"
	"fmt"
)

var (
	ErrSecretNotFound  = errors.New("secret.not_found")
	ErrVersionConflict = errors.New("secret.version_conflict")
)

var (
	ErrRequestRequired  = errors.New("secret.request_required")
	ErrLoginRequired    = errors.New("secret.login_required")
	ErrPasswordRequired = errors.New("secret.password_required")
)

func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}
