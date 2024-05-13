package onelogin

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
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
	currentJwks          keyfunc.Keyfunc
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
		logger.WarnContext(ctx, "problem refreshing openid configuration:", slog.Any("err", err))
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

	uri, err := url.ParseRequestURI(v.JwksURI)
	if err != nil {
		return err
	}

	storage, err := jwkset.NewStorageFromHTTP(uri, jwkset.HTTPClientStorageOptions{
		Ctx:                       c.ctx,
		Client:                    c.httpClient,
		RefreshInterval:           refreshInterval,
		HTTPTimeout:               refreshTimeout,
		NoErrorReturnFirstHTTPReq: true,
		RefreshErrorHandler: func(_ context.Context, err error) {
			c.logger.WarnContext(ctx, "problem refreshing jwks", slog.Any("err", err))
		},
	})
	if err != nil {
		return err
	}

	client, err := jwkset.NewHTTPClient(jwkset.HTTPClientOptions{
		HTTPURLs: map[string]jwkset.Storage{
			uri.String(): storage,
		},
		RefreshUnknownKID: rate.NewLimiter(rate.Every(refreshRateLimit), 1),
	})
	if err != nil {
		return err
	}

	c.currentConfiguration = &v
	c.currentJwks, err = keyfunc.New(keyfunc.Options{Ctx: c.ctx, Storage: client})

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
				c.logger.WarnContext(c.ctx, "problem refreshing openid configuration", slog.Any("err", err.Error()))
			}

			lastRefresh = c.now()

		case <-c.ctx.Done():
			return
		}
	}
}
