package localize

import (
	"encoding/json"
	"fmt"
	"strings"

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

func (b Bundle) For(lang Lang) *Localizer {
	return &Localizer{
		i18n.NewLocalizer(b.Bundle, lang.String()),
		false,
		lang,
	}
}

type Localizer struct {
	*i18n.Localizer
	showTranslationKeys bool
	Lang                Lang
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
	if l.showTranslationKeys {
		return fmt.Sprintf("{%s} [%s]", translation, messageID)
	} else {
		return translation
	}
}

func (l Localizer) ShowTranslationKeys() bool {
	return l.showTranslationKeys
}

func (l *Localizer) SetShowTranslationKeys(s bool) {
	l.showTranslationKeys = s
}

func (l *Localizer) Possessive(s string) string {
	if l.Lang == Cy {
		return "Welsh"
	}

	format := "%s’s"

	if strings.HasSuffix(s, "s") {
		format = "%s’"
	}

	return fmt.Sprintf(format, s)
}

func (l *Localizer) Concat(list []string, joiner string) string {
	if l.Lang == Cy {
		return "Welsh"
	}

	switch len(list) {
	case 0:
		return ""
	case 1:
		return list[0]
	default:
		last := len(list) - 1
		return fmt.Sprintf("%s %s %s", strings.Join(list[:last], ", "), l.T(joiner), list[last])
	}
}
