package localization

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLanguageMiddleware(t *testing.T) {
	handler := LanguageMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := GetLanguageFromContext(r.Context())
		w.Write([]byte(lang))
	}))

	tests := []struct {
		name           string
		acceptLanguage string
		expectedLang   string
	}{
		{
			name:           "Russian language",
			acceptLanguage: "ru-RU,ru;q=0.9",
			expectedLang:   "ru",
		},
		{
			name:           "No Accept-Language header",
			acceptLanguage: "",
			expectedLang:   "ru",
		},
		{
			name:           "English language (defaults to Russian)",
			acceptLanguage: "en-US,en;q=0.9",
			expectedLang:   "ru",
		},
		{
			name:           "Russian with quality",
			acceptLanguage: "ru;q=0.8,en;q=0.9",
			expectedLang:   "ru",
		},
		{
			name:           "Invalid Accept-Language",
			acceptLanguage: "invalid",
			expectedLang:   "ru",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tt.acceptLanguage)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedLang, w.Body.String())
		})
	}
}

func TestExtractLanguage(t *testing.T) {
	tests := []struct {
		name           string
		acceptLanguage string
		expected       string
	}{
		{
			name:           "Russian",
			acceptLanguage: "ru",
			expected:       "ru",
		},
		{
			name:           "Russian with region",
			acceptLanguage: "ru-RU",
			expected:       "ru",
		},
		{
			name:           "Empty header",
			acceptLanguage: "",
			expected:       "ru",
		},
		{
			name:           "Multiple languages with Russian first",
			acceptLanguage: "ru,en-US;q=0.9,en;q=0.8",
			expected:       "ru",
		},
		{
			name:           "English (defaults to Russian)",
			acceptLanguage: "en-US",
			expected:       "ru",
		},
		{
			name:           "Invalid format",
			acceptLanguage: "###",
			expected:       "ru",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tt.acceptLanguage)
			}

			result := extractLanguage(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetLanguageFromContext(t *testing.T) {
	t.Run("Language in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), LanguageContextKey, "en")
		lang := GetLanguageFromContext(ctx)
		assert.Equal(t, "en", lang)
	})

	t.Run("No language in context", func(t *testing.T) {
		ctx := context.Background()
		lang := GetLanguageFromContext(ctx)
		assert.Equal(t, "ru", lang)
	})

	t.Run("Invalid type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), LanguageContextKey, 123)
		lang := GetLanguageFromContext(ctx)
		assert.Equal(t, "ru", lang)
	})

	t.Run("Russian in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), LanguageContextKey, "ru")
		lang := GetLanguageFromContext(ctx)
		assert.Equal(t, "ru", lang)
	})
}

func TestLanguageMiddleware_ContextPropagation(t *testing.T) {
	var capturedLang string
	handler := LanguageMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedLang = GetLanguageFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "ru-RU")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, "ru", capturedLang)
	assert.Equal(t, http.StatusOK, w.Code)
}
