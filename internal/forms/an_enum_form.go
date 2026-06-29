package forms

import "net/http"

type EnumForm[T anEnum, O any, U anUnmarshallableEnum[T]] struct {
	Form
	Enum *Enum[T, O, U]
}

func NewEnumForm[T anEnum, O any, U anUnmarshallableEnum[T]](label string, options O) *EnumForm[T, O, U] {
	return &EnumForm[T, O, U]{
		Enum: NewEnum[T, O, U]("enum", label, options).Selected(),
	}
}

func (f *EnumForm[T, O, U]) Parse(r *http.Request) bool {
	return f.ParsePostForm(r, f.Enum)
}
