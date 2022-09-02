package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStart(t *testing.T) {
	w := httptest.NewRecorder()
	appData := AppData{}

	template := &mockTemplate{}
	template.
		On("Func", w, &startData{App: appData}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	Start(nil, template.Func)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestStartWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	appData := AppData{}

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)
	template := &mockTemplate{}
	template.
		On("Func", w, &startData{App: appData}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	Start(logger, template.Func)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mock.AssertExpectationsForObjects(t, template, logger)
}
