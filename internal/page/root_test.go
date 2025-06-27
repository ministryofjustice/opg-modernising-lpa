package page

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoot(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	Root(nil, nil, "http://example.com/somewhere")(testAppData, w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://example.com/somewhere", resp.Header.Get("Location"))
}

func TestRootNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/what", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &rootData{App: testAppData}).
		Return(nil)

	Root(template.Execute, nil, "")(testAppData, w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRootNotFoundTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/what", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &rootData{App: testAppData}).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		ErrorContext(r.Context(), "error rendering page", slog.Any("req", r), slog.Any("err", expectedError))

	Root(template.Execute, logger, "")(testAppData, w, r)
}
