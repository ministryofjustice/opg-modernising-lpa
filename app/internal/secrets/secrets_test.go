package secrets

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
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

func TestPrivateKey(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	b, _ := x509.MarshalPKCS8PrivateKey(priv)
	k := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})

	b64PrivatePem := base64.StdEncoding.EncodeToString(k)

	secretsCache := &mockSecretsCache{}
	secretsCache.
		On("GetSecretString", "private-jwt-key-base64").
		Return(b64PrivatePem, nil)

	c := &Client{cache: secretsCache}

	result, err := c.PrivateKey()
	assert.Nil(t, err)
	assert.Equal(t, priv, result)
}

func TestPrivateKeyWhenGetSecretError(t *testing.T) {
	secretsCache := &mockSecretsCache{}
	secretsCache.
		On("GetSecretString", "private-jwt-key-base64").
		Return("", expectedError)

	c := &Client{cache: secretsCache}

	_, err := c.PrivateKey()
	assert.Equal(t, expectedError, err)
}

func TestPrivateKeyWhenNotBase64(t *testing.T) {
	secretsCache := &mockSecretsCache{}
	secretsCache.
		On("GetSecretString", "private-jwt-key-base64").
		Return("hello", nil)

	c := &Client{cache: secretsCache}

	_, err := c.PrivateKey()
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
	assert.Equal(t, expectedError, err)
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
