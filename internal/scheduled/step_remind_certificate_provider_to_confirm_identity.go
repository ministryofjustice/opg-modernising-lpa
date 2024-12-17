package scheduled

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func (r *Runner) stepRemindCertificateProviderToConfirmIdentity(ctx context.Context, row *Event) error {
	certificateProvider, err := r.certificateProviderStore.One(ctx, row.TargetLpaKey)
	if err != nil {
		return fmt.Errorf("error retrieving certificate provider: %w", err)
	}

	if certificateProvider.Tasks.ConfirmYourIdentity.IsCompleted() {
		return errStepIgnored
	}

	donor, err := r.donorStore.One(ctx, row.TargetLpaKey, row.TargetLpaOwnerKey)
	if err != nil {
		return fmt.Errorf("error retrieving donor: %w", err)
	}

	lpa, err := r.lpaStoreResolvingService.Resolve(ctx, donor)
	if err != nil {
		return fmt.Errorf("error resolving lpa: %w", err)
	}

	beforeExpiry := lpa.ExpiresAt().AddDate(0, -3, 0)
	afterSigning := lpa.CertificateProvider.SignedAt.AddDate(0, 3, 0)

	if r.now().Before(afterSigning) || r.now().Before(beforeExpiry) {
		return errStepIgnored
	}

	if lpa.CertificateProvider.Channel.IsPaper() {
		letterRequest := event.LetterRequested{
			UID:        lpa.LpaUID,
			LetterType: "ADVISE_CERTIFICATE_PROVIDER_TO_CONFIRM_IDENTITY",
			ActorType:  actor.TypeCertificateProvider,
			ActorUID:   lpa.CertificateProvider.UID,
		}

		if err := r.eventClient.SendLetterRequested(ctx, letterRequest); err != nil {
			return fmt.Errorf("could not send certificate provider letter request: %w", err)
		}
	} else {
		localizer := r.bundle.For(certificateProvider.ContactLanguagePreference)
		toCertificateProviderEmail := notify.ToLpaCertificateProvider(certificateProvider, lpa)

		if err := r.notifyClient.SendActorEmail(ctx, toCertificateProviderEmail, lpa.LpaUID, notify.AdviseCertificateProviderToConfirmIdentityEmail{
			DonorFullName:                   lpa.Donor.FullName(),
			DonorFullNamePossessive:         localizer.Possessive(lpa.Donor.FullName()),
			LpaType:                         localizer.T(lpa.Type.String()),
			CertificateProviderFullName:     lpa.CertificateProvider.FullName(),
			DeadlineDate:                    localizer.FormatDate(lpa.ExpiresAt()),
			CertificateProviderStartPageURL: r.appPublicURL + page.PathCertificateProviderStart.Format(),
		}); err != nil {
			return fmt.Errorf("could not send certificate provider email: %w", err)
		}
	}

	if lpa.Donor.Channel.IsPaper() {
		letterRequest := event.LetterRequested{
			UID:        lpa.LpaUID,
			LetterType: "INFORM_DONOR_CERTIFICATE_PROVIDER_HAS_NOT_CONFIRMED_IDENTITY",
			ActorType:  actor.TypeDonor,
			ActorUID:   lpa.Donor.UID,
		}

		if lpa.Correspondent.Address.Line1 != "" {
			letterRequest.ActorType = actor.TypeCorrespondent
			letterRequest.ActorUID = lpa.Correspondent.UID
		}

		if err := r.eventClient.SendLetterRequested(ctx, letterRequest); err != nil {
			return fmt.Errorf("could not send donor letter request: %w", err)
		}
	} else {
		localizer := r.bundle.For(certificateProvider.ContactLanguagePreference)
		toDonorEmail := notify.ToLpaDonor(lpa)

		if err := r.notifyClient.SendActorEmail(ctx, toDonorEmail, lpa.LpaUID, notify.InformDonorCertificateProviderHasNotConfirmedIdentityEmail{
			Greeting:                        r.notifyClient.EmailGreeting(lpa),
			CertificateProviderFullName:     lpa.CertificateProvider.FullName(),
			LpaType:                         localizer.T(lpa.Type.String()),
			DeadlineDate:                    localizer.FormatDate(lpa.ExpiresAt()),
			CertificateProviderStartPageURL: r.appPublicURL + page.PathCertificateProviderStart.Format(),
		}); err != nil {
			return fmt.Errorf("could not send donor email: %w", err)
		}
	}

	return nil
}
