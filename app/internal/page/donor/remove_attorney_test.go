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
		}).
		Return(nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: actor.Attorneys{attorney}}, nil)

	err := RemoveAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRemoveAttorneyErrorOnStore(t *testing.T) {
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

	err := RemoveAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: actor.Attorneys{attorney}}, nil)

	err := RemoveAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseAttorneysSummary, resp.Header.Get("Location"))
}

func TestPostRemoveAttorney(t *testing.T) {
	form := url.Values{
		"remove-attorney": {"yes"},
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
		Return(&page.Lpa{Attorneys: actor.Attorneys{attorneyWithoutAddress, attorneyWithAddress}}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{Attorneys: actor.Attorneys{attorneyWithAddress}}).
		Return(nil)

	err := RemoveAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseAttorneysSummary, resp.Header.Get("Location"))
}

func TestPostRemoveAttorneyWithFormValueNo(t *testing.T) {
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
		Return(&page.Lpa{Attorneys: actor.Attorneys{attorneyWithoutAddress, attorneyWithAddress}}, nil)

	err := RemoveAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseAttorneysSummary, resp.Header.Get("Location"))
}

func TestPostRemoveAttorneyErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"remove-attorney": {"yes"},
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: actor.Attorneys{attorneyWithoutAddress, attorneyWithAddress}}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{Attorneys: actor.Attorneys{attorneyWithAddress}}).
		Return(expectedError)

	err := RemoveAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: actor.Attorneys{attorneyWithoutAddress}}, nil)

	validationError := validation.With("remove-attorney", validation.SelectError{Label: "yesToRemoveAttorney"})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *removeAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveAttorney(nil, template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveAttorneyRemoveLastAttorneyRedirectsToChooseAttorney(t *testing.T) {
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
		Return(&page.Lpa{Attorneys: actor.Attorneys{{ID: "without-address"}}, Tasks: page.Tasks{ChooseAttorneys: page.TaskCompleted}}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{Attorneys: actor.Attorneys{}, Tasks: page.Tasks{ChooseAttorneys: page.TaskInProgress}}).
		Return(nil)

	err := RemoveAttorney(logger, template.Execute, lpaStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseAttorneys, resp.Header.Get("Location"))
}

func TestRemoveAttorneyFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *removeAttorneyForm
		errors validation.List
	}{
		"valid - yes": {
			form: &removeAttorneyForm{
				RemoveAttorney: "yes",
			},
		},
		"valid - no": {
			form: &removeAttorneyForm{
				RemoveAttorney: "no",
			},
		},
		"missing-value": {
			form: &removeAttorneyForm{
				RemoveAttorney: "",
				errorLabel:     "xyz",
			},
			errors: validation.With("remove-attorney", validation.SelectError{Label: "xyz"}),
		},
		"unexpected-value": {
			form: &removeAttorneyForm{
				RemoveAttorney: "not expected",
				errorLabel:     "xyz",
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
