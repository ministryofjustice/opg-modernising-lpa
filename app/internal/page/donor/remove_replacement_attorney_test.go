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

func TestGetRemoveReplacementAttorney(t *testing.T) {
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
		On("Execute", w, &removeReplacementAttorneyData{
			App:      testAppData,
			Attorney: attorney,
			Form:     &removeAttorneyForm{},
			Options:  actor.YesNoValues,
		}).
		Return(nil)

	err := RemoveReplacementAttorney(logger, template.Execute, nil)(testAppData, w, r, &page.Lpa{ReplacementAttorneys: actor.Attorneys{attorney}})

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

	err := RemoveReplacementAttorney(logger, template.Execute, nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", ReplacementAttorneys: actor.Attorneys{attorney}})

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
		lpa        *page.Lpa
		updatedLpa *page.Lpa
		redirect   page.LpaPath
	}{
		"many left": {
			lpa: &page.Lpa{
				ID:                           "lpa-id",
				ReplacementAttorneys:         actor.Attorneys{attorneyWithEmail, attorneyWithAddress, attorneyWithoutAddress},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedLpa: &page.Lpa{
				ID:                           "lpa-id",
				ReplacementAttorneys:         actor.Attorneys{attorneyWithEmail, attorneyWithAddress},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
				Tasks:                        page.Tasks{ChooseReplacementAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseReplacementAttorneysSummary,
		},
		"one left": {
			lpa: &page.Lpa{
				ID:                           "lpa-id",
				ReplacementAttorneys:         actor.Attorneys{attorneyWithAddress, attorneyWithoutAddress},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedLpa: &page.Lpa{
				ID:                   "lpa-id",
				ReplacementAttorneys: actor.Attorneys{attorneyWithAddress},
				Tasks:                page.Tasks{ChooseReplacementAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseReplacementAttorneysSummary,
		},
		"none left": {
			lpa: &page.Lpa{ID: "lpa-id", ReplacementAttorneys: actor.Attorneys{attorneyWithoutAddress}},
			updatedLpa: &page.Lpa{
				ID:                   "lpa-id",
				ReplacementAttorneys: actor.Attorneys{},
			},
			redirect: page.Paths.ChooseReplacementAttorneysSummary,
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

	err := RemoveReplacementAttorney(logger, template.Execute, nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", ReplacementAttorneys: actor.Attorneys{attorneyWithoutAddress, attorneyWithAddress}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveReplacementAttorneyErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"remove-attorney": {actor.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
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

	err := RemoveReplacementAttorney(logger, template.Execute, donorStore)(testAppData, w, r, &page.Lpa{
		WantReplacementAttorneys: actor.Yes,
		ReplacementAttorneys:     actor.Attorneys{attorneyWithoutAddress, attorneyWithAddress},
	})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveReplacementAttorneyFormValidation(t *testing.T) {
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

	validationError := validation.With("remove-attorney", validation.SelectError{Label: "yesToRemoveReplacementAttorney"})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *removeReplacementAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveReplacementAttorney(nil, template.Execute, nil)(testAppData, w, r, &page.Lpa{ReplacementAttorneys: actor.Attorneys{attorneyWithoutAddress}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
