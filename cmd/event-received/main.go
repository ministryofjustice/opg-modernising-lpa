package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-go-common/env"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/app"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/s3"
)

const (
	virusFound = "infected"
)

type uidEvent struct {
	UID string `json:"uid"`
}

type dynamodbClient interface {
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error
	OneByUID(ctx context.Context, uid string, v interface{}) error
	OneByPK(ctx context.Context, pk dynamo.PK, v interface{}) error
	OneBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	Put(ctx context.Context, v interface{}) error
	UpdateReturn(ctx context.Context, pk dynamo.PK, sk dynamo.SK, values map[string]dynamodbtypes.AttributeValue, expression string) (map[string]dynamodbtypes.AttributeValue, error)
	DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error
}

type s3Client interface {
	GetObjectTags(ctx context.Context, key string) ([]types.Tag, error)
}

type DocumentStore interface {
	UpdateScanResults(ctx context.Context, lpaID, objectKey string, virusDetected bool) error
}

type Event struct {
	events.S3Event
	events.CloudWatchEvent
}

func (e Event) isS3Event() bool {
	return len(e.Records) > 0
}

func (e Event) isCloudWatchEvent() bool {
	return e.Source == "aws.cloudwatch" || e.Source == "opg.poas.makeregister" || e.Source == "opg.poas.sirius"
}

func handler(ctx context.Context, event Event) error {
	var (
		tableName             = os.Getenv("LPAS_TABLE")
		notifyIsProduction    = os.Getenv("GOVUK_NOTIFY_IS_PRODUCTION") == "1"
		appPublicURL          = os.Getenv("APP_PUBLIC_URL")
		awsBaseURL            = os.Getenv("AWS_BASE_URL")
		notifyBaseURL         = os.Getenv("GOVUK_NOTIFY_BASE_URL")
		evidenceBucketName    = os.Getenv("UPLOADS_S3_BUCKET_NAME")
		uidBaseURL            = os.Getenv("UID_BASE_URL")
		lpaStoreBaseURL       = os.Getenv("LPA_STORE_BASE_URL")
		eventBusName          = env.Get("EVENT_BUS_NAME", "default")
		searchEndpoint        = os.Getenv("SEARCH_ENDPOINT")
		searchIndexingEnabled = env.Get("SEARCH_INDEXING_DISABLED", "") != "1"
	)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load default config: %w", err)
	}

	if len(awsBaseURL) > 0 {
		cfg.BaseEndpoint = aws.String(awsBaseURL)
	}

	dynamoClient, err := dynamo.NewClient(cfg, tableName)
	if err != nil {
		return fmt.Errorf("failed to create dynamodb client: %w", err)
	}

	if event.isS3Event() {
		s3Client := s3.NewClient(cfg, evidenceBucketName)
		documentStore := app.NewDocumentStore(dynamoClient, nil, nil, nil, nil)

		if err := handleObjectTagsAdded(ctx, dynamoClient, event.S3Event, s3Client, documentStore); err != nil {
			return fmt.Errorf("ObjectTagging:Put: %w", err)
		}

		return nil
	}

	if event.isCloudWatchEvent() {
		factory := &Factory{
			now:                   time.Now,
			cfg:                   cfg,
			dynamoClient:          dynamoClient,
			appPublicURL:          appPublicURL,
			lpaStoreBaseURL:       lpaStoreBaseURL,
			uidBaseURL:            uidBaseURL,
			notifyBaseURL:         notifyBaseURL,
			notifyIsProduction:    notifyIsProduction,
			eventBusName:          eventBusName,
			searchEndpoint:        searchEndpoint,
			searchIndexingEnabled: searchIndexingEnabled,
		}

		handler := &cloudWatchEventHandler{
			dynamoClient: dynamoClient,
			now:          time.Now,
			factory:      factory,
		}

		if err := handler.Handle(ctx, event.CloudWatchEvent); err != nil {
			return fmt.Errorf("%s: %w", event.DetailType, err)
		}

		log.Println("successfully handled ", event.DetailType)
		return nil
	}

	eJson, _ := json.Marshal(event)
	return fmt.Errorf("unknown event type received: %s", string(eJson))
}

func main() {
	lambda.Start(handler)
}
