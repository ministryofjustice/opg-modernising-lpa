package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/golang-jwt/jwt"

	"github.com/ministryofjustice/opg-go-common/env"
)

func Hello() string {
	return "Hello, world!"
}

type PageData struct {
	WebDir      string
	ServiceName string
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

func home(w http.ResponseWriter, r *http.Request) {
	webDir := env.Get("WEB_DIR", "web")

	data := PageData{
		WebDir:      webDir,
		ServiceName: "Modernising LPA",
	}

	files := []string{
		path.Join(webDir, "/template/home.gohtml"),
		path.Join(webDir, "/template/layout/base.gohtml"),
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		log.Fatal(err)
	}

	err = ts.ExecuteTemplate(w, "base", data)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("home")
}

func setToken(w http.ResponseWriter, r *http.Request) {
	log.Println("Setting token")
	log.Println(r.URL.Query().Get("code"))

	body := &TokenRequestBody{
		GrantType:           "authorization_code",
		AuthorizationCode:   "code-value",
		RedirectUri:         "http://localhost:5050/home",
		ClientAssertionType: "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		// TODO - generate a real JWT https://docs.sign-in.service.gov.uk/integrate-with-integration-environment/integrate-with-code-flow/#create-a-jwt-assertion
		ClientAssertion: "THEJWT",
	}

	payloadBuf := new(bytes.Buffer)
	err := json.NewEncoder(payloadBuf).Encode(body)

	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", "http://oidc-proxy:5000/token", payloadBuf)
	if err != nil {
		log.Fatal("Error building req: ", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal("Error POSTing to /token: ", err)
	}

	defer res.Body.Close()

	// Print the body to the stdout
	_, err = io.Copy(os.Stdout, res.Body)

	if err != nil {
		log.Fatal(err)
	}

	var tokenResponse TokenResponseBody

	err = json.NewDecoder(r.Body).Decode(&tokenResponse)
	if err != nil {
		log.Fatalf("Issues parsing token response body: %v", err)
	}

	//awslocal secretsmanager create-secret --name "default/private-jwt-key-base64" --secret-string "$(base64 private.pem)"
	//awslocal secretsmanager create-secret --name "default/public-jwt-key-base64" --secret-string "$(base64 public.pem)"

	token, err := jwt.Parse(tokenResponse.IdToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return hmacSampleSecret, nil
	})

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: token.Raw,
		// TODO - use exp from JWT once we are verifying claims and have access to it
		Expires: time.Now().Add(time.Second * time.Duration(tokenResponse.ExpiresIn)),
	})
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
