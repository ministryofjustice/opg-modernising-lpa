package forms

import (
	"net/url"
	"strings"
)

// Bool is a validatable boolean form field.
type Bool struct {
	Field
	Value      bool
	validators []validator[bool]
}

func NewBool(name, label string) *Bool {
	f := &Bool{}
	f.Name = name
	f.Label = label
	return f
}

func (f *Bool) Set(value bool) {
	if value {
		f.Input = "1"
	} else {
		f.Input = ""
	}
}

func (f *Bool) Parse(values url.Values) {
	f.Input = strings.TrimSpace(values.Get(f.Name))
	f.Value = f.Input == "1"

	for _, validator := range f.validators {
		if error := validator.Validate(f.Value); error != nil {
			f.Error = error
			break
		}
	}
}

// WithError overrides the error returned by the previously defined validator.
func (f *Bool) WithError(replace Error) *Bool {
	if l := len(f.validators); l > 0 {
		f.validators[l-1] = withError[bool]{replace: replace, wrapped: f.validators[l-1]}
	}

	return f
}

type boolSelectedValidator struct {
	label string
}

func (v boolSelectedValidator) Validate(value bool) Error {
	if !value {
		return newSelectError(v.label)
	}

	return nil
}

func (f *Bool) Selected() *Bool {
	f.validators = append(f.validators, boolSelectedValidator{label: f.Field.Label})

	return f
}
