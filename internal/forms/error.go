package forms

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type Localizer interface {
	T(msgid string) string
	Format(msgid string, data map[string]any) string
}

type withError[T any] struct {
	replace Error
	wrapped validator[T]
}

func (v withError[T]) Validate(t T) Error {
	if v.wrapped.Validate(t) != nil {
		return v.replace
	}
	return nil
}

type withErrorLabel[T any] struct {
	replace string
	wrapped validator[T]
}

func (v withErrorLabel[T]) Validate(t T) Error {
	error := v.wrapped.Validate(t)

	if ferror, ok := error.(formattedError); ok {
		ferror.Data["Label"] = v.replace
		return ferror
	}

	return error
}

type Error interface {
	Format(localizer Localizer) string
}

type ErrorMessage string

func (e ErrorMessage) Format(l Localizer) string {
	return l.T(string(e))
}

type formattedError struct {
	Key  string
	Data map[string]any
}

func (e formattedError) Format(l Localizer) string {
	return l.Format(e.Key, e.Data)
}

func newEmptyError(label string) formattedError {
	return formattedError{
		Key:  "errorEnter",
		Data: map[string]any{"Label": label},
	}
}

func newTooLongError(label string, length int) formattedError {
	return formattedError{
		Key: "errorStringTooLong",
		Data: map[string]any{
			"Label":  label,
			"Length": length,
		},
	}
}

func newSelectError(label string) formattedError {
	return formattedError{
		Key:  "errorSelect",
		Data: map[string]any{"Label": label},
	}
}

func newPhoneError(label string) formattedError {
	return formattedError{
		Key:  "errorPhone",
		Data: map[string]any{"Label": label},
	}
}

func newDateMustBeRealError(label string) formattedError {
	return formattedError{
		Key:  "errorDateMustBeReal",
		Data: map[string]any{"Label": label},
	}
}

func newDateMustBePastError(label string) formattedError {
	return formattedError{
		Key:  "errorDateMustBePast",
		Data: map[string]any{"Label": label},
	}
}

func newDateMustBeBeforeYearsError(label string, years int) formattedError {
	return formattedError{
		Key:  "errorDateMustBeBeforeYears",
		Data: map[string]any{"Label": label, "Years": years},
	}
}

type dateMissingError struct {
	Label string
	// need to highlight the correct fields, if all then EnterError should be used
	MissingDay   bool
	MissingMonth bool
	MissingYear  bool
}

func (e dateMissingError) Format(l Localizer) string {
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
		"Label": l.T(e.Label),
		// TODO: check whether this produces the correct Welsh
		"Missing": l.T("a") + " " + strings.Join(missing, " "+l.T("and")+" "),
	})
}

func lowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
