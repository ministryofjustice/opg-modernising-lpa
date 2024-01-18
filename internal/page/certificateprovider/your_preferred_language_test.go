package certificateprovider

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourPreferredLanguage(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id", ContactLanguagePreference: localize.Cy}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourPreferredLanguageData{
			App: testAppData,
			Form: &form.LanguagePreferenceForm{
				Preference: localize.Cy,
			},
			Options:   localize.LangValues,
			FieldName: form.FieldNames.LanguagePreference,
			Donor:     &actor.DonorProvidedDetails{},
		}).
		Return(nil)

	err := YourPreferredLanguage(template.Execute, certificateProviderStore, donorStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourPreferredLanguageWhenCertificateProviderStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := YourPreferredLanguage(nil, certificateProviderStore, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourPreferredLanguageWhenDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id", ContactLanguagePreference: localize.Cy}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.DonorProvidedDetails{}, expectedError)

	err := YourPreferredLanguage(nil, certificateProviderStore, donorStore)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourPreferredLanguageWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id", ContactLanguagePreference: localize.Cy}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := YourPreferredLanguage(template.Execute, certificateProviderStore, donorStore)(testAppData, w, r)

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
				Get(r.Context()).
				Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)
			certificateProviderStore.EXPECT().
				Put(r.Context(), &actor.CertificateProviderProvidedDetails{LpaID: "lpa-id", ContactLanguagePreference: lang}).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				GetAny(r.Context()).
				Return(&actor.DonorProvidedDetails{}, nil)

			err := YourPreferredLanguage(nil, certificateProviderStore, donorStore)(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.CertificateProvider.ConfirmYourDetails.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostYourPreferredLanguageWhenAttorneyStoreError(t *testing.T) {
	formValues := url.Values{form.FieldNames.LanguagePreference: {localize.En.String()}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(formValues.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	err := YourPreferredLanguage(nil, certificateProviderStore, donorStore)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourPreferredLanguageWhenInvalidData(t *testing.T) {
	formValues := url.Values{form.FieldNames.LanguagePreference: {"not-a-lang"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(formValues.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourPreferredLanguageData{
			App: testAppData,
			Form: &form.LanguagePreferenceForm{
				Error:      errors.New("invalid Lang 'not-a-lang'"),
				ErrorLabel: "whichLanguageYoudLikeUsToUseWhenWeContactYou",
			},
			Options:   localize.LangValues,
			FieldName: form.FieldNames.LanguagePreference,
			Errors:    validation.With(form.FieldNames.LanguagePreference, validation.SelectError{Label: "whichLanguageYoudLikeUsToUseWhenWeContactYou"}),
			Donor:     &actor.DonorProvidedDetails{},
		}).
		Return(nil)

	err := YourPreferredLanguage(template.Execute, certificateProviderStore, donorStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
