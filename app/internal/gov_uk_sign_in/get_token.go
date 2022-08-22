package govuksignin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/golang-jwt/jwt"
)

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

func (c *Client) GetToken() (*jwt.Token, error) {
	log.Println("GetToken()")

	// Build body for POST to OIDC /token
	body := &TokenRequestBody{
		GrantType:           "authorization_code",
		AuthorizationCode:   "code-value",
		RedirectUri:         "http://localhost:5050/home",
		ClientAssertionType: "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		// TODO - generate a real JWT https://docs.sign-in.service.gov.uk/integrate-with-integration-environment/integrate-with-code-flow/#create-a-jwt-assertion
		ClientAssertion: "THEJWT",
	}

	encodedPostBody := new(bytes.Buffer)
	err := json.NewEncoder(encodedPostBody).Encode(body)

	if err != nil {
		return &jwt.Token{}, err
	}

	// Build request for POST OIDC /token
	req, err := c.NewRequest("POST", c.DiscoverData.TokenEndpoint.Path, encodedPostBody)
	if err != nil {
		return &jwt.Token{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// POST to OIDC /token
	res, err := c.httpClient.Do(req)

	if err != nil {
		return &jwt.Token{}, err
	}

	pubKeyBytes, err := getPublicKey()
	if err != nil {
		return &jwt.Token{}, err
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)

	if err != nil {
		panic("failed to parse public key: " + err.Error())
	}

	// Parse response from OIDC /token
	defer res.Body.Close()

	var tokenResponse TokenResponseBody

	err = json.NewDecoder(res.Body).Decode(&tokenResponse)
	if err != nil {
		return &jwt.Token{}, err
	}

	// Parse JWT from OIDC /token
	log.Println(tokenResponse.IdToken)

	token, err := jwt.Parse(tokenResponse.IdToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		log.Println("Checking key")
		log.Println(pubKey)

		// TODO - add in any extra checks on JWT here
		return pubKey, nil
	})

	log.Println(err)

	return token, err
}

func getPublicKey() ([]byte, error) {
	// TODO move AWS code into aws package
	// Get public key from AWS secrets manager
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("eu-west-1"),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
		Endpoint:    aws.String("http://localstack:4566"),
	})

	if err != nil {
		return []byte{}, fmt.Errorf("problem initialising new AWS session: %v", err)
	}

	svc := secretsmanager.New(
		sess,
		aws.NewConfig().WithRegion("eu-west-1"),
	)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String("public-jwt-key-base64"),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		return []byte{}, fmt.Errorf("problem initialising new AWS session: %v", err)
	}

	// Base64 Decode public key
	var base64PublicKey string
	if result.SecretString != nil {
		base64PublicKey = *result.SecretString
	}

	publicKey, err := base64.StdEncoding.DecodeString(base64PublicKey)

	if err != nil {
		return []byte{}, fmt.Errorf("problem initialising new AWS session: %v", err)
	}

	return publicKey, nil
}
