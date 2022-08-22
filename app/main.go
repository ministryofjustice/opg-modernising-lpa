package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"

	govuksignin "github.com/ministryofjustice/opg-modernising-lpa/internal/gov_uk_sign_in"

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

func home(w http.ResponseWriter, r *http.Request) {
	emailCookie, err := r.Cookie("user-email")

	userEmail := ""
	signInUrl := ""

	// Building login URL
	if err != nil {
		log.Println("setting sign in link to http://app:5000/login")
		log.Printf("cookie err is: %v", err)
		signInUrl = "http://app:5000/login"
	} else {
		log.Printf("setting userEmail to %s from cookie", emailCookie.Value)
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
	log.Println("/auth/callback")

	issuer, err := url.Parse(env.Get("ISSUER", "http://sign-in-mock:5060"))

	if err != nil {
		log.Fatalf("Issues parsing issuer URL: %v", err)
	}

	signInClient := govuksignin.NewClient(http.DefaultClient, issuer.String())

	jwt, err := signInClient.GetToken()

	if err != nil {
		log.Fatalf("Error getting token: %v", err)
	}

	log.Println("Setting token")

	http.SetCookie(w, &http.Cookie{
		Name:  "sign-in-token",
		Value: jwt.Raw,
		// TODO - use exp from JWT once we are verifying claims and have access to it
		Expires: time.Now().Add(time.Second * time.Duration(5)),
	})

	userInfo, err := signInClient.GetUserInfo(jwt)

	if err != nil {
		log.Fatalf("Error getting user info: %v", err)
	}

	log.Printf("setting user-email cookie to %s", userInfo.Email)

	//Store user email from /userinfo
	http.SetCookie(w, &http.Cookie{
		Name:  "user-email",
		Value: userInfo.Email,
		// TODO - use exp from JWT once we are verifying claims and have access to it
		Expires: time.Now().Add(time.Second * time.Duration(5)),
	})

	http.Redirect(w, r, "http://app:5000/home", http.StatusFound)
}

func login(w http.ResponseWriter, r *http.Request) {
	log.Println("/login")
	issuer, err := url.Parse(env.Get("ISSUER", "http://sign-in-mock:5060"))
	clientID := env.Get("CLIENT_ID", "client-id-value")
	appHost := env.Get("APP_HOST", "http://app")
	appPort := env.Get("APP_PORT", "5000")

	if err != nil {
		log.Fatalf("Issues parsing issuer URL: %v", err)
	}

	signInClient := govuksignin.NewClient(http.DefaultClient, issuer.String())

	redirectURL := fmt.Sprintf("%s:%s%s", appHost, appPort, signInClient.AuthCallbackPath)
	err = signInClient.AuthorizeAndRedirect(redirectURL, clientID, "state-value", "nonce-value", "scope-value")

	if err != nil {
		log.Fatalf("Error GETting authorize: %v", err)
	}
	//discoverEndpoint := issuer.String() + "/.well-known/openid-configuration"
	//
	//log.Println(discoverEndpoint)
	//
	//// Call out to discovery endpoint
	//req, err := http.NewRequest("GET", discoverEndpoint, nil)
	//if err != nil {
	//	log.Fatalf("Error building discover request: %v", err)
	//}
	//
	//res, err := http.DefaultClient.Do(req)
	//
	//if err != nil {
	//	log.Fatalf("Error GETting discover data: %v", err)
	//}
	//
	//defer res.Body.Close()
	//
	//// Add all endpoints needed for future calls to a struct
	//err = json.NewDecoder(res.Body).Decode(&discoverData)
	//if err != nil {
	//	log.Println(res.Body)
	//	log.Fatalf("Issues parsing discover response body: %v", err)
	//}

	//authorizeUrl, err := url.Parse(signInClient.DiscoverData.AuthorizationEndpoint)
	//
	//if err != nil {
	//	log.Fatalf("Issues parsing auth endpoint URL: %v", err)
	//}
	//
	//if issuer.Host != authorizeUrl.Host {
	//	log.Fatalf("Host of authorize URL does not match issuer. Wanted %s, Got: %s", issuer.Host, authorizeUrl.Host)
	//}
	//
	//q := authorizeUrl.Query()
	////TODO use env var host and port once added
	//q.Set("redirect_uri", "http://app:5000"+callbackPath)
	//q.Set("client_id", clientID)
	//q.Set("state", "state-value")
	//q.Set("nonce", "nonce-value")
	//q.Set("scope", "scope-value")
	//authorizeUrl.RawQuery = q.Encode()
	//
	//// Call out to authorize endpoint
	//req, err = http.NewRequest("GET", authorizeUrl.String(), nil)
	//if err != nil {
	//	log.Fatalf("Error building authorise request: %v", err)
	//}
	//
	//res, err = http.DefaultClient.Do(req)
	//
	//if err != nil {
	//	log.Fatalf("Error GETting discover data: %v", err)
	//}
}

func main() {
	//TODO move port and host to env vars
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./web/static/"))

	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/login", login)
	mux.HandleFunc("/home", home)
	mux.HandleFunc("/auth/callback", setToken)

	err := http.ListenAndServe(":5000", mux)

	if err != nil {
		log.Fatal(err)
	}
}
