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

func TestGetChooseAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAttorneysData{
			App:         testAppData,
			Donor:       &actor.DonorProvidedDetails{},
			Form:        &chooseAttorneysForm{},
			ShowDetails: true,
		}).
		Return(nil)

	err := ChooseAttorneys(template.Execute, nil, mockUuidString)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseAttorneysFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneys(nil, nil, mockUuidString)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "John", ID: "1"},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetChooseAttorneysDobWarningIsAlwaysShown(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=1", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAttorneysData{
			App: testAppData,
			Donor: &actor.DonorProvidedDetails{
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{ID: "1", DateOfBirth: date.New("1900", "1", "2")},
				}},
			},
			Form: &chooseAttorneysForm{
				Dob: date.New("1900", "1", "2"),
			},
			ShowDetails: false,
			DobWarning:  "dateOfBirthIsOver100",
		}).
		Return(nil)

	err := ChooseAttorneys(template.Execute, nil, mockUuidString)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{},
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{ID: "1", DateOfBirth: date.New("1900", "1", "2")},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseAttorneysWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := ChooseAttorneys(template.Execute, nil, mockUuidString)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneysAttorneyDoesNotExist(t *testing.T) {
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
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeAttorney, actor.TypeDonor, "Jane", "Doe").String()},
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

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), &actor.DonorProvidedDetails{
					LpaID: "lpa-id",
					Donor: actor.Donor{
						FirstNames: "Jane",
						LastName:   "Doe",
					},
					Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{tc.attorney}},
					Tasks:     actor.DonorTasks{ChooseAttorneys: actor.TaskCompleted},
				}).
				Return(nil)

			err := ChooseAttorneys(nil, donorStore, mockUuidString)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{
					FirstNames: "Jane",
					LastName:   "Doe",
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.ChooseAttorneysAddress.Format("lpa-id")+"?id=123", resp.Header.Get("Location"))
		})
	}
}

func TestPostChooseAttorneysAttorneyExists(t *testing.T) {
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
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeAttorney, actor.TypeDonor, "Jane", "Doe").String()},
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

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), &actor.DonorProvidedDetails{
					LpaID:     "lpa-id",
					Donor:     actor.Donor{FirstNames: "Jane", LastName: "Doe"},
					Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{tc.attorney}},
					Tasks:     actor.DonorTasks{ChooseAttorneys: actor.TaskCompleted},
				}).
				Return(nil)

			err := ChooseAttorneys(nil, donorStore, mockUuidString)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{FirstNames: "Jane", LastName: "Doe"},
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{
						FirstNames: "John",
						ID:         "123",
						Address:    place.Address{Line1: "abc"},
					},
				}},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.ChooseAttorneysAddress.Format("lpa-id")+"?id=123", resp.Header.Get("Location"))
		})
	}
}

func TestPostChooseAttorneysNameWarningOnlyShownWhenAttorneyAndFormNamesAreDifferent(t *testing.T) {
	form := url.Values{
		"first-names":         {"Jane"},
		"last-name":           {"Doe"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"2000"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &actor.DonorProvidedDetails{
			LpaID: "lpa-id",
			Donor: actor.Donor{FirstNames: "Jane", LastName: "Doe"},
			Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
				{
					FirstNames:  "Jane",
					LastName:    "Doe",
					ID:          "123",
					Address:     place.Address{Line1: "abc"},
					DateOfBirth: date.New("2000", "1", "2"),
				},
			}},
			Tasks: actor.DonorTasks{ChooseAttorneys: actor.TaskCompleted},
		}).
		Return(nil)

	err := ChooseAttorneys(nil, donorStore, mockUuidString)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		Donor: actor.Donor{FirstNames: "Jane", LastName: "Doe"},
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "Jane", LastName: "Doe", ID: "123", Address: place.Address{Line1: "abc"}},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseAttorneysAddress.Format("lpa-id")+"?id=123", resp.Header.Get("Location"))
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
				return assert.Equal(t, "", data.DobWarning) &&
					assert.Nil(t, data.NameWarning) &&
					assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
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
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning) &&
					assert.Nil(t, data.NameWarning) &&
					assert.True(t, data.Errors.None())
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
			dataMatcher: func(t *testing.T, data *chooseAttorneysData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning) &&
					assert.Nil(t, data.NameWarning) &&
					assert.Equal(t, validation.With("last-name", validation.EnterError{Label: "lastName"}), data.Errors)
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
			dataMatcher: func(t *testing.T, data *chooseAttorneysData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning) &&
					assert.Nil(t, data.NameWarning) &&
					assert.True(t, data.Errors.None())
			},
		},
		"name warning": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"email":               {"name@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
			},
			dataMatcher: func(t *testing.T, data *chooseAttorneysData) bool {
				return assert.Equal(t, "", data.DobWarning) &&
					assert.Equal(t, actor.NewSameNameWarning(actor.TypeAttorney, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.True(t, data.Errors.None())
			},
		},
		"name warning ignored but other errors": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"ignore-name-warning": {"errorDonorMatchesActor|anAttorney|Jane|Doe"},
			},
			dataMatcher: func(t *testing.T, data *chooseAttorneysData) bool {
				return assert.Equal(t, "", data.DobWarning) &&
					assert.Equal(t, actor.NewSameNameWarning(actor.TypeAttorney, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.Equal(t, validation.With("date-of-birth", validation.DateMissingError{Label: "dateOfBirth", MissingDay: false, MissingMonth: false, MissingYear: true}), data.Errors)
			},
		},
		"other name warning ignored": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"email":               {"name@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
				"ignore-name-warning": {"errorReplacementAttorneyMatchesActor|anAttorney|Jane|Doe"},
			},
			dataMatcher: func(t *testing.T, data *chooseAttorneysData) bool {
				return assert.Equal(t, "", data.DobWarning) &&
					assert.Equal(t, actor.NewSameNameWarning(actor.TypeAttorney, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.True(t, data.Errors.None())
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			template := newMockTemplate(t)
			template.
				On("Execute", w, mock.MatchedBy(func(data *chooseAttorneysData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := ChooseAttorneys(template.Execute, nil, mockUuidString)(testAppData, w, r, &actor.DonorProvidedDetails{
				Donor: actor.Donor{FirstNames: "Jane", LastName: "Doe"},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostChooseAttorneysWhenStoreErrors(t *testing.T) {
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := ChooseAttorneys(nil, donorStore, mockUuidString)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
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
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readChooseAttorneysForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("john@example.com", result.Email)
	assert.Equal(date.New("1990", "1", "2"), result.Dob)
}

func TestChooseAttorneysFormValidate(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form   *chooseAttorneysForm
		errors validation.List
	}{
		"valid": {
			form: &chooseAttorneysForm{
				FirstNames: "A",
				LastName:   "B",
				Dob:        validDob,
			},
		},
		"max length": {
			form: &chooseAttorneysForm{
				FirstNames: strings.Repeat("x", 53),
				LastName:   strings.Repeat("x", 61),
				Dob:        validDob,
			},
		},
		"missing all": {
			form: &chooseAttorneysForm{},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}).
				With("date-of-birth", validation.EnterError{Label: "dateOfBirth"}),
		},
		"too long": {
			form: &chooseAttorneysForm{
				FirstNames: strings.Repeat("x", 54),
				LastName:   strings.Repeat("x", 62),
				Dob:        validDob,
			},
			errors: validation.
				With("first-names", validation.StringTooLongError{Label: "firstNames", Length: 53}).
				With("last-name", validation.StringTooLongError{Label: "lastName", Length: 61}),
		},
		"future dob": {
			form: &chooseAttorneysForm{
				FirstNames: "A",
				LastName:   "B",
				Dob:        now.AddDate(0, 0, 1),
			},
			errors: validation.With("date-of-birth", validation.DateMustBePastError{Label: "dateOfBirth"}),
		},
		"invalid dob": {
			form: &chooseAttorneysForm{
				FirstNames: "A",
				LastName:   "B",
				Dob:        date.New("2000", "22", "2"),
			},
			errors: validation.With("date-of-birth", validation.DateMustBeRealError{Label: "dateOfBirth"}),
		},
		"invalid missing dob": {
			form: &chooseAttorneysForm{
				FirstNames: "A",
				LastName:   "B",
				Dob:        date.New("1", "", "1"),
			},
			errors: validation.With("date-of-birth", validation.DateMissingError{Label: "dateOfBirth", MissingMonth: true}),
		},
		"invalid email": {
			form: &chooseAttorneysForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "person@",
				Dob:        validDob,
			},
			errors: validation.With("email", validation.EmailError{Label: "email"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}

func TestChooseAttorneysFormDobWarning(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form    *chooseAttorneysForm
		warning string
	}{
		"valid": {
			form: &chooseAttorneysForm{
				Dob: validDob,
			},
		},
		"future dob": {
			form: &chooseAttorneysForm{
				Dob: now.AddDate(0, 0, 1),
			},
		},
		"dob is 18": {
			form: &chooseAttorneysForm{
				Dob: now.AddDate(-18, 0, 0),
			},
		},
		"dob under 18": {
			form: &chooseAttorneysForm{
				Dob: now.AddDate(-18, 0, 1),
			},
			warning: "attorneyDateOfBirthIsUnder18",
		},
		"dob is 100": {
			form: &chooseAttorneysForm{
				Dob: now.AddDate(-100, 0, 0),
			},
		},
		"dob over 100": {
			form: &chooseAttorneysForm{
				Dob: now.AddDate(-100, 0, -1),
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

func TestAttorneyMatches(t *testing.T) {
	donor := &actor.DonorProvidedDetails{
		Donor: actor.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "c", LastName: "d"},
			{ID: "123", FirstNames: "e", LastName: "f"},
		}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "g", LastName: "h"},
			{FirstNames: "i", LastName: "j"},
		}},
		CertificateProvider: actor.CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: actor.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{FirstNames: "o", LastName: "p"},
		},
		AuthorisedSignatory: actor.AuthorisedSignatory{FirstNames: "a", LastName: "s"},
		IndependentWitness:  actor.IndependentWitness{FirstNames: "i", LastName: "w"},
	}

	assert.Equal(t, actor.TypeNone, attorneyMatches(donor, "123", "x", "y"))
	assert.Equal(t, actor.TypeDonor, attorneyMatches(donor, "123", "a", "b"))
	assert.Equal(t, actor.TypeAttorney, attorneyMatches(donor, "123", "c", "d"))
	assert.Equal(t, actor.TypeNone, attorneyMatches(donor, "123", "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, attorneyMatches(donor, "123", "g", "h"))
	assert.Equal(t, actor.TypeReplacementAttorney, attorneyMatches(donor, "123", "I", "J"))
	assert.Equal(t, actor.TypeCertificateProvider, attorneyMatches(donor, "123", "k", "l"))
	assert.Equal(t, actor.TypePersonToNotify, attorneyMatches(donor, "123", "M", "N"))
	assert.Equal(t, actor.TypePersonToNotify, attorneyMatches(donor, "123", "o", "p"))
	assert.Equal(t, actor.TypeAuthorisedSignatory, attorneyMatches(donor, "123", "a", "s"))
	assert.Equal(t, actor.TypeIndependentWitness, attorneyMatches(donor, "123", "i", "w"))
}

func TestAttorneyMatchesEmptyNamesIgnored(t *testing.T) {
	donor := &actor.DonorProvidedDetails{
		Donor: actor.Donor{FirstNames: "", LastName: ""},
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{ID: "123", FirstNames: "", LastName: ""},
			{FirstNames: "", LastName: ""},
		}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		CertificateProvider: actor.CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: actor.PeopleToNotify{
			{FirstNames: "", LastName: ""},
		},
	}

	assert.Equal(t, actor.TypeNone, attorneyMatches(donor, "123", "", ""))
}
