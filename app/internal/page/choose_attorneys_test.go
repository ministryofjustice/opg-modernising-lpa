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

func TestGetChooseAttorneys(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysData{
			App:  appData,
			Form: &chooseAttorneysForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneys(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetChooseAttorneysWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneys(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChooseAttorneysFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{
			Attorney: Attorney{
				FirstNames: "John",
			},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysData{
			App: appData,
			Form: &chooseAttorneysForm{
				FirstNames: "John",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneys(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetChooseAttorneysWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysData{
			App:  appData,
			Form: &chooseAttorneysForm{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneys(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostChooseAttorneys(t *testing.T) {
	testCases := map[string]struct {
		form     url.Values
		attorney Attorney
	}{
		"valid": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"email":               {"john@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {strconv.Itoa(time.Now().Year() - 40)},
			},
			attorney: Attorney{
				FirstNames:  "John",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: time.Date(time.Now().Year()-40, time.January, 2, 0, 0, 0, 0, time.UTC),
				Address:     place.Address{Line1: "abc"},
			},
		},
		"warning ignored": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"email":               {"john@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-warning":      {"dateOfBirthIsOver100"},
			},
			attorney: Attorney{
				FirstNames:  "John",
				LastName:    "Doe",
				Email:       "john@example.com",
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
					Attorney: Attorney{
						FirstNames: "John",
						Address:    place.Address{Line1: "abc"},
					},
				}, nil)
			lpaStore.
				On("Put", mock.Anything, "session-id", Lpa{
					Attorney: tc.attorney,
				}).
				Return(nil)

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := ChooseAttorneys(nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/choose-attorneys-address", resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostChooseAttorneysWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *chooseAttorneysData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name":           {"Doe"},
				"email":               {"name@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
			},
			dataMatcher: func(t *testing.T, data *chooseAttorneysData) bool {
				return assert.Equal(t, map[string]string{"first-names": "enterFirstNames"}, data.Errors)
			},
		},
		"dob warning": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"email":               {"name@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
			},
			dataMatcher: func(t *testing.T, data *chooseAttorneysData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
		"dob warning ignored but other errors": {
			form: url.Values{
				"first-names":         {"John"},
				"email":               {"name@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-warning":      {"dateOfBirthIsOver100"},
			},
			dataMatcher: func(t *testing.T, data *chooseAttorneysData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
		"other dob warning ignored": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"email":               {"name@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-warning":      {"dateOfBirthIsUnder18"},
			},
			dataMatcher: func(t *testing.T, data *chooseAttorneysData) bool {
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
				On("Func", w, mock.MatchedBy(func(data *chooseAttorneysData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := ChooseAttorneys(template.Func, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, lpaStore, template)
		})
	}
}

func TestPostChooseAttorneysWhenStoreErrors(t *testing.T) {
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

	err := ChooseAttorneys(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestReadChooseAttorneysForm(t *testing.T) {
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

	result := readChooseAttorneysForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("john@example.com", result.Email)
	assert.Equal("2", result.Dob.Day)
	assert.Equal("1", result.Dob.Month)
	assert.Equal("1990", result.Dob.Year)
	assert.Equal(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC), result.DateOfBirth)
	assert.Nil(result.DateOfBirthError)
}

func TestChooseAttorneysFormValidate(t *testing.T) {
	now := time.Now().UTC().Round(24 * time.Hour)
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form   *chooseAttorneysForm
		errors map[string]string
	}{
		"valid": {
			form: &chooseAttorneysForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "person@example.com",
				Dob: Date{
					Day:   "C",
					Month: "D",
					Year:  "E",
				},
				DateOfBirth: validDob,
			},
			errors: map[string]string{},
		},
		"max length": {
			form: &chooseAttorneysForm{
				FirstNames: strings.Repeat("x", 53),
				LastName:   strings.Repeat("x", 61),
				Email:      "person@example.com",
				Dob: Date{
					Day:   "C",
					Month: "D",
					Year:  "E",
				},
				DateOfBirth: validDob,
			},
			errors: map[string]string{},
		},
		"missing all": {
			form: &chooseAttorneysForm{},
			errors: map[string]string{
				"first-names":   "enterFirstNames",
				"last-name":     "enterLastName",
				"date-of-birth": "enterDateOfBirth",
				"email":         "enterEmail",
			},
		},
		"too long": {
			form: &chooseAttorneysForm{
				FirstNames: strings.Repeat("x", 54),
				LastName:   strings.Repeat("x", 62),
				Email:      "person@example.com",
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
			},
		},
		"future dob": {
			form: &chooseAttorneysForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "person@example.com",
				Dob: Date{
					Day:   "1",
					Month: "1",
					Year:  "1",
				},
				DateOfBirth: now.AddDate(0, 0, 1),
			},
			errors: map[string]string{
				"date-of-birth": "dateOfBirthIsFuture",
			},
		},
		"invalid-dob": {
			form: &chooseAttorneysForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "person@example.com",
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
			form: &chooseAttorneysForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "person@example.com",
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
func TestChooseAttorneysFormDobWarning(t *testing.T) {
	now := time.Now().UTC().Round(24 * time.Hour)
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form    *chooseAttorneysForm
		warning string
	}{
		"valid": {
			form: &chooseAttorneysForm{
				DateOfBirth: validDob,
			},
		},
		"future-dob": {
			form: &chooseAttorneysForm{
				DateOfBirth: now.AddDate(0, 0, 1),
			},
		},
		"dob-is-18": {
			form: &chooseAttorneysForm{
				DateOfBirth: now.AddDate(-18, 0, 0),
			},
		},
		"dob-under-18": {
			form: &chooseAttorneysForm{
				DateOfBirth: now.AddDate(-18, 0, 1),
			},
			warning: "dateOfBirthIsUnder18",
		},
		"dob-is-100": {
			form: &chooseAttorneysForm{
				DateOfBirth: now.AddDate(-100, 0, 0),
			},
		},
		"dob-over-100": {
			form: &chooseAttorneysForm{
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
