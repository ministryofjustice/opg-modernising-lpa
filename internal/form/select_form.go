package form

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enumType interface {
	Empty() bool
}

type unmarshallable[T enumType] interface {
	UnmarshalText([]byte) error
	*T
}

type SelectForm[T enumType, TOptions any, TP unmarshallable[T]] struct {
	Selected   T
	FieldName  string
	Options    TOptions
	ErrorLabel string
}

func NewSelectForm[T enumType, TOptions any, TP unmarshallable[T]](
	selected T,
	values TOptions,
	errorLabel string,
) *SelectForm[T, TOptions, TP] {
	return &SelectForm[T, TOptions, TP]{
		Selected:   selected,
		FieldName:  FieldNames.Select,
		Options:    values,
		ErrorLabel: errorLabel,
	}
}

func NewEmptySelectForm[T enumType, TOptions any, TP unmarshallable[T]](
	values TOptions,
	errorLabel string,
) *SelectForm[T, TOptions, TP] {
	return &SelectForm[T, TOptions, TP]{
		FieldName:  FieldNames.Select,
		Options:    values,
		ErrorLabel: errorLabel,
	}
}

func (f *SelectForm[T, TOptions, TP]) Read(r *http.Request) {
	_ = TP(&f.Selected).UnmarshalText([]byte(PostFormString(r, f.FieldName)))
}

func (f *SelectForm[T, TOptions, TP]) Validate() validation.List {
	var errors validation.List

	errors.Enum(f.FieldName, f.ErrorLabel, f.Selected,
		validation.Selected())

	return errors
}
