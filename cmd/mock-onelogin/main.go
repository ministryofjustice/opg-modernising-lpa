package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-go-common/env"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	port               = env.Get("PORT", "8080")
	publicURL          = env.Get("PUBLIC_URL", "http://localhost:8080")
	internalURL        = env.Get("INTERNAL_URL", "http://mock-onelogin:8080")
	clientId           = env.Get("CLIENT_ID", "theClientId")
	serviceRedirectUrl = env.Get("REDIRECT_URL", "http://localhost:5050/auth/redirect")

	tokenSigningKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tokenSigningKid    = randomString("kid-", 8)

	sessions      = map[string]sessionData{}
	tokens        = map[string]sessionData{}
	emailOverride = ""
)

type sessionData struct {
	user     string
	nonce    string
	identity bool
	sub      string
}

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
	Sub             string `json:"sub"`
	Email           string `json:"email"`
	EmailVerified   bool   `json:"email_verified"`
	Phone           string `json:"phone"`
	PhoneVerified   bool   `json:"phone_verified"`
	UpdatedAt       int    `json:"updated_at"`
	CoreIdentityJWT string `json:"https://vocab.account.gov.uk/v1/coreIdentityJWT,omitempty"`
}

func openIDConfig(c OpenIdConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(c)
	}
}

func jwks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		publicKey := tokenSigningKey.PublicKey

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []map[string]interface{}{
				{
					"kty": "EC",
					"use": "sig",
					"crv": "P-256",
					"kid": tokenSigningKid,
					"x":   base64.RawURLEncoding.EncodeToString(publicKey.X.Bytes()),
					"y":   base64.RawURLEncoding.EncodeToString(publicKey.Y.Bytes()),
					"alg": "ES256",
				},
			},
		})
	}
}

func token(clientId, issuer string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.PostFormValue("code")
		accessToken := randomString("token-", 10)

		session := sessions[code]
		delete(sessions, code)
		tokens[accessToken] = session

		t, err := createSignedToken(session.nonce, clientId, issuer)
		if err != nil {
			log.Fatalf("Error creating JWT: %s", err)
		}

		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			IDToken:     t,
		})
	}
}

func authorize() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wantsIdentity := r.FormValue("vtr") == "[Cl.Cm.P2]" && r.FormValue("claims") == `{"userinfo":{"https://vocab.account.gov.uk/v1/coreIdentityJWT": null}}`

		if r.Method == http.MethodGet && wantsIdentity {
			io.WriteString(w, `<!doctype html>
<style>body { font-family: sans-serif; font-size: 21px; margin: 2rem; } label { display: inline-block; padding: .5rem 0; margin-bottom: 1rem; } button { font-family: inherit; font-size: 21px; padding: .3rem .5rem; } input { transform: scale(1.2); margin-right: .5rem; }</style>
<h1>Mock GOV.UK One Login</h1>
<form method="post">
<label><input type="radio" name="user" value="donor" />Sam Smith (donor)</label><br/>
<label><input type="radio" name="user" value="certificate-provider" />Charlie Cooper (certificate provider)</label><br/>
<label><input type="radio" name="user" value="random" />Somebody Else (a random person)</label><br/>
<button type="submit">Sign in</button>
</form>`)
			return
		}

		if r.Method == http.MethodGet {
			io.WriteString(w, `<!doctype html>
<style>body { font-family: sans-serif; font-size: 21px; margin: 2rem; } label { padding: .5rem 0; display: block; } button { font-family: inherit; font-size: 21px; padding: .3rem .5rem; margin-top: 1rem; display: block; }</style>
<h1>Mock GOV.UK One Login</h1>
<form method="post">

<p>Sign in using a OneLogin sub (to ignore, leave empty)</p>
<label for="sub">OneLogin sub</label>
<input type="text" name="sub" id=f-sub />

<p>Set email in OneLogin UserInfo (leave empty to set as test email address)</p>
<label for="email">Email</label>
<input type="text" name="email" id=f-email />

<button type="submit">Sign in</button>
</form>`)
			return
		}

		redirectUri := r.FormValue("redirect_uri")
		if redirectUri == "" {
			log.Fatal("Required query param 'redirect_uri' missing from request")
		}

		if redirectUri != serviceRedirectUrl {
			log.Fatalf("redirect_uri does not match pre-defined redirect URL (in RL this is set with GDS at a service level). Got %s, want %s", redirectUri, serviceRedirectUrl)
		}

		u, parseErr := url.Parse(redirectUri)
		if parseErr != nil {
			log.Fatalf("Error parsing redirect_uri: %s", parseErr)
		}

		q := u.Query()

		code := randomString("code-", 10)
		q.Set("code", code)
		q.Set("state", r.FormValue("state"))

		sessions[code] = sessionData{
			nonce:    r.FormValue("nonce"),
			user:     r.FormValue("user"),
			identity: wantsIdentity,
			sub:      r.FormValue("sub"),
		}

		emailOverride = r.FormValue("email")

		u.RawQuery = q.Encode()

		log.Printf("Redirecting to %s", u.String())
		http.Redirect(w, r, u.String(), 302)
	}
}

func userInfo(privateKey *ecdsa.PrivateKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := tokens[strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")]

		sub := randomString("sub-", 12)

		if token.sub != "" {
			sub = token.sub
		}

		email := "simulate-delivered@notifications.service.gov.uk"
		if emailOverride != "" {
			email = emailOverride
			emailOverride = ""
		}

		userInfo := UserInfoResponse{
			Sub:           sub,
			Email:         email,
			EmailVerified: true,
			Phone:         "01406946277",
			PhoneVerified: true,
			UpdatedAt:     1311280970,
		}

		if token.identity {
			givenName, familyName, birthDate := userDetails(token.user)

			userInfo.CoreIdentityJWT, _ = jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": time.Now().Add(-time.Minute).Unix(),
				"vc": map[string]any{
					"type": []string{},
					"credentialSubject": map[string]any{
						"name": []map[string]any{
							{
								"validFrom": "2000-01-01",
								"nameParts": []map[string]any{
									{"type": "GivenName", "value": givenName},
									{"type": "FamilyName", "value": familyName},
								},
							},
						},
						"birthDate": []map[string]any{
							{
								"value": birthDate,
							},
						},
					},
				},
			}).SignedString(privateKey)
		}

		log.Printf("Logging in with sub %s and email %s", sub, email)
		json.NewEncoder(w).Encode(userInfo)
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

	identityPrivateKeyBytes, _ := base64.StdEncoding.DecodeString("LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSVBheDJBYW92aXlQWDF3cndmS2FWckxEOHdQbkpJcUlicTMzZm8rWHdBZDdvQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFSlEyVmtpZWtzNW9rSTIxY1Jma0FhOXVxN0t4TTZtMmpaWUJ4cHJsVVdCWkNFZnhxMjdwVQp0Qzd5aXplVlRiZUVqUnlJaStYalhPQjFBbDhPbHFtaXJnPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo=")
	identityPrivateKey, _ := jwt.ParseECPrivateKeyFromPEM(identityPrivateKeyBytes)

	http.HandleFunc("/.well-known/openid-configuration", openIDConfig(c))
	http.HandleFunc("/.well-known/jwks", jwks())
	http.HandleFunc("/authorize", authorize())
	http.HandleFunc("/token", token(clientId, c.Issuer))
	http.HandleFunc("/userinfo", userInfo(identityPrivateKey))

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

func randomString(prefix string, length int) string {
	return prefix + stringWithCharset(length, charset)
}

func createSignedToken(nonce, clientId, issuer string) (string, error) {
	t := jwt.New(jwt.SigningMethodES256)

	t.Header["kid"] = tokenSigningKid

	t.Claims = jwt.MapClaims{
		"sub":   randomString("sub-", 10),
		"iss":   issuer,
		"nonce": nonce,
		"aud":   clientId,
		"exp":   time.Now().Add(time.Minute * 5).Unix(),
		"iat":   time.Now().Unix(),
	}

	return t.SignedString(tokenSigningKey)
}

func userDetails(key string) (givenName, familyName, birthDate string) {
	switch key {
	case "donor":
		return "Sam", "Smith", "2000-01-02"
	case "certificate-provider":
		return "Charlie", "Cooper", "1990-01-02"
	default:
		return "Someone", "Else", "2000-01-02"
	}
}
