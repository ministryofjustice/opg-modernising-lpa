package secrets

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

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

	c := &Client{svc: secretsManager, cache: map[string]string{}}

	result, err := c.Secret(ctx, name)
	assert.Nil(t, err)
	assert.Equal(t, "a-fake-key", result)
}

func TestSecretWhenError(t *testing.T) {
	name := "a-test"

	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, name).
		Return("", expectedError)

	c := &Client{svc: secretsManager, cache: map[string]string{}}

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

	c := &Client{svc: secretsManager, cache: map[string]string{}}

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

	c := &Client{svc: secretsManager, cache: map[string]string{}}

	_, err := c.SecretBytes(ctx, name)
	assert.True(t, errors.Is(err, expectedError))
}

func TestSecretBytesWhenNotBase64(t *testing.T) {
	name := "a-test"

	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, name).
		Return("hello", nil)

	c := &Client{svc: secretsManager, cache: map[string]string{}}

	_, err := c.SecretBytes(ctx, name)
	assert.NotNil(t, err)
}

func TestCookieSessionKeys(t *testing.T) {
	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, "cookie-session-keys").
		Return(`["aGV5","YW5vdGhlcg=="]`, nil)

	c := &Client{svc: secretsManager, cache: map[string]string{}}

	result, err := c.CookieSessionKeys(ctx)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{[]byte("hey"), []byte("another")}, result)
}

func TestCookieSessionKeysWhenGetSecretError(t *testing.T) {
	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, "cookie-session-keys").
		Return("", expectedError)

	c := &Client{svc: secretsManager, cache: map[string]string{}}

	_, err := c.CookieSessionKeys(ctx)
	assert.True(t, errors.Is(err, expectedError))
}

func TestCookieSessionKeysWhenNotJSON(t *testing.T) {
	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, "cookie-session-keys").
		Return("oh", nil)

	c := &Client{svc: secretsManager, cache: map[string]string{}}

	_, err := c.CookieSessionKeys(ctx)
	assert.NotNil(t, err)
}

func TestCookieSessionKeysNotBase64(t *testing.T) {
	secretsManager := &mockSecretsManager{}
	secretsManager.
		On("GetSecretValue", ctx, "cookie-session-keys").
		Return(`["oh"]`, nil)

	c := &Client{svc: secretsManager, cache: map[string]string{}}

	_, err := c.CookieSessionKeys(ctx)
	assert.NotNil(t, err)
}
