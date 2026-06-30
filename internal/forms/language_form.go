package forms

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type LanguageForm struct {
	Form
	Language *Enum[localize.Lang, localize.LangOptions, *localize.Lang]
}

func NewLanguageForm(label string) *LanguageForm {
	return &LanguageForm{
		Language: NewEnum[localize.Lang]("language", label, localize.LangValues).Selected(),
	}
}

func (f *LanguageForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r, f.Language)
}
