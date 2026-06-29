package forms

import (
	"net/url"
	"strings"
)

type anEnum interface {
	String() string
}

type anUnmarshallableEnum[T anEnum] interface {
	UnmarshalText([]byte) error
	*T
}

// Enum is a validatable enum form field.
type Enum[T anEnum, O any, U anUnmarshallableEnum[T]] struct {
	Field
	Value      T
	Options    O
	parseError error
	validators []validator[anEnum]
}

func NewEnum[T anEnum, O any, U anUnmarshallableEnum[T]](name, label string, values O) *Enum[T, O, U] {
	f := &Enum[T, O, U]{}
	f.Name = name
	f.Label = label
	f.Options = values
	return f
}

func (f *Enum[T, O, U]) Set(e T) {
	f.Input = e.String()
}

func (f *Enum[T, O, U]) Parse(values url.Values) {
	f.Input = strings.TrimSpace(values.Get(f.Name))

	f.Value = *new(T) // reset value before parsing
	f.parseError = U(&f.Value).UnmarshalText([]byte(f.Input))

	for _, validator := range f.validators {
		if error := validator.Validate(f.Value); error != nil {
			f.Error = error
			break
		}
	}
}

// WithError overrides the error returned by the previously defined validator.
func (f *Enum[T, O, U]) WithError(replace Error) *Enum[T, O, U] {
	if l := len(f.validators); l > 0 {
		f.validators[l-1] = withError[anEnum]{replace: replace, wrapped: f.validators[l-1]}
	}

	return f
}

type enumSelectedValidator struct {
	label      string
	parseError *error // pointer as the value is set after this value is constructed
}

func (v enumSelectedValidator) Validate(e anEnum) Error {
	if *v.parseError != nil {
		return newSelectError(v.label)
	}

	et, ok := e.(interface{ Empty() bool })
	if ok && et.Empty() {
		return newSelectError(v.label)
	}

	return nil
}

func (f *Enum[T, O, U]) Selected() *Enum[T, O, U] {
	f.validators = append(f.validators, enumSelectedValidator{
		label:      f.Field.Label,
		parseError: &f.parseError,
	})

	return f
}
