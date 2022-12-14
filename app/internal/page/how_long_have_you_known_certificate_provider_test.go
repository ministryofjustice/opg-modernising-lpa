package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowLongHaveYouKnownCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App: appData,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowLongHaveYouKnownCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	certificateProvider := CertificateProvider{RelationshipLength: "gte-2-years"}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{CertificateProvider: certificateProvider}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App:                 appData,
			CertificateProvider: certificateProvider,
			HowLong:             "gte-2-years",
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowLongHaveYouKnownCertificateProviderWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowLongHaveYouKnownCertificateProvider(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetHowLongHaveYouKnownCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App: appData,
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowLongHaveYouKnownCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			Attorneys:                 []Attorney{{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)}},
			HowAttorneysMakeDecisions: Jointly,
		}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			Attorneys:                 []Attorney{{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)}},
			HowAttorneysMakeDecisions: Jointly,
			CertificateProvider:       CertificateProvider{RelationshipLength: "gte-2-years"},
			Tasks:                     Tasks{CertificateProvider: TaskCompleted},
		}).
		Return(nil)

	form := url.Values{
		"how-long": {"gte-2-years"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowLongHaveYouKnownCertificateProvider(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.CheckYourLpa, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowLongHaveYouKnownCertificateProviderWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", mock.Anything).
		Return(expectedError)

	form := url.Values{
		"how-long": {"gte-2-years"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowLongHaveYouKnownCertificateProvider(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowLongHaveYouKnownCertificateProviderWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App: appData,
			Errors: map[string]string{
				"how-long": "selectHowLongHaveYouKnownCertificateProvider",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadHowLongHaveYouKnownCertificateProviderForm(t *testing.T) {
	form := url.Values{
		"how-long": {"gte-2-years"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readHowLongHaveYouKnownCertificateProviderForm(r)

	assert.Equal(t, "gte-2-years", result.HowLong)
}

func TestHowLongHaveYouKnownCertificateProviderFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howLongHaveYouKnownCertificateProviderForm
		errors map[string]string
	}{
		"gte-2-years": {
			form: &howLongHaveYouKnownCertificateProviderForm{
				HowLong: "gte-2-years",
			},
			errors: map[string]string{},
		},
		"lt-2-years": {
			form: &howLongHaveYouKnownCertificateProviderForm{
				HowLong: "lt-2-years",
			},
			errors: map[string]string{
				"how-long": "mustHaveKnownCertificateProviderTwoYears",
			},
		},
		"missing": {
			form: &howLongHaveYouKnownCertificateProviderForm{},
			errors: map[string]string{
				"how-long": "selectHowLongHaveYouKnownCertificateProvider",
			},
		},
		"invalid": {
			form: &howLongHaveYouKnownCertificateProviderForm{
				HowLong: "what",
			},
			errors: map[string]string{
				"how-long": "selectHowLongHaveYouKnownCertificateProvider",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
