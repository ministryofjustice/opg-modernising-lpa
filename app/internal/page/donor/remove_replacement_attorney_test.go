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
		}).
		Return(nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{attorney}}, nil)

	err := RemoveReplacementAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRemoveReplacementAttorneyErrorOnStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	logger := newMockLogger(t)
	logger.
		On("Print", "error getting lpa from store: err").
		Return(nil)

	template := newMockTemplate(t)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := RemoveReplacementAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{attorney}}, nil)

	err := RemoveReplacementAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))
}

func TestPostRemoveReplacementAttorney(t *testing.T) {
	attorneyWithEmail := actor.Attorney{ID: "with-email", Email: "a"}
	attorneyWithAddress := actor.Attorney{ID: "with-address", Address: place.Address{Line1: "1 Road way"}}
	attorneyWithoutAddress := actor.Attorney{ID: "without-address"}

	testcases := map[string]struct {
		lpa        *page.Lpa
		updatedLpa *page.Lpa
		redirect   string
	}{
		"many left": {
			lpa: &page.Lpa{
				ReplacementAttorneys:         actor.Attorneys{attorneyWithEmail, attorneyWithAddress, attorneyWithoutAddress},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedLpa: &page.Lpa{
				ReplacementAttorneys:         actor.Attorneys{attorneyWithEmail, attorneyWithAddress},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
				Tasks:                        page.Tasks{ChooseReplacementAttorneys: page.TaskInProgress},
			},
			redirect: page.Paths.ChooseReplacementAttorneysSummary,
		},
		"one left": {
			lpa: &page.Lpa{
				ReplacementAttorneys:         actor.Attorneys{attorneyWithAddress, attorneyWithoutAddress},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedLpa: &page.Lpa{
				ReplacementAttorneys: actor.Attorneys{attorneyWithAddress},
				Tasks:                page.Tasks{ChooseReplacementAttorneys: page.TaskInProgress},
			},
			redirect: page.Paths.ChooseReplacementAttorneysSummary,
		},
		"none left": {
			lpa: &page.Lpa{ReplacementAttorneys: actor.Attorneys{attorneyWithoutAddress}},
			updatedLpa: &page.Lpa{
				ReplacementAttorneys: actor.Attorneys{},
			},
			redirect: page.Paths.DoYouWantReplacementAttorneys,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {

			form := url.Values{
				"remove-attorney": {"yes"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			logger := newMockLogger(t)
			template := newMockTemplate(t)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.lpa, nil)
			lpaStore.
				On("Put", r.Context(), tc.updatedLpa).
				Return(nil)

			err := RemoveReplacementAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+tc.redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostRemoveReplacementAttorneyWithFormValueNo(t *testing.T) {
	form := url.Values{
		"remove-attorney": {"no"},
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{attorneyWithoutAddress, attorneyWithAddress}}, nil)

	err := RemoveReplacementAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))
}

func TestPostRemoveReplacementAttorneyErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"remove-attorney": {"yes"},
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			WantReplacementAttorneys: "yes",
			ReplacementAttorneys:     actor.Attorneys{attorneyWithoutAddress, attorneyWithAddress},
		}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := RemoveReplacementAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{attorneyWithoutAddress}}, nil)

	validationError := validation.With("remove-attorney", validation.SelectError{Label: "yesToRemoveReplacementAttorney"})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *removeReplacementAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveReplacementAttorney(nil, template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveReplacementAttorneyRemoveLastAttorneyRedirectsToChooseReplacementAttorney(t *testing.T) {
	form := url.Values{
		"remove-attorney": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)
	template := newMockTemplate(t)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			WantReplacementAttorneys: "yes",
			ReplacementAttorneys:     actor.Attorneys{{ID: "without-address"}},
			Tasks:                    page.Tasks{ChooseReplacementAttorneys: page.TaskCompleted},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			WantReplacementAttorneys: "yes",
			ReplacementAttorneys:     actor.Attorneys{},
			Tasks:                    page.Tasks{ChooseReplacementAttorneys: page.TaskInProgress},
		}).
		Return(nil)

	err := RemoveReplacementAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.DoYouWantReplacementAttorneys, resp.Header.Get("Location"))
}
