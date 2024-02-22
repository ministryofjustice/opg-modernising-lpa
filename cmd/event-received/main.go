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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/app"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/s3"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

const (
	virusFound = "infected"
)

type uidEvent struct {
	UID string `json:"uid"`
}

type dynamodbClient interface {
	One(ctx context.Context, pk, sk string, v interface{}) error
	OneByUID(ctx context.Context, uid string, v interface{}) error
	Put(ctx context.Context, v interface{}) error
	UpdateReturn(ctx context.Context, pk, sk string, values map[string]dynamodbtypes.AttributeValue, expression string) (map[string]dynamodbtypes.AttributeValue, error)
	DeleteOne(ctx context.Context, pk, sk string) error
}

type s3Client interface {
	GetObjectTags(ctx context.Context, key string) ([]types.Tag, error)
}

type shareCodeSender interface {
	SendCertificateProviderPrompt(context.Context, page.AppData, *actor.DonorProvidedDetails) error
}

type DocumentStore interface {
	UpdateScanResults(ctx context.Context, lpaID, objectKey string, virusDetected bool) error
}

type UidStore interface {
	Set(ctx context.Context, lpaID, sessionID, organisationID, uid string) error
}

type UidClient interface {
	CreateCase(context.Context, *uid.CreateCaseRequestBody) (string, error)
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
		tableName          = os.Getenv("LPAS_TABLE")
		notifyIsProduction = os.Getenv("GOVUK_NOTIFY_IS_PRODUCTION") == "1"
		appPublicURL       = os.Getenv("APP_PUBLIC_URL")
		awsBaseURL         = os.Getenv("AWS_BASE_URL")
		notifyBaseURL      = os.Getenv("GOVUK_NOTIFY_BASE_URL")
		evidenceBucketName = os.Getenv("UPLOADS_S3_BUCKET_NAME")
		uidBaseURL         = os.Getenv("UID_BASE_URL")
		eventBusName       = env.Get("EVENT_BUS_NAME", "default")
	)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load default config: %w", err)
	}

	if len(awsBaseURL) > 0 {
		cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           awsBaseURL,
				SigningRegion: "eu-west-1",
			}, nil
		})
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
		handler := &cloudWatchEventHandler{
			dynamoClient:       dynamoClient,
			now:                time.Now,
			uidBaseURL:         uidBaseURL,
			cfg:                cfg,
			notifyIsProduction: notifyIsProduction,
			notifyBaseURL:      notifyBaseURL,
			appPublicURL:       appPublicURL,
			eventBusName:       eventBusName,
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
