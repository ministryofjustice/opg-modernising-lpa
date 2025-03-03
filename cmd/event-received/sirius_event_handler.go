package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
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

	case "priority-correspondence-sent":
		return handlePriorityCorrespondenceSent(ctx, factory.DynamoClient(), cloudWatchEvent, factory.Now())

	case "immaterial-change-confirmed":
		lpaStoreClient, err := factory.LpaStoreClient()
		if err != nil {
			return fmt.Errorf("failed to instantiaite lpaStoreClient: %w", err)
		}

		return handleChangeConfirmed(ctx, factory.DynamoClient(), factory.CertificateProviderStore(), cloudWatchEvent, factory.Now(), lpaStoreClient, false)

	case "material-change-confirmed":
		lpaStoreClient, err := factory.LpaStoreClient()
		if err != nil {
			return fmt.Errorf("failed to instantiaite lpaStoreClient: %w", err)
		}

		return handleChangeConfirmed(ctx, factory.DynamoClient(), factory.CertificateProviderStore(), cloudWatchEvent, factory.Now(), lpaStoreClient, true)

	case "certificate-provider-identity-check-failed":
		notifyClient, err := factory.NotifyClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to instantiaite notifyClient: %w", err)
		}

		bundle, err := factory.Bundle()
		if err != nil {
			return fmt.Errorf("failed to instantiaite bundle: %w", err)
		}

		return handleCertificateProviderIdentityCheckedFailed(ctx, factory.DynamoClient(), notifyClient, bundle, factory.AppPublicURL(), cloudWatchEvent)
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
	var v feeApprovedEvent
	if err := json.Unmarshal(e.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	donor, err := getDonorByLpaUID(ctx, client, v.UID)
	if err != nil {
		return fmt.Errorf("failed to get donor: %w", err)
	}

	if donor.Tasks.PayForLpa.IsCompleted() || donor.Tasks.PayForLpa.IsApproved() {
		return nil
	}

	donor.FeeType = v.ApprovedType

	if donor.FeeAmount() <= 0 {
		donor.Tasks.PayForLpa = task.PaymentStateCompleted

		if donor.Tasks.SignTheLpa.IsCompleted() {
			if err := lpaStoreClient.SendLpa(ctx, donor.LpaUID, lpastore.CreateLpaFromDonorProvided(donor)); err != nil {
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

		if donor.Voucher.Allowed && donor.VoucherInvitedAt.IsZero() {
			if err := shareCodeSender.SendVoucherAccessCode(ctx, donor, appData); err != nil {
				return err
			}

			donor.VoucherInvitedAt = now()
		}
	} else {
		donor.Tasks.PayForLpa = task.PaymentStateApproved
	}

	donor.ReducedFeeApprovedAt = now()

	if err := putDonor(ctx, donor, now, client); err != nil {
		return fmt.Errorf("failed to update donor provided details: %w", err)
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
		return fmt.Errorf("failed to get donor: %w", err)
	}

	if donor.Tasks.PayForLpa.IsMoreEvidenceRequired() {
		return nil
	}

	donor.Tasks.PayForLpa = task.PaymentStateMoreEvidenceRequired
	donor.MoreEvidenceRequiredAt = now()

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
		return fmt.Errorf("failed to get donor: %w", err)
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

	lpa, err := lpaStoreClient.Lpa(ctx, v.UID)
	if err != nil {
		return err
	}

	// There is no certificate provider record yet, so assume English
	to := notify.ToLpaCertificateProvider(&certificateproviderdata.Provided{ContactLanguagePreference: localize.En}, lpa)

	if err := shareCodeSender.SendCertificateProviderInvite(ctx, appData, sharecode.CertificateProviderInvite{
		LpaKey:                      lpa.LpaKey,
		LpaOwnerKey:                 lpa.LpaOwnerKey,
		LpaUID:                      lpa.LpaUID,
		Type:                        lpa.Type,
		DonorFirstNames:             lpa.Donor.FirstNames,
		DonorFullName:               lpa.Donor.FullName(),
		CertificateProviderUID:      lpa.CertificateProvider.UID,
		CertificateProviderFullName: lpa.CertificateProvider.FullName(),
	}, to); err != nil {
		return fmt.Errorf("failed to send share code to certificate provider: %w", err)
	}

	lpaID := uuidString()

	donor := &donordata.Provided{
		PK:                           dynamo.LpaKey(lpaID),
		SK:                           dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
		LpaID:                        lpaID,
		LpaUID:                       v.UID,
		CreatedAt:                    now(),
		Version:                      1,
		CertificateProviderInvitedAt: now(),
	}

	transaction := dynamo.NewTransaction().
		Create(donor).
		Create(scheduled.Event{
			PK:                dynamo.ScheduledDayKey(donor.CertificateProviderInvitedAt.AddDate(0, 3, 1)),
			SK:                dynamo.ScheduledKey(donor.CertificateProviderInvitedAt.AddDate(0, 3, 1), uuidString()),
			CreatedAt:         now(),
			At:                donor.CertificateProviderInvitedAt.AddDate(0, 3, 1),
			Action:            scheduled.ActionRemindCertificateProviderToComplete,
			TargetLpaKey:      donor.PK,
			TargetLpaOwnerKey: donor.SK,
			LpaUID:            donor.LpaUID,
		}).
		Create(dynamo.Keys{PK: dynamo.UIDKey(v.UID), SK: dynamo.MetadataKey("")}).
		Create(dynamo.Keys{PK: donor.PK, SK: dynamo.ReservedKey(dynamo.DonorKey)})

	if err := client.WriteTransaction(ctx, transaction); err != nil {
		return err
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

	lpa, err := lpaStoreClient.Lpa(ctx, v.UID)
	if err != nil {
		return fmt.Errorf("failed to retrieve lpa: %w", err)
	}

	if lpa.CertificateProvider.Channel.IsPaper() {
		shareCodeSender, err := factory.ShareCodeSender(ctx)
		if err != nil {
			return err
		}

		appData, err := factory.AppData()
		if err != nil {
			return err
		}

		dynamoClient := factory.DynamoClient()

		donor, err := getDonorByLpaUID(ctx, dynamoClient, v.UID)
		if err != nil {
			return fmt.Errorf("failed to get donor: %w", err)
		}

		now := factory.Now()
		donor.AttorneysInvitedAt = now()

		if err := factory.ScheduledStore().DeleteAllActionByUID(ctx, []scheduled.Action{
			scheduled.ActionRemindCertificateProviderToComplete,
			scheduled.ActionRemindCertificateProviderToConfirmIdentity,
		}, v.UID); err != nil {
			return fmt.Errorf("failed to delete scheduled events: %w", err)
		}

		if err := shareCodeSender.SendAttorneys(ctx, appData, lpa); err != nil {
			return fmt.Errorf("failed to send share codes to attorneys: %w", err)
		}

		if err := putDonor(ctx, donor, now, dynamoClient); err != nil {
			return fmt.Errorf("failed to put donor: %w", err)
		}
	}

	return nil
}

type priorityCorrespondenceSentEvent struct {
	UID      string    `json:"uid"`
	SentDate time.Time `json:"sentDate"`
}

func handlePriorityCorrespondenceSent(ctx context.Context, client dynamodbClient, event *events.CloudWatchEvent, now func() time.Time) error {
	var v priorityCorrespondenceSentEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	donor, err := getDonorByLpaUID(ctx, client, v.UID)
	if err != nil {
		return fmt.Errorf("failed to get donor: %w", err)
	}

	donor.PriorityCorrespondenceSentAt = v.SentDate

	if err := putDonor(ctx, donor, now, client); err != nil {
		return fmt.Errorf("failed to update donor: %w", err)
	}

	return nil
}

type changeConfirmedEvent struct {
	UID       string     `json:"uid"`
	ActorType actor.Type `json:"actorType"`
	ActorUID  string     `json:"actorUID"`
}

func handleChangeConfirmed(ctx context.Context, client dynamodbClient, certificateProviderStore CertificateProviderStore, event *events.CloudWatchEvent, now func() time.Time, lpaStoreClient LpaStoreClient, materialChange bool) error {
	var v changeConfirmedEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	switch v.ActorType {
	case actor.TypeDonor:
		donor, err := getDonorByLpaUID(ctx, client, v.UID)
		if err != nil {
			return fmt.Errorf("failed to get donor: %w", err)
		}

		if donor.Tasks.ConfirmYourIdentity.IsPending() && (donor.ContinueWithMismatchedIdentity || !donor.SignedAt.IsZero()) {
			if materialChange {
				donor.MaterialChangeConfirmedAt = now()
				donor.Tasks.ConfirmYourIdentity = task.IdentityStateProblem
			} else {
				donor.ImmaterialChangeConfirmedAt = now()
				donor.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted

				if err := lpaStoreClient.SendDonorConfirmIdentity(ctx, donor); err != nil {
					if !errors.Is(err, lpastore.ErrNotFound) {
						return fmt.Errorf("failed to send donor confirmed identity to lpa store: %w", err)
					}
				}
			}

			if err := putDonor(ctx, donor, now, client); err != nil {
				return fmt.Errorf("failed to update donor: %w", err)
			}
		}
	case actor.TypeCertificateProvider:
		certificateProvider, err := certificateProviderStore.OneByUID(ctx, v.UID)
		if err != nil {
			return fmt.Errorf("failed to get certificate provider: %w", err)
		}

		if certificateProvider.Tasks.ConfirmYourIdentity.IsPending() && certificateProvider.IdentityDetailsMismatched {
			if materialChange {
				certificateProvider.MaterialChangeConfirmedAt = now()
				certificateProvider.Tasks.ConfirmYourIdentity = task.IdentityStateProblem
			} else {
				certificateProvider.ImmaterialChangeConfirmedAt = now()
				certificateProvider.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted

				if err := lpaStoreClient.SendCertificateProviderConfirmIdentity(ctx, v.UID, certificateProvider); err != nil {
					if !errors.Is(err, lpastore.ErrNotFound) {
						return fmt.Errorf("failed to send certificate provider confirmed identity to lpa store: %w", err)
					}
				}
			}

			if err := certificateProviderStore.Put(ctx, certificateProvider); err != nil {
				return fmt.Errorf("failed to update certificate provider: %w", err)
			}
		}
	default:
		return fmt.Errorf("invalid actorType, got %s, want donor or certificateProvider", v.ActorType.String())
	}

	return nil
}

func handleCertificateProviderIdentityCheckedFailed(ctx context.Context, dynamoClient dynamodbClient, notifyClient NotifyClient, bundle Bundle, appPublicURL string, event *events.CloudWatchEvent) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	provided, err := getDonorByLpaUID(ctx, dynamoClient, v.UID)
	if err != nil {
		return fmt.Errorf("failed to get donor: %w", err)
	}

	localizer := bundle.For(provided.Donor.ContactLanguagePreference)

	return notifyClient.SendActorEmail(ctx, notify.ToDonor(provided), v.UID, notify.InformDonorPaperCertificateProviderIdentityCheckFailed{
		CertificateProviderFullName: provided.CertificateProvider.FullName(),
		LpaType:                     localize.LowerFirst(localizer.T(provided.Type.String())),
		DonorStartPageURL:           appPublicURL + localizer.Lang().URL(page.PathStart.Format()),
	})
}
