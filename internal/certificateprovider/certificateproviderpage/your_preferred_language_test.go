package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
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
			Form: &form.LanguagePreferenceForm{
				Preference: localize.Cy,
			},
			Options:   localize.LangValues,
			FieldName: form.FieldNames.LanguagePreference,
			Lpa:       &lpadata.Lpa{},
		}).
		Return(nil)

	err := YourPreferredLanguage(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id", ContactLanguagePreference: localize.Cy}, &lpadata.Lpa{})

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

	err := YourPreferredLanguage(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id", ContactLanguagePreference: localize.Cy}, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourPreferredLanguage(t *testing.T) {
	testCases := []localize.Lang{localize.En, localize.Cy}

	for _, lang := range testCases {
		t.Run(lang.String(), func(t *testing.T) {
			formValues := url.Values{form.FieldNames.LanguagePreference: {lang.String()}}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(formValues.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				Put(r.Context(), &certificateproviderdata.Provided{LpaID: "lpa-id", ContactLanguagePreference: lang}).
				Return(nil)

			err := YourPreferredLanguage(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"}, nil)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, certificateprovider.PathConfirmYourDetails.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostYourPreferredLanguageWhenCertificateProviderStoreError(t *testing.T) {
	formValues := url.Values{form.FieldNames.LanguagePreference: {localize.En.String()}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(formValues.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := YourPreferredLanguage(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{}, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourPreferredLanguageWhenInvalidData(t *testing.T) {
	formValues := url.Values{form.FieldNames.LanguagePreference: {"not-a-lang"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(formValues.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourPreferredLanguageData{
			App: testAppData,
			Form: &form.LanguagePreferenceForm{
				ErrorLabel: "whichLanguageYouWouldLikeUsToUseWhenWeContactYou",
			},
			Options:   localize.LangValues,
			FieldName: form.FieldNames.LanguagePreference,
			Errors:    validation.With(form.FieldNames.LanguagePreference, validation.SelectError{Label: "whichLanguageYouWouldLikeUsToUseWhenWeContactYou"}),
			Lpa:       &lpadata.Lpa{},
		}).
		Return(nil)

	err := YourPreferredLanguage(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
