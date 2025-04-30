package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
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

func TestPostChooseAttorneysWhenDOBWarning(t *testing.T) {
	form := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"email":               {"name@example.com"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1900"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+testUID.String(), strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	appData := appcontext.Data{Page: "/a"}
	err := ChooseAttorneys(nil, donorStore)(appData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathWarningInterruption.Format("lpa-id")+"?id="+testUID.String()+"&warningFrom=%2Fa", resp.Header.Get("Location"))
}

func TestPostChooseAttorneysWhenNameWarning(t *testing.T) {
	form := url.Values{
		"first-names":         {"Jane"},
		"last-name":           {"Doe"},
		"email":               {"name@example.com"},
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
		Return(nil)

	appData := appcontext.Data{Page: "/a"}
	err := ChooseAttorneys(nil, donorStore)(appData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, donor.PathWarningInterruption.Format("lpa-id")+"?id="+testUID.String()+"&warningFrom=%2Fa", resp.Header.Get("Location"))
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
