package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourDetails(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourDetailsData{
			App:  appData,
			Form: &yourDetailsForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourDetails(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourDetails(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetYourDetailsFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{
			You: Person{
				FirstNames: "John",
			},
		}, nil)

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

	err := YourDetails(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourDetailsData{
			App:  appData,
			Form: &yourDetailsForm{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourDetails(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostYourDetails(t *testing.T) {
	testCases := map[string]struct {
		form   url.Values
		person Person
	}{
		"valid": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {strconv.Itoa(time.Now().Year() - 40)},
			},
			person: Person{
				FirstNames:  "John",
				LastName:    "Doe",
				DateOfBirth: time.Date(time.Now().Year()-40, time.January, 2, 0, 0, 0, 0, time.UTC),
				Address:     place.Address{Line1: "abc"},
			},
		},
		"warning ignored": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-warning":      {"dateOfBirthIsOver100"},
			},
			person: Person{
				FirstNames:  "John",
				LastName:    "Doe",
				DateOfBirth: time.Date(1900, time.January, 2, 0, 0, 0, 0, time.UTC),
				Address:     place.Address{Line1: "abc"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(Lpa{
					You: Person{
						FirstNames: "John",
						Address:    place.Address{Line1: "abc"},
					},
				}, nil)
			lpaStore.
				On("Put", mock.Anything, "session-id", Lpa{
					You: tc.person,
				}).
				Return(nil)

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := YourDetails(nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, yourAddressPath, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostYourDetailsWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *yourDetailsData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, map[string]string{"first-names": "enterFirstNames"}, data.Errors)
			},
		},
		"dob warning": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
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
				"first-names":         {"John"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-warning":      {"dateOfBirthIsOver100"},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
		"other dob warning ignored": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-warning":      {"dateOfBirthIsUnder18"},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(Lpa{}, nil)

			template := &mockTemplate{}
			template.
				On("Func", w, mock.MatchedBy(func(data *yourDetailsData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := YourDetails(template.Func, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, lpaStore, template)
		})
	}
}

func TestPostYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{
			You: Person{
				FirstNames: "John",
				Address:    place.Address{Line1: "abc"},
			},
		}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", Lpa{
			You: Person{
				FirstNames:  "John",
				LastName:    "Doe",
				DateOfBirth: time.Date(1990, time.January, 2, 0, 0, 0, 0, time.UTC),
				Address:     place.Address{Line1: "abc"},
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

	err := YourDetails(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
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
		"ignore-warning":      {"xyz"},
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
	assert.Equal("xyz", result.IgnoreWarning)
}

func TestYourDetailsFormValidate(t *testing.T) {
	now := time.Now().UTC().Round(24 * time.Hour)
	validDob := now.AddDate(-18, 0, -1)

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
				DateOfBirth: validDob,
			},
			errors: map[string]string{},
		},
		"max-length": {
			form: &yourDetailsForm{
				FirstNames: strings.Repeat("x", 53),
				LastName:   strings.Repeat("x", 61),
				OtherNames: strings.Repeat("x", 50),
				Dob: Date{
					Day:   "C",
					Month: "D",
					Year:  "E",
				},
				DateOfBirth: validDob,
			},
			errors: map[string]string{},
		},
		"missing-all": {
			form: &yourDetailsForm{},
			errors: map[string]string{
				"first-names":   "enterFirstNames",
				"last-name":     "enterLastName",
				"date-of-birth": "enterDateOfBirth",
			},
		},
		"too-long": {
			form: &yourDetailsForm{
				FirstNames: strings.Repeat("x", 54),
				LastName:   strings.Repeat("x", 62),
				OtherNames: strings.Repeat("x", 51),
				Dob: Date{
					Day:   "C",
					Month: "D",
					Year:  "E",
				},
				DateOfBirth: validDob,
			},
			errors: map[string]string{
				"first-names": "firstNamesTooLong",
				"last-name":   "lastNameTooLong",
				"other-names": "otherNamesTooLong",
			},
		},
		"future-dob": {
			form: &yourDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Dob: Date{
					Day:   "C",
					Month: "D",
					Year:  "E",
				},
				DateOfBirth: now.AddDate(0, 0, 1),
			},
			errors: map[string]string{
				"date-of-birth": "dateOfBirthIsFuture",
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
				"date-of-birth": "enterDateOfBirth",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}

func TestYourDetailsFormDobWarning(t *testing.T) {
	now := time.Now().UTC().Round(24 * time.Hour)
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form    *yourDetailsForm
		warning string
	}{
		"valid": {
			form: &yourDetailsForm{
				DateOfBirth: validDob,
			},
		},
		"future-dob": {
			form: &yourDetailsForm{
				DateOfBirth: now.AddDate(0, 0, 1),
			},
		},
		"dob-is-18": {
			form: &yourDetailsForm{
				DateOfBirth: now.AddDate(-18, 0, 0),
			},
		},
		"dob-under-18": {
			form: &yourDetailsForm{
				DateOfBirth: now.AddDate(-18, 0, 1),
			},
			warning: "dateOfBirthIsUnder18",
		},
		"dob-is-100": {
			form: &yourDetailsForm{
				DateOfBirth: now.AddDate(-100, 0, 0),
			},
		},
		"dob-over-100": {
			form: &yourDetailsForm{
				DateOfBirth: now.AddDate(-100, 0, -1),
			},
			warning: "dateOfBirthIsOver100",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.warning, tc.form.DobWarning())
		})
	}
}
