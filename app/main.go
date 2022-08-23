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
	appPort := env.Get("APP_PORT", "8080")
	appPublicURL := env.Get("APP_PUBLIC_URL", "http://localhost:5050")
	signInPublicURL := env.Get("GOV_UK_SIGN_IN_PUBLIC_URL", "http://localhost:7012")

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
	mux.Handle("/login", page.Login(*signInClient, appPublicURL, clientID, signInPublicURL))
	mux.Handle("/home", page.Home(tmpls.Get("home.gohtml"), fmt.Sprintf("%s/login", appPublicURL)))
	mux.Handle("/auth/callback", page.SetToken(*signInClient, appPublicURL, clientID, RandomString(12)))

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
