package page

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/notify"
)

var useTestCode = false

type ShareCodeSender struct {
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
	useTestCode = true
}

func (s *ShareCodeSender) SendCertificateProvider(ctx context.Context, template notify.TemplateId, appData AppData, identity bool, lpa *Lpa) error {
	var shareCode string

	if useTestCode {
		shareCode = "abcdef123456"
		useTestCode = false
	} else {
		shareCode = s.randomString(12)
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

func (s *ShareCodeSender) SendAttorneys(ctx context.Context, appData AppData, lpa *Lpa) error {
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

func (s *ShareCodeSender) sendAttorney(ctx context.Context, template notify.TemplateId, appData AppData, lpa *Lpa, attorney actor.Attorney, isReplacement bool) error {
	if attorney.Email == "" {
		return nil
	}

	var shareCode string

	if useTestCode {
		shareCode = "abcdef123456"
		useTestCode = false
	} else {
		shareCode = s.randomString(12)
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
