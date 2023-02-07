package secrets

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")
var ctx = context.TODO()

type mockSecretsManager struct {
	mock.Mock
}

func (m *mockSecretsManager) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	args := m.Called(ctx, *params.SecretId)
	x := args.String(0)
	return &secretsmanager.GetSecretValueOutput{SecretString: &x}, args.Error(1)
}

func TestSecret(t *testing.T) {
	name := "a-test"

	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, name).
		Return("a-fake-key", nil)

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

	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, name).
		Return("a-fake-key", nil)

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

	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, name).
		Return("", expectedError)

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

	for i := 0; i < 8; i++ {
		result, err = c.Secret(ctx, name)
	}
	assert.Nil(t, err)
	assert.Equal(t, "an-old-value", result)
	assert.Equal(t, int64(10), item.errorCount)
	assert.Equal(t, int64(55000005000), item.untilNano)

	result, err = c.Secret(ctx, name)
	assert.Nil(t, err)
	assert.Equal(t, "an-old-value", result)
	assert.Equal(t, int64(10), item.errorCount)
	assert.Equal(t, int64(65000005000), item.untilNano)
}

func TestSecretWhenCachedAndErroredSuccessResets(t *testing.T) {
	name := "a-test"

	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, name).
		Return("a-fake-key", nil)

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

	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, name).
		Return("", expectedError)

	c := &Client{svc: secretsManager, cache: map[string]*cacheItem{}}

	result, err := c.Secret(ctx, name)
	assert.Equal(t, "", result)
	assert.True(t, errors.Is(err, expectedError))
}

func TestSecretBytes(t *testing.T) {
	name := "a-test"
	key := []byte("hello")

	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, name).
		Return(base64.StdEncoding.EncodeToString(key), nil)

	c := &Client{svc: secretsManager, cache: map[string]*cacheItem{}}

	result, err := c.SecretBytes(ctx, name)
	assert.Nil(t, err)
	assert.Equal(t, key, result)
}

func TestSecretBytesWhenGetSecretError(t *testing.T) {
	name := "a-test"
	key := []byte("hello")

	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, name).
		Return(base64.StdEncoding.EncodeToString(key), expectedError)

	c := &Client{svc: secretsManager, cache: map[string]*cacheItem{}}

	_, err := c.SecretBytes(ctx, name)
	assert.True(t, errors.Is(err, expectedError))
}

func TestSecretBytesWhenNotBase64(t *testing.T) {
	name := "a-test"

	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, name).
		Return("hello", nil)

	c := &Client{svc: secretsManager, cache: map[string]*cacheItem{}}

	_, err := c.SecretBytes(ctx, name)
	assert.NotNil(t, err)
}

func TestCookieSessionKeys(t *testing.T) {
	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, "cookie-session-keys").
		Return(`["aGV5","YW5vdGhlcg=="]`, nil)

	c := &Client{svc: secretsManager, cache: map[string]*cacheItem{}}

	result, err := c.CookieSessionKeys(ctx)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{[]byte("hey"), []byte("another")}, result)
}

func TestCookieSessionKeysWhenGetSecretError(t *testing.T) {
	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, "cookie-session-keys").
		Return("", expectedError)

	c := &Client{svc: secretsManager, cache: map[string]*cacheItem{}}

	_, err := c.CookieSessionKeys(ctx)
	assert.True(t, errors.Is(err, expectedError))
}

func TestCookieSessionKeysWhenNotJSON(t *testing.T) {
	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, "cookie-session-keys").
		Return("oh", nil)

	c := &Client{svc: secretsManager, cache: map[string]*cacheItem{}}

	_, err := c.CookieSessionKeys(ctx)
	assert.NotNil(t, err)
}

func TestCookieSessionKeysNotBase64(t *testing.T) {
	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, "cookie-session-keys").
		Return(`["oh"]`, nil)

	c := &Client{svc: secretsManager, cache: map[string]*cacheItem{}}

	_, err := c.CookieSessionKeys(ctx)
	assert.NotNil(t, err)
}
