package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterMobileNumber(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &mobileNumberData{
			App:  testAppData,
			Lpa:  lpa,
			Form: &mobileNumberForm{},
		}).
		Return(nil)

	err := EnterMobileNumber(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterMobileNumberFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
		CertificateProviderProvidedDetails: actor.CertificateProvider{
			Mobile: "07535111222",
		},
	}
	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

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

	err := EnterMobileNumber(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterMobileNumberWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := EnterMobileNumber(nil, lpaStore)(testAppData, w, r)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &mobileNumberData{
			App:  testAppData,
			Lpa:  lpa,
			Form: &mobileNumberForm{},
		}).
		Return(expectedError)

	err := EnterMobileNumber(template.Execute, lpaStore)(testAppData, w, r)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			ID: "lpa-id",
			CertificateProviderProvidedDetails: actor.CertificateProvider{
				Mobile: "07535111222",
			},
		}).
		Return(nil)

	err := EnterMobileNumber(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderYourAddress, resp.Header.Get("Location"))

}

func TestPostEnterMobileNumberWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	form := url.Values{
		"mobile": {"0123456"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)

	dataMatcher := func(t *testing.T, data *mobileNumberData) bool {
		return assert.Equal(t, validation.With("mobile", validation.EnterError{Label: "aValidUkMobileLike"}), data.Errors)
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *mobileNumberData) bool {
			return dataMatcher(t, data)
		})).
		Return(nil)

	err := EnterMobileNumber(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterMobileNumberWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"mobile": {"07535111222"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := EnterMobileNumber(nil, lpaStore)(testAppData, w, r)
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
				With("mobile", validation.EnterError{Label: "yourUkMobile"}),
		},
		"invalid-mobile-format": {
			form: &mobileNumberForm{
				Mobile: "123",
			},
			errors: validation.With("mobile", validation.EnterError{Label: "aValidUkMobileLike"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
