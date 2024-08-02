package page

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	testcases := map[bool]*errorData{
		true:  {App: TestAppData, Err: expectedError},
		false: {App: TestAppData},
	}

	for showErrors, data := range testcases {
		t.Run(fmt.Sprint(showErrors), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequestWithContext(appcontext.ContextWithData(context.Background(), TestAppData), http.MethodGet, "/", nil)

			logger := newMockLogger(t)
			logger.EXPECT().
				ErrorContext(r.Context(), "request error", slog.Any("req", r), slog.Any("err", expectedError))

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, data).
				Return(nil)

			Error(template.Execute, logger, showErrors)(w, r, expectedError)
			resp := w.Result()

			assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		})
	}
}

func TestErrorWithErrCsrfInvalid(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(appcontext.ContextWithData(context.Background(), TestAppData), http.MethodGet, "/", nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		ErrorContext(r.Context(), "request error", slog.Any("req", r), slog.Any("err", ErrCsrfInvalid))

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &errorData{App: TestAppData}).
		Return(nil)

	Error(template.Execute, logger, false)(w, r, ErrCsrfInvalid)
	resp := w.Result()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestErrorWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(appcontext.ContextWithData(context.Background(), TestAppData), http.MethodGet, "/", nil)

	templateError := errors.New("template error")

	logger := newMockLogger(t)
	logger.EXPECT().
		ErrorContext(r.Context(), "request error", slog.Any("req", r), slog.Any("err", expectedError))
	logger.EXPECT().
		ErrorContext(r.Context(), "error rendering page", slog.Any("req", r), slog.Any("err", templateError))

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &errorData{App: TestAppData}).
		Return(templateError)

	Error(template.Execute, logger, false)(w, r, expectedError)
	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
