package secrets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const (
	GovUkNotify           = "gov-uk-notify-api-key"
	GovUkPay              = "gov-uk-pay-api-key"
	GovUkSignInPrivateKey = "private-jwt-key-base64"
	OrdnanceSurvey        = "os-postcode-lookup-api-key"
	YotiPrivateKey        = "yoti-private-key"

	cookieSessionKeys = "cookie-session-keys"
)

type secretsManager interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type Client struct {
	svc secretsManager

	mu    sync.Mutex
	cache map[string]string
}

func NewClient(cfg aws.Config) (*Client, error) {
	svc := secretsmanager.NewFromConfig(cfg)

	return &Client{svc: svc, cache: map[string]string{}}, nil
}

func (c *Client) Secret(ctx context.Context, name string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.cache[name]
	if ok {
		return value, nil
	}

	result, err := c.svc.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	})
	if err != nil {
		return "", fmt.Errorf("error retrieving secret '%s': %w", name, err)
	}

	c.cache[name] = *result.SecretString

	return *result.SecretString, nil
}

func (c *Client) SecretBytes(ctx context.Context, name string) ([]byte, error) {
	secret, err := c.Secret(ctx, name)
	if err != nil {
		return nil, err
	}

	keyBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 secret '%s': %w", name, err)
	}

	return keyBytes, nil
}

func (c *Client) CookieSessionKeys(ctx context.Context) ([][]byte, error) {
	secret, err := c.Secret(ctx, cookieSessionKeys)
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
