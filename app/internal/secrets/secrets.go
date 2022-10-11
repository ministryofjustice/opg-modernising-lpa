package secrets

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
	"github.com/golang-jwt/jwt"
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

func (c *Client) PrivateKey() (*rsa.PrivateKey, error) {
	secret, err := c.cache.GetSecretString("private-jwt-key-base64")
	if err != nil {
		return nil, err
	}

	keyBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 secret: %w", err)
	}

	return jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
}

func (c *Client) CookieSessionKeys() ([][]byte, error) {
	secret, err := c.cache.GetSecretString("cookie-session-keys")
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

func (c *Client) PayApiKey() (string, error) {
	secret, err := c.cache.GetSecretString("gov-uk-pay-api-key")
	if err != nil {
		return "", err
	}

	return secret, nil
}

func (c *Client) YotiPrivateKey() ([]byte, error) {
	secret, err := c.cache.GetSecretString("yoti-private-key")
	if err != nil {
		return nil, fmt.Errorf("get yoti sandbox key: %w", err)
	}

	keyBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 yoti key: %w", err)
	}

	return keyBytes, nil
}

func (c *Client) OrdnanceSurveyApiKey() (string, error) {
	secret, err := c.cache.GetSecretString("os-postcode-lookup-api-key")
	if err != nil {
		return "", err
	}

	return secret, nil
}
