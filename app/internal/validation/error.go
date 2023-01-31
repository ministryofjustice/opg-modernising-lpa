package validation

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type FormattableError interface {
	Format(Localizer) string
}

type SelectedError struct {
	Label string
}

func (e SelectedError) Format(l Localizer) string {
	return l.Format("errorSelect", map[string]any{
		"Label": l.T(e.Label),
	})
}

type CustomError struct {
	Label string
}

func (e CustomError) Format(l Localizer) string {
	return l.T(e.Label)
}

type SelectError struct {
	Label string
}

func (e SelectError) Format(l Localizer) string {
	return l.Format("errorSelect", map[string]any{
		"Label": lowerFirst(l.T(e.Label)),
	})
}

type EnterError struct {
	Label string
}

func (e EnterError) Format(l Localizer) string {
	return l.Format("errorEnter", map[string]any{
		"Label": lowerFirst(l.T(e.Label)),
	})
}

type StringTooLongError struct {
	Label  string
	Length int
}

func (e StringTooLongError) Format(l Localizer) string {
	return l.Format("errorStringTooLong", map[string]any{
		"Label":  l.T(e.Label),
		"Length": e.Length,
	})
}

type StringLengthError struct {
	Label  string
	Length int
}

func (e StringLengthError) Format(l Localizer) string {
	return l.Format("errorStringLength", map[string]any{
		"Label":  l.T(e.Label),
		"Length": e.Length,
	})
}

type MobileError struct {
	Label string
}

func (e MobileError) Format(l Localizer) string {
	return l.Format("errorMobile", map[string]any{
		"Label": l.T(e.Label),
	})
}

type EmailError struct {
	Label string
}

func (e EmailError) Format(l Localizer) string {
	return l.Format("errorEmail", map[string]any{
		"Label": l.T(e.Label),
	})
}

type DateMissingError struct {
	Label string
	// need to highlight the correct fields, if all then EnterError should be used
	MissingDay   bool
	MissingMonth bool
	MissingYear  bool
}

func (e DateMissingError) Format(l Localizer) string {
	var missing []string
	if e.MissingDay {
		missing = append(missing, l.T("day"))
	}
	if e.MissingMonth {
		missing = append(missing, l.T("month"))
	}
	if e.MissingYear {
		missing = append(missing, l.T("year"))
	}

	return l.Format("errorDateMissing", map[string]any{
		"Label":   l.T(e.Label),
		"Missing": l.T("a") + " " + strings.Join(missing, " "+l.T("and")+" "),
	})
}

type DateMustBeRealError struct {
	Label string
}

func (e DateMustBeRealError) Format(l Localizer) string {
	return l.Format("errorDateMustBeReal", map[string]any{
		"Label": l.T(e.Label),
	})
}

type DateMustBePastError struct {
	Label string
}

func (e DateMustBePastError) Format(l Localizer) string {
	return l.Format("errorDateMustBePast", map[string]any{
		"Label": l.T(e.Label),
	})
}

type AddressSelectedError struct {
	Label string
}

func (e AddressSelectedError) Format(l Localizer) string {
	return l.Format("errorSelect", map[string]any{
		"Label": lowerFirst(l.T(e.Label)),
	})
}

func lowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
