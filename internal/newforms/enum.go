package newforms

import (
	"net/http"
	"net/url"
	"strings"
)

type enumType interface {
	String() string
}

type enumEmptyType interface {
	Empty() bool
}

type unmarshallable[T enumType] interface {
	UnmarshalText([]byte) error
	*T
}

type Enum[T enumType, TO any, TP unmarshallable[T]] struct {
	Field
	Options    TO
	Value      T
	preprocess func(string) string
	validators []func(enumType) Error
	orDefault  *T
}

func NewEnum[T enumType, TO any, TP unmarshallable[T]](name, label string, options TO) *Enum[T, TO, TP] {
	f := &Enum[T, TO, TP]{}
	f.Name = name
	f.Label = label
	f.Options = options
	return f
}

func (f *Enum[T, TO, TP]) SetInput(value T) {
	f.Input = value.String()
}

func (f *Enum[T, TO, TP]) Selected() *Enum[T, TO, TP] {
	f.validators = append(f.validators, func(e enumType) Error {
		et, ok := e.(enumEmptyType)
		if ok && et.Empty() {
			return SelectError{Field: f.Field}
		}

		return nil
	})

	return f
}

func (f *Enum[T, TO, TP]) OrDefault(value T) *Enum[T, TO, TP] {
	f.orDefault = &value
	return f
}

func (f *Enum[T, TO, TP]) Parse(values url.Values) {
	f.Input = strings.TrimSpace(values.Get(f.Name))
	if err := TP(&f.Value).UnmarshalText([]byte(f.Input)); err != nil {
		if f.orDefault != nil {
			f.Value = *f.orDefault
		} else {
			f.Error = SelectError{Field: f.Field}
			return
		}
	}

	for _, validator := range f.validators {
		if error := validator(f.Value); error != nil {
			f.Error = error
			break
		}
	}
}

type EnumForm[T enumType, TO any, TP unmarshallable[T]] struct {
	Enum   *Enum[T, TO, TP]
	Errors []Field
}

func NewEnumForm[T enumType, TO any, TP unmarshallable[T]](label string, options TO) *EnumForm[T, TO, TP] {
	return &EnumForm[T, TO, TP]{
		Enum: NewEnum[T, TO, TP]("enum", label, options).Selected(),
	}
}

func (f *EnumForm[T, TO, TP]) Parse(r *http.Request) bool {
	f.Errors = ParsePostForm(r, f.Enum)

	return len(f.Errors) == 0
}
