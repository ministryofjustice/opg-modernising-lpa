package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheControlHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	CacheControlHeaders(http.NotFoundHandler()).ServeHTTP(w, r)

	resp := w.Result()
	assert.Equal(t, "max-age=2592000", resp.Header.Get("Cache-Control"))
}
