package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCertificateProviderDetails(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &certificateProviderDetailsData{
			App:  appData,
			Form: &certificateProviderDetailsForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := CertificateProviderDetails(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetCertificateProviderDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := CertificateProviderDetails(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetCertificateProviderDetailsFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{
			CertificateProvider: CertificateProvider{
				FirstNames: "John",
			},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &certificateProviderDetailsData{
			App: appData,
			Form: &certificateProviderDetailsForm{
				FirstNames: "John",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := CertificateProviderDetails(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetCertificateProviderDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &certificateProviderDetailsData{
			App:  appData,
			Form: &certificateProviderDetailsForm{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := CertificateProviderDetails(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostCertificateProviderDetails(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{
			CertificateProvider: CertificateProvider{
				FirstNames: "John",
			},
		}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", Lpa{
			CertificateProvider: CertificateProvider{
				FirstNames:  "John",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: time.Date(1990, time.January, 2, 0, 0, 0, 0, time.UTC),
			},
		}).
		Return(nil)

	form := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"email":               {"john@example.com"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := CertificateProviderDetails(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, howDoYouKnowYourCertificateProviderPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostCertificateProviderDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", mock.Anything).
		Return(expectedError)

	form := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"email":               {"john@example.com"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := CertificateProviderDetails(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostCertificateProviderDetailsWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *certificateProviderDetailsData) bool {
			return assert.Equal(t, map[string]string{"first-names": "enterFirstNames"}, data.Errors)
		})).
		Return(nil)

	form := url.Values{
		"last-name":           {"Doe"},
		"email":               {"john@example.com"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := CertificateProviderDetails(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadCertificateProviderDetailsForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"first-names":         {"  John "},
		"last-name":           {"Doe"},
		"email":               {"john@example.com"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readCertificateProviderDetailsForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("john@example.com", result.Email)
	assert.Equal("2", result.Dob.Day)
	assert.Equal("1", result.Dob.Month)
	assert.Equal("1990", result.Dob.Year)
	assert.Equal(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC), result.DateOfBirth)
	assert.Nil(result.DateOfBirthError)
}

func TestCertificateProviderDetailsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *certificateProviderDetailsForm
		errors map[string]string
	}{
		"valid": {
			form: &certificateProviderDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "H",
				Dob: Date{
					Day:   "C",
					Month: "D",
					Year:  "E",
				},
				DateOfBirth: time.Now(),
			},
			errors: map[string]string{},
		},
		"missing-all": {
			form: &certificateProviderDetailsForm{},
			errors: map[string]string{
				"first-names":   "enterFirstNames",
				"last-name":     "enterLastName",
				"date-of-birth": "dateOfBirthYear",
				"email":         "enterEmail",
			},
		},
		"invalid-dob": {
			form: &certificateProviderDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "C",
				Dob: Date{
					Day:   "1",
					Month: "1",
					Year:  "1",
				},
				DateOfBirthError: expectedError,
			},
			errors: map[string]string{
				"date-of-birth": "dateOfBirthMustBeReal",
			},
		},
		"invalid-missing-dob": {
			form: &certificateProviderDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "C",
				Dob: Date{
					Day:  "1",
					Year: "1",
				},
				DateOfBirthError: expectedError,
			},
			errors: map[string]string{
				"date-of-birth": "dateOfBirthMonth",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
