package onelogin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
)

const (
	openidConfigurationEndpoint = "/.well-known/openid-configuration"
	refreshInterval             = 24 * time.Hour
	refreshRateLimit            = time.Minute
	refreshTimeout              = 30 * time.Second
)

var ErrConfigurationMissing = errors.New("openid configuration missing")

type openidConfiguration struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	Issuer                string `json:"issuer"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
}

type configurationClient struct {
	ctx            context.Context
	logger         Logger
	httpClient     *http.Client
	issuer         string
	now            func() time.Time
	refreshRequest chan (struct{})

	currentConfiguration *openidConfiguration
	currentJwks          *keyfunc.JWKS
}

func getConfiguration(ctx context.Context, logger Logger, httpClient *http.Client, issuer string) *configurationClient {
	client := &configurationClient{
		ctx:        ctx,
		logger:     logger,
		httpClient: httpClient,
		issuer:     issuer,
		now:        time.Now,
		// only allow a single request to be waiting
		refreshRequest: make(chan struct{}, 1),
	}

	if err := client.refresh(); err != nil {
		logger.Print("error refreshing openid configuration:", err)
	}

	go client.backgroundRefresh()

	return client
}

func (c *configurationClient) AuthorizationEndpoint() (string, error) {
	if c.currentConfiguration == nil {
		c.requestRefresh()
		return "", ErrConfigurationMissing
	}

	return c.currentConfiguration.AuthorizationEndpoint, nil
}

func (c *configurationClient) EndSessionEndpoint() (string, error) {
	if c.currentConfiguration == nil {
		c.requestRefresh()
		return "", ErrConfigurationMissing
	}

	return c.currentConfiguration.EndSessionEndpoint, nil
}

func (c *configurationClient) UserinfoEndpoint() (string, error) {
	if c.currentConfiguration == nil {
		c.requestRefresh()
		return "", ErrConfigurationMissing
	}

	return c.currentConfiguration.UserinfoEndpoint, nil
}

func (c *configurationClient) ForExchange() (tokenEndpoint string, keyfunc jwt.Keyfunc, issuer string, err error) {
	if c.currentConfiguration == nil || c.currentJwks == nil {
		c.requestRefresh()
		return "", nil, "", ErrConfigurationMissing
	}

	return c.currentConfiguration.TokenEndpoint, c.currentJwks.Keyfunc, c.currentConfiguration.Issuer, nil
}

// requestRefresh will request that the configuration is refreshed, if no other request is waiting
func (c *configurationClient) requestRefresh() {
	select {
	case c.refreshRequest <- struct{}{}:
	default:
	}
}

// refresh updates the current configuration
func (c *configurationClient) refresh() error {
	ctx, cancel := context.WithTimeout(c.ctx, refreshTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.issuer+openidConfigurationEndpoint, nil)
	if err != nil {
		return err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var v openidConfiguration
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return err
	}

	c.currentConfiguration = &v
	c.currentJwks, err = keyfunc.Get(c.currentConfiguration.JwksURI, keyfunc.Options{
		Client: c.httpClient,
		Ctx:    c.ctx,
		RefreshErrorHandler: func(err error) {
			c.logger.Print("error refreshing jwks: ", err)
		},
		RefreshInterval:   refreshInterval,
		RefreshRateLimit:  refreshRateLimit,
		RefreshTimeout:    refreshTimeout,
		RefreshUnknownKID: true,
	})

	return err
}

// backgroundRefresh triggers refresh periodically on refreshInterval, or when requested limited by refreshRateLimit
func (c *configurationClient) backgroundRefresh() {
	lastRefresh := c.now().Add(-refreshRateLimit)

	for {
		select {
		case <-time.After(refreshInterval):
			c.requestRefresh()

		case <-c.refreshRequest:
			if lastRefresh.Add(refreshRateLimit).After(c.now()) {
				continue
			}

			if err := c.refresh(); err != nil {
				c.logger.Print("error refreshing openid configuration: ", err.Error())
			}

			lastRefresh = c.now()

		case <-c.ctx.Done():
			return
		}
	}
}
