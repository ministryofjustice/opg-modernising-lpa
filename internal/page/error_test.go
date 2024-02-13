package page

import (
	"context"
	"errors"
	"log/slog"
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
		Error("request error", slog.Any("req", r), slog.Any("err", expectedError))

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
		Error("request error", slog.Any("req", r), slog.Any("err", ErrCsrfInvalid))

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
		Error("request error", slog.Any("req", r), slog.Any("err", expectedError))
	logger.EXPECT().
		Error("error rendering page", slog.Any("req", r), slog.Any("err", templateError))

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &errorData{App: TestAppData}).
		Return(templateError)

	Error(template.Execute, logger)(w, r, expectedError)
	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
