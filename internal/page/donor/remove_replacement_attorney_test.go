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

func TestGetRemoveReplacementAttorney(t *testing.T) {
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
	template.
		On("Execute", w, &removeAttorneyData{
			App:        testAppData,
			TitleLabel: "doYouWantToRemoveReplacementAttorney",
			Name:       "John Smith",
			Form:       &form.YesNoForm{},
			Options:    form.YesNoValues,
		}).
		Return(nil)

	err := RemoveReplacementAttorney(logger, template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorney}}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRemoveReplacementAttorneyAttorneyDoesNotExist(t *testing.T) {
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

	err := RemoveReplacementAttorney(logger, template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ID: "lpa-id", ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorney}}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveReplacementAttorney(t *testing.T) {
	attorneyWithEmail := actor.Attorney{ID: "with-email", Email: "a"}
	attorneyWithAddress := actor.Attorney{ID: "with-address", Address: place.Address{Line1: "1 Road way"}}
	attorneyWithoutAddress := actor.Attorney{ID: "without-address"}

	testcases := map[string]struct {
		lpa        *actor.DonorProvidedDetails
		updatedLpa *actor.DonorProvidedDetails
		redirect   page.LpaPath
	}{
		"many left": {
			lpa: &actor.DonorProvidedDetails{
				ID:                           "lpa-id",
				ReplacementAttorneys:         actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithEmail, attorneyWithAddress, attorneyWithoutAddress}},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedLpa: &actor.DonorProvidedDetails{
				ID:                           "lpa-id",
				ReplacementAttorneys:         actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithEmail, attorneyWithAddress}},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
				Tasks:                        actor.DonorTasks{ChooseReplacementAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseReplacementAttorneysSummary,
		},
		"one left": {
			lpa: &actor.DonorProvidedDetails{
				ID:                           "lpa-id",
				ReplacementAttorneys:         actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithAddress, attorneyWithoutAddress}},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedLpa: &actor.DonorProvidedDetails{
				ID:                   "lpa-id",
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithAddress}},
				Tasks:                actor.DonorTasks{ChooseReplacementAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseReplacementAttorneysSummary,
		},
		"none left": {
			lpa: &actor.DonorProvidedDetails{ID: "lpa-id", ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithoutAddress}}},
			updatedLpa: &actor.DonorProvidedDetails{
				ID:                   "lpa-id",
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{}},
			},
			redirect: page.Paths.ChooseReplacementAttorneysSummary,
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
			donorStore.
				On("Put", r.Context(), tc.updatedLpa).
				Return(nil)

			err := RemoveReplacementAttorney(logger, template.Execute, donorStore)(testAppData, w, r, tc.lpa)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostRemoveReplacementAttorneyWithFormValueNo(t *testing.T) {
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

	err := RemoveReplacementAttorney(logger, template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ID: "lpa-id", ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithoutAddress, attorneyWithAddress}}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveReplacementAttorneyErrorOnPutStore(t *testing.T) {
	f := url.Values{
		"yes-no": {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)

	logger := newMockLogger(t)
	logger.
		On("Print", "error removing replacement Attorney from LPA: err").
		Return(nil)

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
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := RemoveReplacementAttorney(logger, template.Execute, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		WantReplacementAttorneys: form.Yes,
		ReplacementAttorneys:     actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithoutAddress, attorneyWithAddress}},
	})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveReplacementAttorneyFormValidation(t *testing.T) {
	f := url.Values{
		"yes-no": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyWithoutAddress := actor.Attorney{
		ID:      "without-address",
		Address: place.Address{},
	}

	validationError := validation.With("yes-no", validation.SelectError{Label: "yesToRemoveReplacementAttorney"})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *removeAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveReplacementAttorney(nil, template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithoutAddress}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
