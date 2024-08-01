package donorpage

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourPreferredLanguage(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourPreferredLanguageData{
			App: testAppData,
			Form: &yourPreferredLanguageForm{
				Contact: localize.Cy,
			},
			Options: localize.LangValues,
		}).
		Return(nil)

	err := YourPreferredLanguage(template.Execute, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaID: "lpa-id", Donor: donordata.Donor{ContactLanguagePreference: localize.Cy}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourPreferredLanguageWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := YourPreferredLanguage(template.Execute, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaID: "lpa-id", Donor: donordata.Donor{ContactLanguagePreference: localize.Cy}})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourPreferredLanguage(t *testing.T) {
	testCases := []localize.Lang{localize.En, localize.Cy}

	for _, lang := range testCases {
		t.Run(lang.String(), func(t *testing.T) {
			formValues := url.Values{"contact-language": {lang.String()}, "lpa-language": {lang.String()}}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(formValues.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.DonorProvidedDetails{
					LpaID: "lpa-id",
					Donor: donordata.Donor{ContactLanguagePreference: lang, LpaLanguagePreference: lang},
				}).
				Return(nil)

			err := YourPreferredLanguage(nil, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaID: "lpa-id"})

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.YourLegalRightsAndResponsibilitiesIfYouMakeLpa.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostYourPreferredLanguageWhenDonorStoreError(t *testing.T) {
	formValues := url.Values{"contact-language": {localize.En.String()}, "lpa-language": {localize.En.String()}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(formValues.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := YourPreferredLanguage(nil, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaID: "lpa-id"})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourPreferredLanguageWhenInvalidData(t *testing.T) {
	formValues := url.Values{"contact-language": {"not-a-lang"}, "lpa-language": {localize.En.String()}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(formValues.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourPreferredLanguageData{
			App: testAppData,
			Form: &yourPreferredLanguageForm{
				Lpa:          localize.En,
				ContactError: errors.New("invalid Lang 'not-a-lang'"),
			},
			Options: localize.LangValues,
			Errors:  validation.With("contact-language", validation.SelectError{Label: "whichLanguageYouWouldLikeUsToUseWhenWeContactYou"}),
		}).
		Return(nil)

	err := YourPreferredLanguage(template.Execute, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaID: "lpa-id"})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadYourPreferredLanguageForm(t *testing.T) {
	form := url.Values{"contact-language": {localize.En.String()}, "lpa-language": {localize.Cy.String()}}
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	assert.Equal(t, &yourPreferredLanguageForm{Contact: localize.En, Lpa: localize.Cy}, readYourPreferredLanguageForm(r))
}

func TestLanguagePreferenceFormValidate(t *testing.T) {
	testcases := map[string]struct {
		form   *yourPreferredLanguageForm
		errors validation.List
	}{
		"valid": {
			form: &yourPreferredLanguageForm{},
		},
		"invalid": {
			form: &yourPreferredLanguageForm{ContactError: errors.New("err"), LpaError: errors.New("arr")},
			errors: validation.With("contact-language", validation.SelectError{Label: "whichLanguageYouWouldLikeUsToUseWhenWeContactYou"}).
				With("lpa-language", validation.SelectError{Label: "theLanguageInWhichYouWouldLikeYourLpaRegistered"}),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
