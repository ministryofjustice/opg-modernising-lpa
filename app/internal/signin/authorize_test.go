package signin

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthCodeURL(t *testing.T) {
	c := NewClient(http.DefaultClient, nil)
	c.discoverData = DiscoverResponse{
		AuthorizationEndpoint: "http://example.org/authorize",
	}

	got := c.AuthCodeURL("/redirect", "123", "state", "nonce", "scope", "http://example.org")

	want := "http://example.org/authorize?client_id=123&nonce=nonce&redirect_uri=%2Fredirect&scope=scope&state=state"

	assert.Equal(t, want, got)
}
