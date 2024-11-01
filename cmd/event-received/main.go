// Event received is an AWS Lambda function to handle incoming events.
package main

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/document"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/s3"
)

const (
	virusFound = "infected"
)

type factory interface {
	Now() func() time.Time
	DynamoClient() dynamodbClient
	UuidString() func() string
	AppData() (appcontext.Data, error)
	ShareCodeSender(ctx context.Context) (ShareCodeSender, error)
	LpaStoreClient() (LpaStoreClient, error)
	UidStore() (UidStore, error)
	UidClient() UidClient
	EventClient() EventClient
}

type Handler interface {
	Handle(context.Context, factory, *events.CloudWatchEvent) error
}

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
	CreateOnly(ctx context.Context, v interface{}) error
}

type s3Client interface {
	GetObjectTags(ctx context.Context, key string) ([]types.Tag, error)
}

type DocumentStore interface {
	UpdateScanResults(ctx context.Context, lpaID, objectKey string, virusDetected bool) error
}

type EventClient interface {
	SendApplicationUpdated(ctx context.Context, event event.ApplicationUpdated) error
	SendCertificateProviderStarted(ctx context.Context, event event.CertificateProviderStarted) error
}

type Event struct {
	S3Event  *events.S3Event
	SQSEvent *events.SQSEvent
}

func (e *Event) UnmarshalJSON(data []byte) error {
	var s3 events.S3Event
	if err := json.Unmarshal(data, &s3); err == nil && len(s3.Records) > 0 && s3.Records[0].S3.Bucket.Name != "" {
		e.S3Event = &s3
		return nil
	}

	var sqs events.SQSEvent
	if err := json.Unmarshal(data, &sqs); err == nil && len(sqs.Records) > 0 && sqs.Records[0].MessageId != "" {
		e.SQSEvent = &sqs
		return nil
	}

	return errors.New("unknown event type")
}

func handler(ctx context.Context, event Event) (map[string]any, error) {
	var (
		tableName             = os.Getenv("LPAS_TABLE")
		notifyIsProduction    = os.Getenv("GOVUK_NOTIFY_IS_PRODUCTION") == "1"
		appPublicURL          = os.Getenv("APP_PUBLIC_URL")
		awsBaseURL            = os.Getenv("AWS_BASE_URL")
		notifyBaseURL         = os.Getenv("GOVUK_NOTIFY_BASE_URL")
		evidenceBucketName    = os.Getenv("UPLOADS_S3_BUCKET_NAME")
		uidBaseURL            = os.Getenv("UID_BASE_URL")
		lpaStoreBaseURL       = os.Getenv("LPA_STORE_BASE_URL")
		lpaStoreSecretARN     = os.Getenv("LPA_STORE_SECRET_ARN")
		eventBusName          = cmp.Or(os.Getenv("EVENT_BUS_NAME"), "default")
		searchEndpoint        = os.Getenv("SEARCH_ENDPOINT")
		searchIndexName       = cmp.Or(os.Getenv("SEARCH_INDEX_NAME"), "lpas")
		searchIndexingEnabled = os.Getenv("SEARCH_INDEXING_DISABLED") != "1"
	)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil).
		WithAttrs([]slog.Attr{
			slog.String("service_name", "opg-modernising-lpa/event-received"),
		}))

	result := map[string]any{}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return result, fmt.Errorf("failed to load default config: %w", err)
	}

	if len(awsBaseURL) > 0 {
		cfg.BaseEndpoint = aws.String(awsBaseURL)
	}

	dynamoClient, err := dynamo.NewClient(cfg, tableName)
	if err != nil {
		return result, fmt.Errorf("failed to create dynamodb client: %w", err)
	}

	if event.S3Event != nil {
		s3Client := s3.NewClient(cfg, evidenceBucketName)
		documentStore := document.NewStore(dynamoClient, nil, nil)

		if err := handleObjectTagsAdded(ctx, dynamoClient, event.S3Event, s3Client, documentStore); err != nil {
			return result, fmt.Errorf("ObjectTagging:Put: %w", err)
		}

		return result, nil
	}

	factory := &Factory{
		logger:                logger,
		now:                   time.Now,
		uuidString:            random.UuidString,
		cfg:                   cfg,
		dynamoClient:          dynamoClient,
		appPublicURL:          appPublicURL,
		lpaStoreBaseURL:       lpaStoreBaseURL,
		lpaStoreSecretARN:     lpaStoreSecretARN,
		uidBaseURL:            uidBaseURL,
		notifyBaseURL:         notifyBaseURL,
		notifyIsProduction:    notifyIsProduction,
		eventBusName:          eventBusName,
		searchEndpoint:        searchEndpoint,
		searchIndexName:       searchIndexName,
		searchIndexingEnabled: searchIndexingEnabled,
	}

	if event.SQSEvent != nil {
		batchItemFailures := []map[string]any{}
		for _, record := range event.SQSEvent.Records {
			var cloud *events.CloudWatchEvent
			if err := json.Unmarshal([]byte(record.Body), &cloud); err != nil {
				logger.ErrorContext(ctx, "could not unmarshal event", slog.String("messageID", record.MessageId), slog.Any("err", err))
				batchItemFailures = append(batchItemFailures, map[string]any{"itemIdentifier": record.MessageId})
				continue
			}

			if err := handleCloudWatchEvent(ctx, factory, cloud); err != nil {
				logger.ErrorContext(ctx, "error processing event", slog.String("messageID", record.MessageId), slog.Any("err", err))
				batchItemFailures = append(batchItemFailures, map[string]any{"itemIdentifier": record.MessageId})
				continue
			}
		}

		result["batchItemFailures"] = batchItemFailures
		return result, nil
	}

	return result, nil
}

func handleCloudWatchEvent(ctx context.Context, factory *Factory, event *events.CloudWatchEvent) error {
	var handler Handler
	switch event.Source {
	case "opg.poas.sirius":
		handler = &siriusEventHandler{}
	case "opg.poas.makeregister":
		handler = &makeregisterEventHandler{}
	case "opg.poas.lpastore":
		handler = &lpastoreEventHandler{}
	}

	if handler == nil {
		eJson, _ := json.Marshal(event)
		return fmt.Errorf("unknown event received: %s", string(eJson))
	}

	if err := handler.Handle(ctx, factory, event); err != nil {
		return fmt.Errorf("%s: %w", event.DetailType, err)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
