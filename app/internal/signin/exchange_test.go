package signin

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSecretsClient struct {
	mock.Mock
}

func (m *mockSecretsClient) PrivateKey() (*rsa.PrivateKey, error) {
	args := m.Called()
	return args.Get(0).(*rsa.PrivateKey), args.Error(1)
}

func TestGetToken(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.New(rand.NewSource(99)), 2048)

	idToken, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"sub": "hey"}).SignedString(privateKey)

	response := tokenResponseBody{
		AccessToken:  "a",
		RefreshToken: "b",
		TokenType:    "Bearer",
		ExpiresIn:    1,
		IdToken:      idToken,
	}

	data, _ := json.Marshal(response)

	httpClient := &mockHttpClient{}
	httpClient.
		On("Do", mock.MatchedBy(func(r *http.Request) bool {
			return assert.Equal(t, http.MethodPost, r.Method) &&
				assert.Equal(t, "http://token", r.URL.String()) &&
				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type")) // && body
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(data)),
		}, nil)

	secretsClient := &mockSecretsClient{}
	secretsClient.
		On("PrivateKey").
		Return(privateKey, nil)

	client := &Client{
		httpClient:    httpClient,
		secretsClient: secretsClient,
		openidConfiguration: openidConfiguration{
			TokenEndpoint: "http://token",
		},
		randomString: func(i int) string { return "this-is-random" },
	}

	result, err := client.Exchange("my-code")
	assert.Nil(t, err)
	assert.Equal(t, idToken, result)

	mock.AssertExpectationsForObjects(t, httpClient, secretsClient)
}
