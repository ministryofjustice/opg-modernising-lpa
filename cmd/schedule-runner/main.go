package main

import (
	"cmp"
	"context"
	"fmt"
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
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
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
)

func handleRunSchedule(ctx context.Context) error {
	//start := time.Now()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load default config: %w", err)
	}

	//if xrayEnabled {
	//	ctx := context.Background()
	//
	//	resource, err := lambdadetector.NewResourceDetector().Detect(ctx)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	shutdown, err := telemetry.Setup(ctx, strings.Contains(notifyBaseURL, "mock-notify"), resource)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	defer shutdown(ctx)
	//
	//	httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)
	//}

	httpClient := &http.Client{Timeout: time.Second * 30}

	if len(awsBaseURL) > 0 {
		cfg.BaseEndpoint = aws.String(awsBaseURL)
	}

	//otelaws.AppendMiddlewares(&cfg.APIOptions)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil).
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

	//// Calculate adaptive flush timeout based on execution time
	//executionTime := time.Since(start)
	//flushTimeout := 200 * time.Millisecond // minimum timeout
	//if executionTime > time.Second {
	//	// For longer executions, give more time for flush
	//	flushTimeout = 500 * time.Millisecond
	//}
	//
	//// Force flush before returning
	//if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
	//	flushCtx, cancel := context.WithTimeout(ctx, flushTimeout)
	//	defer cancel()
	//	_ = tp.ForceFlush(flushCtx)
	//}

	return nil
}

func main() {
	if xrayEnabled {
		ctx := context.Background()
		tp, err := xrayconfig.NewTracerProvider(ctx)
		if err != nil {
			fmt.Printf("error creating tracer provider: %v", err)
		}

		defer func(ctx context.Context) {
			err := tp.Shutdown(ctx)
			if err != nil {
				fmt.Printf("error shutting down tracer provider: %v", err)
			}
		}(ctx)

		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(xray.Propagator{})

		lambda.Start(otellambda.InstrumentHandler(handleRunSchedule, xrayconfig.WithRecommendedOptions(tp)...))
	} else {
		lambda.Start(handleRunSchedule)
	}

	//handler := otellambda.InstrumentHandler(handleRunSchedule,
	//	otellambda.WithTracerProvider(otel.GetTracerProvider()),
	//)
	//lambda.Start(handler)
}
