package secrets

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

const (
	GovUkNotify           = "gov-uk-notify-api-key"
	GovUkPay              = "gov-uk-pay-api-key"
	GovUkSignInPrivateKey = "private-jwt-key-base64"
	YotiPrivateKey        = "yoti-private-key"

	cookieSessionKeys = "cookie-session-keys"
)

type secretsCache interface {
	GetSecretString(secretId string) (string, error)
}

type Client struct {
	cache secretsCache
}

func NewClient(sess *session.Session) (*Client, error) {
	cache, err := secretcache.New(func(c *secretcache.Cache) { c.Client = secretsmanager.New(sess) })
	if err != nil {
		return nil, err
	}

	return &Client{cache: cache}, nil
}

func (c *Client) Secret(name string) (string, error) {
	secret, err := c.cache.GetSecretString(name)
	if err != nil {
		return "", fmt.Errorf("error retrieving secret '%s': %w", name, err)
	}

	return secret, nil
}

func (c *Client) SecretBytes(name string) ([]byte, error) {
	secret, err := c.Secret(name)
	if err != nil {
		return nil, err
	}

	keyBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 secret '%s': %w", name, err)
	}

	return keyBytes, nil
}

func (c *Client) CookieSessionKeys() ([][]byte, error) {
	secret, err := c.Secret(cookieSessionKeys)
	if err != nil {
		return nil, err
	}

	var v []string
	if err := json.Unmarshal([]byte(secret), &v); err != nil {
		return nil, err
	}

	keys := make([][]byte, len(v))
	for i, data := range v {
		enc, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return nil, err
		}
		keys[i] = []byte(enc)
	}

	return keys, nil
}
