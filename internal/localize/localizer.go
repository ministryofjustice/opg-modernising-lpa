package localize

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
)

type Message struct {
	S string

	// when plural
	One   string
	Other string

	// for Welsh only
	Two  string
	Few  string
	Many string
}

func (m *Message) UnmarshalJSON(text []byte) error {
	var s string
	if err := json.Unmarshal(text, &s); err == nil {
		m.S = s
		return nil
	}

	var v map[string]string
	if err := json.Unmarshal(text, &v); err == nil {
		m.One = v["one"]
		m.Other = v["other"]
		m.Two = v["two"]
		m.Few = v["few"]
		m.Many = v["many"]
		return nil
	}

	return errors.New("message malformed")
}

type Messages map[string]Message

func (m Messages) Find(key string) (string, bool) {
	if msg, ok := m[key]; ok {
		return msg.S, true
	}

	return "", false
}

func (m Messages) FindPlural(key string, count int) (string, bool) {
	msg, ok := m[key]
	if !ok {
		return "", false
	}

	if count == 1 {
		return msg.One, true
	}

	if count == 2 && msg.Two != "" {
		return msg.Two, true
	}

	if count == 3 && msg.Few != "" {
		return msg.Few, true
	}

	if count == 6 && msg.Many != "" {
		return msg.Many, true
	}

	return msg.Other, true
}

type Bundle struct {
	messages map[string]Messages
}

func NewBundle(paths ...string) *Bundle {
	bundle := &Bundle{messages: map[string]Messages{}}

	for _, path := range paths {
		bundle.LoadMessageFile(path)
	}

	return bundle
}

func (b *Bundle) LoadMessageFile(p string) error {
	data, err := os.ReadFile(p)
	if err != nil {
		return err
	}

	var v map[string]Message
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	lang, _ := strings.CutSuffix(path.Base(p), ".json")

	if lang == "en" {
		if err := verifyEn(v); err != nil {
			return err
		}
	} else if lang == "cy" {
		if err := verifyCy(v); err != nil {
			return err
		}
	} else {
		return errors.New("only supports en or cy")
	}

	b.messages[lang] = v
	return nil
}

func verifyEn(v map[string]Message) error {
	for key, message := range v {
		if message.S != "" {
			continue
		}

		if message.One != "" && message.Other != "" && message.Two == "" && message.Few == "" && message.Many == "" {
			continue
		}

		return fmt.Errorf("problem with key: %s", key)
	}

	return nil
}

func verifyCy(v map[string]Message) error {
	for key, message := range v {
		if message.S != "" {
			continue
		}

		if message.One != "" && message.Other != "" && message.Two != "" && message.Few != "" && message.Many != "" {
			continue
		}

		return fmt.Errorf("problem with key: %s", key)
	}

	return nil
}

func (b *Bundle) For(lang Lang) *Localizer {
	return &Localizer{b.messages[lang.String()], false, lang}
}

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

	return l.translate(msg, messageID)
}

func (l Localizer) Format(messageID string, data map[string]interface{}) string {
	msg, ok := l.messages.Find(messageID)
	if !ok {
		return l.translate(messageID, messageID)
	}

	var buf bytes.Buffer
	template.Must(template.New("").Parse(msg)).Execute(&buf, data)
	return l.translate(buf.String(), messageID)
}

func (l Localizer) Count(messageID string, count int) string {
	msg, ok := l.messages.FindPlural(messageID, count)
	if !ok {
		return l.translate(messageID, messageID)
	}

	var buf bytes.Buffer
	template.Must(template.New("").Parse(msg)).Execute(&buf, map[string]int{"PluralCount": count})
	return l.translate(buf.String(), messageID)
}

func (l Localizer) FormatCount(messageID string, count int, data map[string]any) string {
	msg, ok := l.messages.FindPlural(messageID, count)
	if !ok {
		return l.translate(messageID, messageID)
	}

	data["PluralCount"] = count
	var buf bytes.Buffer
	template.Must(template.New("").Parse(msg)).Execute(&buf, data)
	return l.translate(buf.String(), messageID)
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

func LowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
