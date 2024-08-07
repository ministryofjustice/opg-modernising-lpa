package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseAttorneys(t *testing.T) {
	testcases := map[string]struct {
		lpaType                   lpadata.LpaType
		replacementAttorneys      donordata.Attorneys
		expectedShowTrustCorpLink bool
	}{
		"property and affairs": {
			lpaType:                   lpadata.LpaTypePropertyAndAffairs,
			expectedShowTrustCorpLink: true,
		},
		"personal welfare": {
			lpaType:                   lpadata.LpaTypePersonalWelfare,
			expectedShowTrustCorpLink: false,
		},
		"property and affairs with lay replacement attorney": {
			lpaType:                   lpadata.LpaTypePropertyAndAffairs,
			replacementAttorneys:      donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
			expectedShowTrustCorpLink: true,
		},
		"personal welfare with lay replacement attorney": {
			lpaType:                   lpadata.LpaTypePersonalWelfare,
			replacementAttorneys:      donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
			expectedShowTrustCorpLink: false,
		},
		"property and affairs with replacement trust corporation": {
			lpaType:                   lpadata.LpaTypePropertyAndAffairs,
			replacementAttorneys:      donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{Name: "a"}},
			expectedShowTrustCorpLink: false,
		},
		"personal welfare with replacement trust corporation": {
			lpaType:                   lpadata.LpaTypePersonalWelfare,
			replacementAttorneys:      donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{Name: "a"}},
			expectedShowTrustCorpLink: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?id="+testUID.String(), nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &chooseAttorneysData{
					App: testAppData,
					Donor: &donordata.Provided{
						Type:                 tc.lpaType,
						ReplacementAttorneys: tc.replacementAttorneys,
					},
					Form:                     &chooseAttorneysForm{},
					ShowDetails:              true,
					ShowTrustCorporationLink: tc.expectedShowTrustCorpLink,
				}).
				Return(nil)

			err := ChooseAttorneys(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
				Type:                 tc.lpaType,
				ReplacementAttorneys: tc.replacementAttorneys,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetChooseAttorneysWhenNoID(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneys(nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "John", UID: testUID},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetChooseAttorneysDobWarningIsAlwaysShown(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+testUID.String(), nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAttorneysData{
			App: testAppData,
			Donor: &donordata.Provided{
				Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
					{UID: testUID, DateOfBirth: date.New("1900", "1", "2")},
				}},
			},
			Form: &chooseAttorneysForm{
				Dob: date.New("1900", "1", "2"),
			},
			ShowDetails: false,
			DobWarning:  "dateOfBirthIsOver100",
		}).
		Return(nil)

	err := ChooseAttorneys(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{UID: testUID, DateOfBirth: date.New("1900", "1", "2")},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseAttorneysWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+testUID.String(), nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ChooseAttorneys(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneysAttorneyDoesNotExist(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form     url.Values
		attorney donordata.Attorney
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
			attorney: donordata.Attorney{
				FirstNames:  "John",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				UID:         testUID,
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
			attorney: donordata.Attorney{
				FirstNames:  "John",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: date.New("1900", "1", "2"),
				UID:         testUID,
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
			attorney: donordata.Attorney{
				FirstNames:  "Jane",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				UID:         testUID,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/?id="+testUID.String(), strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID: "lpa-id",
					Donor: donordata.Donor{
						FirstNames: "Jane",
						LastName:   "Doe",
					},
					Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{tc.attorney}},
					Tasks:     donordata.Tasks{ChooseAttorneys: task.StateInProgress},
				}).
				Return(nil)

			err := ChooseAttorneys(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					FirstNames: "Jane",
					LastName:   "Doe",
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathChooseAttorneysAddress.Format("lpa-id")+"?id="+testUID.String(), resp.Header.Get("Location"))
		})
	}
}

func TestPostChooseAttorneysAttorneyExists(t *testing.T) {
	uid := actoruid.New()
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form     url.Values
		attorney donordata.Attorney
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
			attorney: donordata.Attorney{
				FirstNames:  "John",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				Address:     place.Address{Line1: "abc"},
				UID:         uid,
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
			attorney: donordata.Attorney{
				FirstNames:  "John",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: date.New("1900", "1", "2"),
				Address:     place.Address{Line1: "abc"},
				UID:         uid,
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
			attorney: donordata.Attorney{
				FirstNames:  "Jane",
				LastName:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				Address:     place.Address{Line1: "abc"},
				UID:         uid,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:     "lpa-id",
					Donor:     donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
					Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{tc.attorney}},
					Tasks:     donordata.Tasks{ChooseAttorneys: task.StateCompleted},
				}).
				Return(nil)

			err := ChooseAttorneys(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
				Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
					{
						FirstNames: "John",
						UID:        uid,
						Address:    place.Address{Line1: "abc"},
					},
				}},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathChooseAttorneysAddress.Format("lpa-id")+"?id="+uid.String(), resp.Header.Get("Location"))
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

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
			Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
				{
					FirstNames:  "Jane",
					LastName:    "Doe",
					UID:         uid,
					Address:     place.Address{Line1: "abc"},
					DateOfBirth: date.New("2000", "1", "2"),
				},
			}},
			Tasks: donordata.Tasks{ChooseAttorneys: task.StateCompleted},
		}).
		Return(nil)

	err := ChooseAttorneys(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "Jane", LastName: "Doe", UID: uid, Address: place.Address{Line1: "abc"}},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseAttorneysAddress.Format("lpa-id")+"?id="+uid.String(), resp.Header.Get("Location"))
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
			r, _ := http.NewRequest(http.MethodPost, "/?id="+testUID.String(), strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, mock.MatchedBy(func(data *chooseAttorneysData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := ChooseAttorneys(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
				Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
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
	r, _ := http.NewRequest(http.MethodPost, "/?id="+testUID.String(), strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ChooseAttorneys(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

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
				Dob: now.AddDate(-18, 0, 2),
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
	uid := actoruid.New()

	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "c", LastName: "d"},
			{UID: uid, FirstNames: "e", LastName: "f"},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "g", LastName: "h"},
			{FirstNames: "i", LastName: "j"},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{FirstNames: "o", LastName: "p"},
		},
		AuthorisedSignatory: donordata.AuthorisedSignatory{FirstNames: "a", LastName: "s"},
		IndependentWitness:  donordata.IndependentWitness{FirstNames: "i", LastName: "w"},
	}

	assert.Equal(t, actor.TypeNone, attorneyMatches(donor, uid, "x", "y"))
	assert.Equal(t, actor.TypeDonor, attorneyMatches(donor, uid, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, attorneyMatches(donor, uid, "c", "d"))
	assert.Equal(t, actor.TypeNone, attorneyMatches(donor, uid, "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, attorneyMatches(donor, uid, "g", "h"))
	assert.Equal(t, actor.TypeReplacementAttorney, attorneyMatches(donor, uid, "I", "J"))
	assert.Equal(t, actor.TypeCertificateProvider, attorneyMatches(donor, uid, "k", "l"))
	assert.Equal(t, actor.TypePersonToNotify, attorneyMatches(donor, uid, "M", "N"))
	assert.Equal(t, actor.TypePersonToNotify, attorneyMatches(donor, uid, "o", "p"))
	assert.Equal(t, actor.TypeAuthorisedSignatory, attorneyMatches(donor, uid, "a", "s"))
	assert.Equal(t, actor.TypeIndependentWitness, attorneyMatches(donor, uid, "i", "w"))
}

func TestAttorneyMatchesEmptyNamesIgnored(t *testing.T) {
	uid := actoruid.New()

	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "", LastName: ""},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{UID: uid, FirstNames: "", LastName: ""},
			{FirstNames: "", LastName: ""},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "", LastName: ""},
		},
	}

	assert.Equal(t, actor.TypeNone, attorneyMatches(donor, uid, "", ""))
}
