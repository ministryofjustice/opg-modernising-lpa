package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCertificateProviderDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &certificateProviderDetailsData{
			App:  testAppData,
			Form: &certificateProviderDetailsForm{},
		}).
		Return(nil)

	err := CertificateProviderDetails(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCertificateProviderDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := CertificateProviderDetails(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCertificateProviderDetailsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			CertificateProviderDetails: page.CertificateProviderDetails{
				FirstNames: "John",
			},
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &certificateProviderDetailsData{
			App: testAppData,
			Form: &certificateProviderDetailsForm{
				FirstNames: "John",
			},
		}).
		Return(nil)

	err := CertificateProviderDetails(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCertificateProviderDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &certificateProviderDetailsData{
			App:  testAppData,
			Form: &certificateProviderDetailsForm{},
		}).
		Return(expectedError)

	err := CertificateProviderDetails(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCertificateProviderDetails(t *testing.T) {
	testCases := map[string]struct {
		form                       url.Values
		certificateProviderDetails page.CertificateProviderDetails
	}{
		"valid": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Rey"},
				"mobile":              {"07535111111"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
			},
			certificateProviderDetails: page.CertificateProviderDetails{
				FirstNames:  "John",
				LastName:    "Rey",
				Mobile:      "07535111111",
				DateOfBirth: date.New("1990", "1", "2"),
			},
		},
		"name warning ignored": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"mobile":              {"07535111111"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeCertificateProvider, actor.TypeDonor, "Jane", "Doe").String()},
			},
			certificateProviderDetails: page.CertificateProviderDetails{
				FirstNames:  "Jane",
				LastName:    "Doe",
				Mobile:      "07535111111",
				DateOfBirth: date.New("1990", "1", "2"),
			},
		},
		"similar name warning ignored": {
			form: url.Values{
				"first-names":                 {"Joyce"},
				"last-name":                   {"Doe"},
				"mobile":                      {"07535111111"},
				"date-of-birth-day":           {"2"},
				"date-of-birth-month":         {"1"},
				"date-of-birth-year":          {"1990"},
				"ignore-similar-name-warning": {"yes"},
			},
			certificateProviderDetails: page.CertificateProviderDetails{
				FirstNames:  "Joyce",
				LastName:    "Doe",
				Mobile:      "07535111111",
				DateOfBirth: date.New("1990", "1", "2"),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					Donor: actor.Donor{
						FirstNames: "Jane",
						LastName:   "Doe",
					},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					Donor: actor.Donor{
						FirstNames: "Jane",
						LastName:   "Doe",
					},
					CertificateProviderDetails: tc.certificateProviderDetails,
				}).
				Return(nil)

			err := CertificateProviderDetails(nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+page.Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole, resp.Header.Get("Location"))
		})
	}
}

func TestPostCertificateProviderDetailsWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		existingLpa *page.Lpa
		dataMatcher func(t *testing.T, data *certificateProviderDetailsData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name":           {"Doe"},
				"mobile":              {"07535111111"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
			},
			existingLpa: &page.Lpa{},
			dataMatcher: func(t *testing.T, data *certificateProviderDetailsData) bool {
				return assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
			},
		},
		"name warning": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"mobile":              {"07535111111"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
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
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
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
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
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
				"first-names":         {"Joyce"},
				"last-name":           {"Doe"},
				"mobile":              {"07535111111"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
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

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.existingLpa, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, mock.MatchedBy(func(data *certificateProviderDetailsData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := CertificateProviderDetails(template.Execute, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostCertificateProviderDetailsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"mobile":              {"07535111111"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := CertificateProviderDetails(nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestReadCertificateProviderDetailsForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"first-names":                 {"  John "},
		"last-name":                   {"Doe"},
		"mobile":                      {"07535111111"},
		"date-of-birth-day":           {"2"},
		"date-of-birth-month":         {"1"},
		"date-of-birth-year":          {"1990"},
		"ignore-name-warning":         {"a warning"},
		"ignore-similar-name-warning": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readCertificateProviderDetailsForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("07535111111", result.Mobile)
	assert.Equal(date.New("1990", "1", "2"), result.Dob)
	assert.Equal("a warning", result.IgnoreNameWarning)
	assert.Equal(true, result.IgnoreSimilarNameWarning)
}

func TestCertificateProviderDetailsFormValidate(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form   *certificateProviderDetailsForm
		errors validation.List
	}{
		"valid": {
			form: &certificateProviderDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Mobile:     "07535111111",
				Dob:        validDob,
			},
		},
		"missing-all": {
			form: &certificateProviderDetailsForm{},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}).
				With("date-of-birth", validation.EnterError{Label: "dateOfBirth"}).
				With("mobile", validation.EnterError{Label: "mobile"}),
		},
		"invalid-dob": {
			form: &certificateProviderDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Mobile:     "07535111111",
				Dob:        date.New("2000", "22", "2"),
			},
			errors: validation.With("date-of-birth", validation.DateMustBeRealError{Label: "dateOfBirth"}),
		},
		"invalid-missing-dob": {
			form: &certificateProviderDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Mobile:     "07535111111",
				Dob:        date.New("2000", "", "1"),
			},
			errors: validation.With("date-of-birth", validation.DateMissingError{Label: "dateOfBirth", MissingMonth: true}),
		},
		"invalid-incorrect-mobile-format": {
			form: &certificateProviderDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Mobile:     "0753511111",
				Dob:        validDob,
			},
			errors: validation.With("mobile", validation.MobileError{Label: "mobile"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}

func TestUkMobileFormatValidation(t *testing.T) {
	form := &certificateProviderDetailsForm{
		FirstNames: "A",
		LastName:   "B",
		Dob:        date.Today().AddDate(-18, 0, -1),
	}

	testCases := map[string]struct {
		Mobile string
		Error  validation.List
	}{
		"valid local format": {
			Mobile: "07535111222",
		},
		"valid international format": {
			Mobile: "+447535111222",
		},
		"valid local format spaces": {
			Mobile: "  0 7 5 3 5 1 1 1 2 2 2 ",
		},
		"valid international format spaces": {
			Mobile: "  + 4 4 7 5 3 5 1 1 1 2 2 2 ",
		},
		"invalid local too short": {
			Mobile: "0753511122",
			Error:  validation.With("mobile", validation.MobileError{Label: "mobile"}),
		},
		"invalid local too long": {
			Mobile: "075351112223",
			Error:  validation.With("mobile", validation.MobileError{Label: "mobile"}),
		},
		"invalid international too short": {
			Mobile: "+44753511122",
			Error:  validation.With("mobile", validation.MobileError{Label: "mobile"}),
		},
		"invalid international too long": {
			Mobile: "+4475351112223",
			Error:  validation.With("mobile", validation.MobileError{Label: "mobile"}),
		},
		"invalid contains alpha chars": {
			Mobile: "+44753511122a",
			Error:  validation.With("mobile", validation.MobileError{Label: "mobile"}),
		},
		"invalid local not uk": {
			Mobile: "09535111222",
			Error:  validation.With("mobile", validation.MobileError{Label: "mobile"}),
		},
		"invalid international not uk": {
			Mobile: "+449535111222",
			Error:  validation.With("mobile", validation.MobileError{Label: "mobile"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form.Mobile = tc.Mobile
			assert.Equal(t, tc.Error, form.Validate())
		})
	}
}

func TestCertificateProviderMatches(t *testing.T) {
	lpa := &page.Lpa{
		Donor: actor.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: actor.Attorneys{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
		},
		ReplacementAttorneys: actor.Attorneys{
			{FirstNames: "g", LastName: "h"},
			{FirstNames: "i", LastName: "j"},
		},
		CertificateProviderDetails: page.CertificateProviderDetails{FirstNames: "k", LastName: "l"},
		PeopleToNotify: actor.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{FirstNames: "o", LastName: "p"},
		},
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
}
