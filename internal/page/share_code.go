package page

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
)

type ShareCodeSender struct {
	useTestCode    bool
	shareCodeStore ShareCodeStore
	notifyClient   NotifyClient
	appPublicURL   string
	randomString   func(int) string
}

func NewShareCodeSender(shareCodeStore ShareCodeStore, notifyClient NotifyClient, appPublicURL string, randomString func(int) string) *ShareCodeSender {
	return &ShareCodeSender{
		shareCodeStore: shareCodeStore,
		notifyClient:   notifyClient,
		appPublicURL:   appPublicURL,
		randomString:   randomString,
	}
}

func (s *ShareCodeSender) UseTestCode() {
	s.useTestCode = true
}

func (s *ShareCodeSender) SendCertificateProviderInvite(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails) error {
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
		return fmt.Errorf("creating sharecode failed: %w", err)
	}

	if _, err := s.notifyClient.SendEmail(ctx, donor.CertificateProvider.Email, notify.CertificateProviderInviteEmail{
		CertificateProviderFullName: donor.CertificateProvider.FullName(),
		DonorFullName:               donor.Donor.FullName(),
		LpaType:                     appData.Localizer.T(donor.Type.LegalTermTransKey()),
		CertificateProviderStartURL: fmt.Sprintf("%s%s", s.appPublicURL, Paths.CertificateProviderStart),
		ShareCode:                   shareCode,
		DonorFirstNames:             donor.Donor.FirstNames,
		DonorFirstNamesPossessive:   appData.Localizer.Possessive(donor.Donor.FirstNames),
		WhatLpaCovers:               appData.Localizer.T(donor.Type.WhatLPACoversTransKey()),
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}

func (s *ShareCodeSender) SendCertificateProviderPrompt(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails) error {
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
		return fmt.Errorf("creating sharecode failed: %w", err)
	}

	if _, err := s.notifyClient.SendEmail(ctx, donor.CertificateProvider.Email, notify.CertificateProviderProvideCertificatePromptEmail{
		CertificateProviderFullName: donor.CertificateProvider.FullName(),
		DonorFullName:               donor.Donor.FullName(),
		LpaType:                     appData.Localizer.T(donor.Type.LegalTermTransKey()),
		CertificateProviderStartURL: fmt.Sprintf("%s%s", s.appPublicURL, Paths.CertificateProviderStart),
		ShareCode:                   shareCode,
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}

func (s *ShareCodeSender) SendAttorneys(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails) error {
	if err := s.sendTrustCorporation(ctx, appData, donor, donor.Attorneys.TrustCorporation, false); err != nil {
		return err
	}
	if err := s.sendReplacementTrustCorporation(ctx, appData, donor, donor.ReplacementAttorneys.TrustCorporation, true); err != nil {
		return err
	}

	for _, attorney := range donor.Attorneys.Attorneys {
		if err := s.sendAttorney(ctx, appData, donor, attorney, false); err != nil {
			return err
		}
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		if err := s.sendReplacementAttorney(ctx, appData, donor, attorney, true); err != nil {
			return err
		}
	}

	return nil
}

func (s *ShareCodeSender) sendAttorney(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails, attorney actor.Attorney, isReplacement bool) error {
	if attorney.Email == "" {
		return nil
	}

	shareCode := s.randomString(12)
	if s.useTestCode {
		shareCode = "abcdef123456"
		s.useTestCode = false
	}

	if err := s.shareCodeStore.Put(ctx, actor.TypeAttorney, shareCode, actor.ShareCodeData{
		SessionID:             appData.SessionID,
		LpaID:                 appData.LpaID,
		AttorneyID:            attorney.ID,
		IsReplacementAttorney: isReplacement,
	}); err != nil {
		return fmt.Errorf("creating attorney share failed: %w", err)
	}

	if _, err := s.notifyClient.SendEmail(ctx, attorney.Email, notify.InitialOriginalAttorneyEmail{
		ShareCode:                 shareCode,
		AttorneyFullName:          attorney.FullName(),
		DonorFirstNames:           donor.Donor.FirstNames,
		DonorFirstNamesPossessive: appData.Localizer.Possessive(donor.Donor.FirstNames),
		DonorFullName:             donor.Donor.FullName(),
		LpaType:                   appData.Localizer.T(donor.Type.LegalTermTransKey()),
		AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}

func (s *ShareCodeSender) sendReplacementAttorney(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails, attorney actor.Attorney, isReplacement bool) error {
	if attorney.Email == "" {
		return nil
	}

	shareCode := s.randomString(12)
	if s.useTestCode {
		shareCode = "abcdef123456"
		s.useTestCode = false
	}

	if err := s.shareCodeStore.Put(ctx, actor.TypeAttorney, shareCode, actor.ShareCodeData{
		SessionID:             appData.SessionID,
		LpaID:                 appData.LpaID,
		AttorneyID:            attorney.ID,
		IsReplacementAttorney: isReplacement,
	}); err != nil {
		return fmt.Errorf("creating attorney share failed: %w", err)
	}

	if _, err := s.notifyClient.SendEmail(ctx, attorney.Email, notify.InitialReplacementAttorneyEmail{
		ShareCode:                 shareCode,
		AttorneyFullName:          attorney.FullName(),
		DonorFirstNames:           donor.Donor.FirstNames,
		DonorFirstNamesPossessive: appData.Localizer.Possessive(donor.Donor.FirstNames),
		DonorFullName:             donor.Donor.FullName(),
		LpaType:                   appData.Localizer.T(donor.Type.LegalTermTransKey()),
		AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}

func (s *ShareCodeSender) sendTrustCorporation(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails, trustCorporation actor.TrustCorporation, isReplacement bool) error {
	if trustCorporation.Email == "" {
		return nil
	}

	shareCode := s.randomString(12)
	if s.useTestCode {
		shareCode = "abcdef123456"
		s.useTestCode = false
	}

	if err := s.shareCodeStore.Put(ctx, actor.TypeAttorney, shareCode, actor.ShareCodeData{
		SessionID:             appData.SessionID,
		LpaID:                 appData.LpaID,
		IsTrustCorporation:    true,
		IsReplacementAttorney: isReplacement,
	}); err != nil {
		return fmt.Errorf("creating trust corporation share failed: %w", err)
	}

	if _, err := s.notifyClient.SendEmail(ctx, trustCorporation.Email, notify.InitialOriginalAttorneyEmail{
		ShareCode:                 shareCode,
		AttorneyFullName:          trustCorporation.Name,
		DonorFirstNames:           donor.Donor.FirstNames,
		DonorFirstNamesPossessive: appData.Localizer.Possessive(donor.Donor.FirstNames),
		DonorFullName:             donor.Donor.FullName(),
		LpaType:                   appData.Localizer.T(donor.Type.LegalTermTransKey()),
		AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}

func (s *ShareCodeSender) sendReplacementTrustCorporation(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails, trustCorporation actor.TrustCorporation, isReplacement bool) error {
	if trustCorporation.Email == "" {
		return nil
	}

	shareCode := s.randomString(12)
	if s.useTestCode {
		shareCode = "abcdef123456"
		s.useTestCode = false
	}

	if err := s.shareCodeStore.Put(ctx, actor.TypeAttorney, shareCode, actor.ShareCodeData{
		SessionID:             appData.SessionID,
		LpaID:                 appData.LpaID,
		IsTrustCorporation:    true,
		IsReplacementAttorney: isReplacement,
	}); err != nil {
		return fmt.Errorf("creating trust corporation share failed: %w", err)
	}

	if _, err := s.notifyClient.SendEmail(ctx, trustCorporation.Email, notify.InitialReplacementAttorneyEmail{
		ShareCode:                 shareCode,
		AttorneyFullName:          trustCorporation.Name,
		DonorFirstNames:           donor.Donor.FirstNames,
		DonorFirstNamesPossessive: appData.Localizer.Possessive(donor.Donor.FirstNames),
		DonorFullName:             donor.Donor.FullName(),
		LpaType:                   appData.Localizer.T(donor.Type.LegalTermTransKey()),
		AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}
