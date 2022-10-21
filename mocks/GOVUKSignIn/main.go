package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/ministryofjustice/opg-go-common/env"
)

var (
	port        = env.Get("PORT", "8080")
	publicURL   = env.Get("PUBLIC_URL", "http://localhost:8080")
	internalURL = env.Get("INTERNAL_URL", "http://sign-in-mock:8080")
	clientId    = env.Get("CLIENT_ID", "theClientId")
	awsBaseUrl  = env.Get("AWS_BASE_URL", "http://localstack:4566")

	nonce         string
	signingKid    = "my-kid"
	signingKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
)

type OpenIdConfig struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	Issuer                string `json:"issuer"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	IDToken     string `json:"id_token"`
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

func stringWithCharset(length int, charset string) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes)
}

func randomString(length int) string {
	return stringWithCharset(length, charset)
}

func createSignedToken(clientId, issuer string) (string, error) {
	t := jwt.New(jwt.SigningMethodES256)

	t.Header["kid"] = signingKid

	t.Claims = jwt.MapClaims{
		"sub":   fmt.Sprintf("%s-sub", randomString(10)),
		"iss":   issuer,
		"nonce": nonce,
		"aud":   clientId,
		"exp":   time.Now().Add(time.Minute * 5).Unix(),
		"iat":   time.Now().Unix(),
	}

	return t.SignedString(signingKey)
}

func openIDConfig(c OpenIdConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(c)
	}
}

func jwks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		publicKey := signingKey.PublicKey

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []map[string]interface{}{
				{
					"kty": "EC",
					"use": "sig",
					"crv": "P-256",
					"kid": signingKid,
					"x":   base64.URLEncoding.EncodeToString(publicKey.X.Bytes()),
					"y":   base64.URLEncoding.EncodeToString(publicKey.Y.Bytes()),
					"alg": "ES256",
				},
			},
		})
	}
}

func token(clientId, issuer string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := createSignedToken(clientId, issuer)
		if err != nil {
			log.Fatalf("Error creating JWT: %s", err)
		}

		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: "access-token-value",
			TokenType:   "Bearer",
			IDToken:     t,
		})
	}
}

func authorize() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("/authorize")

		nonce = r.FormValue("nonce")

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
		json.NewEncoder(w).Encode(UserInfoResponse{
			Sub:           randomString(12),
			Email:         "simulate-delivered@notifications.service.gov.uk",
			EmailVerified: true,
			Phone:         "01406946277",
			PhoneVerified: true,
			UpdatedAt:     1311280970,
		})
	}
}

func main() {
	flag.Parse()

	c := OpenIdConfig{
		Issuer:                publicURL,
		AuthorizationEndpoint: publicURL + "/authorize",
		TokenEndpoint:         internalURL + "/token",
		UserinfoEndpoint:      internalURL + "/userinfo",
		JwksURI:               internalURL + "/.well-known/jwks",
	}

	http.HandleFunc("/.well-known/openid-configuration", openIDConfig(c))
	http.HandleFunc("/.well-known/jwks", jwks())
	http.HandleFunc("/authorize", authorize())
	http.HandleFunc("/token", token(clientId, c.Issuer))
	http.HandleFunc("/userinfo", userInfo())

	log.Println("GOV UK Sign in mock initialized")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), logRoute(http.DefaultServeMux)); err != nil {
		panic(err)
	}
}

func logRoute(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	}
}
