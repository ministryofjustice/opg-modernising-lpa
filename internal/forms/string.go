package forms

import (
	"net/url"
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

func (f *String) SetInput(s string) {
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
