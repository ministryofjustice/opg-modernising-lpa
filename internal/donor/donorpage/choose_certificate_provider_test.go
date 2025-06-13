package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviders := []donordata.CertificateProvider{{FirstNames: "John"}}

	service := newMockCertificateProviderService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return(certificateProviders, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseCertificateProviderData{
			App:                  testAppData,
			Form:                 &chooseCertificateProviderForm{},
			Donor:                &donordata.Provided{},
			CertificateProviders: certificateProviders,
		}).
		Return(nil)

	err := ChooseCertificateProvider(template.Execute, service)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseCertificateProviderWhenNoReusableCertificateProviders(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	service := newMockCertificateProviderService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return(nil, nil)

	err := ChooseCertificateProvider(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathCertificateProviderDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetChooseCertificateProviderWhenError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	service := newMockCertificateProviderService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return(nil, expectedError)

	err := ChooseCertificateProvider(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestGetChooseCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	service := newMockCertificateProviderService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return([]donordata.CertificateProvider{{FirstNames: "John"}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ChooseCertificateProvider(template.Execute, service)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseCertificateProvider(t *testing.T) {
	form := url.Values{
		"option": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviders := []donordata.CertificateProvider{{FirstNames: "John"}, {FirstNames: "Dave"}}

	service := newMockCertificateProviderService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return(certificateProviders, nil)
	service.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:               "lpa-id",
			CertificateProvider: certificateProviders[1],
		}).
		Return(nil)

	err := ChooseCertificateProvider(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathCertificateProviderSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseCertificateProviderWhenNew(t *testing.T) {
	form := url.Values{
		"option": {"new"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviders := []donordata.CertificateProvider{{FirstNames: "John"}, {FirstNames: "Dave"}}

	service := newMockCertificateProviderService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return(certificateProviders, nil)

	err := ChooseCertificateProvider(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathCertificateProviderDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseCertificateProviderWhenServiceError(t *testing.T) {
	form := url.Values{
		"option": {"0"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := newMockCertificateProviderService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return([]donordata.CertificateProvider{{}}, nil)
	service.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ChooseCertificateProvider(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestReadChooseCertificateProviderForm(t *testing.T) {
	form := url.Values{
		"option": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readChooseCertificateProviderForm(r)

	assert.False(t, result.New)
	assert.Equal(t, 1, result.Index)
	assert.Nil(t, result.Err)
}

func TestReadChooseCertificateProviderFormWhenNew(t *testing.T) {
	form := url.Values{
		"option": {"new"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readChooseCertificateProviderForm(r)

	assert.True(t, result.New)
	assert.NotNil(t, result.Err)
}

func TestChooseCertificateProviderFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *chooseCertificateProviderForm
		errors validation.List
	}{
		"new": {
			form: &chooseCertificateProviderForm{New: true, Err: expectedError},
		},
		"index": {
			form: &chooseCertificateProviderForm{Index: 1},
		},
		"error": {
			form:   &chooseCertificateProviderForm{Err: expectedError},
			errors: validation.With("option", validation.SelectError{Label: "aCertificateProviderOrToAddANewCertificateProvider"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
