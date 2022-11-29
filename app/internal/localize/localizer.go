package localize

import (
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Bundle struct {
	*i18n.Bundle
}

func NewBundle(paths ...string) Bundle {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	for _, path := range paths {
		bundle.LoadMessageFile(path)
	}

	return Bundle{bundle}
}

func (b Bundle) For(lang ...string) Localizer {
	return Localizer{i18n.NewLocalizer(b.Bundle, lang...)}
}

type Localizer struct {
	*i18n.Localizer
}

func (l Localizer) T(messageID string) string {
	return l.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID})
}

func (l Localizer) Format(messageID string, data map[string]interface{}) string {
	return l.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID, TemplateData: data})
}

func (l Localizer) Count(messageID string, count int) string {
	return l.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID, PluralCount: count})
}

func (l Localizer) FormatCount(messageID string, count int, data map[string]interface{}) string {
	data["PluralCount"] = count
	return l.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID, PluralCount: count, TemplateData: data})
}
