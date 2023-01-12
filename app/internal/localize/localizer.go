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
		if l.ShowTransKeys {
			return fmt.Sprintf("%s [%s]", messageID, messageID)
		} else {
			return messageID
		}
	}

	if l.ShowTransKeys {
		return fmt.Sprintf("%s [%s]", msg, messageID)
	} else {
		return msg
	}
}

func (l Localizer) Format(messageID string, data map[string]interface{}) string {
	if l.ShowTransKeys {
		return fmt.Sprintf("%s [%s]", l.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID, TemplateData: data}), messageID)
	} else {
		return l.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID, TemplateData: data})
	}
}

func (l Localizer) Count(messageID string, count int) string {
	if l.ShowTransKeys {
		return fmt.Sprintf("%s [%s]", l.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID, PluralCount: count}), messageID)
	} else {
		return l.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID, PluralCount: count})
	}
}

func (l Localizer) FormatCount(messageID string, count int, data map[string]interface{}) string {
	data["PluralCount"] = count

	if l.ShowTransKeys {
		return fmt.Sprintf("%s [%s]", l.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID, PluralCount: count, TemplateData: data}), messageID)
	} else {
		return l.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID, PluralCount: count, TemplateData: data})
	}
}
