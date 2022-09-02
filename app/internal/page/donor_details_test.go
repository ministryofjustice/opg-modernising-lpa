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

func TestGetDonorDetails(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &donorDetailsData{
			App:  appData,
			Form: &donorDetailsForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	DonorDetails(nil, template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetDonorDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, &donorDetailsData{
			App:  appData,
			Form: &donorDetailsForm{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	DonorDetails(logger, template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, logger)
}

func TestPostDonorDetails(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{
			Donor: Donor{FirstNames: "John", LastName: "Doe", DateOfBirth: time.Date(1990, time.January, 2, 0, 0, 0, 0, time.UTC)},
		}).
		Return(nil)

	form := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	DonorDetails(nil, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/donor-address", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostDonorDetailsWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *donorDetailsData) bool {
			return assert.Equal(t, map[string]string{"first-names": "enterFirstNames"}, data.Errors)
		})).
		Return(nil)

	form := url.Values{
		"last-name":           {"Doe"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	DonorDetails(nil, template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadDonorDetailsForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"first-names":         {"  John "},
		"last-name":           {"Doe"},
		"other-names":         {"Somebody"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readDonorDetailsForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("Somebody", result.OtherNames)
	assert.Equal("2", result.DobDay)
	assert.Equal("1", result.DobMonth)
	assert.Equal("1990", result.DobYear)
	assert.Equal(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC), result.DateOfBirth)
	assert.Nil(result.DateOfBirthError)
}

func TestDonorDetailsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *donorDetailsForm
		errors map[string]string
	}{
		"valid": {
			form: &donorDetailsForm{
				FirstNames:  "A",
				LastName:    "B",
				DobDay:      "C",
				DobMonth:    "D",
				DobYear:     "E",
				DateOfBirth: time.Now(),
			},
			errors: map[string]string{},
		},
		"missing-all": {
			form: &donorDetailsForm{},
			errors: map[string]string{
				"first-names":   "enterFirstNames",
				"last-name":     "enterLastName",
				"date-of-birth": "dateOfBirthYear",
			},
		},
		"invalid-dob": {
			form: &donorDetailsForm{
				FirstNames:       "A",
				LastName:         "B",
				DobDay:           "1",
				DobMonth:         "1",
				DobYear:          "1",
				DateOfBirthError: expectedError,
			},
			errors: map[string]string{
				"date-of-birth": "dateOfBirthMustBeReal",
			},
		},
		"invalid-missing-dob": {
			form: &donorDetailsForm{
				FirstNames:       "A",
				LastName:         "B",
				DobDay:           "1",
				DobYear:          "1",
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
