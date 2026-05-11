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

type FormattedError struct {
	Key  string
	Data map[string]any
}

func (e FormattedError) Format(l Localizer) string {
	return l.Format(e.Key, e.Data)
}

func newEmptyError(label string) FormattedError {
	return FormattedError{
		Key:  "errorEnter",
		Data: map[string]any{"Label": label},
	}
}

func newTooLongError(label string, length int) FormattedError {
	return FormattedError{
		Key: "errorStringTooLong",
		Data: map[string]any{
			"Label":  label,
			"Length": length,
		},
	}
}

func newLengthError(label string, length int) FormattedError {
	return FormattedError{
		Key: "errorStringLength",
		Data: map[string]any{
			"Label":  label,
			"Length": length,
		},
	}
}

func NewSelectError(label string) FormattedError {
	return FormattedError{
		Key: "errorSelect",
		Data: map[string]any{
			"Label": label,
		},
	}
}

func newEmailError(label string) FormattedError {
	return FormattedError{
		Key: "errorEmail",
		Data: map[string]any{
			"Label": label,
		},
	}
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
		missing = append(missing, lowerFirst(l.T("day")))
	}
	if e.MissingMonth {
		missing = append(missing, lowerFirst(l.T("month")))
	}
	if e.MissingYear {
		missing = append(missing, lowerFirst(l.T("year")))
	}

	return l.Format("errorDateMissing", map[string]any{
		"Label":   l.T(e.Label),
		"Missing": l.T("a") + " " + strings.Join(missing, " "+l.T("and")+" "),
	})
}

func newDateMustBeRealError(label string) FormattedError {
	return FormattedError{
		Key: "errorDateMustBeReal",
		Data: map[string]any{
			"Label": label,
		},
	}
}

func newDateMustBePastError(label string) FormattedError {
	return FormattedError{
		Key: "errorDateMustBePast",
		Data: map[string]any{
			"Label": label,
		},
	}
}

func newPhoneError(label string) FormattedError {
	return FormattedError{
		Key: "errorPhone",
		Data: map[string]any{
			"Label": label,
		},
	}
}

func newMobileError(label string) FormattedError {
	return FormattedError{
		Key: "errorMobile",
		Data: map[string]any{
			"Label": label,
		},
	}
}

func newNoLinksError(label string) FormattedError {
	return FormattedError{
		Key: "errorNoLinks",
		Data: map[string]any{
			"Label": label,
		},
	}
}

func NewIncorrectError(label string) FormattedError {
	return FormattedError{
		Key: "errorIncorrect",
		Data: map[string]any{
			"Label": label,
		},
	}
}

func lowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
