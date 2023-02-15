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

	logger := &MockLogger{}

	attorney := actor.Attorney{
		ID: "123",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	template := &MockTemplate{}
	template.
		On("Func", w, &removeReplacementAttorneyData{
			App:      TestAppData,
			Attorney: attorney,
			Form:     &removeAttorneyForm{},
		}).
		Return(nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{attorney}}, nil)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(TestAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetRemoveReplacementAttorneyErrorOnStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	logger := &MockLogger{}
	logger.
		On("Print", "error getting lpa from store: err").
		Return(nil)

	template := &MockTemplate{}

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, ExpectedError)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(TestAppData, w, r)

	resp := w.Result()

	assert.Equal(t, ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, logger)
}

func TestGetRemoveReplacementAttorneyAttorneyDoesNotExist(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=invalid-id", nil)

	logger := &MockLogger{}
	template := &MockTemplate{}

	attorney := actor.Attorney{
		ID: "123",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{attorney}}, nil)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(TestAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostRemoveReplacementAttorney(t *testing.T) {
	form := url.Values{
		"remove-attorney": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := &MockLogger{}
	template := &MockTemplate{}

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

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{attorneyWithoutAddress, attorneyWithAddress}}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{ReplacementAttorneys: actor.Attorneys{attorneyWithAddress}}).
		Return(nil)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(TestAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostRemoveReplacementAttorneyWithFormValueNo(t *testing.T) {
	form := url.Values{
		"remove-attorney": {"no"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := &MockLogger{}
	template := &MockTemplate{}

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

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{attorneyWithoutAddress, attorneyWithAddress}}, nil)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(TestAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostRemoveReplacementAttorneyErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"remove-attorney": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := &MockTemplate{}

	logger := &MockLogger{}
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

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{attorneyWithoutAddress, attorneyWithAddress}}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{ReplacementAttorneys: actor.Attorneys{attorneyWithAddress}}).
		Return(ExpectedError)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(TestAppData, w, r)

	resp := w.Result()

	assert.Equal(t, ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template, logger)
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

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{attorneyWithoutAddress}}, nil)

	validationError := validation.With("remove-attorney", validation.SelectError{Label: "yesToRemoveReplacementAttorney"})

	template := &MockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *removeReplacementAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveReplacementAttorney(nil, template.Func, lpaStore)(TestAppData, w, r)
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
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := &MockLogger{}
	template := &MockTemplate{}

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{{ID: "without-address"}}, Tasks: page.Tasks{ChooseReplacementAttorneys: page.TaskCompleted}}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{ReplacementAttorneys: actor.Attorneys{}, Tasks: page.Tasks{ChooseReplacementAttorneys: page.TaskInProgress}}).
		Return(nil)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(TestAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.DoYouWantReplacementAttorneys, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
