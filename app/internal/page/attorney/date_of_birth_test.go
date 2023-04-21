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
		lpa     *page.Lpa
		appData page.AppData
	}{
		"attorney": {
			lpa: &page.Lpa{
				ID: "lpa-id",
				AttorneyProvidedDetails: actor.Attorneys{{
					ID: "attorney-id",
				}},
			},
			appData: testAppData,
		},
		"replacement attorney": {
			lpa: &page.Lpa{
				ID: "lpa-id",
				ReplacementAttorneyProvidedDetails: actor.Attorneys{{
					ID: "attorney-id",
				}},
			},
			appData: testReplacementAppData,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.lpa, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &dateOfBirthData{
					App:  tc.appData,
					Lpa:  tc.lpa,
					Form: &dateOfBirthForm{},
				}).
				Return(nil)

			err := DateOfBirth(template.Execute, lpaStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetDateOfBirthWhenAttorneyDetailsDontExist(t *testing.T) {
	testcases := map[string]struct {
		appData page.AppData
		lpa     *page.Lpa
	}{
		"attorney": {
			appData: testAppData,
			lpa: &page.Lpa{
				ReplacementAttorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
				},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				Attorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.lpa, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &dateOfBirthData{
					App:  tc.appData,
					Lpa:  tc.lpa,
					Form: &dateOfBirthForm{},
				}).
				Return(nil)

			err := DateOfBirth(template.Execute, lpaStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetDateOfBirthFromStore(t *testing.T) {
	testcases := map[string]struct {
		appData page.AppData
		lpa     *page.Lpa
	}{
		"attorney": {
			appData: testAppData,
			lpa: &page.Lpa{
				Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
				AttorneyProvidedDetails: actor.Attorneys{{
					ID:          "attorney-id",
					DateOfBirth: date.New("1997", "1", "2"),
				}},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				ReplacementAttorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
				ReplacementAttorneyProvidedDetails: actor.Attorneys{{
					ID:          "attorney-id",
					DateOfBirth: date.New("1997", "1", "2"),
				}},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.lpa, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &dateOfBirthData{
					App: tc.appData,
					Lpa: tc.lpa,
					Form: &dateOfBirthForm{
						Dob: date.New("1997", "1", "2"),
					},
				}).
				Return(nil)

			err := DateOfBirth(template.Execute, lpaStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetDateOfBirthWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := DateOfBirth(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDateOfBirthWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
		AttorneyProvidedDetails: actor.Attorneys{{
			ID: "attorney-id",
		}},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &dateOfBirthData{
			App:  testAppData,
			Lpa:  lpa,
			Form: &dateOfBirthForm{},
		}).
		Return(expectedError)

	err := DateOfBirth(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDateOfBirth(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form       url.Values
		lpa        *page.Lpa
		updatedLpa *page.Lpa
		appData    page.AppData
	}{
		"valid": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			lpa: &page.Lpa{
				AttorneyProvidedDetails: actor.Attorneys{{ID: "attorney-id"}},
			},
			updatedLpa: &page.Lpa{
				AttorneyProvidedDetails: actor.Attorneys{{
					ID:          "attorney-id",
					DateOfBirth: date.New(validBirthYear, "1", "2"),
				}},
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
			lpa: &page.Lpa{
				AttorneyProvidedDetails: actor.Attorneys{{ID: "attorney-id"}},
			},
			updatedLpa: &page.Lpa{
				AttorneyProvidedDetails: actor.Attorneys{{
					ID:          "attorney-id",
					DateOfBirth: date.New("1900", "1", "2"),
				}},
			},
			appData: testAppData,
		},
		"replacement attorney valid": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			lpa: &page.Lpa{
				ReplacementAttorneyProvidedDetails: actor.Attorneys{{ID: "attorney-id"}},
			},
			updatedLpa: &page.Lpa{
				ReplacementAttorneyProvidedDetails: actor.Attorneys{{
					ID:          "attorney-id",
					DateOfBirth: date.New(validBirthYear, "1", "2"),
				}},
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
			lpa: &page.Lpa{
				ReplacementAttorneyProvidedDetails: actor.Attorneys{{ID: "attorney-id"}},
			},
			updatedLpa: &page.Lpa{
				ReplacementAttorneyProvidedDetails: actor.Attorneys{{
					ID:          "attorney-id",
					DateOfBirth: date.New("1900", "1", "2"),
				}},
			},
			appData: testReplacementAppData,
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
				Return(tc.lpa, nil)
			lpaStore.
				On("Put", r.Context(), tc.updatedLpa).
				Return(nil)

			err := DateOfBirth(nil, lpaStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.NextPage, resp.Header.Get("Location"))
		})
	}
}

func TestPostDateOfBirthWhenAttorneyDetailsDontExist(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form      url.Values
		attorneys actor.Attorneys
	}{
		"valid": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			attorneys: actor.Attorneys{{
				ID:          "attorney-id",
				DateOfBirth: date.New(validBirthYear, "1", "2"),
			}},
		},
		"warning ignored": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
			},
			attorneys: actor.Attorneys{{
				ID:          "attorney-id",
				DateOfBirth: date.New("1900", "1", "2"),
			}},
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
					ID:                      "lpa-id",
					AttorneyProvidedDetails: tc.attorneys,
				}).
				Return(nil)

			err := DateOfBirth(nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.NextPage, resp.Header.Get("Location"))
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
				return assert.Equal(t, validation.With("date-of-birth", validation.DateMustBeRealError{Label: "dateOfBirth"}), data.Errors)
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

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					ID: "lpa-id",
					AttorneyProvidedDetails: actor.Attorneys{{
						ID: "attorney-id",
					}},
				}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, mock.MatchedBy(func(data *dateOfBirthData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := DateOfBirth(template.Execute, lpaStore)(testAppData, w, r)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			AttorneyProvidedDetails: actor.Attorneys{{
				ID: "attorney-id",
			}},
		}, expectedError)

	err := DateOfBirth(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
		form   *dateOfBirthForm
		errors validation.List
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
				With("date-of-birth", validation.EnterError{Label: "dateOfBirth"}),
		},
		"future-dob": {
			form: &dateOfBirthForm{
				Dob: now.AddDate(0, 0, 1),
			},
			errors: validation.With("date-of-birth", validation.DateMustBePastError{Label: "dateOfBirth"}),
		},
		"dob-under-18": {
			form: &dateOfBirthForm{
				Dob: now.AddDate(0, 0, -1),
			},
			errors: validation.With("date-of-birth", validation.CustomError{Label: "youAttorneyAreUnder18Error"}),
		},
		"invalid-dob": {
			form: &dateOfBirthForm{
				Dob: date.New("2000", "22", "2"),
			},
			errors: validation.With("date-of-birth", validation.DateMustBeRealError{Label: "dateOfBirth"}),
		},
		"invalid-missing-dob": {
			form: &dateOfBirthForm{
				Dob: date.New("1", "", "1"),
			},
			errors: validation.With("date-of-birth", validation.DateMissingError{Label: "dateOfBirth", MissingMonth: true}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
