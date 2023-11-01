package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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
)

const (
	virusFound                    = "infected"
	evidenceReceivedEventName     = "evidence-received"
	feeApprovedEventName          = "fee-approved"
	feeDeniedEventName            = "fee-denied"
	moreEvidenceRequiredEventName = "more-evidence-required"
	objectTagsAddedEventName      = "ObjectTagging:Put"
)

type uidEvent struct {
	UID string `json:"uid"`
}

//go:generate mockery --testonly --inpackage --name dynamodbClient --structname mockDynamodbClient
type dynamodbClient interface {
	One(ctx context.Context, pk, sk string, v interface{}) error
	OneByUID(ctx context.Context, uid string, v interface{}) error
	Put(ctx context.Context, v interface{}) error
}

//go:generate mockery --testonly --inpackage --name s3Client --structname mockS3Client
type s3Client interface {
	GetObjectTags(ctx context.Context, key string) ([]types.Tag, error)
}

//go:generate mockery --testonly --inpackage --name shareCodeSender --structname mockShareCodeSender
type shareCodeSender interface {
	SendCertificateProvider(context.Context, notify.Template, page.AppData, bool, *page.Lpa) error
}

//go:generate mockery --testonly --inpackage --name DocumentStore --structname mockDocumentStore
type DocumentStore interface {
	UpdateScanResults(ctx context.Context, PK, SK string, virusDetected bool) error
}

type Event struct {
	events.S3Event
	events.CloudWatchEvent
}

func (e Event) isS3Event() bool {
	return len(e.Records) > 0
}

func (e Event) isCloudWatchEvent() bool {
	return e.Source == "aws.cloudwatch"
}

func Handler(ctx context.Context, event Event) error {
	tableName := os.Getenv("LPAS_TABLE")
	notifyIsProduction := os.Getenv("GOVUK_NOTIFY_IS_PRODUCTION") == "1"
	appPublicURL := os.Getenv("APP_PUBLIC_URL")
	awsBaseURL := os.Getenv("AWS_BASE_URL")
	notifyBaseURL := os.Getenv("GOVUK_NOTIFY_BASE_URL")
	evidenceBucketName := os.Getenv("UPLOADS_S3_BUCKET_NAME")

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

	s3Client := s3.NewClient(cfg, evidenceBucketName)

	dynamoClient, err := dynamo.NewClient(cfg, tableName)
	if err != nil {
		return fmt.Errorf("failed to create dynamodb client: %w", err)
	}

	secretsClient, err := secrets.NewClient(cfg, time.Hour)
	if err != nil {
		return fmt.Errorf("failed to create secrets client: %w", err)
	}

	notifyApiKey, err := secretsClient.Secret(ctx, secrets.GovUkNotify)
	if err != nil {
		return fmt.Errorf("failed to get notify API secret: %w", err)
	}

	notifyClient, err := notify.New(notifyIsProduction, notifyBaseURL, notifyApiKey, http.DefaultClient)

	documentStore := app.NewDocumentStore(dynamoClient, s3Client)

	bundle := localize.NewBundle("./lang/en.json", "./lang/cy.json")

	//TODO do this in handleFeeApproved when/if we save lang preference in LPA
	appData := page.AppData{Localizer: bundle.For(localize.En)}

	shareCodeSender := page.NewShareCodeSender(app.NewShareCodeStore(dynamoClient), notifyClient, appPublicURL, random.String)
	now := time.Now

	if event.isS3Event() {
		return handleObjectTagsAdded(ctx, dynamoClient, event.S3Event, s3Client, documentStore)
	}

	if event.isCloudWatchEvent() {
		switch event.DetailType {
		case evidenceReceivedEventName:
			return handleEvidenceReceived(ctx, dynamoClient, event.CloudWatchEvent)
		case feeApprovedEventName:
			return handleFeeApproved(ctx, dynamoClient, event.CloudWatchEvent, shareCodeSender, appData, now)
		case moreEvidenceRequiredEventName:
			return handleMoreEvidenceRequired(ctx, dynamoClient, event.CloudWatchEvent, now)
		case feeDeniedEventName:
			return handleFeeDenied(ctx, dynamoClient, event.CloudWatchEvent, now)
		default:
			return fmt.Errorf("unknown event received: %s", event.DetailType)
		}
	}

	eJson, _ := json.Marshal(event)
	return fmt.Errorf("unknown event type received: %s", string(eJson))
}

func handleEvidenceReceived(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal '%s' detail: %w", evidenceReceivedEventName, err)
	}

	var key dynamo.Key
	if err := client.OneByUID(ctx, v.UID, &key); err != nil {
		return fmt.Errorf("failed to resolve uid for '%s': %w", evidenceReceivedEventName, err)
	}

	if key.PK == "" {
		return fmt.Errorf("PK missing from LPA in response to '%s'", evidenceReceivedEventName)
	}

	if err := client.Put(ctx, map[string]string{"PK": key.PK, "SK": "#EVIDENCE_RECEIVED"}); err != nil {
		return fmt.Errorf("failed to persist evidence received for '%s': %w", evidenceReceivedEventName, err)
	}

	return nil
}

func handleFeeApproved(ctx context.Context, dynamoClient dynamodbClient, event events.CloudWatchEvent, shareCodeSender shareCodeSender, appData page.AppData, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal '%s' detail: %w", feeApprovedEventName, err)
	}

	lpa, err := getLpaByUID(ctx, dynamoClient, v.UID, feeApprovedEventName)
	if err != nil {
		return err
	}

	lpa.Tasks.PayForLpa = actor.PaymentTaskCompleted
	lpa.UpdatedAt = now()

	if err := dynamoClient.Put(ctx, lpa); err != nil {
		return fmt.Errorf("failed to update LPA task status for '%s': %w", feeApprovedEventName, err)
	}

	if err := shareCodeSender.SendCertificateProvider(ctx, notify.CertificateProviderInviteEmail, appData, false, &lpa); err != nil {
		return fmt.Errorf("failed to send share code to certificate provider for '%s': %w", feeApprovedEventName, err)
	}

	return nil
}

func handleMoreEvidenceRequired(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal '%s' detail: %w", moreEvidenceRequiredEventName, err)
	}

	lpa, err := getLpaByUID(ctx, client, v.UID, moreEvidenceRequiredEventName)
	if err != nil {
		return err
	}

	lpa.Tasks.PayForLpa = actor.PaymentTaskMoreEvidenceRequired
	lpa.UpdatedAt = now()

	if err := client.Put(ctx, lpa); err != nil {
		return fmt.Errorf("failed to update LPA task status for '%s': %w", moreEvidenceRequiredEventName, err)
	}

	return nil
}

func handleFeeDenied(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal '%s' detail: %w", feeDeniedEventName, err)
	}

	lpa, err := getLpaByUID(ctx, client, v.UID, feeDeniedEventName)
	if err != nil {
		return err
	}

	lpa.Tasks.PayForLpa = actor.PaymentTaskDenied
	lpa.UpdatedAt = now()

	if err := client.Put(ctx, lpa); err != nil {
		return fmt.Errorf("failed to update LPA task status for '%s': %w", feeDeniedEventName, err)
	}

	return nil
}

func handleObjectTagsAdded(ctx context.Context, dynamodbClient dynamodbClient, event events.S3Event, s3Client s3Client, documentStore DocumentStore) error {
	objectKey := event.Records[0].S3.Object.Key
	if objectKey == "" {
		return fmt.Errorf("object key missing in event in '%s'", objectTagsAddedEventName)
	}

	tags, err := s3Client.GetObjectTags(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("failed to get tags for object in '%s': %w", objectTagsAddedEventName, err)
	}

	hasScannedTag := false
	hasVirus := false

	for _, tag := range tags {
		if *tag.Key == "virus-scan-status" {
			hasScannedTag = true
			hasVirus = *tag.Value == virusFound
			break
		}
	}

	if !hasScannedTag {
		return nil
	}

	parts := strings.Split(objectKey, "/")

	var lpaKey dynamo.Key
	if err := dynamodbClient.OneByUID(ctx, parts[0], &lpaKey); err != nil {
		return fmt.Errorf("failed to resolve uid for '%s': %w", objectTagsAddedEventName, err)
	}

	err = documentStore.UpdateScanResults(ctx, lpaKey.PK, "#DOCUMENT#"+objectKey, hasVirus)
	if err != nil {
		return fmt.Errorf("failed to update scan results for '%s': %w", objectTagsAddedEventName, err)
	}

	return nil
}

func getLpaByUID(ctx context.Context, client dynamodbClient, uid, eventName string) (page.Lpa, error) {
	var key dynamo.Key
	if err := client.OneByUID(ctx, uid, &key); err != nil {
		return page.Lpa{}, fmt.Errorf("failed to resolve uid for '%s': %w", eventName, err)
	}

	if key.PK == "" {
		return page.Lpa{}, fmt.Errorf("PK missing from LPA in response to '%s'", eventName)
	}

	var lpa page.Lpa
	if err := client.One(ctx, key.PK, key.SK, &lpa); err != nil {
		return page.Lpa{}, fmt.Errorf("failed to get LPA for '%s': %w", eventName, err)
	}

	return lpa, nil
}

func main() {
	lambda.Start(Handler)
}
