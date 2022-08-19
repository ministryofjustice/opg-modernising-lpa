package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/zitadel/oidc/pkg/client/rp"
	httphelper "github.com/zitadel/oidc/pkg/http"
	"github.com/zitadel/oidc/pkg/oidc"

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

func authorize(w http.ResponseWriter, r *http.Request) {

}

func setToken(w http.ResponseWriter, r *http.Request) {
	log.Println("Setting token")

	//
	//// Build body for POST to OIDC /token
	//body := &TokenRequestBody{
	//	GrantType:           "authorization_code",
	//	AuthorizationCode:   "code-value",
	//	RedirectUri:         "http://localhost:5050/home",
	//	ClientAssertionType: "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	//	// TODO - generate a real JWT https://docs.sign-in.service.gov.uk/integrate-with-integration-environment/integrate-with-code-flow/#create-a-jwt-assertion
	//	ClientAssertion: "THEJWT",
	//}
	//
	//encodedPostBody := new(bytes.Buffer)
	//err := json.NewEncoder(encodedPostBody).Encode(body)
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// Build request for POST OIDC /token
	//req, err := http.NewRequest("POST", "http://oidc-proxy:5000/token", encodedPostBody)
	//if err != nil {
	//	log.Fatal("Error building req: ", err)
	//}
	//
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//
	//// POST to OIDC /token
	//res, err := http.DefaultClient.Do(req)
	//
	//if err != nil {
	//	log.Fatal("Error POSTing to /token: ", err)
	//}
	//
	//// Get private key from AWS secrets manager
	//sess, err := session.NewSession(&aws.Config{
	//	Region:      aws.String("eu-west-1"),
	//	Credentials: credentials.NewStaticCredentials("test", "test", ""),
	//	Endpoint:    aws.String("http://localstack:4566"),
	//})
	//
	//if err != nil {
	//	log.Fatalf("Problem initialising new AWS session: %v", err)
	//}
	//
	//svc := secretsmanager.New(
	//	sess,
	//	aws.NewConfig().WithRegion("eu-west-1"),
	//)
	//
	//input := &secretsmanager.GetSecretValueInput{
	//	SecretId: aws.String("private-jwt-key-base64"),
	//}
	//
	//result, err := svc.GetSecretValue(input)
	//if err != nil {
	//	log.Fatalf("Problem get secret '%s': %v", "private-jwt-key-base64", err)
	//}
	//
	//// Base64 Decode private key
	//var base64PrivateKey string
	//if result.SecretString != nil {
	//	base64PrivateKey = *result.SecretString
	//}
	//
	//privateKey, err := base64.StdEncoding.DecodeString(base64PrivateKey)
	//if err != nil {
	//	log.Fatal("error decoding base64 string: ", err)
	//}
	//
	//// Parse response from OIDC /token
	//defer res.Body.Close()
	//
	//var tokenResponse TokenResponseBody
	//
	//err = json.NewDecoder(res.Body).Decode(&tokenResponse)
	//if err != nil {
	//	log.Println(res.Body)
	//	log.Fatalf("Issues parsing token response body: %v", err)
	//}
	//
	//// Parse JWT from OIDC /token
	//token, err := jwt.Parse(tokenResponse.IdToken, func(token *jwt.Token) (interface{}, error) {
	//	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
	//		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	//	}
	//
	//	return privateKey, nil
	//})
	//
	//// Store JWT from OIDC /token
	//http.SetCookie(w, &http.Cookie{
	//	Name:  "sign-in-token",
	//	Value: token.Raw,
	//	// TODO - use exp from JWT once we are verifying claims and have access to it
	//	Expires: time.Now().Add(time.Second * time.Duration(tokenResponse.ExpiresIn)),
	//})
	//
	//// Build GET request to OIDC /userinfo
	//getBody := new(bytes.Buffer)
	//req, err = http.NewRequest("GET", "http://oidc-proxy:5000/userinfo", getBody)
	//if err != nil {
	//	log.Fatal("Error building req: ", err)
	//}
	//
	//var bearer = "Bearer " + token.Raw
	//req.Header.Add("Authorization", bearer)
	//
	//// GET OIDC /userinfo
	//res, err = http.DefaultClient.Do(req)
	//if err != nil {
	//	log.Fatal("Error making request to /userinfo: ", err)
	//}
	//
	//// Parse response from GET OIDC /userinfo
	//defer res.Body.Close()
	//var userinfoResponse UserInfoResponse
	//
	//err = json.NewDecoder(res.Body).Decode(&userinfoResponse)
	//if err != nil {
	//	log.Println(res.Body)
	//	log.Fatalf("Issues parsing userinfo response body: %v", err)
	//}

	// Store user email from /userinfo
	//http.SetCookie(w, &http.Cookie{
	//	Name:  "user-email",
	//	Value: userinfoResponse.Email,
	//	// TODO - use exp from JWT once we are verifying claims and have access to it
	//	Expires: time.Now().Add(time.Second * time.Duration(tokenResponse.ExpiresIn)),
	//})

	http.Redirect(w, r, "http://localhost:5050/home", 302)
}

var (
	callbackPath = "/auth/callback"
	key          = []byte("test1234test1234")
)

func main() {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./web/static/"))

	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	pkFilePath := saveKeyToCwd()

	clientID := env.Get("CLIENT_ID", "clientidvalue")
	clientSecret := env.Get("CLIENT_SECRET", "clientsecret")
	keyPath := env.Get("KEY_PATH", pkFilePath)
	issuer := env.Get("ISSUER", "issuervalue")
	port := env.Get("PORT", "5000")
	scopes := strings.Split(env.Get("SCOPES", "email phone"), " ")

	redirectURI := fmt.Sprintf("http://oidc-mock:%v%v", port, callbackPath)
	cookieHandler := httphelper.NewCookieHandler(key, key, httphelper.WithUnsecure())

	options := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}
	if clientSecret == "" {
		options = append(options, rp.WithPKCE(cookieHandler))
	}
	if keyPath != "" {
		options = append(options, rp.WithJWTProfile(rp.SignerFromKeyPath(keyPath)))
	}

	provider, err := rp.NewRelyingPartyOIDC(issuer, clientID, clientSecret, redirectURI, scopes, options...)
	if err != nil {
		logrus.Fatalf("error creating provider %s", err.Error())
	}

	//generate some state (representing the state of the user in your application,
	//e.g. the page where he was before sending him to login
	state := func() string {
		return uuid.New().String()
	}

	//register the AuthURLHandler at your preferred path
	//the AuthURLHandler creates the auth request and redirects the user to the auth server
	//including state handling with secure cookie and the possibility to use PKCE
	mux.HandleFunc("/login", rp.AuthURLHandler(state, provider))

	//for demonstration purposes the returned userinfo response is written as JSON object onto response
	marshalUserinfo := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens, state string, rp rp.RelyingParty, info oidc.UserInfo) {
		data, err := json.Marshal(info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}

	mux.HandleFunc(callbackPath, rp.CodeExchangeHandler(rp.UserinfoCallback(marshalUserinfo), provider))

	mux.HandleFunc("/home", home)
	mux.HandleFunc("/set_token", setToken)

	err = http.ListenAndServe(":5000", mux)

	if err != nil {
		log.Fatal(err)
	}
}

func getPrivateKey() string {
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
		log.Fatalf("Problem get secret '%s': %v", "private-jwt-key-base64", err)
	}

	return string(privateKey)
}

func saveKeyToCwd() string {
	f, err := os.Create("private_key.pem")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(getPrivateKey())

	if err2 != nil {
		log.Fatal(err2)
	}

	return f.Name()
}
