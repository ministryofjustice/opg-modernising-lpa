package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/golang-jwt/jwt"
	"github.com/ministryofjustice/opg-go-common/env"
)

var (
	port        = env.Get("PORT", "8080")
	publicURL   = env.Get("PUBLIC_URL", "http://localhost:8080")
	internalURL = env.Get("INTERNAL_URL", "http://sign-in-mock:8080")
	clientId    = env.Get("CLIENT_ID", "theClientId")
	awsBaseUrl  = env.Get("AWS_BASE_URL", "http://localstack:4566")
)

type OpenIdConfig struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	Issuer                string `json:"issuer"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresInSeconds int    `json:"expires_in"`
	IDJWTToken       string `json:"id_token"`
}

type UserInfoResponse struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Phone         string `json:"phone"`
	PhoneVerified bool   `json:"phone_verified"`
	UpdatedAt     int    `json:"updated_at"`
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randomString(length int) string {
	return stringWithCharset(length, charset)
}

func createSignedToken(clientId, issuer string) (string, error) {
	t := jwt.New(jwt.GetSigningMethod("RS256"))

	t.Header["sub"] = fmt.Sprintf("%s-sub", randomString(10))
	t.Header["iss"] = issuer
	t.Header["nonce"] = "nonce-value"
	t.Header["aud"] = clientId
	t.Header["exp"] = time.Now().Add(time.Minute * 5).Unix()
	t.Header["iat"] = time.Now().Unix()

	key, err := getPrivateKey()
	if err != nil {
		return "", fmt.Errorf("unable to load key, %w", err)
	}

	return t.SignedString(key)
}

func createConfig(publicURL, internalURL string) OpenIdConfig {
	return OpenIdConfig{
		Issuer:                publicURL,
		AuthorizationEndpoint: publicURL + "/authorize",
		TokenEndpoint:         internalURL + "/token",
		UserinfoEndpoint:      internalURL + "/userinfo",
	}
}

func createUserInfo() UserInfoResponse {
	return UserInfoResponse{
		Sub:           "b2d2d115-1d7e-4579-b9d6-f8e84f4f56ca",
		Email:         "gideon.felix@example.org",
		EmailVerified: true,
		Phone:         "01406946277",
		PhoneVerified: true,
		UpdatedAt:     1311280970,
	}
}

func openIDConfig(c OpenIdConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("/.well-known/openid-configuration")

		payloadBuf := new(bytes.Buffer)
		err := json.NewEncoder(payloadBuf).Encode(c)

		if err != nil {
			log.Fatal(err)
		}

		log.Println(payloadBuf.String())

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(payloadBuf.Bytes())

		if err != nil {
			log.Fatalf("Issues parsing OIDC configuration response: %v", err)
		}
	}
}

func token(clientId, issuer string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("/token")

		t, err := createSignedToken(clientId, issuer)
		if err != nil {
			log.Fatalf("Error creating JWT: %s", err)
		}

		tr := TokenResponse{
			AccessToken:      "access-token-value",
			RefreshToken:     randomString(20),
			TokenType:        "Bearer",
			ExpiresInSeconds: 3600,
			IDJWTToken:       t,
		}

		payloadBuf := new(bytes.Buffer)
		err = json.NewEncoder(payloadBuf).Encode(tr)

		if err != nil {
			log.Fatal(err)
		}

		_, err = w.Write(payloadBuf.Bytes())

		if err != nil {
			log.Fatalf("Issues parsing token response: %v", err)
		}
	}
}

func authorize() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("/authorize")

		redirectUri := r.URL.Query().Get("redirect_uri")
		if redirectUri == "" {
			log.Fatal("Required query param 'redirect_uri' missing from request")
		}

		u, parseErr := url.Parse(redirectUri)
		if parseErr != nil {
			log.Fatalf("Error parsing redirect_uri: %s", parseErr)
		}

		q := u.Query()

		code := randomString(10)
		q.Set("code", code)

		state := r.URL.Query().Get("state")
		if state != "" {
			q.Set("state", state)
		}

		u.RawQuery = q.Encode()

		log.Printf("Redirecting to %s", u.String())

		http.Redirect(w, r, u.String(), 302)
	}
}

func userInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("/userinfo")

		ui := createUserInfo()

		payloadBuf := new(bytes.Buffer)
		err := json.NewEncoder(payloadBuf).Encode(ui)

		if err != nil {
			log.Fatal(err)
		}

		_, err = w.Write(payloadBuf.Bytes())

		if err != nil {
			log.Fatalf("Issues parsing user info response: %v", err)
		}
	}
}

func getPrivateKey() (*rsa.PrivateKey, error) {
	config := &aws.Config{
		Region:      aws.String("eu-west-1"),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
	}

	if len(awsBaseUrl) > 0 {
		config.Endpoint = aws.String(awsBaseUrl)
	}

	// Get private key from AWS secrets manager
	sess, err := session.NewSession(config)

	if err != nil {
		return &rsa.PrivateKey{}, fmt.Errorf("problem initialising new AWS session: %v", err)
	}

	svc := secretsmanager.New(
		sess,
		aws.NewConfig().WithRegion("eu-west-1"),
	)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String("private-jwt-key-base64"),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		return &rsa.PrivateKey{}, fmt.Errorf("problem initialising new AWS session: %v", err)
	}

	// Base64 Decode private key
	var base64PrivateKey string
	if result.SecretString != nil {
		base64PrivateKey = *result.SecretString
	}

	pem, err := base64.StdEncoding.DecodeString(base64PrivateKey)

	if err != nil {
		return &rsa.PrivateKey{}, fmt.Errorf("problem initialising new AWS session: %v", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pem)
	if err != nil {
		return &rsa.PrivateKey{}, fmt.Errorf("unable to parse RSA private key: %w", err)
	}
	return privateKey, nil
}

func main() {
	flag.Parse()

	c := createConfig(publicURL, internalURL)

	http.HandleFunc("/.well-known/openid-configuration", openIDConfig(c))
	http.HandleFunc("/authorize", authorize())
	http.HandleFunc("/token", token(clientId, c.Issuer))
	http.HandleFunc("/userinfo", userInfo())

	log.Println("GOV UK Sign in mock initialized")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		panic(err)
	}
}
