package newforms

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
)

type String struct {
	Field
	Value      string
	preprocess func(string) string
	validators []func(string) Error
}

func NewString(name, label string) *String {
	f := &String{}
	f.Name = name
	f.Label = label
	return f
}

func (f *String) SetInput(s string) {
	f.Input = s
}

func (f *String) Preprocess(fn func(string) string) *String {
	f.preprocess = fn
	return f
}

func (f *String) NotEmpty() *String {
	f.validators = append(f.validators, func(s string) Error {
		if s == "" {
			return EmptyError{Field: f.Field}
		}
		return nil
	})

	return f
}

func (f *String) NotEmptyCustom(error Error) *String {
	f.validators = append(f.validators, func(s string) Error {
		if s == "" {
			return error
		}
		return nil
	})

	return f
}

func (f *String) MaxLength(max int) *String {
	f.validators = append(f.validators, func(s string) Error {
		if len(s) > max {
			return TooLongError{Field: f.Field, Max: max}
		}
		return nil
	})

	return f
}

func (f *String) Length(chars int) *String {
	f.validators = append(f.validators, func(s string) Error {
		if len(s) != chars {
			return LengthError{Field: f.Field, Chars: chars}
		}
		return nil
	})

	return f
}

func (f *String) MatchesCustom(re string, error Error) *String {
	compiled := regexp.MustCompile(re)
	f.validators = append(f.validators, func(s string) Error {
		if !compiled.MatchString(s) {
			return error
		}
		return nil
	})

	return f
}

func (f *String) Email() *String {
	f.validators = append(f.validators, func(s string) Error {
		if s != "" {
			if _, err := mail.ParseAddress(fmt.Sprintf("<%s>", s)); err != nil {
				return EmailError{Field: f.Field}
			}
		}

		return nil
	})
	return f
}

var phoneRegex = regexp.MustCompile(`^\+?\d{4,15}$`)

func (f *String) Phone() *String {
	f.validators = append(f.validators, func(s string) Error {
		if s != "" && !phoneRegex.MatchString(strings.ReplaceAll(s, " ", "")) {
			return PhoneError{Field: f.Field}
		}

		return nil
	})
	return f
}

func (f *String) Custom(fn func(string) Error) *String {
	f.validators = append(f.validators, fn)
	return f
}

func (f *String) Parse(values url.Values) {
	f.Input = strings.TrimSpace(values.Get(f.Name))
	f.Value = f.Input

	if f.preprocess != nil {
		f.Value = f.preprocess(f.Value)
	}

	for _, validator := range f.validators {
		if error := validator(f.Value); error != nil {
			f.Error = error
			break
		}
	}
}
