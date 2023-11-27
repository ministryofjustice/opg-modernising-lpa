package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/app"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/s3"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

const (
	virusFound               = "infected"
	objectTagsAddedEventName = "ObjectTagging:Put"
)

type uidEvent struct {
	UID string `json:"uid"`
}

//go:generate mockery --testonly --inpackage --name dynamodbClient --structname mockDynamodbClient
type dynamodbClient interface {
	One(ctx context.Context, pk, sk string, v interface{}) error
	OneByUID(ctx context.Context, uid string, v interface{}) error
	Put(ctx context.Context, v interface{}) error
	Update(ctx context.Context, pk, sk string, values map[string]dynamodbtypes.AttributeValue, expression string) error
}

//go:generate mockery --testonly --inpackage --name s3Client --structname mockS3Client
type s3Client interface {
	GetObjectTags(ctx context.Context, key string) ([]types.Tag, error)
}

//go:generate mockery --testonly --inpackage --name shareCodeSender --structname mockShareCodeSender
type shareCodeSender interface {
	SendCertificateProvider(context.Context, notify.Template, page.AppData, *actor.DonorProvidedDetails) error
}

//go:generate mockery --testonly --inpackage --name DocumentStore --structname mockDocumentStore
type DocumentStore interface {
	UpdateScanResults(ctx context.Context, lpaID, objectKey string, virusDetected bool) error
}

//go:generate mockery --testonly --inpackage --name UidStore --structname mockUidStore
type UidStore interface {
	Set(ctx context.Context, lpaID, sessionID, uid string) error
}

//go:generate mockery --testonly --inpackage --name UidClient --structname mockUidClient
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

	now := time.Now

	if event.isS3Event() {
		s3Client := s3.NewClient(cfg, evidenceBucketName)
		documentStore := app.NewDocumentStore(dynamoClient, nil, nil, nil, nil)

		if err := handleObjectTagsAdded(ctx, dynamoClient, event.S3Event, s3Client, documentStore); err != nil {
			return fmt.Errorf("ObjectTagging:Put: %w", err)
		}

		return nil
	}

	if event.isCloudWatchEvent() {
		err := fmt.Errorf("unknown cloudwatch event")

		switch event.DetailType {
		case "uid-requested":
			uidStore := app.NewUidStore(dynamoClient, now)
			uidClient := uid.New(uidBaseURL, &http.Client{Timeout: 10 * time.Second}, cfg, v4.NewSigner(), time.Now)

			err = handleUidRequested(ctx, uidStore, uidClient, event.CloudWatchEvent)

		case "evidence-received":
			err = handleEvidenceReceived(ctx, dynamoClient, event.CloudWatchEvent)

		case "reduced-fee-approved":
			bundle := localize.NewBundle("./lang/en.json", "./lang/cy.json")

			//TODO do this in handleFeeApproved when/if we save lang preference in LPA
			appData := page.AppData{Localizer: bundle.For(localize.En)}

			secretsClient, err := secrets.NewClient(cfg, time.Hour)
			if err != nil {
				return fmt.Errorf("failed to create secrets client: %w", err)
			}

			notifyApiKey, err := secretsClient.Secret(ctx, secrets.GovUkNotify)
			if err != nil {
				return fmt.Errorf("failed to get notify API secret: %w", err)
			}

			notifyClient, err := notify.New(notifyIsProduction, notifyBaseURL, notifyApiKey, http.DefaultClient)
			if err != nil {
				return err
			}

			shareCodeSender := page.NewShareCodeSender(app.NewShareCodeStore(dynamoClient), notifyClient, appPublicURL, random.String)

			err = handleFeeApproved(ctx, dynamoClient, event.CloudWatchEvent, shareCodeSender, appData, now)

		case "reduced-fee-declined":
			err = handleFeeDenied(ctx, dynamoClient, event.CloudWatchEvent, now)

		case "move-evidence-required":
			err = handleMoreEvidenceRequired(ctx, dynamoClient, event.CloudWatchEvent, now)
		}

		if err != nil {
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
