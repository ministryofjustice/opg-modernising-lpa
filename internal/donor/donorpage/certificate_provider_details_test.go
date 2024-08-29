package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCertificateProviderDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &certificateProviderDetailsData{
			App:  testAppData,
			Form: &certificateProviderDetailsForm{},
		}).
		Return(nil)

	err := CertificateProviderDetails(template.Execute, nil, testUIDFn)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCertificateProviderDetailsFromStore(t *testing.T) {
	testcases := map[string]struct {
		donor *donordata.Provided
		form  *certificateProviderDetailsForm
	}{
		"uk mobile": {
			donor: &donordata.Provided{
				CertificateProvider: donordata.CertificateProvider{
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
			donor: &donordata.Provided{
				CertificateProvider: donordata.CertificateProvider{
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
			template.EXPECT().
				Execute(w, &certificateProviderDetailsData{
					App:  testAppData,
					Form: tc.form,
				}).
				Return(nil)

			err := CertificateProviderDetails(template.Execute, nil, testUIDFn)(testAppData, w, r, tc.donor)
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
	template.EXPECT().
		Execute(w, &certificateProviderDetailsData{
			App:  testAppData,
			Form: &certificateProviderDetailsForm{},
		}).
		Return(expectedError)

	err := CertificateProviderDetails(template.Execute, nil, testUIDFn)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCertificateProviderDetails(t *testing.T) {
	testCases := map[string]struct {
		form                       url.Values
		certificateProviderDetails donordata.CertificateProvider
	}{
		"valid": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Rey"},
				"mobile":      {"07535111111"},
			},
			certificateProviderDetails: donordata.CertificateProvider{
				UID:        testUID,
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
			certificateProviderDetails: donordata.CertificateProvider{
				UID:            testUID,
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
			certificateProviderDetails: donordata.CertificateProvider{
				UID:        testUID,
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
			certificateProviderDetails: donordata.CertificateProvider{
				UID:        testUID,
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
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID: "lpa-id",
					Donor: donordata.Donor{
						FirstNames: "Jane",
						LastName:   "Doe",
					},
					CertificateProvider: tc.certificateProviderDetails,
					Tasks:               task.DonorTasks{CertificateProvider: task.StateInProgress},
				}).
				Return(nil)

			err := CertificateProviderDetails(nil, donorStore, testUIDFn)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					FirstNames: "Jane",
					LastName:   "Doe",
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathHowDoYouKnowYourCertificateProvider.Format("lpa-id"), resp.Header.Get("Location"))
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
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			Donor: donordata.Donor{
				FirstNames: "Jane",
				LastName:   "Doe",
			},
			CertificateProvider: donordata.CertificateProvider{
				UID:        testUID,
				FirstNames: "John",
				LastName:   "Rey",
				Mobile:     "07535111111",
			},
			Tasks: task.DonorTasks{CertificateProvider: task.StateCompleted},
		}).
		Return(nil)

	err := CertificateProviderDetails(nil, donorStore, testUIDFn)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{
			FirstNames: "Jane",
			LastName:   "Doe",
		},
		Tasks: task.DonorTasks{CertificateProvider: task.StateCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathHowDoYouKnowYourCertificateProvider.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostCertificateProviderDetailsWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form          url.Values
		existingDonor *donordata.Provided
		dataMatcher   func(t *testing.T, data *certificateProviderDetailsData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name": {"Doe"},
				"mobile":    {"07535111111"},
			},
			existingDonor: &donordata.Provided{},
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
			existingDonor: &donordata.Provided{
				Donor: donordata.Donor{
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
			existingDonor: &donordata.Provided{
				Donor: donordata.Donor{
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
			existingDonor: &donordata.Provided{
				Donor: donordata.Donor{
					FirstNames: "John",
					LastName:   "Doe",
				},
			},
			dataMatcher: func(t *testing.T, data *certificateProviderDetailsData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypeCertificateProvider, actor.TypeDonor, "John", "Doe"), data.NameWarning)
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
				Execute(w, mock.MatchedBy(func(data *certificateProviderDetailsData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := CertificateProviderDetails(template.Execute, nil, testUIDFn)(testAppData, w, r, tc.existingDonor)
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
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := CertificateProviderDetails(nil, donorStore, testUIDFn)(testAppData, w, r, &donordata.Provided{})

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
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "g", LastName: "h"},
			{FirstNames: "i", LastName: "j"},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{FirstNames: "o", LastName: "p"},
		},
		AuthorisedSignatory: donordata.AuthorisedSignatory{FirstNames: "a", LastName: "s"},
		IndependentWitness:  donordata.IndependentWitness{FirstNames: "i", LastName: "w"},
	}

	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "x", "y"))
	assert.Equal(t, actor.TypeDonor, certificateProviderMatches(donor, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, certificateProviderMatches(donor, "c", "d"))
	assert.Equal(t, actor.TypeAttorney, certificateProviderMatches(donor, "E", "F"))
	assert.Equal(t, actor.TypeReplacementAttorney, certificateProviderMatches(donor, "g", "h"))
	assert.Equal(t, actor.TypeReplacementAttorney, certificateProviderMatches(donor, "I", "J"))
	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "k", "l"))
	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "m", "n"))
	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "o", "p"))
	assert.Equal(t, actor.TypeAuthorisedSignatory, certificateProviderMatches(donor, "a", "s"))
	assert.Equal(t, actor.TypeIndependentWitness, certificateProviderMatches(donor, "i", "w"))
}

func TestCertificateProviderMatchesEmptyNamesIgnored(t *testing.T) {
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "", LastName: ""},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "", LastName: ""},
		},
	}

	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "", ""))
}
