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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourDateOfBirth(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourDateOfBirthData{
			App:  testAppData,
			Form: &yourDateOfBirthForm{},
		}).
		Return(nil)

	err := YourDateOfBirth(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourDateOfBirthFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourDateOfBirthData{
			App: testAppData,
			Form: &yourDateOfBirthForm{
				Dob: date.New("2000", "1", "2"),
			},
		}).
		Return(nil)

	err := YourDateOfBirth(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{
			DateOfBirth: date.New("2000", "1", "2"),
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourDateOfBirthDobWarningIsAlwaysShown(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourDateOfBirthData{
			App: testAppData,
			Form: &yourDateOfBirthForm{
				Dob: date.New("1900", "01", "02"),
			},
		}).
		Return(nil)

	err := YourDateOfBirth(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{
			DateOfBirth: date.New("1900", "01", "02"),
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourDateOfBirthWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := YourDateOfBirth(template.Execute, nil)(testAppData, w, r, &donordata.Provided{Donor: donordata.Donor{FirstNames: "John", DateOfBirth: date.New("2000", "1", "2")}})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourDateOfBirth(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)
	under18BirthYear := strconv.Itoa(time.Now().Year() - 10)
	over100BirthYear := strconv.Itoa(time.Now().Year() - 101)

	testCases := map[string]struct {
		url      string
		form     url.Values
		person   donordata.Donor
		redirect string
	}{
		"valid": {
			url: "/",
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			person: donordata.Donor{
				DateOfBirth: date.New(validBirthYear, "1", "2"),
			},
			redirect: donor.PathDoYouLiveInTheUK.Format("lpa-id"),
		},
		"making another lpa": {
			url: "/?makingAnotherLPA=1",
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			person: donordata.Donor{
				DateOfBirth: date.New(validBirthYear, "1", "2"),
			},
			redirect: donor.PathWeHaveUpdatedYourDetails.Format("lpa-id") + "?detail=dateOfBirth",
		},
		"under 18": {
			url: "/?from=somewhere",
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {under18BirthYear},
			},
			person: donordata.Donor{
				DateOfBirth: date.New(under18BirthYear, "1", "2"),
			},
			redirect: donor.PathYouHaveToldUsYouAreUnder18.FormatQuery("lpa-id", url.Values{"next": {"somewhere"}}),
		},
		"over 100": {
			url: "/?from=somewhere",
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {over100BirthYear},
			},
			person: donordata.Donor{
				DateOfBirth: date.New(over100BirthYear, "1", "2"),
			},
			redirect: donor.PathWarningInterruption.FormatQuery(
				"lpa-id",
				url.Values{
					"warningFrom": {"/abc"},
					"next":        {donor.PathDoYouLiveInTheUK.Format("lpa-id")},
					"actor":       {actor.TypeDonor.String()},
				},
			),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, tc.url, strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID: "lpa-id",
					Donor: tc.person,
				}).
				Return(nil)

			appData := appcontext.Data{Page: "/abc"}
			err := YourDateOfBirth(nil, donorStore)(appData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					DateOfBirth: date.New("2000", "1", "2"),
				},
				HasSentApplicationUpdatedEvent: true,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostYourDateOfBirthWhenDetailsNotChanged(t *testing.T) {
	testcases := map[string]struct {
		url      string
		redirect donor.Path
	}{
		"making first": {
			url:      "/",
			redirect: donor.PathDoYouLiveInTheUK,
		},
		"making another": {
			url:      "/?makingAnotherLPA=1",
			redirect: donor.PathMakeANewLPA,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {

			validBirthYear := strconv.Itoa(time.Now().Year() - 40)
			f := url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			}

			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, tc.url, strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := YourDateOfBirth(nil, nil)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					DateOfBirth: date.New(validBirthYear, "1", "2"),
				},
				HasSentApplicationUpdatedEvent: true,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostYourDateOfBirthWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *yourDateOfBirthData) bool
	}{
		"validation error": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"9999"},
			},
			dataMatcher: func(t *testing.T, data *yourDateOfBirthData) bool {
				return assert.Equal(t, validation.With("date-of-birth", validation.DateMustBePastError{Label: "dateOfBirth"}), data.Errors)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, mock.MatchedBy(func(data *yourDateOfBirthData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := YourDateOfBirth(template.Execute, nil)(testAppData, w, r, &donordata.Provided{Donor: donordata.Donor{DateOfBirth: date.New("2000", "1", "2")}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostYourDateOfBirthWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := YourDateOfBirth(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{
			DateOfBirth: date.New("2000", "1", "2"),
		},
	})

	assert.Equal(t, expectedError, err)
}

func TestReadYourDateOfBirthForm(t *testing.T) {
	assert := assert.New(t)

	f := url.Values{
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readYourDateOfBirthForm(r)

	assert.Equal(date.New("1990", "1", "2"), result.Dob)
}

func TestYourDateOfBirthFormValidate(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form   *yourDateOfBirthForm
		errors validation.List
	}{
		"valid": {
			form: &yourDateOfBirthForm{
				Dob: validDob,
			},
		},
		"missing": {
			form: &yourDateOfBirthForm{},
			errors: validation.
				With("date-of-birth", validation.EnterError{Label: "dateOfBirth"}),
		},
		"future dob": {
			form: &yourDateOfBirthForm{
				Dob: now.AddDate(0, 0, 1),
			},
			errors: validation.With("date-of-birth", validation.DateMustBePastError{Label: "dateOfBirth"}),
		},
		"invalid dob": {
			form: &yourDateOfBirthForm{
				Dob: date.New("2000", "22", "2"),
			},
			errors: validation.With("date-of-birth", validation.DateMustBeRealError{Label: "dateOfBirth"}),
		},
		"invalid missing dob": {
			form: &yourDateOfBirthForm{
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
