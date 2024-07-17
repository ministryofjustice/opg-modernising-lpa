package localize

import (
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/dustin/go-humanize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
)

type Localizer struct {
	messages            Messages
	showTranslationKeys bool
	Lang                Lang
}

func (l *Localizer) T(messageID string) string {
	msg, ok := l.messages.Find(messageID)
	if !ok {
		return l.translate(messageID, messageID)
	}

	return l.translate(msg.S, messageID)
}

func (l *Localizer) Format(messageID string, data map[string]interface{}) string {
	msg, ok := l.messages.Find(messageID)
	if !ok {
		return l.translate(messageID, messageID)
	}

	return l.translate(msg.Execute(data), messageID)
}

func (l *Localizer) Count(messageID string, count int) string {
	return l.FormatCount(messageID, count, map[string]any{})
}

func (l *Localizer) FormatCount(messageID string, count int, data map[string]any) string {
	msg, ok := l.messages.FindPlural(messageID, count)
	if !ok {
		return l.translate(messageID, messageID)
	}

	data["PluralCount"] = count
	return l.translate(msg.Execute(data), messageID)
}

func (l *Localizer) translate(translation, messageID string) string {
	if l.showTranslationKeys {
		return fmt.Sprintf("{%s} [%s]", translation, messageID)
	} else {
		return translation
	}
}

func (l *Localizer) ShowTranslationKeys() bool {
	return l.showTranslationKeys
}

func (l *Localizer) SetShowTranslationKeys(s bool) {
	l.showTranslationKeys = s
}

func (l *Localizer) Possessive(s string) string {
	if l.Lang == Cy {
		return s
	}

	format := "%s’s"

	if strings.HasSuffix(s, "s") {
		format = "%s’"
	}

	return fmt.Sprintf(format, s)
}

func (l *Localizer) Concat(list []string, joiner string) string {
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

func (l *Localizer) FormatDate(t date.TimeOrDate) string {
	if t.IsZero() {
		return ""
	}

	if l.Lang == Cy {
		return fmt.Sprintf("%d %s %d", t.Day(), monthsCy[t.Month()], t.Year())
	}

	return t.Format("2 January 2006")
}

func (l *Localizer) FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	if l.Lang == Cy {
		amPm := "yb"
		if t.Hour() >= 12 {
			amPm = "yp"
		}

		return fmt.Sprintf("%s%s", t.Format("3:04"), amPm)
	}

	return t.Format("3:04pm")
}

func (l *Localizer) FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	if l.Lang == Cy {
		amPm := "yb"
		if t.Hour() >= 12 {
			amPm = "yp"
		}

		return fmt.Sprintf("%d %s %d am %s%s", t.Day(), monthsCy[t.Month()], t.Year(), t.Format("3:04"), amPm)
	}

	return t.Format("2 January 2006 at 3:04pm")
}

func (l *Localizer) PenceToPounds(pence int) string {
	return fmt.Sprintf("£%s", humanize.CommafWithDigits(float64(pence)/100, 2))
}

func LowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
