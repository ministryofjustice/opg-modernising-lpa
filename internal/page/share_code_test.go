package page

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(donor.Type.String()).
		Return("Property and affairs").
		Once()
	localizer.EXPECT().
		T(donor.Type.WhatLPACoversTransKey()).
		Return("houses and stuff").
		Once()
	localizer.EXPECT().
		Possessive("Jan").
		Return("Jan’s")
	TestAppData.Localizer = localizer

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeCertificateProvider, testRandomString, sharecode.Data{
			LpaKey:      dynamo.LpaKey("lpa"),
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.CertificateProviderInviteEmail{
			ShareCode:                    testRandomString,
			CertificateProviderFullName:  "Joanna Jones",
			DonorFirstNames:              "Jan",
			DonorFullName:                "Jan Smith",
			LpaType:                      "property and affairs",
			CertificateProviderStartURL:  fmt.Sprintf("http://app%s", PathCertificateProviderStart),
			DonorFirstNamesPossessive:    "Jan’s",
			WhatLpaCovers:                "houses and stuff",
			CertificateProviderOptOutURL: fmt.Sprintf("http://app%s", PathCertificateProviderEnterReferenceNumberOptOut),
		}).
		Return(nil)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil)
	err := sender.SendCertificateProviderInvite(ctx, TestAppData, CertificateProviderInvite{
		LpaKey:                      dynamo.LpaKey("lpa"),
		LpaOwnerKey:                 dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
		LpaUID:                      donor.LpaUID,
		Type:                        donor.Type,
		DonorFirstNames:             donor.Donor.FirstNames,
		DonorFullName:               donor.Donor.FullName(),
		CertificateProviderUID:      donor.CertificateProvider.UID,
		CertificateProviderFullName: donor.CertificateProvider.FullName(),
		CertificateProviderEmail:    donor.CertificateProvider.Email,
	})

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
		Type:   lpadata.LpaTypePropertyAndAffairs,
		LpaUID: "lpa-uid",
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T(donor.Type.String()).
				Return("Property and affairs").
				Twice()
			localizer.EXPECT().
				Possessive("Jan").
				Return("Jan’s")
			localizer.EXPECT().
				T(donor.Type.WhatLPACoversTransKey()).
				Return("houses and stuff")
			TestAppData.Localizer = localizer

			ctx := context.Background()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, tc.expectedTestCode, sharecode.Data{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				}).
				Once().
				Return(nil)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, testRandomString, sharecode.Data{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				}).
				Once().
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.CertificateProviderInviteEmail{
					CertificateProviderFullName:  "Joanna Jones",
					DonorFirstNames:              "Jan",
					DonorFullName:                "Jan Smith",
					LpaType:                      "property and affairs",
					CertificateProviderStartURL:  fmt.Sprintf("http://app%s", PathCertificateProviderStart),
					ShareCode:                    tc.expectedTestCode,
					DonorFirstNamesPossessive:    "Jan’s",
					WhatLpaCovers:                "houses and stuff",
					CertificateProviderOptOutURL: fmt.Sprintf("http://app%s", PathCertificateProviderEnterReferenceNumberOptOut),
				}).
				Once().
				Return(nil)
			notifyClient.EXPECT().
				SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.CertificateProviderInviteEmail{
					CertificateProviderFullName:  "Joanna Jones",
					DonorFirstNames:              "Jan",
					DonorFullName:                "Jan Smith",
					LpaType:                      "property and affairs",
					CertificateProviderStartURL:  fmt.Sprintf("http://app%s", PathCertificateProviderStart),
					ShareCode:                    testRandomString,
					DonorFirstNamesPossessive:    "Jan’s",
					WhatLpaCovers:                "houses and stuff",
					CertificateProviderOptOutURL: fmt.Sprintf("http://app%s", PathCertificateProviderEnterReferenceNumberOptOut),
				}).
				Once().
				Return(nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil)

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
				CertificateProviderEmail:    donor.CertificateProvider.Email,
			})
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
				CertificateProviderEmail:    donor.CertificateProvider.Email,
			})
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
	TestAppData.Localizer = localizer

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil)
	err := sender.SendCertificateProviderInvite(ctx, TestAppData, CertificateProviderInvite{
		LpaUID:                      donor.LpaUID,
		Type:                        donor.Type,
		DonorFirstNames:             donor.Donor.FirstNames,
		DonorFullName:               donor.Donor.FullName(),
		CertificateProviderUID:      donor.CertificateProvider.UID,
		CertificateProviderFullName: donor.CertificateProvider.FullName(),
		CertificateProviderEmail:    donor.CertificateProvider.Email,
	})

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderInviteWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", testRandomStringFn, nil)
	err := sender.SendCertificateProviderInvite(ctx, TestAppData, CertificateProviderInvite{})

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
		SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
			ShareCode:                   testRandomString,
			CertificateProviderFullName: "Joanna Jones",
			DonorFullName:               "Jan Smith",
			LpaType:                     "property and affairs",
			CertificateProviderStartURL: fmt.Sprintf("http://app%s", PathCertificateProviderStart),
		}).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeCertificateProvider, testRandomString, sharecode.Data{
			LpaKey:      dynamo.LpaKey("lpa"),
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
		}).
		Return(nil)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil)
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
		Put(ctx, actor.TypeCertificateProvider, testRandomString, sharecode.Data{
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

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", testRandomStringFn, eventClient)
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
				Put(ctx, actor.TypeCertificateProvider, tc.expectedTestCode, sharecode.Data{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				}).
				Once().
				Return(nil)

			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, testRandomString, sharecode.Data{
					LpaKey:      dynamo.LpaKey("lpa"),
					LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				}).
				Once().
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
					CertificateProviderFullName: "Joanna Jones",
					DonorFullName:               "Jan Smith",
					LpaType:                     "property and affairs",
					CertificateProviderStartURL: fmt.Sprintf("http://app%s", PathCertificateProviderStart),
					ShareCode:                   tc.expectedTestCode,
				}).
				Once().
				Return(nil)
			notifyClient.EXPECT().
				SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
					CertificateProviderFullName: "Joanna Jones",
					DonorFullName:               "Jan Smith",
					LpaType:                     "property and affairs",
					CertificateProviderStartURL: fmt.Sprintf("http://app%s", PathCertificateProviderStart),
					ShareCode:                   testRandomString,
				}).
				Once().
				Return(nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil)

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

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", testRandomStringFn, nil)
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

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", testRandomStringFn, eventClient)
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

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil)
	err := sender.SendCertificateProviderPrompt(ctx, TestAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderPromptWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", testRandomStringFn, nil)
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

	donor := &lpastore.Lpa{
		Attorneys: lpastore.Attorneys{
			TrustCorporation: lpastore.TrustCorporation{
				UID:   trustCorporationUID,
				Name:  "Trusty",
				Email: "trusted@example.com",
			},
			Attorneys: []lpastore.Attorney{
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
		ReplacementAttorneys: lpastore.Attorneys{
			TrustCorporation: lpastore.TrustCorporation{
				UID:   replacementTrustCorporationUID,
				Name:  "Untrusty",
				Email: "untrusted@example.com",
			},
			Attorneys: []lpastore.Attorney{
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
		Donor: lpastore.Donor{
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
		T(donor.Type.String()).
		Return("property and affairs")
	localizer.EXPECT().
		Possessive("Jan").
		Return("Jan's")

	TestAppData.Localizer = localizer

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeTrustCorporation, testRandomString, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: trustCorporationUID, IsTrustCorporation: true}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementTrustCorporation, testRandomString, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: replacementTrustCorporationUID, IsTrustCorporation: true, IsReplacementAttorney: true}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, testRandomString, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: attorney1UID}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, testRandomString, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: attorney2UID}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, testRandomString, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: attorney3UID}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementAttorney, testRandomString, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: replacement1UID, IsReplacementAttorney: true}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementAttorney, testRandomString, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa"), ActorUID: replacement2UID, IsReplacementAttorney: true}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, "trusted@example.com", "lpa-uid", notify.InitialOriginalAttorneyEmail{
			ShareCode:                 testRandomString,
			AttorneyFullName:          "Trusty",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", PathAttorneyStart),
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)
	notifyClient.EXPECT().
		SendActorEmail(ctx, "untrusted@example.com", "lpa-uid", notify.InitialReplacementAttorneyEmail{
			ShareCode:                 testRandomString,
			AttorneyFullName:          "Untrusty",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", PathAttorneyStart),
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)
	notifyClient.EXPECT().
		SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.InitialOriginalAttorneyEmail{
			ShareCode:                 testRandomString,
			AttorneyFullName:          "Joanna Jones",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", PathAttorneyStart),
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)
	notifyClient.EXPECT().
		SendActorEmail(ctx, "name2@example.org", "lpa-uid", notify.InitialOriginalAttorneyEmail{
			ShareCode:                 testRandomString,
			AttorneyFullName:          "John Jones",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", PathAttorneyStart),
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", PathAttorneyEnterReferenceNumberOptOut),
		}).
		Return(nil)
	notifyClient.EXPECT().
		SendActorEmail(ctx, "dave@example.com", "lpa-uid", notify.InitialReplacementAttorneyEmail{
			ShareCode:                 testRandomString,
			AttorneyFullName:          "Dave Davis",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", PathAttorneyStart),
			AttorneyOptOutURL:         fmt.Sprintf("http://app%s", PathAttorneyEnterReferenceNumberOptOut),
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

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, eventClient)
	err := sender.SendAttorneys(ctx, TestAppData, donor)

	assert.Nil(t, err)
}

func TestShareCodeSenderSendAttorneysTrustCorporationsNoEmail(t *testing.T) {
	uid1 := actoruid.New()
	uid2 := actoruid.New()

	donor := &lpastore.Lpa{
		Attorneys: lpastore.Attorneys{
			TrustCorporation: lpastore.TrustCorporation{
				UID:  uid1,
				Name: "Trusty",
			},
		},
		ReplacementAttorneys: lpastore.Attorneys{
			TrustCorporation: lpastore.TrustCorporation{
				UID:  uid2,
				Name: "Untrusty",
			},
		},
		Donor: lpastore.Donor{
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
		Put(ctx, actor.TypeTrustCorporation, testRandomString, sharecode.Data{
			LpaOwnerKey:        dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			LpaKey:             dynamo.LpaKey("lpa"),
			ActorUID:           uid1,
			IsTrustCorporation: true,
		}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementTrustCorporation, testRandomString, sharecode.Data{
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

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", testRandomStringFn, eventClient)
	err := sender.SendAttorneys(ctx, TestAppData, donor)

	assert.Nil(t, err)
}

func TestShareCodeSenderSendAttorneysWithTestCode(t *testing.T) {
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

	donor := &lpastore.Lpa{
		Attorneys: lpastore.Attorneys{Attorneys: []lpastore.Attorney{
			{
				FirstNames: "Joanna",
				LastName:   "Jones",
				Email:      "name@example.org",
			},
		}},
		Donor: lpastore.Donor{
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
		T(donor.Type.String()).
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
				Put(ctx, actor.TypeAttorney, tc.expectedTestCode, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa")}).
				Return(nil)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeAttorney, testRandomString, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaKey: dynamo.LpaKey("lpa")}).
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.InitialOriginalAttorneyEmail{
					ShareCode:                 tc.expectedTestCode,
					AttorneyFullName:          "Joanna Jones",
					DonorFirstNames:           "Jan",
					DonorFullName:             "Jan Smith",
					DonorFirstNamesPossessive: "Jan's",
					LpaType:                   "property and affairs",
					AttorneyStartPageURL:      fmt.Sprintf("http://app%s", PathAttorneyStart),
					AttorneyOptOutURL:         fmt.Sprintf("http://app%s", PathAttorneyEnterReferenceNumberOptOut),
				}).
				Return(nil)
			notifyClient.EXPECT().
				SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.InitialOriginalAttorneyEmail{
					ShareCode:                 testRandomString,
					AttorneyFullName:          "Joanna Jones",
					DonorFirstNames:           "Jan",
					DonorFullName:             "Jan Smith",
					DonorFirstNamesPossessive: "Jan's",
					LpaType:                   "property and affairs",
					AttorneyStartPageURL:      fmt.Sprintf("http://app%s", PathAttorneyStart),
					AttorneyOptOutURL:         fmt.Sprintf("http://app%s", PathAttorneyEnterReferenceNumberOptOut),
				}).
				Return(nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil)

			if tc.useTestCode {
				sender.UseTestCode("abcdef123456")
			}

			err := sender.SendAttorneys(ctx, TestAppData, donor)
			assert.Nil(t, err)

			err = sender.SendAttorneys(ctx, TestAppData, donor)
			assert.Nil(t, err)
		})
	}
}

func TestShareCodeSenderSendAttorneysWhenEmailErrors(t *testing.T) {
	ctx := context.Background()

	donor := &lpastore.Lpa{
		Attorneys: lpastore.Attorneys{Attorneys: []lpastore.Attorney{
			{
				FirstNames: "Joanna",
				LastName:   "Jones",
				Email:      "name@example.org",
			},
		}},
		Donor: lpastore.Donor{
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

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil)
	err := sender.SendAttorneys(ctx, TestAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendAttorneysWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", testRandomStringFn, nil)
	err := sender.SendAttorneys(ctx, TestAppData, &lpastore.Lpa{
		Attorneys: lpastore.Attorneys{Attorneys: []lpastore.Attorney{{Email: "hey@example.com"}}},
	})

	assert.Equal(t, expectedError, errors.Unwrap(err))
}
