package govuksignin

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ministryofjustice/opg-go-common/env"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/golang-jwt/jwt"
)

var b64 = base64.URLEncoding.WithPadding(base64.NoPadding)

type TokenRequestBody struct {
	GrantType           string `json:"grant_type"`
	AuthorizationCode   string `json:"code"`
	RedirectUri         string `json:"redirect_uri"`
	ClientAssertionType string `json:"client_assertion_type"`
	ClientAssertion     string `json:"client_assertion"`
}

type TokenResponseBody struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	IdToken      string `json:"id_token"`
}

func (c *Client) GetToken(redirectUri, clientID, JTI string) (*jwt.Token, error) {
	log.Println("GetToken()")

	data, _ := json.Marshal(map[string]interface{}{
		"aud": []string{"https://oidc.integration.account.gov.uk/token"},
		"iss": clientID,
		"sub": clientID,
		"exp": time.Now().Add(5 * time.Minute).Unix(),
		"jti": JTI,
		"iat": time.Now().Unix(),
	})

	privateKey, err := getPrivateKey()
	if err != nil {
		return nil, err
	}

	signedAssertion := signJwt(string(data), privateKey)

	body := &TokenRequestBody{
		GrantType:           "authorization_code",
		AuthorizationCode:   "code-value",
		RedirectUri:         redirectUri,
		ClientAssertionType: "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		ClientAssertion:     signedAssertion,
	}

	encodedPostBody := new(bytes.Buffer)
	err = json.NewEncoder(encodedPostBody).Encode(body)

	if err != nil {
		return nil, err
	}

	req, err := c.NewRequest("POST", c.DiscoverData.TokenEndpoint.Path, encodedPostBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	pubKey, err := getPublicKey()
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var tokenResponse TokenResponseBody

	err = json.NewDecoder(res.Body).Decode(&tokenResponse)
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenResponse.IdToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// TODO - add in any extra checks on JWT here
		return pubKey, nil
	})

	return token, err
}

func signJwt(data string, privateKey *rsa.PrivateKey) string {
	header := `{"alg":"RS256"}`

	toSign := b64.EncodeToString([]byte(header)) + "." + b64.EncodeToString([]byte(data))

	digest := sha256.Sum256([]byte(toSign))
	sig, err := privateKey.Sign(rand.Reader, digest[:], crypto.SHA256)
	if err != nil {
		panic(err)
	}

	return toSign + "." + b64.EncodeToString(sig)
}

func getPublicKey() (*rsa.PublicKey, error) {
	secretOutput, err := getSecret("public-jwt-key-base64")
	if err != nil {
		return nil, err
	}

	keyBytes, err := decodeSecret(secretOutput)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM(keyBytes)

}

func getPrivateKey() (*rsa.PrivateKey, error) {
	secretOutput, err := getSecret("private-jwt-key-base64")
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

func getSecret(secretName string) (*secretsmanager.GetSecretValueOutput, error) {
	awsBaseUrl := env.Get("AWS_BASE_URL", "http://localstack:4566")

	config := &aws.Config{
		Region: aws.String("eu-west-1"),
	}

	if len(awsBaseUrl) > 0 {
		config.Endpoint = aws.String(awsBaseUrl)
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
