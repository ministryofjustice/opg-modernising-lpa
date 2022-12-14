package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemoveReplacementAttorney(t *testing.T) {
	w := httptest.NewRecorder()

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
			Errors:   nil,
			Form:     removeAttorneyForm{},
		}).
		Return(nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{attorney}}, nil)

	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetRemoveReplacementAttorneyErrorOnStore(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	logger.
		On("Print", "error getting lpa from store: err").
		Return(nil)

	template := &mockTemplate{}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, logger)
}

func TestGetRemoveReplacementAttorneyAttorneyDoesNotExist(t *testing.T) {
	w := httptest.NewRecorder()

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
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{attorney}}, nil)

	r, _ := http.NewRequest(http.MethodGet, "/?id=invalid-id", nil)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/choose-replacement-attorneys-summary", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostRemoveReplacementAttorney(t *testing.T) {
	w := httptest.NewRecorder()

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
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{attorneyWithoutAddress, attorneyWithAddress}}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{ReplacementAttorneys: []Attorney{attorneyWithAddress}}).
		Return(nil)

	form := url.Values{
		"remove-attorney": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/choose-replacement-attorneys-summary", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostRemoveReplacementAttorneyWithFormValueNo(t *testing.T) {
	w := httptest.NewRecorder()

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
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{attorneyWithoutAddress, attorneyWithAddress}}, nil)

	form := url.Values{
		"remove-attorney": {"no"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/choose-replacement-attorneys-summary", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostRemoveReplacementAttorneyErrorOnPutStore(t *testing.T) {
	w := httptest.NewRecorder()

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
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{attorneyWithoutAddress, attorneyWithAddress}}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{ReplacementAttorneys: []Attorney{attorneyWithAddress}}).
		Return(expectedError)

	form := url.Values{
		"remove-attorney": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template, logger)
}

func TestRemoveReplacementAttorneyFormValidation(t *testing.T) {
	w := httptest.NewRecorder()

	attorneyWithoutAddress := Attorney{
		ID:      "without-address",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{attorneyWithoutAddress}}, nil)

	validationError := map[string]string{
		"remove-attorney": "selectRemoveAttorney",
	}

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *removeReplacementAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	form := url.Values{
		"remove-attorney": {""},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := RemoveReplacementAttorney(nil, template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestRemoveReplacementAttorneyRemoveLastAttorneyRedirectsToChooseReplacementAttorney(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	template := &mockTemplate{}

	attorneyWithoutAddress := Attorney{
		ID:      "without-address",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{attorneyWithoutAddress}}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{ReplacementAttorneys: []Attorney{}}).
		Return(nil)

	form := url.Values{
		"remove-attorney": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := RemoveReplacementAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/do-you-want-replacement-attorneys", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
