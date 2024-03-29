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

func TestGetYourName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourNameData{
			App:  testAppData,
			Form: &yourNameForm{},
		}).
		Return(nil)

	err := YourName(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourNameFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourNameData{
			App: testAppData,
			Form: &yourNameForm{
				FirstNames: "John",
				LastName:   "Doe",
				OtherNames: "Fawn",
			},
		}).
		Return(nil)

	err := YourName(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{
			FirstNames: "John",
			LastName:   "Doe",
			OtherNames: "Fawn",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourNameWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := YourName(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{Donor: actor.Donor{FirstNames: "John"}})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourName(t *testing.T) {
	testCases := map[string]struct {
		form   url.Values
		person actor.Donor
	}{
		"valid": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
				"other-names": {"Fawn"},
			},
			person: actor.Donor{
				FirstNames: "John",
				LastName:   "Doe",
				OtherNames: "Fawn",
			},
		},
		"warning ignored": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
			},
			person: actor.Donor{
				FirstNames: "John",
				LastName:   "Doe",
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
					LpaID: "lpa-id",
					Donor: tc.person,
				}).
				Return(nil)

			err := YourName(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{
					FirstNames: "John",
				},
				HasSentApplicationUpdatedEvent: true,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.WeHaveUpdatedYourDetails.Format("lpa-id")+"?detail=name", resp.Header.Get("Location"))
		})
	}
}

func TestPostYourNameWhenDetailsNotChanged(t *testing.T) {
	f := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
		"other-names": {"Fawn"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := YourName(nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		Donor: actor.Donor{
			FirstNames: "John",
			LastName:   "Doe",
			OtherNames: "Fawn",
		},
		HasSentApplicationUpdatedEvent: true,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.MakeANewLPA.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostYourNameWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *yourNameData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name": {"Doe"},
			},
			dataMatcher: func(t *testing.T, data *yourNameData) bool {
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
			template.
				On("Execute", w, mock.MatchedBy(func(data *yourNameData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := YourName(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostYourNameWhenStoreErrors(t *testing.T) {
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

	err := YourName(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{
			FirstNames: "John",
			Address:    place.Address{Line1: "abc"},
		},
	})

	assert.Equal(t, expectedError, err)
}
