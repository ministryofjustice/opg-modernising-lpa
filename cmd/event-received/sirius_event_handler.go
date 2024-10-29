package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

type siriusEventHandler struct{}

func (h *siriusEventHandler) Handle(ctx context.Context, factory factory, cloudWatchEvent *events.CloudWatchEvent) error {
	switch cloudWatchEvent.DetailType {
	case "evidence-received":
		return handleEvidenceReceived(ctx, factory.DynamoClient(), cloudWatchEvent)

	case "reduced-fee-approved":
		appData, err := factory.AppData()
		if err != nil {
			return err
		}

		shareCodeSender, err := factory.ShareCodeSender(ctx)
		if err != nil {
			return err
		}

		lpaStoreClient, err := factory.LpaStoreClient()
		if err != nil {
			return err
		}

		return handleFeeApproved(ctx, factory.DynamoClient(), cloudWatchEvent, shareCodeSender, lpaStoreClient, factory.EventClient(), appData, factory.Now())

	case "reduced-fee-declined":
		return handleFeeDenied(ctx, factory.DynamoClient(), cloudWatchEvent, factory.Now())

	case "further-info-requested":
		return handleFurtherInfoRequested(ctx, factory.DynamoClient(), cloudWatchEvent, factory.Now())

	case "donor-submission-completed":
		appData, err := factory.AppData()
		if err != nil {
			return err
		}

		shareCodeSender, err := factory.ShareCodeSender(ctx)
		if err != nil {
			return err
		}

		lpaStoreClient, err := factory.LpaStoreClient()
		if err != nil {
			return err
		}

		return handleDonorSubmissionCompleted(ctx, factory.DynamoClient(), cloudWatchEvent, shareCodeSender, appData, lpaStoreClient, factory.UuidString(), factory.Now())

	case "certificate-provider-submission-completed":
		return handleCertificateProviderSubmissionCompleted(ctx, cloudWatchEvent, factory)

	default:
		return fmt.Errorf("unknown sirius event")
	}
}

func handleEvidenceReceived(ctx context.Context, client dynamodbClient, event *events.CloudWatchEvent) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	var key dynamo.Keys
	if err := client.OneByUID(ctx, v.UID, &key); err != nil {
		return fmt.Errorf("failed to resolve uid: %w", err)
	}

	if key.PK == nil {
		return fmt.Errorf("PK missing from LPA in response")
	}

	if err := client.Put(ctx, map[string]string{"PK": key.PK.PK(), "SK": dynamo.EvidenceReceivedKey().SK()}); err != nil {
		return fmt.Errorf("failed to persist evidence received: %w", err)
	}

	return nil
}

func handleFeeApproved(
	ctx context.Context,
	client dynamodbClient,
	e *events.CloudWatchEvent,
	shareCodeSender ShareCodeSender,
	lpaStoreClient LpaStoreClient,
	eventClient EventClient,
	appData appcontext.Data,
	now func() time.Time,
) error {
	var v uidEvent
	if err := json.Unmarshal(e.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	donor, err := getDonorByLpaUID(ctx, client, v.UID)
	if err != nil {
		return err
	}

	if donor.Tasks.PayForLpa.IsCompleted() || donor.Tasks.PayForLpa.IsApproved() {
		return nil
	}

	if donor.FeeAmount() == 0 {
		donor.Tasks.PayForLpa = task.PaymentStateCompleted

		if donor.Tasks.ConfirmYourIdentityAndSign.IsCompleted() {
			if err := lpaStoreClient.SendLpa(ctx, donor); err != nil {
				return fmt.Errorf("failed to send to lpastore: %w", err)
			}

			if err := eventClient.SendCertificateProviderStarted(ctx, event.CertificateProviderStarted{
				UID: v.UID,
			}); err != nil {
				return fmt.Errorf("failed to send certificate-provider-started event: %w", err)
			}

			if err := shareCodeSender.SendCertificateProviderPrompt(ctx, appData, donor); err != nil {
				return fmt.Errorf("failed to send share code to certificate provider: %w", err)
			}
		}
	} else {
		donor.Tasks.PayForLpa = task.PaymentStateApproved
	}

	if err := putDonor(ctx, donor, now, client); err != nil {
		return fmt.Errorf("failed to update LPA task status: %w", err)
	}

	return nil
}

func handleFurtherInfoRequested(ctx context.Context, client dynamodbClient, event *events.CloudWatchEvent, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	donor, err := getDonorByLpaUID(ctx, client, v.UID)
	if err != nil {
		return err
	}

	if donor.Tasks.PayForLpa.IsMoreEvidenceRequired() {
		return nil
	}

	donor.Tasks.PayForLpa = task.PaymentStateMoreEvidenceRequired

	if err := putDonor(ctx, donor, now, client); err != nil {
		return fmt.Errorf("failed to update LPA task status: %w", err)
	}

	return nil
}

func handleFeeDenied(ctx context.Context, client dynamodbClient, event *events.CloudWatchEvent, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	donor, err := getDonorByLpaUID(ctx, client, v.UID)
	if err != nil {
		return err
	}

	if donor.Tasks.PayForLpa.IsDenied() {
		return nil
	}

	donor.FeeType = pay.FullFee
	donor.Tasks.PayForLpa = task.PaymentStateDenied

	if err := putDonor(ctx, donor, now, client); err != nil {
		return fmt.Errorf("failed to update LPA task status: %w", err)
	}

	return nil
}

func handleDonorSubmissionCompleted(ctx context.Context, client dynamodbClient, event *events.CloudWatchEvent, shareCodeSender ShareCodeSender, appData appcontext.Data, lpaStoreClient LpaStoreClient, uuidString func() string, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	var key dynamo.Keys
	if err := client.OneByUID(ctx, v.UID, &key); !errors.Is(err, dynamo.NotFoundError{}) {
		return err
	}

	lpaID := uuidString()

	if err := client.Put(ctx, &donordata.Provided{
		PK:        dynamo.LpaKey(lpaID),
		SK:        dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
		LpaID:     lpaID,
		LpaUID:    v.UID,
		CreatedAt: now(),
		Version:   1,
	}); err != nil {
		return err
	}

	lpa, err := lpaStoreClient.Lpa(ctx, v.UID)
	if err != nil {
		return err
	}

	if err := shareCodeSender.SendCertificateProviderInvite(ctx, appData, sharecode.CertificateProviderInvite{
		LpaKey:                      lpa.LpaKey,
		LpaOwnerKey:                 lpa.LpaOwnerKey,
		LpaUID:                      lpa.LpaUID,
		Type:                        lpa.Type,
		DonorFirstNames:             lpa.Donor.FirstNames,
		DonorFullName:               lpa.Donor.FullName(),
		CertificateProviderUID:      lpa.CertificateProvider.UID,
		CertificateProviderFullName: lpa.CertificateProvider.FullName(),
		CertificateProviderEmail:    lpa.CertificateProvider.Email,
	}); err != nil {
		return fmt.Errorf("failed to send share code to certificate provider: %w", err)
	}

	return nil
}

func handleCertificateProviderSubmissionCompleted(ctx context.Context, event *events.CloudWatchEvent, factory factory) error {
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

	if donor.CertificateProvider.Channel.IsPaper() {
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
