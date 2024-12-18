package scheduled

import (
	"context"
	"errors"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func (r *Runner) stepRemindAttorneyToComplete(ctx context.Context, row *Event) error {
	donor, err := r.donorStore.One(ctx, row.TargetLpaKey, row.TargetLpaOwnerKey)
	if err != nil {
		return fmt.Errorf("error retrieving donor: %w", err)
	}

	lpa, err := r.lpaStoreResolvingService.Resolve(ctx, donor)
	if err != nil {
		return fmt.Errorf("error resolving lpa: %w", err)
	}

	beforeExpiry := lpa.ExpiresAt().AddDate(0, -3, 0)
	afterInvite := lpa.AttorneysInvitedAt.AddDate(0, 3, 0)

	if r.now().Before(afterInvite) || r.now().Before(beforeExpiry) {
		return errStepIgnored
	}

	attorneys, err := r.attorneyStore.All(ctx, donor.LpaUID)
	if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return fmt.Errorf("error retrieving attorney: %w", err)
	}

	attorneyMap := map[actoruid.UID]*attorneydata.Provided{}
	for _, attorney := range attorneys {
		attorneyMap[attorney.UID] = attorney
	}

	ran := false

	for _, attorney := range lpa.Attorneys.Attorneys {
		if provided, ok := attorneyMap[attorney.UID]; !ok || !provided.Signed() {
			ran = true
			if err := r.stepRemindAttorneyToCompleteAttorney(ctx, lpa, actor.TypeAttorney, attorney, provided); err != nil {
				return err
			}
		}
	}

	if trustCorporation := lpa.Attorneys.TrustCorporation; !lpa.Attorneys.TrustCorporation.UID.IsZero() {
		if provided, ok := attorneyMap[trustCorporation.UID]; !ok || !provided.Signed() {
			ran = true
			if err := r.stepRemindAttorneyToCompleteTrustCorporation(ctx, lpa, actor.TypeTrustCorporation, trustCorporation, provided); err != nil {
				return err
			}
		}
	}

	for _, attorney := range lpa.ReplacementAttorneys.Attorneys {
		if provided, ok := attorneyMap[attorney.UID]; !ok || !provided.Signed() {
			ran = true
			if err := r.stepRemindAttorneyToCompleteAttorney(ctx, lpa, actor.TypeReplacementAttorney, attorney, provided); err != nil {
				return err
			}
		}
	}

	if trustCorporation := lpa.ReplacementAttorneys.TrustCorporation; !trustCorporation.UID.IsZero() {
		if provided, ok := attorneyMap[trustCorporation.UID]; !ok || !provided.Signed() {
			ran = true
			if err := r.stepRemindAttorneyToCompleteTrustCorporation(ctx, lpa, actor.TypeReplacementTrustCorporation, trustCorporation, provided); err != nil {
				return err
			}
		}
	}

	if !ran {
		return errStepIgnored
	}

	return nil
}

func (r *Runner) stepRemindAttorneyToCompleteAttorney(ctx context.Context, lpa *lpadata.Lpa, actorType actor.Type, attorney lpadata.Attorney, provided *attorneydata.Provided) error {
	if attorney.Channel.IsPaper() {
		letterRequest := event.LetterRequested{
			UID:        lpa.LpaUID,
			LetterType: "ADVISE_ATTORNEY_TO_SIGN_OR_OPT_OUT",
			ActorType:  actorType,
			ActorUID:   attorney.UID,
		}

		if err := r.eventClient.SendLetterRequested(ctx, letterRequest); err != nil {
			return fmt.Errorf("could not send attorney letter request: %w", err)
		}
	} else {
		localizer := r.bundle.For(localize.En)
		if provided != nil && !provided.ContactLanguagePreference.Empty() {
			localizer = r.bundle.For(provided.ContactLanguagePreference)
		}

		toAttorneyEmail := notify.ToLpaAttorney(attorney)

		if err := r.notifyClient.SendActorEmail(ctx, toAttorneyEmail, lpa.LpaUID, notify.AdviseAttorneyToSignOrOptOutEmail{
			DonorFullName:           lpa.Donor.FullName(),
			DonorFullNamePossessive: localizer.Possessive(lpa.Donor.FullName()),
			LpaType:                 localizer.T(lpa.Type.String()),
			AttorneyFullName:        attorney.FullName(),
			InvitedDate:             localizer.FormatDate(lpa.AttorneysInvitedAt),
			DeadlineDate:            localizer.FormatDate(lpa.ExpiresAt()),
			AttorneyStartPageURL:    r.appPublicURL + page.PathAttorneyStart.Format(),
		}); err != nil {
			return fmt.Errorf("could not send attorney email: %w", err)
		}
	}

	if lpa.Donor.Channel.IsPaper() {
		letterRequest := event.LetterRequested{
			UID:        lpa.LpaUID,
			LetterType: "INFORM_DONOR_ATTORNEY_HAS_NOT_ACTED",
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
		localizer := r.bundle.For(lpa.Donor.ContactLanguagePreference)
		toDonorEmail := notify.ToLpaDonor(lpa)

		if err := r.notifyClient.SendActorEmail(ctx, toDonorEmail, lpa.LpaUID, notify.InformDonorAttorneyHasNotActedEmail{
			Greeting:             r.notifyClient.EmailGreeting(lpa),
			AttorneyFullName:     attorney.FullName(),
			LpaType:              localizer.T(lpa.Type.String()),
			InvitedDate:          localizer.FormatDate(lpa.AttorneysInvitedAt),
			DeadlineDate:         localizer.FormatDate(lpa.ExpiresAt()),
			AttorneyStartPageURL: r.appPublicURL + page.PathAttorneyStart.Format(),
		}); err != nil {
			return fmt.Errorf("could not send donor email: %w", err)
		}
	}

	return nil
}

func (r *Runner) stepRemindAttorneyToCompleteTrustCorporation(ctx context.Context, lpa *lpadata.Lpa, actorType actor.Type, trustCorporation lpadata.TrustCorporation, provided *attorneydata.Provided) error {
	if trustCorporation.Channel.IsPaper() {
		letterRequest := event.LetterRequested{
			UID:        lpa.LpaUID,
			LetterType: "ADVISE_ATTORNEY_TO_SIGN_OR_OPT_OUT",
			ActorType:  actorType,
			ActorUID:   trustCorporation.UID,
		}

		if err := r.eventClient.SendLetterRequested(ctx, letterRequest); err != nil {
			return fmt.Errorf("could not send certificate provider letter request: %w", err)
		}
	} else {
		localizer := r.bundle.For(localize.En)
		if provided != nil && !provided.ContactLanguagePreference.Empty() {
			localizer = r.bundle.For(provided.ContactLanguagePreference)
		}

		toAttorneyEmail := notify.ToLpaTrustCorporation(trustCorporation)

		if err := r.notifyClient.SendActorEmail(ctx, toAttorneyEmail, lpa.LpaUID, notify.AdviseAttorneyToSignOrOptOutEmail{
			DonorFullName:           lpa.Donor.FullName(),
			DonorFullNamePossessive: localizer.Possessive(lpa.Donor.FullName()),
			LpaType:                 localizer.T(lpa.Type.String()),
			AttorneyFullName:        trustCorporation.Name,
			InvitedDate:             localizer.FormatDate(lpa.AttorneysInvitedAt),
			DeadlineDate:            localizer.FormatDate(lpa.ExpiresAt()),
			AttorneyStartPageURL:    r.appPublicURL + page.PathAttorneyStart.Format(),
		}); err != nil {
			return fmt.Errorf("could not send trust corporation email: %w", err)
		}
	}

	if lpa.Donor.Channel.IsPaper() {
		letterRequest := event.LetterRequested{
			UID:        lpa.LpaUID,
			LetterType: "INFORM_DONOR_ATTORNEY_HAS_NOT_ACTED",
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
		localizer := r.bundle.For(lpa.Donor.ContactLanguagePreference)
		toDonorEmail := notify.ToLpaDonor(lpa)

		if err := r.notifyClient.SendActorEmail(ctx, toDonorEmail, lpa.LpaUID, notify.InformDonorAttorneyHasNotActedEmail{
			Greeting:             r.notifyClient.EmailGreeting(lpa),
			AttorneyFullName:     trustCorporation.Name,
			LpaType:              localizer.T(lpa.Type.String()),
			InvitedDate:          localizer.FormatDate(lpa.AttorneysInvitedAt),
			DeadlineDate:         localizer.FormatDate(lpa.ExpiresAt()),
			AttorneyStartPageURL: r.appPublicURL + page.PathAttorneyStart.Format(),
		}); err != nil {
			return fmt.Errorf("could not send donor email: %w", err)
		}
	}

	return nil
}
