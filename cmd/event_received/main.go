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
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/app"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
)

type evidenceReceivedEvent struct {
	UID string `json:"uid"`
}

//go:generate mockery --testonly --inpackage --name dynamodbClient --structname mockDynamodbClient
type dynamodbClient interface {
	Put(ctx context.Context, v interface{}) error
	GetOneByUID(context.Context, string, interface{}) error
}

//go:generate mockery --testonly --inpackage --name shareCodeSender --structname mockShareCodeSender
type shareCodeSender interface {
	SendCertificateProvider(context.Context, notify.Template, page.AppData, bool, *page.Lpa) error
}

func Handler(ctx context.Context, event events.CloudWatchEvent) error {
	tableName := os.Getenv("LPAS_TABLE")
	notifyIsProduction := os.Getenv("GOVUK_NOTIFY_IS_PRODUCTION") == "1"
	notifyBaseURL := os.Getenv("GOVUK_NOTIFY_BASE_URL")
	appPublicURL := os.Getenv("APP_PUBLIC_URL")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load default config: %w", err)
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

	bundle := localize.NewBundle("../lang/en.json", "../lang/cy.json")

	//TODO do this in handleFeeApproved when/if we save lang preference in LPA
	appData := page.AppData{Localizer: bundle.For(localize.En)}

	shareCodeSender := page.NewShareCodeSender(app.NewShareCodeStore(dynamoClient), notifyClient, appPublicURL, random.String)

	switch event.DetailType {
	case "evidence-received":
		return handleEvidenceReceived(ctx, dynamoClient, event)
	case "fee-approved":
		return handleFeeApproved(ctx, dynamoClient, event, shareCodeSender, appData)
	default:
		return fmt.Errorf("unknown event received: %s", event.DetailType)
	}
}

func handleEvidenceReceived(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent) error {
	var v evidenceReceivedEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal 'evidence-received' detail: %w", err)
	}

	var lpa page.Lpa
	err := client.GetOneByUID(ctx, v.UID, &lpa)
	if err != nil {
		return fmt.Errorf("failed to resolve uid for 'evidence-received': %w", err)
	}

	log.Println(lpa)

	item, err := attributevalue.MarshalMap(map[string]any{"PK": lpa.PK, "SK": "#EVIDENCE_RECEIVED"})
	if err != nil {
		return fmt.Errorf("failed to marshal item in response to 'evidence-received': %w", err)
	}

	if err := client.Put(ctx, &dynamodb.PutItemInput{Item: item}); err != nil {
		return fmt.Errorf("failed to persist evidence received for 'evidence-received': %w", err)
	}

	return nil
}

func handleFeeApproved(ctx context.Context, dynamoClient dynamodbClient, event events.CloudWatchEvent, shareCodeSender shareCodeSender, appData page.AppData) error {
	var v evidenceReceivedEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal 'fee-approved' detail: %w", err)
	}

	var lpa page.Lpa
	err := dynamoClient.GetOneByUID(ctx, v.UID, &lpa)
	if err != nil {
		return fmt.Errorf("failed to resolve uid for 'fee-approved': %w", err)
	}

	lpa.Tasks.PayForLpa = actor.PaymentTaskApproved

	if err := dynamoClient.Put(ctx, lpa); err != nil {
		return fmt.Errorf("failed to update LPA task status for 'fee-approved': %w", err)
	}

	if err := shareCodeSender.SendCertificateProvider(ctx, notify.CertificateProviderInviteEmail, appData, false, &lpa); err != nil {
		return fmt.Errorf("failed to send share code to certificate provider for 'fee-approved': %w", err)
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
