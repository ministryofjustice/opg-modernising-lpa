package main

import (
	"context"
	"fmt"
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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
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
	logger := logging.New(os.Stdout, "opg-modernising-lpa")

	issuer, err := url.Parse(env.Get("GOV_UK_SIGN_IN_URL", "http://sign-in-mock:5060"))
	if err != nil {
		logger.Fatal(err)
	}

	clientID := env.Get("CLIENT_ID", "client-id-value")
	appPort := env.Get("APP_PORT", "8080")
	appPublicURL := env.Get("APP_PUBLIC_URL", "http://localhost:5050")
	signInPublicURL := env.Get("GOV_UK_SIGN_IN_PUBLIC_URL", "http://localhost:7012")

	webDir := env.Get("WEB_DIR", "web")

	tmpls, err := template.Parse(webDir+"/template", map[string]interface{}{
		"isEnglish": func(lang page.Lang) bool {
			return lang == page.En
		},
		"isWelsh": func(lang page.Lang) bool {
			return lang == page.Cy
		},
	})
	if err != nil {
		logger.Fatal(err)
	}

	bundle := localize.NewBundle("lang/en.json", "lang/cy.json")

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir(webDir + "/static/"))

	signInClient := govuksignin.NewClient(http.DefaultClient, issuer.String(), "/auth/callback")
	err = signInClient.Init()
	if err != nil {
		logger.Fatal(err)
	}

	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.Handle("/login", page.Login(*signInClient, appPublicURL, clientID, signInPublicURL))
	mux.Handle("/home", page.Home(tmpls.Get("home.gohtml"), fmt.Sprintf("%s/login", appPublicURL), bundle.For("en"), page.En))
	mux.Handle("/auth/callback", page.SetToken(*signInClient, appPublicURL, clientID, RandomString(12)))

	mux.Handle("/cy/", page.App(logger, bundle.For("cy"), page.Cy, tmpls))
	mux.Handle("/", page.App(logger, bundle.For("en"), page.En, tmpls))

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
