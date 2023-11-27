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

func (s *ShareCodeSender) SendCertificateProvider(ctx context.Context, template notify.Template, appData AppData, identity bool, donor *actor.DonorProvidedDetails) error {
	shareCode := s.randomString(12)
	if s.useTestCode {
		shareCode = "abcdef123456"
		s.useTestCode = false
	}

	if err := s.shareCodeStore.Put(ctx, actor.TypeCertificateProvider, shareCode, actor.ShareCodeData{
		LpaID:           appData.LpaID,
		Identity:        identity,
		DonorFullname:   donor.Donor.FullName(),
		DonorFirstNames: donor.Donor.FirstNames,
		SessionID:       appData.SessionID,
	}); err != nil {
		return fmt.Errorf("creating sharecode failed: %w", err)
	}

	personalisation := map[string]string{
		"cpFullName":                  donor.CertificateProvider.FullName(),
		"donorFullName":               donor.Donor.FullName(),
		"lpaLegalTerm":                appData.Localizer.T(donor.Type.LegalTermTransKey()),
		"certificateProviderStartURL": fmt.Sprintf("%s%s", s.appPublicURL, Paths.CertificateProviderStart),
		"shareCode":                   shareCode,
	}

	if template == notify.CertificateProviderInviteEmail {
		personalisation["donorFirstNames"] = donor.Donor.FirstNames
		personalisation["donorFirstNamesPossessive"] = appData.Localizer.Possessive(donor.Donor.FirstNames)
		personalisation["whatLPACovers"] = appData.Localizer.T(donor.Type.WhatLPACoversTransKey())
	}

	if _, err := s.notifyClient.Email(ctx, notify.Email{
		TemplateID:      s.notifyClient.TemplateID(template),
		EmailAddress:    donor.CertificateProvider.Email,
		Personalisation: personalisation,
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}

func (s *ShareCodeSender) SendAttorneys(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails) error {
	if err := s.sendTrustCorporation(ctx, notify.TrustCorporationInviteEmail, appData, donor, donor.Attorneys.TrustCorporation, false); err != nil {
		return err
	}
	if err := s.sendTrustCorporation(ctx, notify.ReplacementTrustCorporationInviteEmail, appData, donor, donor.ReplacementAttorneys.TrustCorporation, true); err != nil {
		return err
	}

	for _, attorney := range donor.Attorneys.Attorneys {
		if err := s.sendAttorney(ctx, notify.AttorneyInviteEmail, appData, donor, attorney, false); err != nil {
			return err
		}
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		if err := s.sendAttorney(ctx, notify.ReplacementAttorneyInviteEmail, appData, donor, attorney, true); err != nil {
			return err
		}
	}

	return nil
}

func (s *ShareCodeSender) sendAttorney(ctx context.Context, template notify.Template, appData AppData, donor *actor.DonorProvidedDetails, attorney actor.Attorney, isReplacement bool) error {
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

	if _, err := s.notifyClient.Email(ctx, notify.Email{
		TemplateID:   s.notifyClient.TemplateID(template),
		EmailAddress: attorney.Email,
		Personalisation: map[string]string{
			"shareCode":                 shareCode,
			"attorneyFullName":          attorney.FullName(),
			"donorFirstNames":           donor.Donor.FirstNames,
			"donorFirstNamesPossessive": appData.Localizer.Possessive(donor.Donor.FirstNames),
			"donorFullName":             donor.Donor.FullName(),
			"lpaLegalTerm":              appData.Localizer.T(donor.Type.LegalTermTransKey()),
			"landingPageLink":           fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
		},
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}

func (s *ShareCodeSender) sendTrustCorporation(ctx context.Context, template notify.Template, appData AppData, donor *actor.DonorProvidedDetails, trustCorporation actor.TrustCorporation, isReplacement bool) error {
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

	if _, err := s.notifyClient.Email(ctx, notify.Email{
		TemplateID:   s.notifyClient.TemplateID(template),
		EmailAddress: trustCorporation.Email,
		Personalisation: map[string]string{
			"shareCode":                 shareCode,
			"attorneyFullName":          trustCorporation.Name,
			"donorFirstNames":           donor.Donor.FirstNames,
			"donorFirstNamesPossessive": appData.Localizer.Possessive(donor.Donor.FirstNames),
			"donorFullName":             donor.Donor.FullName(),
			"lpaLegalTerm":              appData.Localizer.T(donor.Type.LegalTermTransKey()),
			"landingPageLink":           fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
		},
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}
