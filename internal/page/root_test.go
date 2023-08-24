package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoot(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	Root(nil, nil)(TestAppData, w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, Paths.Start.Format(), resp.Header.Get("Location"))
}

func TestRootNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/what", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &rootData{App: TestAppData}).
		Return(nil)

	Root(template.Execute, nil)(TestAppData, w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRootNotFoundTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/what", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &rootData{App: TestAppData}).
		Return(ExpectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", "Error rendering page: "+ExpectedError.Error())

	Root(template.Execute, logger)(TestAppData, w, r)
}
