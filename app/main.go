package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/golang-jwt/jwt"

	"github.com/ministryofjustice/opg-go-common/env"
)

func Hello() string {
	return "Hello, world!"
}

type PageData struct {
	WebDir      string
	ServiceName string
	UserEmail   string
	SignInURL   string
}

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

type UserInfoResponse struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Phone         string `json:"phone"`
	PhoneVerified bool   `json:"phone_verified"`
	UpdatedAt     int    `json:"updated_at"`
}

func home(w http.ResponseWriter, r *http.Request) {
	emailCookie, err := r.Cookie("user-email")

	userEmail := ""
	signInUrl := ""

	// Building login URL
	if err != nil {
		u, parseErr := url.Parse("http://localhost:7011/authorize")

		if parseErr != nil {
			log.Fatal(parseErr)
		}

		q := u.Query()
		q.Set("redirect_uri", "http://localhost:5050/set_token")
		q.Set("client_id", "client-credentials-mock-client")
		q.Set("state", "state-value")
		q.Set("nonce", "nonce-value")
		q.Set("scope", "scope-value")
		u.RawQuery = q.Encode()

		signInUrl = u.String()
	} else {
		userEmail = emailCookie.Value
	}

	// Building template
	webDir := env.Get("WEB_DIR", "web")

	data := PageData{
		WebDir:      webDir,
		ServiceName: "Modernising LPA",
		UserEmail:   userEmail,
		SignInURL:   signInUrl,
	}

	files := []string{
		path.Join(webDir, "/template/home.gohtml"),
		path.Join(webDir, "/template/layout/base.gohtml"),
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		log.Fatal(err)
	}

	// Serve template
	err = ts.ExecuteTemplate(w, "base", data)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("home")
}

func setToken(w http.ResponseWriter, r *http.Request) {
	log.Println("Setting token")

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
		log.Fatal(err)
	}

	// Build request for POST OIDC /token
	req, err := http.NewRequest("POST", "http://oidc-proxy:5000/token", encodedPostBody)
	if err != nil {
		log.Fatal("Error building req: ", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// POST to OIDC /token
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal("Error POSTing to /token: ", err)
	}

	// Get private key from AWS secrets manager
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("eu-west-1"),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
		Endpoint:    aws.String("http://localstack:4566"),
	})

	if err != nil {
		log.Fatalf("Problem initialising new AWS session: %v", err)
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
		log.Fatalf("Problem get secret '%s': %v", "private-jwt-key-base64", err)
	}

	// Base64 Decode private key
	var base64PrivateKey string
	if result.SecretString != nil {
		base64PrivateKey = *result.SecretString
	}

	privateKey, err := base64.StdEncoding.DecodeString(base64PrivateKey)
	if err != nil {
		log.Fatal("error decoding base64 string: ", err)
	}

	// Parse response from OIDC /token
	defer res.Body.Close()

	var tokenResponse TokenResponseBody

	err = json.NewDecoder(res.Body).Decode(&tokenResponse)
	if err != nil {
		log.Println(res.Body)
		log.Fatalf("Issues parsing token response body: %v", err)
	}

	// Parse JWT from OIDC /token
	token, err := jwt.Parse(tokenResponse.IdToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return privateKey, nil
	})

	// Store JWT from OIDC /token
	http.SetCookie(w, &http.Cookie{
		Name:  "sign-in-token",
		Value: token.Raw,
		// TODO - use exp from JWT once we are verifying claims and have access to it
		Expires: time.Now().Add(time.Second * time.Duration(tokenResponse.ExpiresIn)),
	})

	// Build GET request to OIDC /userinfo
	getBody := new(bytes.Buffer)
	req, err = http.NewRequest("GET", "http://oidc-proxy:5000/userinfo", getBody)
	if err != nil {
		log.Fatal("Error building req: ", err)
	}

	var bearer = "Bearer " + token.Raw
	req.Header.Add("Authorization", bearer)

	// GET OIDC /userinfo
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Error making request to /userinfo: ", err)
	}

	// Parse response from GET OIDC /userinfo
	defer res.Body.Close()
	var userinfoResponse UserInfoResponse

	err = json.NewDecoder(res.Body).Decode(&userinfoResponse)
	if err != nil {
		log.Println(res.Body)
		log.Fatalf("Issues parsing userinfo response body: %v", err)
	}

	// Store user email from /userinfo
	http.SetCookie(w, &http.Cookie{
		Name:  "user-email",
		Value: userinfoResponse.Email,
		// TODO - use exp from JWT once we are verifying claims and have access to it
		Expires: time.Now().Add(time.Second * time.Duration(tokenResponse.ExpiresIn)),
	})

	http.Redirect(w, r, "http://localhost:5050/home", 302)
}

func main() {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./web/static/"))

	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/home", home)
	mux.HandleFunc("/set_token", setToken)

	err := http.ListenAndServe(":5000", mux)

	if err != nil {
		log.Fatal(err)
	}
}
