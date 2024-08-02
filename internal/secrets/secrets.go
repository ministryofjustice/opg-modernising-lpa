// Package secrets provides a client for AWS Secrets Manager.
package secrets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const (
	GovUkNotify                    = "gov-uk-notify-api-key"
	GovUkPay                       = "gov-uk-pay-api-key"
	GovUkOneLoginPrivateKey        = "private-jwt-key-base64"
	GovUkOneLoginIdentityPublicKey = "gov-uk-onelogin-identity-public-key"
	OrdnanceSurvey                 = "os-postcode-lookup-api-key"
	LpaStoreJwtSecretKey           = "lpa-store-jwt-secret-key"

	cookieSessionKeys = "cookie-session-keys"

	delay = time.Second
)

type secretsManager interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type cacheItem struct {
	untilNano  int64
	errorCount int64
	value      string
}

func (i *cacheItem) isFresh() bool {
	return time.Now().UnixNano() < i.untilNano
}

type Client struct {
	svc secretsManager
	ttl int64

	mu    sync.Mutex
	cache map[string]*cacheItem
}

func NewClient(cfg aws.Config, ttl time.Duration) (*Client, error) {
	svc := secretsmanager.NewFromConfig(cfg)

	return &Client{svc: svc, ttl: ttl.Nanoseconds(), cache: map[string]*cacheItem{}}, nil
}

// Secret retrieves the named secret from the cache. If not in the cache or the
// item is stale then the secret is retrieved from Secrets Manager. On failure
// the stale secret will be returned, and if that isn't possible an error.
func (c *Client) Secret(ctx context.Context, name string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.cache[name]
	if found && item.isFresh() {
		item.errorCount = 0
		return item.value, nil
	}

	result, err := c.svc.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	})
	if err != nil {
		if found {
			item.errorCount++
			item.untilNano += item.errorCount * delay.Nanoseconds()

			return item.value, nil
		}

		return "", fmt.Errorf("error retrieving secret '%s': %w", name, err)
	}

	if found {
		item.errorCount = 0
		item.untilNano = time.Now().UnixNano() + c.ttl
		item.value = *result.SecretString
	} else {
		c.cache[name] = &cacheItem{untilNano: time.Now().UnixNano() + c.ttl, value: *result.SecretString}
	}

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
