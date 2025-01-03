package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
)

type lpastoreEventHandler struct{}

type lpaUpdatedEvent struct {
	UID        string `json:"uid"`
	ChangeType string `json:"changeType"`
}

func (h *lpastoreEventHandler) Handle(ctx context.Context, factory factory, cloudWatchEvent *events.CloudWatchEvent) error {
	if cloudWatchEvent.DetailType == "lpa-updated" {
		var v lpaUpdatedEvent
		if err := json.Unmarshal(cloudWatchEvent.Detail, &v); err != nil {
			return fmt.Errorf("failed to unmarshal detail: %w", err)
		}

		switch v.ChangeType {
		case "CREATE":
			lpaStoreClient, err := factory.LpaStoreClient()
			if err != nil {
				return fmt.Errorf("could not create LpaStoreClient: %w", err)
			}

			bundle, err := factory.Bundle()
			if err != nil {
				return fmt.Errorf("could not load Bundle: %w", err)
			}

			notifyClient, err := factory.NotifyClient(ctx)
			if err != nil {
				return fmt.Errorf("could not create NotifyClient: %w", err)
			}

			return handleCreate(ctx, factory.DynamoClient(), lpaStoreClient, notifyClient, bundle, v)

		case "CERTIFICATE_PROVIDER_SIGN":
			lpaStoreClient, err := factory.LpaStoreClient()
			if err != nil {
				return fmt.Errorf("could not create LpaStoreClient: %w", err)
			}

			bundle, err := factory.Bundle()
			if err != nil {
				return fmt.Errorf("could not load Bundle: %w", err)
			}

			notifyClient, err := factory.NotifyClient(ctx)
			if err != nil {
				return fmt.Errorf("could not create NotifyClient: %w", err)
			}

			return handleCertificateProviderSign(ctx, factory.DynamoClient(), lpaStoreClient, notifyClient, bundle, v)

		case "REGISTER":
			lpaStoreClient, err := factory.LpaStoreClient()
			if err != nil {
				return fmt.Errorf("could not create LpaStoreClient: %w", err)
			}

			return handleRegister(ctx, factory.DynamoClient(), lpaStoreClient, factory.EventClient(), v)

		case "STATUTORY_WAITING_PERIOD":
			return handleStatutoryWaitingPeriod(ctx, factory.DynamoClient(), factory.Now(), v)

		case "CANNOT_REGISTER":
			return handleCannotRegister(ctx, factory.ScheduledStore(), v)

		default:
			return nil
		}
	}

	return fmt.Errorf("unknown lpastore event")
}

func handleCreate(ctx context.Context, client dynamodbClient, lpaStoreClient LpaStoreClient, notifyClient NotifyClient, bundle Bundle, v lpaUpdatedEvent) error {
	lpa, err := lpaStoreClient.Lpa(ctx, v.UID)
	if err != nil {
		return fmt.Errorf("error getting lpa: %w", err)
	}

	localizer := bundle.For(lpa.Donor.ContactLanguagePreference)

	if lpa.Donor.Channel.IsPaper() {
		if lpa.Donor.Mobile != "" {
			if err := notifyClient.SendActorSMS(ctx, notify.ToLpaDonor(lpa), v.UID, notify.PaperDonorLpaSubmittedSMS{
				LpaType: localizer.T(lpa.Type.String()),
			}); err != nil {
				return fmt.Errorf("error sending sms: %w", err)
			}
		}

		return nil
	}

	donor, err := getDonorByLpaUID(ctx, client, v.UID)
	if err != nil {
		return fmt.Errorf("error getting donor: %w", err)
	}

	if err := notifyClient.SendActorEmail(ctx, notify.ToDonor(donor), v.UID, notify.DigitalDonorLpaSubmittedEmail{
		Greeting: notifyClient.EmailGreeting(lpa),
		LpaType:  localizer.T(lpa.Type.String()),
	}); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	return nil
}

func handleCertificateProviderSign(ctx context.Context, client dynamodbClient, lpaStoreClient LpaStoreClient, notifyClient NotifyClient, bundle Bundle, v lpaUpdatedEvent) error {
	lpa, err := lpaStoreClient.Lpa(ctx, v.UID)
	if err != nil {
		return fmt.Errorf("error getting lpa: %w", err)
	}

	localizer := bundle.For(lpa.Donor.ContactLanguagePreference)

	if lpa.Donor.Channel.IsPaper() {
		if lpa.Donor.Mobile != "" {
			if err := notifyClient.SendActorSMS(ctx, notify.ToLpaDonor(lpa), v.UID, notify.PaperDonorCertificateProvidedSMS{
				CertificateProviderFullName: lpa.CertificateProvider.FullName(),
				LpaType:                     localizer.T(lpa.Type.String()),
			}); err != nil {
				return fmt.Errorf("error sending sms: %w", err)
			}
		}

		return nil
	}

	donor, err := getDonorByLpaUID(ctx, client, v.UID)
	if err != nil {
		return fmt.Errorf("error getting donor: %w", err)
	}

	if err := notifyClient.SendActorEmail(ctx, notify.ToDonor(donor), v.UID, notify.DigitalDonorCertificateProvidedEmail{
		Greeting:                    notifyClient.EmailGreeting(lpa),
		CertificateProviderFullName: lpa.CertificateProvider.FullName(),
		LpaType:                     localizer.T(lpa.Type.String()),
	}); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	return nil
}

func handleRegister(ctx context.Context, client dynamodbClient, lpaStoreClient LpaStoreClient, eventClient EventClient, v lpaUpdatedEvent) error {
	lpa, err := lpaStoreClient.Lpa(ctx, v.UID)
	if err != nil {
		return fmt.Errorf("error getting lpa: %w", err)
	}

	var links []dashboarddata.LpaLink
	if err := client.AllByLpaUIDAndPartialSK(ctx, v.UID, dynamo.SubKey(""), &links); err != nil {
		return fmt.Errorf("error getting all subs for uid: %w", err)
	}

	data := event.LpaAccessGranted{
		UID:     v.UID,
		LpaType: lpa.Type.String(),
	}

	for _, link := range links {
		if !link.ActorType.IsDonor() &&
			!link.ActorType.IsAttorney() && !link.ActorType.IsReplacementAttorney() &&
			!link.ActorType.IsTrustCorporation() && !link.ActorType.IsReplacementTrustCorporation() {
			continue
		}

		sub, _ := base64.StdEncoding.DecodeString(link.UserSub())

		data.Actors = append(data.Actors, event.LpaAccessGrantedActor{
			SubjectID: string(sub),
			ActorUID:  link.UID.String(),
		})
	}

	return eventClient.SendLpaAccessGranted(ctx, data)
}

func handleStatutoryWaitingPeriod(ctx context.Context, client dynamodbClient, now func() time.Time, event lpaUpdatedEvent) error {
	donor, err := getDonorByLpaUID(ctx, client, event.UID)
	if err != nil {
		return err
	}

	donor.StatutoryWaitingPeriodAt = now()

	if err := putDonor(ctx, donor, now, client); err != nil {
		return fmt.Errorf("failed to update donor details: %w", err)
	}

	return nil
}

func handleCannotRegister(ctx context.Context, store ScheduledStore, event lpaUpdatedEvent) error {
	return store.DeleteAllByUID(ctx, event.UID)
}
