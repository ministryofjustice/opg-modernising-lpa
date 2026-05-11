package newforms

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type Localizer interface {
	T(msgid string) string
	Format(msgid string, data map[string]any) string
}

type Error interface {
	Format(localizer Localizer) string
}

type CustomError string

func (e CustomError) Format(_ Localizer) string {
	return string(e)
}

type LocalizedError string

func (e LocalizedError) Format(l Localizer) string {
	return l.T(string(e))
}

type EmptyError struct {
	Field Field
}

func (e EmptyError) Format(l Localizer) string {
	return l.Format("errorEnter", map[string]any{
		"Label": l.T(e.Field.Label),
	})
}

type TooLongError struct {
	Field Field
	Max   int
}

func (e TooLongError) Format(l Localizer) string {
	return l.Format("errorStringTooLong", map[string]any{
		"Label":  l.T(e.Field.Label),
		"Length": e.Max,
	})
}

type LengthError struct {
	Field Field
	Chars int
}

func (e LengthError) Format(l Localizer) string {
	return l.Format("errorStringLength", map[string]any{
		"Label":  l.T(e.Field.Label),
		"Length": e.Chars,
	})
}

type SelectError struct {
	Field Field
}

func (e SelectError) Format(l Localizer) string {
	return l.Format("errorSelect", map[string]any{
		"Label": l.T(e.Field.Label),
	})
}

type EmailError struct {
	Field Field
}

func (e EmailError) Format(l Localizer) string {
	return l.Format("errorEmail", map[string]any{
		"Label": l.T(e.Field.Label),
	})
}

type DateMissingError struct {
	Field Field
	// need to highlight the correct fields, if all then EnterError should be used
	MissingDay   bool
	MissingMonth bool
	MissingYear  bool
}

func (e DateMissingError) Format(l Localizer) string {
	var missing []string
	if e.MissingDay {
		missing = append(missing, lowerFirst(l.T("day")))
	}
	if e.MissingMonth {
		missing = append(missing, lowerFirst(l.T("month")))
	}
	if e.MissingYear {
		missing = append(missing, lowerFirst(l.T("year")))
	}

	return l.Format("errorDateMissing", map[string]any{
		"Label":   l.T(e.Field.Label),
		"Missing": l.T("a") + " " + strings.Join(missing, " "+l.T("and")+" "),
	})
}

type DateMustBeRealError struct {
	Field Field
}

func (e DateMustBeRealError) Format(l Localizer) string {
	return l.Format("errorDateMustBeReal", map[string]any{
		"Label": l.T(e.Field.Label),
	})
}

type DateMustBePastError struct {
	Field Field
}

func (e DateMustBePastError) Format(l Localizer) string {
	return l.Format("errorDateMustBePast", map[string]any{
		"Label": l.T(e.Field.Label),
	})
}

type PhoneError struct {
	Field Field
}

func (e PhoneError) Format(l Localizer) string {
	return l.Format("errorPhone", map[string]any{
		"Label": l.T(e.Field.Label),
	})
}

func lowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
