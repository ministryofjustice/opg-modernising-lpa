package page

import (
	"context"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/notify"
)

type SMSSender struct {
	notifyClient NotifyClient
}

func NewSMSSender(client NotifyClient) *SMSSender {
	return &SMSSender{
		notifyClient: client,
	}
}

func (s *SMSSender) PaperCertificateProviderMeetingPrompt(ctx context.Context, lpa *Lpa, appData AppData, template notify.TemplateId) error {
	_, err := s.notifyClient.Sms(ctx, notify.Sms{
		PhoneNumber: lpa.CertificateProvider.Mobile,
		TemplateID:  s.notifyClient.TemplateID(template),
		Personalisation: map[string]string{
			"donorFullName":   lpa.Donor.FullName(),
			"lpaType":         appData.Localizer.T(lpa.Type.LegalTermTransKey()),
			"donorFirstNames": lpa.Donor.FirstNames,
		},
	})

	return err
}
