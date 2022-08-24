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
	DiscoverData     DiscoverResponse
	AuthCallbackPath string
	secretsClient    SecretsClient
}

type DiscoverResponse struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	Issuer                string `json:"issuer"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
}

type SecretsClient interface {
	PublicKey() (*rsa.PublicKey, error)
	PrivateKey() (*rsa.PrivateKey, error)
}

func NewClient(httpClient Doer, authCallbackPath string, secretsClient SecretsClient) *Client {
	client := &Client{
		httpClient:       httpClient,
		AuthCallbackPath: authCallbackPath,
		secretsClient:    secretsClient,
	}

	return client
}
