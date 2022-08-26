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

	secretsClient := &mockSecretsClient{}
	secretsClient.
		On("PrivateKey").
		Return(privateKey, nil)

	httpClient := &mockHttpClient{}
	httpClient.
		On("Do", mock.MatchedBy(func(r *http.Request) bool {
			var body tokenRequestBody
			json.NewDecoder(r.Body).Decode(&body)

			clientAssertion, _ := jwt.Parse(body.ClientAssertion, func(token *jwt.Token) (interface{}, error) {
				return &privateKey.PublicKey, nil
			})

			claims := clientAssertion.Claims.(jwt.MapClaims)

			return assert.Equal(t, http.MethodPost, r.Method) &&
				assert.Equal(t, "http://token", r.URL.String()) &&
				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type")) &&
				assert.Equal(t, "authorization_code", body.GrantType) &&
				assert.Equal(t, "my-code", body.AuthorizationCode) &&
				assert.Equal(t, "http://redirect", body.RedirectUri) &&
				assert.Equal(t, "urn:ietf:params:oauth:client-assertion-type:jwt-bearer", body.ClientAssertionType) &&
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
	assert.Equal(t, idToken, result)

	mock.AssertExpectationsForObjects(t, httpClient, secretsClient)
}

func TestGetTokenWhenPrivateKeyError(t *testing.T) {
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

func TestGetTokenWhenTokenRequestError(t *testing.T) {
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

func TestGetTokenWhenWrongSigningMethod(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.New(rand.NewSource(99)), 2048)

	idToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "hey"}).SignedString([]byte("my-key"))

	response := tokenResponseBody{
		AccessToken:  "a",
		RefreshToken: "b",
		TokenType:    "Bearer",
		ExpiresIn:    1,
		IdToken:      idToken,
	}

	data, _ := json.Marshal(response)

	secretsClient := &mockSecretsClient{}
	secretsClient.
		On("PrivateKey").
		Return(privateKey, nil)

	httpClient := &mockHttpClient{}
	httpClient.
		On("Do", mock.Anything).
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
		randomString: func(i int) string { return "this-is-random" },
	}

	_, err := client.Exchange("my-code")
	assert.Equal(t, "unexpected signing method: HS256", err.Error())

	mock.AssertExpectationsForObjects(t, httpClient, secretsClient)
}

func TestGetTokenWhenWrongKey(t *testing.T) {
	random := rand.New(rand.NewSource(99))
	privateKey, _ := rsa.GenerateKey(random, 2048)
	otherPrivateKey, _ := rsa.GenerateKey(random, 2048)

	idToken, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"sub": "hey"}).SignedString(otherPrivateKey)

	response := tokenResponseBody{
		AccessToken:  "a",
		RefreshToken: "b",
		TokenType:    "Bearer",
		ExpiresIn:    1,
		IdToken:      idToken,
	}

	data, _ := json.Marshal(response)

	secretsClient := &mockSecretsClient{}
	secretsClient.
		On("PrivateKey").
		Return(privateKey, nil)

	httpClient := &mockHttpClient{}
	httpClient.
		On("Do", mock.Anything).
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
		randomString: func(i int) string { return "this-is-random" },
	}

	_, err := client.Exchange("my-code")
	assert.Equal(t, "crypto/rsa: verification error", err.Error())

	mock.AssertExpectationsForObjects(t, httpClient, secretsClient)
}
