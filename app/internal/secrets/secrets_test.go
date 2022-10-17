package secrets

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

type mockSecretsCache struct {
	mock.Mock
}

func (m *mockSecretsCache) GetSecretString(name string) (string, error) {
	args := m.Called(name)
	return args.String(0), args.Error(1)
}

func TestSecret(t *testing.T) {
	name := "a-test"

	secretsCache := &mockSecretsCache{}
	secretsCache.
		On("GetSecretString", name).
		Return("a-fake-key", nil)

	c := &Client{cache: secretsCache}

	result, err := c.Secret(name)
	assert.Nil(t, err)
	assert.Equal(t, "a-fake-key", result)
}

func TestSecretWhenError(t *testing.T) {
	name := "a-test"

	secretsCache := &mockSecretsCache{}
	secretsCache.
		On("GetSecretString", name).
		Return("", expectedError)

	c := &Client{cache: secretsCache}

	result, err := c.Secret(name)
	assert.Equal(t, "", result)
	assert.True(t, errors.Is(err, expectedError))
}

func TestSecretBytes(t *testing.T) {
	name := "a-test"
	key := []byte("hello")

	secretsCache := &mockSecretsCache{}
	secretsCache.
		On("GetSecretString", name).
		Return(base64.StdEncoding.EncodeToString(key), nil)

	c := &Client{cache: secretsCache}

	result, err := c.SecretBytes(name)
	assert.Nil(t, err)
	assert.Equal(t, key, result)
}

func TestSecretBytesWhenGetSecretError(t *testing.T) {
	name := "a-test"
	key := []byte("hello")

	secretsCache := &mockSecretsCache{}
	secretsCache.
		On("GetSecretString", name).
		Return(base64.StdEncoding.EncodeToString(key), expectedError)

	c := &Client{cache: secretsCache}

	_, err := c.SecretBytes(name)
	assert.True(t, errors.Is(err, expectedError))
}

func TestSecretBytesWhenNotBase64(t *testing.T) {
	name := "a-test"

	secretsCache := &mockSecretsCache{}
	secretsCache.
		On("GetSecretString", name).
		Return("hello", nil)

	c := &Client{cache: secretsCache}

	_, err := c.SecretBytes(name)
	assert.NotNil(t, err)
}

func TestCookieSessionKeys(t *testing.T) {
	secretsCache := &mockSecretsCache{}
	secretsCache.
		On("GetSecretString", "cookie-session-keys").
		Return(`["aGV5","YW5vdGhlcg=="]`, nil)

	c := &Client{cache: secretsCache}

	result, err := c.CookieSessionKeys()
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{[]byte("hey"), []byte("another")}, result)
}

func TestCookieSessionKeysWhenGetSecretError(t *testing.T) {
	secretsCache := &mockSecretsCache{}
	secretsCache.
		On("GetSecretString", "cookie-session-keys").
		Return("", expectedError)

	c := &Client{cache: secretsCache}

	_, err := c.CookieSessionKeys()
	assert.True(t, errors.Is(err, expectedError))
}

func TestCookieSessionKeysWhenNotJSON(t *testing.T) {
	secretsCache := &mockSecretsCache{}
	secretsCache.
		On("GetSecretString", "cookie-session-keys").
		Return("oh", nil)

	c := &Client{cache: secretsCache}

	_, err := c.CookieSessionKeys()
	assert.NotNil(t, err)
}

func TestCookieSessionKeysNotBase64(t *testing.T) {
	secretsCache := &mockSecretsCache{}
	secretsCache.
		On("GetSecretString", "cookie-session-keys").
		Return(`["oh"]`, nil)

	c := &Client{cache: secretsCache}

	_, err := c.CookieSessionKeys()
	assert.NotNil(t, err)
}
