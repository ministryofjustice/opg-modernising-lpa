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
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
)

func handleRunSchedule(ctx context.Context) error {
	var (
		eventBusName          = cmp.Or(os.Getenv("EVENT_BUS_NAME"), "default")
		notifyBaseURL         = os.Getenv("GOVUK_NOTIFY_BASE_URL")
		notifyIsProduction    = os.Getenv("GOVUK_NOTIFY_IS_PRODUCTION") == "1"
		searchEndpoint        = os.Getenv("SEARCH_ENDPOINT")
		searchIndexName       = cmp.Or(os.Getenv("SEARCH_INDEX_NAME"), "lpas")
		searchIndexingEnabled = os.Getenv("SEARCH_INDEXING_DISABLED") != "1"
		tableName             = os.Getenv("LPAS_TABLE")
	)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil).
		WithAttrs([]slog.Attr{
			slog.String("service_name", "opg-modernising-lpa/schedule-runner"),
		}))

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load default config: %w", err)
	}

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

	notifyClient, err := notify.New(logger, notifyIsProduction, notifyBaseURL, notifyApiKey, http.DefaultClient, event.NewClient(cfg, eventBusName), bundle)
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
	lambda.Start(handleRunSchedule)
}
