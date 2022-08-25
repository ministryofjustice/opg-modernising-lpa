package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"

	"github.com/ministryofjustice/opg-go-common/env"
	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/signin"
)

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
	awsBaseUrl := env.Get("AWS_BASE_URL", "http://localstack:4566")

	secretsClient, err := secrets.NewClient(awsBaseUrl)
	if err != nil {
		logger.Fatal(err)
	}

	signInClient := signin.NewClient(http.DefaultClient, &secretsClient)
	err = signInClient.Discover(issuer.String() + "/.well-known/openid-configuration")
	if err != nil {
		logger.Fatal(err)
	}

	signInCallbackEndpoint := "/auth/callback"
	redirectURL := fmt.Sprintf("%s%s", appPublicURL, signInCallbackEndpoint)

	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.Handle("/login", page.Login(*signInClient, appPublicURL, clientID, signInPublicURL, redirectURL))
	mux.Handle("/home", page.Home(tmpls.Get("home.gohtml"), fmt.Sprintf("%s/login", appPublicURL), bundle.For("en"), page.En))
	mux.Handle(signInCallbackEndpoint, page.SigninCallback(*signInClient, appPublicURL, clientID, random.String(12)))

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
