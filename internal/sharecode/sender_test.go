package sharecode

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	testAppData = appcontext.Data{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
	}
	testStringCode = "123"
	testHashedCode = sharecodedata.HashedFromString(testStringCode)
	testGenerateFn = func() (sharecodedata.PlainText, sharecodedata.Hashed) {
		return sharecodedata.PlainText(testStringCode), testHashedCode
	}
)

func TestShareCodeSenderSendCertificateProviderInvite(t *testing.T) {
	donor := &donordata.Provided{
		PK: dynamo.LpaKey("lpa"),
		SK: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
		},
		Donor: donordata.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:   lpadata.LpaTypePropertyAndAffairs,
		LpaUID: "lpa-uid",
	}

	to := notify.ToCertificateProvider(donor.CertificateProvider)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(donor.Type.String()).
		Return("Property and affairs").
		Once()
	localizer.EXPECT().
		T("whatPropertyAndAffairsCovers").
		Return("houses and stuff").
		Once()
	localizer.EXPECT().
		Possessive("Jan").
		Return("Jan’s")
	testAppData.Localizer = localizer

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeCertificateProvider, testHashedCode, sharecodedata.Link{
			LpaKey:      dynamo.LpaKey("lpa"),
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			LpaUID:      "lpa-uid",
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, to, "lpa-uid", notify.CertificateProviderInviteEmail{
			ShareCode:                    testStringCode,
			CertificateProviderFullName:  "Joanna Jones",
			DonorFirstNames:              "Jan",
			DonorFullName:                "Jan Smith",
			LpaType:                      "property and affairs",
			CertificateProviderStartURL:  "http://example.com/certificate-provider",
			DonorFirstNamesPossessive:    "Jan’s",
			WhatLpaCovers:                "houses and stuff",
			CertificateProviderOptOutURL: fmt.Sprintf("http://app%s", page.PathCertificateProviderEnterReferenceNumberOptOut),
		}).
		Return(nil)

	sender := &Sender{
		shareCodeStore:              shareCodeStore,
		notifyClient:                notifyClient,
		appPublicURL:                "http://app",
		certificateProviderStartURL: "http://example.com/certificate-provider",
		generate:                    testGenerateFn,
	}
	err := sender.SendCertificateProviderInvite(ctx, testAppData, donor)

	assert.Nil(t, err)
}

func TestShareCodeSenderSendCertificateProviderInviteWithTestCode(t *testing.T) {
	testcases := map[string]struct {
		useTestCode      bool
		expectedTestCode string
	}{
		"with test code": {
			useTestCode:      true,
			expectedTestCode: "abcdef123456",
		},
		"without test code": {
			useTestCode:      false,
			expectedTestCode: testStringCode,
		},
	}

	donor := &donordata.Provided{
		PK: dynamo.LpaKey("lpa"),
		SK: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
		},
		Donor: donordata.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:   lpadata.LpaTypePersonalWelfare,
		LpaUID: "lpa-uid",
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T(donor.Type.String()).
				Return("Personal welfare").
				Twice()
			localizer.EXPECT().
				Possessive("Jan").
				Return("Jan’s")
			localizer.EXPECT().
				T("whatPersonalWelfareCovers").
				Return("health and stuff")
			testAppData.Localizer = localizer

			to := notify.ToCertificateProvider(donor.CertificateProvider)
			ctx := context.Background()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, sharecodedata.HashedFromString(tc.expectedTestCode), sharecodedata.Link{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
					LpaUID:      "lpa-uid",
				}).
				Once().
				Return(nil)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, testHashedCode, sharecodedata.Link{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
					LpaUID:      "lpa-uid",
				}).
				Once().
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(ctx, to, "lpa-uid", notify.CertificateProviderInviteEmail{
					CertificateProviderFullName:  "Joanna Jones",
					DonorFirstNames:              "Jan",
					DonorFullName:                "Jan Smith",
					LpaType:                      "personal welfare",
					CertificateProviderStartURL:  "http://example.com/certificate-provider",
					ShareCode:                    tc.expectedTestCode,
					DonorFirstNamesPossessive:    "Jan’s",
					WhatLpaCovers:                "health and stuff",
					CertificateProviderOptOutURL: fmt.Sprintf("http://app%s", page.PathCertificateProviderEnterReferenceNumberOptOut),
				}).
				Once().
				Return(nil)
			notifyClient.EXPECT().
				SendActorEmail(ctx, to, "lpa-uid", notify.CertificateProviderInviteEmail{
					CertificateProviderFullName:  "Joanna Jones",
					DonorFirstNames:              "Jan",
					DonorFullName:                "Jan Smith",
					LpaType:                      "personal welfare",
					CertificateProviderStartURL:  "http://example.com/certificate-provider",
					ShareCode:                    testStringCode,
					DonorFirstNamesPossessive:    "Jan’s",
					WhatLpaCovers:                "health and stuff",
					CertificateProviderOptOutURL: fmt.Sprintf("http://app%s", page.PathCertificateProviderEnterReferenceNumberOptOut),
				}).
				Once().
				Return(nil)

			sender := &Sender{
				shareCodeStore:              shareCodeStore,
				notifyClient:                notifyClient,
				appPublicURL:                "http://app",
				certificateProviderStartURL: "http://example.com/certificate-provider",
				generate:                    testGenerateFn,
			}

			if tc.useTestCode {
				sender.UseTestCode("abcdef123456")
			}

			err := sender.SendCertificateProviderInvite(ctx, testAppData, donor)
			assert.Nil(t, err)

			err = sender.SendCertificateProviderInvite(ctx, testAppData, donor)
			assert.Nil(t, err)
		})
	}
}

func TestShareCodeSenderSendCertificateProviderInviteWhenEmailErrors(t *testing.T) {
	ctx := context.Background()

	donor := &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
		},
		Donor: donordata.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type: lpadata.LpaTypePropertyAndAffairs,
	}

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("")
	localizer.EXPECT().
		Possessive("Jan").
		Return("Jan’s")
	testAppData.Localizer = localizer

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := &Sender{
		shareCodeStore: shareCodeStore,
		notifyClient:   notifyClient,
		generate:       testGenerateFn,
	}
	err := sender.SendCertificateProviderInvite(ctx, testAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderInviteWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := &Sender{
		shareCodeStore: shareCodeStore,
		generate:       testGenerateFn,
	}
	err := sender.SendCertificateProviderInvite(ctx, testAppData, &donordata.Provided{})

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderPromptOnline(t *testing.T) {
	donor := &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
			CarryOutBy: lpadata.ChannelOnline,
		},
		Donor: donordata.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:   lpadata.LpaTypePropertyAndAffairs,
		LpaUID: "lpa-uid",
		PK:     dynamo.LpaKey("lpa"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(donor.Type.String()).
		Return("Property and affairs").
		Once()
	testAppData.Localizer = localizer

	ctx := context.Background()

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToCertificateProvider(donor.CertificateProvider), "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
			ShareCode:                   testStringCode,
			CertificateProviderFullName: "Joanna Jones",
			DonorFullName:               "Jan Smith",
			LpaType:                     "property and affairs",
			CertificateProviderStartURL: "http://example.com/certificate-provider",
		}).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeCertificateProvider, testHashedCode, sharecodedata.Link{
			LpaKey:      dynamo.LpaKey("lpa"),
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			LpaUID:      "lpa-uid",
		}).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(ctx).
		Return(nil, expectedError)

	sender := &Sender{
		shareCodeStore:              shareCodeStore,
		notifyClient:                notifyClient,
		certificateProviderStartURL: "http://example.com/certificate-provider",
		generate:                    testGenerateFn,
		certificateProviderStore:    certificateProviderStore,
	}
	err := sender.SendCertificateProviderPrompt(ctx, testAppData, donor)

	assert.Nil(t, err)
}

func TestShareCodeSenderSendCertificateProviderPromptOnlineWhenStarted(t *testing.T) {
	donor := &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
			CarryOutBy: lpadata.ChannelOnline,
		},
		Donor: donordata.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:   lpadata.LpaTypePropertyAndAffairs,
		LpaUID: "lpa-uid",
		PK:     dynamo.LpaKey("lpa"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	certificateProvider := &certificateproviderdata.Provided{
		Email:                     "correct@example.com",
		ContactLanguagePreference: localize.Cy,
	}

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(donor.Type.String()).
		Return("Property and affairs").
		Once()
	testAppData.Localizer = localizer

	ctx := context.Background()

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToProvidedCertificateProvider(certificateProvider, donor.CertificateProvider), "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
			ShareCode:                   testStringCode,
			CertificateProviderFullName: "Joanna Jones",
			DonorFullName:               "Jan Smith",
			LpaType:                     "property and affairs",
			CertificateProviderStartURL: "http://example.com/certificate-provider",
		}).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeCertificateProvider, testHashedCode, sharecodedata.Link{
			LpaKey:      dynamo.LpaKey("lpa"),
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			LpaUID:      "lpa-uid",
		}).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(ctx).
		Return(certificateProvider, nil)

	sender := &Sender{
		shareCodeStore:              shareCodeStore,
		notifyClient:                notifyClient,
		certificateProviderStartURL: "http://example.com/certificate-provider",
		generate:                    testGenerateFn,
		certificateProviderStore:    certificateProviderStore,
	}
	err := sender.SendCertificateProviderPrompt(ctx, testAppData, donor)

	assert.Nil(t, err)
}

func TestShareCodeSenderSendCertificateProviderPromptPaper(t *testing.T) {
	actorUID := actoruid.New()

	donor := &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{
			UID:        actorUID,
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
			CarryOutBy: lpadata.ChannelPaper,
		},
		Donor: donordata.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:   lpadata.LpaTypePropertyAndAffairs,
		LpaUID: "lpa-uid",
		PK:     dynamo.LpaKey("lpa"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeCertificateProvider, testHashedCode, sharecodedata.Link{
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			LpaKey:      dynamo.LpaKey("lpa"),
			LpaUID:      "lpa-uid",
			ActorUID:    actorUID,
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaperFormRequested(ctx, event.PaperFormRequested{
			UID:        "lpa-uid",
			ActorType:  actor.TypeCertificateProvider.String(),
			ActorUID:   actorUID,
			AccessCode: testStringCode,
		}).
		Return(nil)

	sender := &Sender{
		shareCodeStore: shareCodeStore,
		generate:       testGenerateFn,
		eventClient:    eventClient,
	}
	err := sender.SendCertificateProviderPrompt(ctx, testAppData, donor)

	assert.Nil(t, err)
}

func TestShareCodeSenderSendCertificateProviderPromptWithTestCode(t *testing.T) {
	testcases := map[string]struct {
		useTestCode      bool
		expectedTestCode string
	}{
		"with test code": {
			useTestCode:      true,
			expectedTestCode: "abcdef123456",
		},
		"without test code": {
			useTestCode:      false,
			expectedTestCode: testStringCode,
		},
	}

	donor := &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
		},
		Donor: donordata.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:   lpadata.LpaTypePropertyAndAffairs,
		LpaUID: "lpa-uid",
		PK:     dynamo.LpaKey("lpa"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T(donor.Type.String()).
				Return("Property and affairs").
				Twice()

			testAppData.Localizer = localizer

			ctx := context.Background()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, sharecodedata.HashedFromString(tc.expectedTestCode), sharecodedata.Link{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
					LpaUID:      "lpa-uid",
				}).
				Once().
				Return(nil)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, testHashedCode, sharecodedata.Link{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
					LpaUID:      "lpa-uid",
				}).
				Once().
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToCertificateProvider(donor.CertificateProvider), "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
					CertificateProviderFullName: "Joanna Jones",
					DonorFullName:               "Jan Smith",
					LpaType:                     "property and affairs",
					CertificateProviderStartURL: "http://example.com/certificate-provider",
					ShareCode:                   tc.expectedTestCode,
				}).
				Once().
				Return(nil)
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToCertificateProvider(donor.CertificateProvider), "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
					CertificateProviderFullName: "Joanna Jones",
					DonorFullName:               "Jan Smith",
					LpaType:                     "property and affairs",
					CertificateProviderStartURL: "http://example.com/certificate-provider",
					ShareCode:                   testStringCode,
				}).
				Once().
				Return(nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				GetAny(ctx).
				Return(nil, expectedError)

			sender := &Sender{
				shareCodeStore:              shareCodeStore,
				notifyClient:                notifyClient,
				certificateProviderStartURL: "http://example.com/certificate-provider",
				generate:                    testGenerateFn,
				certificateProviderStore:    certificateProviderStore,
			}

			if tc.useTestCode {
				sender.UseTestCode("abcdef123456")
			}

			err := sender.SendCertificateProviderPrompt(ctx, testAppData, donor)
			assert.Nil(t, err)

			err = sender.SendCertificateProviderPrompt(ctx, testAppData, donor)
			assert.Nil(t, err)
		})
	}
}

func TestShareCodeSenderSendCertificateProviderPromptPaperWhenShareCodeStoreError(t *testing.T) {
	donor := &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{
			CarryOutBy: lpadata.ChannelPaper,
		},
	}

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := &Sender{
		shareCodeStore: shareCodeStore,
		generate:       testGenerateFn,
	}
	err := sender.SendCertificateProviderPrompt(ctx, testAppData, donor)

	assert.ErrorIs(t, err, expectedError)
}

func TestShareCodeSenderSendCertificateProviderPromptPaperWhenEventClientError(t *testing.T) {
	donor := &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{
			CarryOutBy: lpadata.ChannelPaper,
		},
	}

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaperFormRequested(mock.Anything, mock.Anything).
		Return(expectedError)

	sender := &Sender{
		shareCodeStore: shareCodeStore,
		generate:       testGenerateFn,
		eventClient:    eventClient,
	}
	err := sender.SendCertificateProviderPrompt(ctx, testAppData, donor)

	assert.Equal(t, expectedError, err)
}

func TestShareCodeSenderSendCertificateProviderPromptWhenEmailErrors(t *testing.T) {
	ctx := context.Background()

	donor := &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
		},
		Donor: donordata.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type: lpadata.LpaTypePropertyAndAffairs,
	}

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("")

	testAppData.Localizer = localizer

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(ctx).
		Return(nil, expectedError)

	sender := &Sender{
		shareCodeStore:           shareCodeStore,
		notifyClient:             notifyClient,
		generate:                 testGenerateFn,
		certificateProviderStore: certificateProviderStore,
	}
	err := sender.SendCertificateProviderPrompt(ctx, testAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderPromptWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := &Sender{
		shareCodeStore: shareCodeStore,
		generate:       testGenerateFn,
	}
	err := sender.SendCertificateProviderPrompt(ctx, testAppData, &donordata.Provided{})

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendOnlineCertificateProviderPrompt(t *testing.T) {
	lpaKey := dynamo.LpaKey("lpa")
	lpaOwnerKey := dynamo.LpaOwnerKey(dynamo.DonorKey("donor"))

	lpa := &lpadata.Lpa{
		CertificateProvider: lpadata.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
			Channel:    lpadata.ChannelOnline,
		},
		Donor: lpadata.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:   lpadata.LpaTypePropertyAndAffairs,
		LpaUID: "lpa-uid",
	}

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(lpa.Type.String()).
		Return("Property and affairs").
		Once()
	testAppData.Localizer = localizer

	ctx := context.Background()

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaCertificateProvider(&certificateproviderdata.Provided{ContactLanguagePreference: localize.En}, lpa), "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
			ShareCode:                   testStringCode,
			CertificateProviderFullName: "Joanna Jones",
			DonorFullName:               "Jan Smith",
			LpaType:                     "property and affairs",
			CertificateProviderStartURL: "http://example.com/certificate-provider",
		}).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeCertificateProvider, testHashedCode, sharecodedata.Link{
			LpaKey:      dynamo.LpaKey("lpa"),
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			LpaUID:      "lpa-uid",
		}).
		Return(nil)

	sender := &Sender{
		shareCodeStore:              shareCodeStore,
		notifyClient:                notifyClient,
		certificateProviderStartURL: "http://example.com/certificate-provider",
		generate:                    testGenerateFn,
	}
	err := sender.SendLpaCertificateProviderPrompt(ctx, testAppData, lpaKey, lpaOwnerKey, lpa)

	assert.Nil(t, err)
}

func TestShareCodeSenderSendOnlineCertificateProviderPromptWhenShareCodeErrors(t *testing.T) {
	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := &Sender{
		shareCodeStore: shareCodeStore,
		generate:       testGenerateFn,
	}
	err := sender.SendLpaCertificateProviderPrompt(context.Background(), testAppData, dynamo.LpaKey(""), dynamo.LpaOwnerKey(dynamo.DonorKey("")), &lpadata.Lpa{})

	assert.ErrorIs(t, err, expectedError)
}

func TestShareCodeSenderSendOnlineCertificateProviderPromptWhenEmailErrors(t *testing.T) {
	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("Property and affairs").
		Once()
	testAppData.Localizer = localizer

	ctx := context.Background()

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	sender := &Sender{
		shareCodeStore:              shareCodeStore,
		notifyClient:                notifyClient,
		certificateProviderStartURL: "http://example.com/certificate-provider",
		generate:                    testGenerateFn,
	}
	err := sender.SendLpaCertificateProviderPrompt(ctx, testAppData, dynamo.LpaKey(""), dynamo.LpaOwnerKey(dynamo.DonorKey("")), &lpadata.Lpa{})

	assert.ErrorIs(t, err, expectedError)
}

func TestShareCodeSenderSendAttorneys(t *testing.T) {
	trustCorporationUID := actoruid.New()
	replacementTrustCorporationUID := actoruid.New()
	attorney1UID := actoruid.New()
	attorney2UID := actoruid.New()
	attorney3UID := actoruid.New()
	replacement1UID := actoruid.New()
	replacement2UID := actoruid.New()

	lpa := &lpadata.Lpa{
		Attorneys: lpadata.Attorneys{
			TrustCorporation: lpadata.TrustCorporation{
				UID:   trustCorporationUID,
				Name:  "Trusty",
				Email: "trusted@example.com",
			},
			Attorneys: []lpadata.Attorney{
				{
					UID:        attorney1UID,
					FirstNames: "Joanna",
					LastName:   "Jones",
					Email:      "name@example.org",
				},
				{
					UID:        attorney2UID,
					FirstNames: "John",
					LastName:   "Jones",
					Email:      "name2@example.org",
				},
				{
					UID:        attorney3UID,
					FirstNames: "Nope",
					LastName:   "Jones",
				},
			},
		},
		ReplacementAttorneys: lpadata.Attorneys{
			TrustCorporation: lpadata.TrustCorporation{
				UID:   replacementTrustCorporationUID,
				Name:  "Untrusty",
				Email: "untrusted@example.com",
			},
			Attorneys: []lpadata.Attorney{
				{
					UID:        replacement1UID,
					FirstNames: "Dave",
					LastName:   "Davis",
					Email:      "dave@example.com",
				},
				{
					UID:        replacement2UID,
					FirstNames: "Donny",
					LastName:   "Davis",
				},
			},
		},
		Donor: lpadata.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:        lpadata.LpaTypePropertyAndAffairs,
		LpaUID:      "lpa-uid",
		LpaKey:      dynamo.LpaKey("lpa"),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(lpa.Type.String()).
		Return("property and affairs")
	localizer.EXPECT().
		Possessive("Jan").
		Return("Jan's")

	testAppData.Localizer = localizer

	ctx := context.Background()

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(ctx, scheduled.Event{
			At:                testNow.AddDate(0, 3, 1),
			Action:            scheduled.ActionRemindAttorneyToComplete,
			TargetLpaKey:      dynamo.LpaKey("lpa"),
			TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			LpaUID:            "lpa-uid",
		}, scheduled.Event{
			At:                lpa.ExpiresAt().AddDate(0, -3, 1),
			Action:            scheduled.ActionRemindAttorneyToComplete,
			TargetLpaKey:      dynamo.LpaKey("lpa"),
			TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			LpaUID:            "lpa-uid",
		}).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeTrustCorporation, testHashedCode, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), LpaUID: "lpa-uid", ActorUID: trustCorporationUID, IsTrustCorporation: true}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementTrustCorporation, testHashedCode, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), LpaUID: "lpa-uid", ActorUID: replacementTrustCorporationUID, IsTrustCorporation: true, IsReplacementAttorney: true}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, testHashedCode, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), LpaUID: "lpa-uid", ActorUID: attorney1UID}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, testHashedCode, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), LpaUID: "lpa-uid", ActorUID: attorney2UID}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, testHashedCode, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), LpaUID: "lpa-uid", ActorUID: attorney3UID}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementAttorney, testHashedCode, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), LpaUID: "lpa-uid", ActorUID: replacement1UID, IsReplacementAttorney: true}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementAttorney, testHashedCode, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), LpaUID: "lpa-uid", ActorUID: replacement2UID, IsReplacementAttorney: true}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaTrustCorporation(lpa.Attorneys.TrustCorporation), "lpa-uid", notify.InitialOriginalAttorneyEmail{
			ShareCode:                 testStringCode,
			AttorneyFullName:          "Trusty",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      "http://example.com/attorney",
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", page.PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaTrustCorporation(lpa.ReplacementAttorneys.TrustCorporation), "lpa-uid", notify.InitialReplacementAttorneyEmail{
			ShareCode:                 testStringCode,
			AttorneyFullName:          "Untrusty",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      "http://example.com/attorney",
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", page.PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaAttorney(lpa.Attorneys.Attorneys[0]), "lpa-uid", notify.InitialOriginalAttorneyEmail{
			ShareCode:                 testStringCode,
			AttorneyFullName:          "Joanna Jones",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      "http://example.com/attorney",
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", page.PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaAttorney(lpa.Attorneys.Attorneys[1]), "lpa-uid", notify.InitialOriginalAttorneyEmail{
			ShareCode:                 testStringCode,
			AttorneyFullName:          "John Jones",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      "http://example.com/attorney",
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", page.PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaAttorney(lpa.ReplacementAttorneys.Attorneys[0]), "lpa-uid", notify.InitialReplacementAttorneyEmail{
			ShareCode:                 testStringCode,
			AttorneyFullName:          "Dave Davis",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      "http://example.com/attorney",
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", page.PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaperFormRequested(ctx, event.PaperFormRequested{
			UID:        "lpa-uid",
			ActorType:  "attorney",
			ActorUID:   attorney3UID,
			AccessCode: testStringCode,
		}).
		Return(nil)
	eventClient.EXPECT().
		SendPaperFormRequested(ctx, event.PaperFormRequested{
			UID:        "lpa-uid",
			ActorType:  "replacementAttorney",
			ActorUID:   replacement2UID,
			AccessCode: testStringCode,
		}).
		Return(nil)
	eventClient.EXPECT().
		SendAttorneyStarted(ctx, event.AttorneyStarted{
			LpaUID:   lpa.LpaUID,
			ActorUID: attorney1UID,
		}).
		Return(nil)
	eventClient.EXPECT().
		SendAttorneyStarted(ctx, event.AttorneyStarted{
			LpaUID:   lpa.LpaUID,
			ActorUID: attorney2UID,
		}).
		Return(nil)
	eventClient.EXPECT().
		SendAttorneyStarted(ctx, event.AttorneyStarted{
			LpaUID:   lpa.LpaUID,
			ActorUID: attorney3UID,
		}).
		Return(nil)
	eventClient.EXPECT().
		SendAttorneyStarted(ctx, event.AttorneyStarted{
			LpaUID:   lpa.LpaUID,
			ActorUID: trustCorporationUID,
		}).
		Return(nil)
	eventClient.EXPECT().
		SendAttorneyStarted(ctx, event.AttorneyStarted{
			LpaUID:   lpa.LpaUID,
			ActorUID: replacementTrustCorporationUID,
		}).
		Return(nil)
	eventClient.EXPECT().
		SendAttorneyStarted(ctx, event.AttorneyStarted{
			LpaUID:   lpa.LpaUID,
			ActorUID: replacement1UID,
		}).
		Return(nil)
	eventClient.EXPECT().
		SendAttorneyStarted(ctx, event.AttorneyStarted{
			LpaUID:   lpa.LpaUID,
			ActorUID: replacement2UID,
		}).
		Return(nil)

	sender := &Sender{
		shareCodeStore:   shareCodeStore,
		notifyClient:     notifyClient,
		appPublicURL:     "http://app",
		attorneyStartURL: "http://example.com/attorney",
		generate:         testGenerateFn,
		eventClient:      eventClient,
		scheduledStore:   scheduledStore,
		now:              testNowFn,
	}
	err := sender.SendAttorneys(ctx, testAppData, lpa)

	assert.Nil(t, err)
}

func TestShareCodeSenderSendAttorneysTrustCorporationsNoEmail(t *testing.T) {
	uid1 := actoruid.New()
	uid2 := actoruid.New()

	donor := &lpadata.Lpa{
		Attorneys: lpadata.Attorneys{
			TrustCorporation: lpadata.TrustCorporation{
				UID:  uid1,
				Name: "Trusty",
			},
		},
		ReplacementAttorneys: lpadata.Attorneys{
			TrustCorporation: lpadata.TrustCorporation{
				UID:  uid2,
				Name: "Untrusty",
			},
		},
		Donor: lpadata.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:        lpadata.LpaTypePropertyAndAffairs,
		LpaUID:      "lpa-uid",
		LpaKey:      dynamo.LpaKey("lpa"),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	ctx := context.Background()

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeTrustCorporation, testHashedCode, sharecodedata.Link{
			LpaOwnerKey:        dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			LpaKey:             dynamo.LpaKey("lpa"),
			LpaUID:             "lpa-uid",
			ActorUID:           uid1,
			IsTrustCorporation: true,
		}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementTrustCorporation, testHashedCode, sharecodedata.Link{
			LpaOwnerKey:           dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			LpaKey:                dynamo.LpaKey("lpa"),
			LpaUID:                "lpa-uid",
			ActorUID:              uid2,
			IsTrustCorporation:    true,
			IsReplacementAttorney: true,
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaperFormRequested(ctx, event.PaperFormRequested{
			UID:        "lpa-uid",
			ActorType:  "trustCorporation",
			ActorUID:   uid1,
			AccessCode: testStringCode,
		}).
		Return(nil)
	eventClient.EXPECT().
		SendPaperFormRequested(ctx, event.PaperFormRequested{
			UID:        "lpa-uid",
			ActorType:  "replacementTrustCorporation",
			ActorUID:   uid2,
			AccessCode: testStringCode,
		}).
		Return(nil)
	eventClient.EXPECT().
		SendAttorneyStarted(ctx, event.AttorneyStarted{
			LpaUID:   donor.LpaUID,
			ActorUID: uid1,
		}).
		Return(nil)
	eventClient.EXPECT().
		SendAttorneyStarted(ctx, event.AttorneyStarted{
			LpaUID:   donor.LpaUID,
			ActorUID: uid2,
		}).
		Return(nil)

	sender := &Sender{
		shareCodeStore: shareCodeStore,
		scheduledStore: scheduledStore,
		appPublicURL:   "http://app",
		generate:       testGenerateFn,
		eventClient:    eventClient,
		now:            testNowFn,
	}
	err := sender.SendAttorneys(ctx, testAppData, donor)

	assert.Nil(t, err)
}

func TestShareCodeSenderSendAttorneysWithTestCode(t *testing.T) {
	uid := actoruid.New()

	testcases := map[string]struct {
		useTestCode      bool
		expectedTestCode string
	}{
		"with test code": {
			useTestCode:      true,
			expectedTestCode: "abcdef123456",
		},
		"without test code": {
			useTestCode:      false,
			expectedTestCode: testStringCode,
		},
	}

	lpa := &lpadata.Lpa{
		Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{
			{
				FirstNames: "Joanna",
				LastName:   "Jones",
				Email:      "name@example.org",
				UID:        uid,
			},
		}},
		Donor: lpadata.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:        lpadata.LpaTypePropertyAndAffairs,
		LpaUID:      "lpa-uid",
		LpaKey:      dynamo.LpaKey("lpa"),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(lpa.Type.String()).
		Return("property and affairs")
	localizer.EXPECT().
		Possessive("Jan").
		Return("Jan's")

	testAppData.Localizer = localizer

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			scheduledStore := newMockScheduledStore(t)
			scheduledStore.EXPECT().
				Create(mock.Anything, mock.Anything, mock.Anything).
				Return(nil)

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeAttorney, sharecodedata.HashedFromString(tc.expectedTestCode), sharecodedata.Link{
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"),
					LpaUID:   "lpa-uid",
					ActorUID: uid,
				}).
				Return(nil)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeAttorney, testHashedCode, sharecodedata.Link{
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"),
					LpaUID:   "lpa-uid",
					ActorUID: uid,
				}).
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaAttorney(lpa.Attorneys.Attorneys[0]), "lpa-uid", notify.InitialOriginalAttorneyEmail{
					ShareCode:                 tc.expectedTestCode,
					AttorneyFullName:          "Joanna Jones",
					DonorFirstNames:           "Jan",
					DonorFullName:             "Jan Smith",
					DonorFirstNamesPossessive: "Jan's",
					LpaType:                   "property and affairs",
					AttorneyStartPageURL:      "http://example.com/attorney",
					AttorneyOptOutURL:         fmt.Sprintf("http://app%s", page.PathAttorneyEnterReferenceNumberOptOut),
				}).
				Return(nil)
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaAttorney(lpa.Attorneys.Attorneys[0]), "lpa-uid", notify.InitialOriginalAttorneyEmail{
					ShareCode:                 testStringCode,
					AttorneyFullName:          "Joanna Jones",
					DonorFirstNames:           "Jan",
					DonorFullName:             "Jan Smith",
					DonorFirstNamesPossessive: "Jan's",
					LpaType:                   "property and affairs",
					AttorneyStartPageURL:      "http://example.com/attorney",
					AttorneyOptOutURL:         fmt.Sprintf("http://app%s", page.PathAttorneyEnterReferenceNumberOptOut),
				}).
				Return(nil)

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendAttorneyStarted(ctx, event.AttorneyStarted{
					LpaUID:   lpa.LpaUID,
					ActorUID: uid,
				}).
				Return(nil)
			eventClient.EXPECT().
				SendAttorneyStarted(ctx, event.AttorneyStarted{
					LpaUID:   lpa.LpaUID,
					ActorUID: uid,
				}).
				Return(nil)

			sender := &Sender{
				shareCodeStore:   shareCodeStore,
				scheduledStore:   scheduledStore,
				notifyClient:     notifyClient,
				appPublicURL:     "http://app",
				attorneyStartURL: "http://example.com/attorney",
				generate:         testGenerateFn,
				eventClient:      eventClient,
				now:              testNowFn,
			}

			if tc.useTestCode {
				sender.UseTestCode("abcdef123456")
			}

			err := sender.SendAttorneys(ctx, testAppData, lpa)
			assert.Nil(t, err)

			err = sender.SendAttorneys(ctx, testAppData, lpa)
			assert.Nil(t, err)
		})
	}
}

func TestShareCodeSenderSendAttorneysWhenEmailErrors(t *testing.T) {
	ctx := context.Background()
	uid := actoruid.New()

	donor := &lpadata.Lpa{
		Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{
			{
				FirstNames: "Joanna",
				LastName:   "Jones",
				Email:      "name@example.org",
				UID:        uid,
			},
		}},
		Donor: lpadata.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type: lpadata.LpaTypePropertyAndAffairs,
	}

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(donor.Type.String()).
		Return("property and affairs")
	localizer.EXPECT().
		Possessive("Jan").
		Return("Jan's")
	testAppData.Localizer = localizer

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendAttorneyStarted(mock.Anything, mock.Anything).
		Return(nil)

	sender := &Sender{
		shareCodeStore: shareCodeStore,
		notifyClient:   notifyClient,
		appPublicURL:   "http://app",
		generate:       testGenerateFn,
		eventClient:    eventClient,
		scheduledStore: scheduledStore,
		now:            testNowFn,
	}
	err := sender.SendAttorneys(ctx, testAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendAttorneysWhenScheduledStoreErrors(t *testing.T) {
	ctx := context.Background()

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := &Sender{
		generate:       testGenerateFn,
		scheduledStore: scheduledStore,
		now:            testNowFn,
	}
	err := sender.SendAttorneys(ctx, testAppData, &lpadata.Lpa{
		Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{Email: "hey@example.com"}}},
	})

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendAttorneysWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := &Sender{
		shareCodeStore: shareCodeStore,
		appPublicURL:   "http://app",
		generate:       testGenerateFn,
		scheduledStore: scheduledStore,
		now:            testNowFn,
	}
	err := sender.SendAttorneys(ctx, testAppData, &lpadata.Lpa{
		Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{Email: "hey@example.com"}}},
	})

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendAttorneysWhenEventClientErrors(t *testing.T) {
	uid := actoruid.New()

	testcases := map[string]*lpadata.Lpa{
		"original attorneys": {
			Attorneys: lpadata.Attorneys{
				Attorneys: []lpadata.Attorney{{UID: uid}},
			},
		},
		"replacement attorneys": {
			ReplacementAttorneys: lpadata.Attorneys{
				Attorneys: []lpadata.Attorney{{UID: uid}},
			},
		},
		"original trust corporation": {
			Attorneys: lpadata.Attorneys{
				TrustCorporation: lpadata.TrustCorporation{UID: uid, Name: "a"},
			},
		},
		"replacement trust corporation": {
			ReplacementAttorneys: lpadata.Attorneys{
				TrustCorporation: lpadata.TrustCorporation{UID: uid, Name: "a"},
			},
		},
	}

	ctx := context.Background()

	for name, lpa := range testcases {
		t.Run(name, func(t *testing.T) {
			scheduledStore := newMockScheduledStore(t)
			scheduledStore.EXPECT().
				Create(mock.Anything, mock.Anything, mock.Anything).
				Return(nil)

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil)

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendAttorneyStarted(mock.Anything, mock.Anything).
				Return(expectedError)

			sender := &Sender{
				shareCodeStore: shareCodeStore,
				appPublicURL:   "http://app",
				generate:       testGenerateFn,
				eventClient:    eventClient,
				scheduledStore: scheduledStore,
				now:            testNowFn,
			}
			err := sender.SendAttorneys(ctx, testAppData, lpa)

			assert.Equal(t, expectedError, err)
		})
	}
}

func TestSendVoucherInvite(t *testing.T) {
	ctx := context.Background()
	uid := actoruid.New()

	testcases := map[string]struct {
		setupNotifyClient    func(*mockNotifyClient, *donordata.Provided)
		setupLocalizer       func(*mockLocalizer)
		donor                donordata.Donor
		correspondent        donordata.Correspondent
		voucherCodeSentBySMS bool
		voucherCodeSentTo    string
	}{
		"sms": {
			setupNotifyClient: func(nc *mockNotifyClient, provided *donordata.Provided) {
				nc.EXPECT().
					SendActorSMS(ctx, notify.ToDonorOnly(provided), "lpa-uid", notify.VouchingShareCodeSMS{
						ShareCode:                 testStringCode,
						DonorFullNamePossessive:   "Possessive full name",
						LpaType:                   "translated type",
						VoucherFullName:           "c d",
						DonorFirstNamesPossessive: "Possessive first names",
					}).
					Return(nil)
				nc.EXPECT().
					SendActorEmail(ctx, notify.ToVoucher(provided.Voucher), "lpa-uid",
						notify.VoucherInviteEmail{
							VoucherFullName:           "c d",
							DonorFullName:             "a b",
							DonorFirstNamesPossessive: "Possessive first names",
							DonorFirstNames:           "a",
							LpaType:                   "translated type",
							VoucherStartPageURL:       "http://app" + page.PathVoucherStart.Format(),
						}).
					Return(nil)
			},
			setupLocalizer: func(l *mockLocalizer) {
				l.EXPECT().
					T(lpadata.LpaTypePersonalWelfare.String()).
					Return("translated type").
					Twice()
				l.EXPECT().
					Possessive("a").
					Return("Possessive first names").
					Twice()
				l.EXPECT().
					Possessive("a b").
					Return("Possessive full name")
			},
			donor: donordata.Donor{
				FirstNames: "a",
				LastName:   "b",
				Mobile:     "123",
				Email:      "donor@example.com",
			},
			voucherCodeSentBySMS: true,
			voucherCodeSentTo:    "123",
		},
		"email": {
			setupNotifyClient: func(nc *mockNotifyClient, provided *donordata.Provided) {
				nc.EXPECT().
					SendActorEmail(ctx, notify.ToDonorOnly(provided), "lpa-uid",
						notify.VouchingShareCodeEmail{
							ShareCode:       testStringCode,
							VoucherFullName: "c d",
							DonorFullName:   "a b",
							LpaType:         "translated type",
						}).
					Return(nil)
				nc.EXPECT().
					SendActorEmail(ctx, notify.ToVoucher(provided.Voucher), "lpa-uid",
						notify.VoucherInviteEmail{
							VoucherFullName:           "c d",
							DonorFullName:             "a b",
							DonorFirstNamesPossessive: "Possessive first names",
							DonorFirstNames:           "a",
							LpaType:                   "translated type",
							VoucherStartPageURL:       "http://app" + page.PathVoucherStart.Format(),
						}).
					Return(nil)
			},
			setupLocalizer: func(l *mockLocalizer) {
				l.EXPECT().
					T(lpadata.LpaTypePersonalWelfare.String()).
					Return("translated type")
				l.EXPECT().
					Possessive("a").
					Return("Possessive first names")
				l.EXPECT().
					T(lpadata.LpaTypePersonalWelfare.String()).
					Return("translated type")
			},
			donor: donordata.Donor{
				FirstNames: "a",
				LastName:   "b",
				Email:      "donor@example.com",
			},
			voucherCodeSentBySMS: false,
			voucherCodeSentTo:    "donor@example.com",
		},
		"email has correspondent": {
			setupNotifyClient: func(nc *mockNotifyClient, provided *donordata.Provided) {
				nc.EXPECT().
					SendActorEmail(ctx, notify.ToDonorOnly(provided), "lpa-uid",
						notify.VouchingShareCodeEmail{
							ShareCode:       testStringCode,
							VoucherFullName: "c d",
							DonorFullName:   "a b",
							LpaType:         "translated type",
						}).
					Return(nil)
				nc.EXPECT().
					SendActorEmail(ctx, notify.ToCorrespondent(provided), "lpa-uid",
						notify.CorrespondentInformedVouchingInProgress{
							CorrespondentFullName:   "corr espond",
							DonorFullName:           "a b",
							DonorFullNamePossessive: "Possessive full name",
							LpaType:                 "translated type",
						}).
					Return(nil)
				nc.EXPECT().
					SendActorEmail(ctx, notify.ToVoucher(provided.Voucher), "lpa-uid",
						notify.VoucherInviteEmail{
							VoucherFullName:           "c d",
							DonorFullName:             "a b",
							DonorFirstNamesPossessive: "Possessive first names",
							DonorFirstNames:           "a",
							LpaType:                   "translated type",
							VoucherStartPageURL:       "http://app" + page.PathVoucherStart.Format(),
						}).
					Return(nil)
			},
			setupLocalizer: func(l *mockLocalizer) {
				l.EXPECT().
					T(lpadata.LpaTypePersonalWelfare.String()).
					Return("translated type")
				l.EXPECT().
					Possessive("a").
					Return("Possessive first names")
				l.EXPECT().
					Possessive("a b").
					Return("Possessive full name")
				l.EXPECT().
					T(lpadata.LpaTypePersonalWelfare.String()).
					Return("translated type")
			},
			donor: donordata.Donor{
				FirstNames: "a",
				LastName:   "b",
				Email:      "donor@example.com",
			},
			correspondent: donordata.Correspondent{
				FirstNames: "corr",
				LastName:   "espond",
				Email:      "correspondent@example.com",
			},
			voucherCodeSentBySMS: false,
			voucherCodeSentTo:    "donor@example.com",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			provided := &donordata.Provided{
				PK:            dynamo.LpaKey("lpa"),
				SK:            dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				LpaUID:        "lpa-uid",
				Type:          lpadata.LpaTypePersonalWelfare,
				Donor:         tc.donor,
				Correspondent: tc.correspondent,
				Voucher: donordata.Voucher{
					UID:        uid,
					FirstNames: "c",
					LastName:   "d",
					Email:      "voucher@example.com",
				},
			}

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeVoucher, testHashedCode, sharecodedata.Link{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
					LpaUID:      "lpa-uid",
					ActorUID:    uid,
				}).
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			tc.setupNotifyClient(notifyClient, provided)

			sender := &Sender{
				shareCodeStore: shareCodeStore,
				notifyClient:   notifyClient,
				appPublicURL:   "http://app",
				generate:       testGenerateFn,
				now:            testNowFn,
			}

			localizer := newMockLocalizer(t)
			tc.setupLocalizer(localizer)

			appData := testAppData
			appData.Localizer = localizer

			err := sender.SendVoucherInvite(ctx, provided, appData)
			assert.Nil(t, err)

			assert.Equal(t, testNow, provided.VoucherInvitedAt)
			assert.Equal(t, tc.voucherCodeSentBySMS, provided.VoucherCodeSentBySMS)
			assert.Equal(t, tc.voucherCodeSentTo, provided.VoucherCodeSentTo)
		})
	}
}

func TestSendVoucherAccessCodeWhenShareCodeStoreError(t *testing.T) {
	ctx := context.Background()
	uid := actoruid.New()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := &Sender{
		shareCodeStore: shareCodeStore,
		appPublicURL:   "http://app",
		generate:       testGenerateFn,
		now:            testNowFn,
	}

	err := sender.SendVoucherAccessCode(ctx, &donordata.Provided{
		PK:     dynamo.LpaKey("lpa"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
		LpaUID: "lpa-uid",
		Type:   lpadata.LpaTypePersonalWelfare,
		Donor: donordata.Donor{
			FirstNames: "a",
			LastName:   "b",
			Mobile:     "123",
			Email:      "a@example.com",
		},
		Voucher: donordata.Voucher{
			UID:        uid,
			FirstNames: "c",
			LastName:   "d",
		},
	}, testAppData)

	assert.Equal(t, fmt.Errorf("creating share failed: %w", expectedError), err)
}

func TestSendVoucherInviteWhenNotifyClientError(t *testing.T) {
	testcases := map[string]struct {
		email              string
		correspondentEmail string
		mobile             string
		notifyClient       func() *mockNotifyClient
		localizer          func() *mockLocalizer
		error              error
	}{
		"sms": {
			mobile: "123",
			notifyClient: func() *mockNotifyClient {
				nc := newMockNotifyClient(t)
				nc.EXPECT().
					SendActorSMS(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError).
					Once()
				return nc
			},
			localizer: func() *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					T(mock.Anything).
					Return("translated type")
				l.EXPECT().
					Possessive(mock.Anything).
					Return("Possessive first names")
				l.EXPECT().
					Possessive(mock.Anything).
					Return("Possessive full name")
				return l
			},
			error: fmt.Errorf("sms failed: %w", expectedError),
		},
		"email": {
			email: "a@example.com",
			notifyClient: func() *mockNotifyClient {
				nc := newMockNotifyClient(t)
				nc.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError).
					Once()
				return nc
			},
			localizer: func() *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					T(mock.Anything).
					Return("translated type")
				return l
			},
			error: fmt.Errorf("email failed: %w", expectedError),
		},
		"voucher email": {
			mobile: "123",
			notifyClient: func() *mockNotifyClient {
				nc := newMockNotifyClient(t)
				nc.EXPECT().
					SendActorSMS(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).
					Once()
				nc.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError).
					Once()
				return nc
			},
			localizer: func() *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					T(mock.Anything).
					Return("translated type").
					Times(2)
				l.EXPECT().
					Possessive(mock.Anything).
					Return("Possessive first names").
					Times(3)
				return l
			},
			error: fmt.Errorf("email failed: %w", expectedError),
		},
		"correspondent email": {
			mobile:             "123",
			correspondentEmail: "exampel@example.com",
			notifyClient: func() *mockNotifyClient {
				nc := newMockNotifyClient(t)
				nc.EXPECT().
					SendActorSMS(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).
					Once()
				nc.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError).
					Once()
				return nc
			},
			localizer: func() *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					T(mock.Anything).
					Return("translated type").
					Times(2)
				l.EXPECT().
					Possessive(mock.Anything).
					Return("Possessive first names").
					Times(3)
				return l
			},
			error: fmt.Errorf("email failed: %w", expectedError),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			uid := actoruid.New()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil)

			testAppData.Localizer = tc.localizer()

			sender := &Sender{
				shareCodeStore: shareCodeStore,
				notifyClient:   tc.notifyClient(),
				appPublicURL:   "http://app",
				generate:       testGenerateFn,
				now:            testNowFn,
			}

			err := sender.SendVoucherInvite(ctx, &donordata.Provided{
				PK:     dynamo.LpaKey("lpa"),
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePersonalWelfare,
				Donor: donordata.Donor{
					FirstNames: "a",
					LastName:   "b",
					Mobile:     tc.mobile,
					Email:      tc.email,
				},
				Correspondent: donordata.Correspondent{
					Email: tc.correspondentEmail,
				},
				Voucher: donordata.Voucher{
					UID:        uid,
					FirstNames: "c",
					LastName:   "d",
				},
			}, testAppData)

			assert.Equal(t, tc.error, err)
		})
	}
}
