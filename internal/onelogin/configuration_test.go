package onelogin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfiguration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedConfiguration := &openidConfiguration{
		AuthorizationEndpoint: "http://example.org/authorize",
		TokenEndpoint:         "http://example.org/token",
		Issuer:                "http://example.org",
		UserinfoEndpoint:      "http://example.org/userinfo",
		EndSessionEndpoint:    "http://example.org/sign-out",
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

	conf := getConfiguration(ctx, nil, http.DefaultClient, oidcServer.URL)

	assert.Equal(t, expectedConfiguration, conf.currentConfiguration)
}

func TestConfigurationClientEndpoints(t *testing.T) {
	client := &configurationClient{
		currentConfiguration: &openidConfiguration{
			AuthorizationEndpoint: "AuthorizationEndpoint",
			EndSessionEndpoint:    "EndSessionEndpoint",
			UserinfoEndpoint:      "UserinfoEndpoint",
		},
	}

	testcases := map[string]func() (string, error){
		"AuthorizationEndpoint": client.AuthorizationEndpoint,
		"EndSessionEndpoint":    client.EndSessionEndpoint,
		"UserinfoEndpoint":      client.UserinfoEndpoint,
	}

	for expected, fn := range testcases {
		t.Run(expected, func(t *testing.T) {
			endpoint, err := fn()
			assert.Nil(t, err)
			assert.Equal(t, expected, endpoint)
		})
	}
}

func TestConfigurationClientEndpointsWhenMissing(t *testing.T) {
	ch := make(chan struct{}, 1)
	client := &configurationClient{
		refreshRequest: ch,
	}

	testcases := map[string]func() (string, error){
		"AuthorizationEndpoint": client.AuthorizationEndpoint,
		"EndSessionEndpoint":    client.EndSessionEndpoint,
		"UserinfoEndpoint":      client.UserinfoEndpoint,
	}

	for expected, fn := range testcases {
		t.Run(expected, func(t *testing.T) {
			_, err := fn()
			assert.Equal(t, ErrConfigurationMissing, err)

			select {
			case <-ch:
			default:
				t.Fail()
			}
		})
	}
}

func TestConfigurationClientForExchange(t *testing.T) {
	client := &configurationClient{
		currentConfiguration: &openidConfiguration{
			TokenEndpoint: "TokenEndpoint",
			Issuer:        "Issuer",
		},
		currentJwks: &keyfunc.JWKS{},
	}

	tokenEndpoint, keyfunc, issuer, err := client.ForExchange()
	assert.Nil(t, err)
	assert.Equal(t, "TokenEndpoint", tokenEndpoint)
	assert.NotNil(t, keyfunc)
	assert.Equal(t, "Issuer", issuer)
}

func TestConfigurationClientForExchangeWhenMissing(t *testing.T) {
	ch := make(chan struct{}, 1)

	testcases := map[string]*configurationClient{
		"configuration": {
			currentJwks:    &keyfunc.JWKS{},
			refreshRequest: ch,
		},
		"jwks": {
			currentConfiguration: &openidConfiguration{
				TokenEndpoint: "TokenEndpoint",
				Issuer:        "Issuer",
			},
			refreshRequest: ch,
		},
	}

	for name, client := range testcases {
		t.Run(name, func(t *testing.T) {
			_, _, _, err := client.ForExchange()
			assert.Equal(t, ErrConfigurationMissing, err)

			select {
			case <-ch:
			default:
				t.Fail()
			}
		})
	}
}

func TestConfigurationClientBackgroundRefresh(t *testing.T) {
	ch := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())

	expectedConfiguration := &openidConfiguration{
		AuthorizationEndpoint: "http://example.org/authorize",
		TokenEndpoint:         "http://example.org/token",
		Issuer:                "http://example.org",
		UserinfoEndpoint:      "http://example.org/userinfo",
		EndSessionEndpoint:    "http://example.org/sign-out",
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

	client := &configurationClient{
		logger:         newMockLogger(t),
		httpClient:     http.DefaultClient,
		ctx:            ctx,
		refreshRequest: ch,
		now:            time.Now,
		issuer:         oidcServer.URL,
	}

	ch <- struct{}{}

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	client.backgroundRefresh()

	assert.Equal(t, expectedConfiguration, client.currentConfiguration)
}

func TestConfigurationClientBackgroundRefreshCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &configurationClient{ctx: ctx, now: time.Now}

	cancel()
	client.backgroundRefresh()
}

func TestConfigurationClientBackgroundRefreshRateLimit(t *testing.T) {
	requestCount := 0
	ch := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())

	oidcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer oidcServer.Close()

	logger := newMockLogger(t)
	logger.EXPECT().Print(mock.Anything, mock.Anything)

	client := &configurationClient{
		httpClient:     http.DefaultClient,
		logger:         logger,
		ctx:            ctx,
		refreshRequest: ch,
		now:            time.Now,
		issuer:         oidcServer.URL,
	}

	ch <- struct{}{}
	go func() {
		ch <- struct{}{}
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	client.backgroundRefresh()

	assert.Equal(t, 1, requestCount)
}
