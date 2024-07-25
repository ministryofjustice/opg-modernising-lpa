package onelogin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthCodeURL(t *testing.T) {
	expected := "http://auth?client_id=123&nonce=nonce&redirect_uri=http%3A%2F%2Fredirect&response_type=code&scope=openid+email&state=state&ui_locales=cy"

	c := &Client{
		redirectURL: "http://redirect",
		clientID:    "123",
		openidConfiguration: &configurationClient{
			currentConfiguration: &openidConfiguration{
				AuthorizationEndpoint: "http://auth",
			},
		},
	}
	actual, err := c.AuthCodeURL("state", "nonce", "cy", false)

	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestAuthCodeURLForIdentity(t *testing.T) {
	expected := "http://auth?claims=%7B%22userinfo%22%3A%7B%22https%3A%2F%2Fvocab.account.gov.uk%2Fv1%2FcoreIdentityJWT%22%3Anull%2C%22https%3A%2F%2Fvocab.account.gov.uk%2Fv1%2FreturnCode%22%3Anull%2C%22https%3A%2F%2Fvocab.account.gov.uk%2Fv1%2Faddress%22%3Anull%7D%7D&client_id=123&nonce=nonce&redirect_uri=http%3A%2F%2Fredirect&response_type=code&scope=openid+email&state=state&ui_locales=cy&vtr=%5B%22Cl.Cm.P2%22%5D"

	c := &Client{
		redirectURL: "http://redirect",
		clientID:    "123",
		openidConfiguration: &configurationClient{
			currentConfiguration: &openidConfiguration{
				AuthorizationEndpoint: "http://auth",
			},
		},
	}
	actual, err := c.AuthCodeURL("state", "nonce", "cy", true)

	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestAuthCodeURLWhenConfigurationMissing(t *testing.T) {
	c := &Client{
		openidConfiguration: &configurationClient{},
	}
	_, err := c.AuthCodeURL("state", "nonce", "cy", false)

	assert.Equal(t, ErrConfigurationMissing, err)
}

func TestEndSessionURL(t *testing.T) {
	expected := "http://end?id_token_hint=id-token&post_logout_redirect_uri=http%3A%2F%2Fafter"

	c := &Client{
		openidConfiguration: &configurationClient{
			currentConfiguration: &openidConfiguration{
				EndSessionEndpoint: "http://end",
			},
		},
	}

	actual, err := c.EndSessionURL("id-token", "http://after")

	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestEndSessionURLWhenConfigurationMissing(t *testing.T) {
	c := &Client{
		openidConfiguration: &configurationClient{},
	}
	_, err := c.EndSessionURL("id-token", "http://after")

	assert.Equal(t, ErrConfigurationMissing, err)
}

func TestCheckHealth(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer s.Close()

	c := &Client{
		httpClient: http.DefaultClient,
		openidConfiguration: &configurationClient{
			issuer: s.URL,
		},
	}

	assert.Nil(t, c.CheckHealth(context.Background()))
}

func TestCheckHealthWhenError(t *testing.T) {
	c := &Client{
		httpClient: http.DefaultClient,
		openidConfiguration: &configurationClient{
			issuer: "some-rubbish",
		},
	}

	assert.NotNil(t, c.CheckHealth(context.Background()))
}
