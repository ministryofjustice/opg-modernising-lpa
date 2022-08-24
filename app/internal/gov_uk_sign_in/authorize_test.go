package govuksignin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthorizeAndRedirect(t *testing.T) {
	// Make a Client
	c := NewClient(http.DefaultClient, "http://base.url", "http://example.org")
	c.DiscoverData = buildTestDiscoverResponse()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	c.AuthorizeAndRedirect(w, r, "/redirect", "123", "state", "nonce", "scope", "http://example.org")
	resp := w.Result()

	want := "http://example.org/authorize?client_id=123&nonce=nonce&redirect_uri=%2Fredirect&scope=scope&state=state"
	got, _ := resp.Location()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, want, got.String())
}
