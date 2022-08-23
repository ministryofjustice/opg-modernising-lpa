package main

import (
	"context"
	"fmt"
	html "html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	govuksignin "github.com/ministryofjustice/opg-modernising-lpa/internal/gov_uk_sign_in"

	"github.com/ministryofjustice/opg-go-common/env"
	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

//func login(w http.ResponseWriter, r *http.Request) {
//	log.Println("/login")
//	signInClient := govuksignin.NewClient(http.DefaultClient, issuer.String())
//
//	redirectURL := fmt.Sprintf("%s%s", appBaseURL, signInClient.AuthCallbackPath)
//	err := signInClient.AuthorizeAndRedirect(redirectURL, clientID, "state-value", "nonce-value", "scope-value")
//
//	if err != nil {
//		log.Fatalf("Error GETting authorize: %v", err)
//	}
//}

//func setToken(w http.ResponseWriter, r *http.Request) {
//	log.Println("/auth/callback")
//
//	signInClient := govuksignin.NewClient(http.DefaultClient, issuer.String())
//
//	jwt, err := signInClient.GetToken(fmt.Sprintf("%s:%s", appBaseURL, "/home"))
//
//	if err != nil {
//		log.Fatalf("Error getting token: %v", err)
//	}
//
//	userInfo, err := signInClient.GetUserInfo(jwt)
//
//	if err != nil {
//		log.Fatalf("Error getting user info: %v", err)
//	}
//
//	redirectURL, err := url.Parse(fmt.Sprintf("%s/home", appBaseURL))
//
//	if err != nil {
//		log.Fatalf("Error parsing redirect URL: %v", err)
//	}
//
//	redirectURL.Query().Add("email", userInfo.Email)
//
//	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
//}

//func home(w http.ResponseWriter, r *http.Request) {
//	requestURI, err := url.Parse(r.RequestURI)
//
//	if err != nil {
//		log.Fatalf("Error parsing requestURI: %v", err)
//	}
//
//	var userEmail string
//	var signInUrl string
//
//	// Building login URL
//	if requestURI.Query().Get("user_email") != "" {
//		log.Printf("setting userEmail to %s from query", requestURI.Query().Get("user_email"))
//		userEmail = requestURI.Query().Get("user_email")
//	} else {
//		log.Printf("user email not set - setting login")
//		signInUrl = fmt.Sprintf("%s/login", appBaseURL)
//	}
//
//	// Building template
//	webDir := env.Get("WEB_DIR", "web")
//
//	data := PageData{
//		WebDir:      webDir,
//		ServiceName: "Modernising LPA",
//		UserEmail:   userEmail,
//		SignInURL:   signInUrl,
//	}
//
//	files := []string{
//		path.Join(webDir, "/template/home.gohtml"),
//		path.Join(webDir, "/template/layout/base.gohtml"),
//	}
//
//	ts, err := template.ParseFiles(files...)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Serve template
//	err = ts.ExecuteTemplate(w, "base", data)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	log.Println("home")
//}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

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

func main() {
	issuer, err := url.Parse(env.Get("GOV_UK_SIGN_IN_URL", "http://sign-in-mock:5060"))

	if err != nil {
		log.Fatalf("Issues parsing issuer URL: %v", err)
	}

	clientID := env.Get("CLIENT_ID", "client-id-value")
	appHost := env.Get("APP_HOST", "http://app")
	appPort := env.Get("APP_PORT", "8080")
	appBaseURL := fmt.Sprintf("%s:%s", appHost, appPort)

	logger := logging.New(os.Stdout, "opg-modernise-lpa")
	webDir := env.Get("WEB_DIR", "web")

	tmpls, err := template.Parse(webDir+"/template", html.FuncMap{})
	if err != nil {
		logger.Fatal(err)
	}

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir(webDir + "/static/"))

	signInClient := govuksignin.NewClient(http.DefaultClient, issuer.String())

	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	mux.Handle("/", page.Start(tmpls.Get("start.gohtml")))
	mux.Handle("/login", page.Login(*signInClient, appBaseURL, clientID))
	mux.Handle("/home", page.Home(tmpls.Get("home.gohtml"), appBaseURL))
	mux.Handle("/auth/callback", page.SetToken(*signInClient, appBaseURL, clientID, RandomString(12)))

	server := &http.Server{
		Addr:              ":" + appPort,
		Handler:           mux,
		ReadHeaderTimeout: 20 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Fatal(err)
		}
	}()

	logger.Print("Running at :" + appPort)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	sig := <-c
	logger.Print("signal received: ", sig)

	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(tc); err != nil {
		logger.Print(err)
	}
}
