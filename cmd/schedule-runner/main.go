package main

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	awslambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lambda"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	awsBaseURL   = os.Getenv("AWS_BASE_URL")
	eventBusName = cmp.Or(os.Getenv("EVENT_BUS_NAME"), "default")
	// TODO remove in MLPAB-2690
	metricsEnabled        = os.Getenv("METRICS_ENABLED") == "1"
	notifyBaseURL         = os.Getenv("GOVUK_NOTIFY_BASE_URL")
	notifyIsProduction    = os.Getenv("GOVUK_NOTIFY_IS_PRODUCTION") == "1"
	searchEndpoint        = os.Getenv("SEARCH_ENDPOINT")
	searchIndexName       = cmp.Or(os.Getenv("SEARCH_INDEX_NAME"), "lpas")
	searchIndexingEnabled = os.Getenv("SEARCH_INDEXING_DISABLED") != "1"
	tableName             = os.Getenv("LPAS_TABLE")
	xrayEnabled           = os.Getenv("XRAY_ENABLED") == "1"
	lpaStoreBaseURL       = os.Getenv("LPA_STORE_BASE_URL")
	lpaStoreSecretARN     = os.Getenv("LPA_STORE_SECRET_ARN")
	appPublicURL          = os.Getenv("APP_PUBLIC_URL")

	Tag string

	httpClient *http.Client
	cfg        aws.Config
	logger     *slog.Logger
)

func handleRunSchedule(ctx context.Context) error {
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

	lambdaClient := lambda.New(cfg, v4.NewSigner(), httpClient, time.Now)
	lpaStoreClient := lpastore.New(lpaStoreBaseURL, secretsClient, lpaStoreSecretARN, lambdaClient)

	scheduledStore := scheduled.NewStore(dynamoClient)
	donorStore := donor.NewStore(dynamoClient, eventClient, logger, searchClient)
	certificateProviderStore := certificateprovider.NewStore(dynamoClient)
	lpaStoreResolvingService := lpastore.NewResolvingService(donorStore, lpaStoreClient)

	if Tag == "" {
		Tag = os.Getenv("TAG")
	}

	client := cloudwatch.NewFromConfig(cfg)
	metricsClient := telemetry.NewMetricsClient(client, Tag)

	runner := scheduled.NewRunner(
		logger,
		scheduledStore,
		donorStore,
		certificateProviderStore,
		lpaStoreResolvingService,
		notifyClient,
		eventClient,
		bundle,
		metricsClient,
		metricsEnabled,
		appPublicURL,
	)

	if err = runner.Run(ctx); err != nil {
		logger.Error("runner error", slog.Any("err", err))
		return err
	}

	return nil
}

func main() {
	ctx := context.Background()

	httpClient = &http.Client{Timeout: time.Second * 30}

	logger = slog.New(telemetry.NewSlogHandler(slog.
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
		})).
		WithAttrs([]slog.Attr{
			slog.String("service_name", "opg-modernising-lpa/schedule-runner"),
		}))

	var err error
	cfg, err = config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to load default config", slog.Any("err", err))
		return
	}

	if len(awsBaseURL) > 0 {
		cfg.BaseEndpoint = aws.String(awsBaseURL)
	}

	var tp *trace.TracerProvider
	if xrayEnabled {
		tp, err = telemetry.SetupLambda(ctx, &cfg.APIOptions)
		if err != nil {
			logger.WarnContext(ctx, "error creating tracer provider", slog.Any("err", err))
		}
	}

	if tp != nil {
		httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)

		defer func(ctx context.Context) {
			if err := tp.Shutdown(ctx); err != nil {
				logger.WarnContext(ctx, "error shutting down tracer provider", slog.Any("err", err))
			}
		}(ctx)

		awslambda.Start(otellambda.InstrumentHandler(handleRunSchedule, xrayconfig.WithRecommendedOptions(tp)...))
	} else {
		awslambda.Start(handleRunSchedule)
	}
}
