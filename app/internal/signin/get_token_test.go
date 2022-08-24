package signin

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSecretsClient struct {
	mock.Mock
}

func (m *mockSecretsClient) PublicKey() (*rsa.PublicKey, error) {
	args := m.Called()
	return args.Get(0).(*rsa.PublicKey), args.Error(1)
}

func (m *mockSecretsClient) PrivateKey() (*rsa.PrivateKey, error) {
	args := m.Called()
	return args.Get(0).(*rsa.PrivateKey), args.Error(1)
}

func TestGetToken(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.New(rand.NewSource(99)), 2048)

	idToken := signJwt(`{"sub":"hey"}`, privateKey)

	response := TokenResponseBody{
		AccessToken:  "a",
		RefreshToken: "b",
		TokenType:    "Bearer",
		ExpiresIn:    1,
		IdToken:      idToken,
	}

	data, _ := json.Marshal(response)

	client := &mockHttpClient{}
	client.
		On("Do", mock.MatchedBy(func(r *http.Request) bool {
			return assert.Equal(t, http.MethodPost, r.Method) &&
				assert.Equal(t, "http://token", r.URL.String()) &&
				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(data)),
		}, nil)

	secretsClient := &mockSecretsClient{}
	secretsClient.
		On("PrivateKey").
		Return(privateKey, nil)
	secretsClient.
		On("PublicKey").
		Return(&privateKey.PublicKey, nil)

	c := NewClient(client, "http://example.org", secretsClient)
	c.DiscoverData = DiscoverResponse{
		TokenEndpoint: "http://token",
	}

	result, err := c.GetToken("http://redirect", "clientId", "jti", "my-code")

	assert.Nil(t, err)
	assert.Equal(t, idToken, result)

	mock.AssertExpectationsForObjects(t, client, secretsClient)
}
