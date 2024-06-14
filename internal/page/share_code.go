package page

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
)

type ShareCodeSender struct {
	testCode       string
	shareCodeStore ShareCodeStore
	notifyClient   NotifyClient
	appPublicURL   string
	randomString   func(int) string
	eventClient    EventClient
}

func NewShareCodeSender(shareCodeStore ShareCodeStore, notifyClient NotifyClient, appPublicURL string, randomString func(int) string, eventClient EventClient) *ShareCodeSender {
	return &ShareCodeSender{
		shareCodeStore: shareCodeStore,
		notifyClient:   notifyClient,
		appPublicURL:   appPublicURL,
		randomString:   randomString,
		eventClient:    eventClient,
	}
}

func (s *ShareCodeSender) UseTestCode(shareCode string) {
	s.testCode = shareCode
}

type CertificateProviderInvite struct {
	LpaKey                      dynamo.LpaKeyType
	LpaOwnerKey                 dynamo.LpaOwnerKeyType
	LpaUID                      string
	Type                        actor.LpaType
	DonorFirstNames             string
	DonorFullName               string
	CertificateProviderUID      actoruid.UID
	CertificateProviderFullName string
	CertificateProviderEmail    string
}

func (s *ShareCodeSender) SendCertificateProviderInvite(ctx context.Context, appData AppData, invite CertificateProviderInvite) error {
	shareCode, err := s.createShareCode(ctx, invite.LpaKey, invite.LpaOwnerKey, invite.CertificateProviderUID, actor.TypeCertificateProvider)
	if err != nil {
		return err
	}

	return s.sendEmail(ctx, invite.CertificateProviderEmail, invite.LpaUID, notify.CertificateProviderInviteEmail{
		CertificateProviderFullName:  invite.CertificateProviderFullName,
		DonorFullName:                invite.DonorFullName,
		LpaType:                      localize.LowerFirst(appData.Localizer.T(invite.Type.String())),
		CertificateProviderStartURL:  fmt.Sprintf("%s%s", s.appPublicURL, Paths.CertificateProviderStart),
		DonorFirstNames:              invite.DonorFirstNames,
		DonorFirstNamesPossessive:    appData.Localizer.Possessive(invite.DonorFirstNames),
		WhatLpaCovers:                appData.Localizer.T(invite.Type.WhatLPACoversTransKey()),
		ShareCode:                    shareCode,
		CertificateProviderOptOutURL: fmt.Sprintf("%s%s", s.appPublicURL, Paths.CertificateProvider.EnterReferenceNumberOptOut),
	})
}

func (s *ShareCodeSender) SendCertificateProviderPrompt(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails) error {
	shareCode, err := s.createShareCode(ctx, donor.PK, donor.SK, donor.CertificateProvider.UID, actor.TypeCertificateProvider)
	if err != nil {
		return err
	}

	if donor.CertificateProvider.CarryOutBy.IsPaper() {
		return s.sendPaperForm(ctx, donor.LpaUID, actor.TypeCertificateProvider, donor.CertificateProvider.UID, shareCode)
	}

	return s.sendEmail(ctx, donor.CertificateProvider.Email, donor.LpaUID, notify.CertificateProviderProvideCertificatePromptEmail{
		CertificateProviderFullName: donor.CertificateProvider.FullName(),
		DonorFullName:               donor.Donor.FullName(),
		LpaType:                     localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
		CertificateProviderStartURL: fmt.Sprintf("%s%s", s.appPublicURL, Paths.CertificateProviderStart),
		ShareCode:                   shareCode,
	})
}

func (s *ShareCodeSender) SendAttorneys(ctx context.Context, appData AppData, donor *lpastore.Lpa) error {
	if err := s.sendTrustCorporation(ctx, appData, donor, donor.Attorneys.TrustCorporation); err != nil {
		return err
	}
	if err := s.sendReplacementTrustCorporation(ctx, appData, donor, donor.ReplacementAttorneys.TrustCorporation); err != nil {
		return err
	}

	for _, attorney := range donor.Attorneys.Attorneys {
		if err := s.sendOriginalAttorney(ctx, appData, donor, attorney); err != nil {
			return err
		}
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		if err := s.sendReplacementAttorney(ctx, appData, donor, attorney); err != nil {
			return err
		}
	}

	return nil
}

func (s *ShareCodeSender) sendOriginalAttorney(ctx context.Context, appData AppData, lpa *lpastore.Lpa, attorney lpastore.Attorney) error {
	shareCode, err := s.createShareCode(ctx, lpa.LpaKey, lpa.LpaOwnerKey, attorney.UID, actor.TypeAttorney)
	if err != nil {
		return err
	}

	if attorney.Email == "" {
		return s.sendPaperForm(ctx, lpa.LpaUID, actor.TypeAttorney, attorney.UID, shareCode)
	}

	return s.sendEmail(ctx, attorney.Email, lpa.LpaUID,
		notify.InitialOriginalAttorneyEmail{
			AttorneyFullName:          attorney.FullName(),
			DonorFirstNames:           lpa.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(lpa.Donor.FirstNames),
			DonorFullName:             lpa.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
			AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
			ShareCode:                 shareCode,
		})
}

func (s *ShareCodeSender) sendReplacementAttorney(ctx context.Context, appData AppData, lpa *lpastore.Lpa, attorney lpastore.Attorney) error {
	shareCode, err := s.createShareCode(ctx, lpa.LpaKey, lpa.LpaOwnerKey, attorney.UID, actor.TypeReplacementAttorney)
	if err != nil {
		return err
	}

	if attorney.Email == "" {
		return s.sendPaperForm(ctx, lpa.LpaUID, actor.TypeReplacementAttorney, attorney.UID, shareCode)
	}

	return s.sendEmail(ctx, attorney.Email, lpa.LpaUID,
		notify.InitialReplacementAttorneyEmail{
			AttorneyFullName:          attorney.FullName(),
			DonorFirstNames:           lpa.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(lpa.Donor.FirstNames),
			DonorFullName:             lpa.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
			AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
			ShareCode:                 shareCode,
		})
}

func (s *ShareCodeSender) sendTrustCorporation(ctx context.Context, appData AppData, lpa *lpastore.Lpa, trustCorporation lpastore.TrustCorporation) error {
	if trustCorporation.Name == "" {
		return nil
	}

	shareCode, err := s.createShareCode(ctx, lpa.LpaKey, lpa.LpaOwnerKey, trustCorporation.UID, actor.TypeTrustCorporation)
	if err != nil {
		return err
	}

	if trustCorporation.Email == "" {
		return s.sendPaperForm(ctx, lpa.LpaUID, actor.TypeTrustCorporation, trustCorporation.UID, shareCode)
	}

	return s.sendEmail(ctx, trustCorporation.Email, lpa.LpaUID,
		notify.InitialOriginalAttorneyEmail{
			AttorneyFullName:          trustCorporation.Name,
			DonorFirstNames:           lpa.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(lpa.Donor.FirstNames),
			DonorFullName:             lpa.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
			AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
			ShareCode:                 shareCode,
		})
}

func (s *ShareCodeSender) sendReplacementTrustCorporation(ctx context.Context, appData AppData, lpa *lpastore.Lpa, trustCorporation lpastore.TrustCorporation) error {
	if trustCorporation.Name == "" {
		return nil
	}

	shareCode, err := s.createShareCode(ctx, lpa.LpaKey, lpa.LpaOwnerKey, trustCorporation.UID, actor.TypeReplacementTrustCorporation)
	if err != nil {
		return err
	}

	if trustCorporation.Email == "" {
		return s.sendPaperForm(ctx, lpa.LpaUID, actor.TypeReplacementTrustCorporation, trustCorporation.UID, shareCode)
	}

	return s.sendEmail(ctx, trustCorporation.Email, lpa.LpaUID,
		notify.InitialReplacementAttorneyEmail{
			AttorneyFullName:          trustCorporation.Name,
			DonorFirstNames:           lpa.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(lpa.Donor.FirstNames),
			DonorFullName:             lpa.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
			AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
			ShareCode:                 shareCode,
		})
}

func (s *ShareCodeSender) createShareCode(ctx context.Context, lpaKey dynamo.LpaKeyType, lpaOwnerKey dynamo.LpaOwnerKeyType, actorUID actoruid.UID, actorType actor.Type) (string, error) {
	shareCode := s.randomString(12)
	if s.testCode != "" {
		shareCode = s.testCode
		s.testCode = ""
	}

	shareCodeData := actor.ShareCodeData{
		LpaKey:                lpaKey,
		LpaOwnerKey:           lpaOwnerKey,
		ActorUID:              actorUID,
		IsReplacementAttorney: actorType == actor.TypeReplacementAttorney || actorType == actor.TypeReplacementTrustCorporation,
		IsTrustCorporation:    actorType == actor.TypeTrustCorporation || actorType == actor.TypeReplacementTrustCorporation,
	}

	if err := s.shareCodeStore.Put(ctx, actorType, shareCode, shareCodeData); err != nil {
		return "", fmt.Errorf("creating share failed: %w", err)
	}

	return shareCode, nil
}

func (s *ShareCodeSender) sendEmail(ctx context.Context, to string, lpaUID string, email notify.Email) error {
	if err := s.notifyClient.SendActorEmail(ctx, to, lpaUID, email); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}

func (s *ShareCodeSender) sendPaperForm(ctx context.Context, lpaUID string, actorType actor.Type, actorUID actoruid.UID, shareCode string) error {
	return s.eventClient.SendPaperFormRequested(ctx, event.PaperFormRequested{
		UID:        lpaUID,
		ActorType:  actorType.String(),
		ActorUID:   actorUID,
		AccessCode: shareCode,
	})
}
