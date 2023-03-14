package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCertificateProviderYourDetails(t *testing.T) {
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
		On("Execute", w, &yourDetailsData{
			App:  testAppData,
			Lpa:  lpa,
			Form: &yourDetailsForm{},
		}).
		Return(nil)

	err := YourDetails(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCertificateProviderYourDetailsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
		CertificateProviderProvidedDetails: actor.CertificateProvider{
			Mobile:      "07535111222",
			DateOfBirth: date.New("1997", "1", "2"),
		},
	}
	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourDetailsData{
			App: testAppData,
			Lpa: lpa,
			Form: &yourDetailsForm{
				Mobile: "07535111222",
				Dob:    date.New("1997", "1", "2"),
			},
		}).
		Return(nil)

	err := YourDetails(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCertificateProviderYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := YourDetails(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCertificateProviderYourDetailsWhenTemplateErrors(t *testing.T) {
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
		On("Execute", w, &yourDetailsData{
			App:  testAppData,
			Lpa:  lpa,
			Form: &yourDetailsForm{},
		}).
		Return(expectedError)

	err := YourDetails(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCertificateProviderYourDetails(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form url.Values
		cp   actor.CertificateProvider
	}{
		"valid": {
			form: url.Values{
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			cp: actor.CertificateProvider{
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				Mobile:      "07535111222",
			},
		},
		"warning ignored": {
			form: url.Values{
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
			},
			cp: actor.CertificateProvider{
				DateOfBirth: date.New("1900", "1", "2"),
				Mobile:      "07535111222",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{ID: "lpa-id"}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					ID:                                 "lpa-id",
					CertificateProviderProvidedDetails: tc.cp,
				}).
				Return(nil)

			err := YourDetails(nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.CertificateProviderYourAddress, resp.Header.Get("Location"))
		})
	}
}

func TestPostCertificateProviderYourDetailsWhenInputRequired(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *yourDetailsData) bool
	}{
		"validation error": {
			form: url.Values{
				"mobile":              {"0123456"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, validation.With("mobile", validation.EnterError{Label: "aValidUkMobileLike"}), data.Errors)
			},
		},
		"dob warning": {
			form: url.Values{
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
		"dob warning ignored but other errors": {
			form: url.Values{
				"mobile":              {"0123456"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning) &&
					assert.Equal(t, validation.With("mobile", validation.EnterError{Label: "aValidUkMobileLike"}), data.Errors)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{ID: "lpa-id"}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, mock.MatchedBy(func(data *yourDetailsData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := YourDetails(template.Execute, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostYourDetailsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"mobile":              {"07535111222"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1999"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := YourDetails(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadYourDetailsForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"mobile":              {"07535111222"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
		"ignore-dob-warning":  {"xyz"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readYourDetailsForm(r)

	assert.Equal("07535111222", result.Mobile)
	assert.Equal(date.New("1990", "1", "2"), result.Dob)
	assert.Equal("xyz", result.IgnoreDobWarning)
}

func TestYourDetailsFormValidate(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form   *yourDetailsForm
		errors validation.List
	}{
		"valid": {
			form: &yourDetailsForm{
				Dob:              validDob,
				Mobile:           "07535999222",
				IgnoreDobWarning: "xyz",
			},
		},
		"missing-all": {
			form: &yourDetailsForm{},
			errors: validation.
				With("date-of-birth", validation.EnterError{Label: "yourDateOfBirth"}).
				With("mobile", validation.EnterError{Label: "mobile"}),
		},
		"future-dob": {
			form: &yourDetailsForm{
				Mobile: "07535999222",
				Dob:    now.AddDate(0, 0, 1),
			},
			errors: validation.With("date-of-birth", validation.DateMustBePastError{Label: "yourDateOfBirth"}),
		},
		"dob-under-18": {
			form: &yourDetailsForm{
				Mobile: "07535999222",
				Dob:    now.AddDate(0, 0, -1),
			},
			errors: validation.With("date-of-birth", validation.CustomError{Label: "youAreUnder18Error"}),
		},
		"invalid-dob": {
			form: &yourDetailsForm{
				Mobile: "07535999222",
				Dob:    date.New("2000", "22", "2"),
			},
			errors: validation.With("date-of-birth", validation.DateMustBeRealError{Label: "aDateOfBirth"}),
		},
		"invalid-missing-dob": {
			form: &yourDetailsForm{
				Mobile: "07535999222",
				Dob:    date.New("1", "", "1"),
			},
			errors: validation.With("date-of-birth", validation.DateMissingError{Label: "yourDateOfBirth", MissingMonth: true}),
		},
		"invalid-mobile-format": {
			form: &yourDetailsForm{
				Mobile: "123",
				Dob:    validDob,
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
