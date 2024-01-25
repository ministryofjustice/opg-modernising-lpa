package page

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
)

type shareCodeEmail interface {
	WithShareCode(string) notify.Email
}

type ShareCodeSender struct {
	useTestCode    bool
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

func (s *ShareCodeSender) UseTestCode() {
	s.useTestCode = true
}

func (s *ShareCodeSender) SendCertificateProviderInvite(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails) error {
	id, err := s.sendCertificateProvider(ctx, appData, donor, notify.CertificateProviderInviteEmail{
		CertificateProviderFullName: donor.CertificateProvider.FullName(),
		DonorFullName:               donor.Donor.FullName(),
		LpaType:                     localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
		CertificateProviderStartURL: fmt.Sprintf("%s%s", s.appPublicURL, Paths.CertificateProviderStart),
		DonorFirstNames:             donor.Donor.FirstNames,
		DonorFirstNamesPossessive:   appData.Localizer.Possessive(donor.Donor.FirstNames),
		WhatLpaCovers:               appData.Localizer.T(donor.Type.WhatLPACoversTransKey()),
	})

	if err != nil {
		return err
	}

	return s.eventClient.SendNotificationSent(ctx, event.NotificationSent{
		UID:            donor.LpaUID,
		NotificationID: id,
	})
}

func (s *ShareCodeSender) SendCertificateProviderPrompt(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails) error {
	_, err := s.sendCertificateProvider(ctx, appData, donor, notify.CertificateProviderProvideCertificatePromptEmail{
		CertificateProviderFullName: donor.CertificateProvider.FullName(),
		DonorFullName:               donor.Donor.FullName(),
		LpaType:                     localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
		CertificateProviderStartURL: fmt.Sprintf("%s%s", s.appPublicURL, Paths.CertificateProviderStart),
	})

	return err
}

func (s *ShareCodeSender) sendCertificateProvider(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails, email shareCodeEmail) (string, error) {
	shareCode := s.randomString(12)
	if s.useTestCode {
		shareCode = "abcdef123456"
		s.useTestCode = false
	}

	if err := s.shareCodeStore.Put(ctx, actor.TypeCertificateProvider, shareCode, actor.ShareCodeData{
		LpaID:           appData.LpaID,
		DonorFullname:   donor.Donor.FullName(),
		DonorFirstNames: donor.Donor.FirstNames,
		SessionID:       appData.SessionID,
	}); err != nil {
		return "", fmt.Errorf("creating sharecode failed: %w", err)
	}

	id, err := s.notifyClient.SendEmail(ctx, donor.CertificateProvider.Email, email.WithShareCode(shareCode))
	if err != nil {
		return "", fmt.Errorf("email failed: %w", err)
	}

	return id, nil
}

func (s *ShareCodeSender) SendAttorneys(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails) error {
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

func (s *ShareCodeSender) sendOriginalAttorney(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails, attorney actor.Attorney) error {
	if attorney.Email == "" {
		return nil
	}

	return s.sendAttorney(ctx, attorney.Email,
		notify.InitialOriginalAttorneyEmail{
			AttorneyFullName:          attorney.FullName(),
			DonorFirstNames:           donor.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(donor.Donor.FirstNames),
			DonorFullName:             donor.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
			AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
		},
		actor.ShareCodeData{
			SessionID:  appData.SessionID,
			LpaID:      appData.LpaID,
			AttorneyID: attorney.ID,
		})
}

func (s *ShareCodeSender) sendReplacementAttorney(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails, attorney actor.Attorney) error {
	if attorney.Email == "" {
		return nil
	}

	return s.sendAttorney(ctx, attorney.Email,
		notify.InitialReplacementAttorneyEmail{
			AttorneyFullName:          attorney.FullName(),
			DonorFirstNames:           donor.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(donor.Donor.FirstNames),
			DonorFullName:             donor.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
			AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
		}, actor.ShareCodeData{
			SessionID:             appData.SessionID,
			LpaID:                 appData.LpaID,
			AttorneyID:            attorney.ID,
			IsReplacementAttorney: true,
		})
}

func (s *ShareCodeSender) sendTrustCorporation(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails, trustCorporation actor.TrustCorporation) error {
	if trustCorporation.Email == "" {
		return nil
	}

	return s.sendAttorney(ctx, trustCorporation.Email,
		notify.InitialOriginalAttorneyEmail{
			AttorneyFullName:          trustCorporation.Name,
			DonorFirstNames:           donor.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(donor.Donor.FirstNames),
			DonorFullName:             donor.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
			AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
		},
		actor.ShareCodeData{
			SessionID:          appData.SessionID,
			LpaID:              appData.LpaID,
			IsTrustCorporation: true,
		})
}

func (s *ShareCodeSender) sendReplacementTrustCorporation(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails, trustCorporation actor.TrustCorporation) error {
	if trustCorporation.Email == "" {
		return nil
	}

	return s.sendAttorney(ctx, trustCorporation.Email,
		notify.InitialReplacementAttorneyEmail{
			AttorneyFullName:          trustCorporation.Name,
			DonorFirstNames:           donor.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(donor.Donor.FirstNames),
			DonorFullName:             donor.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
			AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
		},
		actor.ShareCodeData{
			SessionID:             appData.SessionID,
			LpaID:                 appData.LpaID,
			IsTrustCorporation:    true,
			IsReplacementAttorney: true,
		})
}

func (s *ShareCodeSender) sendAttorney(ctx context.Context, to string, email shareCodeEmail, shareCodeData actor.ShareCodeData) error {
	shareCode := s.randomString(12)
	if s.useTestCode {
		shareCode = "abcdef123456"
		s.useTestCode = false
	}

	if err := s.shareCodeStore.Put(ctx, actor.TypeAttorney, shareCode, shareCodeData); err != nil {
		return fmt.Errorf("creating attorney share failed: %w", err)
	}

	if _, err := s.notifyClient.SendEmail(ctx, to, email.WithShareCode(shareCode)); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}
