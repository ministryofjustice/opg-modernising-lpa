package donor

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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysData{
			App:  TestAppData,
			Form: &chooseAttorneysForm{},
		}).
		Return(nil)

	err := ChooseReplacementAttorneys(template.Func, lpaStore, MockRandom)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetChooseReplacementAttorneysWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, ExpectedError)

	err := ChooseReplacementAttorneys(nil, lpaStore, MockRandom)(TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChooseReplacementAttorneysFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			ReplacementAttorneys: actor.Attorneys{
				{FirstNames: "John", ID: "1"},
			},
		}, nil)

	template := &MockTemplate{}

	err := ChooseReplacementAttorneys(template.Func, lpaStore, MockRandom)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChooseReplacementAttorneysWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysData{
			App:  TestAppData,
			Form: &chooseAttorneysForm{},
		}).
		Return(ExpectedError)

	err := ChooseReplacementAttorneys(template.Func, lpaStore, MockRandom)(TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostChooseReplacementAttorneysAttorneyDoesNotExists(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form     url.Values
		attorney actor.Attorney
	}{
		"valid": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"email":               {"john@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			attorney: actor.Attorney{
				FirstNames:  "John",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				ID:          "123",
			},
		},
		"dob warning ignored": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"email":               {"john@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
			},
			attorney: actor.Attorney{
				FirstNames:  "John",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: date.New("1900", "1", "2"),
				ID:          "123",
			},
		},
		"name warning ignored": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"email":               {"john@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeReplacementAttorney, actor.TypeDonor, "Jane", "Doe").String()},
			},
			attorney: actor.Attorney{
				FirstNames:  "Jane",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				ID:          "123",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := &MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					You: actor.Person{FirstNames: "Jane", LastName: "Doe"},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					You:                  actor.Person{FirstNames: "Jane", LastName: "Doe"},
					ReplacementAttorneys: actor.Attorneys{tc.attorney},
					Tasks:                page.Tasks{ChooseReplacementAttorneys: page.TaskInProgress},
				}).
				Return(nil)

			err := ChooseReplacementAttorneys(nil, lpaStore, MockRandom)(TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneysAddress+"?id=123", resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostChooseReplacementAttorneysAttorneyExists(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form     url.Values
		attorney actor.Attorney
	}{
		"valid": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"email":               {"john@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			attorney: actor.Attorney{
				FirstNames:  "John",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				Address:     place.Address{Line1: "abc"},
				ID:          "123",
			},
		},
		"dob warning ignored": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"email":               {"john@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
			},
			attorney: actor.Attorney{
				FirstNames:  "John",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: date.New("1900", "1", "2"),
				Address:     place.Address{Line1: "abc"},
				ID:          "123",
			},
		},
		"name warning ignored": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"email":               {"john@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeReplacementAttorney, actor.TypeDonor, "Jane", "Doe").String()},
			},
			attorney: actor.Attorney{
				FirstNames:  "Jane",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				Address:     place.Address{Line1: "abc"},
				ID:          "123",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := &MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					You: actor.Person{FirstNames: "Jane", LastName: "Doe"},
					ReplacementAttorneys: actor.Attorneys{
						{
							FirstNames: "John",
							ID:         "123",
							Address:    place.Address{Line1: "abc"},
						},
					},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					You:                  actor.Person{FirstNames: "Jane", LastName: "Doe"},
					ReplacementAttorneys: actor.Attorneys{tc.attorney},
				}).
				Return(nil)

			err := ChooseReplacementAttorneys(nil, lpaStore, MockRandom)(TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneysAddress+"?id=123", resp.Header.Get("Location"))
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
			"/lpa/lpa-id/test",
		},
		"without from value": {
			"/?from=&id=123",
			"/lpa/lpa-id" + page.Paths.ChooseReplacementAttorneysAddress + "?id=123",
		},
		"missing from key": {
			"/?id=123",
			"/lpa/lpa-id" + page.Paths.ChooseReplacementAttorneysAddress + "?id=123",
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			form := url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"email":               {"john@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, tc.requestUrl, strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := &MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					ReplacementAttorneys: actor.Attorneys{
						{FirstNames: "John", Address: place.Address{Line1: "abc"}, ID: "123"},
					},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					ReplacementAttorneys: actor.Attorneys{
						{
							ID:          "123",
							FirstNames:  "John",
							LastName:    "Doe",
							Email:       "john@example.com",
							DateOfBirth: date.New("1990", "1", "2"),
							Address:     place.Address{Line1: "abc"},
						},
					},
				}).
				Return(nil)

			err := ChooseReplacementAttorneys(nil, lpaStore, MockRandom)(TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedNextUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostChooseReplacementAttorneysWhenInputRequired(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

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
				return assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
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
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
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
				"ignore-dob-warning":  {"attorneyDateOfBirthIsUnder18"},
			},
			dataMatcher: func(t *testing.T, data *chooseReplacementAttorneysData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
		"name warning": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"email":               {"name@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			dataMatcher: func(t *testing.T, data *chooseReplacementAttorneysData) bool {
				return assert.Equal(t, "", data.DobWarning) &&
					assert.Equal(t, actor.NewSameNameWarning(actor.TypeReplacementAttorney, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.True(t, data.Errors.None())
			},
		},
		"name warning ignored but other errors": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
				"ignore-name-warning": {"errorDonorMatchesActor|aReplacementAttorney|Jane|Doe"},
			},
			dataMatcher: func(t *testing.T, data *chooseReplacementAttorneysData) bool {
				return assert.Equal(t, "", data.DobWarning) &&
					assert.Equal(t, actor.NewSameNameWarning(actor.TypeReplacementAttorney, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.Equal(t, validation.With("email", validation.EnterError{Label: "email"}), data.Errors)
			},
		},
		"other name warning ignored": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"email":               {"name@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
				"ignore-name-warning": {"errorAttorneyMatchesActor|aReplacementAttorney|Jane|Doe"},
			},
			dataMatcher: func(t *testing.T, data *chooseReplacementAttorneysData) bool {
				return assert.Equal(t, "", data.DobWarning) &&
					assert.Equal(t, actor.NewSameNameWarning(actor.TypeReplacementAttorney, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.True(t, data.Errors.None())
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := &MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					You: actor.Person{FirstNames: "Jane", LastName: "Doe"},
				}, nil)

			template := &MockTemplate{}
			template.
				On("Func", w, mock.MatchedBy(func(data *chooseReplacementAttorneysData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := ChooseReplacementAttorneys(template.Func, lpaStore, MockRandom)(TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, lpaStore, template)
		})
	}
}

func TestPostChooseReplacementAttorneysWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"email":               {"john@example.com"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(ExpectedError)

	err := ChooseReplacementAttorneys(nil, lpaStore, MockRandom)(TestAppData, w, r)

	assert.Equal(t, ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestReplacementAttorneyMatches(t *testing.T) {
	lpa := &page.Lpa{
		You: actor.Person{FirstNames: "a", LastName: "b"},
		Attorneys: actor.Attorneys{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
		},
		ReplacementAttorneys: actor.Attorneys{
			{FirstNames: "g", LastName: "h"},
			{ID: "123", FirstNames: "i", LastName: "j"},
		},
		CertificateProvider: actor.CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: actor.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{FirstNames: "o", LastName: "p"},
		},
	}

	assert.Equal(t, actor.TypeNone, replacementAttorneyMatches(lpa, "123", "x", "y"))
	assert.Equal(t, actor.TypeDonor, replacementAttorneyMatches(lpa, "123", "a", "b"))
	assert.Equal(t, actor.TypeAttorney, replacementAttorneyMatches(lpa, "123", "c", "d"))
	assert.Equal(t, actor.TypeAttorney, replacementAttorneyMatches(lpa, "123", "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, replacementAttorneyMatches(lpa, "123", "g", "h"))
	assert.Equal(t, actor.TypeNone, replacementAttorneyMatches(lpa, "123", "i", "j"))
	assert.Equal(t, actor.TypeCertificateProvider, replacementAttorneyMatches(lpa, "123", "k", "l"))
	assert.Equal(t, actor.TypePersonToNotify, replacementAttorneyMatches(lpa, "123", "m", "n"))
	assert.Equal(t, actor.TypePersonToNotify, replacementAttorneyMatches(lpa, "123", "o", "p"))
}
