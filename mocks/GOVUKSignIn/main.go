package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/golang-jwt/jwt"
	"github.com/ministryofjustice/opg-go-common/env"
)

type OpenIdConfig struct {
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	RegistrationEndpoint                       string   `json:"registration_endpoint"`
	Issuer                                     string   `json:"issuer"`
	JwksUri                                    string   `json:"jwks_uri"`
	ScopesSupported                            []string `json:"scopes_supported"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	GrantTypesSupported                        []string `json:"grant_types_supported"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported"`
	ServiceDocumentation                       string   `json:"service_documentation"`
	RequestUriParameterSupported               bool     `json:"request_uri_parameter_supported"`
	Trustmarks                                 string   `json:"trustmarks"`
	SubjectTypesSupported                      []string `json:"subject_types_supported"`
	UserinfoEndpoint                           string   `json:"userinfo_endpoint"`
	EndSessionEndpoint                         string   `json:"end_session_endpoint"`
	IdTokenSigningAlgValuesSupported           []string `json:"id_token_signing_alg_values_supported"`
	ClaimTypesSupported                        []string `json:"claim_types_supported"`
	ClaimsSupported                            []string `json:"claims_supported"`
	BackchannelLogoutSupported                 bool     `json:"backchannel_logout_supported"`
	BackchannelLogoutSessionSupported          bool     `json:"backchannel_logout_session_supported"`
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

var privateKeyCache = &sync.Map{}

const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandomString(length int) string {
	return StringWithCharset(length, charset)
}

func loadPrivateKey(pemPath string) (*rsa.PrivateKey, error) {
	if key, ok := privateKeyCache.Load(pemPath); ok {
		return key.(*rsa.PrivateKey), nil
	}

	pem, err := ioutil.ReadFile(pemPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load private key pem file from %s, %w", pemPath, err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pem)
	if err != nil {
		return nil, fmt.Errorf("unable to parse RSA private key from data in %s, %w", pemPath, err)
	}

	privateKeyCache.Store(pemPath, privateKey)

	return privateKey, nil
}

func createSignedToken(clientId, issuer string) (string, error) {
	t := jwt.New(jwt.GetSigningMethod("RS256"))

	t.Header["sub"] = fmt.Sprintf("%s-sub", RandomString(10))
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

func createConfig(baseUri string) OpenIdConfig {
	return OpenIdConfig{
		AuthorizationEndpoint:             fmt.Sprintf("%s/authorize", baseUri),
		TokenEndpoint:                     fmt.Sprintf("%s/token", baseUri),
		RegistrationEndpoint:              fmt.Sprintf("%s/connect/register", baseUri),
		Issuer:                            fmt.Sprintf("%s", baseUri),
		JwksUri:                           fmt.Sprintf("%s/.well-known/jwks.json", baseUri),
		ScopesSupported:                   []string{"openid", "email", "phone", "offline_access"},
		ResponseTypesSupported:            []string{"code"},
		GrantTypesSupported:               []string{"authorization_code"},
		TokenEndpointAuthMethodsSupported: []string{"private_key_jwt"},
		TokenEndpointAuthSigningAlgValuesSupported: []string{"RS256", "RS384", "RS512", "PS256", "PS384", "PS512"},
		ServiceDocumentation:                       "https://docs.sign-in.service.gov.uk/",
		RequestUriParameterSupported:               true,
		Trustmarks:                                 fmt.Sprintf("%s/trustmark", baseUri),
		SubjectTypesSupported:                      []string{"public", "pairwise"},
		UserinfoEndpoint:                           fmt.Sprintf("%s/userinfo", baseUri),
		EndSessionEndpoint:                         fmt.Sprintf("%s/logout", baseUri),
		IdTokenSigningAlgValuesSupported:           []string{"ES256"},
		ClaimTypesSupported:                        []string{"normal"},
		ClaimsSupported:                            []string{"sub", "email", "email_verified", "phone_number", "phone_number_verified"},
		BackchannelLogoutSupported:                 true,
		BackchannelLogoutSessionSupported:          true,
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
			RefreshToken:     RandomString(20),
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

		code := RandomString(10)
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

		log.Println(payloadBuf.String())

		_, err = w.Write(payloadBuf.Bytes())

		if err != nil {
			log.Fatalf("Issues parsing user info response: %v", err)
		}
	}
}

func getPrivateKey() (*rsa.PrivateKey, error) {
	// TODO move AWS code into aws package
	// Get private key from AWS secrets manager
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("eu-west-1"),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
		Endpoint:    aws.String("http://localstack:4566"),
	})

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
	var (
		port        = flag.String("port", env.Get("PROXY_PORT", "5060"), "The port to run the mock on")
		clientId    = flag.String("clientid", env.Get("CLIENT_ID", "theClientId"), "The client ID set up when registering with Gov UK Sign in")
		mockBaseUri = flag.String("mockbaseuri", env.Get("MOCK_BASE_URI", fmt.Sprintf("http://sign-in-mock:%s", *port)), "The client ID set up when registering with Gov UK Sign in")
	)
	log.Println("Initializing GOV UK Sign in mock")

	flag.Parse()

	c := createConfig(*mockBaseUri)

	http.HandleFunc("/.well-known/openid-configuration", openIDConfig(c))
	http.HandleFunc("/authorize", authorize())
	http.HandleFunc("/token", token(*clientId, c.Issuer))
	http.HandleFunc("/userinfo", userInfo())

	log.Println("GOV UK Sign in mock initialized")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil); err != nil {
		panic(err)
	}
}
