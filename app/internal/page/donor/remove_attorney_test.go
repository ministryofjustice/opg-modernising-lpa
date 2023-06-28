package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemoveAttorney(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	logger := newMockLogger(t)

	attorney := actor.Attorney{
		ID: "123",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &removeAttorneyData{
			App:      testAppData,
			Attorney: attorney,
			Errors:   nil,
			Form:     &removeAttorneyForm{},
			Options:  actor.YesNoValues,
		}).
		Return(nil)

	err := RemoveAttorney(logger, template.Execute, nil)(testAppData, w, r, &page.Lpa{Attorneys: actor.Attorneys{attorney}})

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

	err := RemoveAttorney(logger, template.Execute, nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", Attorneys: actor.Attorneys{attorney}})

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
		lpa        *page.Lpa
		updatedLpa *page.Lpa
		redirect   page.LpaPath
	}{
		"many left": {
			lpa: &page.Lpa{
				ID:                "lpa-id",
				Attorneys:         actor.Attorneys{attorneyWithEmail, attorneyWithAddress, attorneyWithoutAddress},
				AttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedLpa: &page.Lpa{
				ID:                "lpa-id",
				Attorneys:         actor.Attorneys{attorneyWithEmail, attorneyWithAddress},
				AttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
				Tasks:             page.Tasks{ChooseAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseAttorneysSummary,
		},
		"one left": {
			lpa: &page.Lpa{
				ID:                "lpa-id",
				Attorneys:         actor.Attorneys{attorneyWithAddress, attorneyWithoutAddress},
				AttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedLpa: &page.Lpa{
				ID:        "lpa-id",
				Attorneys: actor.Attorneys{attorneyWithAddress},
				Tasks:     page.Tasks{ChooseAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseAttorneysSummary,
		},
		"none left": {
			lpa: &page.Lpa{ID: "lpa-id", Attorneys: actor.Attorneys{attorneyWithoutAddress}},
			updatedLpa: &page.Lpa{
				ID:        "lpa-id",
				Attorneys: actor.Attorneys{},
			},
			redirect: page.Paths.ChooseAttorneysSummary,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"remove-attorney": {actor.Yes.String()},
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

			err := RemoveAttorney(logger, template.Execute, donorStore)(testAppData, w, r, tc.lpa)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostRemoveAttorneyWithFormValueNo(t *testing.T) {
	form := url.Values{
		"remove-attorney": {actor.No.String()},
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

	err := RemoveAttorney(logger, template.Execute, nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", Attorneys: actor.Attorneys{attorneyWithoutAddress, attorneyWithAddress}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveAttorneyErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"remove-attorney": {actor.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)

	logger := newMockLogger(t)
	logger.
		On("Print", "error removing Attorney from LPA: err").
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

	err := RemoveAttorney(logger, template.Execute, donorStore)(testAppData, w, r, &page.Lpa{Attorneys: actor.Attorneys{attorneyWithoutAddress, attorneyWithAddress}})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveAttorneyFormValidation(t *testing.T) {
	form := url.Values{
		"remove-attorney": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyWithoutAddress := actor.Attorney{
		ID:      "without-address",
		Address: place.Address{},
	}

	validationError := validation.With("remove-attorney", validation.SelectError{Label: "yesToRemoveAttorney"})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *removeAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveAttorney(nil, template.Execute, nil)(testAppData, w, r, &page.Lpa{Attorneys: actor.Attorneys{attorneyWithoutAddress}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveAttorneyFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *removeAttorneyForm
		errors validation.List
	}{
		"valid": {
			form: &removeAttorneyForm{},
		},
		"invalid": {
			form: &removeAttorneyForm{
				Error:      expectedError,
				errorLabel: "xyz",
			},
			errors: validation.With("remove-attorney", validation.SelectError{Label: "xyz"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
