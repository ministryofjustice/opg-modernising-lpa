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

type PageData struct {
	WebDir      string
	ServiceName string
	UserEmail   string
	SignInURL   string
}

func home(w http.ResponseWriter, r *http.Request) {
	emailCookie, err := r.Cookie("user-email")

	var userEmail string
	var signInUrl string

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
