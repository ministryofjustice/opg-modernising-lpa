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

func (s *ShareCodeSender) SendCertificateProvider(ctx context.Context, template notify.Template, appData AppData, identity bool, lpa *actor.DonorProvidedDetails) error {
	shareCode := s.randomString(12)
	if s.useTestCode {
		shareCode = "abcdef123456"
		s.useTestCode = false
	}

	if err := s.shareCodeStore.Put(ctx, actor.TypeCertificateProvider, shareCode, actor.ShareCodeData{
		LpaID:           appData.LpaID,
		Identity:        identity,
		DonorFullname:   lpa.Donor.FullName(),
		DonorFirstNames: lpa.Donor.FirstNames,
		SessionID:       appData.SessionID,
	}); err != nil {
		return fmt.Errorf("creating sharecode failed: %w", err)
	}

	if _, err := s.notifyClient.Email(ctx, notify.Email{
		TemplateID:   s.notifyClient.TemplateID(template),
		EmailAddress: lpa.CertificateProvider.Email,
		Personalisation: map[string]string{
			"cpFullName":                  lpa.CertificateProvider.FullName(),
			"donorFullName":               lpa.Donor.FullName(),
			"lpaLegalTerm":                appData.Localizer.T(lpa.Type.LegalTermTransKey()),
			"donorFirstNames":             lpa.Donor.FirstNames,
			"certificateProviderStartURL": fmt.Sprintf("%s%s", s.appPublicURL, Paths.CertificateProviderStart),
			"donorFirstNamesPossessive":   appData.Localizer.Possessive(lpa.Donor.FirstNames),
			"shareCode":                   shareCode,
		},
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}

func (s *ShareCodeSender) SendAttorneys(ctx context.Context, appData AppData, lpa *actor.DonorProvidedDetails) error {
	if err := s.sendTrustCorporation(ctx, notify.TrustCorporationInviteEmail, appData, lpa, lpa.Attorneys.TrustCorporation, false); err != nil {
		return err
	}
	if err := s.sendTrustCorporation(ctx, notify.ReplacementTrustCorporationInviteEmail, appData, lpa, lpa.ReplacementAttorneys.TrustCorporation, true); err != nil {
		return err
	}

	for _, attorney := range lpa.Attorneys.Attorneys {
		if err := s.sendAttorney(ctx, notify.AttorneyInviteEmail, appData, lpa, attorney, false); err != nil {
			return err
		}
	}

	for _, attorney := range lpa.ReplacementAttorneys.Attorneys {
		if err := s.sendAttorney(ctx, notify.ReplacementAttorneyInviteEmail, appData, lpa, attorney, true); err != nil {
			return err
		}
	}

	return nil
}

func (s *ShareCodeSender) sendAttorney(ctx context.Context, template notify.Template, appData AppData, lpa *actor.DonorProvidedDetails, attorney actor.Attorney, isReplacement bool) error {
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
			"donorFirstNames":           lpa.Donor.FirstNames,
			"donorFirstNamesPossessive": appData.Localizer.Possessive(lpa.Donor.FirstNames),
			"donorFullName":             lpa.Donor.FullName(),
			"lpaLegalTerm":              appData.Localizer.T(lpa.Type.LegalTermTransKey()),
			"landingPageLink":           fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
		},
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}

func (s *ShareCodeSender) sendTrustCorporation(ctx context.Context, template notify.Template, appData AppData, lpa *actor.DonorProvidedDetails, trustCorporation actor.TrustCorporation, isReplacement bool) error {
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
			"donorFirstNames":           lpa.Donor.FirstNames,
			"donorFirstNamesPossessive": appData.Localizer.Possessive(lpa.Donor.FirstNames),
			"donorFullName":             lpa.Donor.FullName(),
			"lpaLegalTerm":              appData.Localizer.T(lpa.Type.LegalTermTransKey()),
			"landingPageLink":           fmt.Sprintf("%s%s", s.appPublicURL, Paths.Attorney.Start),
		},
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}
