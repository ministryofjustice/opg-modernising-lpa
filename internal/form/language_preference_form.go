package form

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type LanguagePreferenceForm struct {
	Preference localize.Lang
	Error      error
	ErrorLabel string
}

func ReadLanguagePreferenceForm(r *http.Request, errorLabel string) *LanguagePreferenceForm {
	preference, err := localize.ParseLang(PostFormString(r, FieldNames.LanguagePreference))

	return &LanguagePreferenceForm{
		Preference: preference,
		Error:      err,
		ErrorLabel: errorLabel,
	}
}

func (f *LanguagePreferenceForm) Validate() validation.List {
	var errors validation.List

	errors.Error(FieldNames.LanguagePreference, f.ErrorLabel, f.Error,
		validation.Selected())

	return errors
}
