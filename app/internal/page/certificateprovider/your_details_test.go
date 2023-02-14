package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"

	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCertificateProviderYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
	}

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	sessionStore := &page.MockSessionsStore{}
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{
			"certificate-provider": &sesh.CertificateProviderSession{
				Sub:            "random",
				DonorSessionID: "session-id",
				LpaID:          "lpa-id",
			},
		}}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &cpYourDetailsData{
			App:  page.TestAppData,
			Lpa:  lpa,
			Form: &cpYourDetailsForm{},
		}).
		Return(nil)

	err := YourDetails(template.Func, lpaStore, sessionStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore, sessionStore)
}

func TestGetCertificateProviderYourDetailsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
		CertificateProviderProvidedDetails: actor.CertificateProvider{
			Email:       "a@example.org",
			Mobile:      "07535111222",
			DateOfBirth: date.New("1997", "1", "2"),
		},
	}
	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	sessionStore := &page.MockSessionsStore{}
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{
			"certificate-provider": &sesh.CertificateProviderSession{
				Sub:            "random",
				DonorSessionID: "session-id",
				LpaID:          "lpa-id",
			},
		}}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &cpYourDetailsData{
			App: page.TestAppData,
			Lpa: lpa,
			Form: &cpYourDetailsForm{
				Email:  "a@example.org",
				Mobile: "07535111222",
				Dob:    date.New("1997", "1", "2"),
			},
		}).
		Return(nil)

	err := YourDetails(template.Func, lpaStore, sessionStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore, sessionStore)
}

func TestGetCertificateProviderYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, page.ExpectedError)

	err := YourDetails(nil, lpaStore, nil)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetCertificateProviderYourDetailsWhenSessionStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	sessionStore := &page.MockSessionsStore{}
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{}, page.ExpectedError)

	err := YourDetails(nil, lpaStore, sessionStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, sessionStore)
}

func TestGetCertificateProviderYourDetailsWhenLpaIdMismatch(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)

	sessionStore := &page.MockSessionsStore{}
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{
			"certificate-provider": &sesh.CertificateProviderSession{
				Sub:            "random",
				DonorSessionID: "session-id",
				LpaID:          "not-lpa-id",
			},
		}}, nil)

	err := YourDetails(nil, lpaStore, sessionStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.CertificateProviderStart, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, sessionStore)
}

func TestGetCertificateProviderYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
	}

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	sessionStore := &page.MockSessionsStore{}
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{
			"certificate-provider": &sesh.CertificateProviderSession{
				Sub:            "random",
				DonorSessionID: "session-id",
				LpaID:          "lpa-id",
			},
		}}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &cpYourDetailsData{
			App:  page.TestAppData,
			Lpa:  lpa,
			Form: &cpYourDetailsForm{},
		}).
		Return(page.ExpectedError)

	err := YourDetails(template.Func, lpaStore, sessionStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore, sessionStore)
}

func TestPostCertificateProviderYourDetails(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form url.Values
		cp   actor.CertificateProvider
	}{
		"valid": {
			form: url.Values{
				"email":               {"name@example.com"},
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			cp: actor.CertificateProvider{
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				Email:       "name@example.com",
				Mobile:      "07535111222",
			},
		},
		"warning ignored": {
			form: url.Values{
				"email":               {"name@example.com"},
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
			},
			cp: actor.CertificateProvider{
				DateOfBirth: date.New("1900", "1", "2"),
				Email:       "name@example.com",
				Mobile:      "07535111222",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := &page.MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{ID: "lpa-id"}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					ID:                                 "lpa-id",
					CertificateProviderProvidedDetails: tc.cp,
				}).
				Return(nil)

			sessionStore := &page.MockSessionsStore{}
			sessionStore.
				On("Get", r, "session").
				Return(&sessions.Session{Values: map[any]any{"certificate-provider": &sesh.CertificateProviderSession{Sub: "xyz", LpaID: "lpa-id"}}}, nil)

			err := YourDetails(nil, lpaStore, sessionStore)(page.TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+page.Paths.CertificateProviderYourAddress, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore, sessionStore)
		})
	}
}

func TestPostCertificateProviderYourDetailsWhenInputRequired(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *cpYourDetailsData) bool
	}{
		"validation error": {
			form: url.Values{
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			dataMatcher: func(t *testing.T, data *cpYourDetailsData) bool {
				return assert.Equal(t, validation.With("email", validation.EnterError{Label: "email"}), data.Errors)
			},
		},
		"dob warning": {
			form: url.Values{
				"email":               {"name@example.com"},
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
			},
			dataMatcher: func(t *testing.T, data *cpYourDetailsData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
		"dob warning ignored but other errors": {
			form: url.Values{
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
			},
			dataMatcher: func(t *testing.T, data *cpYourDetailsData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning) &&
					assert.Equal(t, validation.With("email", validation.EnterError{Label: "email"}), data.Errors)
			},
		},
		"other dob warning ignored": {
			form: url.Values{
				"email":               {"name@example.com"},
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsUnder18"},
			},
			dataMatcher: func(t *testing.T, data *cpYourDetailsData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := &page.MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{ID: "lpa-id"}, nil)

			template := &page.MockTemplate{}
			template.
				On("Func", w, mock.MatchedBy(func(data *cpYourDetailsData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			sessionStore := &page.MockSessionsStore{}
			sessionStore.
				On("Get", r, "session").
				Return(&sessions.Session{Values: map[any]any{"certificate-provider": &sesh.CertificateProviderSession{Sub: "xyz", LpaID: "lpa-id"}}}, nil)

			err := YourDetails(template.Func, lpaStore, sessionStore)(page.TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, lpaStore, sessionStore)
		})
	}
}

func TestPostCpYourDetailsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"email":               {"name@example.com"},
		"mobile":              {"07535111222"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1999"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, page.ExpectedError)

	sessionStore := &page.MockSessionsStore{}

	err := YourDetails(nil, lpaStore, sessionStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, sessionStore)
}

func TestPostCpYourDetailsWhenSessionProblem(t *testing.T) {
	testCases := map[string]struct {
		session *sessions.Session
		error   error
	}{
		"store error": {
			session: &sessions.Session{Values: map[any]any{"certificate-provider": &sesh.CertificateProviderSession{Sub: "xyz", LpaID: "lpa-id"}}},
			error:   page.ExpectedError,
		},
		"missing certificate provider session": {
			session: &sessions.Session{},
			error:   nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"email":               {"name@example.com"},
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := &page.MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{}, nil)

			sessionStore := &page.MockSessionsStore{}
			sessionStore.
				On("Get", mock.Anything, "session").
				Return(tc.session, tc.error)

			err := YourDetails(nil, lpaStore, sessionStore)(page.TestAppData, w, r)

			assert.NotNil(t, err)
			mock.AssertExpectationsForObjects(t, lpaStore, sessionStore)
		})
	}
}

func TestReadCpYourDetailsForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"email":               {"name@example.com"},
		"mobile":              {"07535111222"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
		"ignore-dob-warning":  {"xyz"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readCpYourDetailsForm(r)

	assert.Equal("name@example.com", result.Email)
	assert.Equal("07535111222", result.Mobile)
	assert.Equal(date.New("1990", "1", "2"), result.Dob)
	assert.Equal("xyz", result.IgnoreDobWarning)
}

func TestCpYourDetailsFormValidate(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form   *cpYourDetailsForm
		errors validation.List
	}{
		"valid": {
			form: &cpYourDetailsForm{
				Dob:              validDob,
				Mobile:           "07535999222",
				Email:            "name@example.org",
				IgnoreDobWarning: "xyz",
			},
		},
		"missing-all": {
			form: &cpYourDetailsForm{},
			errors: validation.
				With("date-of-birth", validation.EnterError{Label: "dateOfBirth"}).
				With("mobile", validation.EnterError{Label: "mobile"}).
				With("email", validation.EnterError{Label: "email"}),
		},
		"future-dob": {
			form: &cpYourDetailsForm{
				Mobile: "07535999222",
				Email:  "name@example.org",
				Dob:    now.AddDate(0, 0, 1),
			},
			errors: validation.With("date-of-birth", validation.DateMustBePastError{Label: "dateOfBirth"}),
		},
		"invalid-dob": {
			form: &cpYourDetailsForm{
				Mobile: "07535999222",
				Email:  "name@example.org",
				Dob:    date.New("2000", "22", "2"),
			},
			errors: validation.With("date-of-birth", validation.DateMustBeRealError{Label: "dateOfBirth"}),
		},
		"invalid-missing-dob": {
			form: &cpYourDetailsForm{
				Mobile: "07535999222",
				Email:  "name@example.org",
				Dob:    date.New("1", "", "1"),
			},
			errors: validation.With("date-of-birth", validation.DateMissingError{Label: "dateOfBirth", MissingMonth: true}),
		},
		"invalid-mobile-format": {
			form: &cpYourDetailsForm{
				Mobile: "123",
				Email:  "name@example.org",
				Dob:    validDob,
			},
			errors: validation.With("mobile", validation.MobileError{Label: "mobile"}),
		},
		"invalid-email-format": {
			form: &cpYourDetailsForm{
				Mobile: "07535999222",
				Email:  "name@",
				Dob:    validDob,
			},
			errors: validation.With("email", validation.EmailError{Label: "email"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
