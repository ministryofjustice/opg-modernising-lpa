package onelogin

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
)

var expectedError = errors.New("err")

const openidConfigurationEndpoint = "/.well-known/openid-configuration"

type openidConfiguration struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	Issuer                string `json:"issuer"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
}

//go:generate mockery --testonly --inpackage --name Doer --structname mockHttpClient
type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

type Logger interface {
	Print(v ...interface{})
}

//go:generate mockery --testonly --inpackage --name SecretsClient --structname mockSecretsClient
type SecretsClient interface {
	SecretBytes(ctx context.Context, name string) ([]byte, error)
}

type IdentityPublicKeyFunc func(context.Context) (*ecdsa.PublicKey, error)

type Client struct {
	ctx                   context.Context
	logger                Logger
	httpClient            Doer
	openidConfiguration   openidConfiguration
	secretsClient         SecretsClient
	randomString          func(int) string
	jwks                  *keyfunc.JWKS
	identityPublicKeyFunc IdentityPublicKeyFunc

	clientID    string
	redirectURL string
}

func Discover(ctx context.Context, logger Logger, httpClient *http.Client, secretsClient SecretsClient, issuer, clientID, redirectURL string, identityPublicKeyFunc IdentityPublicKeyFunc) (*Client, error) {
	c := &Client{
		ctx:                   ctx,
		logger:                logger,
		httpClient:            httpClient,
		secretsClient:         secretsClient,
		randomString:          random.String,
		identityPublicKeyFunc: identityPublicKeyFunc,
		clientID:              clientID,
		redirectURL:           redirectURL,
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

	c.jwks, err = keyfunc.Get(c.openidConfiguration.JwksURI, keyfunc.Options{
		Client: httpClient,
		Ctx:    c.ctx,
		RefreshErrorHandler: func(err error) {
			c.logger.Print("error refreshing jwks:", err)
		},
		RefreshInterval:   24 * time.Hour,
		RefreshRateLimit:  5 * time.Minute,
		RefreshTimeout:    30 * time.Second,
		RefreshUnknownKID: true,
	})

	return c, err
}

func (c *Client) AuthCodeURL(state, nonce, locale string, identity bool) string {
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
		q.Add("vtr", "[Cl.Cm.P2]")
		q.Add("claims", `{"userinfo":{"https://vocab.account.gov.uk/v1/coreIdentityJWT": null}}`)
	}

	return c.openidConfiguration.AuthorizationEndpoint + "?" + q.Encode()
}

func (c *Client) EndSessionURL(idToken, postLogoutURL string) string {
	return c.openidConfiguration.EndSessionEndpoint + "?" + url.Values{
		"id_token_hint":            {idToken},
		"post_logout_redirect_uri": {postLogoutURL},
	}.Encode()
}
