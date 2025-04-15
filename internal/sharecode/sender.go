package sharecode

import (
	"context"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
)

type Localizer interface {
	localize.Localizer
}

type ShareCodeStore interface {
	Get(ctx context.Context, actorType actor.Type, shareCode sharecodedata.Hashed) (sharecodedata.Link, error)
	Put(ctx context.Context, actorType actor.Type, shareCode sharecodedata.Hashed, data sharecodedata.Link) error
}

type NotifyClient interface {
	SendActorEmail(context context.Context, to notify.ToEmail, lpaUID string, email notify.Email) error
	SendActorSMS(context context.Context, to notify.ToMobile, lpaUID string, sms notify.SMS) error
}

type EventClient interface {
	SendAttorneyStarted(ctx context.Context, event event.AttorneyStarted) error
	SendNotificationSent(ctx context.Context, notificationSentEvent event.NotificationSent) error
	SendPaperFormRequested(ctx context.Context, paperFormRequestedEvent event.PaperFormRequested) error
}

type CertificateProviderStore interface {
	GetAny(ctx context.Context) (*certificateproviderdata.Provided, error)
}

type ScheduledStore interface {
	Create(ctx context.Context, rows ...scheduled.Event) error
}

type Sender struct {
	testCode                    string
	shareCodeStore              ShareCodeStore
	certificateProviderStore    CertificateProviderStore
	scheduledStore              ScheduledStore
	notifyClient                NotifyClient
	appPublicURL                string
	certificateProviderStartURL string
	attorneyStartURL            string
	eventClient                 EventClient
	generate                    func() (sharecodedata.PlainText, sharecodedata.Hashed)
	now                         func() time.Time
}

func NewSender(shareCodeStore ShareCodeStore, notifyClient NotifyClient, appPublicURL, certificateProviderStartURL, attorneyStartURL string, eventClient EventClient, certificateProviderStore CertificateProviderStore, scheduledStore ScheduledStore) *Sender {
	return &Sender{
		shareCodeStore:              shareCodeStore,
		notifyClient:                notifyClient,
		appPublicURL:                appPublicURL,
		certificateProviderStartURL: certificateProviderStartURL,
		attorneyStartURL:            attorneyStartURL,
		eventClient:                 eventClient,
		certificateProviderStore:    certificateProviderStore,
		scheduledStore:              scheduledStore,
		generate:                    sharecodedata.Generate,
		now:                         time.Now,
	}
}

func (s *Sender) UseTestCode(shareCode string) {
	s.testCode = shareCode
}

type CertificateProviderInvite struct {
	LpaKey                      dynamo.LpaKeyType
	LpaOwnerKey                 dynamo.LpaOwnerKeyType
	LpaUID                      string
	Type                        lpadata.LpaType
	DonorFirstNames             string
	DonorFullName               string
	CertificateProviderUID      actoruid.UID
	CertificateProviderFullName string
}

func (s *Sender) SendCertificateProviderInvite(ctx context.Context, appData appcontext.Data, invite CertificateProviderInvite, to notify.ToEmail) error {
	shareCode, err := s.createShareCode(ctx, invite.LpaKey, invite.LpaOwnerKey, invite.LpaUID, invite.CertificateProviderUID, actor.TypeCertificateProvider)
	if err != nil {
		return err
	}

	whatLpaCovers := "whatPropertyAndAffairsCovers"
	if invite.Type.IsPersonalWelfare() {
		whatLpaCovers = "whatPersonalWelfareCovers"
	}

	return s.sendEmail(ctx, to, invite.LpaUID, notify.CertificateProviderInviteEmail{
		CertificateProviderFullName:  invite.CertificateProviderFullName,
		DonorFullName:                invite.DonorFullName,
		LpaType:                      localize.LowerFirst(appData.Localizer.T(invite.Type.String())),
		CertificateProviderStartURL:  s.certificateProviderStartURL,
		DonorFirstNames:              invite.DonorFirstNames,
		DonorFirstNamesPossessive:    appData.Localizer.Possessive(invite.DonorFirstNames),
		WhatLpaCovers:                appData.Localizer.T(whatLpaCovers),
		ShareCode:                    shareCode.Plain(),
		CertificateProviderOptOutURL: fmt.Sprintf("%s%s", s.appPublicURL, page.PathCertificateProviderEnterReferenceNumberOptOut),
	})
}

func (s *Sender) SendCertificateProviderPrompt(ctx context.Context, appData appcontext.Data, donor *donordata.Provided) error {
	shareCode, err := s.createShareCode(ctx, donor.PK, donor.SK, donor.LpaUID, donor.CertificateProvider.UID, actor.TypeCertificateProvider)
	if err != nil {
		return err
	}

	if donor.CertificateProvider.CarryOutBy.IsPaper() {
		return s.sendPaperForm(ctx, donor.LpaUID, actor.TypeCertificateProvider, donor.CertificateProvider.UID, shareCode)
	}

	to := notify.ToCertificateProvider(donor.CertificateProvider)
	if certificateProvider, err := s.certificateProviderStore.GetAny(ctx); err == nil {
		to = notify.ToProvidedCertificateProvider(certificateProvider, donor.CertificateProvider)
	}

	return s.sendEmail(ctx, to, donor.LpaUID, notify.CertificateProviderProvideCertificatePromptEmail{
		CertificateProviderFullName: donor.CertificateProvider.FullName(),
		DonorFullName:               donor.Donor.FullName(),
		LpaType:                     localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
		CertificateProviderStartURL: s.certificateProviderStartURL,
		ShareCode:                   shareCode.Plain(),
	})
}

func (s *Sender) SendAttorneys(ctx context.Context, appData appcontext.Data, lpa *lpadata.Lpa) error {
	if err := s.scheduledStore.Create(ctx, scheduled.Event{
		At:                s.now().AddDate(0, 3, 1),
		Action:            scheduled.ActionRemindAttorneyToComplete,
		TargetLpaKey:      lpa.LpaKey,
		TargetLpaOwnerKey: lpa.LpaOwnerKey,
		LpaUID:            lpa.LpaUID,
	}, scheduled.Event{
		At:                lpa.ExpiresAt().AddDate(0, -3, 1),
		Action:            scheduled.ActionRemindAttorneyToComplete,
		TargetLpaKey:      lpa.LpaKey,
		TargetLpaOwnerKey: lpa.LpaOwnerKey,
		LpaUID:            lpa.LpaUID,
	}); err != nil {
		return fmt.Errorf("error scheduling attorneys prompt: %w", err)
	}

	if err := s.sendTrustCorporation(ctx, appData, lpa, lpa.Attorneys.TrustCorporation); err != nil {
		return err
	}
	if err := s.sendReplacementTrustCorporation(ctx, appData, lpa, lpa.ReplacementAttorneys.TrustCorporation); err != nil {
		return err
	}

	for _, attorney := range lpa.Attorneys.Attorneys {
		if err := s.sendOriginalAttorney(ctx, appData, lpa, attorney); err != nil {
			return err
		}
	}

	for _, attorney := range lpa.ReplacementAttorneys.Attorneys {
		if err := s.sendReplacementAttorney(ctx, appData, lpa, attorney); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sender) SendVoucherInvite(ctx context.Context, provided *donordata.Provided, appData appcontext.Data) error {
	if err := s.SendVoucherAccessCode(ctx, provided, appData); err != nil {
		return err
	}

	if provided.Correspondent.Email != "" {
		if err := s.sendEmail(ctx, notify.ToCorrespondent(provided), provided.LpaUID, notify.CorrespondentInformedVouchingInProgress{
			CorrespondentFullName:   provided.Correspondent.FullName(),
			DonorFullName:           provided.Donor.FullName(),
			DonorFullNamePossessive: appData.Localizer.Possessive(provided.Donor.FullName()),
			LpaType:                 appData.Localizer.T(provided.Type.String()),
		}); err != nil {
			return err
		}
	}

	return s.sendEmail(ctx, notify.ToVoucher(provided.Voucher), provided.LpaUID, notify.VoucherInviteEmail{
		VoucherFullName:           provided.Voucher.FullName(),
		DonorFullName:             provided.Donor.FullName(),
		DonorFirstNamesPossessive: appData.Localizer.Possessive(provided.Donor.FirstNames),
		DonorFirstNames:           provided.Donor.FirstNames,
		LpaType:                   appData.Localizer.T(provided.Type.String()),
		VoucherStartPageURL:       s.appPublicURL + page.PathVoucherStart.Format(),
	})
}

func (s *Sender) SendVoucherAccessCode(ctx context.Context, provided *donordata.Provided, appData appcontext.Data) error {
	shareCode, err := s.createShareCode(ctx, provided.PK, provided.SK, provided.LpaUID, provided.Voucher.UID, actor.TypeVoucher)
	if err != nil {
		return err
	}

	provided.VoucherInvitedAt = s.now()

	if provided.Donor.Mobile != "" {
		provided.VoucherCodeSentBySMS = true
		provided.VoucherCodeSentTo = provided.Donor.Mobile

		if err := s.sendSMS(ctx, notify.ToDonorOnly(provided), provided.LpaUID, notify.VouchingShareCodeSMS{
			ShareCode:                 shareCode.Plain(),
			DonorFullNamePossessive:   appData.Localizer.Possessive(provided.Donor.FullName()),
			LpaType:                   appData.Localizer.T(provided.Type.String()),
			VoucherFullName:           provided.Voucher.FullName(),
			DonorFirstNamesPossessive: appData.Localizer.Possessive(provided.Donor.FirstNames),
		}); err != nil {
			return err
		}
	} else {
		provided.VoucherCodeSentBySMS = false
		provided.VoucherCodeSentTo = provided.Donor.Email

		if err := s.sendEmail(ctx, notify.ToDonorOnly(provided), provided.LpaUID, notify.VouchingShareCodeEmail{
			ShareCode:       shareCode.Plain(),
			VoucherFullName: provided.Voucher.FullName(),
			DonorFullName:   provided.Donor.FullName(),
			LpaType:         appData.Localizer.T(provided.Type.String()),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sender) sendOriginalAttorney(ctx context.Context, appData appcontext.Data, lpa *lpadata.Lpa, attorney lpadata.Attorney) error {
	shareCode, err := s.createShareCode(ctx, lpa.LpaKey, lpa.LpaOwnerKey, lpa.LpaUID, attorney.UID, actor.TypeAttorney)
	if err != nil {
		return err
	}

	if err := s.sendAttorneyStarted(ctx, lpa.LpaUID, attorney.UID); err != nil {
		return err
	}

	if attorney.Email == "" {
		return s.sendPaperForm(ctx, lpa.LpaUID, actor.TypeAttorney, attorney.UID, shareCode)
	}

	return s.sendEmail(ctx, notify.ToLpaAttorney(attorney), lpa.LpaUID,
		notify.InitialOriginalAttorneyEmail{
			AttorneyFullName:          attorney.FullName(),
			DonorFirstNames:           lpa.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(lpa.Donor.FirstNames),
			DonorFullName:             lpa.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
			AttorneyStartPageURL:      s.attorneyStartURL,
			ShareCode:                 shareCode.Plain(),
			AttorneyOptOutURL:         s.appPublicURL + page.PathAttorneyEnterReferenceNumberOptOut.Format(),
		})
}

func (s *Sender) sendReplacementAttorney(ctx context.Context, appData appcontext.Data, lpa *lpadata.Lpa, attorney lpadata.Attorney) error {
	shareCode, err := s.createShareCode(ctx, lpa.LpaKey, lpa.LpaOwnerKey, lpa.LpaUID, attorney.UID, actor.TypeReplacementAttorney)
	if err != nil {
		return err
	}

	if err := s.sendAttorneyStarted(ctx, lpa.LpaUID, attorney.UID); err != nil {
		return err
	}

	if attorney.Email == "" {
		return s.sendPaperForm(ctx, lpa.LpaUID, actor.TypeReplacementAttorney, attorney.UID, shareCode)
	}

	return s.sendEmail(ctx, notify.ToLpaAttorney(attorney), lpa.LpaUID,
		notify.InitialReplacementAttorneyEmail{
			AttorneyFullName:          attorney.FullName(),
			DonorFirstNames:           lpa.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(lpa.Donor.FirstNames),
			DonorFullName:             lpa.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
			AttorneyStartPageURL:      s.attorneyStartURL,
			ShareCode:                 shareCode.Plain(),
			AttorneyOptOutURL:         s.appPublicURL + page.PathAttorneyEnterReferenceNumberOptOut.Format(),
		})
}

func (s *Sender) sendTrustCorporation(ctx context.Context, appData appcontext.Data, lpa *lpadata.Lpa, trustCorporation lpadata.TrustCorporation) error {
	if trustCorporation.Name == "" {
		return nil
	}

	shareCode, err := s.createShareCode(ctx, lpa.LpaKey, lpa.LpaOwnerKey, lpa.LpaUID, trustCorporation.UID, actor.TypeTrustCorporation)
	if err != nil {
		return err
	}

	if err := s.sendAttorneyStarted(ctx, lpa.LpaUID, trustCorporation.UID); err != nil {
		return err
	}

	if trustCorporation.Email == "" {
		return s.sendPaperForm(ctx, lpa.LpaUID, actor.TypeTrustCorporation, trustCorporation.UID, shareCode)
	}

	return s.sendEmail(ctx, notify.ToLpaTrustCorporation(trustCorporation), lpa.LpaUID,
		notify.InitialOriginalAttorneyEmail{
			AttorneyFullName:          trustCorporation.Name,
			DonorFirstNames:           lpa.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(lpa.Donor.FirstNames),
			DonorFullName:             lpa.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
			AttorneyStartPageURL:      s.attorneyStartURL,
			ShareCode:                 shareCode.Plain(),
			AttorneyOptOutURL:         s.appPublicURL + page.PathAttorneyEnterReferenceNumberOptOut.Format(),
		})
}

func (s *Sender) sendReplacementTrustCorporation(ctx context.Context, appData appcontext.Data, lpa *lpadata.Lpa, trustCorporation lpadata.TrustCorporation) error {
	if trustCorporation.Name == "" {
		return nil
	}

	shareCode, err := s.createShareCode(ctx, lpa.LpaKey, lpa.LpaOwnerKey, lpa.LpaUID, trustCorporation.UID, actor.TypeReplacementTrustCorporation)
	if err != nil {
		return err
	}

	if err := s.sendAttorneyStarted(ctx, lpa.LpaUID, trustCorporation.UID); err != nil {
		return err
	}

	if trustCorporation.Email == "" {
		return s.sendPaperForm(ctx, lpa.LpaUID, actor.TypeReplacementTrustCorporation, trustCorporation.UID, shareCode)
	}

	return s.sendEmail(ctx, notify.ToLpaTrustCorporation(trustCorporation), lpa.LpaUID,
		notify.InitialReplacementAttorneyEmail{
			AttorneyFullName:          trustCorporation.Name,
			DonorFirstNames:           lpa.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(lpa.Donor.FirstNames),
			DonorFullName:             lpa.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
			AttorneyStartPageURL:      s.attorneyStartURL,
			ShareCode:                 shareCode.Plain(),
			AttorneyOptOutURL:         s.appPublicURL + page.PathAttorneyEnterReferenceNumberOptOut.Format(),
		})
}

func (s *Sender) createShareCode(ctx context.Context, lpaKey dynamo.LpaKeyType, lpaOwnerKey dynamo.LpaOwnerKeyType, lpaUID string, actorUID actoruid.UID, actorType actor.Type) (sharecodedata.PlainText, error) {
	plainCode, hashedCode := s.generate()

	if s.testCode != "" {
		plainCode = sharecodedata.PlainText(s.testCode)
		hashedCode = sharecodedata.HashedFromString(s.testCode)
		s.testCode = ""
	}

	shareCodeData := sharecodedata.Link{
		LpaKey:                lpaKey,
		LpaOwnerKey:           lpaOwnerKey,
		LpaUID:                lpaUID,
		ActorUID:              actorUID,
		IsReplacementAttorney: actorType == actor.TypeReplacementAttorney || actorType == actor.TypeReplacementTrustCorporation,
		IsTrustCorporation:    actorType == actor.TypeTrustCorporation || actorType == actor.TypeReplacementTrustCorporation,
	}

	if err := s.shareCodeStore.Put(ctx, actorType, hashedCode, shareCodeData); err != nil {
		return "", fmt.Errorf("creating share failed: %w", err)
	}

	return plainCode, nil
}

func (s *Sender) sendEmail(ctx context.Context, to notify.ToEmail, lpaUID string, email notify.Email) error {
	if err := s.notifyClient.SendActorEmail(ctx, to, lpaUID, email); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}

func (s *Sender) sendSMS(ctx context.Context, to notify.ToMobile, lpaUID string, sms notify.SMS) error {
	if err := s.notifyClient.SendActorSMS(ctx, to, lpaUID, sms); err != nil {
		return fmt.Errorf("sms failed: %w", err)
	}

	return nil
}

func (s *Sender) sendPaperForm(ctx context.Context, lpaUID string, actorType actor.Type, actorUID actoruid.UID, shareCode sharecodedata.PlainText) error {
	return s.eventClient.SendPaperFormRequested(ctx, event.PaperFormRequested{
		UID:        lpaUID,
		ActorType:  actorType.String(),
		ActorUID:   actorUID,
		AccessCode: shareCode.Plain(),
	})
}

func (s *Sender) sendAttorneyStarted(ctx context.Context, lpaUID string, actorUID actoruid.UID) error {
	return s.eventClient.SendAttorneyStarted(ctx, event.AttorneyStarted{
		LpaUID:   lpaUID,
		ActorUID: actorUID,
	})
}
