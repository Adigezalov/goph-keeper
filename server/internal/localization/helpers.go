package localization

import (
	"net/http"
)

func LocalizedError(w http.ResponseWriter, r *http.Request, statusCode int, messageID string, templateData map[string]interface{}) {
	lang := GetLanguageFromContext(r.Context())
	message := Translate(lang, messageID, templateData)
	http.Error(w, message, statusCode)
}

func LocalizedErrorWithFallback(w http.ResponseWriter, r *http.Request, statusCode int, messageID string, fallback string, templateData map[string]interface{}) {
	lang := GetLanguageFromContext(r.Context())
	message := Translate(lang, messageID, templateData)
	if message == messageID {
		message = fallback
	}
	http.Error(w, message, statusCode)
}
