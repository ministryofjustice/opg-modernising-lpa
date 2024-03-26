package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type factory interface {
	AppData() (page.AppData, error)
	ShareCodeSender(ctx context.Context) (ShareCodeSender, error)
	LpaStoreClient() (LpaStoreClient, error)
	UidStore() (UidStore, error)
	UidClient() UidClient
}

type cloudWatchEventHandler struct {
	dynamoClient dynamodbClient
	now          func() time.Time
	factory      factory
}

func (h *cloudWatchEventHandler) Handle(ctx context.Context, cloudWatchEvent events.CloudWatchEvent) error {
	switch cloudWatchEvent.DetailType {
	case "uid-requested":
		uidStore, err := h.factory.UidStore()
		if err != nil {
			return err
		}

		uidClient := h.factory.UidClient()

		return handleUidRequested(ctx, uidStore, uidClient, cloudWatchEvent)

	case "evidence-received":
		return handleEvidenceReceived(ctx, h.dynamoClient, cloudWatchEvent)

	case "reduced-fee-approved":
		appData, err := h.factory.AppData()
		if err != nil {
			return err
		}

		shareCodeSender, err := h.factory.ShareCodeSender(ctx)
		if err != nil {
			return err
		}

		lpaStoreClient, err := h.factory.LpaStoreClient()
		if err != nil {
			return err
		}

		return handleFeeApproved(ctx, h.dynamoClient, cloudWatchEvent, shareCodeSender, lpaStoreClient, appData, h.now)

	case "reduced-fee-declined":
		return handleFeeDenied(ctx, h.dynamoClient, cloudWatchEvent, h.now)

	case "more-evidence-required":
		return handleMoreEvidenceRequired(ctx, h.dynamoClient, cloudWatchEvent, h.now)

	case "donor-submission-completed":
		appData, err := h.factory.AppData()
		if err != nil {
			return err
		}

		shareCodeSender, err := h.factory.ShareCodeSender(ctx)
		if err != nil {
			return err
		}

		lpaStoreClient, err := h.factory.LpaStoreClient()
		if err != nil {
			return err
		}

		return handleDonorSubmissionCompleted(ctx, h.dynamoClient, cloudWatchEvent, shareCodeSender, appData, lpaStoreClient)

	case "certificate-provider-submission-completed":
		return handleCertificateProviderSubmissionCompleted(ctx, cloudWatchEvent, h.factory)

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

	if err := uidStore.Set(ctx, v.LpaID, v.DonorSessionID, v.OrganisationID, uid); err != nil {
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

func handleFeeApproved(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent, shareCodeSender ShareCodeSender, lpaStoreClient LpaStoreClient, appData page.AppData, now func() time.Time) error {
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

	if err := shareCodeSender.SendCertificateProviderPrompt(ctx, appData, donor); err != nil {
		return fmt.Errorf("failed to send share code to certificate provider: %w", err)
	}

	if err := lpaStoreClient.SendLpa(ctx, donor); err != nil {
		return fmt.Errorf("failed to send to lpastore: %w", err)
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

func handleDonorSubmissionCompleted(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent, shareCodeSender ShareCodeSender, appData page.AppData, lpaStoreClient LpaStoreClient) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	var key dynamo.Key
	if err := client.OneByUID(ctx, v.UID, &key); !errors.Is(err, dynamo.NotFoundError{}) {
		return err
	}

	donor, err := lpaStoreClient.Lpa(ctx, v.UID)
	if err != nil {
		return err
	}

	if donor.CertificateProvider.CarryOutBy.IsOnline() {
		if err := shareCodeSender.SendCertificateProviderInvite(ctx, appData, donor); err != nil {
			return fmt.Errorf("failed to send share code to certificate provider: %w", err)
		}
	}

	return nil
}

func handleCertificateProviderSubmissionCompleted(ctx context.Context, event events.CloudWatchEvent, factory factory) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	lpaStoreClient, err := factory.LpaStoreClient()
	if err != nil {
		return err
	}

	donor, err := lpaStoreClient.Lpa(ctx, v.UID)
	if err != nil {
		return fmt.Errorf("failed to retrieve lpa: %w", err)
	}

	if donor.CertificateProvider.CarryOutBy.IsPaper() {
		shareCodeSender, err := factory.ShareCodeSender(ctx)
		if err != nil {
			return err
		}

		appData, err := factory.AppData()
		if err != nil {
			return err
		}

		if err := shareCodeSender.SendAttorneys(ctx, appData, donor); err != nil {
			return fmt.Errorf("failed to send share codes to attorneys: %w", err)
		}
	}

	return nil
}
