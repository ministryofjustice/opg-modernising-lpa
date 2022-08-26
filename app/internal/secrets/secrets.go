package secrets

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/golang-jwt/jwt"
)

type Client struct {
	sm secretsmanageriface.SecretsManagerAPI
}

func NewClient(baseURL string) (*Client, error) {
	config := &aws.Config{}
	if len(baseURL) > 0 {
		config.Endpoint = aws.String(baseURL)
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("error initialising new AWS session: %w", err)
	}

	return &Client{
		sm: secretsmanager.New(sess),
	}, nil
}

func (c *Client) PrivateKey() (*rsa.PrivateKey, error) {
	secretOutput, err := c.getSecret("private-jwt-key-base64")
	if err != nil {
		return nil, err
	}

	if secretOutput.SecretString == nil {
		return nil, fmt.Errorf("secret string was nil")
	}

	keyBytes, err := base64.StdEncoding.DecodeString(*secretOutput.SecretString)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 secret: %w", err)
	}

	return jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
}

func (c *Client) getSecret(secretName string) (*secretsmanager.GetSecretValueOutput, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := c.sm.GetSecretValue(input)
	if err != nil {
		return nil, fmt.Errorf("error getting secret: %w", err)
	}

	return result, nil
}
