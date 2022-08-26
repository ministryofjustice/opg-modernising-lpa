package signin

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v4"
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

func TestExchange(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.New(rand.NewSource(99)), 2048)

	response := tokenResponseBody{
		AccessToken: "a",
		TokenType:   "Bearer",
		IdToken:     "b",
	}

	data, _ := json.Marshal(response)

	secretsClient := &mockSecretsClient{}
	secretsClient.
		On("PrivateKey").
		Return(privateKey, nil)

	httpClient := &mockHttpClient{}
	httpClient.
		On("Do", mock.MatchedBy(func(r *http.Request) bool {
			clientAssertion, _ := jwt.Parse(r.FormValue("client_assertion"), func(token *jwt.Token) (interface{}, error) {
				return &privateKey.PublicKey, nil
			})

			claims := clientAssertion.Claims.(jwt.MapClaims)

			return assert.Equal(t, http.MethodPost, r.Method) &&
				assert.Equal(t, "http://token", r.URL.String()) &&
				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type")) &&
				assert.Equal(t, "client-id", r.FormValue("client_id")) &&
				assert.Equal(t, "authorization_code", r.FormValue("grant_type")) &&
				assert.Equal(t, "my-code", r.FormValue("code")) &&
				assert.Equal(t, "http://redirect", r.FormValue("redirect_uri")) &&
				assert.Equal(t, "urn:ietf:params:oauth:client-assertion-type:jwt-bearer", r.FormValue("client_assertion_type")) &&
				assert.Equal(t, []interface{}{"https://oidc.integration.account.gov.uk/token"}, claims["aud"]) &&
				assert.Equal(t, "client-id", claims["iss"]) &&
				assert.Equal(t, "client-id", claims["sub"]) &&
				assert.Equal(t, "this-is-random", claims["jti"])
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(data)),
		}, nil)

	client := &Client{
		httpClient:    httpClient,
		secretsClient: secretsClient,
		openidConfiguration: openidConfiguration{
			TokenEndpoint: "http://token",
		},
		clientID:     "client-id",
		redirectURL:  "http://redirect",
		randomString: func(i int) string { return "this-is-random" },
	}

	result, err := client.Exchange("my-code")
	assert.Nil(t, err)
	assert.Equal(t, "a", result)

	mock.AssertExpectationsForObjects(t, httpClient, secretsClient)
}

func TestExchangeWhenPrivateKeyError(t *testing.T) {
	expectedError := errors.New("err")

	secretsClient := &mockSecretsClient{}
	secretsClient.
		On("PrivateKey").
		Return(&rsa.PrivateKey{}, expectedError)

	client := &Client{
		secretsClient: secretsClient,
	}

	_, err := client.Exchange("my-code")
	assert.Equal(t, expectedError, err)

	mock.AssertExpectationsForObjects(t, secretsClient)
}

func TestExchangeWhenTokenRequestError(t *testing.T) {
	expectedError := errors.New("err")

	privateKey, _ := rsa.GenerateKey(rand.New(rand.NewSource(99)), 2048)

	secretsClient := &mockSecretsClient{}
	secretsClient.
		On("PrivateKey").
		Return(privateKey, nil)

	httpClient := &mockHttpClient{}
	httpClient.
		On("Do", mock.Anything).
		Return(&http.Response{}, expectedError)

	client := &Client{
		httpClient:    httpClient,
		secretsClient: secretsClient,
		openidConfiguration: openidConfiguration{
			TokenEndpoint: "http://token",
		},
		randomString: func(i int) string { return "this-is-random" },
	}

	_, err := client.Exchange("my-code")
	assert.Equal(t, expectedError, err)

	mock.AssertExpectationsForObjects(t, httpClient, secretsClient)
}
