package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/env"
	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/signin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
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

	cleanup := initTracer()
	defer cleanup(context.Background())

	// load aws config
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Fatal(err)
	}

	// instrument all aws clients
	otelaws.AppendMiddlewares(&cfg.APIOptions)

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

func initTracer() func(context.Context) error {
	logger := logging.New(os.Stdout, "opg-modernising-lpa-xray-init")

	var (
		serviceName  = env.Get("SERVICE_NAME", "modernising-lpa")
		collectorURL = env.Get("OTEL_EXPORTER_OTLP_ENDPOINT", "0.0.0.0:4317")
		insecure     = env.Get("INSECURE_MODE", "true")
	)

	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if len(insecure) > 0 {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(collectorURL),
		),
	)

	if err != nil {
		logger.Fatal(err)
	}

	// // Create a new ID Generator
	// idg := xray.NewIDGenerator()

	// // Instantiate a new ECS Resource detector
	// ecsResourceDetector := ecs.NewResourceDetector()
	// resource, err := ecsResourceDetector.Detect(context.Background())
	// if err != nil {
	// 	log.Fatal(err)
	// }

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		logger.Fatal(err)
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			// sdktrace.WithIDGenerator(idg),
			sdktrace.WithResource(resources),
		),
	)
	return exporter.Shutdown
}
