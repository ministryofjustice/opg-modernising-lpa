package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowShouldReplacementAttorneysStepIn(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldReplacementAttorneysStepInData{
			App: appData,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}
