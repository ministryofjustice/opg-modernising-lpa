// Package onelogin provides a client for GOV.UK's One Login service.
package onelogin

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
)

var expectedError = errors.New("err")

type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

type Logger interface {
	WarnContext(ctx context.Context, msg string, args ...any)
}

type SecretsClient interface {
	SecretBytes(ctx context.Context, name string) ([]byte, error)
}

type IdentityPublicKeyFunc func(context.Context) (*ecdsa.PublicKey, error)

type Client struct {
	ctx                 context.Context
	logger              Logger
	httpClient          Doer
	openidConfiguration *configurationClient
	secretsClient       SecretsClient
	randomString        func(int) string
	didClient           *didClient

	clientID    string
	redirectURL string
}

func New(ctx context.Context, logger Logger, httpClient *http.Client, secretsClient SecretsClient, issuer, identityURL, clientID, redirectURL string) *Client {
	return &Client{
		ctx:                 ctx,
		logger:              logger,
		httpClient:          httpClient,
		secretsClient:       secretsClient,
		randomString:        random.String,
		clientID:            clientID,
		redirectURL:         redirectURL,
		openidConfiguration: getConfiguration(ctx, logger, httpClient, issuer),
		didClient:           getDID(ctx, logger, httpClient, identityURL),
	}
}

func (c *Client) AuthCodeURL(state, nonce, locale string, identity bool) (string, error) {
	q := url.Values{
		"response_type": {"code"},
		"scope":         {"openid email"},
		"redirect_uri":  {c.redirectURL},
		"client_id":     {c.clientID},
		"state":         {state},
		"nonce":         {nonce},
		"ui_locales":    {locale},
	}

	if identity {
		q.Add("vtr", `["Cl.Cm.P2"]`)
		q.Add("claims", `{"userinfo":{"https://vocab.account.gov.uk/v1/coreIdentityJWT": null,"https://vocab.account.gov.uk/v1/returnCode": null,"https://vocab.account.gov.uk/v1/address": null}}`)
	}

	endpoint, err := c.openidConfiguration.AuthorizationEndpoint()
	if err != nil {
		return "", err
	}

	return endpoint + "?" + q.Encode(), nil
}

func (c *Client) EndSessionURL(idToken, postLogoutURL string) (string, error) {
	endpoint, err := c.openidConfiguration.EndSessionEndpoint()
	if err != nil {
		return "", err
	}

	return endpoint + "?" + url.Values{
		"id_token_hint":            {idToken},
		"post_logout_redirect_uri": {postLogoutURL},
	}.Encode(), nil
}

func (c *Client) CheckHealth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.openidConfiguration.issuer, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	return resp.Body.Close()
}
