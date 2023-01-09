package onelogin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockHttpClient struct {
	mock.Mock
}

func (m *mockHttpClient) Do(r *http.Request) (*http.Response, error) {
	args := m.Called(r)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestDiscover(t *testing.T) {
	expectedConfiguration := openidConfiguration{
		AuthorizationEndpoint: "http://example.org/authorize",
		TokenEndpoint:         "http://example.org/token",
		Issuer:                "http://example.org",
		UserinfoEndpoint:      "http://example.org/userinfo",
	}

	oidcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(expectedConfiguration)
		case "/.well-known/jwks":
			w.Write([]byte(`{"keys":[{"kty":"EC","use":"sig","crv":"P-256","kid":"644af598b780f54106ca0f3c017341bc230c4f8373f35f32e18e3e40cc7acff6","x":"5URVCgH4HQgkg37kiipfOGjyVft0R5CdjFJahRoJjEw","y":"QzrvsnDy3oY1yuz55voaAq9B1M5tfhgW3FBjh_n_F0U","alg":"ES256"},{"kty":"EC","use":"sig","crv":"P-256","kid":"e1f5699d068448882e7866b49d24431b2f21bf1a8f3c2b2dde8f4066f0506f1b","x":"BJnIZvnzJ9D_YRu5YL8a3CXjBaa5AxlX1xSeWDLAn9k","y":"x4FU3lRtkeDukSWVJmDuw2nHVFVIZ8_69n4bJ6ik4bQ","alg":"ES256"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer oidcServer.Close()

	expectedConfiguration.JwksURI = oidcServer.URL + "/.well-known/jwks"

	c, err := Discover(context.Background(), nil, http.DefaultClient, nil, oidcServer.URL, "client-id", "http://redirect")

	assert.Nil(t, err)
	assert.Equal(t, expectedConfiguration, c.openidConfiguration)
}

func TestAuthCodeURL(t *testing.T) {
	expected := "http://auth?client_id=123&nonce=nonce&redirect_uri=http%3A%2F%2Fredirect&response_type=code&scope=openid+email&state=state&ui_locales=cy"

	c := &Client{
		redirectURL: "http://redirect",
		clientID:    "123",
		openidConfiguration: openidConfiguration{
			AuthorizationEndpoint: "http://auth",
		},
	}
	actual := c.AuthCodeURL("state", "nonce", "cy")

	assert.Equal(t, expected, actual)
}
