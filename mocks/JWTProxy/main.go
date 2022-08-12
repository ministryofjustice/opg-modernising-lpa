package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/ministryofjustice/opg-go-common/env"
)

type OpenIdConfigResponse struct {
	Issuer string `json:"issuer"`
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresInSeconds int    `json:"expires_in"`
	IDJWTToken       string `json:"id_token"`
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

func loadPrivateKey(pemPath string) (*ecdsa.PrivateKey, error) {
	if key, ok := privateKeyCache.Load(pemPath); ok {
		return key.(*ecdsa.PrivateKey), nil
	}

	pem, err := ioutil.ReadFile(pemPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load private key pem file from %s, %w", pemPath, err)
	}

	privateKey, err := jwt.ParseECPrivateKeyFromPEM(pem)
	if err != nil {
		return nil, fmt.Errorf("unable to parse ec private key from data in %s, %w", pemPath, err)
	}

	privateKeyCache.Store(pemPath, privateKey)

	return privateKey, nil
}

func createToken(keyPath, clientId, issuer string) (string, error) {
	t := jwt.New(jwt.GetSigningMethod("RS256"))

	t.Header["sub"] = fmt.Sprintf("%s-sub", RandomString(10))
	t.Header["iss"] = issuer
	t.Header["nonce"] = "nonce-value"
	t.Header["aud"] = clientId
	t.Header["exp"] = time.Now().Add(time.Minute * 5).Unix()
	t.Header["iat"] = time.Now().Unix()

	key, err := loadPrivateKey(keyPath)
	if err != nil {
		return "", fmt.Errorf("unable to load key, %w", err)
	}

	return t.SignedString(key)
}

// Given a request send it to the appropriate url
func proxyRequest(privKeyPath, clientId, issuer string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := createToken(privKeyPath, clientId, issuer)
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
			log.Fatalf("Issues parsing OIDC configuration response: %v", err)
		}
	}
}

func main() {
	var (
		port           = flag.String("port", env.Get("PROXY_PORT", "5060"), "The port to run the proxy on")
		privateKeyPath = flag.String("privkey", env.Get("PROXY_PRIVATE_KEY", "key.pem"), "The path to an RSA256 private key file")
		clientId       = flag.String("clientid", env.Get("CLIENT_ID", "theClientId"), "The client ID set up when registering with Gov UK Sign in")
	)

	flag.Parse()

	r, err := http.Get("http://localhost:7011/.well-known/openid-configuration")

	if err != nil {
		log.Fatal(err)
	}

	defer r.Body.Close()

	var openIdResponse OpenIdConfigResponse

	err = json.NewDecoder(r.Body).Decode(&openIdResponse)
	if err != nil {
		log.Fatalf("Issues parsing OIDC configuration response: %v", err)
	}

	// start server
	http.HandleFunc("/", proxyRequest(*privateKeyPath, *clientId, *openIdResponse.Issuer))

	if err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil); err != nil {
		panic(err)
	}
}
