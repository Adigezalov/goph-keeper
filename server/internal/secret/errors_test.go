package secret

import (
	"errors"
	"testing"
)

func TestWrapError_WithError(t *testing.T) {
	originalErr := errors.New("original error")
	message := "wrapped message"

	wrappedErr := WrapError(originalErr, message)

	if wrappedErr == nil {
		t.Fatal("Expected non-nil error")
	}

	expectedMsg := "wrapped message: original error"
	if wrappedErr.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, wrappedErr.Error())
	}

	if !errors.Is(wrappedErr, originalErr) {
		t.Error("Wrapped error should unwrap to original error")
	}
}

func TestWrapError_WithNilError(t *testing.T) {
	message := "some message"

	wrappedErr := WrapError(nil, message)

	if wrappedErr != nil {
		t.Errorf("Expected nil error, got %v", wrappedErr)
	}
}

func TestWrapError_MultipleLevels(t *testing.T) {
	baseErr := errors.New("base error")
	firstWrap := WrapError(baseErr, "first wrap")
	secondWrap := WrapError(firstWrap, "second wrap")

	if secondWrap == nil {
		t.Fatal("Expected non-nil error")
	}

	expectedMsg := "second wrap: first wrap: base error"
	if secondWrap.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, secondWrap.Error())
	}

	if !errors.Is(secondWrap, baseErr) {
		t.Error("Multiple wrapped error should unwrap to base error")
	}
}

func TestErrorVariables(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrSecretNotFound",
			err:      ErrSecretNotFound,
			expected: "secret.not_found",
		},
		{
			name:     "ErrVersionConflict",
			err:      ErrVersionConflict,
			expected: "secret.version_conflict",
		},
		{
			name:     "ErrRequestRequired",
			err:      ErrRequestRequired,
			expected: "secret.request_required",
		},
		{
			name:     "ErrLoginRequired",
			err:      ErrLoginRequired,
			expected: "secret.login_required",
		},
		{
			name:     "ErrPasswordRequired",
			err:      ErrPasswordRequired,
			expected: "secret.password_required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Expected error message '%s', got '%s'", tt.expected, tt.err.Error())
			}
		})
	}
}
