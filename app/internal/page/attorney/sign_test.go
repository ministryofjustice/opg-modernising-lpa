package attorney

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSign(t *testing.T) {
	testcases := map[string]struct {
		appData page.AppData
		lpa     *page.Lpa
		data    *signData
	}{
		"attorney use when registered": {
			appData: testAppData,
			lpa: &page.Lpa{
				Submitted:           time.Now(),
				WhenCanTheLpaBeUsed: page.CanBeUsedWhenHasCapacity,
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				}},
			},
			data: &signData{
				App:                         testAppData,
				Form:                        &signForm{},
				Attorney:                    actor.Attorney{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
				LpaCanBeUsedWhenHasCapacity: true,
			},
		},
		"attorney use when capacity lost": {
			appData: testAppData,
			lpa: &page.Lpa{
				Submitted:           time.Now(),
				WhenCanTheLpaBeUsed: page.CanBeUsedWhenCapacityLost,
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				}},
			},
			data: &signData{
				App:      testAppData,
				Form:     &signForm{},
				Attorney: actor.Attorney{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
			},
		},
		"replacement attorney use when registered": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				Submitted:           time.Now(),
				WhenCanTheLpaBeUsed: page.CanBeUsedWhenHasCapacity,
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				}},
			},
			data: &signData{
				App:                         testReplacementAppData,
				Form:                        &signForm{},
				Attorney:                    actor.Attorney{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
				IsReplacement:               true,
				LpaCanBeUsedWhenHasCapacity: true,
			},
		},
		"replacement attorney use when capacity lost": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				Submitted:           time.Now(),
				WhenCanTheLpaBeUsed: page.CanBeUsedWhenCapacityLost,
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				}},
			},
			data: &signData{
				App:           testReplacementAppData,
				Form:          &signForm{},
				Attorney:      actor.Attorney{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
				IsReplacement: true,
			},
		},
		"trust corporation": {
			appData: testTrustCorporationAppData,
			lpa: &page.Lpa{
				Submitted:           time.Now(),
				WhenCanTheLpaBeUsed: page.CanBeUsedWhenHasCapacity,
				Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{
					Name: "Corp",
				}},
			},
			data: &signData{
				App:                         testTrustCorporationAppData,
				Form:                        &signForm{},
				TrustCorporation:            actor.TrustCorporation{Name: "Corp"},
				LpaCanBeUsedWhenHasCapacity: true,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			template := newMockTemplate(t)
			template.
				On("Execute", w, tc.data).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(tc.lpa, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.
				On("GetAny", mock.Anything).
				Return(&actor.CertificateProviderProvidedDetails{
					Certificate: actor.Certificate{Agreed: time.Now()},
				}, nil)

			err := Sign(template.Execute, donorStore, certificateProviderStore, nil, nil)(tc.appData, w, r, &actor.AttorneyProvidedDetails{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetSignCantSignYet(t *testing.T) {
	testcases := map[string]struct {
		appData             page.AppData
		lpa                 *page.Lpa
		certificateProvider *actor.CertificateProviderProvidedDetails
	}{
		"submitted but not certified": {
			appData: testAppData,
			lpa: &page.Lpa{
				Submitted: time.Now(),
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				}},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
		},
		"certified but not submitted": {
			appData: testAppData,
			lpa: &page.Lpa{
				WhenCanTheLpaBeUsed: page.CanBeUsedWhenCapacityLost,
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				}},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{
				Certificate: actor.Certificate{Agreed: time.Now()},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(tc.lpa, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.
				On("GetAny", mock.Anything).
				Return(tc.certificateProvider, nil)

			err := Sign(nil, donorStore, certificateProviderStore, nil, nil)(tc.appData, w, r, &actor.AttorneyProvidedDetails{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestGetSignWhenAttorneyDoesNotExist(t *testing.T) {
	testcases := map[string]struct {
		appData page.AppData
		lpa     *page.Lpa
	}{
		"attorney": {
			appData: testAppData,
			lpa: &page.Lpa{
				Submitted: time.Now(),
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
				}},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				Submitted: time.Now(),
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
				}},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(tc.lpa, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.
				On("GetAny", mock.Anything).
				Return(&actor.CertificateProviderProvidedDetails{
					Certificate: actor.Certificate{Agreed: time.Now()},
				}, nil)

			err := Sign(nil, donorStore, certificateProviderStore, nil, nil)(tc.appData, w, r, &actor.AttorneyProvidedDetails{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.Start.Format(), resp.Header.Get("Location"))
		})
	}
}

func TestGetSignOnDonorStoreError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := Sign(template.Execute, donorStore, nil, nil, nil)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetSignOnTemplateError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{
			Submitted: time.Now(),
			Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "attorney-id"}}},
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{
			Certificate: actor.Certificate{Agreed: time.Now()},
		}, nil)

	err := Sign(template.Execute, donorStore, certificateProviderStore, nil, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostSign(t *testing.T) {
	now := time.Now()

	testcases := map[string]struct {
		url             string
		appData         page.AppData
		form            url.Values
		lpa             *page.Lpa
		updatedAttorney *actor.AttorneyProvidedDetails
		redirect        page.AttorneyPath
	}{
		"attorney": {
			appData: testAppData,
			form:    url.Values{"confirm": {"1"}},
			lpa: &page.Lpa{
				Submitted: now,
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}}},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				LpaID:     "lpa-id",
				Confirmed: now,
				Tasks:     actor.AttorneyTasks{SignTheLpa: actor.TaskCompleted},
			},
			redirect: page.Paths.Attorney.WhatHappensNext,
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			form:    url.Values{"confirm": {"1"}},
			lpa: &page.Lpa{
				Submitted:            now,
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}}},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				LpaID:     "lpa-id",
				Confirmed: now,
				Tasks:     actor.AttorneyTasks{SignTheLpa: actor.TaskCompleted},
			},
			redirect: page.Paths.Attorney.WhatHappensNext,
		},
		"trust corporation": {
			appData: testTrustCorporationAppData,
			form: url.Values{
				"first-names":        {"a"},
				"last-name":          {"b"},
				"professional-title": {"c"},
				"confirm":            {"1"},
			},
			lpa: &page.Lpa{
				Submitted: now,
				Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "Corp"}},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				LpaID: "lpa-id",
				AuthorisedSignatories: [2]actor.AuthorisedSignatory{{
					FirstNames:        "a",
					LastName:          "b",
					ProfessionalTitle: "c",
					Confirmed:         now,
				}},
				Tasks: actor.AttorneyTasks{SignTheLpa: actor.TaskCompleted},
			},
			redirect: page.Paths.Attorney.WouldLikeSecondSignatory,
		},
		"replacement trust corporation": {
			appData: testReplacementTrustCorporationAppData,
			form: url.Values{
				"first-names":        {"a"},
				"last-name":          {"b"},
				"professional-title": {"c"},
				"confirm":            {"1"},
			},
			lpa: &page.Lpa{
				Submitted:            now,
				ReplacementAttorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "Corp"}},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				LpaID: "lpa-id",
				AuthorisedSignatories: [2]actor.AuthorisedSignatory{{
					FirstNames:        "a",
					LastName:          "b",
					ProfessionalTitle: "c",
					Confirmed:         now,
				}},
				Tasks: actor.AttorneyTasks{SignTheLpa: actor.TaskCompleted},
			},
			redirect: page.Paths.Attorney.WouldLikeSecondSignatory,
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
			lpa: &page.Lpa{
				Submitted: now,
				Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "Corp"}},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				LpaID: "lpa-id",
				AuthorisedSignatories: [2]actor.AuthorisedSignatory{{}, {
					FirstNames:        "a",
					LastName:          "b",
					ProfessionalTitle: "c",
					Confirmed:         now,
				}},
				Tasks: actor.AttorneyTasks{SignTheLpaSecond: actor.TaskCompleted},
			},
			redirect: page.Paths.Attorney.WhatHappensNext,
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
			lpa: &page.Lpa{
				Submitted:            now,
				ReplacementAttorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "Corp"}},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				LpaID: "lpa-id",
				AuthorisedSignatories: [2]actor.AuthorisedSignatory{{}, {
					FirstNames:        "a",
					LastName:          "b",
					ProfessionalTitle: "c",
					Confirmed:         now,
				}},
				Tasks: actor.AttorneyTasks{SignTheLpaSecond: actor.TaskCompleted},
			},
			redirect: page.Paths.Attorney.WhatHappensNext,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, tc.url, strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			w := httptest.NewRecorder()

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(tc.lpa, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Put", r.Context(), tc.updatedAttorney).
				Return(nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.
				On("GetAny", mock.Anything).
				Return(&actor.CertificateProviderProvidedDetails{
					Certificate: actor.Certificate{Agreed: time.Now()},
				}, nil)

			err := Sign(nil, donorStore, certificateProviderStore, attorneyStore, func() time.Time { return now })(tc.appData, w, r, &actor.AttorneyProvidedDetails{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostSignWhenStoreError(t *testing.T) {
	form := url.Values{
		"confirm": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	w := httptest.NewRecorder()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{
			Submitted: time.Now(),
			Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}}},
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{
			Certificate: actor.Certificate{Agreed: time.Now()},
		}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := Sign(nil, donorStore, certificateProviderStore, attorneyStore, time.Now)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostSignOnValidationError(t *testing.T) {
	form := url.Values{}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	w := httptest.NewRecorder()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{
			Submitted: time.Now(),
			Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}}},
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{
			Certificate: actor.Certificate{Agreed: time.Now()},
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &signData{
			App:      testAppData,
			Form:     &signForm{},
			Attorney: actor.Attorney{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
			Errors:   validation.With("confirm", validation.CustomError{Label: "youMustSelectTheBoxToSignAttorney"}),
		}).
		Return(nil)

	err := Sign(template.Execute, donorStore, certificateProviderStore, nil, time.Now)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
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
