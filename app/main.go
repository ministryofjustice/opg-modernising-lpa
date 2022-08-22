package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ministryofjustice/opg-go-common/env"
	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func main() {
	logger := logging.New(os.Stdout, "opg-modernising-lpa")

	port := env.Get("PORT", "8080")
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

	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	mux.Handle("/cy/", page.App(logger, bundle.For("cy"), page.Cy, tmpls))
	mux.Handle("/", page.App(logger, bundle.For("en"), page.En, tmpls))

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

	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(tc); err != nil {
		logger.Print(err)
	}
}
