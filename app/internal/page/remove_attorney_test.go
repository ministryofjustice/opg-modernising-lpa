package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemoveAttorney(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}

	template := &mockTemplate{}
	template.
		On("Func", w, &removeAttorneyData{
			App:      appData,
			Attorney: attorneyWithAddress,
			Errors:   nil,
			Form:     removeAttorneyForm{},
		}).
		Return(nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{Attorneys: []Attorney{attorneyWithAddress}}, nil)

	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	err := RemoveAttorney(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
