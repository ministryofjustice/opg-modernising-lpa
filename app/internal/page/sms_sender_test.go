package page

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPaperCertificateProviderMeetingPrompt(t *testing.T) {
	lpa := &Lpa{
		Donor:               actor.Donor{FirstNames: "Teneil", LastName: "Throssell"},
		Type:                LpaTypePropertyFinance,
		CertificateProvider: actor.CertificateProvider{Mobile: "07700900000"},
	}

	ctx := context.Background()

	localizer := newMockLocalizer(t)
	localizer.
		On("T", lpa.Type.LegalTermTransKey()).
		Return("property and affairs")

	TestAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("TemplateID", notify.TemplateId(99)).
		Return("template-id")
	notifyClient.
		On("Sms", ctx, notify.Sms{
			PhoneNumber: lpa.CertificateProvider.Mobile,
			TemplateID:  "template-id",
			Personalisation: map[string]string{
				"donorFullName":   "Teneil Throssell",
				"lpaType":         "property and affairs",
				"donorFirstNames": "Teneil",
			},
		}).
		Return("", nil)

	sender := NewSMSSender(notifyClient)
	err := sender.PaperCertificateProviderMeetingPrompt(ctx, lpa, TestAppData, notify.TemplateId(99))

	assert.Nil(t, err)
}

func TestPaperCertificateProviderMeetingPromptWhenSmsError(t *testing.T) {
	localizer := newMockLocalizer(t)
	localizer.
		On("T", mock.Anything).
		Return("property and affairs")

	TestAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("TemplateID", mock.Anything).
		Return("template-id")
	notifyClient.
		On("Sms", mock.Anything, mock.Anything).
		Return("", expectedError)

	sender := NewSMSSender(notifyClient)
	err := sender.PaperCertificateProviderMeetingPrompt(context.Background(), &Lpa{}, TestAppData, notify.TemplateId(99))

	assert.Equal(t, expectedError, err)
}
