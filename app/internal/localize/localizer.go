package localize

import (
	"encoding/json"
	"fmt"

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
	return Localizer{i18n.NewLocalizer(b.Bundle, lang...), false}
}

type Localizer struct {
	*i18n.Localizer
	ShowTransKeys bool
}

func (l Localizer) T(messageID string) string {
	msg, err := l.Localize(&i18n.LocalizeConfig{MessageID: messageID})

	if err != nil {
		return l.translate(messageID, messageID)
	}

	return l.translate(msg, messageID)
}

func (l Localizer) Format(messageID string, data map[string]interface{}) string {
	return l.translate(l.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID, TemplateData: data}), messageID)
}

func (l Localizer) Count(messageID string, count int) string {
	return l.translate(l.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID, PluralCount: count}), messageID)
}

func (l Localizer) FormatCount(messageID string, count int, data map[string]interface{}) string {
	data["PluralCount"] = count
	return l.translate(l.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID, PluralCount: count, TemplateData: data}), messageID)
}

func (l Localizer) translate(translation, messageID string) string {
	if l.ShowTransKeys {
		return fmt.Sprintf("{%s} [%s]", translation, messageID)
	} else {
		return translation
	}
}
