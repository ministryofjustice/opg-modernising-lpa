package secrets

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
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

func (m *mockSecretsManagerClient) GetSecretValue(i *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	args := m.Called(i)
	return args.Get(0).(*secretsmanager.GetSecretValueOutput), args.Error(1)
}

func TestPrivateKey(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	b, _ := x509.MarshalPKCS8PrivateKey(priv)
	k := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})

	b64PrivatePem := base64.StdEncoding.EncodeToString(k)

	secretsClient := mockSecretsManagerClient{}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String("private-jwt-key-base64"),
	}

	output := &secretsmanager.GetSecretValueOutput{
		SecretString: &b64PrivatePem,
	}

	secretsClient.
		On("GetSecretValue", input).
		Return(output, nil)

	c := Client{sm: &secretsClient}

	got, err := c.PrivateKey()

	assert.Nil(t, err)
	assert.Equal(t, priv, got)
}
