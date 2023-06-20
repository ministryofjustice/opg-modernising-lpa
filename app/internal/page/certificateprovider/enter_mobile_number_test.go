package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterMobileNumber(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(lpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &mobileNumberData{
			App:  testAppData,
			Lpa:  lpa,
			Form: &mobileNumberForm{},
		}).
		Return(nil)

	err := EnterMobileNumber(template.Execute, donorStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterMobileNumberFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(lpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{Mobile: "07535111222"}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &mobileNumberData{
			App: testAppData,
			Lpa: lpa,
			Form: &mobileNumberForm{
				Mobile: "07535111222",
			},
		}).
		Return(nil)

	err := EnterMobileNumber(template.Execute, donorStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterMobileNumberWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := EnterMobileNumber(nil, donorStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterMobileNumberWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := EnterMobileNumber(nil, donorStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterMobileNumberWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(lpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &mobileNumberData{
			App:  testAppData,
			Lpa:  lpa,
			Form: &mobileNumberForm{},
		}).
		Return(expectedError)

	err := EnterMobileNumber(template.Execute, donorStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterMobileNumber(t *testing.T) {
	w := httptest.NewRecorder()
	form := url.Values{
		"mobile": {"07535111222"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)
	certificateProviderStore.
		On("Put", r.Context(), &actor.CertificateProviderProvidedDetails{LpaID: "lpa-id", Mobile: "07535111222"}).
		Return(nil)

	err := EnterMobileNumber(nil, donorStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.WhatYoullNeedToConfirmYourIdentity.Format("lpa-id"), resp.Header.Get("Location"))

}

func TestPostEnterMobileNumberWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	form := url.Values{
		"mobile": {"0123456"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	dataMatcher := func(t *testing.T, data *mobileNumberData) bool {
		return assert.Equal(t, validation.With("mobile", validation.MobileError{Label: "mobile"}), data.Errors)
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *mobileNumberData) bool {
			return dataMatcher(t, data)
		})).
		Return(nil)

	err := EnterMobileNumber(template.Execute, donorStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterMobileNumberWhenCertificateProviderStoreErrors(t *testing.T) {
	form := url.Values{
		"mobile": {"07535111222"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)
	certificateProviderStore.
		On("Put", r.Context(), &actor.CertificateProviderProvidedDetails{Mobile: "07535111222"}).
		Return(expectedError)

	err := EnterMobileNumber(nil, donorStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadMobileNumberForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"mobile": {"07535111222"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readMobileNumberForm(r)

	assert.Equal("07535111222", result.Mobile)
}

func TestMobileNumberFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *mobileNumberForm
		errors validation.List
	}{
		"valid": {
			form: &mobileNumberForm{
				Mobile: "07535999222",
			},
		},
		"missing": {
			form: &mobileNumberForm{},
			errors: validation.
				With("mobile", validation.EnterError{Label: "mobile"}),
		},
		"invalid-mobile-format": {
			form: &mobileNumberForm{
				Mobile: "123",
			},
			errors: validation.With("mobile", validation.MobileError{Label: "mobile"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
