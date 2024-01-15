package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemoveAttorney(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	logger := newMockLogger(t)

	attorney := actor.Attorney{
		ID:         "123",
		FirstNames: "John",
		LastName:   "Smith",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &removeAttorneyData{
			App:        testAppData,
			TitleLabel: "removeAnAttorney",
			Name:       "John Smith",
			Form:       &form.YesNoForm{},
			Options:    form.YesNoValues,
		}).
		Return(nil)

	err := RemoveAttorney(logger, template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorney}}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRemoveAttorneyAttorneyDoesNotExist(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=invalid-id", nil)

	logger := newMockLogger(t)

	template := newMockTemplate(t)

	attorney := actor.Attorney{
		ID: "123",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	err := RemoveAttorney(logger, template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id", Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorney}}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveAttorney(t *testing.T) {
	attorneyWithEmail := actor.Attorney{ID: "with-email", Email: "a"}
	attorneyWithAddress := actor.Attorney{ID: "with-address", Address: place.Address{Line1: "1 Road way"}}
	attorneyWithoutAddress := actor.Attorney{ID: "without-address"}

	testcases := map[string]struct {
		donor        *actor.DonorProvidedDetails
		updatedDonor *actor.DonorProvidedDetails
		redirect     page.LpaPath
	}{
		"many left": {
			donor: &actor.DonorProvidedDetails{
				LpaID:             "lpa-id",
				Attorneys:         actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithEmail, attorneyWithAddress, attorneyWithoutAddress}},
				AttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedDonor: &actor.DonorProvidedDetails{
				LpaID:             "lpa-id",
				Attorneys:         actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithEmail, attorneyWithAddress}},
				AttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
				Tasks:             actor.DonorTasks{ChooseAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseAttorneysSummary,
		},
		"one left": {
			donor: &actor.DonorProvidedDetails{
				LpaID:             "lpa-id",
				Attorneys:         actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithAddress, attorneyWithoutAddress}},
				AttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedDonor: &actor.DonorProvidedDetails{
				LpaID:     "lpa-id",
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithAddress}},
				Tasks:     actor.DonorTasks{ChooseAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseAttorneysSummary,
		},
		"none left": {
			donor: &actor.DonorProvidedDetails{LpaID: "lpa-id", Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithoutAddress}}},
			updatedDonor: &actor.DonorProvidedDetails{
				LpaID:     "lpa-id",
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{}},
			},
			redirect: page.Paths.ChooseAttorneysSummary,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"yes-no": {form.Yes.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			logger := newMockLogger(t)
			template := newMockTemplate(t)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), tc.updatedDonor).
				Return(nil)

			err := RemoveAttorney(logger, template.Execute, donorStore)(testAppData, w, r, tc.donor)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostRemoveAttorneyWithFormValueNo(t *testing.T) {
	form := url.Values{
		"yes-no": {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)
	template := newMockTemplate(t)

	attorneyWithAddress := actor.Attorney{
		ID: "with-address",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	attorneyWithoutAddress := actor.Attorney{
		ID:      "without-address",
		Address: place.Address{},
	}

	err := RemoveAttorney(logger, template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id", Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithoutAddress, attorneyWithAddress}}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveAttorneyErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"yes-no": {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)

	logger := newMockLogger(t)
	logger.EXPECT().
		Print("error removing Attorney from LPA: err").
		Return()

	attorneyWithAddress := actor.Attorney{
		ID: "with-address",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	attorneyWithoutAddress := actor.Attorney{
		ID:      "without-address",
		Address: place.Address{},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := RemoveAttorney(logger, template.Execute, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithoutAddress, attorneyWithAddress}}})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveAttorneyFormValidation(t *testing.T) {
	form := url.Values{
		"yes-no": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyWithoutAddress := actor.Attorney{
		ID:      "without-address",
		Address: place.Address{},
	}

	validationError := validation.With("yes-no", validation.SelectError{Label: "yesToRemoveAttorney"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *removeAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveAttorney(nil, template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithoutAddress}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
