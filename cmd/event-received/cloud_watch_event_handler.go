package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/app"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lambda"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type cloudWatchEventHandler struct {
	dynamoClient       dynamodbClient
	now                func() time.Time
	uidBaseURL         string
	cfg                aws.Config
	notifyIsProduction bool
	notifyBaseURL      string
	appPublicURL       string
}

func (h *cloudWatchEventHandler) Handle(ctx context.Context, event events.CloudWatchEvent) error {
	switch event.DetailType {
	case "uid-requested":
		uidStore := app.NewUidStore(h.dynamoClient, h.now)
		uidClient := uid.New(h.uidBaseURL, lambda.New(h.cfg, v4.NewSigner(), &http.Client{Timeout: 10 * time.Second}, time.Now))

		return handleUidRequested(ctx, uidStore, uidClient, event)

	case "evidence-received":
		return handleEvidenceReceived(ctx, h.dynamoClient, event)

	case "reduced-fee-approved":
		bundle, _ := localize.NewBundle("./lang/en.json", "./lang/cy.json")

		//TODO do this in handleFeeApproved when/if we save lang preference in LPA
		appData := page.AppData{Localizer: bundle.For(localize.En)}

		secretsClient, err := secrets.NewClient(h.cfg, time.Hour)
		if err != nil {
			return fmt.Errorf("failed to create secrets client: %w", err)
		}

		notifyApiKey, err := secretsClient.Secret(ctx, secrets.GovUkNotify)
		if err != nil {
			return fmt.Errorf("failed to get notify API secret: %w", err)
		}

		notifyClient, err := notify.New(h.notifyIsProduction, h.notifyBaseURL, notifyApiKey, http.DefaultClient)
		if err != nil {
			return err
		}

		shareCodeSender := page.NewShareCodeSender(app.NewShareCodeStore(h.dynamoClient), notifyClient, h.appPublicURL, random.String)

		return handleFeeApproved(ctx, h.dynamoClient, event, shareCodeSender, appData, h.now)

	case "reduced-fee-declined":
		return handleFeeDenied(ctx, h.dynamoClient, event, h.now)

	case "more-evidence-required":
		return handleMoreEvidenceRequired(ctx, h.dynamoClient, event, h.now)

	default:
		return fmt.Errorf("unknown cloudwatch event")
	}
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

	if err := uidStore.Set(ctx, v.LpaID, v.DonorSessionID, uid); err != nil {
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

func handleFeeApproved(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent, shareCodeSender shareCodeSender, appData page.AppData, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	donor, err := getDonorByLpaUID(ctx, client, v.UID)
	if err != nil {
		return err
	}

	donor.Tasks.PayForLpa = actor.PaymentTaskCompleted

	if err := putDonor(ctx, donor, now, client); err != nil {
		return fmt.Errorf("failed to update LPA task status: %w", err)
	}

	if err := shareCodeSender.SendCertificateProviderPrompt(ctx, appData, &donor); err != nil {
		return fmt.Errorf("failed to send share code to certificate provider: %w", err)
	}

	return nil
}

func handleMoreEvidenceRequired(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	donor, err := getDonorByLpaUID(ctx, client, v.UID)
	if err != nil {
		return err
	}

	donor.Tasks.PayForLpa = actor.PaymentTaskMoreEvidenceRequired

	if err := putDonor(ctx, donor, now, client); err != nil {
		return fmt.Errorf("failed to update LPA task status: %w", err)
	}

	return nil
}

func handleFeeDenied(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	donor, err := getDonorByLpaUID(ctx, client, v.UID)
	if err != nil {
		return err
	}

	donor.Tasks.PayForLpa = actor.PaymentTaskDenied

	if err := putDonor(ctx, donor, now, client); err != nil {
		return fmt.Errorf("failed to update LPA task status: %w", err)
	}

	return nil
}
