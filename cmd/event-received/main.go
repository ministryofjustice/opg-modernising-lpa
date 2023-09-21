package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/app"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
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

//go:generate mockery --testonly --inpackage --name shareCodeSender --structname mockShareCodeSender
type shareCodeSender interface {
	SendCertificateProvider(context.Context, notify.Template, page.AppData, bool, *page.Lpa) error
}

func Handler(ctx context.Context, event events.CloudWatchEvent) error {
	tableName := os.Getenv("LPAS_TABLE")
	notifyIsProduction := os.Getenv("GOVUK_NOTIFY_IS_PRODUCTION") == "1"
	appPublicURL := os.Getenv("APP_PUBLIC_URL")
	awsBaseURL := os.Getenv("AWS_BASE_URL")
	notifyBaseURL := os.Getenv("GOVUK_NOTIFY_BASE_URL")

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

	secretsClient, err := secrets.NewClient(cfg, time.Hour)
	if err != nil {
		return fmt.Errorf("failed to create secrets client: %w", err)
	}

	notifyApiKey, err := secretsClient.Secret(ctx, secrets.GovUkNotify)
	if err != nil {
		return fmt.Errorf("failed to get notify API secret: %w", err)
	}

	notifyClient, err := notify.New(notifyIsProduction, notifyBaseURL, notifyApiKey, http.DefaultClient)

	bundle := localize.NewBundle("./lang/en.json", "./lang/cy.json")

	//TODO do this in handleFeeApproved when/if we save lang preference in LPA
	appData := page.AppData{Localizer: bundle.For(localize.En)}

	shareCodeSender := page.NewShareCodeSender(app.NewShareCodeStore(dynamoClient), notifyClient, appPublicURL, random.String)

	switch event.DetailType {
	case "evidence-received":
		return handleEvidenceReceived(ctx, dynamoClient, event)
	case "fee-approved":
		return handleFeeApproved(ctx, dynamoClient, event, shareCodeSender, appData)
	case "more-evidence-required":
		return handleMoreEvidenceRequired(ctx, dynamoClient, event)
	case "fee-denied":
		return handleFeeDenied(ctx, dynamoClient, event)
	default:
		return fmt.Errorf("unknown event received: %s", event.DetailType)
	}
}

func handleEvidenceReceived(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal 'evidence-received' detail: %w", err)
	}

	var key dynamo.Key
	if err := client.OneByUID(ctx, v.UID, &key); err != nil {
		return fmt.Errorf("failed to resolve uid for 'evidence-received': %w", err)
	}

	if key.PK == "" {
		return errors.New("PK missing from LPA in response to 'evidence-received'")
	}

	if err := client.Put(ctx, map[string]string{"PK": key.PK, "SK": "#EVIDENCE_RECEIVED"}); err != nil {
		return fmt.Errorf("failed to persist evidence received for 'evidence-received': %w", err)
	}

	return nil
}

func handleFeeApproved(ctx context.Context, dynamoClient dynamodbClient, event events.CloudWatchEvent, shareCodeSender shareCodeSender, appData page.AppData) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal 'fee-approved' detail: %w", err)
	}

	var key dynamo.Key
	if err := dynamoClient.OneByUID(ctx, v.UID, &key); err != nil {
		return fmt.Errorf("failed to resolve uid for 'fee-approved': %w", err)
	}

	var lpa page.Lpa
	if err := dynamoClient.One(ctx, key.PK, key.SK, &lpa); err != nil {
		return fmt.Errorf("failed to get LPA for 'fee-approved': %w", err)
	}

	lpa.Tasks.PayForLpa = actor.PaymentTaskCompleted

	if err := dynamoClient.Put(ctx, lpa); err != nil {
		return fmt.Errorf("failed to update LPA task status for 'fee-approved': %w", err)
	}

	if err := shareCodeSender.SendCertificateProvider(ctx, notify.CertificateProviderInviteEmail, appData, false, &lpa); err != nil {
		return fmt.Errorf("failed to send share code to certificate provider for 'fee-approved': %w", err)
	}

	return nil
}

func handleMoreEvidenceRequired(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal 'more-evidence-required' detail: %w", err)
	}

	var key dynamo.Key
	if err := client.OneByUID(ctx, v.UID, &key); err != nil {
		return fmt.Errorf("failed to resolve uid for 'more-evidence-required': %w", err)
	}

	if key.PK == "" {
		return errors.New("PK missing from LPA in response to 'more-evidence-required'")
	}

	var lpa page.Lpa
	if err := client.One(ctx, key.PK, key.SK, &lpa); err != nil {
		return fmt.Errorf("failed to get LPA for 'more-evidence-required': %w", err)
	}

	lpa.Tasks.PayForLpa = actor.PaymentTaskMoreEvidenceRequired

	if err := client.Put(ctx, lpa); err != nil {
		return fmt.Errorf("failed to update LPA task status for 'more-evidence-required': %w", err)
	}

	return nil
}

func handleFeeDenied(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal 'fee-denied' detail: %w", err)
	}

	var key dynamo.Key
	if err := client.OneByUID(ctx, v.UID, &key); err != nil {
		return fmt.Errorf("failed to resolve uid for 'fee-denied': %w", err)
	}

	if key.PK == "" {
		return errors.New("PK missing from LPA in response to 'fee-denied'")
	}

	var lpa page.Lpa
	if err := client.One(ctx, key.PK, key.SK, &lpa); err != nil {
		return fmt.Errorf("failed to get LPA for 'fee-denied': %w", err)
	}

	lpa.Tasks.PayForLpa = actor.PaymentTaskDenied

	if err := client.Put(ctx, lpa); err != nil {
		return fmt.Errorf("failed to update LPA task status for 'fee-denied': %w", err)
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
