package signin

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
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
	body, _ := json.Marshal(expectedConfiguration)

	client := &mockHttpClient{}
	client.
		On("Do", mock.MatchedBy(func(r *http.Request) bool {
			return assert.Equal(t, http.MethodGet, r.Method) && assert.Equal(t, "http://base.uri/.well-known/openid-configuration", r.URL.String())
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}, nil)

	c, err := Discover(client, nil, "http://base.uri", "client-id", "http://redirect")

	assert.Nil(t, err)
	assert.Equal(t, expectedConfiguration, c.openidConfiguration)
	mock.AssertExpectationsForObjects(t, client)
}

func TestAuthCodeURL(t *testing.T) {
	expected := "http://auth?client_id=123&nonce=nonce&redirect_uri=http%3A%2F%2Fredirect&response_type=code&scope=openid+email&state=state"

	c := &Client{
		redirectURL: "http://redirect",
		clientID:    "123",
		openidConfiguration: openidConfiguration{
			AuthorizationEndpoint: "http://auth",
		},
	}
	actual := c.AuthCodeURL("state", "nonce")

	assert.Equal(t, expected, actual)
}
