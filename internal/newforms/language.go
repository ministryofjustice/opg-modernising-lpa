package newforms

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type Language = Enum[localize.Lang, localize.LangOptions, *localize.Lang]

func NewLanguage(label string) *Language {
	return NewEnum[localize.Lang]("language-preference", label, localize.LangValues)
}

type LanguageForm struct {
	Form
	Language *Language
}

func NewLanguageForm(label string) *LanguageForm {
	return &LanguageForm{
		Language: NewLanguage(label).Selected(),
	}
}

func (f *LanguageForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r, f.Language)
}
