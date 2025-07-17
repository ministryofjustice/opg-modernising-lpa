package donorpage

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterAttorney(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+testUID.String(), nil)

	provided := &donordata.Provided{
		LpaID: "lpa-id",
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterAttorneyData{
			App:                      testAppData,
			Donor:                    provided,
			Form:                     &enterAttorneyForm{},
			ShowTrustCorporationLink: true,
		}).
		Return(nil)

	service := newMockAttorneyService(t)
	service.EXPECT().
		IsReplacement().
		Return(false)
	service.EXPECT().
		CanAddTrustCorporation(provided).
		Return(true)

	err := EnterAttorney(template.Execute, service)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterAttorneyWhenNoID(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := EnterAttorney(nil, testAttorneyService(t))(testAppData, w, r, &donordata.Provided{
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

func TestGetEnterAttorneyWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+testUID.String(), nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := EnterAttorney(template.Execute, testAttorneyService(t))(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAttorneyWhenAttorneyDoesNotExist(t *testing.T) {
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

			provided := &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					FirstNames: "Jane",
					LastName:   "Doe",
				},
			}

			service := testAttorneyService(t)
			service.EXPECT().
				Put(r.Context(), provided, tc.attorney).
				Return(nil)

			err := EnterAttorney(nil, service)(testAppData, w, r, provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathChooseAttorneysAddress.Format("lpa-id")+"?id="+testUID.String(), resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterAttorneyWhenAttorneyExists(t *testing.T) {
	uid := actoruid.New()
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	attorney := donordata.Attorney{
		FirstNames:  "John",
		LastName:    "Doe",
		Email:       "john@example.com",
		DateOfBirth: date.New(validBirthYear, "1", "2"),
		Address:     place.Address{Line1: "abc"},
		UID:         uid,
	}

	testcases := map[bool]struct {
		provided *donordata.Provided
		nextPath donor.Path
	}{
		false: {
			provided: &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
				Attorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{{
						FirstNames: "John",
						UID:        uid,
						Address:    place.Address{Line1: "abc"},
					}},
				},
			},
			nextPath: donor.PathChooseAttorneysAddress,
		},
		true: {
			provided: &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
				ReplacementAttorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{{
						FirstNames: "John",
						UID:        uid,
						Address:    place.Address{Line1: "abc"},
					}},
				},
			},
			nextPath: donor.PathChooseReplacementAttorneysAddress,
		},
	}

	for isReplacement, tc := range testcases {
		t.Run(fmt.Sprint(isReplacement), func(t *testing.T) {
			form := url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"email":               {"john@example.com"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			service := newMockAttorneyService(t)
			service.EXPECT().
				IsReplacement().
				Return(isReplacement)
			service.EXPECT().
				Put(r.Context(), tc.provided, attorney).
				Return(nil)
			service.EXPECT().
				CanAddTrustCorporation(tc.provided).
				Return(false)

			err := EnterAttorney(nil, service)(testAppData, w, r, tc.provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.nextPath.Format("lpa-id")+"?id="+uid.String(), resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterAttorneyWhenDOBWarning(t *testing.T) {
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

	service := testAttorneyService(t)
	service.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	appData := appcontext.Data{Page: "/abc"}
	err := EnterAttorney(nil, service)(appData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathWarningInterruption.FormatQuery("lpa-id", url.Values{
		"id":          {testUID.String()},
		"warningFrom": {"/abc"},
		"next": {donor.PathChooseAttorneysAddress.FormatQuery(
			"lpa-id",
			url.Values{"id": {testUID.String()}}),
		},
		"actor": {actor.TypeAttorney.String()},
	}), resp.Header.Get("Location"))
}

func TestPostEnterAttorneyWhenNameWarning(t *testing.T) {
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

	service := testAttorneyService(t)
	service.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	appData := appcontext.Data{Page: "/abc"}
	err := EnterAttorney(nil, service)(appData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, donor.PathWarningInterruption.FormatQuery("lpa-id", url.Values{
		"id":          {testUID.String()},
		"warningFrom": {"/abc"},
		"next": {donor.PathChooseAttorneysAddress.FormatQuery(
			"lpa-id",
			url.Values{"id": {testUID.String()}}),
		},
		"actor": {actor.TypeAttorney.String()},
	}), resp.Header.Get("Location"))
}

func TestPostEnterAttorneyWhenServiceErrors(t *testing.T) {
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

	service := testAttorneyService(t)
	service.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterAttorney(nil, service)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestReadEnterAttorneyForm(t *testing.T) {
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

	result := readEnterAttorneyForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("john@example.com", result.Email)
	assert.Equal(date.New("1990", "1", "2"), result.Dob)
}

func TestEnterAttorneyFormValidate(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form   *enterAttorneyForm
		errors validation.List
	}{
		"valid": {
			form: &enterAttorneyForm{
				FirstNames: "A",
				LastName:   "B",
				Dob:        validDob,
			},
		},
		"max length": {
			form: &enterAttorneyForm{
				FirstNames: strings.Repeat("x", 53),
				LastName:   strings.Repeat("x", 61),
				Dob:        validDob,
			},
		},
		"missing all": {
			form: &enterAttorneyForm{},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}).
				With("date-of-birth", validation.EnterError{Label: "dateOfBirth"}),
		},
		"too long": {
			form: &enterAttorneyForm{
				FirstNames: strings.Repeat("x", 54),
				LastName:   strings.Repeat("x", 62),
				Dob:        validDob,
			},
			errors: validation.
				With("first-names", validation.StringTooLongError{Label: "firstNames", Length: 53}).
				With("last-name", validation.StringTooLongError{Label: "lastName", Length: 61}),
		},
		"future dob": {
			form: &enterAttorneyForm{
				FirstNames: "A",
				LastName:   "B",
				Dob:        now.AddDate(0, 0, 1),
			},
			errors: validation.With("date-of-birth", validation.DateMustBePastError{Label: "dateOfBirth"}),
		},
		"invalid dob": {
			form: &enterAttorneyForm{
				FirstNames: "A",
				LastName:   "B",
				Dob:        date.New("2000", "22", "2"),
			},
			errors: validation.With("date-of-birth", validation.DateMustBeRealError{Label: "dateOfBirth"}),
		},
		"invalid missing dob": {
			form: &enterAttorneyForm{
				FirstNames: "A",
				LastName:   "B",
				Dob:        date.New("1", "", "1"),
			},
			errors: validation.With("date-of-birth", validation.DateMissingError{Label: "dateOfBirth", MissingMonth: true}),
		},
		"invalid email": {
			form: &enterAttorneyForm{
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
