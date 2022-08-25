package signin

import (
	"crypto/rsa"
	"net/http"
)

type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

type Client struct {
	httpClient       Doer
	discoverData     DiscoverResponse
	authCallbackPath string
	secretsClient    SecretsClient
}

type DiscoverResponse struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	Issuer                string `json:"issuer"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
}

type SecretsClient interface {
	PrivateKey() (*rsa.PrivateKey, error)
}

func NewClient(httpClient Doer, secretsClient SecretsClient) *Client {
	client := &Client{
		httpClient:    httpClient,
		secretsClient: secretsClient,
	}

	return client
}
