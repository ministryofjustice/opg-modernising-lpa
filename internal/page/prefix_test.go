package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteToPrefix(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/lpa/123/somewhere%2Fwhat", nil)

	RouteToPrefix("/lpa/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/somewhere/what", r.URL.Path)
		assert.Equal(t, "/somewhere%2Fwhat", r.URL.RawPath)

		w.WriteHeader(http.StatusTeapot)
	}), nil).ServeHTTP(w, r)

	res := w.Result()

	assert.Equal(t, http.StatusTeapot, res.StatusCode)
}

func TestRouteToPrefixWithoutID(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/lpa/", nil)

	notFoundHandler := newMockHandler(t)
	notFoundHandler.EXPECT().
		Execute(AppData{}, w, r).
		Return(nil)

	RouteToPrefix("/lpa/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}), notFoundHandler.Execute).ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}
