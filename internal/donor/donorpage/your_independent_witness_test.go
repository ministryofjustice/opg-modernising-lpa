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

	err := YourIndependentWitness(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
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

	err := YourIndependentWitness(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
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

	err := YourIndependentWitness(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourIndependentWitness(t *testing.T) {
	testCases := map[string]struct {
		form             url.Values
		person           donordata.IndependentWitness
		expectedRedirect string
	}{
		"valid": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
			},
			person: donordata.IndependentWitness{
				UID:        testUID,
				FirstNames: "John",
				LastName:   "Doe",
			},
			expectedRedirect: donor.PathYourIndependentWitnessMobile.Format("lpa-id"),
		},
		"when name matches": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Smith"},
			},
			person: donordata.IndependentWitness{
				UID:        testUID,
				FirstNames: "John",
				LastName:   "Smith",
			},
			expectedRedirect: donor.PathWarningInterruption.FormatQuery(
				"lpa-id",
				url.Values{
					"warningFrom": {"/abc"},
					"next":        {donor.PathYourIndependentWitnessMobile.Format("lpa-id")},
					"actor":       {actor.TypeIndependentWitness.String()},
				}),
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

			appData := appcontext.Data{Page: "/abc"}
			err := YourIndependentWitness(nil, donorStore, testUIDFn)(appData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{FirstNames: "John", LastName: "Smith"},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostYourIndependentWitnessWhenSigned(t *testing.T) {
	f := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	updated := &donordata.Provided{
		LpaID:    "lpa-id",
		SignedAt: testNow,
		IndependentWitness: donordata.IndependentWitness{
			UID:        testUID,
			FirstNames: "John",
			LastName:   "Doe",
		},
		Tasks: donordata.Tasks{ChooseYourSignatory: task.StateInProgress},
	}
	updated.UpdateCheckedHash()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), updated).
		Return(nil)

	err := YourIndependentWitness(nil, donorStore, testUIDFn)(testAppData, w, r, &donordata.Provided{
		LpaID:    "lpa-id",
		SignedAt: testNow,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathYourIndependentWitnessMobile.Format("lpa-id"), resp.Header.Get("Location"))
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
				UID:        testUID,
				FirstNames: "John",
				LastName:   "Doe",
			},
			Tasks: donordata.Tasks{ChooseYourSignatory: task.StateCompleted},
		}).
		Return(nil)

	err := YourIndependentWitness(nil, donorStore, testUIDFn)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		IndependentWitness: donordata.IndependentWitness{
			FirstNames: "John",
		},
		Tasks: donordata.Tasks{ChooseYourSignatory: task.StateCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathYourIndependentWitnessMobile.Format("lpa-id"), resp.Header.Get("Location"))
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

			err := YourIndependentWitness(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
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

	err := YourIndependentWitness(nil, donorStore, testUIDFn)(testAppData, w, r, &donordata.Provided{
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
		"first-names": {"  John "},
		"last-name":   {"Doe"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readYourIndependentWitnessForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
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
