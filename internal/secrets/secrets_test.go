package secrets

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
)

var expectedError = errors.New("err")
var ctx = context.TODO()

func TestSecret(t *testing.T) {
	name := "a-test"
	secret := "a-fake-key"

	secretsManager := newMockSecretsManager(t)
	secretsManager.EXPECT().
		GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &name}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: &secret}, nil)

	cache := map[string]*cacheItem{}

	c := &Client{svc: secretsManager, ttl: time.Minute.Nanoseconds(), cache: cache}

	result, err := c.Secret(ctx, name)
	assert.Nil(t, err)
	assert.Equal(t, "a-fake-key", result)

	item, ok := cache[name]
	assert.True(t, ok)
	assert.Equal(t, "a-fake-key", item.value)
	assert.Equal(t, int64(0), item.errorCount)
	assert.InDelta(t, time.Now().UnixNano()+time.Minute.Nanoseconds(), item.untilNano, float64(time.Millisecond.Nanoseconds()))
}

func TestSecretWhenCached(t *testing.T) {
	name := "a-test"

	c := &Client{cache: map[string]*cacheItem{
		name: {
			value:     "a-fake-key",
			untilNano: time.Now().UnixNano() + time.Millisecond.Nanoseconds(),
		},
	}}

	result, err := c.Secret(ctx, name)
	assert.Nil(t, err)
	assert.Equal(t, "a-fake-key", result)
}

func TestSecretWhenCachedNotFresh(t *testing.T) {
	name := "a-test"
	secret := "a-fake-key"

	secretsManager := newMockSecretsManager(t)
	secretsManager.EXPECT().
		GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &name}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: &secret}, nil)

	c := &Client{
		svc: secretsManager,
		cache: map[string]*cacheItem{
			name: {
				value:     "an-old-value",
				untilNano: 5000,
			},
		},
	}

	result, err := c.Secret(ctx, name)
	assert.Nil(t, err)
	assert.Equal(t, "a-fake-key", result)
}

func TestSecretWhenCachedNotFreshButServiceErrors(t *testing.T) {
	name := "a-test"

	secretsManager := newMockSecretsManager(t)
	secretsManager.EXPECT().
		GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &name}).
		Return(nil, expectedError)

	item := &cacheItem{
		value:     "an-old-value",
		untilNano: 5000,
	}

	c := &Client{
		svc: secretsManager,
		cache: map[string]*cacheItem{
			name: item,
		},
	}

	result, err := c.Secret(ctx, name)
	assert.Nil(t, err)
	assert.Equal(t, "an-old-value", result)
	assert.Equal(t, int64(1), item.errorCount)
	assert.Equal(t, int64(1000005000), item.untilNano)

	result, err = c.Secret(ctx, name)
	assert.Nil(t, err)
	assert.Equal(t, "an-old-value", result)
	assert.Equal(t, int64(2), item.errorCount)
	assert.Equal(t, int64(3000005000), item.untilNano)
}

func TestSecretWhenCachedAndErroredSuccessResets(t *testing.T) {
	name := "a-test"
	secret := "a-fake-key"

	secretsManager := newMockSecretsManager(t)
	secretsManager.EXPECT().
		GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &name}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: &secret}, nil)

	item := &cacheItem{
		value:      "an-old-value",
		untilNano:  5000,
		errorCount: 5,
	}

	c := &Client{
		svc: secretsManager,
		ttl: time.Minute.Nanoseconds(),
		cache: map[string]*cacheItem{
			name: item,
		},
	}

	result, err := c.Secret(ctx, name)
	assert.Nil(t, err)
	assert.Equal(t, "a-fake-key", result)
	assert.Equal(t, int64(0), item.errorCount)
	assert.InDelta(t, time.Now().UnixNano()+time.Minute.Nanoseconds(), item.untilNano, float64(time.Millisecond.Nanoseconds()))
}

func TestSecretWhenError(t *testing.T) {
	name := "a-test"

	secretsManager := newMockSecretsManager(t)
	secretsManager.EXPECT().
		GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &name}).
		Return(nil, expectedError)

	c := &Client{svc: secretsManager, cache: map[string]*cacheItem{}}

	result, err := c.Secret(ctx, name)
	assert.Equal(t, "", result)
	assert.True(t, errors.Is(err, expectedError))
}

func TestSecretBytes(t *testing.T) {
	name := "a-test"
	key := []byte("hello")
	secret := base64.StdEncoding.EncodeToString(key)

	secretsManager := newMockSecretsManager(t)
	secretsManager.EXPECT().
		GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &name}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: &secret}, nil)

	c := &Client{svc: secretsManager, cache: map[string]*cacheItem{}}

	result, err := c.SecretBytes(ctx, name)
	assert.Nil(t, err)
	assert.Equal(t, key, result)
}

func TestSecretBytesWhenError(t *testing.T) {
	testcases := map[string]struct {
		output *secretsmanager.GetSecretValueOutput
		err    error
	}{
		"errors": {
			err: expectedError,
		},
		"not base64": {
			output: &secretsmanager.GetSecretValueOutput{SecretString: aws.String("hello")},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			name := "a-test"

			secretsManager := newMockSecretsManager(t)
			secretsManager.EXPECT().
				GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &name}).
				Return(tc.output, tc.err)

			c := &Client{svc: secretsManager, cache: map[string]*cacheItem{}}

			_, err := c.SecretBytes(ctx, name)
			assert.NotNil(t, err)
		})
	}
}

func TestCookieSessionKeys(t *testing.T) {
	name := "cookie-session-keys"
	secret := `["aGV5","YW5vdGhlcg=="]`

	secretsManager := newMockSecretsManager(t)
	secretsManager.EXPECT().
		GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &name}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: &secret}, nil)

	c := &Client{svc: secretsManager, cache: map[string]*cacheItem{}}

	result, err := c.CookieSessionKeys(ctx)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{[]byte("hey"), []byte("another")}, result)
}

func TestCookieSessionKeysWhenGetSecretError(t *testing.T) {
	testcases := map[string]struct {
		output *secretsmanager.GetSecretValueOutput
		err    error
	}{
		"errors": {
			err: expectedError,
		},
		"not base64": {
			output: &secretsmanager.GetSecretValueOutput{SecretString: aws.String(`["hello"]`)},
		},
		"not json": {
			output: &secretsmanager.GetSecretValueOutput{SecretString: aws.String("hello")},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			name := "cookie-session-keys"

			secretsManager := newMockSecretsManager(t)
			secretsManager.EXPECT().
				GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &name}).
				Return(tc.output, tc.err)

			c := &Client{svc: secretsManager, cache: map[string]*cacheItem{}}

			_, err := c.CookieSessionKeys(ctx)
			assert.NotNil(t, err)
		})
	}
}
