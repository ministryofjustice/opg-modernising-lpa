package attorney

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSign(t *testing.T) {
	testcases := map[string]struct {
		appData            page.AppData
		lpa                *page.Lpa
		isReplacement      bool
		usedWhenRegistered bool
	}{
		"attorney use when registered": {
			appData: testAppData,
			lpa: &page.Lpa{
				Submitted:           time.Now(),
				WhenCanTheLpaBeUsed: page.UsedWhenRegistered,
				Attorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				},
			},
			usedWhenRegistered: true,
		},
		"attorney use when capacity lost": {
			appData: testAppData,
			lpa: &page.Lpa{
				Submitted:           time.Now(),
				WhenCanTheLpaBeUsed: page.UsedWhenCapacityLost,
				Attorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				},
			},
		},
		"replacement attorney use when registered": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				Submitted:           time.Now(),
				WhenCanTheLpaBeUsed: page.UsedWhenRegistered,
				ReplacementAttorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				},
			},
			isReplacement:      true,
			usedWhenRegistered: true,
		},
		"replacement attorney use when capacity lost": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				Submitted:           time.Now(),
				WhenCanTheLpaBeUsed: page.UsedWhenCapacityLost,
				ReplacementAttorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				},
			},
			isReplacement: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			template := newMockTemplate(t)
			template.
				On("Execute", w, &signData{
					App:                        tc.appData,
					Form:                       &signForm{},
					Attorney:                   actor.Attorney{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					IsReplacement:              tc.isReplacement,
					LpaCanBeUsedWhenRegistered: tc.usedWhenRegistered,
				}).
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

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Get", r.Context()).
				Return(&actor.AttorneyProvidedDetails{}, nil)

			err := Sign(template.Execute, donorStore, certificateProviderStore, attorneyStore)(tc.appData, w, r)
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
				Attorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
		},
		"certified but not submitted": {
			appData: testAppData,
			lpa: &page.Lpa{
				WhenCanTheLpaBeUsed: page.UsedWhenCapacityLost,
				Attorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				},
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

			err := Sign(nil, donorStore, certificateProviderStore, nil)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.TaskList, resp.Header.Get("Location"))
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
				ReplacementAttorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
				},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				Submitted: time.Now(),
				Attorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
				},
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

			err := Sign(nil, donorStore, certificateProviderStore, nil)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.Start, resp.Header.Get("Location"))
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

	err := Sign(template.Execute, donorStore, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetSignOnAttorneyStoreError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{
			Submitted: time.Now(),
			Attorneys: actor.Attorneys{{ID: "attorney-id"}},
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{
			Certificate: actor.Certificate{Agreed: time.Now()},
		}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, expectedError)

	err := Sign(nil, donorStore, certificateProviderStore, attorneyStore)(testAppData, w, r)
	assert.Equal(t, expectedError, err)
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
			Attorneys: actor.Attorneys{{ID: "attorney-id"}},
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{
			Certificate: actor.Certificate{Agreed: time.Now()},
		}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)

	err := Sign(template.Execute, donorStore, certificateProviderStore, attorneyStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostSign(t *testing.T) {
	now := time.Now()

	testcases := map[string]struct {
		appData         page.AppData
		lpa             *page.Lpa
		updatedAttorney *actor.AttorneyProvidedDetails
	}{
		"attorney": {
			appData: testAppData,
			lpa: &page.Lpa{
				Submitted: now,
				Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				Confirmed: true,
				Tasks:     actor.AttorneyTasks{SignTheLpa: actor.TaskCompleted},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				Submitted:            now,
				ReplacementAttorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				Confirmed: true,
				Tasks:     actor.AttorneyTasks{SignTheLpa: actor.TaskCompleted},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"confirm": {"1"},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			w := httptest.NewRecorder()

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(tc.lpa, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Get", r.Context()).
				Return(&actor.AttorneyProvidedDetails{}, nil)
			attorneyStore.
				On("Put", r.Context(), tc.updatedAttorney).
				Return(nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.
				On("GetAny", mock.Anything).
				Return(&actor.CertificateProviderProvidedDetails{
					Certificate: actor.Certificate{Agreed: time.Now()},
				}, nil)

			err := Sign(nil, donorStore, certificateProviderStore, attorneyStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.WhatHappensNext, resp.Header.Get("Location"))
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
			Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{
			Certificate: actor.Certificate{Agreed: time.Now()},
		}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)
	attorneyStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := Sign(nil, donorStore, certificateProviderStore, attorneyStore)(testAppData, w, r)
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
			Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{
			Certificate: actor.Certificate{Agreed: time.Now()},
		}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &signData{
			App:      testAppData,
			Form:     &signForm{},
			Attorney: actor.Attorney{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
			Errors:   validation.With("confirm", validation.CustomError{Label: "youMustSelectTheBoxToSignAttorney"}),
		}).
		Return(nil)

	err := Sign(template.Execute, donorStore, certificateProviderStore, attorneyStore)(testAppData, w, r)
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
		form          signForm
		isReplacement bool
		errors        validation.List
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
		"false for attorney": {
			form:   signForm{},
			errors: validation.With("confirm", validation.CustomError{Label: "youMustSelectTheBoxToSignAttorney"}),
		},
		"false for replacement attorney": {
			form:          signForm{},
			errors:        validation.With("confirm", validation.CustomError{Label: "youMustSelectTheBoxToSignReplacementAttorney"}),
			isReplacement: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate(tc.isReplacement))
		})
	}
}
