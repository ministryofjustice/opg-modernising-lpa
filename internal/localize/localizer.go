// Package localize provides functionality for English and Welsh language
// content.
package localize

import (
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
)

type Localizer interface {
	Concat(list []string, joiner string) string
	Count(messageID string, count int) string
	Format(messageID string, data map[string]interface{}) string
	FormatCount(messageID string, count int, data map[string]any) string
	FormatDate(t date.TimeOrDate) string
	FormatDateTime(t time.Time) string
	FormatTime(t time.Time) string
	Lang() Lang
	Possessive(s string) string
	ShowTranslationKeys() bool
	SetShowTranslationKeys(s bool)
	T(messageID string) string
}

type DefaultLocalizer struct {
	messages            Messages
	showTranslationKeys bool
	lang                Lang
}

func (l *DefaultLocalizer) T(messageID string) string {
	msg, ok := l.messages.Find(messageID)
	if !ok {
		return l.translate(messageID, messageID)
	}

	return l.translate(msg.S, messageID)
}

func (l *DefaultLocalizer) Format(messageID string, data map[string]interface{}) string {
	msg, ok := l.messages.Find(messageID)
	if !ok {
		return l.translate(messageID, messageID)
	}

	return l.translate(msg.Execute(data), messageID)
}

func (l *DefaultLocalizer) Count(messageID string, count int) string {
	return l.FormatCount(messageID, count, map[string]any{})
}

func (l *DefaultLocalizer) FormatCount(messageID string, count int, data map[string]any) string {
	msg, ok := l.messages.FindPlural(messageID, count)
	if !ok {
		return l.translate(messageID, messageID)
	}

	data["PluralCount"] = count
	return l.translate(msg.Execute(data), messageID)
}

func (l *DefaultLocalizer) translate(translation, messageID string) string {
	if l.showTranslationKeys {
		return fmt.Sprintf("{%s} [%s]", translation, messageID)
	} else {
		return translation
	}
}

func (l *DefaultLocalizer) ShowTranslationKeys() bool {
	return l.showTranslationKeys
}

func (l *DefaultLocalizer) SetShowTranslationKeys(s bool) {
	l.showTranslationKeys = s
}

func (l *DefaultLocalizer) Possessive(s string) string {
	if l.lang == Cy {
		return s
	}

	format := "%s’s"

	if strings.HasSuffix(s, "s") {
		format = "%s’"
	}

	return fmt.Sprintf(format, s)
}

func (l *DefaultLocalizer) Concat(list []string, joiner string) string {
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

var monthsCy = map[time.Month]string{
	time.January:   "Ionawr",
	time.February:  "Chwefror",
	time.March:     "Mawrth",
	time.April:     "Ebrill",
	time.May:       "Mai",
	time.June:      "Mehefin",
	time.July:      "Gorffennaf",
	time.August:    "Awst",
	time.September: "Medi",
	time.October:   "Hydref",
	time.November:  "Tachwedd",
	time.December:  "Rhagfyr",
}

func (l *DefaultLocalizer) FormatDate(t date.TimeOrDate) string {
	if t.IsZero() {
		return ""
	}

	if l.lang == Cy {
		return fmt.Sprintf("%d %s %d", t.Day(), monthsCy[t.Month()], t.Year())
	}

	return t.Format("2 January 2006")
}

func (l *DefaultLocalizer) FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	if l.lang == Cy {
		amPm := "yb"
		if t.Hour() >= 12 {
			amPm = "yp"
		}

		return fmt.Sprintf("%s%s", t.Format("3:04"), amPm)
	}

	return t.Format("3:04pm")
}

func (l *DefaultLocalizer) FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	if l.lang == Cy {
		amPm := "yb"
		if t.Hour() >= 12 {
			amPm = "yp"
		}

		return fmt.Sprintf("%d %s %d am %s%s", t.Day(), monthsCy[t.Month()], t.Year(), t.Format("3:04"), amPm)
	}

	return t.Format("2 January 2006 at 3:04pm")
}

func (l *DefaultLocalizer) Lang() Lang {
	return l.lang
}

func LowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
