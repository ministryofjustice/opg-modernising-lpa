package certificateprovider

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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterDateOfBirth(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpastore.Lpa{
		LpaID: "lpa-id",
	}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dateOfBirthData{
			App:  testAppData,
			Lpa:  donor,
			Form: &dateOfBirthForm{},
		}).
		Return(nil)

	err := EnterDateOfBirth(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &actor.CertificateProviderProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterDateOfBirthFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dateOfBirthData{
			App: testAppData,
			Lpa: &lpastore.Lpa{},
			Form: &dateOfBirthForm{
				Dob: date.New("1997", "1", "2"),
			},
		}).
		Return(nil)

	err := EnterDateOfBirth(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &actor.CertificateProviderProvidedDetails{DateOfBirth: date.New("1997", "1", "2")})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterDateOfBirthWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, expectedError)

	err := EnterDateOfBirth(nil, lpaStoreResolvingService, nil)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterDateOfBirthWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpastore.Lpa{
		LpaID: "lpa-id",
	}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dateOfBirthData{
			App:  testAppData,
			Lpa:  donor,
			Form: &dateOfBirthForm{},
		}).
		Return(expectedError)

	err := EnterDateOfBirth(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &actor.CertificateProviderProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterDateOfBirth(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form      url.Values
		retrieved *actor.CertificateProviderProvidedDetails
		updated   *actor.CertificateProviderProvidedDetails
	}{
		"valid": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			retrieved: &actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"},
			updated: &actor.CertificateProviderProvidedDetails{
				LpaID:       "lpa-id",
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				Tasks: actor.CertificateProviderTasks{
					ConfirmYourDetails: actor.TaskInProgress,
				},
			},
		},
		"previously completed": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			retrieved: &actor.CertificateProviderProvidedDetails{
				LpaID: "lpa-id",
				Tasks: actor.CertificateProviderTasks{
					ConfirmYourDetails: actor.TaskCompleted,
				},
			},
			updated: &actor.CertificateProviderProvidedDetails{
				LpaID:       "lpa-id",
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				Tasks: actor.CertificateProviderTasks{
					ConfirmYourDetails: actor.TaskCompleted,
				},
			},
		},
		"warning ignored": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
			},
			retrieved: &actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"},
			updated: &actor.CertificateProviderProvidedDetails{
				LpaID:       "lpa-id",
				DateOfBirth: date.New("1900", "1", "2"),
				Tasks: actor.CertificateProviderTasks{
					ConfirmYourDetails: actor.TaskInProgress,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(&lpastore.Lpa{LpaID: "lpa-id"}, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				Put(r.Context(), tc.updated).
				Return(nil)

			err := EnterDateOfBirth(nil, lpaStoreResolvingService, certificateProviderStore)(testAppData, w, r, tc.retrieved)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.CertificateProvider.YourPreferredLanguage.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterDateOfBirthWhenProfessionalCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1980"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{LpaID: "lpa-id", CertificateProvider: lpastore.CertificateProvider{Relationship: actor.Professionally}}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), &actor.CertificateProviderProvidedDetails{
			LpaID:       "lpa-id",
			DateOfBirth: date.New("1980", "1", "2"),
			Tasks: actor.CertificateProviderTasks{
				ConfirmYourDetails: actor.TaskInProgress,
			},
		}).
		Return(nil)

	err := EnterDateOfBirth(nil, lpaStoreResolvingService, certificateProviderStore)(testAppData, w, r, &actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.WhatIsYourHomeAddress.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterDateOfBirthWhenInputRequired(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *dateOfBirthData) bool
	}{
		"validation error": {
			form: url.Values{
				"date-of-birth-day":   {"55"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			dataMatcher: func(t *testing.T, data *dateOfBirthData) bool {
				return assert.Equal(t, validation.With("date-of-birth", validation.DateMustBeRealError{Label: "dateOfBirth"}), data.Errors)
			},
		},
		"dob warning": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
			},
			dataMatcher: func(t *testing.T, data *dateOfBirthData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(&lpastore.Lpa{LpaID: "lpa-id"}, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, mock.MatchedBy(func(data *dateOfBirthData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := EnterDateOfBirth(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &actor.CertificateProviderProvidedDetails{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostYourDetailsWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	form := url.Values{
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1999"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, expectedError)

	err := EnterDateOfBirth(nil, lpaStoreResolvingService, nil)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadDateOfBirthForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
		"ignore-dob-warning":  {"xyz"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readDateOfBirthForm(r)

	assert.Equal(date.New("1990", "1", "2"), result.Dob)
	assert.Equal("xyz", result.IgnoreDobWarning)
}

func TestDateOfBirthFormValidate(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form   *dateOfBirthForm
		errors validation.List
	}{
		"valid": {
			form: &dateOfBirthForm{
				Dob:              validDob,
				IgnoreDobWarning: "xyz",
			},
		},
		"missing": {
			form: &dateOfBirthForm{},
			errors: validation.
				With("date-of-birth", validation.EnterError{Label: "dateOfBirth"}),
		},
		"future-dob": {
			form: &dateOfBirthForm{
				Dob: now.AddDate(0, 0, 1),
			},
			errors: validation.With("date-of-birth", validation.DateMustBePastError{Label: "dateOfBirth"}),
		},
		"dob-under-18": {
			form: &dateOfBirthForm{
				Dob: now.AddDate(0, 0, -1),
			},
			errors: validation.With("date-of-birth", validation.CustomError{Label: "youAreUnder18Error"}),
		},
		"invalid-dob": {
			form: &dateOfBirthForm{
				Dob: date.New("2000", "22", "2"),
			},
			errors: validation.With("date-of-birth", validation.DateMustBeRealError{Label: "dateOfBirth"}),
		},
		"invalid-missing-dob": {
			form: &dateOfBirthForm{
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
