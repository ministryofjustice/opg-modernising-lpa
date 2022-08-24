package secrets

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/golang-jwt/jwt"
)

type Client struct {
	BaseURL string
}

func (c *Client) PublicKey() (*rsa.PublicKey, error) {
	secretOutput, err := c.getSecret("public-jwt-key-base64")
	if err != nil {
		return nil, err
	}

	keyBytes, err := decodeSecret(secretOutput)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM(keyBytes)
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
	config := &aws.Config{
		Region: aws.String("eu-west-1"),
	}

	if len(c.BaseURL) > 0 {
		config.Endpoint = aws.String(c.BaseURL)
		config.Credentials = credentials.NewStaticCredentials("test", "test", "")
	}

	sess, err := session.NewSession(config)

	if err != nil {
		return nil, fmt.Errorf("error initialising new AWS session: %v", err)
	}

	svc := secretsmanager.New(
		sess,
		aws.NewConfig().WithRegion("eu-west-1"),
	)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		return nil, fmt.Errorf("error getting secret: %v", err)
	}

	return result, nil
}
