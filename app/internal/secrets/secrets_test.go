package secrets

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/stretchr/testify/mock"
)

type mockSecretsManagerClient struct {
	mock.Mock
	secretsmanageriface.SecretsManagerAPI
}

func (m *mockSecretsManagerClient) GetSecretValue(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*secretsmanager.GetSecretValueOutput), args.Error(1)
}

func TestPrivateKey(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	b, _ := x509.MarshalPKCS8PrivateKey(priv)
	k := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})

	b64PrivatePem := base64.StdEncoding.EncodeToString(k)

	secretsClient := &mockSecretsManagerClient{}
	secretsClient.
		On("GetSecretValue", &secretsmanager.GetSecretValueInput{SecretId: aws.String("private-jwt-key-base64")}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: aws.String(b64PrivatePem)}, nil)

	c := Client{sm: secretsClient}

	result, err := c.PrivateKey()
	assert.Nil(t, err)
	assert.Equal(t, priv, result)
}

func TestPrivateKeyWhenGetSecretError(t *testing.T) {
	expectedError := errors.New("err")

	secretsClient := &mockSecretsManagerClient{}
	secretsClient.
		On("GetSecretValue", &secretsmanager.GetSecretValueInput{SecretId: aws.String("private-jwt-key-base64")}).
		Return(&secretsmanager.GetSecretValueOutput{}, expectedError)

	c := Client{sm: secretsClient}

	_, err := c.PrivateKey()
	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestPrivateKeyWhenMissing(t *testing.T) {
	secretsClient := &mockSecretsManagerClient{}
	secretsClient.
		On("GetSecretValue", &secretsmanager.GetSecretValueInput{SecretId: aws.String("private-jwt-key-base64")}).
		Return(&secretsmanager.GetSecretValueOutput{}, nil)

	c := Client{sm: secretsClient}

	_, err := c.PrivateKey()
	assert.NotNil(t, err)
}

func TestPrivateKeyWhenNotBase64(t *testing.T) {
	secretsClient := &mockSecretsManagerClient{}
	secretsClient.
		On("GetSecretValue", &secretsmanager.GetSecretValueInput{SecretId: aws.String("private-jwt-key-base64")}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: aws.String("hello")}, nil)

	c := Client{sm: secretsClient}

	_, err := c.PrivateKey()
	assert.NotNil(t, err)
}
