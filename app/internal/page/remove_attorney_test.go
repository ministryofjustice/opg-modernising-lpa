package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemoveAttorney(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	logger := &mockLogger{}

	attorney := Attorney{
		ID: "123",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	template := &mockTemplate{}
	template.
		On("Func", w, &removeAttorneyData{
			App:      appData,
			Attorney: attorney,
			Errors:   nil,
			Form:     &removeAttorneyForm{},
		}).
		Return(nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorney}}, nil)

	err := RemoveAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetRemoveAttorneyErrorOnStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	logger := &mockLogger{}
	logger.
		On("Print", "error getting lpa from store: err").
		Return(nil)

	template := &mockTemplate{}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := RemoveAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, logger)
}

func TestGetRemoveAttorneyAttorneyDoesNotExist(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=invalid-id", nil)

	logger := &mockLogger{}

	template := &mockTemplate{}

	attorney := Attorney{
		ID: "123",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorney}}, nil)

	err := RemoveAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChooseAttorneysSummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostRemoveAttorney(t *testing.T) {
	form := url.Values{
		"remove-attorney": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	logger := &mockLogger{}
	template := &mockTemplate{}

	attorneyWithAddress := Attorney{
		ID: "with-address",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	attorneyWithoutAddress := Attorney{
		ID:      "without-address",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorneyWithoutAddress, attorneyWithAddress}}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{Attorneys: []Attorney{attorneyWithAddress}}).
		Return(nil)

	err := RemoveAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChooseAttorneysSummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostRemoveAttorneyWithFormValueNo(t *testing.T) {
	form := url.Values{
		"remove-attorney": {"no"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	logger := &mockLogger{}
	template := &mockTemplate{}

	attorneyWithAddress := Attorney{
		ID: "with-address",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	attorneyWithoutAddress := Attorney{
		ID:      "without-address",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorneyWithoutAddress, attorneyWithAddress}}, nil)

	err := RemoveAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChooseAttorneysSummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostRemoveAttorneyErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"remove-attorney": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	template := &mockTemplate{}

	logger := &mockLogger{}
	logger.
		On("Print", "error removing Attorney from LPA: err").
		Return(nil)

	attorneyWithAddress := Attorney{
		ID: "with-address",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	attorneyWithoutAddress := Attorney{
		ID:      "without-address",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorneyWithoutAddress, attorneyWithAddress}}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{Attorneys: []Attorney{attorneyWithAddress}}).
		Return(expectedError)

	err := RemoveAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template, logger)
}

func TestRemoveAttorneyFormValidation(t *testing.T) {
	form := url.Values{
		"remove-attorney": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	attorneyWithoutAddress := Attorney{
		ID:      "without-address",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorneyWithoutAddress}}, nil)

	validationError := validation.With("remove-attorney", "selectRemoveAttorney")

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *removeAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveAttorney(nil, template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestRemoveAttorneyRemoveLastAttorneyRedirectsToChooseAttorney(t *testing.T) {
	form := url.Values{
		"remove-attorney": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	logger := &mockLogger{}
	template := &mockTemplate{}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{{ID: "without-address"}}, Tasks: Tasks{ChooseAttorneys: TaskCompleted}}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{Attorneys: []Attorney{}, Tasks: Tasks{ChooseAttorneys: TaskInProgress}}).
		Return(nil)

	err := RemoveAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChooseAttorneys, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
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
			},
			errors: validation.With("remove-attorney", "selectRemoveAttorney"),
		},
		"unexpected-value": {
			form: &removeAttorneyForm{
				RemoveAttorney: "not expected",
			},
			errors: validation.With("remove-attorney", "selectRemoveAttorney"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
