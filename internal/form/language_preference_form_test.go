package form

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestReadLanguagePreferenceForm(t *testing.T) {
	form := url.Values{FieldNames.LanguagePreference: {localize.En.String()}}
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	assert.Equal(t, &LanguagePreferenceForm{Preference: localize.En, ErrorLabel: "a-label"}, ReadLanguagePreferenceForm(r, "a-label"))
}

func TestLanguagePreferenceFormValidate(t *testing.T) {
	testcases := map[string]struct {
		form   *LanguagePreferenceForm
		errors validation.List
	}{
		"valid": {
			form: &LanguagePreferenceForm{Preference: localize.En},
		},
		"invalid": {
			form:   &LanguagePreferenceForm{ErrorLabel: "a-label"},
			errors: validation.With(FieldNames.LanguagePreference, validation.SelectError{Label: "a-label"}),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
