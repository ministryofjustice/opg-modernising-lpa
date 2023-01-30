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

func TestGetRemoveReplacementAttorney(t *testing.T) {
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
		On("Func", w, &removeReplacementAttorneyData{
			App:      appData,
			Attorney: attorney,
			Form:     &removeAttorneyForm{},
		}).
		Return(nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{ReplacementAttorneys: []Attorney{attorney}}, nil)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetRemoveReplacementAttorneyErrorOnStore(t *testing.T) {
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

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, logger)
}

func TestGetRemoveReplacementAttorneyAttorneyDoesNotExist(t *testing.T) {
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
		Return(&Lpa{ReplacementAttorneys: []Attorney{attorney}}, nil)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostRemoveReplacementAttorney(t *testing.T) {
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
		Return(&Lpa{ReplacementAttorneys: []Attorney{attorneyWithoutAddress, attorneyWithAddress}}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{ReplacementAttorneys: []Attorney{attorneyWithAddress}}).
		Return(nil)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostRemoveReplacementAttorneyWithFormValueNo(t *testing.T) {
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
		Return(&Lpa{ReplacementAttorneys: []Attorney{attorneyWithoutAddress, attorneyWithAddress}}, nil)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostRemoveReplacementAttorneyErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"remove-attorney": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	template := &mockTemplate{}

	logger := &mockLogger{}
	logger.
		On("Print", "error removing replacement Attorney from LPA: err").
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
		Return(&Lpa{ReplacementAttorneys: []Attorney{attorneyWithoutAddress, attorneyWithAddress}}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{ReplacementAttorneys: []Attorney{attorneyWithAddress}}).
		Return(expectedError)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template, logger)
}

func TestRemoveReplacementAttorneyFormValidation(t *testing.T) {
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
		Return(&Lpa{ReplacementAttorneys: []Attorney{attorneyWithoutAddress}}, nil)

	validationError := validation.With("remove-attorney", "selectRemoveAttorney")

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *removeReplacementAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveReplacementAttorney(nil, template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestRemoveReplacementAttorneyRemoveLastAttorneyRedirectsToChooseReplacementAttorney(t *testing.T) {
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
		Return(&Lpa{ReplacementAttorneys: []Attorney{{ID: "without-address"}}, Tasks: Tasks{ChooseReplacementAttorneys: TaskCompleted}}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{ReplacementAttorneys: []Attorney{}, Tasks: Tasks{ChooseReplacementAttorneys: TaskInProgress}}).
		Return(nil)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.DoYouWantReplacementAttorneys, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
