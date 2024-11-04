package main

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	awsBaseURL            = os.Getenv("AWS_BASE_URL")
	eventBusName          = cmp.Or(os.Getenv("EVENT_BUS_NAME"), "default")
	notifyBaseURL         = os.Getenv("GOVUK_NOTIFY_BASE_URL")
	notifyIsProduction    = os.Getenv("GOVUK_NOTIFY_IS_PRODUCTION") == "1"
	searchEndpoint        = os.Getenv("SEARCH_ENDPOINT")
	searchIndexName       = cmp.Or(os.Getenv("SEARCH_INDEX_NAME"), "lpas")
	searchIndexingEnabled = os.Getenv("SEARCH_INDEXING_DISABLED") != "1"
	tableName             = os.Getenv("LPAS_TABLE")
	xrayEnabled           = os.Getenv("XRAY_ENABLED") == "1"

	httpClient *http.Client
	cfg        aws.Config
	logHandler slog.Handler
)

func handleRunSchedule(ctx context.Context) error {
	logger := slog.New(logHandler.
		WithAttrs([]slog.Attr{
			slog.String("service_name", "opg-modernising-lpa/schedule-runner"),
		}))

	secretsClient, err := secrets.NewClient(cfg, time.Hour)
	if err != nil {
		return err
	}

	notifyApiKey, err := secretsClient.Secret(ctx, secrets.GovUkNotify)
	if err != nil {
		return fmt.Errorf("failed to get notify API secret: %w", err)
	}

	bundle, err := localize.NewBundle("./lang/en.json", "./lang/cy.json")
	if err != nil {
		return err
	}

	notifyClient, err := notify.New(logger, notifyIsProduction, notifyBaseURL, notifyApiKey, httpClient, event.NewClient(cfg, eventBusName), bundle)
	if err != nil {
		return err
	}

	dynamoClient, err := dynamo.NewClient(cfg, tableName)
	if err != nil {
		return fmt.Errorf("failed to create dynamodb client: %w", err)
	}

	eventClient := event.NewClient(cfg, eventBusName)

	searchClient, err := search.NewClient(cfg, searchEndpoint, searchIndexName, searchIndexingEnabled)
	if err != nil {
		return err
	}

	donorStore := donor.NewStore(dynamoClient, eventClient, logger, searchClient)
	scheduledStore := scheduled.NewStore(dynamoClient)

	runner := scheduled.NewRunner(logger, scheduledStore, donorStore, notifyClient)

	if err := runner.Run(ctx); err != nil {
		logger.Error("runner error", slog.Any("err", err))
		return err
	}

	return nil
}

func main() {
	ctx := context.Background()

	var err error
	cfg, err = config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("failed to load default config: %v", err)
		return
	}

	httpClient = &http.Client{Timeout: time.Second * 30}

	if len(awsBaseURL) > 0 {
		cfg.BaseEndpoint = aws.String(awsBaseURL)
	}

	if xrayEnabled {
		tp, err := telemetry.SetupLambda(ctx)
		if err != nil {
			log.Printf("error creating tracer provider: %v", err)
		}

		otelaws.AppendMiddlewares(&cfg.APIOptions)
		telemetry.AppendMiddlewares(&cfg.APIOptions)
		httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)

		defer func(ctx context.Context) {
			if err := tp.Shutdown(ctx); err != nil {
				log.Printf("error shutting down tracer provider: %v", err)
			}
		}(ctx)

		logHandler = telemetry.NewSlogHandler(slog.
			NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					switch a.Value.Kind() {
					case slog.KindAny:
						switch v := a.Value.Any().(type) {
						case *http.Request:
							return slog.Group(a.Key,
								slog.String("method", v.Method),
								slog.String("uri", v.URL.String()))
						}
					}

					return a
				},
			}))

		lambda.Start(otellambda.InstrumentHandler(handleRunSchedule, xrayconfig.WithRecommendedOptions(tp)...))
	} else {
		logHandler = slog.NewJSONHandler(os.Stdout, nil)
		lambda.Start(handleRunSchedule)
	}
}
