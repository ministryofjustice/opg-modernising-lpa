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
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/app"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
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
	SendCertificateProvider(context.Context, notify.Template, page.AppData, bool, *page.Lpa) error
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
	return e.Source == "aws.cloudwatch"
}

func Handler(ctx context.Context, event Event) error {
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
		documentStore := app.NewDocumentStore(dynamoClient, s3Client, random.UuidString)

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

		case "fee-approved":
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

		case "fee-denined":
			err = handleFeeDenied(ctx, dynamoClient, event.CloudWatchEvent, now)

		case "move-evidence-required":
			err = handleMoreEvidenceRequired(ctx, dynamoClient, event.CloudWatchEvent, now)
		}

		if err != nil {
			return fmt.Errorf("%s: %w", event.DetailType, err)
		}

		return nil
	}

	eJson, _ := json.Marshal(event)
	return fmt.Errorf("unknown event type received: %s", string(eJson))
}

func handleUidRequested(ctx context.Context, uidStore UidStore, uidClient UidClient, e events.CloudWatchEvent) error {
	var v event.UidRequested
	if err := json.Unmarshal(e.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	uid, err := uidClient.CreateCase(ctx, &uid.CreateCaseRequestBody{Type: v.Type, Donor: v.Donor})
	if err != nil {
		return fmt.Errorf("failed to create case: %w", err)
	}

	if err := uidStore.Set(ctx, v.ID, v.DonorSessionID, uid); err != nil {
		return fmt.Errorf("failed to set uid: %w", err)
	}

	return nil
}

func handleEvidenceReceived(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	var key dynamo.Key
	if err := client.OneByUID(ctx, v.UID, &key); err != nil {
		return fmt.Errorf("failed to resolve uid: %w", err)
	}

	if key.PK == "" {
		return fmt.Errorf("PK missing from LPA in response")
	}

	if err := client.Put(ctx, map[string]string{"PK": key.PK, "SK": "#EVIDENCE_RECEIVED"}); err != nil {
		return fmt.Errorf("failed to persist evidence received: %w", err)
	}

	return nil
}

func handleFeeApproved(ctx context.Context, dynamoClient dynamodbClient, event events.CloudWatchEvent, shareCodeSender shareCodeSender, appData page.AppData, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	lpa, err := getLpaByUID(ctx, dynamoClient, v.UID)
	if err != nil {
		return err
	}

	lpa.Tasks.PayForLpa = actor.PaymentTaskCompleted
	lpa.UpdatedAt = now()

	if err := dynamoClient.Put(ctx, lpa); err != nil {
		return fmt.Errorf("failed to update LPA task status: %w", err)
	}

	if err := shareCodeSender.SendCertificateProvider(ctx, notify.CertificateProviderInviteEmail, appData, false, &lpa); err != nil {
		return fmt.Errorf("failed to send share code to certificate provider: %w", err)
	}

	return nil
}

func handleMoreEvidenceRequired(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	lpa, err := getLpaByUID(ctx, client, v.UID)
	if err != nil {
		return err
	}

	lpa.Tasks.PayForLpa = actor.PaymentTaskMoreEvidenceRequired
	lpa.UpdatedAt = now()

	if err := client.Put(ctx, lpa); err != nil {
		return fmt.Errorf("failed to update LPA task status: %w", err)
	}

	return nil
}

func handleFeeDenied(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	lpa, err := getLpaByUID(ctx, client, v.UID)
	if err != nil {
		return err
	}

	lpa.Tasks.PayForLpa = actor.PaymentTaskDenied
	lpa.UpdatedAt = now()

	if err := client.Put(ctx, lpa); err != nil {
		return fmt.Errorf("failed to update LPA task status: %w", err)
	}

	return nil
}

func handleObjectTagsAdded(ctx context.Context, dynamodbClient dynamodbClient, event events.S3Event, s3Client s3Client, documentStore DocumentStore) error {
	objectKey := event.Records[0].S3.Object.Key
	if objectKey == "" {
		return fmt.Errorf("object key missing")
	}

	tags, err := s3Client.GetObjectTags(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("failed to get tags for object: %w", err)
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

	lpa, err := getLpaByUID(ctx, dynamodbClient, parts[0])
	if err != nil {
		return err
	}

	err = documentStore.UpdateScanResults(ctx, lpa.ID, objectKey, hasVirus)
	if err != nil {
		return fmt.Errorf("failed to update scan results: %w", err)
	}

	return nil
}

func getLpaByUID(ctx context.Context, client dynamodbClient, uid string) (page.Lpa, error) {
	var key dynamo.Key
	if err := client.OneByUID(ctx, uid, &key); err != nil {
		return page.Lpa{}, fmt.Errorf("failed to resolve uid: %w", err)
	}

	if key.PK == "" {
		return page.Lpa{}, fmt.Errorf("PK missing from LPA in response")
	}

	var lpa page.Lpa
	if err := client.One(ctx, key.PK, key.SK, &lpa); err != nil {
		return page.Lpa{}, fmt.Errorf("failed to get LPA: %w", err)
	}

	return lpa, nil
}

func main() {
	lambda.Start(Handler)
}
