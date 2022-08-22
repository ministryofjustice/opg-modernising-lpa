package page

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTemplate struct {
	mock.Mock
}

func (m *mockTemplate) Func(w io.Writer, data interface{}) error {
	args := m.Called(w, data)
	return args.Error(0)
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Print(v ...interface{}) {
	m.Called(v...)
}

func TestStart(t *testing.T) {
	w := httptest.NewRecorder()

	localizer := localize.Localizer{}

	template := &mockTemplate{}
	template.
		On("Func", w, startData{L: localizer, Lang: En}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	Start(nil, localizer, En, template.Func)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestStartWhenTemplateErrors(t *testing.T) {
	expectedError := errors.New("err")
	w := httptest.NewRecorder()

	localizer := localize.Localizer{}

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)
	template := &mockTemplate{}
	template.
		On("Func", w, startData{L: localizer, Lang: En}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	Start(logger, localizer, En, template.Func)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mock.AssertExpectationsForObjects(t, template, logger)
}
