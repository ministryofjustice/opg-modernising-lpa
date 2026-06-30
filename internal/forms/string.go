package forms

import (
	"net/url"
	"regexp"
	"strings"
)

// String is a validatable string form field.
type String struct {
	Field
	Value      string
	validators []validator[string]
}

func NewString(name, label string) *String {
	f := &String{}
	f.Name = name
	f.Label = label
	return f
}

func (f *String) Set(s string) {
	f.Input = strings.TrimSpace(s)
}

func (f *String) Parse(values url.Values) {
	f.Input = strings.TrimSpace(values.Get(f.Name))
	f.Value = f.Input

	for _, validator := range f.validators {
		if error := validator.Validate(f.Value); error != nil {
			f.Error = error
			break
		}
	}
}

// WithError overrides the error returned by the previously defined validator.
func (f *String) WithError(replace Error) *String {
	if l := len(f.validators); l > 0 {
		f.validators[l-1] = withError[string]{replace: replace, wrapped: f.validators[l-1]}
	}

	return f
}

// WithErrorLabel overrides the label used by the error returned by the
// previously defined validator.
func (f *String) WithErrorLabel(replace string) *String {
	if l := len(f.validators); l > 0 {
		f.validators[l-1] = withErrorLabel[string]{replace: replace, wrapped: f.validators[l-1]}
	}

	return f
}

type notEmptyValidator struct{ Label string }

func (v notEmptyValidator) Validate(s string) Error {
	if s == "" {
		return newEmptyError(v.Label)
	}
	return nil
}

func (f *String) NotEmpty() *String {
	f.validators = append(f.validators, notEmptyValidator{Label: f.Field.Label})

	return f
}

type maxLengthValidator struct {
	Label string
	Max   int
}

func (v maxLengthValidator) Validate(s string) Error {
	if len(s) > v.Max {
		return newTooLongError(v.Label, v.Max)
	}
	return nil
}

func (f *String) MaxLength(max int) *String {
	f.validators = append(f.validators, maxLengthValidator{Label: f.Field.Label, Max: max})

	return f
}

var phoneRegex = regexp.MustCompile(`^\+?\d{4,15}$`)

type phoneValidator struct {
	Label string
}

func (v phoneValidator) Validate(s string) Error {
	if s != "" && !phoneRegex.MatchString(strings.ReplaceAll(s, " ", "")) {
		return newPhoneError(v.Label)
	}

	return nil
}

func (f *String) Phone() *String {
	f.validators = append(f.validators, phoneValidator{Label: f.Field.Label})

	return f
}
