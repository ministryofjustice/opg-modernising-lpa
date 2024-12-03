// Package lpastore provides a client for the LPA store.
package lpastore

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
)

const (
	issuer                     = "opg.poas.makeregister"
	statusActive               = "active"
	statusInactive             = "inactive"
	statusRemoved              = "removed"
	appointmentTypeOriginal    = "original"
	appointmentTypeReplacement = "replacement"
)

var ErrNotFound = errors.New("lpa not found in lpa-store")

type responseError struct {
	name string
	body any
}

func (e responseError) Error() string { return e.name }
func (e responseError) Title() string { return e.name }
func (e responseError) Data() any     { return e.body }

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type SecretsClient interface {
	Secret(ctx context.Context, name string) (string, error)
}

type Client struct {
	baseURL       string
	secretsClient SecretsClient
	secretARN     string
	doer          Doer
	now           func() time.Time
}

func New(baseURL string, secretsClient SecretsClient, secretARN string, lambdaClient Doer) *Client {
	return &Client{
		baseURL:       baseURL,
		secretsClient: secretsClient,
		secretARN:     secretARN,
		doer:          lambdaClient,
		now:           time.Now,
	}
}

func (c *Client) do(ctx context.Context, actorUID actoruid.UID, req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:   issuer,
		IssuedAt: jwt.NewNumericDate(c.now()),
		Subject:  actorUID.PrefixedString(),
	})

	secretKey, err := c.secretsClient.Secret(ctx, c.secretARN)
	if err != nil {
		return nil, err
	}

	auth, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Jwt-Authorization", "Bearer "+auth)

	return c.doer.Do(req)
}

func (c *Client) CheckHealth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health-check", nil)
	if err != nil {
		return err
	}

	resp, err := c.doer.Do(req)
	if err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return responseError{name: fmt.Sprintf("expected 200 response but got %d", resp.StatusCode)}
	}

	return nil
}
