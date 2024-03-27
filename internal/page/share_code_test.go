package page

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShareCodeSenderSendCertificateProviderInvite(t *testing.T) {
	donor := &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:    actor.LpaTypePropertyAndAffairs,
		LpaUID:  "lpa-uid",
		Channel: actor.Online,
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
		Put(ctx, actor.TypeCertificateProvider, testRandomString, actor.ShareCodeData{
			LpaID:        "lpa-id",
			SessionID:    "session-id",
			DonorChannel: actor.Online,
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.CertificateProviderInviteEmail{
			ShareCode:                   testRandomString,
			CertificateProviderFullName: "Joanna Jones",
			DonorFirstNames:             "Jan",
			DonorFullName:               "Jan Smith",
			LpaType:                     "property and affairs",
			CertificateProviderStartURL: fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
			DonorFirstNamesPossessive:   "Jan’s",
			WhatLpaCovers:               "houses and stuff",
		}).
		Return(nil)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil)
	err := sender.SendCertificateProviderInvite(ctx, TestAppData, donor)

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

	donor := &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:   actor.LpaTypePropertyAndAffairs,
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
				Put(ctx, actor.TypeCertificateProvider, tc.expectedTestCode, actor.ShareCodeData{
					LpaID:     "lpa-id",
					SessionID: "session-id",
				}).
				Once().
				Return(nil)

			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, testRandomString, actor.ShareCodeData{
					LpaID:     "lpa-id",
					SessionID: "session-id",
				}).
				Once().
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.CertificateProviderInviteEmail{
					CertificateProviderFullName: "Joanna Jones",
					DonorFirstNames:             "Jan",
					DonorFullName:               "Jan Smith",
					LpaType:                     "property and affairs",
					CertificateProviderStartURL: fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
					ShareCode:                   tc.expectedTestCode,
					DonorFirstNamesPossessive:   "Jan’s",
					WhatLpaCovers:               "houses and stuff",
				}).
				Once().
				Return(nil)
			notifyClient.EXPECT().
				SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.CertificateProviderInviteEmail{
					CertificateProviderFullName: "Joanna Jones",
					DonorFirstNames:             "Jan",
					DonorFullName:               "Jan Smith",
					LpaType:                     "property and affairs",
					CertificateProviderStartURL: fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
					ShareCode:                   testRandomString,
					DonorFirstNamesPossessive:   "Jan’s",
					WhatLpaCovers:               "houses and stuff",
				}).
				Once().
				Return(nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil)

			if tc.useTestCode {
				sender.UseTestCode("abcdef123456")
			}

			err := sender.SendCertificateProviderInvite(ctx, TestAppData, donor)
			assert.Nil(t, err)

			err = sender.SendCertificateProviderInvite(ctx, TestAppData, donor)
			assert.Nil(t, err)
		})
	}
}

func TestShareCodeSenderSendCertificateProviderInviteWhenEmailErrors(t *testing.T) {
	ctx := context.Background()

	donor := &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type: actor.LpaTypePropertyAndAffairs,
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
	err := sender.SendCertificateProviderInvite(ctx, TestAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderInviteWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", testRandomStringFn, nil)
	err := sender.SendCertificateProviderInvite(ctx, TestAppData, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderPromptOnline(t *testing.T) {
	donor := &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
			CarryOutBy: actor.Online,
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:    actor.LpaTypePropertyAndAffairs,
		LpaUID:  "lpa-uid",
		Channel: actor.Online,
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
			CertificateProviderStartURL: fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
		}).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeCertificateProvider, testRandomString, actor.ShareCodeData{
			LpaID:        "lpa-id",
			SessionID:    "session-id",
			DonorChannel: actor.Online,
		}).
		Return(nil)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", testRandomStringFn, nil)
	err := sender.SendCertificateProviderPrompt(ctx, TestAppData, donor)

	assert.Nil(t, err)
}

func TestShareCodeSenderSendCertificateProviderPromptPaper(t *testing.T) {
	actorUID := actoruid.New()

	donor := &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{
			UID:        actorUID,
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
			CarryOutBy: actor.Paper,
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:   actor.LpaTypePropertyAndAffairs,
		LpaUID: "lpa-uid",
	}

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeCertificateProvider, testRandomString, actor.ShareCodeData{
			SessionID: "session-id",
			LpaID:     "lpa-id",
			ActorUID:  actorUID,
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

	donor := &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:   actor.LpaTypePropertyAndAffairs,
		LpaUID: "lpa-uid",
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
				Put(ctx, actor.TypeCertificateProvider, tc.expectedTestCode, actor.ShareCodeData{
					LpaID:     "lpa-id",
					SessionID: "session-id",
				}).
				Once().
				Return(nil)

			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, testRandomString, actor.ShareCodeData{
					LpaID:     "lpa-id",
					SessionID: "session-id",
				}).
				Once().
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
					CertificateProviderFullName: "Joanna Jones",
					DonorFullName:               "Jan Smith",
					LpaType:                     "property and affairs",
					CertificateProviderStartURL: fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
					ShareCode:                   tc.expectedTestCode,
				}).
				Once().
				Return(nil)
			notifyClient.EXPECT().
				SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
					CertificateProviderFullName: "Joanna Jones",
					DonorFullName:               "Jan Smith",
					LpaType:                     "property and affairs",
					CertificateProviderStartURL: fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
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
	donor := &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{
			CarryOutBy: actor.Paper,
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
	donor := &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{
			CarryOutBy: actor.Paper,
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

	donor := &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type: actor.LpaTypePropertyAndAffairs,
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
	err := sender.SendCertificateProviderPrompt(ctx, TestAppData, &actor.DonorProvidedDetails{})

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

	donor := &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{
			TrustCorporation: actor.TrustCorporation{
				UID:   trustCorporationUID,
				Name:  "Trusty",
				Email: "trusted@example.com",
			},
			Attorneys: []actor.Attorney{
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
		ReplacementAttorneys: actor.Attorneys{
			TrustCorporation: actor.TrustCorporation{
				UID:   replacementTrustCorporationUID,
				Name:  "Untrusty",
				Email: "untrusted@example.com",
			},
			Attorneys: []actor.Attorney{
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
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:    actor.LpaTypePropertyAndAffairs,
		LpaUID:  "lpa-uid",
		Channel: actor.Online,
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
		Put(ctx, actor.TypeTrustCorporation, testRandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", ActorUID: trustCorporationUID, IsTrustCorporation: true, DonorChannel: actor.Online}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementTrustCorporation, testRandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", ActorUID: replacementTrustCorporationUID, IsTrustCorporation: true, IsReplacementAttorney: true, DonorChannel: actor.Online}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, testRandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", ActorUID: attorney1UID, DonorChannel: actor.Online}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, testRandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", ActorUID: attorney2UID, DonorChannel: actor.Online}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, testRandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", ActorUID: attorney3UID, DonorChannel: actor.Online}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementAttorney, testRandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", ActorUID: replacement1UID, IsReplacementAttorney: true, DonorChannel: actor.Online}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementAttorney, testRandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", ActorUID: replacement2UID, IsReplacementAttorney: true, DonorChannel: actor.Online}).
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
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
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
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
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
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
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
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
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
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
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

	donor := &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{
			TrustCorporation: actor.TrustCorporation{
				UID:  uid1,
				Name: "Trusty",
			},
		},
		ReplacementAttorneys: actor.Attorneys{
			TrustCorporation: actor.TrustCorporation{
				UID:  uid2,
				Name: "Untrusty",
			},
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:    actor.LpaTypePropertyAndAffairs,
		LpaUID:  "lpa-uid",
		Channel: actor.Online,
	}

	ctx := context.Background()
	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeTrustCorporation, testRandomString, actor.ShareCodeData{
			SessionID:          "session-id",
			LpaID:              "lpa-id",
			ActorUID:           uid1,
			IsTrustCorporation: true,
			DonorChannel:       actor.Online,
		}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeReplacementTrustCorporation, testRandomString, actor.ShareCodeData{
			SessionID:             "session-id",
			LpaID:                 "lpa-id",
			ActorUID:              uid2,
			IsTrustCorporation:    true,
			IsReplacementAttorney: true,
			DonorChannel:          actor.Online,
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

	donor := &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{
				FirstNames: "Joanna",
				LastName:   "Jones",
				Email:      "name@example.org",
			},
		}},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type:   actor.LpaTypePropertyAndAffairs,
		LpaUID: "lpa-uid",
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
				Put(ctx, actor.TypeAttorney, tc.expectedTestCode, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id"}).
				Return(nil)
			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeAttorney, testRandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id"}).
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
					AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
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
					AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
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

	donor := &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{
				FirstNames: "Joanna",
				LastName:   "Jones",
				Email:      "name@example.org",
			},
		}},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type: actor.LpaTypePropertyAndAffairs,
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
	err := sender.SendAttorneys(ctx, TestAppData, &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{Email: "hey@example.com"}}},
	})

	assert.Equal(t, expectedError, errors.Unwrap(err))
}
