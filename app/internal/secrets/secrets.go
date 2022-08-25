package secrets

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/golang-jwt/jwt"
)

type Client struct {
	sm *secretsmanager.SecretsManager
}

func (c *Client) PrivateKey() (*rsa.PrivateKey, error) {
	secretOutput, err := c.getSecret("private-jwt-key-base64")
	if err != nil {
		return nil, err
	}

	keyBytes, err := decodeSecret(secretOutput)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
}

func decodeSecret(secretOutput *secretsmanager.GetSecretValueOutput) ([]byte, error) {
	var base64PublicKey string
	if secretOutput.SecretString != nil {
		base64PublicKey = *secretOutput.SecretString
	}

	publicKey, err := base64.StdEncoding.DecodeString(base64PublicKey)

	if err != nil {
		return nil, fmt.Errorf("error decoding base64 secret: %v", err)
	}

	return publicKey, nil
}

func (c *Client) getSecret(secretName string) (*secretsmanager.GetSecretValueOutput, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := c.sm.GetSecretValue(input)
	if err != nil {
		return nil, fmt.Errorf("error getting secret: %v", err)
	}

	return result, nil
}

func NewClient(baseURL string) (Client, error) {
	config := &aws.Config{
		Region: aws.String("eu-west-1"),
	}

	if len(baseURL) > 0 {
		config.Endpoint = aws.String(baseURL)
	}

	sess, err := session.NewSession(config)

	if err != nil {
		return Client{}, fmt.Errorf("error initialising new AWS session: %v", err)
	}

	c := Client{sm: secretsmanager.New(
		sess,
		aws.NewConfig().WithRegion("eu-west-1"),
	)}

	return c, nil
}
