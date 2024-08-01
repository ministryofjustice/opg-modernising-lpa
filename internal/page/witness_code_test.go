package page

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWitnessCodeSenderSendToCertificateProvider(t *testing.T) {
	testCases := map[string]struct {
		randomCode          string
		expectedWitnessCode string
		useTestCode         bool
	}{
		"random code": {
			randomCode:          "4321",
			expectedWitnessCode: "4321",
		},
		"test code": {
			randomCode:          "4321",
			expectedWitnessCode: "1234",
			useTestCode:         true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			now := time.Now()
			ctx := context.Background()
			UseTestWitnessCode = tc.useTestCode

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorSMS(ctx, "0777", "lpa-uid", notify.WitnessCodeSMS{
					WitnessCode:   tc.expectedWitnessCode,
					DonorFullName: "Joe Jones’",
					LpaType:       "property and affairs",
				}).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(ctx, &actor.DonorProvidedDetails{
					LpaUID:                   "lpa-uid",
					Donor:                    actor.Donor{FirstNames: "Joe", LastName: "Jones"},
					CertificateProvider:      donordata.CertificateProvider{Mobile: "0777"},
					CertificateProviderCodes: actor.WitnessCodes{{Code: tc.expectedWitnessCode, Created: now}},
					Type:                     actor.LpaTypePropertyAndAffairs,
				}).
				Return(nil)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("property-and-affairs").
				Return("property and affairs")
			localizer.EXPECT().
				Possessive("Joe Jones").
				Return("Joe Jones’")

			sender := &WitnessCodeSender{
				donorStore:   donorStore,
				notifyClient: notifyClient,
				randomCode:   func(int) string { return tc.randomCode },
				now:          func() time.Time { return now },
			}
			err := sender.SendToCertificateProvider(ctx, &actor.DonorProvidedDetails{
				LpaUID:              "lpa-uid",
				Donor:               actor.Donor{FirstNames: "Joe", LastName: "Jones"},
				CertificateProvider: donordata.CertificateProvider{Mobile: "0777"},
				Type:                actor.LpaTypePropertyAndAffairs,
			}, localizer)

			assert.Nil(t, err)
		})
	}
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
	notifyClient.EXPECT().
		SendActorSMS(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("property-and-affairs").
		Return("property and affairs")
	localizer.EXPECT().
		Possessive("Joe Jones").
		Return("Joe Jones’")

	sender := &WitnessCodeSender{
		donorStore:   donorStore,
		notifyClient: notifyClient,
		randomCode:   func(int) string { return "1234" },
		now:          time.Now,
	}
	err := sender.SendToCertificateProvider(context.Background(), &actor.DonorProvidedDetails{
		CertificateProvider: donordata.CertificateProvider{Mobile: "0777"},
		Donor:               actor.Donor{FirstNames: "Joe", LastName: "Jones"},
		Type:                actor.LpaTypePropertyAndAffairs,
	}, localizer)

	assert.Equal(t, expectedError, err)
}

func TestWitnessCodeSenderSendToCertificateProviderWhenDonorStoreErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	sender := &WitnessCodeSender{
		donorStore: donorStore,
		randomCode: func(int) string { return "1234" },
		now:        time.Now,
	}
	err := sender.SendToCertificateProvider(context.Background(), &actor.DonorProvidedDetails{
		CertificateProvider: donordata.CertificateProvider{Mobile: "0777"},
		Donor:               actor.Donor{FirstNames: "Joe", LastName: "Jones"},
		Type:                actor.LpaTypePropertyAndAffairs,
	}, nil)

	assert.Equal(t, expectedError, err)
}

func TestWitnessCodeSenderSendToIndependentWitness(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorSMS(ctx, "0777", "lpa-uid", notify.WitnessCodeSMS{
			WitnessCode:   "1234",
			DonorFullName: "Joe Jones’",
			LpaType:       "property and affairs",
		}).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(ctx, &actor.DonorProvidedDetails{
			LpaUID:                  "lpa-uid",
			Donor:                   actor.Donor{FirstNames: "Joe", LastName: "Jones"},
			IndependentWitness:      actor.IndependentWitness{Mobile: "0777"},
			IndependentWitnessCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
			Type:                    actor.LpaTypePropertyAndAffairs,
		}).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("property-and-affairs").
		Return("property and affairs")
	localizer.EXPECT().
		Possessive("Joe Jones").
		Return("Joe Jones’")

	sender := &WitnessCodeSender{
		donorStore:   donorStore,
		notifyClient: notifyClient,
		randomCode:   func(int) string { return "1234" },
		now:          func() time.Time { return now },
	}
	err := sender.SendToIndependentWitness(ctx, &actor.DonorProvidedDetails{
		LpaUID:             "lpa-uid",
		Donor:              actor.Donor{FirstNames: "Joe", LastName: "Jones"},
		IndependentWitness: actor.IndependentWitness{Mobile: "0777"},
		Type:               actor.LpaTypePropertyAndAffairs,
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
	notifyClient.EXPECT().
		SendActorSMS(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("property-and-affairs").
		Return("property and affairs")
	localizer.EXPECT().
		Possessive("Joe Jones").
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
		Type:               actor.LpaTypePropertyAndAffairs,
	}, localizer)

	assert.Equal(t, expectedError, err)
}

func TestWitnessCodeSenderSendToIndependentWitnessWhenDonorStoreErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	sender := &WitnessCodeSender{
		donorStore: donorStore,
		randomCode: func(int) string { return "1234" },
		now:        time.Now,
	}
	err := sender.SendToIndependentWitness(context.Background(), &actor.DonorProvidedDetails{
		IndependentWitness: actor.IndependentWitness{Mobile: "0777"},
		Donor:              actor.Donor{FirstNames: "Joe", LastName: "Jones"},
		Type:               actor.LpaTypePropertyAndAffairs,
	}, nil)

	assert.Equal(t, expectedError, err)
}
