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

func TestGetEnterReplacementAttorney(t *testing.T) {
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
				Execute(w, &enterReplacementAttorneyData{
					App: testAppData,
					Donor: &donordata.Provided{
						Type:      tc.lpaType,
						Attorneys: tc.attorneys,
					},
					Form:                     &enterAttorneyForm{},
					ShowTrustCorporationLink: tc.expectedShowTrustCorpLink,
				}).
				Return(nil)

			err := EnterReplacementAttorney(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
				Type:      tc.lpaType,
				Attorneys: tc.attorneys,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetEnterReplacementAttorneyWhenNoID(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := EnterReplacementAttorney(nil, nil, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{FirstNames: "John", UID: actoruid.New()}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetEnterReplacementAttorneyWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+testUID.String(), nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := EnterReplacementAttorney(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReplacementAttorneyAttorneyDoesNotExists(t *testing.T) {
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

			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				PutAttorney(r.Context(), tc.attorney).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:                "lpa-id",
					Donor:                donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
					ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{tc.attorney}},
					Tasks:                donordata.Tasks{ChooseReplacementAttorneys: task.StateInProgress},
				}).
				Return(nil)

			err := EnterReplacementAttorney(nil, donorStore, reuseStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathChooseReplacementAttorneysAddress.Format("lpa-id")+"?id="+testUID.String(), resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterReplacementAttorneyAttorneyExists(t *testing.T) {
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

			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				PutAttorney(r.Context(), tc.attorney).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:                "lpa-id",
					Donor:                donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
					ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{tc.attorney}},
					Tasks:                donordata.Tasks{ChooseReplacementAttorneys: task.StateCompleted},
				}).
				Return(nil)

			err := EnterReplacementAttorney(nil, donorStore, reuseStore)(testAppData, w, r, &donordata.Provided{
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

func TestPostEnterReplacementAttorneyWhenDOBWarning(t *testing.T) {
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

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutAttorney(mock.Anything, mock.Anything).
		Return(nil)

	appData := appcontext.Data{Page: "/abc"}
	err := EnterReplacementAttorney(nil, donorStore, reuseStore)(appData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathWarningInterruption.FormatQuery("lpa-id", url.Values{
		"id":          {testUID.String()},
		"warningFrom": {"/abc"},
		"next": {donor.PathChooseReplacementAttorneysAddress.FormatQuery(
			"lpa-id",
			url.Values{"id": {testUID.String()}}),
		},
		"actor": {actor.TypeReplacementAttorney.String()},
	}), resp.Header.Get("Location"))
}

func TestPostEnterReplacementAttorneyWhenNameWarning(t *testing.T) {
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

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutAttorney(mock.Anything, mock.Anything).
		Return(nil)

	appData := appcontext.Data{Page: "/abc"}
	err := EnterReplacementAttorney(nil, donorStore, reuseStore)(appData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathWarningInterruption.FormatQuery("lpa-id", url.Values{
		"id":          {testUID.String()},
		"warningFrom": {"/abc"},
		"next": {donor.PathChooseReplacementAttorneysAddress.FormatQuery(
			"lpa-id",
			url.Values{"id": {testUID.String()}}),
		},
		"actor": {actor.TypeReplacementAttorney.String()},
	}), resp.Header.Get("Location"))
}

func TestPostEnterReplacementAttorneyWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *enterReplacementAttorneyData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
			},
			dataMatcher: func(t *testing.T, data *enterReplacementAttorneyData) bool {
				return assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
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
				Execute(w, mock.MatchedBy(func(data *enterReplacementAttorneyData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := EnterReplacementAttorney(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostEnterReplacementAttorneyAttorneyWhenReuseStoreErrors(t *testing.T) {
	uid := actoruid.New()
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

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

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutAttorney(mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterReplacementAttorney(nil, nil, reuseStore)(testAppData, w, r, &donordata.Provided{
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
	assert.Equal(t, expectedError, err)
}

func TestPostEnterReplacementAttorneyWhenDonorStoreErrors(t *testing.T) {
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

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutAttorney(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := EnterReplacementAttorney(nil, donorStore, reuseStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}
