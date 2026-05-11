package newforms

import (
	"net/http"
	"net/url"
	"strings"
)

type Bool struct {
	Field
	Value      bool
	validators []func(bool) Error
}

func NewBool(name, label string) *Bool {
	f := &Bool{}
	f.Name = name
	f.Label = label
	return f
}

func (f *Bool) SetInput(value bool) {
	if value {
		f.Input = "1"
	}
}

func (f *Bool) True(msg string) *Bool {
	f.validators = append(f.validators, func(b bool) Error {
		if !b {
			return CustomError(msg)
		}
		return nil
	})

	return f
}

func (f *Bool) Parse(values url.Values) {
	f.Input = strings.TrimSpace(values.Get(f.Name))
	f.Value = f.Input == "1"

	for _, validator := range f.validators {
		if error := validator(f.Value); error != nil {
			f.Error = error
			break
		}
	}
}

type BoolForm struct {
	Bool   *Bool
	Errors []Field
}

func NewBoolForm(label, errorMessage string) *BoolForm {
	return &BoolForm{
		Bool: NewBool("bool", label).True(errorMessage),
	}
}

func (f *BoolForm) Parse(r *http.Request) bool {
	f.Errors = ParsePostForm(r, f.Bool)

	return len(f.Errors) == 0
}
