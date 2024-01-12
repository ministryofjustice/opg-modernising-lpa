package page

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ContextWithAppData(context.Background(), TestAppData), http.MethodGet, "/", nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Request(r, expectedError)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &errorData{App: TestAppData}).
		Return(nil)

	Error(template.Execute, logger)(w, r, expectedError)
	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestErrorWithErrCsrfInvalid(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ContextWithAppData(context.Background(), TestAppData), http.MethodGet, "/", nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Request(r, ErrCsrfInvalid)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &errorData{App: TestAppData}).
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
	logger.EXPECT().
		Request(r, expectedError)
	logger.EXPECT().
		Request(r, fmt.Errorf("Error rendering page: %w", templateError))

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &errorData{App: TestAppData}).
		Return(templateError)

	Error(template.Execute, logger)(w, r, expectedError)
	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
