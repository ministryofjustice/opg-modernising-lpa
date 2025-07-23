package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
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

	err := CertificateProviderDetails(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
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

			err := CertificateProviderDetails(template.Execute, nil)(testAppData, w, r, tc.donor)
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

	err := CertificateProviderDetails(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
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
				FirstNames:     "John",
				LastName:       "Rey",
				Mobile:         "+337575757",
				HasNonUKMobile: true,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			service := newMockCertificateProviderService(t)
			service.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID: "lpa-id",
					Donor: donordata.Donor{
						FirstNames: "Jane",
						LastName:   "Doe",
					},
					CertificateProvider: tc.certificateProviderDetails,
				}).
				Return(nil)

			err := CertificateProviderDetails(nil, service)(testAppData, w, r, &donordata.Provided{
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

func TestPostCertificateProviderDetailsWhenSharesDetail(t *testing.T) {
	f := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
		"mobile":      {"07535111111"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := newMockCertificateProviderService(t)
	service.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			Donor: donordata.Donor{
				FirstNames: "Jane",
				LastName:   "Doe",
				Address:    testAddress,
			},
			CertificateProvider: donordata.CertificateProvider{
				UID:        testUID,
				FirstNames: "John",
				LastName:   "Doe",
				Mobile:     "07535111111",
				Address:    testAddress,
			},
		}).
		Return(nil)

	appData := appcontext.Data{Page: "/abc"}
	err := CertificateProviderDetails(nil, service)(appData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{
			FirstNames: "Jane",
			LastName:   "Doe",
			Address:    testAddress,
		},
		CertificateProvider: donordata.CertificateProvider{
			UID:        testUID,
			FirstNames: "Bob",
			LastName:   "Doe",
			Address:    testAddress,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathWarningInterruption.FormatQuery("lpa-id",
		url.Values{
			"warningFrom": {"/abc"},
			"next":        {donor.PathHowDoYouKnowYourCertificateProvider.Format("lpa-id")},
			"actor":       {actor.TypeCertificateProvider.String()},
		},
	), resp.Header.Get("Location"))
}

func TestPostCertificateProviderDetailsWhenSigned(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Rey"},
		"mobile":      {"07535111111"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	updated := &donordata.Provided{
		LpaID:    "lpa-id",
		SignedAt: testNow,
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "John",
			LastName:   "Rey",
			Mobile:     "07535111111",
		},
	}
	updated.UpdateCheckedHash()

	service := newMockCertificateProviderService(t)
	service.EXPECT().
		Put(r.Context(), updated).
		Return(nil)

	err := CertificateProviderDetails(nil, service)(testAppData, w, r, &donordata.Provided{
		LpaID:    "lpa-id",
		SignedAt: testNow,
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

			err := CertificateProviderDetails(template.Execute, nil)(testAppData, w, r, tc.existingDonor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostCertificateProviderDetailsWhenServiceErrors(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
		"mobile":      {"07535111111"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := newMockCertificateProviderService(t)
	service.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := CertificateProviderDetails(nil, service)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestReadCertificateProviderDetailsForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"first-names": {"  John "},
		"last-name":   {"Doe"},
		"mobile":      {"07535111111"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readCertificateProviderDetailsForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("07535111111", result.Mobile)
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
