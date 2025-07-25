// Package localize provides functionality for English and Welsh language
// content.
package localize

import (
	"fmt"
	"strings"
	"time"
	_ "time/tzdata" // To ensure timezone database is available
	"unicode"
	"unicode/utf8"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
)

var london, _ = time.LoadLocation("Europe/London")

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

// defaultLocalizer is instantiated via Bundle.For()
type defaultLocalizer struct {
	messages            Messages
	showTranslationKeys bool
	lang                Lang
}

func (l *defaultLocalizer) T(messageID string) string {
	msg, ok := l.messages.Find(messageID)
	if !ok {
		return l.translate(messageID, messageID)
	}

	return l.translate(msg.S, messageID)
}

func (l *defaultLocalizer) Format(messageID string, data map[string]interface{}) string {
	msg, ok := l.messages.Find(messageID)
	if !ok {
		return l.translate(messageID, messageID)
	}

	return l.translate(msg.Execute(data), messageID)
}

func (l *defaultLocalizer) Count(messageID string, count int) string {
	return l.FormatCount(messageID, count, map[string]any{})
}

func (l *defaultLocalizer) FormatCount(messageID string, count int, data map[string]any) string {
	msg, ok := l.messages.FindPlural(messageID, count)
	if !ok {
		return l.translate(messageID, messageID)
	}

	data["PluralCount"] = count
	return l.translate(msg.Execute(data), messageID)
}

func (l *defaultLocalizer) translate(translation, messageID string) string {
	if l.showTranslationKeys {
		return fmt.Sprintf("{%s} [%s]", translation, messageID)
	} else {
		return translation
	}
}

func (l *defaultLocalizer) ShowTranslationKeys() bool {
	return l.showTranslationKeys
}

func (l *defaultLocalizer) SetShowTranslationKeys(s bool) {
	l.showTranslationKeys = s
}

func (l *defaultLocalizer) Possessive(s string) string {
	if l.lang == Cy {
		return s
	}

	format := "%s’s"

	if strings.HasSuffix(s, "s") {
		format = "%s’"
	}

	return fmt.Sprintf(format, s)
}

func (l *defaultLocalizer) Concat(list []string, joiner string) string {
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

func (l *defaultLocalizer) FormatDate(t date.TimeOrDate) string {
	if t.IsZero() {
		return ""
	}

	if l.lang == Cy {
		return fmt.Sprintf("%d %s %d", t.Day(), monthsCy[t.Month()], t.Year())
	}

	return t.Format("2 January 2006")
}

func (l *defaultLocalizer) FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	londonLoc, _ := time.LoadLocation("Europe/London")
	lt := t.In(londonLoc)

	if l.lang == Cy {
		amPm := "yb"
		if lt.Hour() >= 12 {
			amPm = "yp"
		}

		return fmt.Sprintf("%s%s", lt.Format("3:04"), amPm)
	}

	return lt.Format("3:04pm")
}

func (l *defaultLocalizer) FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	lt := t.In(london)

	if l.lang == Cy {
		amPm := "yb"
		if lt.Hour() >= 12 {
			amPm = "yp"
		}

		return fmt.Sprintf("%d %s %d am %s%s", lt.Day(), monthsCy[lt.Month()], lt.Year(), lt.Format("3:04"), amPm)
	}

	return lt.Format("3:04pm on 2 January 2006")
}

func (l *defaultLocalizer) Lang() Lang {
	return l.lang
}

func LowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
