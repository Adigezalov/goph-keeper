package localization

import (
	"context"
	"net/http"
	"strings"

	"golang.org/x/text/language"
)

type contextKey string

const LanguageContextKey contextKey = "language"

func LanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := extractLanguage(r)
		ctx := context.WithValue(r.Context(), LanguageContextKey, lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractLanguage(r *http.Request) string {
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang == "" {
		return "ru"
	}

	tags, _, err := language.ParseAcceptLanguage(acceptLang)
	if err != nil || len(tags) == 0 {
		return "ru"
	}

	lang := tags[0].String()
	lang = strings.ToLower(lang)

	if strings.HasPrefix(lang, "ru") {
		return "ru"
	}

	return "ru"
}

func GetLanguageFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(LanguageContextKey).(string); ok {
		return lang
	}
	return "ru"
}
