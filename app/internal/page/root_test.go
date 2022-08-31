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

	Root()(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, startPath, resp.Header.Get("Location"))
}

func TestRootNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/what", nil)

	Root()(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
