package signin

import (
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/url"
)

const openidConfigurationEndpoint = "/.well-known/openid-configuration"

type openidConfiguration struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	Issuer                string `json:"issuer"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
}

type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

type SecretsClient interface {
	PrivateKey() (*rsa.PrivateKey, error)
}

type Client struct {
	httpClient          Doer
	openidConfiguration openidConfiguration
	authCallbackPath    string
	secretsClient       SecretsClient
	randomString        func(int) string

	clientID    string
	redirectURL string
}

func Discover(httpClient Doer, secretsClient SecretsClient, issuer, clientID, redirectURL string) (*Client, error) {
	c := &Client{
		httpClient:    httpClient,
		secretsClient: secretsClient,
		randomString:  randomString,
		clientID:      clientID,
		redirectURL:   redirectURL,
	}

	req, err := http.NewRequest("GET", issuer+openidConfigurationEndpoint, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&c.openidConfiguration); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) AuthCodeURL(state, nonce string) string {
	q := url.Values{
		"response_type": {"code"},
		"scope":         {"openid email"},
		"redirect_uri":  {c.redirectURL},
		"client_id":     {c.clientID},
		"state":         {state},
		"nonce":         {nonce},
	}

	return c.openidConfiguration.AuthorizationEndpoint + "?" + q.Encode()
}
