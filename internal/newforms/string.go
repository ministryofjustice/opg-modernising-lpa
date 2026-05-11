package newforms

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
)

var (
	mobileRegex      = regexp.MustCompile(`^(?:07|\+?447)\d{9}$`)
	nonUKMobileRegex = regexp.MustCompile(`^\+\d{4,15}$`)
	phoneRegex       = regexp.MustCompile(`^\+?\d{4,15}$`)
)

type String struct {
	Field
	Value      string
	preprocess []func(string) string
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

func (f *String) Replace(oldnew ...string) *String {
	f.preprocess = append(f.preprocess, func(s string) string {
		replacer := strings.NewReplacer(oldnew...)

		return replacer.Replace(s)
	})

	return f
}

func (f *String) NotEmpty() *String {
	f.validators = append(f.validators, func(s string) Error {
		if s == "" {
			return newEmptyError(f.Field.Label)
		}
		return nil
	})

	return f
}

func (f *String) MaxLength(max int) *String {
	f.validators = append(f.validators, func(s string) Error {
		if len(s) > max {
			return newTooLongError(f.Field.Label, max)
		}
		return nil
	})

	return f
}

func (f *String) Length(chars int, error ...Error) *String {
	f.validators = append(f.validators, func(s string) Error {
		if len(s) != chars {
			if len(error) == 1 {
				return error[0]
			}
			return newLengthError(f.Field.Label, chars)
		}
		return nil
	})

	return f
}

func (f *String) Email() *String {
	f.validators = append(f.validators, func(s string) Error {
		if s != "" {
			if _, err := mail.ParseAddress(fmt.Sprintf("<%s>", s)); err != nil {
				return newEmailError(f.Field.Label)
			}
		}

		return nil
	})
	return f
}

func (f *String) Phone() *String {
	f.validators = append(f.validators, func(s string) Error {
		if s != "" && !phoneRegex.MatchString(strings.ReplaceAll(s, " ", "")) {
			return newPhoneError(f.Field.Label)
		}

		return nil
	})
	return f
}

func (f *String) Mobile() *String {
	f.validators = append(f.validators, func(s string) Error {
		if s != "" && !mobileRegex.MatchString(strings.ReplaceAll(s, " ", "")) {
			return newMobileError(f.Field.Label)
		}

		return nil
	})
	return f
}

func (f *String) NonUKMobile() *String {
	f.validators = append(f.validators, func(s string) Error {
		if s != "" && !nonUKMobileRegex.MatchString(strings.ReplaceAll(s, " ", "")) {
			return newMobileError(f.Field.Label)
		}

		return nil
	})
	return f
}

func (f *String) NoLinks(e ...Error) *String {
	f.validators = append(f.validators, func(s string) Error {
		if strings.Contains(s, "://") {
			if len(e) == 1 {
				return e[0]
			} else {
				return newNoLinksError(f.Field.Label)
			}
		}

		return nil
	})
	return f
}

func (f *String) Parse(values url.Values) {
	f.Input = strings.TrimSpace(values.Get(f.Name))
	f.Value = f.Input

	for _, preprocess := range f.preprocess {
		f.Value = preprocess(f.Value)
	}

	for _, validator := range f.validators {
		if error := validator(f.Value); error != nil {
			f.Error = error
			break
		}
	}
}
