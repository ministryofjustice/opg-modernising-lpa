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

func TestGetYourDetails(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourDetailsData{
			App:  appData,
			Form: &yourDetailsForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourDetails(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestGetYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourDetails(nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestGetYourDetailsFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{
		data: Lpa{
			You: Person{
				FirstNames: "John",
			},
		},
	}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourDetailsData{
			App: appData,
			Form: &yourDetailsForm{
				FirstNames: "John",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourDetails(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestGetYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourDetailsData{
			App:  appData,
			Form: &yourDetailsForm{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourDetails(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestPostYourDetails(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{
		data: Lpa{
			You: Person{
				FirstNames: "John",
				Address:    Address{Line1: "abc"},
			},
		},
	}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{
			You: Person{
				FirstNames:  "John",
				LastName:    "Doe",
				DateOfBirth: time.Date(1990, time.January, 2, 0, 0, 0, 0, time.UTC),
				Address:     Address{Line1: "abc"},
			},
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

	err := YourDetails(nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, yourAddressPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{
		data: Lpa{
			You: Person{
				FirstNames: "John",
				Address:    Address{Line1: "abc"},
			},
		},
	}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{
			You: Person{
				FirstNames:  "John",
				LastName:    "Doe",
				DateOfBirth: time.Date(1990, time.January, 2, 0, 0, 0, 0, time.UTC),
				Address:     Address{Line1: "abc"},
			},
		}).
		Return(expectedError)

	form := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := YourDetails(nil, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostYourDetailsWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *yourDetailsData) bool {
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

	err := YourDetails(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadYourDetailsForm(t *testing.T) {
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

	result := readYourDetailsForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("Somebody", result.OtherNames)
	assert.Equal("2", result.Dob.Day)
	assert.Equal("1", result.Dob.Month)
	assert.Equal("1990", result.Dob.Year)
	assert.Equal(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC), result.DateOfBirth)
	assert.Nil(result.DateOfBirthError)
}

func TestYourDetailsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *yourDetailsForm
		errors map[string]string
	}{
		"valid": {
			form: &yourDetailsForm{
				FirstNames: "A",
				LastName:   "B",
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
			form: &yourDetailsForm{},
			errors: map[string]string{
				"first-names":   "enterFirstNames",
				"last-name":     "enterLastName",
				"date-of-birth": "dateOfBirthYear",
			},
		},
		"invalid-dob": {
			form: &yourDetailsForm{
				FirstNames: "A",
				LastName:   "B",
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
			form: &yourDetailsForm{
				FirstNames: "A",
				LastName:   "B",
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
