package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourDetailsData{
			App:  appData,
			Form: &yourDetailsForm{},
		}).
		Return(nil)

	err := YourDetails(template.Func, lpaStore, nil)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := YourDetails(nil, lpaStore, nil)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetYourDetailsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
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

	err := YourDetails(template.Func, lpaStore, nil)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourDetailsData{
			App:  appData,
			Form: &yourDetailsForm{},
		}).
		Return(expectedError)

	err := YourDetails(template.Func, lpaStore, nil)(appData, w, r)
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
				Email:       "name@example.com",
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
				Email:       "name@example.com",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&Lpa{
					You: Person{
						FirstNames: "John",
						Address:    place.Address{Line1: "abc"},
					},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &Lpa{
					You:   tc.person,
					Tasks: Tasks{YourDetails: TaskInProgress},
				}).
				Return(nil)

			sessionStore := &mockSessionsStore{}
			sessionStore.
				On("Get", r, "session").
				Return(&sessions.Session{Values: map[interface{}]interface{}{"email": "name@example.com"}}, nil)

			err := YourDetails(nil, lpaStore, sessionStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+Paths.YourAddress, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore, sessionStore)
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
				return assert.Equal(t, validation.With("first-names", "enterFirstNames"), data.Errors)
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
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			template := &mockTemplate{}
			template.
				On("Func", w, mock.MatchedBy(func(data *yourDetailsData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&Lpa{}, nil)

			sessionStore := &mockSessionsStore{}
			sessionStore.
				On("Get", mock.Anything, "session").
				Return(&sessions.Session{Values: map[interface{}]interface{}{"email": "name@example.com"}}, nil)

			err := YourDetails(template.Func, lpaStore, sessionStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, template, lpaStore, sessionStore)
		})
	}
}

func TestPostYourDetailsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			You: Person{
				FirstNames: "John",
				Address:    place.Address{Line1: "abc"},
			},
		}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "session").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"email": "name@example.com"}}, nil)

	err := YourDetails(nil, lpaStore, sessionStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, sessionStore)
}

func TestPostYourDetailsWhenSessionProblem(t *testing.T) {
	testCases := map[string]struct {
		session *sessions.Session
		error   error
	}{
		"store error": {
			session: &sessions.Session{Values: map[interface{}]interface{}{"email": "name@example.com"}},
			error:   expectedError,
		},
		"missing email": {
			session: &sessions.Session{},
			error:   nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&Lpa{}, nil)

			sessionStore := &mockSessionsStore{}
			sessionStore.
				On("Get", mock.Anything, "session").
				Return(tc.session, tc.error)

			err := YourDetails(nil, lpaStore, sessionStore)(appData, w, r)

			assert.NotNil(t, err)
			mock.AssertExpectationsForObjects(t, lpaStore, sessionStore)
		})
	}
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
		errors validation.List
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
		},
		"missing-all": {
			form: &yourDetailsForm{},
			errors: validation.
				With("first-names", "enterFirstNames").
				With("last-name", "enterLastName").
				With("date-of-birth", "enterDateOfBirth"),
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
			errors: validation.
				With("first-names", "firstNamesTooLong").
				With("last-name", "lastNameTooLong").
				With("other-names", "otherNamesTooLong"),
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
			errors: validation.With("date-of-birth", "dateOfBirthIsFuture"),
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
			errors: validation.With("date-of-birth", "dateOfBirthMustBeReal"),
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
			errors: validation.With("date-of-birth", "enterDateOfBirth"),
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
