package sharecode

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testRandomString = "123"

var (
	TestAppData = appcontext.Data{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
	}
	testRandomStringFn = func(int) string { return testRandomString }
)

func TestShareCodeSenderSendCertificateProviderInvite(t *testing.T) {
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
	TestAppData.Localizer = localizer

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeCertificateProvider, testRandomString, sharecodedata.Link{
			LpaKey:      dynamo.LpaKey("lpa"),
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, to, "lpa-uid", notify.CertificateProviderInviteEmail{
			ShareCode:                    testRandomString,
			CertificateProviderFullName:  "Joanna Jones",
			DonorFirstNames:              "Jan",
			DonorFullName:                "Jan Smith",
			LpaType:                      "property and affairs",
			CertificateProviderStartURL:  fmt.Sprintf("http://app%s", page.PathCertificateProviderStart),
			DonorFirstNamesPossessive:    "Jan’s",
			WhatLpaCovers:                "houses and stuff",
			CertificateProviderOptOutURL: fmt.Sprintf("http://app%s", page.PathCertificateProviderEnterReferenceNumberOptOut),
		}).
		Return(nil)

	sender := NewSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil, nil)
	err := sender.SendCertificateProviderInvite(ctx, TestAppData, CertificateProviderInvite{
		LpaKey:                      dynamo.LpaKey("lpa"),
		LpaOwnerKey:                 dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
		LpaUID:                      donor.LpaUID,
		Type:                        donor.Type,
		DonorFirstNames:             donor.Donor.FirstNames,
		DonorFullName:               donor.Donor.FullName(),
		CertificateProviderUID:      donor.CertificateProvider.UID,
		CertificateProviderFullName: donor.CertificateProvider.FullName(),
	}, to)

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
			expectedTestCode: testRandomString,
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
			TestAppData.Localizer = localizer

			to := notify.ToCertificateProvider(donor.CertificateProvider)
			ctx := context.Background()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, tc.expectedTestCode, sharecodedata.Link{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				}).
				Once().
				Return(nil)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, testRandomString, sharecodedata.Link{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
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
					CertificateProviderStartURL:  fmt.Sprintf("http://app%s", page.PathCertificateProviderStart),
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
					CertificateProviderStartURL:  fmt.Sprintf("http://app%s", page.PathCertificateProviderStart),
					ShareCode:                    testRandomString,
					DonorFirstNamesPossessive:    "Jan’s",
					WhatLpaCovers:                "health and stuff",
					CertificateProviderOptOutURL: fmt.Sprintf("http://app%s", page.PathCertificateProviderEnterReferenceNumberOptOut),
				}).
				Once().
				Return(nil)

			sender := NewSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil, nil)

			if tc.useTestCode {
				sender.UseTestCode("abcdef123456")
			}

			err := sender.SendCertificateProviderInvite(ctx, TestAppData, CertificateProviderInvite{
				LpaKey:                      dynamo.LpaKey("lpa"),
				LpaOwnerKey:                 dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				LpaUID:                      donor.LpaUID,
				Type:                        donor.Type,
				DonorFirstNames:             donor.Donor.FirstNames,
				DonorFullName:               donor.Donor.FullName(),
				CertificateProviderUID:      donor.CertificateProvider.UID,
				CertificateProviderFullName: donor.CertificateProvider.FullName(),
			}, to)
			assert.Nil(t, err)

			err = sender.SendCertificateProviderInvite(ctx, TestAppData, CertificateProviderInvite{
				LpaKey:                      dynamo.LpaKey("lpa"),
				LpaOwnerKey:                 dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				LpaUID:                      donor.LpaUID,
				Type:                        donor.Type,
				DonorFirstNames:             donor.Donor.FirstNames,
				DonorFullName:               donor.Donor.FullName(),
				CertificateProviderUID:      donor.CertificateProvider.UID,
				CertificateProviderFullName: donor.CertificateProvider.FullName(),
			}, to)
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

	to := notify.ToCertificateProvider(donor.CertificateProvider)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("")
	localizer.EXPECT().
		Possessive("Jan").
		Return("Jan’s")
	TestAppData.Localizer = localizer

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil, nil)
	err := sender.SendCertificateProviderInvite(ctx, TestAppData, CertificateProviderInvite{
		LpaUID:                      donor.LpaUID,
		Type:                        donor.Type,
		DonorFirstNames:             donor.Donor.FirstNames,
		DonorFullName:               donor.Donor.FullName(),
		CertificateProviderUID:      donor.CertificateProvider.UID,
		CertificateProviderFullName: donor.CertificateProvider.FullName(),
	}, to)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderInviteWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewSender(shareCodeStore, nil, "http://app", testRandomStringFn, nil, nil)
	err := sender.SendCertificateProviderInvite(ctx, TestAppData, CertificateProviderInvite{}, notify.ToCustomEmail(localize.En, ""))

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
	TestAppData.Localizer = localizer

	ctx := context.Background()

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToCertificateProvider(donor.CertificateProvider), "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
			ShareCode:                   testRandomString,
			CertificateProviderFullName: "Joanna Jones",
			DonorFullName:               "Jan Smith",
			LpaType:                     "property and affairs",
			CertificateProviderStartURL: fmt.Sprintf("http://app%s", page.PathCertificateProviderStart),
		}).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeCertificateProvider, testRandomString, sharecodedata.Link{
			LpaKey:      dynamo.LpaKey("lpa"),
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
		}).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(ctx).
		Return(nil, expectedError)

	sender := NewSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil, certificateProviderStore)
	err := sender.SendCertificateProviderPrompt(ctx, TestAppData, donor)

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
		Put(ctx, actor.TypeCertificateProvider, testRandomString, sharecodedata.Link{
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			LpaKey:      dynamo.LpaKey("lpa"),
			ActorUID:    actorUID,
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaperFormRequested(ctx, event.PaperFormRequested{
			UID:        "lpa-uid",
			ActorType:  actor.TypeCertificateProvider.String(),
			ActorUID:   actorUID,
			AccessCode: testRandomString,
		}).
		Return(nil)

	sender := NewSender(shareCodeStore, nil, "http://app", testRandomStringFn, eventClient, nil)
	err := sender.SendCertificateProviderPrompt(ctx, TestAppData, donor)

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
			expectedTestCode: testRandomString,
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

			TestAppData.Localizer = localizer

			ctx := context.Background()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, tc.expectedTestCode, sharecodedata.Link{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				}).
				Once().
				Return(nil)

			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, testRandomString, sharecodedata.Link{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				}).
				Once().
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToCertificateProvider(donor.CertificateProvider), "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
					CertificateProviderFullName: "Joanna Jones",
					DonorFullName:               "Jan Smith",
					LpaType:                     "property and affairs",
					CertificateProviderStartURL: fmt.Sprintf("http://app%s", page.PathCertificateProviderStart),
					ShareCode:                   tc.expectedTestCode,
				}).
				Once().
				Return(nil)
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToCertificateProvider(donor.CertificateProvider), "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
					CertificateProviderFullName: "Joanna Jones",
					DonorFullName:               "Jan Smith",
					LpaType:                     "property and affairs",
					CertificateProviderStartURL: fmt.Sprintf("http://app%s", page.PathCertificateProviderStart),
					ShareCode:                   testRandomString,
				}).
				Once().
				Return(nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				GetAny(ctx).
				Return(nil, expectedError)

			sender := NewSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil, certificateProviderStore)

			if tc.useTestCode {
				sender.UseTestCode("abcdef123456")
			}

			err := sender.SendCertificateProviderPrompt(ctx, TestAppData, donor)
			assert.Nil(t, err)

			err = sender.SendCertificateProviderPrompt(ctx, TestAppData, donor)
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

	sender := NewSender(shareCodeStore, nil, "http://app", testRandomStringFn, nil, nil)
	err := sender.SendCertificateProviderPrompt(ctx, TestAppData, donor)

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

	sender := NewSender(shareCodeStore, nil, "http://app", testRandomStringFn, eventClient, nil)
	err := sender.SendCertificateProviderPrompt(ctx, TestAppData, donor)

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

	TestAppData.Localizer = localizer

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

	sender := NewSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil, certificateProviderStore)
	err := sender.SendCertificateProviderPrompt(ctx, TestAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderPromptWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewSender(shareCodeStore, nil, "http://app", testRandomStringFn, nil, nil)
	err := sender.SendCertificateProviderPrompt(ctx, TestAppData, &donordata.Provided{})

	assert.Equal(t, expectedError, errors.Unwrap(err))
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

	TestAppData.Localizer = localizer

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeTrustCorporation, testRandomString, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: trustCorporationUID, IsTrustCorporation: true}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementTrustCorporation, testRandomString, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: replacementTrustCorporationUID, IsTrustCorporation: true, IsReplacementAttorney: true}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, testRandomString, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: attorney1UID}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, testRandomString, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: attorney2UID}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, testRandomString, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: attorney3UID}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementAttorney, testRandomString, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: replacement1UID, IsReplacementAttorney: true}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementAttorney, testRandomString, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: replacement2UID, IsReplacementAttorney: true}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaTrustCorporation(lpa.Attorneys.TrustCorporation), "lpa-uid", notify.InitialOriginalAttorneyEmail{
			ShareCode:                 testRandomString,
			AttorneyFullName:          "Trusty",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", page.PathAttorneyStart),
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", page.PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaTrustCorporation(lpa.ReplacementAttorneys.TrustCorporation), "lpa-uid", notify.InitialReplacementAttorneyEmail{
			ShareCode:                 testRandomString,
			AttorneyFullName:          "Untrusty",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", page.PathAttorneyStart),
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", page.PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaAttorney(lpa.Attorneys.Attorneys[0]), "lpa-uid", notify.InitialOriginalAttorneyEmail{
			ShareCode:                 testRandomString,
			AttorneyFullName:          "Joanna Jones",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", page.PathAttorneyStart),
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", page.PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaAttorney(lpa.Attorneys.Attorneys[1]), "lpa-uid", notify.InitialOriginalAttorneyEmail{
			ShareCode:                 testRandomString,
			AttorneyFullName:          "John Jones",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", page.PathAttorneyStart),
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", page.PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaAttorney(lpa.ReplacementAttorneys.Attorneys[0]), "lpa-uid", notify.InitialReplacementAttorneyEmail{
			ShareCode:                 testRandomString,
			AttorneyFullName:          "Dave Davis",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", page.PathAttorneyStart),
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", page.PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaperFormRequested(ctx, event.PaperFormRequested{
			UID:        "lpa-uid",
			ActorType:  "attorney",
			ActorUID:   attorney3UID,
			AccessCode: testRandomString,
		}).
		Return(nil)
	eventClient.EXPECT().
		SendPaperFormRequested(ctx, event.PaperFormRequested{
			UID:        "lpa-uid",
			ActorType:  "replacementAttorney",
			ActorUID:   replacement2UID,
			AccessCode: testRandomString,
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

	sender := NewSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, eventClient, nil)
	err := sender.SendAttorneys(ctx, TestAppData, lpa)

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
	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeTrustCorporation, testRandomString, sharecodedata.Link{
			LpaOwnerKey:        dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			LpaKey:             dynamo.LpaKey("lpa"),
			ActorUID:           uid1,
			IsTrustCorporation: true,
		}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementTrustCorporation, testRandomString, sharecodedata.Link{
			LpaOwnerKey:           dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			LpaKey:                dynamo.LpaKey("lpa"),
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
			AccessCode: testRandomString,
		}).
		Return(nil)
	eventClient.EXPECT().
		SendPaperFormRequested(ctx, event.PaperFormRequested{
			UID:        "lpa-uid",
			ActorType:  "replacementTrustCorporation",
			ActorUID:   uid2,
			AccessCode: testRandomString,
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

	sender := NewSender(shareCodeStore, nil, "http://app", testRandomStringFn, eventClient, nil)
	err := sender.SendAttorneys(ctx, TestAppData, donor)

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
			expectedTestCode: testRandomString,
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

	TestAppData.Localizer = localizer

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeAttorney, tc.expectedTestCode, sharecodedata.Link{
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"),
					ActorUID: uid,
				}).
				Return(nil)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeAttorney, testRandomString, sharecodedata.Link{
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"),
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
					AttorneyStartPageURL:      fmt.Sprintf("http://app%s", page.PathAttorneyStart),
					AttorneyOptOutURL:         fmt.Sprintf("http://app%s", page.PathAttorneyEnterReferenceNumberOptOut),
				}).
				Return(nil)
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaAttorney(lpa.Attorneys.Attorneys[0]), "lpa-uid", notify.InitialOriginalAttorneyEmail{
					ShareCode:                 testRandomString,
					AttorneyFullName:          "Joanna Jones",
					DonorFirstNames:           "Jan",
					DonorFullName:             "Jan Smith",
					DonorFirstNamesPossessive: "Jan's",
					LpaType:                   "property and affairs",
					AttorneyStartPageURL:      fmt.Sprintf("http://app%s", page.PathAttorneyStart),
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

			sender := NewSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, eventClient, nil)

			if tc.useTestCode {
				sender.UseTestCode("abcdef123456")
			}

			err := sender.SendAttorneys(ctx, TestAppData, lpa)
			assert.Nil(t, err)

			err = sender.SendAttorneys(ctx, TestAppData, lpa)
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
	TestAppData.Localizer = localizer

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

	sender := NewSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, eventClient, nil)
	err := sender.SendAttorneys(ctx, TestAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendAttorneysWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewSender(shareCodeStore, nil, "http://app", testRandomStringFn, nil, nil)
	err := sender.SendAttorneys(ctx, TestAppData, &lpadata.Lpa{
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
			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil)

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendAttorneyStarted(mock.Anything, mock.Anything).
				Return(expectedError)

			sender := NewSender(shareCodeStore, nil, "http://app", testRandomStringFn, eventClient, nil)
			err := sender.SendAttorneys(ctx, TestAppData, lpa)

			assert.Equal(t, expectedError, err)
		})
	}
}

func TestSendVoucherAccessCode(t *testing.T) {
	ctx := context.Background()
	uid := actoruid.New()

	testcases := map[string]struct {
		notifyClient func(*donordata.Provided) *mockNotifyClient
		localizer    func() *mockLocalizer
		donor        donordata.Donor
	}{
		"sms": {
			notifyClient: func(provided *donordata.Provided) *mockNotifyClient {
				nc := newMockNotifyClient(t)
				nc.EXPECT().
					SendActorSMS(ctx, notify.ToDonor(provided), "lpa-uid", notify.VouchingShareCodeSMS{
						ShareCode:                 testRandomString,
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
				return nc
			},
			localizer: func() *mockLocalizer {
				l := newMockLocalizer(t)
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
				return l
			},
			donor: donordata.Donor{
				FirstNames: "a",
				LastName:   "b",
				Mobile:     "123",
				Email:      "donor@example.com",
			},
		},
		"email": {
			notifyClient: func(provided *donordata.Provided) *mockNotifyClient {
				nc := newMockNotifyClient(t)
				nc.EXPECT().
					SendActorEmail(ctx, notify.ToDonor(provided), "lpa-uid",
						notify.VouchingShareCodeEmail{
							ShareCode:       testRandomString,
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
				return nc
			},
			localizer: func() *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					T(lpadata.LpaTypePersonalWelfare.String()).
					Return("translated type")
				l.EXPECT().
					Possessive("a").
					Return("Possessive first names")
				l.EXPECT().
					T(lpadata.LpaTypePersonalWelfare.String()).
					Return("translated type")
				return l
			},
			donor: donordata.Donor{
				FirstNames: "a",
				LastName:   "b",
				Email:      "donor@example.com",
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			provided := &donordata.Provided{
				PK:     dynamo.LpaKey("lpa"),
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePersonalWelfare,
				Donor:  tc.donor,
				Voucher: donordata.Voucher{
					UID:        uid,
					FirstNames: "c",
					LastName:   "d",
					Email:      "voucher@example.com",
				},
			}

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeVoucher, testRandomString, sharecodedata.Link{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
					ActorUID:    uid,
				}).
				Return(nil)

			sender := NewSender(shareCodeStore, tc.notifyClient(provided), "http://app", testRandomStringFn, nil, nil)
			TestAppData.Localizer = tc.localizer()

			err := sender.SendVoucherAccessCode(ctx, provided, TestAppData)
			assert.Nil(t, err)
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

	sender := NewSender(shareCodeStore, nil, "http://app", testRandomStringFn, nil, nil)

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
	}, TestAppData)

	assert.Equal(t, fmt.Errorf("creating share failed: %w", expectedError), err)
}

func TestSendVoucherAccessCodeWhenNotifyClientError(t *testing.T) {
	testcases := map[string]struct {
		email        string
		mobile       string
		notifyClient func() *mockNotifyClient
		localizer    func() *mockLocalizer
		error        error
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
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			uid := actoruid.New()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil)

			TestAppData.Localizer = tc.localizer()

			sender := NewSender(shareCodeStore, tc.notifyClient(), "http://app", testRandomStringFn, nil, nil)

			err := sender.SendVoucherAccessCode(ctx, &donordata.Provided{
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
				Voucher: donordata.Voucher{
					UID:        uid,
					FirstNames: "c",
					LastName:   "d",
				},
			}, TestAppData)

			assert.Equal(t, tc.error, err)
		})
	}
}
