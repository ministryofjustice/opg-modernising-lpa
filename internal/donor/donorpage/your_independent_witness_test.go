package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourIndependentWitness(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourIndependentWitnessData{
			App:  testAppData,
			Form: &yourIndependentWitnessForm{},
		}).
		Return(nil)

	err := YourIndependentWitness(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourIndependentWitnessFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourIndependentWitnessData{
			App: testAppData,
			Form: &yourIndependentWitnessForm{
				FirstNames: "John",
			},
		}).
		Return(nil)

	err := YourIndependentWitness(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		IndependentWitness: donordata.IndependentWitness{
			FirstNames: "John",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourIndependentWitnessWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := YourIndependentWitness(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourIndependentWitness(t *testing.T) {
	testCases := map[string]struct {
		form   url.Values
		person donordata.IndependentWitness
	}{
		"valid": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
			},
			person: donordata.IndependentWitness{
				FirstNames: "John",
				LastName:   "Doe",
			},
		},
		"warning ignored": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Smith"},
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeIndependentWitness, actor.TypeDonor, "John", "Smith").String()},
			},
			person: donordata.IndependentWitness{
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
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:              "lpa-id",
					Donor:              donordata.Donor{FirstNames: "John", LastName: "Smith"},
					IndependentWitness: tc.person,
					Tasks:              donordata.Tasks{ChooseYourSignatory: task.StateInProgress},
				}).
				Return(nil)

			err := YourIndependentWitness(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{FirstNames: "John", LastName: "Smith"},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.YourIndependentWitnessMobile.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostYourIndependentWitnessWhenTaskCompleted(t *testing.T) {
	f := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			IndependentWitness: donordata.IndependentWitness{
				FirstNames: "John",
				LastName:   "Doe",
			},
			Tasks: donordata.Tasks{ChooseYourSignatory: task.StateCompleted},
		}).
		Return(nil)

	err := YourIndependentWitness(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		IndependentWitness: donordata.IndependentWitness{
			FirstNames: "John",
		},
		Tasks: donordata.Tasks{ChooseYourSignatory: task.StateCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.YourIndependentWitnessMobile.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostYourIndependentWitnessWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *yourIndependentWitnessData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name": {"Doe"},
			},
			dataMatcher: func(t *testing.T, data *yourIndependentWitnessData) bool {
				return assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
			},
		},
		"name warning": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
			},
			dataMatcher: func(t *testing.T, data *yourIndependentWitnessData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypeIndependentWitness, actor.TypeDonor, "John", "Doe"), data.NameWarning)
			},
		},
		"name warning ignored but other errors": {
			form: url.Values{
				"first-names":         {"John"},
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeIndependentWitness, actor.TypeDonor, "John", "Doe").String()},
			},
			dataMatcher: func(t *testing.T, data *yourIndependentWitnessData) bool {
				return assert.Equal(t, validation.With("last-name", validation.EnterError{Label: "lastName"}), data.Errors)
			},
		},
		"other name warning ignored": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeIndependentWitness, actor.TypeDonor, "John", "John").String()},
			},
			dataMatcher: func(t *testing.T, data *yourIndependentWitnessData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypeIndependentWitness, actor.TypeDonor, "John", "Doe"), data.NameWarning)
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
				Execute(w, mock.MatchedBy(func(data *yourIndependentWitnessData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := YourIndependentWitness(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
				Donor: donordata.Donor{
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

func TestPostYourIndependentWitnessWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := YourIndependentWitness(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{
			FirstNames: "John",
			Address:    place.Address{Line1: "abc"},
		},
	})

	assert.Equal(t, expectedError, err)
}

func TestReadYourIndependentWitnessForm(t *testing.T) {
	assert := assert.New(t)

	f := url.Values{
		"first-names":         {"  John "},
		"last-name":           {"Doe"},
		"ignore-name-warning": {"xyz"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readYourIndependentWitnessForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("xyz", result.IgnoreNameWarning)
}

func TestYourIndependentWitnessFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *yourIndependentWitnessForm
		errors validation.List
	}{
		"valid": {
			form: &yourIndependentWitnessForm{
				FirstNames: "A",
				LastName:   "B",
			},
		},
		"max length": {
			form: &yourIndependentWitnessForm{
				FirstNames: strings.Repeat("x", 53),
				LastName:   strings.Repeat("x", 61),
			},
		},
		"missing all": {
			form: &yourIndependentWitnessForm{},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}),
		},
		"too long": {
			form: &yourIndependentWitnessForm{
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

func TestIndependentWitnessMatches(t *testing.T) {
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

	assert.Equal(t, actor.TypeNone, independentWitnessMatches(donor, "x", "y"))
	assert.Equal(t, actor.TypeDonor, independentWitnessMatches(donor, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, independentWitnessMatches(donor, "C", "D"))
	assert.Equal(t, actor.TypeAttorney, independentWitnessMatches(donor, "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, independentWitnessMatches(donor, "G", "H"))
	assert.Equal(t, actor.TypeReplacementAttorney, independentWitnessMatches(donor, "i", "j"))
	assert.Equal(t, actor.TypeCertificateProvider, independentWitnessMatches(donor, "k", "l"))
	assert.Equal(t, actor.TypeNone, independentWitnessMatches(donor, "m", "n"))
	assert.Equal(t, actor.TypeNone, independentWitnessMatches(donor, "O", "P"))
	assert.Equal(t, actor.TypeAuthorisedSignatory, independentWitnessMatches(donor, "a", "s"))
	assert.Equal(t, actor.TypeNone, independentWitnessMatches(donor, "i", "w"))
}

func TestIndependentWitnessMatchesEmptyNamesIgnored(t *testing.T) {
	donor := &donordata.Provided{
		Attorneys:            donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
		PeopleToNotify:       donordata.PeopleToNotify{{}},
	}

	assert.Equal(t, actor.TypeNone, independentWitnessMatches(donor, "", ""))
}
