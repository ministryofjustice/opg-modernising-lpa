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

func TestGetChooseReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysData{
			App:  appData,
			Form: &chooseAttorneysForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseReplacementAttorneys(template.Func, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetChooseReplacementAttorneysWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseReplacementAttorneys(nil, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChooseReplacementAttorneysFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			ReplacementAttorneys: []Attorney{
				{FirstNames: "John", ID: "1"},
			},
		}, nil)

	template := &mockTemplate{}

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseReplacementAttorneys(template.Func, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/choose-replacement-attorneys-summary", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChooseReplacementAttorneysWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysData{
			App:  appData,
			Form: &chooseAttorneysForm{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseReplacementAttorneys(template.Func, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostChooseReplacementAttorneysAttorneyDoesNotExists(t *testing.T) {
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
				ID:          "123",
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
				ID:          "123",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{}, nil)
			lpaStore.
				On("Put", mock.Anything, "session-id", &Lpa{
					ReplacementAttorneys: []Attorney{tc.attorney},
				}).
				Return(nil)

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := ChooseReplacementAttorneys(nil, lpaStore, mockRandom)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/choose-replacement-attorneys-address?id=123", resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostChooseReplacementAttorneysAttorneyExists(t *testing.T) {
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
				ID:          "123",
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
				ID:          "123",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{
					ReplacementAttorneys: []Attorney{
						{
							FirstNames: "John",
							ID:         "123",
							Address:    place.Address{Line1: "abc"},
						},
					},
				}, nil)
			lpaStore.
				On("Put", mock.Anything, "session-id", &Lpa{
					ReplacementAttorneys: []Attorney{tc.attorney},
				}).
				Return(nil)

			r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := ChooseReplacementAttorneys(nil, lpaStore, mockRandom)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/choose-replacement-attorneys-address?id=123", resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostChooseReplacementAttorneysFromAnotherPage(t *testing.T) {
	testcases := map[string]struct {
		requestUrl      string
		expectedNextUrl string
	}{
		"with from value": {
			"/?from=/test&id=123",
			"/test",
		},
		"without from value": {
			"/?from=&id=123",
			"/choose-replacement-attorneys-address?id=123",
		},
		"missing from key": {
			"/?id=123",
			"/choose-replacement-attorneys-address?id=123",
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{
					ReplacementAttorneys: []Attorney{
						{FirstNames: "John", Address: place.Address{Line1: "abc"}, ID: "123"},
					},
				}, nil)
			lpaStore.
				On("Put", mock.Anything, "session-id", &Lpa{
					ReplacementAttorneys: []Attorney{
						{
							ID:          "123",
							FirstNames:  "John",
							LastName:    "Doe",
							Email:       "john@example.com",
							DateOfBirth: time.Date(1990, time.January, 2, 0, 0, 0, 0, time.UTC),
							Address:     place.Address{Line1: "abc"},
						},
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

			r, _ := http.NewRequest(http.MethodPost, tc.requestUrl, strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := ChooseReplacementAttorneys(nil, lpaStore, mockRandom)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedNextUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostChooseReplacementAttorneysWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *chooseReplacementAttorneysData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name":           {"Doe"},
				"email":               {"name@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
			},
			dataMatcher: func(t *testing.T, data *chooseReplacementAttorneysData) bool {
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
			dataMatcher: func(t *testing.T, data *chooseReplacementAttorneysData) bool {
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
			dataMatcher: func(t *testing.T, data *chooseReplacementAttorneysData) bool {
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
				"ignore-warning":      {"attorneyDateOfBirthIsUnder18"},
			},
			dataMatcher: func(t *testing.T, data *chooseReplacementAttorneysData) bool {
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
				Return(&Lpa{}, nil)

			template := &mockTemplate{}
			template.
				On("Func", w, mock.MatchedBy(func(data *chooseReplacementAttorneysData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := ChooseReplacementAttorneys(template.Func, lpaStore, mockRandom)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, lpaStore, template)
		})
	}
}

func TestPostChooseReplacementAttorneysWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
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

	err := ChooseReplacementAttorneys(nil, lpaStore, mockRandom)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}
