package form

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type LanguagePreferenceForm struct {
	Preference localize.Lang
	ErrorLabel string
}

func ReadLanguagePreferenceForm(r *http.Request, errorLabel string) *LanguagePreferenceForm {
	preference, _ := localize.ParseLang(PostFormString(r, FieldNames.LanguagePreference))

	return &LanguagePreferenceForm{
		Preference: preference,
		ErrorLabel: errorLabel,
	}
}

func (f *LanguagePreferenceForm) Validate() validation.List {
	var errors validation.List

	errors.Enum(FieldNames.LanguagePreference, f.ErrorLabel, f.Preference,
		validation.Selected())

	return errors
}
