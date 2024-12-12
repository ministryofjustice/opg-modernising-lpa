package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSign(t *testing.T) {
	signedAt := time.Now()

	testcases := map[string]struct {
		appData appcontext.Data
		lpa     *lpadata.Lpa
		data    *signData
	}{
		"attorney use when registered": {
			appData: testAppData,
			lpa: &lpadata.Lpa{
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				WhenCanTheLpaBeUsed:              lpadata.CanBeUsedWhenHasCapacity,
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{
					{UID: testUID, FirstNames: "Bob", LastName: "Smith"},
					{UID: actoruid.New(), FirstNames: "Dave", LastName: "Smith"},
				}},
				CertificateProvider: lpadata.CertificateProvider{
					SignedAt: &signedAt,
				},
			},
			data: &signData{
				App:                         testAppData,
				Form:                        &signForm{},
				Attorney:                    lpadata.Attorney{UID: testUID, FirstNames: "Bob", LastName: "Smith"},
				LpaCanBeUsedWhenHasCapacity: true,
			},
		},
		"attorney use when capacity lost": {
			appData: testAppData,
			lpa: &lpadata.Lpa{
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				WhenCanTheLpaBeUsed:              lpadata.CanBeUsedWhenCapacityLost,
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{
					{UID: testUID, FirstNames: "Bob", LastName: "Smith"},
					{UID: actoruid.New(), FirstNames: "Dave", LastName: "Smith"},
				}},
				CertificateProvider: lpadata.CertificateProvider{
					SignedAt: &signedAt,
				},
			},
			data: &signData{
				App:      testAppData,
				Form:     &signForm{},
				Attorney: lpadata.Attorney{UID: testUID, FirstNames: "Bob", LastName: "Smith"},
			},
		},
		"replacement attorney use when registered": {
			appData: testReplacementAppData,
			lpa: &lpadata.Lpa{
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				WhenCanTheLpaBeUsed:              lpadata.CanBeUsedWhenHasCapacity,
				ReplacementAttorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{
					{UID: testUID, FirstNames: "Bob", LastName: "Smith"},
					{UID: actoruid.New(), FirstNames: "Dave", LastName: "Smith"},
				}},
				CertificateProvider: lpadata.CertificateProvider{
					SignedAt: &signedAt,
				},
			},
			data: &signData{
				App:                         testReplacementAppData,
				Form:                        &signForm{},
				Attorney:                    lpadata.Attorney{UID: testUID, FirstNames: "Bob", LastName: "Smith"},
				IsReplacement:               true,
				LpaCanBeUsedWhenHasCapacity: true,
			},
		},
		"replacement attorney use when capacity lost": {
			appData: testReplacementAppData,
			lpa: &lpadata.Lpa{
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				WhenCanTheLpaBeUsed:              lpadata.CanBeUsedWhenCapacityLost,
				ReplacementAttorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{
					{UID: testUID, FirstNames: "Bob", LastName: "Smith"},
					{UID: actoruid.New(), FirstNames: "Dave", LastName: "Smith"},
				}},
				CertificateProvider: lpadata.CertificateProvider{
					SignedAt: &signedAt,
				},
			},
			data: &signData{
				App:           testReplacementAppData,
				Form:          &signForm{},
				Attorney:      lpadata.Attorney{UID: testUID, FirstNames: "Bob", LastName: "Smith"},
				IsReplacement: true,
			},
		},
		"trust corporation": {
			appData: testTrustCorporationAppData,
			lpa: &lpadata.Lpa{
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				WhenCanTheLpaBeUsed:              lpadata.CanBeUsedWhenHasCapacity,
				Attorneys: lpadata.Attorneys{TrustCorporation: lpadata.TrustCorporation{
					Name: "Corp",
				}},
				CertificateProvider: lpadata.CertificateProvider{
					SignedAt: &signedAt,
				},
			},
			data: &signData{
				App:                         testTrustCorporationAppData,
				Form:                        &signForm{},
				TrustCorporation:            lpadata.TrustCorporation{Name: "Corp"},
				LpaCanBeUsedWhenHasCapacity: true,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, tc.data).
				Return(nil)

			err := Sign(template.Execute, nil, nil, nil)(tc.appData, w, r, &attorneydata.Provided{}, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetSignWhenSigned(t *testing.T) {
	testcases := map[string]*attorneydata.Provided{
		"attorney": {
			LpaID:    "lpa-id",
			SignedAt: time.Now(),
		},
		"trust corporation": {
			LpaID:                    "lpa-id",
			IsTrustCorporation:       true,
			WouldLikeSecondSignatory: form.No,
			AuthorisedSignatories: [2]attorneydata.TrustCorporationSignatory{
				{SignedAt: time.Now()},
			},
		},
	}

	for name, attorneyProvidedDetails := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			err := Sign(nil, nil, nil, nil)(testAppData, w, r, attorneyProvidedDetails, &lpadata.Lpa{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, attorney.PathWhatHappensNext.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestGetSignCantSignYet(t *testing.T) {
	uid := actoruid.New()
	signedAt := time.Now()

	testcases := map[string]struct {
		appData appcontext.Data
		lpa     *lpadata.Lpa
	}{
		"submitted but not certified": {
			appData: testAppData,
			lpa: &lpadata.Lpa{
				SignedAt: time.Now(),
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{
					{UID: uid, FirstNames: "Bob", LastName: "Smith"},
					{UID: actoruid.New(), FirstNames: "Dave", LastName: "Smith"},
				}},
			},
		},
		"certified but not submitted": {
			appData: testAppData,
			lpa: &lpadata.Lpa{
				WhenCanTheLpaBeUsed: lpadata.CanBeUsedWhenCapacityLost,
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{
					{UID: uid, FirstNames: "Bob", LastName: "Smith"},
					{UID: actoruid.New(), FirstNames: "Dave", LastName: "Smith"},
				}},
				CertificateProvider: lpadata.CertificateProvider{
					SignedAt: &signedAt,
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			err := Sign(nil, nil, nil, nil)(tc.appData, w, r, &attorneydata.Provided{LpaID: "lpa-id"}, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, attorney.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestGetSignWhenAttorneyDoesNotExist(t *testing.T) {
	uid := actoruid.New()
	signedAt := time.Now()

	testcases := map[string]struct {
		appData appcontext.Data
		lpa     *lpadata.Lpa
	}{
		"attorney": {
			appData: testAppData,
			lpa: &lpadata.Lpa{
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				ReplacementAttorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{
					{UID: uid, FirstNames: "Bob", LastName: "Smith"},
				}},
				CertificateProvider: lpadata.CertificateProvider{
					SignedAt: &signedAt,
				},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			lpa: &lpadata.Lpa{
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{
					{UID: uid, FirstNames: "Bob", LastName: "Smith"},
				}},
				CertificateProvider: lpadata.CertificateProvider{
					SignedAt: &signedAt,
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			err := Sign(nil, nil, nil, nil)(tc.appData, w, r, &attorneydata.Provided{}, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.PathAttorneyStart.Format(), resp.Header.Get("Location"))
		})
	}
}

func TestGetSignOnTemplateError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	signedAt := time.Now()

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := Sign(template.Execute, nil, nil, nil)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{
		SignedAt:                         time.Now(),
		WitnessedByCertificateProviderAt: time.Now(),
		Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: testUID}}},
		CertificateProvider:              lpadata.CertificateProvider{SignedAt: &signedAt},
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostSign(t *testing.T) {
	lpaSignedAt := time.Now().Add(-time.Minute)
	now := time.Now()

	testcases := map[string]struct {
		url             string
		appData         appcontext.Data
		form            url.Values
		provided        *attorneydata.Provided
		lpa             *lpadata.Lpa
		updatedAttorney *attorneydata.Provided
	}{
		"attorney": {
			appData:  testAppData,
			form:     url.Values{"confirm": {"1"}},
			provided: &attorneydata.Provided{UID: testUID, LpaID: "lpa-id"},
			lpa: &lpadata.Lpa{
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: testUID, FirstNames: "Bob", LastName: "Smith"}}},
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &lpaSignedAt},
			},
			updatedAttorney: &attorneydata.Provided{
				UID:      testUID,
				LpaID:    "lpa-id",
				SignedAt: now,
				Tasks:    attorneydata.Tasks{SignTheLpa: task.StateCompleted},
			},
		},
		"attorney with set mobile": {
			appData: testAppData,
			form:    url.Values{"confirm": {"1"}},
			provided: &attorneydata.Provided{
				UID:      testUID,
				LpaID:    "lpa-id",
				Phone:    "0777",
				PhoneSet: true,
			},
			lpa: &lpadata.Lpa{
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: testUID, FirstNames: "Bob", LastName: "Smith", Mobile: "0888"}}},
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &lpaSignedAt},
			},
			updatedAttorney: &attorneydata.Provided{
				UID:      testUID,
				LpaID:    "lpa-id",
				SignedAt: now,
				Phone:    "0777",
				PhoneSet: true,
				Tasks:    attorneydata.Tasks{SignTheLpa: task.StateCompleted},
			},
		},
		"attorney with removed mobile": {
			appData: testAppData,
			form:    url.Values{"confirm": {"1"}},
			provided: &attorneydata.Provided{
				UID:      testUID,
				LpaID:    "lpa-id",
				PhoneSet: true,
			},
			lpa: &lpadata.Lpa{
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: testUID, FirstNames: "Bob", LastName: "Smith", Mobile: "0888"}}},
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &lpaSignedAt},
			},
			updatedAttorney: &attorneydata.Provided{
				UID:      testUID,
				LpaID:    "lpa-id",
				SignedAt: now,
				PhoneSet: true,
				Tasks:    attorneydata.Tasks{SignTheLpa: task.StateCompleted},
			},
		},
		"attorney with donor provided mobile": {
			appData: testAppData,
			form:    url.Values{"confirm": {"1"}},
			provided: &attorneydata.Provided{
				UID:   testUID,
				LpaID: "lpa-id",
			},
			lpa: &lpadata.Lpa{
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: testUID, FirstNames: "Bob", LastName: "Smith", Mobile: "0888"}}},
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &lpaSignedAt},
			},
			updatedAttorney: &attorneydata.Provided{
				UID:      testUID,
				LpaID:    "lpa-id",
				SignedAt: now,
				Phone:    "0888",
				Tasks:    attorneydata.Tasks{SignTheLpa: task.StateCompleted},
			},
		},
		"replacement attorney": {
			appData:  testReplacementAppData,
			form:     url.Values{"confirm": {"1"}},
			provided: &attorneydata.Provided{UID: testUID, LpaID: "lpa-id"},
			lpa: &lpadata.Lpa{
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				ReplacementAttorneys:             lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: testUID, FirstNames: "Bob", LastName: "Smith"}}},
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &lpaSignedAt},
			},
			updatedAttorney: &attorneydata.Provided{
				UID:      testUID,
				LpaID:    "lpa-id",
				SignedAt: now,
				Tasks:    attorneydata.Tasks{SignTheLpa: task.StateCompleted},
			},
		},
		"second trust corporation": {
			url:     "/?second",
			appData: testTrustCorporationAppData,
			form: url.Values{
				"first-names":        {"a"},
				"last-name":          {"b"},
				"professional-title": {"c"},
				"confirm":            {"1"},
			},
			provided: &attorneydata.Provided{UID: testUID, LpaID: "lpa-id"},
			lpa: &lpadata.Lpa{
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Attorneys:                        lpadata.Attorneys{TrustCorporation: lpadata.TrustCorporation{UID: testUID, Name: "Corp"}},
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &lpaSignedAt},
			},
			updatedAttorney: &attorneydata.Provided{
				UID:   testUID,
				LpaID: "lpa-id",
				AuthorisedSignatories: [2]attorneydata.TrustCorporationSignatory{{}, {
					FirstNames:        "a",
					LastName:          "b",
					ProfessionalTitle: "c",
					SignedAt:          now,
				}},
				Tasks: attorneydata.Tasks{SignTheLpaSecond: task.StateCompleted},
			},
		},
		"second replacment trust corporation": {
			url:     "/?second",
			appData: testReplacementTrustCorporationAppData,
			form: url.Values{
				"first-names":        {"a"},
				"last-name":          {"b"},
				"professional-title": {"c"},
				"confirm":            {"1"},
			},
			provided: &attorneydata.Provided{UID: testUID, LpaID: "lpa-id"},
			lpa: &lpadata.Lpa{
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				ReplacementAttorneys:             lpadata.Attorneys{TrustCorporation: lpadata.TrustCorporation{UID: testUID, Name: "Corp"}},
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &lpaSignedAt},
			},
			updatedAttorney: &attorneydata.Provided{
				UID:   testUID,
				LpaID: "lpa-id",
				AuthorisedSignatories: [2]attorneydata.TrustCorporationSignatory{{}, {
					FirstNames:        "a",
					LastName:          "b",
					ProfessionalTitle: "c",
					SignedAt:          now,
				}},
				Tasks: attorneydata.Tasks{SignTheLpaSecond: task.StateCompleted},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, tc.url, strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			w := httptest.NewRecorder()

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				Put(r.Context(), tc.updatedAttorney).
				Return(nil)

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				SendAttorney(r.Context(), tc.lpa, tc.updatedAttorney).
				Return(nil)

			err := Sign(nil, attorneyStore, lpaStoreClient, func() time.Time { return now })(tc.appData, w, r, tc.provided, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, attorney.PathWhatHappensNext.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostSignWhenSignedInLpaStore(t *testing.T) {
	lpaSignedAt := time.Now().Add(-time.Minute)
	now := time.Now()
	attorneySignedAt := time.Now().Add(-time.Hour)

	testcases := map[string]struct {
		url             string
		appData         appcontext.Data
		form            url.Values
		lpa             *lpadata.Lpa
		updatedAttorney *attorneydata.Provided
	}{
		"attorney": {
			appData: testAppData,
			form:    url.Values{"confirm": {"1"}},
			lpa: &lpadata.Lpa{
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: testUID, FirstNames: "Bob", LastName: "Smith", SignedAt: &attorneySignedAt}}},
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &lpaSignedAt},
			},
			updatedAttorney: &attorneydata.Provided{
				LpaID:    "lpa-id",
				SignedAt: attorneySignedAt,
				Tasks:    attorneydata.Tasks{SignTheLpa: task.StateCompleted},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			form:    url.Values{"confirm": {"1"}},
			lpa: &lpadata.Lpa{
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				ReplacementAttorneys:             lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: testUID, FirstNames: "Bob", LastName: "Smith", SignedAt: &attorneySignedAt}}},
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &lpaSignedAt},
			},
			updatedAttorney: &attorneydata.Provided{
				LpaID:    "lpa-id",
				SignedAt: attorneySignedAt,
				Tasks:    attorneydata.Tasks{SignTheLpa: task.StateCompleted},
			},
		},
		"second trust corporation": {
			url:     "/?second",
			appData: testTrustCorporationAppData,
			form: url.Values{
				"first-names":        {"a"},
				"last-name":          {"b"},
				"professional-title": {"c"},
				"confirm":            {"1"},
			},
			lpa: &lpadata.Lpa{
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Attorneys:                        lpadata.Attorneys{TrustCorporation: lpadata.TrustCorporation{Name: "Corp", Signatories: []lpadata.TrustCorporationSignatory{{}, {SignedAt: attorneySignedAt}}}},
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &lpaSignedAt},
			},
			updatedAttorney: &attorneydata.Provided{
				LpaID: "lpa-id",
				AuthorisedSignatories: [2]attorneydata.TrustCorporationSignatory{{}, {
					FirstNames:        "a",
					LastName:          "b",
					ProfessionalTitle: "c",
					SignedAt:          attorneySignedAt,
				}},
				Tasks: attorneydata.Tasks{SignTheLpaSecond: task.StateCompleted},
			},
		},
		"second replacment trust corporation": {
			url:     "/?second",
			appData: testReplacementTrustCorporationAppData,
			form: url.Values{
				"first-names":        {"a"},
				"last-name":          {"b"},
				"professional-title": {"c"},
				"confirm":            {"1"},
			},
			lpa: &lpadata.Lpa{
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				ReplacementAttorneys:             lpadata.Attorneys{TrustCorporation: lpadata.TrustCorporation{Name: "Corp", Signatories: []lpadata.TrustCorporationSignatory{{}, {SignedAt: attorneySignedAt}}}},
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &lpaSignedAt},
			},
			updatedAttorney: &attorneydata.Provided{
				LpaID: "lpa-id",
				AuthorisedSignatories: [2]attorneydata.TrustCorporationSignatory{{}, {
					FirstNames:        "a",
					LastName:          "b",
					ProfessionalTitle: "c",
					SignedAt:          attorneySignedAt,
				}},
				Tasks: attorneydata.Tasks{SignTheLpaSecond: task.StateCompleted},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, tc.url, strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			w := httptest.NewRecorder()

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				Put(r.Context(), tc.updatedAttorney).
				Return(nil)

			err := Sign(nil, attorneyStore, nil, func() time.Time { return now })(tc.appData, w, r, &attorneydata.Provided{LpaID: "lpa-id"}, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, attorney.PathWhatHappensNext.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostSignWhenWantSecondSignatory(t *testing.T) {
	lpaSignedAt := time.Now().Add(-time.Minute)
	now := time.Now()

	testcases := map[string]struct {
		url             string
		appData         appcontext.Data
		form            url.Values
		lpa             *lpadata.Lpa
		updatedAttorney *attorneydata.Provided
	}{
		"trust corporation": {
			appData: testTrustCorporationAppData,
			form: url.Values{
				"first-names":        {"a"},
				"last-name":          {"b"},
				"professional-title": {"c"},
				"confirm":            {"1"},
			},
			lpa: &lpadata.Lpa{
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Attorneys:                        lpadata.Attorneys{TrustCorporation: lpadata.TrustCorporation{Name: "Corp"}},
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &lpaSignedAt},
			},
			updatedAttorney: &attorneydata.Provided{
				LpaID: "lpa-id",
				AuthorisedSignatories: [2]attorneydata.TrustCorporationSignatory{{
					FirstNames:        "a",
					LastName:          "b",
					ProfessionalTitle: "c",
					SignedAt:          now,
				}},
				Tasks: attorneydata.Tasks{SignTheLpa: task.StateCompleted},
			},
		},
		"replacement trust corporation": {
			appData: testReplacementTrustCorporationAppData,
			form: url.Values{
				"first-names":        {"a"},
				"last-name":          {"b"},
				"professional-title": {"c"},
				"confirm":            {"1"},
			},
			lpa: &lpadata.Lpa{
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				ReplacementAttorneys:             lpadata.Attorneys{TrustCorporation: lpadata.TrustCorporation{Name: "Corp"}},
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &lpaSignedAt},
			},
			updatedAttorney: &attorneydata.Provided{
				LpaID: "lpa-id",
				AuthorisedSignatories: [2]attorneydata.TrustCorporationSignatory{{
					FirstNames:        "a",
					LastName:          "b",
					ProfessionalTitle: "c",
					SignedAt:          now,
				}},
				Tasks: attorneydata.Tasks{SignTheLpa: task.StateCompleted},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, tc.url, strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			w := httptest.NewRecorder()

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				Put(r.Context(), tc.updatedAttorney).
				Return(nil)

			err := Sign(nil, attorneyStore, nil, func() time.Time { return now })(tc.appData, w, r, &attorneydata.Provided{LpaID: "lpa-id"}, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, attorney.PathWouldLikeSecondSignatory.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostSignWhenLpaStoreClientErrors(t *testing.T) {
	form := url.Values{"confirm": {"1"}}
	signedAt := time.Now()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendAttorney(r.Context(), mock.Anything, mock.Anything).
		Return(expectedError)

	err := Sign(nil, nil, lpaStoreClient, time.Now)(testAppData, w, r, &attorneydata.Provided{LpaID: "lpa-id"}, &lpadata.Lpa{
		SignedAt:                         time.Now(),
		WitnessedByCertificateProviderAt: time.Now(),
		Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: testUID, FirstNames: "Bob", LastName: "Smith"}}},
		CertificateProvider:              lpadata.CertificateProvider{SignedAt: &signedAt},
	})
	assert.Equal(t, expectedError, err)
}

func TestPostSignWhenStoreError(t *testing.T) {
	form := url.Values{
		"confirm": {"1"},
	}
	signedAt := time.Now()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendAttorney(r.Context(), mock.Anything, mock.Anything).
		Return(nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := Sign(nil, attorneyStore, lpaStoreClient, time.Now)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{
		SignedAt:                         time.Now(),
		WitnessedByCertificateProviderAt: time.Now(),
		Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: testUID, FirstNames: "Bob", LastName: "Smith"}}},
		CertificateProvider:              lpadata.CertificateProvider{SignedAt: &signedAt},
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostSignOnValidationError(t *testing.T) {
	form := url.Values{}
	signedAt := time.Now()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &signData{
			App:      testAppData,
			Form:     &signForm{},
			Attorney: lpadata.Attorney{UID: testUID, FirstNames: "Bob", LastName: "Smith"},
			Errors:   validation.With("confirm", validation.CustomError{Label: "youMustSelectTheBoxToSignAttorney"}),
		}).
		Return(nil)

	err := Sign(template.Execute, nil, nil, time.Now)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{
		SignedAt:                         time.Now(),
		WitnessedByCertificateProviderAt: time.Now(),
		Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: testUID, FirstNames: "Bob", LastName: "Smith"}}},
		CertificateProvider:              lpadata.CertificateProvider{SignedAt: &signedAt},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadSignForm(t *testing.T) {
	form := url.Values{
		"confirm": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	assert.Equal(t, &signForm{Confirm: true}, readSignForm(r))
}

func TestValidateSignForm(t *testing.T) {
	testCases := map[string]struct {
		form               signForm
		isTrustCorporation bool
		isReplacement      bool
		errors             validation.List
	}{
		"true for attorney": {
			form: signForm{
				Confirm: true,
			},
		},
		"true for replacement attorney": {
			form: signForm{
				Confirm: true,
			},
			isReplacement: true,
		},
		"true for trust corporation": {
			form: signForm{
				FirstNames:        "a",
				LastName:          "b",
				ProfessionalTitle: "c",
				Confirm:           true,
			},
			isTrustCorporation: true,
		},
		"false for attorney": {
			form:   signForm{},
			errors: validation.With("confirm", validation.CustomError{Label: "youMustSelectTheBoxToSignAttorney"}),
		},
		"false for replacement attorney": {
			form:          signForm{},
			errors:        validation.With("confirm", validation.CustomError{Label: "youMustSelectTheBoxToSignReplacementAttorney"}),
			isReplacement: true,
		},
		"empty trust corporation": {
			form: signForm{},
			errors: validation.With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}).
				With("professional-title", validation.EnterError{Label: "professionalTitle"}).
				With("confirm", validation.CustomError{Label: "youMustSelectTheBoxToSignAttorney"}),
			isTrustCorporation: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate(tc.isTrustCorporation, tc.isReplacement))
		})
	}
}
