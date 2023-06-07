package attorney

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

func TestGetDateOfBirth(t *testing.T) {
	testcases := map[string]struct {
		appData page.AppData
	}{
		"attorney": {
			appData: testAppData,
		},
		"replacement attorney": {
			appData: testReplacementAppData,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Get", r.Context()).
				Return(&actor.AttorneyProvidedDetails{}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &dateOfBirthData{
					App:  tc.appData,
					Form: &dateOfBirthForm{},
				}).
				Return(nil)

			err := DateOfBirth(template.Execute, attorneyStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetDateOfBirthWhenAttorneyDetailsDontExist(t *testing.T) {
	testcases := map[string]struct {
		appData page.AppData
	}{
		"attorney": {
			appData: testAppData,
		},
		"replacement attorney": {
			appData: testReplacementAppData,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Get", r.Context()).
				Return(&actor.AttorneyProvidedDetails{}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &dateOfBirthData{
					App:  tc.appData,
					Form: &dateOfBirthForm{},
				}).
				Return(nil)

			err := DateOfBirth(template.Execute, attorneyStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetDateOfBirthFromStore(t *testing.T) {
	testcases := map[string]struct {
		appData  page.AppData
		attorney *actor.AttorneyProvidedDetails
	}{
		"attorney": {
			appData: testAppData,
		},
		"replacement attorney": {
			appData: testReplacementAppData,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Get", r.Context()).
				Return(&actor.AttorneyProvidedDetails{
					DateOfBirth: date.New("1997", "1", "2"),
				}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &dateOfBirthData{
					App: tc.appData,
					Form: &dateOfBirthForm{
						Dob: date.New("1997", "1", "2"),
					},
				}).
				Return(nil)

			err := DateOfBirth(template.Execute, attorneyStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetDateOfBirthWhenAttorneyStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, expectedError)

	err := DateOfBirth(nil, attorneyStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDateOfBirthWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := DateOfBirth(template.Execute, attorneyStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDateOfBirth(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form            url.Values
		updatedAttorney *actor.AttorneyProvidedDetails
		appData         page.AppData
	}{
		"valid": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				DateOfBirth: date.New(validBirthYear, "1", "2"),
			},
			appData: testAppData,
		},
		"warning ignored": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				DateOfBirth: date.New("1900", "1", "2"),
			},
			appData: testAppData,
		},
		"replacement attorney valid": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				DateOfBirth: date.New(validBirthYear, "1", "2"),
			},
			appData: testReplacementAppData,
		},
		"replacement attorney warning ignored": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				DateOfBirth: date.New("1900", "1", "2"),
			},
			appData: testReplacementAppData,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Get", r.Context()).
				Return(&actor.AttorneyProvidedDetails{}, nil)
			attorneyStore.
				On("Put", r.Context(), tc.updatedAttorney).
				Return(nil)

			err := DateOfBirth(nil, attorneyStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/attorney/lpa-id"+page.Paths.Attorney.MobileNumber, resp.Header.Get("Location"))
		})
	}
}

func TestPostDateOfBirthWhenAttorneyDetailsDontExist(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form            url.Values
		providedDetails *actor.AttorneyProvidedDetails
	}{
		"valid": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			providedDetails: &actor.AttorneyProvidedDetails{
				DateOfBirth: date.New(validBirthYear, "1", "2"),
			},
		},
		"warning ignored": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
			},
			providedDetails: &actor.AttorneyProvidedDetails{
				DateOfBirth: date.New("1900", "1", "2"),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Get", r.Context()).
				Return(&actor.AttorneyProvidedDetails{}, nil)
			attorneyStore.
				On("Put", r.Context(), tc.providedDetails).
				Return(nil)

			err := DateOfBirth(nil, attorneyStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/attorney/lpa-id"+page.Paths.Attorney.MobileNumber, resp.Header.Get("Location"))
		})
	}
}

func TestPostDateOfBirthWhenInputRequired(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *dateOfBirthData) bool
	}{
		"validation error": {
			form: url.Values{
				"date-of-birth-day":   {"55"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			dataMatcher: func(t *testing.T, data *dateOfBirthData) bool {
				return assert.Equal(t, validation.With("date-of-birth", validation.DateMustBeRealError{Label: "yourDateOfBirth"}), data.Errors)
			},
		},
		"dob warning": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
			},
			dataMatcher: func(t *testing.T, data *dateOfBirthData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Get", r.Context()).
				Return(&actor.AttorneyProvidedDetails{}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, mock.MatchedBy(func(data *dateOfBirthData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := DateOfBirth(template.Execute, attorneyStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostYourDetailsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1999"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)
	attorneyStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := DateOfBirth(nil, attorneyStore)(testAppData, w, r)
	assert.Equal(t, expectedError, err)
}

func TestReadDateOfBirthForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
		"ignore-dob-warning":  {"xyz"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readDateOfBirthForm(r)

	assert.Equal(date.New("1990", "1", "2"), result.Dob)
	assert.Equal("xyz", result.IgnoreDobWarning)
}

func TestDateOfBirthFormValidate(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form          *dateOfBirthForm
		isReplacement bool
		errors        validation.List
	}{
		"valid": {
			form: &dateOfBirthForm{
				Dob:              validDob,
				IgnoreDobWarning: "xyz",
			},
		},
		"missing": {
			form: &dateOfBirthForm{},
			errors: validation.
				With("date-of-birth", validation.EnterError{Label: "yourDateOfBirth"}),
		},
		"future": {
			form: &dateOfBirthForm{
				Dob: now.AddDate(0, 0, 1),
			},
			errors: validation.With("date-of-birth", validation.DateMustBePastError{Label: "yourDateOfBirth"}),
		},
		"under 18 attorney": {
			form: &dateOfBirthForm{
				Dob: now.AddDate(0, 0, -1),
			},
			errors: validation.With("date-of-birth", validation.CustomError{Label: "youAttorneyAreUnder18Error"}),
		},
		"under 18 replacement attorney": {
			form: &dateOfBirthForm{
				Dob: now.AddDate(0, 0, -1),
			},
			isReplacement: true,
			errors:        validation.With("date-of-birth", validation.CustomError{Label: "youReplacementAttorneyAreUnder18Error"}),
		},
		"invalid": {
			form: &dateOfBirthForm{
				Dob: date.New("2000", "22", "2"),
			},
			errors: validation.With("date-of-birth", validation.DateMustBeRealError{Label: "yourDateOfBirth"}),
		},
		"missing part": {
			form: &dateOfBirthForm{
				Dob: date.New("1", "", "1"),
			},
			errors: validation.With("date-of-birth", validation.DateMissingError{Label: "yourDateOfBirth", MissingMonth: true}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate(tc.isReplacement))
		})
	}
}
