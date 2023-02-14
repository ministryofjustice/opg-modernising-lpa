package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowLongHaveYouKnownCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App: page.TestAppData,
		}).
		Return(nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowLongHaveYouKnownCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := actor.CertificateProvider{RelationshipLength: "gte-2-years"}

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{CertificateProvider: certificateProvider}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App:                 page.TestAppData,
			CertificateProvider: certificateProvider,
			HowLong:             "gte-2-years",
		}).
		Return(nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowLongHaveYouKnownCertificateProviderWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, page.ExpectedError)

	err := HowLongHaveYouKnownCertificateProvider(nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetHowLongHaveYouKnownCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App: page.TestAppData,
		}).
		Return(page.ExpectedError)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowLongHaveYouKnownCertificateProvider(t *testing.T) {
	form := url.Values{
		"how-long": {"gte-2-years"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Attorneys:                 actor.Attorneys{{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: date.New("1990", "1", "1")}},
			HowAttorneysMakeDecisions: page.Jointly,
			Tasks:                     page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			Attorneys:                 actor.Attorneys{{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: date.New("1990", "1", "1")}},
			HowAttorneysMakeDecisions: page.Jointly,
			CertificateProvider:       actor.CertificateProvider{RelationshipLength: "gte-2-years"},
			Tasks:                     page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted, CertificateProvider: page.TaskCompleted},
		}).
		Return(nil)

	err := HowLongHaveYouKnownCertificateProvider(nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.DoYouWantToNotifyPeople, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowLongHaveYouKnownCertificateProviderWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"how-long": {"gte-2-years"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(page.ExpectedError)

	err := HowLongHaveYouKnownCertificateProvider(nil, lpaStore)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowLongHaveYouKnownCertificateProviderWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App:    page.TestAppData,
			Errors: validation.With("how-long", validation.SelectError{Label: "howLongYouHaveKnownCertificateProvider"}),
		}).
		Return(nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, lpaStore)(page.TestAppData, w, r)
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
	r.Header.Add("Content-Type", page.FormUrlEncoded)

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
