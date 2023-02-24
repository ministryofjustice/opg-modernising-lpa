package page

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ContextWithAppData(context.Background(), TestAppData), http.MethodGet, "/", nil)

	logger := newMockLogger(t)
	logger.
		On("Print", ExpectedError)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &errorData{App: TestAppData}).
		Return(nil)

	Error(template.Execute, logger)(w, r, ExpectedError)
	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestErrorWithErrCsrfInvalid(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ContextWithAppData(context.Background(), TestAppData), http.MethodGet, "/", nil)

	logger := newMockLogger(t)
	logger.
		On("Print", ErrCsrfInvalid)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &errorData{App: TestAppData}).
		Return(nil)

	Error(template.Execute, logger)(w, r, ErrCsrfInvalid)
	resp := w.Result()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestErrorWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ContextWithAppData(context.Background(), TestAppData), http.MethodGet, "/", nil)

	templateError := errors.New("template error")

	logger := newMockLogger(t)
	logger.
		On("Print", ExpectedError)
	logger.
		On("Print", "Error rendering page: template error")

	template := newMockTemplate(t)
	template.
		On("Execute", w, &errorData{App: TestAppData}).
		Return(templateError)

	Error(template.Execute, logger)(w, r, ExpectedError)
	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
