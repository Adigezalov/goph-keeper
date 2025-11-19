package localization

import (
	"embed"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

//go:embed locales/*.toml
var localesFS embed.FS

var bundle *i18n.Bundle
var defaultLang = language.Russian

func init() {
	bundle = i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	entries, err := localesFS.ReadDir("locales")
	if err != nil {
		panic(fmt.Sprintf("failed to read locales directory: %v", err))
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := localesFS.ReadFile("locales/" + entry.Name())
		if err != nil {
			panic(fmt.Sprintf("failed to read locale file %s: %v", entry.Name(), err))
		}

		bundle.MustParseMessageFileBytes(data, entry.Name())
	}
}

func GetLocalizer(lang string) *i18n.Localizer {
	tag, err := language.Parse(lang)
	if err != nil {
		tag = defaultLang
	}

	return i18n.NewLocalizer(bundle, tag.String())
}

func GetDefaultLocalizer() *i18n.Localizer {
	return i18n.NewLocalizer(bundle, defaultLang.String())
}

func Translate(lang string, messageID string, templateData map[string]interface{}) string {
	localizer := GetLocalizer(lang)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		defaultLocalizer := GetDefaultLocalizer()
		msg, err = defaultLocalizer.Localize(&i18n.LocalizeConfig{
			MessageID:    messageID,
			TemplateData: templateData,
		})
		if err != nil {
			return messageID
		}
	}
	return msg
}
