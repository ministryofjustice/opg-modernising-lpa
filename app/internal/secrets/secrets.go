package secrets

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
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
	secret, err := c.getSecret("private-jwt-key-base64")
	if err != nil {
		return nil, err
	}

	keyBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 secret: %w", err)
	}

	return jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
}

func (c *Client) CookieSessionKeys() ([][]byte, error) {
	secret, err := c.getSecret("cookie-session-keys")
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

func (c *Client) getSecret(secretName string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := c.sm.GetSecretValue(input)
	if err != nil {
		return "", fmt.Errorf("error getting secret: %w", err)
	}

	if result.SecretString == nil {
		return "", fmt.Errorf("secret string was nil")
	}

	return *result.SecretString, nil
}
