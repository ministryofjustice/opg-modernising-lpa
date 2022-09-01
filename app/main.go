package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/env"
	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/signin"
)

func main() {
	ctx := context.Background()
	logger := logging.New(os.Stdout, "opg-modernising-lpa")

	var (
		port         = env.Get("APP_PORT", "8080")
		appPublicURL = env.Get("APP_PUBLIC_URL", "http://localhost:5050")
		webDir       = env.Get("WEB_DIR", "web")
		awsBaseUrl   = env.Get("AWS_BASE_URL", "")
		clientID     = env.Get("CLIENT_ID", "client-id-value")
		issuer       = env.Get("ISSUER", "http://sign-in-mock:7012")
	)

	tmpls, err := template.Parse(webDir+"/template", map[string]interface{}{
		"isEnglish": func(lang page.Lang) bool {
			return lang == page.En
		},
		"isWelsh": func(lang page.Lang) bool {
			return lang == page.Cy
		},
		"input": func(top interface{}, name, label string, value interface{}, attrs ...interface{}) map[string]interface{} {
			field := map[string]interface{}{
				"top":   top,
				"name":  name,
				"label": label,
				"value": value,
			}

			if len(attrs)%2 != 0 {
				panic("must have even number of attrs")
			}

			for i := 0; i < len(attrs); i += 2 {
				field[attrs[i].(string)] = attrs[i+1]
			}

			return field
		},
		"errorMessage": func(top interface{}, name string) map[string]interface{} {
			return map[string]interface{}{
				"top":  top,
				"name": name,
			}
		},
		"details": func(top interface{}, name, detail string) map[string]interface{} {
			return map[string]interface{}{
				"top":    top,
				"name":   name,
				"detail": detail,
			}
		},
		"inc": func(i int) int {
			return i + 1
		},
		"link": func(lang page.Lang, path string) string {
			if lang == page.Cy {
				return "/cy" + path
			}

			return path
		},
	})
	if err != nil {
		logger.Fatal(err)
	}

	bundle := localize.NewBundle("lang/en.json", "lang/cy.json")

	secretsClient, err := secrets.NewClient(awsBaseUrl)
	if err != nil {
		logger.Fatal(err)
	}

	sessionKeys, err := secretsClient.CookieSessionKeys()
	if err != nil {
		logger.Fatal(err)
	}

	sessionStore := sessions.NewCookieStore(sessionKeys...)

	redirectURL := fmt.Sprintf("%s%s", appPublicURL, page.AuthRedirectPath)

	signInClient, err := signin.Discover(ctx, logger, http.DefaultClient, secretsClient, issuer, clientID, redirectURL)
	if err != nil {
		logger.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(webDir+"/static/"))))
	mux.Handle(page.AuthRedirectPath, page.AuthRedirect(logger, signInClient, sessionStore))
	mux.Handle(page.AuthPath, page.Login(logger, signInClient, sessionStore, random.String))
	mux.Handle("/cookies-consent", page.CookieConsent())
	mux.Handle("/cy/", http.StripPrefix("/cy", page.App(logger, bundle.For("cy"), page.Cy, tmpls, sessionStore)))
	mux.Handle("/", page.App(logger, bundle.For("en"), page.En, tmpls, sessionStore))

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 20 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Fatal(err)
		}
	}()

	logger.Print("Running at :" + port)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	sig := <-c
	logger.Print("signal received: ", sig)

	tc, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := server.Shutdown(tc); err != nil {
		logger.Print(err)
	}
}
