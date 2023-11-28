package page

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWitnessCodeSenderSendToCertificateProvider(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("TemplateID", notify.WitnessCodeSMS).
		Return("template-id")
	notifyClient.
		On("Sms", ctx, notify.Sms{
			PhoneNumber: "0777",
			TemplateID:  "template-id",
			Personalisation: map[string]string{
				"WitnessCode":   "1234",
				"DonorFullName": "Joe Jones’",
				"LpaType":       "property and affairs",
			},
		}).
		Return("sms-id", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", ctx, &actor.DonorProvidedDetails{
			Donor:                    actor.Donor{FirstNames: "Joe", LastName: "Jones"},
			CertificateProvider:      actor.CertificateProvider{Mobile: "0777"},
			CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
			Type:                     actor.LpaTypePropertyFinance,
		}).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.
		On("T", "pfaLegalTerm").
		Return("property and affairs")
	localizer.
		On("Possessive", "Joe Jones").
		Return("Joe Jones’")

	sender := &WitnessCodeSender{
		donorStore:   donorStore,
		notifyClient: notifyClient,
		randomCode:   func(int) string { return "1234" },
		now:          func() time.Time { return now },
	}
	err := sender.SendToCertificateProvider(ctx, &actor.DonorProvidedDetails{
		Donor:               actor.Donor{FirstNames: "Joe", LastName: "Jones"},
		CertificateProvider: actor.CertificateProvider{Mobile: "0777"},
		Type:                actor.LpaTypePropertyFinance,
	}, localizer)

	assert.Nil(t, err)
}

func TestWitnessCodeSenderSendToCertificateProviderWhenTooRecentlySent(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	sender := &WitnessCodeSender{now: func() time.Time { return now }}
	err := sender.SendToCertificateProvider(ctx, &actor.DonorProvidedDetails{
		CertificateProviderCodes: actor.WitnessCodes{{Created: now.Add(-time.Minute)}},
	}, nil)

	assert.Equal(t, ErrTooManyWitnessCodeRequests, err)
}

func TestWitnessCodeSenderSendToCertificateProviderWhenNotifyClientErrors(t *testing.T) {
	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("TemplateID", mock.Anything).
		Return("template-id")
	notifyClient.
		On("Sms", mock.Anything, mock.Anything).
		Return("", ExpectedError)

	localizer := newMockLocalizer(t)
	localizer.
		On("T", "pfaLegalTerm").
		Return("property and affairs")
	localizer.
		On("Possessive", "Joe Jones").
		Return("Joe Jones’")

	sender := &WitnessCodeSender{
		notifyClient: notifyClient,
		randomCode:   func(int) string { return "1234" },
		now:          time.Now,
	}
	err := sender.SendToCertificateProvider(context.Background(), &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{Mobile: "0777"},
		Donor:               actor.Donor{FirstNames: "Joe", LastName: "Jones"},
		Type:                actor.LpaTypePropertyFinance,
	}, localizer)

	assert.Equal(t, ExpectedError, err)
}

func TestWitnessCodeSenderSendToCertificateProviderWhenDonorStoreErrors(t *testing.T) {
	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("TemplateID", mock.Anything).
		Return("template-id")
	notifyClient.
		On("Sms", mock.Anything, mock.Anything).
		Return("sms-id", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", mock.Anything, mock.Anything).
		Return(ExpectedError)

	localizer := newMockLocalizer(t)
	localizer.
		On("T", "pfaLegalTerm").
		Return("property and affairs")
	localizer.
		On("Possessive", "Joe Jones").
		Return("Joe Jones’")

	sender := &WitnessCodeSender{
		donorStore:   donorStore,
		notifyClient: notifyClient,
		randomCode:   func(int) string { return "1234" },
		now:          time.Now,
	}
	err := sender.SendToCertificateProvider(context.Background(), &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{Mobile: "0777"},
		Donor:               actor.Donor{FirstNames: "Joe", LastName: "Jones"},
		Type:                actor.LpaTypePropertyFinance,
	}, localizer)

	assert.Equal(t, ExpectedError, err)
}

func TestWitnessCodeSenderSendToIndependentWitness(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("TemplateID", notify.WitnessCodeSMS).
		Return("template-id")
	notifyClient.
		On("Sms", ctx, notify.Sms{
			PhoneNumber: "0777",
			TemplateID:  "template-id",
			Personalisation: map[string]string{
				"WitnessCode":   "1234",
				"DonorFullName": "Joe Jones’",
				"LpaType":       "property and affairs",
			},
		}).
		Return("sms-id", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", ctx, &actor.DonorProvidedDetails{
			Donor:                   actor.Donor{FirstNames: "Joe", LastName: "Jones"},
			IndependentWitness:      actor.IndependentWitness{Mobile: "0777"},
			IndependentWitnessCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
			Type:                    actor.LpaTypePropertyFinance,
		}).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.
		On("T", "pfaLegalTerm").
		Return("property and affairs")
	localizer.
		On("Possessive", "Joe Jones").
		Return("Joe Jones’")

	sender := &WitnessCodeSender{
		donorStore:   donorStore,
		notifyClient: notifyClient,
		randomCode:   func(int) string { return "1234" },
		now:          func() time.Time { return now },
	}
	err := sender.SendToIndependentWitness(ctx, &actor.DonorProvidedDetails{
		Donor:              actor.Donor{FirstNames: "Joe", LastName: "Jones"},
		IndependentWitness: actor.IndependentWitness{Mobile: "0777"},
		Type:               actor.LpaTypePropertyFinance,
	}, localizer)

	assert.Nil(t, err)
}

func TestWitnessCodeSenderSendToIndependentWitnessWhenTooRecentlySent(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	sender := &WitnessCodeSender{now: func() time.Time { return now }}
	err := sender.SendToIndependentWitness(ctx, &actor.DonorProvidedDetails{
		IndependentWitnessCodes: actor.WitnessCodes{{Created: now.Add(-time.Minute)}},
	}, nil)

	assert.Equal(t, ErrTooManyWitnessCodeRequests, err)
}

func TestWitnessCodeSenderSendToIndependentWitnessWhenNotifyClientErrors(t *testing.T) {
	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("TemplateID", mock.Anything).
		Return("template-id")
	notifyClient.
		On("Sms", mock.Anything, mock.Anything).
		Return("", ExpectedError)

	localizer := newMockLocalizer(t)
	localizer.
		On("T", "pfaLegalTerm").
		Return("property and affairs")
	localizer.
		On("Possessive", "Joe Jones").
		Return("Joe Jones’")

	sender := &WitnessCodeSender{
		notifyClient: notifyClient,
		randomCode:   func(int) string { return "1234" },
		now:          time.Now,
	}
	err := sender.SendToIndependentWitness(context.Background(), &actor.DonorProvidedDetails{
		IndependentWitness: actor.IndependentWitness{Mobile: "0777"},
		Donor:              actor.Donor{FirstNames: "Joe", LastName: "Jones"},
		Type:               actor.LpaTypePropertyFinance,
	}, localizer)

	assert.Equal(t, ExpectedError, err)
}

func TestWitnessCodeSenderSendToIndependentWitnessWhenDonorStoreErrors(t *testing.T) {
	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("TemplateID", mock.Anything).
		Return("template-id")
	notifyClient.
		On("Sms", mock.Anything, mock.Anything).
		Return("sms-id", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", mock.Anything, mock.Anything).
		Return(ExpectedError)

	localizer := newMockLocalizer(t)
	localizer.
		On("T", "pfaLegalTerm").
		Return("property and affairs")
	localizer.
		On("Possessive", "Joe Jones").
		Return("Joe Jones’")

	sender := &WitnessCodeSender{
		donorStore:   donorStore,
		notifyClient: notifyClient,
		randomCode:   func(int) string { return "1234" },
		now:          time.Now,
	}
	err := sender.SendToIndependentWitness(context.Background(), &actor.DonorProvidedDetails{
		IndependentWitness: actor.IndependentWitness{Mobile: "0777"},
		Donor:              actor.Donor{FirstNames: "Joe", LastName: "Jones"},
		Type:               actor.LpaTypePropertyFinance,
	}, localizer)

	assert.Equal(t, ExpectedError, err)
}
