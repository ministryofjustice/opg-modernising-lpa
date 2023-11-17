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
		On("TemplateID", notify.SignatureCodeSMS).
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
		On("Put", ctx, &Lpa{
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
	err := sender.SendToCertificateProvider(ctx, &Lpa{
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
	err := sender.SendToCertificateProvider(ctx, &Lpa{
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
	err := sender.SendToCertificateProvider(context.Background(), &Lpa{
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
	err := sender.SendToCertificateProvider(context.Background(), &Lpa{
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
		On("TemplateID", notify.SignatureCodeSMS).
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
		On("Put", ctx, &Lpa{
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
	err := sender.SendToIndependentWitness(ctx, &Lpa{
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
	err := sender.SendToIndependentWitness(ctx, &Lpa{
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
	err := sender.SendToIndependentWitness(context.Background(), &Lpa{
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
	err := sender.SendToIndependentWitness(context.Background(), &Lpa{
		IndependentWitness: actor.IndependentWitness{Mobile: "0777"},
		Donor:              actor.Donor{FirstNames: "Joe", LastName: "Jones"},
		Type:               actor.LpaTypePropertyFinance,
	}, localizer)

	assert.Equal(t, ExpectedError, err)
}

func TestWitnessCodeHasExpired(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		duration time.Duration
		expected bool
	}{
		"now": {
			duration: 0,
			expected: false,
		},
		"14m59s ago": {
			duration: 14*time.Minute + 59*time.Second,
			expected: false,
		},
		"15m ago": {
			duration: 15 * time.Minute,
			expected: true,
		},
		"15m01s ago": {
			duration: 15*time.Minute + time.Second,
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := Lpa{
				CertificateProviderCodes: actor.WitnessCodes{
					{Code: "a", Created: now.Add(-tc.duration)},
				},
			}

			code, _ := lpa.CertificateProviderCodes.Find("a")
			assert.Equal(t, tc.expected, code.HasExpired())
		})
	}
}

func TestWitnessCodesFind(t *testing.T) {
	codes := actor.WitnessCodes{
		{Code: "new", Created: time.Now()},
		{Code: "expired", Created: time.Now().Add(-16 * time.Minute)},
		{Code: "almost ignored", Created: time.Now().Add(-2*time.Hour + time.Second)},
		{Code: "ignored", Created: time.Now().Add(-2 * time.Hour)},
	}

	testcases := map[string]bool{
		"wrong":          false,
		"new":            true,
		"expired":        true,
		"almost ignored": true,
		"ignored":        false,
	}

	for code, expected := range testcases {
		t.Run(code, func(t *testing.T) {
			_, ok := codes.Find(code)
			assert.Equal(t, expected, ok)
		})
	}
}

func TestWitnessCodesCanRequest(t *testing.T) {
	now := time.Now()

	testcases := map[string]struct {
		codes    actor.WitnessCodes
		expected bool
	}{
		"empty": {
			expected: true,
		},
		"after 1 minute": {
			codes:    actor.WitnessCodes{{Created: now.Add(-time.Minute - time.Second)}},
			expected: true,
		},
		"within 1 minute": {
			codes:    actor.WitnessCodes{{Created: now.Add(-time.Minute)}},
			expected: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.codes.CanRequest(now))
		})
	}
}
