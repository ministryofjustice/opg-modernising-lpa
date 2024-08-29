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

func TestGetChooseReplacementAttorneys(t *testing.T) {
	testcases := map[string]struct {
		lpaType                   lpadata.LpaType
		attorneys                 donordata.Attorneys
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
		"property and affairs with lay attorney": {
			lpaType:                   lpadata.LpaTypePropertyAndAffairs,
			attorneys:                 donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
			expectedShowTrustCorpLink: true,
		},
		"personal welfare with lay attorney": {
			lpaType:                   lpadata.LpaTypePersonalWelfare,
			attorneys:                 donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
			expectedShowTrustCorpLink: false,
		},
		"property and affairs with trust corporation": {
			lpaType:                   lpadata.LpaTypePropertyAndAffairs,
			attorneys:                 donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{Name: "a"}},
			expectedShowTrustCorpLink: false,
		},
		"personal welfare with trust corporation": {
			lpaType:                   lpadata.LpaTypePersonalWelfare,
			attorneys:                 donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{Name: "a"}},
			expectedShowTrustCorpLink: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?id="+testUID.String(), nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &chooseReplacementAttorneysData{
					App: testAppData,
					Donor: &donordata.Provided{
						Type:      tc.lpaType,
						Attorneys: tc.attorneys,
					},
					Form:                     &chooseAttorneysForm{},
					ShowTrustCorporationLink: tc.expectedShowTrustCorpLink,
				}).
				Return(nil)

			err := ChooseReplacementAttorneys(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
				Type:      tc.lpaType,
				Attorneys: tc.attorneys,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetChooseReplacementAttorneysWhenNoID(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseReplacementAttorneys(nil, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{FirstNames: "John", UID: actoruid.New()}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetChooseReplacementAttorneysDobWarningIsAlwaysShown(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+testUID.String(), nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseReplacementAttorneysData{
			App: testAppData,
			Donor: &donordata.Provided{
				ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
					{UID: testUID, DateOfBirth: date.New("1900", "1", "2")},
				}},
			},
			Form: &chooseAttorneysForm{
				Dob: date.New("1900", "1", "2"),
			},
			DobWarning: "dateOfBirthIsOver100",
		}).
		Return(nil)

	err := ChooseReplacementAttorneys(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{UID: testUID, DateOfBirth: date.New("1900", "1", "2")},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementAttorneysWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+testUID.String(), nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ChooseReplacementAttorneys(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysAttorneyDoesNotExists(t *testing.T) {
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
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeReplacementAttorney, actor.TypeDonor, "Jane", "Doe").String()},
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
					LpaID:                "lpa-id",
					Donor:                donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
					ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{tc.attorney}},
					Tasks:                task.DonorTasks{ChooseReplacementAttorneys: task.StateInProgress},
				}).
				Return(nil)

			err := ChooseReplacementAttorneys(nil, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathChooseReplacementAttorneysAddress.Format("lpa-id")+"?id="+testUID.String(), resp.Header.Get("Location"))
		})
	}
}

func TestPostChooseReplacementAttorneysAttorneyExists(t *testing.T) {
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
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeReplacementAttorney, actor.TypeDonor, "Jane", "Doe").String()},
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
					LpaID:                "lpa-id",
					Donor:                donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
					ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{tc.attorney}},
					Tasks:                task.DonorTasks{ChooseReplacementAttorneys: task.StateCompleted},
				}).
				Return(nil)

			err := ChooseReplacementAttorneys(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
				ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
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
			assert.Equal(t, donor.PathChooseReplacementAttorneysAddress.Format("lpa-id")+"?id="+uid.String(), resp.Header.Get("Location"))
		})
	}
}

func TestPostChooseReplacementAttorneysNameWarningOnlyShownWhenAttorneyAndFormNamesAreDifferent(t *testing.T) {
	form := url.Values{
		"first-names":         {"Jane"},
		"last-name":           {"Doe"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"2000"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+testUID.String(), strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
			ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
				{
					FirstNames:  "Jane",
					LastName:    "Doe",
					UID:         testUID,
					Address:     place.Address{Line1: "abc"},
					DateOfBirth: date.New("2000", "1", "2"),
				},
			}},
			Tasks: task.DonorTasks{ChooseReplacementAttorneys: task.StateCompleted},
		}).
		Return(nil)

	err := ChooseReplacementAttorneys(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "Jane", LastName: "Doe", UID: testUID, Address: place.Address{Line1: "abc"}},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseReplacementAttorneysAddress.Format("lpa-id")+"?id="+testUID.String(), resp.Header.Get("Location"))
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
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning) &&
					assert.Nil(t, data.NameWarning) &&
					assert.True(t, data.Errors.None())
			},
		},
		"dob warning ignored but other errors": {
			form: url.Values{
				"first-names":         {"John"},
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
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning) &&
					assert.Nil(t, data.NameWarning) &&
					assert.True(t, data.Errors.None())
			},
		},
		"name warning": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
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
				"ignore-name-warning": {"errorDonorMatchesActor|aReplacementAttorney|Jane|Doe"},
			},
			dataMatcher: func(t *testing.T, data *chooseReplacementAttorneysData) bool {
				return assert.Equal(t, "", data.DobWarning) &&
					assert.Equal(t, actor.NewSameNameWarning(actor.TypeReplacementAttorney, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.Equal(t, validation.With("date-of-birth", validation.DateMissingError{Label: "dateOfBirth", MissingDay: false, MissingMonth: false, MissingYear: true}), data.Errors)
			},
		},
		"other name warning ignored": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
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
			r, _ := http.NewRequest(http.MethodPost, "/?id="+testUID.String(), strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, mock.MatchedBy(func(data *chooseReplacementAttorneysData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := ChooseReplacementAttorneys(template.Execute, nil)(testAppData, w, r, &donordata.Provided{Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
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
	r, _ := http.NewRequest(http.MethodPost, "/?id="+testUID.String(), strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ChooseReplacementAttorneys(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestReplacementAttorneyMatches(t *testing.T) {
	uid := actoruid.New()
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "g", LastName: "h"},
			{UID: uid, FirstNames: "i", LastName: "j"},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{FirstNames: "o", LastName: "p"},
		},
		AuthorisedSignatory: donordata.AuthorisedSignatory{FirstNames: "a", LastName: "s"},
		IndependentWitness:  donordata.IndependentWitness{FirstNames: "i", LastName: "w"},
	}

	assert.Equal(t, actor.TypeNone, replacementAttorneyMatches(donor, uid, "x", "y"))
	assert.Equal(t, actor.TypeDonor, replacementAttorneyMatches(donor, uid, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, replacementAttorneyMatches(donor, uid, "C", "D"))
	assert.Equal(t, actor.TypeAttorney, replacementAttorneyMatches(donor, uid, "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, replacementAttorneyMatches(donor, uid, "g", "h"))
	assert.Equal(t, actor.TypeNone, replacementAttorneyMatches(donor, uid, "i", "j"))
	assert.Equal(t, actor.TypeCertificateProvider, replacementAttorneyMatches(donor, uid, "K", "l"))
	assert.Equal(t, actor.TypePersonToNotify, replacementAttorneyMatches(donor, uid, "m", "n"))
	assert.Equal(t, actor.TypePersonToNotify, replacementAttorneyMatches(donor, uid, "O", "P"))
	assert.Equal(t, actor.TypeAuthorisedSignatory, replacementAttorneyMatches(donor, uid, "a", "s"))
	assert.Equal(t, actor.TypeIndependentWitness, replacementAttorneyMatches(donor, uid, "i", "w"))
}

func TestReplacementAttorneyMatchesEmptyNamesIgnored(t *testing.T) {
	uid := actoruid.New()
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "", LastName: ""},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
			{UID: uid, FirstNames: "", LastName: ""},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "", LastName: ""},
		},
	}

	assert.Equal(t, actor.TypeNone, replacementAttorneyMatches(donor, uid, "", ""))
}
