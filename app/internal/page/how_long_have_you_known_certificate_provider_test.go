package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowLongHaveYouKnownCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App: appData,
		}).
		Return(nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowLongHaveYouKnownCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := actor.CertificateProvider{RelationshipLength: "gte-2-years"}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{CertificateProvider: certificateProvider}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App:                 appData,
			CertificateProvider: certificateProvider,
			HowLong:             "gte-2-years",
		}).
		Return(nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowLongHaveYouKnownCertificateProviderWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := HowLongHaveYouKnownCertificateProvider(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetHowLongHaveYouKnownCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App: appData,
		}).
		Return(expectedError)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowLongHaveYouKnownCertificateProvider(t *testing.T) {
	form := url.Values{
		"how-long": {"gte-2-years"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			Attorneys:                 actor.Attorneys{{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: date.New("1990", "1", "1")}},
			HowAttorneysMakeDecisions: Jointly,
			Tasks:                     Tasks{YourDetails: TaskCompleted, ChooseAttorneys: TaskCompleted},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{
			Attorneys:                 actor.Attorneys{{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: date.New("1990", "1", "1")}},
			HowAttorneysMakeDecisions: Jointly,
			CertificateProvider:       actor.CertificateProvider{RelationshipLength: "gte-2-years"},
			Tasks:                     Tasks{YourDetails: TaskCompleted, ChooseAttorneys: TaskCompleted, CertificateProvider: TaskCompleted},
		}).
		Return(nil)

	err := HowLongHaveYouKnownCertificateProvider(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.DoYouWantToNotifyPeople, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowLongHaveYouKnownCertificateProviderWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"how-long": {"gte-2-years"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := HowLongHaveYouKnownCertificateProvider(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowLongHaveYouKnownCertificateProviderWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App:    appData,
			Errors: validation.With("how-long", validation.SelectError{Label: "howLongYouHaveKnownCertificateProvider"}),
		}).
		Return(nil)

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
		errors validation.List
	}{
		"gte-2-years": {
			form: &howLongHaveYouKnownCertificateProviderForm{
				HowLong: "gte-2-years",
			},
		},
		"lt-2-years": {
			form: &howLongHaveYouKnownCertificateProviderForm{
				HowLong: "lt-2-years",
			},
			errors: validation.With("how-long", validation.CustomError{Label: "mustHaveKnownCertificateProviderTwoYears"}),
		},
		"missing": {
			form:   &howLongHaveYouKnownCertificateProviderForm{},
			errors: validation.With("how-long", validation.SelectError{Label: "howLongYouHaveKnownCertificateProvider"}),
		},
		"invalid": {
			form: &howLongHaveYouKnownCertificateProviderForm{
				HowLong: "what",
			},
			errors: validation.With("how-long", validation.SelectError{Label: "howLongYouHaveKnownCertificateProvider"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
