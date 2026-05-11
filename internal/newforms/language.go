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
	Language *Language
	Errors   []Field
}

func NewLanguageForm(label string) *LanguageForm {
	return &LanguageForm{
		Language: NewLanguage(label).Selected(),
	}
}

func (f *LanguageForm) Parse(r *http.Request) bool {
	f.Errors = ParsePostForm(r, f.Language)

	return len(f.Errors) == 0
}
