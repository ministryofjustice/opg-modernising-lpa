package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourAuthorisedSignatory(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourAuthorisedSignatoryData{
			App:  testAppData,
			Form: &yourAuthorisedSignatoryForm{},
		}).
		Return(nil)

	err := YourAuthorisedSignatory(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourAuthorisedSignatoryFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourAuthorisedSignatoryData{
			App: testAppData,
			Form: &yourAuthorisedSignatoryForm{
				FirstNames: "John",
			},
		}).
		Return(nil)

	err := YourAuthorisedSignatory(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		AuthorisedSignatory: actor.AuthorisedSignatory{
			FirstNames: "John",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourAuthorisedSignatoryWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := YourAuthorisedSignatory(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAuthorisedSignatory(t *testing.T) {
	testCases := map[string]struct {
		form   url.Values
		person actor.AuthorisedSignatory
	}{
		"valid": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
			},
			person: actor.AuthorisedSignatory{
				FirstNames: "John",
				LastName:   "Doe",
			},
		},
		"warning ignored": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Smith"},
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeAuthorisedSignatory, actor.TypeDonor, "John", "Smith").String()},
			},
			person: actor.AuthorisedSignatory{
				FirstNames: "John",
				LastName:   "Smith",
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
				On("Put", r.Context(), &actor.DonorProvidedDetails{
					LpaID:               "lpa-id",
					Donor:               actor.Donor{FirstNames: "John", LastName: "Smith"},
					AuthorisedSignatory: tc.person,
					Tasks:               actor.DonorTasks{ChooseYourSignatory: actor.TaskInProgress},
				}).
				Return(nil)

			err := YourAuthorisedSignatory(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{FirstNames: "John", LastName: "Smith"},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.YourIndependentWitness.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostYourAuthorisedSignatoryWhenTaskCompleted(t *testing.T) {
	f := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &actor.DonorProvidedDetails{
			LpaID: "lpa-id",
			AuthorisedSignatory: actor.AuthorisedSignatory{
				FirstNames: "John",
				LastName:   "Doe",
			},
			Tasks: actor.DonorTasks{ChooseYourSignatory: actor.TaskCompleted},
		}).
		Return(nil)

	err := YourAuthorisedSignatory(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		AuthorisedSignatory: actor.AuthorisedSignatory{
			FirstNames: "John",
		},
		Tasks: actor.DonorTasks{ChooseYourSignatory: actor.TaskCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.YourIndependentWitness.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostYourAuthorisedSignatoryWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *yourAuthorisedSignatoryData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name": {"Doe"},
			},
			dataMatcher: func(t *testing.T, data *yourAuthorisedSignatoryData) bool {
				return assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
			},
		},
		"name warning": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
			},
			dataMatcher: func(t *testing.T, data *yourAuthorisedSignatoryData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypeAuthorisedSignatory, actor.TypeDonor, "John", "Doe"), data.NameWarning)
			},
		},
		"name warning ignored but other errors": {
			form: url.Values{
				"first-names":         {"John"},
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeAuthorisedSignatory, actor.TypeDonor, "John", "Doe").String()},
			},
			dataMatcher: func(t *testing.T, data *yourAuthorisedSignatoryData) bool {
				return assert.Equal(t, validation.With("last-name", validation.EnterError{Label: "lastName"}), data.Errors)
			},
		},
		"other name warning ignored": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeAuthorisedSignatory, actor.TypeDonor, "John", "John").String()},
			},
			dataMatcher: func(t *testing.T, data *yourAuthorisedSignatoryData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypeAuthorisedSignatory, actor.TypeDonor, "John", "Doe"), data.NameWarning)
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
				On("Execute", w, mock.MatchedBy(func(data *yourAuthorisedSignatoryData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := YourAuthorisedSignatory(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
				Donor: actor.Donor{
					FirstNames: "John",
					LastName:   "Doe",
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostYourAuthorisedSignatoryWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := YourAuthorisedSignatory(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{
			FirstNames: "John",
			Address:    place.Address{Line1: "abc"},
		},
	})

	assert.Equal(t, expectedError, err)
}

func TestReadYourAuthorisedSignatoryForm(t *testing.T) {
	assert := assert.New(t)

	f := url.Values{
		"first-names":         {"  John "},
		"last-name":           {"Doe"},
		"ignore-name-warning": {"xyz"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readYourAuthorisedSignatoryForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("xyz", result.IgnoreNameWarning)
}

func TestYourAuthorisedSignatoryFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *yourAuthorisedSignatoryForm
		errors validation.List
	}{
		"valid": {
			form: &yourAuthorisedSignatoryForm{
				FirstNames: "A",
				LastName:   "B",
			},
		},
		"max length": {
			form: &yourAuthorisedSignatoryForm{
				FirstNames: strings.Repeat("x", 53),
				LastName:   strings.Repeat("x", 61),
			},
		},
		"missing all": {
			form: &yourAuthorisedSignatoryForm{},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}),
		},
		"too long": {
			form: &yourAuthorisedSignatoryForm{
				FirstNames: strings.Repeat("x", 54),
				LastName:   strings.Repeat("x", 62),
			},
			errors: validation.
				With("first-names", validation.StringTooLongError{Label: "firstNames", Length: 53}).
				With("last-name", validation.StringTooLongError{Label: "lastName", Length: 61}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}

func TestSignatoryMatches(t *testing.T) {
	lpa := &actor.DonorProvidedDetails{
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

	assert.Equal(t, actor.TypeNone, signatoryMatches(lpa, "x", "y"))
	assert.Equal(t, actor.TypeDonor, signatoryMatches(lpa, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, signatoryMatches(lpa, "C", "D"))
	assert.Equal(t, actor.TypeAttorney, signatoryMatches(lpa, "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, signatoryMatches(lpa, "G", "H"))
	assert.Equal(t, actor.TypeReplacementAttorney, signatoryMatches(lpa, "i", "j"))
	assert.Equal(t, actor.TypeCertificateProvider, signatoryMatches(lpa, "k", "l"))
	assert.Equal(t, actor.TypeNone, signatoryMatches(lpa, "m", "n"))
	assert.Equal(t, actor.TypeNone, signatoryMatches(lpa, "O", "P"))
	assert.Equal(t, actor.TypeNone, signatoryMatches(lpa, "a", "s"))
	assert.Equal(t, actor.TypeIndependentWitness, signatoryMatches(lpa, "i", "w"))
}

func TestSignatoryMatchesEmptyNamesIgnored(t *testing.T) {
	lpa := &actor.DonorProvidedDetails{
		Attorneys:            actor.Attorneys{Attorneys: []actor.Attorney{{}}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{}}},
		PeopleToNotify:       actor.PeopleToNotify{{}},
	}

	assert.Equal(t, actor.TypeNone, signatoryMatches(lpa, "", ""))
}
