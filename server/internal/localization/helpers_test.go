package localization

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalizedError(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		messageID    string
		templateData map[string]interface{}
		language     string
	}{
		{
			name:       "Simple error",
			statusCode: http.StatusBadRequest,
			messageID:  "common.invalid_request_format",
			language:   "ru",
		},
		{
			name:       "Error with template data",
			statusCode: http.StatusNotFound,
			messageID:  "secret.not_found",
			templateData: map[string]interface{}{
				"id": "123",
			},
			language: "ru",
		},
		{
			name:       "Internal server error",
			statusCode: http.StatusInternalServerError,
			messageID:  "common.internal_error",
			language:   "ru",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := context.WithValue(req.Context(), LanguageContextKey, tt.language)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			LocalizedError(w, req, tt.statusCode, tt.messageID, tt.templateData)

			assert.Equal(t, tt.statusCode, w.Code)
			assert.NotEmpty(t, w.Body.String())
		})
	}
}

func TestLocalizedError_WithoutLanguageInContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	LocalizedError(w, req, http.StatusBadRequest, "common.invalid_request_format", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotEmpty(t, w.Body.String())
}

func TestLocalizedErrorWithFallback(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		messageID    string
		fallback     string
		templateData map[string]interface{}
		language     string
	}{
		{
			name:       "With valid message ID",
			statusCode: http.StatusBadRequest,
			messageID:  "common.invalid_request_format",
			fallback:   "Fallback message",
			language:   "ru",
		},
		{
			name:       "With invalid message ID (uses fallback)",
			statusCode: http.StatusBadRequest,
			messageID:  "nonexistent.message.id",
			fallback:   "This is the fallback",
			language:   "ru",
		},
		{
			name:       "With template data",
			statusCode: http.StatusNotFound,
			messageID:  "secret.not_found",
			fallback:   "Secret not found fallback",
			templateData: map[string]interface{}{
				"id": "456",
			},
			language: "ru",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := context.WithValue(req.Context(), LanguageContextKey, tt.language)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			LocalizedErrorWithFallback(w, req, tt.statusCode, tt.messageID, tt.fallback, tt.templateData)

			assert.Equal(t, tt.statusCode, w.Code)
			assert.NotEmpty(t, w.Body.String())
		})
	}
}

func TestLocalizedErrorWithFallback_UsesFallback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), LanguageContextKey, "ru")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	messageID := "completely.nonexistent.message.id.that.does.not.exist"
	fallback := "Expected fallback message"

	LocalizedErrorWithFallback(w, req, http.StatusBadRequest, messageID, fallback, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	// The response should contain the fallback
	assert.Contains(t, w.Body.String(), fallback)
}

func TestLocalizedError_VariousStatusCodes(t *testing.T) {
	statusCodes := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusConflict,
		http.StatusInternalServerError,
	}

	for _, code := range statusCodes {
		t.Run(http.StatusText(code), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := context.WithValue(req.Context(), LanguageContextKey, "ru")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			LocalizedError(w, req, code, "common.internal_error", nil)

			assert.Equal(t, code, w.Code)
		})
	}
}
