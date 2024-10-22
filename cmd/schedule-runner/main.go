package main

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
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
	lambdadetector "go.opentelemetry.io/contrib/detectors/aws/lambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/appengine/log"
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
)

func setup(ctx context.Context, stdOutOverride bool, resource *resource.Resource) (func(context.Context) error, error) {
	var exporter sdktrace.SpanExporter
	var err error

	if stdOutOverride {
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
			stdouttrace.WithWriter(os.Stdout),
		)
	} else {
		exporter, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint("0.0.0.0:4317"),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(
		exporter,
		sdktrace.WithBatchTimeout(100*time.Millisecond),
		sdktrace.WithMaxExportBatchSize(2),
		// Ensure spans are exported before Lambda freezes
		sdktrace.WithExportTimeout(200*time.Millisecond),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithIDGenerator(xray.NewIDGenerator()),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})

	return func(ctx context.Context) error {
		flushCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
		defer cancel()
		if err := tp.ForceFlush(flushCtx); err != nil {
			return fmt.Errorf("failed to flush tracer: %w", err)
		}
		shutdownCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
		defer cancel()
		return tp.Shutdown(shutdownCtx)
	}, nil
}

func handleRunSchedule(ctx context.Context) error {
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

	return nil
}

func main() {
	ctx := context.Background()

	var err error
	cfg, err = config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Errorf(ctx, "failed to load default config: %v", err)
		return
	}

	httpClient = &http.Client{Timeout: time.Second * 30}

	if len(awsBaseURL) > 0 {
		cfg.BaseEndpoint = aws.String(awsBaseURL)
	}

	if xrayEnabled {
		resource, err := lambdadetector.NewResourceDetector().Detect(ctx)
		if err != nil {
			log.Errorf(ctx, "failed to detect resource: %v", err)
			return
		}

		shutdown, err := setup(ctx, strings.Contains(notifyBaseURL, "mock-notify"), resource)
		if err != nil {
			log.Errorf(ctx, "failed to instrument telemetry: %v", err)
			return
		}

		// Wrap handler with proper context and shutdown handling
		handler := func(ctx context.Context) error {
			err := handleRunSchedule(ctx)
			if shutdown != nil {
				if shutdownErr := shutdown(ctx); shutdownErr != nil {
					log.Errorf(ctx, "error shutting down tracer provider: %v", shutdownErr)
				}
			}
			return err
		}

		otelaws.AppendMiddlewares(&cfg.APIOptions)
		httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)

		tp := otellambda.WithTracerProvider(otel.GetTracerProvider())
		lambda.Start(otellambda.InstrumentHandler(handler, tp))
	} else {
		lambda.Start(handleRunSchedule)
	}
}
