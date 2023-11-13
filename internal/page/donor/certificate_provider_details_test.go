package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCertificateProviderDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &certificateProviderDetailsData{
			App:  testAppData,
			Form: &certificateProviderDetailsForm{},
		}).
		Return(nil)

	err := CertificateProviderDetails(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCertificateProviderDetailsFromStore(t *testing.T) {
	testcases := map[string]struct {
		lpa  *page.Lpa
		form *certificateProviderDetailsForm
	}{
		"uk mobile": {
			lpa: &page.Lpa{
				CertificateProvider: actor.CertificateProvider{
					FirstNames: "John",
					Mobile:     "07777",
				},
			},
			form: &certificateProviderDetailsForm{
				FirstNames: "John",
				Mobile:     "07777",
			},
		},
		"non-uk mobile": {
			lpa: &page.Lpa{
				CertificateProvider: actor.CertificateProvider{
					FirstNames:     "John",
					Mobile:         "07777",
					HasNonUKMobile: true,
				},
			},
			form: &certificateProviderDetailsForm{
				FirstNames:     "John",
				NonUKMobile:    "07777",
				HasNonUKMobile: true,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &certificateProviderDetailsData{
					App:  testAppData,
					Form: tc.form,
				}).
				Return(nil)

			err := CertificateProviderDetails(template.Execute, nil)(testAppData, w, r, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetCertificateProviderDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &certificateProviderDetailsData{
			App:  testAppData,
			Form: &certificateProviderDetailsForm{},
		}).
		Return(expectedError)

	err := CertificateProviderDetails(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCertificateProviderDetails(t *testing.T) {
	testCases := map[string]struct {
		form                       url.Values
		certificateProviderDetails actor.CertificateProvider
	}{
		"valid": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Rey"},
				"mobile":      {"07535111111"},
			},
			certificateProviderDetails: actor.CertificateProvider{
				FirstNames: "John",
				LastName:   "Rey",
				Mobile:     "07535111111",
			},
		},
		"valid non uk mobile": {
			form: url.Values{
				"first-names":       {"John"},
				"last-name":         {"Rey"},
				"has-non-uk-mobile": {"1"},
				"non-uk-mobile":     {"+337575757"},
			},
			certificateProviderDetails: actor.CertificateProvider{
				FirstNames:     "John",
				LastName:       "Rey",
				Mobile:         "+337575757",
				HasNonUKMobile: true,
			},
		},
		"name warning ignored": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"mobile":              {"07535111111"},
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeCertificateProvider, actor.TypeDonor, "Jane", "Doe").String()},
			},
			certificateProviderDetails: actor.CertificateProvider{
				FirstNames: "Jane",
				LastName:   "Doe",
				Mobile:     "07535111111",
			},
		},
		"similar name warning ignored": {
			form: url.Values{
				"first-names":                 {"Joyce"},
				"last-name":                   {"Doe"},
				"mobile":                      {"07535111111"},
				"ignore-similar-name-warning": {"yes"},
			},
			certificateProviderDetails: actor.CertificateProvider{
				FirstNames: "Joyce",
				LastName:   "Doe",
				Mobile:     "07535111111",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					ID: "lpa-id",
					Donor: actor.Donor{
						FirstNames: "Jane",
						LastName:   "Doe",
					},
					CertificateProvider: tc.certificateProviderDetails,
					Tasks:               page.Tasks{CertificateProvider: actor.TaskInProgress},
				}).
				Return(nil)

			err := CertificateProviderDetails(nil, donorStore)(testAppData, w, r, &page.Lpa{
				ID: "lpa-id",
				Donor: actor.Donor{
					FirstNames: "Jane",
					LastName:   "Doe",
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.HowDoYouKnowYourCertificateProvider.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostCertificateProviderDetailsWhenAmendingDetailsAfterStateComplete(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Rey"},
		"mobile":      {"07535111111"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID: "lpa-id",
			Donor: actor.Donor{
				FirstNames: "Jane",
				LastName:   "Doe",
			},
			CertificateProvider: actor.CertificateProvider{
				FirstNames: "John",
				LastName:   "Rey",
				Mobile:     "07535111111",
			},
			Tasks: page.Tasks{CertificateProvider: actor.TaskCompleted},
		}).
		Return(nil)

	err := CertificateProviderDetails(nil, donorStore)(testAppData, w, r, &page.Lpa{
		ID: "lpa-id",
		Donor: actor.Donor{
			FirstNames: "Jane",
			LastName:   "Doe",
		},
		Tasks: page.Tasks{CertificateProvider: actor.TaskCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.HowDoYouKnowYourCertificateProvider.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostCertificateProviderDetailsWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		existingLpa *page.Lpa
		dataMatcher func(t *testing.T, data *certificateProviderDetailsData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name": {"Doe"},
				"mobile":    {"07535111111"},
			},
			existingLpa: &page.Lpa{},
			dataMatcher: func(t *testing.T, data *certificateProviderDetailsData) bool {
				return assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
			},
		},
		"name warning": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
				"mobile":      {"07535111111"},
			},
			existingLpa: &page.Lpa{
				Donor: actor.Donor{
					FirstNames: "John",
					LastName:   "Doe",
				},
			},
			dataMatcher: func(t *testing.T, data *certificateProviderDetailsData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypeCertificateProvider, actor.TypeDonor, "John", "Doe"), data.NameWarning)
			},
		},
		"name warning ignored but other errors": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"ignore-name-warning": {"errorDonorMatchesActor|theCertificateProvider|John|Doe"},
			},
			existingLpa: &page.Lpa{
				Donor: actor.Donor{
					FirstNames: "John",
					LastName:   "Doe",
				},
			},
			dataMatcher: func(t *testing.T, data *certificateProviderDetailsData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypeCertificateProvider, actor.TypeDonor, "John", "Doe"), data.NameWarning)
			},
		},
		"other name warning ignored": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"mobile":              {"07535111111"},
				"ignore-name-warning": {"errorAttorneyMatchesActor|theCertificateProvider|John|Doe"},
			},
			existingLpa: &page.Lpa{
				Donor: actor.Donor{
					FirstNames: "John",
					LastName:   "Doe",
				},
			},
			dataMatcher: func(t *testing.T, data *certificateProviderDetailsData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypeCertificateProvider, actor.TypeDonor, "John", "Doe"), data.NameWarning)
			},
		},
		"same last name as donor warning": {
			form: url.Values{
				"first-names": {"Joyce"},
				"last-name":   {"Doe"},
				"mobile":      {"07535111111"},
			},
			existingLpa: &page.Lpa{
				Donor: actor.Donor{
					FirstNames: "John",
					LastName:   "Doe",
				},
			},
			dataMatcher: func(t *testing.T, data *certificateProviderDetailsData) bool {
				assert.True(t, data.SameLastnameAsDonor)
				return assert.Equal(t, (*actor.SameNameWarning)(nil), data.NameWarning)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			template := newMockTemplate(t)
			template.
				On("Execute", w, mock.MatchedBy(func(data *certificateProviderDetailsData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := CertificateProviderDetails(template.Execute, nil)(testAppData, w, r, tc.existingLpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostCertificateProviderDetailsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
		"mobile":      {"07535111111"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := CertificateProviderDetails(nil, donorStore)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestReadCertificateProviderDetailsForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"first-names":                 {"  John "},
		"last-name":                   {"Doe"},
		"mobile":                      {"07535111111"},
		"ignore-name-warning":         {"a warning"},
		"ignore-similar-name-warning": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readCertificateProviderDetailsForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("07535111111", result.Mobile)
	assert.Equal("a warning", result.IgnoreNameWarning)
	assert.Equal(true, result.IgnoreSimilarNameWarning)
}

func TestCertificateProviderDetailsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *certificateProviderDetailsForm
		errors validation.List
	}{
		"valid": {
			form: &certificateProviderDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Mobile:     "07535111111",
			},
		},
		"missing all": {
			form: &certificateProviderDetailsForm{},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}).
				With("mobile", validation.EnterError{Label: "yourCertificateProvidersUkMobileNumber"}),
		},
		"missing when non uk mobile": {
			form: &certificateProviderDetailsForm{HasNonUKMobile: true},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}).
				With("non-uk-mobile", validation.EnterError{Label: "yourCertificateProvidersMobileNumber"}),
		},
		"invalid incorrect mobile format": {
			form: &certificateProviderDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Mobile:     "0753511111",
			},
			errors: validation.With("mobile", validation.CustomError{Label: "enterAMobileNumberInTheCorrectFormat"}),
		},
		"invalid non uk mobile format": {
			form: &certificateProviderDetailsForm{
				FirstNames:     "A",
				LastName:       "B",
				HasNonUKMobile: true,
				NonUKMobile:    "0753511111",
			},
			errors: validation.With("non-uk-mobile", validation.CustomError{Label: "enterAMobileNumberInTheCorrectFormat"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}

func TestCertificateProviderMatches(t *testing.T) {
	lpa := &page.Lpa{
		Donor: actor.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
		}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "g", LastName: "h"},
			{FirstNames: "i", LastName: "j"},
		}},
		CertificateProvider: actor.CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: actor.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{FirstNames: "o", LastName: "p"},
		},
		AuthorisedSignatory: actor.AuthorisedSignatory{FirstNames: "a", LastName: "s"},
		IndependentWitness:  actor.IndependentWitness{FirstNames: "i", LastName: "w"},
	}

	assert.Equal(t, actor.TypeNone, certificateProviderMatches(lpa, "x", "y"))
	assert.Equal(t, actor.TypeDonor, certificateProviderMatches(lpa, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, certificateProviderMatches(lpa, "c", "d"))
	assert.Equal(t, actor.TypeAttorney, certificateProviderMatches(lpa, "E", "F"))
	assert.Equal(t, actor.TypeReplacementAttorney, certificateProviderMatches(lpa, "g", "h"))
	assert.Equal(t, actor.TypeReplacementAttorney, certificateProviderMatches(lpa, "I", "J"))
	assert.Equal(t, actor.TypeNone, certificateProviderMatches(lpa, "k", "l"))
	assert.Equal(t, actor.TypeNone, certificateProviderMatches(lpa, "m", "n"))
	assert.Equal(t, actor.TypeNone, certificateProviderMatches(lpa, "o", "p"))
	assert.Equal(t, actor.TypeAuthorisedSignatory, certificateProviderMatches(lpa, "a", "s"))
	assert.Equal(t, actor.TypeIndependentWitness, certificateProviderMatches(lpa, "i", "w"))
}

func TestCertificateProviderMatchesEmptyNamesIgnored(t *testing.T) {
	lpa := &page.Lpa{
		Donor: actor.Donor{FirstNames: "", LastName: ""},
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		CertificateProvider: actor.CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: actor.PeopleToNotify{
			{FirstNames: "", LastName: ""},
		},
	}

	assert.Equal(t, actor.TypeNone, certificateProviderMatches(lpa, "", ""))
}
