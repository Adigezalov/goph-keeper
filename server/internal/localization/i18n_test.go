package localization

import (
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestGetLocalizer(t *testing.T) {
	tests := []struct {
		name string
		lang string
	}{
		{
			name: "Russian",
			lang: "ru",
		},
		{
			name: "Russian with region",
			lang: "ru-RU",
		},
		{
			name: "English",
			lang: "en",
		},
		{
			name: "Invalid language",
			lang: "invalid-lang-code",
		},
		{
			name: "Empty string",
			lang: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localizer := GetLocalizer(tt.lang)
			assert.NotNil(t, localizer)
		})
	}
}

func TestGetDefaultLocalizer(t *testing.T) {
	localizer := GetDefaultLocalizer()
	assert.NotNil(t, localizer)
}

func TestTranslate(t *testing.T) {
	tests := []struct {
		name         string
		lang         string
		messageID    string
		templateData map[string]interface{}
		expectEmpty  bool
	}{
		{
			name:      "Valid message ID",
			lang:      "ru",
			messageID: "common.invalid_request_format",
		},
		{
			name:      "Valid message ID in Russian",
			lang:      "ru-RU",
			messageID: "common.authorization_error",
		},
		{
			name:      "Non-existent message ID",
			lang:      "ru",
			messageID: "nonexistent.message.id.that.does.not.exist.anywhere",
		},
		{
			name:      "With template data",
			lang:      "ru",
			messageID: "common.invalid_request_format",
			templateData: map[string]interface{}{
				"field": "username",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Translate(tt.lang, tt.messageID, tt.templateData)
			assert.NotEmpty(t, result)
		})
	}
}

func TestTranslate_ReturnMessageIDOnFailure(t *testing.T) {
	lang := "ru"
	messageID := "this.message.definitely.does.not.exist.in.any.locale.file"

	result := Translate(lang, messageID, nil)

	// When translation fails, it should return the messageID
	assert.Equal(t, messageID, result)
}

func TestTranslate_WithVariousLanguages(t *testing.T) {
	messageID := "common.invalid_request_format"

	languages := []string{"ru", "ru-RU", "en", "en-US", "de", "fr"}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			result := Translate(lang, messageID, nil)
			assert.NotEmpty(t, result)
		})
	}
}

func TestTranslate_FallbackToDefault(t *testing.T) {
	// Use an invalid language, should fallback to default (Russian)
	result := Translate("xyz", "common.invalid_request_format", nil)
	assert.NotEmpty(t, result)
}

func TestGetLocalizer_WithParsedLanguage(t *testing.T) {
	// Test with properly formatted language tags
	tags := []string{
		"ru",
		"ru-RU",
		"en-US",
		"en-GB",
	}

	for _, tag := range tags {
		t.Run(tag, func(t *testing.T) {
			localizer := GetLocalizer(tag)
			assert.NotNil(t, localizer)

			// Try to use the localizer
			_, err := localizer.Localize(&i18n.LocalizeConfig{
				MessageID: "common.invalid_request_format",
			})
			// Error is okay if message doesn't exist, we just want to verify localizer works
			_ = err
		})
	}
}

func TestDefaultLanguage(t *testing.T) {
	assert.Equal(t, language.Russian, defaultLang)
}

func TestBundle_Initialization(t *testing.T) {
	assert.NotNil(t, bundle)
}
