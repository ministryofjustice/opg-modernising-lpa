package onelogin

import (
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MicahParks/jwkset"
)

const didDocumentEndpoint = "/.well-known/did.json"

type didDocument struct {
	Context          []string             `json:"@context"`
	ID               string               `json:"id"`
	AssertionMethods []didAssertionMethod `json:"assertionMethod"`
}

type didAssertionMethod struct {
	Type         string            `json:"type"`
	ID           string            `json:"id"`
	Controller   string            `json:"controller"`
	PublicKeyJWK jwkset.JWKMarshal `json:"publicKeyJwk"`
}

type didClient struct {
	ctx              context.Context
	identityURL      string
	http             Doer
	logger           Logger
	now              func() time.Time
	refreshRateLimit time.Duration
	refreshRequest   chan (struct{})

	controllerID     string
	assertionMethods map[string]crypto.PublicKey
}

func getDID(ctx context.Context, logger Logger, httpClient Doer, identityURL string) *didClient {
	client := &didClient{
		ctx:              ctx,
		identityURL:      identityURL,
		http:             httpClient,
		logger:           logger,
		now:              time.Now,
		refreshRateLimit: refreshRateLimit,
		// only allow a single request to be waiting
		refreshRequest: make(chan struct{}, 1),
	}

	go client.backgroundRefresh()

	return client
}

// ForKID retrieves the public key for the given kid.
func (c *didClient) ForKID(kid string) (crypto.PublicKey, error) {
	if c.controllerID == "" {
		c.requestRefresh()
		return nil, ErrConfigurationMissing
	}

	controllerID, _, found := strings.Cut(kid, "#")
	if !found {
		return nil, fmt.Errorf("malformed kid missing '#'")
	}

	if c.controllerID != controllerID {
		return nil, fmt.Errorf("controller id does not match: %s != %s", c.controllerID, controllerID)
	}

	publicKey, ok := c.assertionMethods[kid]
	if !ok {
		return nil, fmt.Errorf("missing jwk for kid %s", kid)
	}

	return publicKey, nil
}

// refresh updates the did documents.
func (c *didClient) refresh() (time.Duration, error) {
	const errRefresh = time.Minute

	ctx, cancel := context.WithTimeout(c.ctx, refreshTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.identityURL+didDocumentEndpoint, nil)
	if err != nil {
		return errRefresh, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return errRefresh, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errRefresh, fmt.Errorf("unexpected response status code %d for %s", resp.StatusCode, c.identityURL+didDocumentEndpoint)
	}

	maxAge, ok := parseCacheControl(resp.Header.Get("Cache-Control"))
	if !ok {
		maxAge = refreshInterval
	}

	var document didDocument
	if err := json.NewDecoder(resp.Body).Decode(&document); err != nil {
		return errRefresh, err
	}

	assertionMethods := map[string]crypto.PublicKey{}

	for _, method := range document.AssertionMethods {
		jwk, err := jwkset.NewJWKFromMarshal(method.PublicKeyJWK, jwkset.JWKMarshalOptions{}, jwkset.JWKValidateOptions{})
		if err != nil {
			return errRefresh, fmt.Errorf("could not unmarshal public key jwk for %s: %w", method.ID, err)
		}

		assertionMethods[method.ID] = jwk.Key().(crypto.PublicKey)
	}

	c.controllerID = document.ID
	c.assertionMethods = assertionMethods

	return maxAge, nil
}

// requestRefresh will request that the DID document is refreshed, if no other request is waiting
func (c *didClient) requestRefresh() {
	select {
	case c.refreshRequest <- struct{}{}:
	default:
	}
}

func (c *didClient) backgroundRefresh() {
	var (
		lastRefresh time.Time
		refreshIn   time.Duration
		err         error
	)

	for {
		select {
		case <-time.After(refreshIn):
			c.requestRefresh()

		case <-c.refreshRequest:
			if lastRefresh.Add(c.refreshRateLimit).After(c.now()) {
				continue
			}

			refreshIn, err = c.refresh()
			if err != nil {
				c.logger.WarnContext(c.ctx, "problem refreshing did document", slog.Any("err", err.Error()))
			}
			lastRefresh = c.now()

		case <-c.ctx.Done():
			return
		}
	}
}

func parseCacheControl(s string) (maxAge time.Duration, ok bool) {
	for _, directive := range strings.Split(s, ",") {
		key, val, _ := strings.Cut(strings.TrimSpace(directive), "=")
		switch key {
		case "max-age":
			i, err := strconv.Atoi(val)
			if err != nil {
				continue
			}

			return time.Duration(i) * time.Second, true
		}
	}

	return 0, false
}
