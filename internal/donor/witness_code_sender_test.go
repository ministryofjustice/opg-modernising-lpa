package donor

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
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
				SendActorSMS(ctx, localize.En, "0777", "lpa-uid", notify.WitnessCodeSMS{
					WitnessCode:   tc.expectedWitnessCode,
					DonorFullName: "Joe Jones’",
					LpaType:       "property and affairs",
				}).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(ctx, &donordata.Provided{
					LpaUID:                   "lpa-uid",
					Donor:                    donordata.Donor{FirstNames: "Joe", LastName: "Jones"},
					CertificateProvider:      donordata.CertificateProvider{Mobile: "0777"},
					CertificateProviderCodes: donordata.WitnessCodes{{Code: tc.expectedWitnessCode, Created: now}},
					Type:                     lpadata.LpaTypePropertyAndAffairs,
				}).
				Return(nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				GetAny(ctx).
				Return(nil, dynamo.NotFoundError{})

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("property-and-affairs").
				Return("property and affairs")
			localizer.EXPECT().
				Possessive("Joe Jones").
				Return("Joe Jones’")

			sender := &WitnessCodeSender{
				donorStore:               donorStore,
				certificateProviderStore: certificateProviderStore,
				notifyClient:             notifyClient,
				localizer:                localizer,
				randomCode:               func(int) string { return tc.randomCode },
				now:                      func() time.Time { return now },
			}
			err := sender.SendToCertificateProvider(ctx, &donordata.Provided{
				LpaUID:              "lpa-uid",
				Donor:               donordata.Donor{FirstNames: "Joe", LastName: "Jones"},
				CertificateProvider: donordata.CertificateProvider{Mobile: "0777"},
				Type:                lpadata.LpaTypePropertyAndAffairs,
			})

			assert.Nil(t, err)
		})
	}
}

func TestWitnessCodeSenderSendToCertificateProviderWhenContactLanguagePreference(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorSMS(ctx, localize.Cy, "0777", "lpa-uid", notify.WitnessCodeSMS{
			WitnessCode:   "1234",
			DonorFullName: "Joe Jones’",
			LpaType:       "property and affairs",
		}).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(ctx, &donordata.Provided{
			LpaUID:                   "lpa-uid",
			Donor:                    donordata.Donor{FirstNames: "Joe", LastName: "Jones"},
			CertificateProvider:      donordata.CertificateProvider{Mobile: "0777"},
			CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: now}},
			Type:                     lpadata.LpaTypePropertyAndAffairs,
		}).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(ctx).
		Return(&certificateproviderdata.Provided{ContactLanguagePreference: localize.Cy}, nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("property-and-affairs").
		Return("property and affairs")
	localizer.EXPECT().
		Possessive("Joe Jones").
		Return("Joe Jones’")

	sender := &WitnessCodeSender{
		donorStore:               donorStore,
		certificateProviderStore: certificateProviderStore,
		notifyClient:             notifyClient,
		localizer:                localizer,
		randomCode:               func(int) string { return "1234" },
		now:                      func() time.Time { return now },
	}
	err := sender.SendToCertificateProvider(ctx, &donordata.Provided{
		LpaUID:              "lpa-uid",
		Donor:               donordata.Donor{FirstNames: "Joe", LastName: "Jones"},
		CertificateProvider: donordata.CertificateProvider{Mobile: "0777"},
		Type:                lpadata.LpaTypePropertyAndAffairs,
	})

	assert.Nil(t, err)
}

func TestWitnessCodeSenderSendToCertificateProviderWhenTooRecentlySent(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	sender := &WitnessCodeSender{now: func() time.Time { return now }}
	err := sender.SendToCertificateProvider(ctx, &donordata.Provided{
		CertificateProviderCodes: donordata.WitnessCodes{{Created: now.Add(-time.Minute)}},
	})

	assert.Equal(t, ErrTooManyWitnessCodeRequests, err)
}

func TestWitnessCodeSenderSendToCertificateProviderWhenNotifyClientErrors(t *testing.T) {
	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorSMS(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(ctx).
		Return(nil, dynamo.NotFoundError{})

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("property-and-affairs").
		Return("property and affairs")
	localizer.EXPECT().
		Possessive("Joe Jones").
		Return("Joe Jones’")

	sender := &WitnessCodeSender{
		donorStore:               donorStore,
		certificateProviderStore: certificateProviderStore,
		notifyClient:             notifyClient,
		localizer:                localizer,
		randomCode:               func(int) string { return "1234" },
		now:                      time.Now,
	}
	err := sender.SendToCertificateProvider(context.Background(), &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{Mobile: "0777"},
		Donor:               donordata.Donor{FirstNames: "Joe", LastName: "Jones"},
		Type:                lpadata.LpaTypePropertyAndAffairs,
	})

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
	err := sender.SendToCertificateProvider(context.Background(), &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{Mobile: "0777"},
		Donor:               donordata.Donor{FirstNames: "Joe", LastName: "Jones"},
		Type:                lpadata.LpaTypePropertyAndAffairs,
	})

	assert.Equal(t, expectedError, err)
}

func TestWitnessCodeSenderSendToIndependentWitness(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorSMS(ctx, localize.En, "0777", "lpa-uid", notify.WitnessCodeSMS{
			WitnessCode:   "1234",
			DonorFullName: "Joe Jones’",
			LpaType:       "property and affairs",
		}).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(ctx, &donordata.Provided{
			LpaUID:                  "lpa-uid",
			Donor:                   donordata.Donor{FirstNames: "Joe", LastName: "Jones"},
			IndependentWitness:      donordata.IndependentWitness{Mobile: "0777"},
			IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: now}},
			Type:                    lpadata.LpaTypePropertyAndAffairs,
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
		localizer:    localizer,
		randomCode:   func(int) string { return "1234" },
		now:          func() time.Time { return now },
	}
	err := sender.SendToIndependentWitness(ctx, &donordata.Provided{
		LpaUID:             "lpa-uid",
		Donor:              donordata.Donor{FirstNames: "Joe", LastName: "Jones"},
		IndependentWitness: donordata.IndependentWitness{Mobile: "0777"},
		Type:               lpadata.LpaTypePropertyAndAffairs,
	})

	assert.Nil(t, err)
}

func TestWitnessCodeSenderSendToIndependentWitnessWhenTooRecentlySent(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	sender := &WitnessCodeSender{now: func() time.Time { return now }}
	err := sender.SendToIndependentWitness(ctx, &donordata.Provided{
		IndependentWitnessCodes: donordata.WitnessCodes{{Created: now.Add(-time.Minute)}},
	})

	assert.Equal(t, ErrTooManyWitnessCodeRequests, err)
}

func TestWitnessCodeSenderSendToIndependentWitnessWhenNotifyClientErrors(t *testing.T) {
	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorSMS(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
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
		localizer:    localizer,
		randomCode:   func(int) string { return "1234" },
		now:          time.Now,
	}
	err := sender.SendToIndependentWitness(context.Background(), &donordata.Provided{
		IndependentWitness: donordata.IndependentWitness{Mobile: "0777"},
		Donor:              donordata.Donor{FirstNames: "Joe", LastName: "Jones"},
		Type:               lpadata.LpaTypePropertyAndAffairs,
	})

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
	err := sender.SendToIndependentWitness(context.Background(), &donordata.Provided{
		IndependentWitness: donordata.IndependentWitness{Mobile: "0777"},
		Donor:              donordata.Donor{FirstNames: "Joe", LastName: "Jones"},
		Type:               lpadata.LpaTypePropertyAndAffairs,
	})

	assert.Equal(t, expectedError, err)
}
