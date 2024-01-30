package page

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
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
		Type:   actor.LpaTypePropertyAndAffairs,
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
		Put(ctx, actor.TypeCertificateProvider, RandomString, actor.ShareCodeData{
			LpaID:           "lpa-id",
			DonorFullname:   "Jan Smith",
			DonorFirstNames: "Jan",
			SessionID:       "session-id",
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.CertificateProviderInviteEmail{
			ShareCode:                   RandomString,
			CertificateProviderFullName: "Joanna Jones",
			DonorFirstNames:             "Jan",
			DonorFullName:               "Jan Smith",
			LpaType:                     "property and affairs",
			CertificateProviderStartURL: fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
			DonorFirstNamesPossessive:   "Jan’s",
			WhatLpaCovers:               "houses and stuff",
		}).
		Return(nil)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandomString, nil)
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
			expectedTestCode: RandomString,
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
					LpaID:           "lpa-id",
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
					SessionID:       "session-id",
				}).
				Once().
				Return(nil)

			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, RandomString, actor.ShareCodeData{
					LpaID:           "lpa-id",
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
					SessionID:       "session-id",
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
					ShareCode:                   RandomString,
					DonorFirstNamesPossessive:   "Jan’s",
					WhatLpaCovers:               "houses and stuff",
				}).
				Once().
				Return(nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandomString, nil)

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

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandomString, nil)
	err := sender.SendCertificateProviderInvite(ctx, TestAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderInviteWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("")
	localizer.EXPECT().
		Possessive(mock.Anything).
		Return("")
	TestAppData.Localizer = localizer

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", MockRandomString, nil)
	err := sender.SendCertificateProviderInvite(ctx, TestAppData, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderPrompt(t *testing.T) {
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

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(donor.Type.String()).
		Return("Property and affairs").
		Once()
	TestAppData.Localizer = localizer

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeCertificateProvider, RandomString, actor.ShareCodeData{
			LpaID:           "lpa-id",
			DonorFullname:   "Jan Smith",
			DonorFirstNames: "Jan",
			SessionID:       "session-id",
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, "name@example.org", "lpa-uid", notify.CertificateProviderProvideCertificatePromptEmail{
			ShareCode:                   RandomString,
			CertificateProviderFullName: "Joanna Jones",
			DonorFullName:               "Jan Smith",
			LpaType:                     "property and affairs",
			CertificateProviderStartURL: fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
		}).
		Return(nil)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandomString, nil)
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
			expectedTestCode: RandomString,
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
					LpaID:           "lpa-id",
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
					SessionID:       "session-id",
				}).
				Once().
				Return(nil)

			shareCodeStore.EXPECT().
				Put(ctx, actor.TypeCertificateProvider, RandomString, actor.ShareCodeData{
					LpaID:           "lpa-id",
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
					SessionID:       "session-id",
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
					ShareCode:                   RandomString,
				}).
				Once().
				Return(nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandomString, nil)

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

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandomString, nil)
	err := sender.SendCertificateProviderPrompt(ctx, TestAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderPromptWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("")

	TestAppData.Localizer = localizer

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", MockRandomString, nil)
	err := sender.SendCertificateProviderPrompt(ctx, TestAppData, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendAttorneys(t *testing.T) {
	donor := &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{
			TrustCorporation: actor.TrustCorporation{
				Name:  "Trusty",
				Email: "trusted@example.com",
			},
			Attorneys: []actor.Attorney{
				{
					ID:         "1",
					FirstNames: "Joanna",
					LastName:   "Jones",
					Email:      "name@example.org",
				},
				{
					ID:         "2",
					FirstNames: "John",
					LastName:   "Jones",
					Email:      "name2@example.org",
				},
				{
					ID:         "3",
					FirstNames: "Nope",
					LastName:   "Jones",
				},
			},
		},
		ReplacementAttorneys: actor.Attorneys{
			TrustCorporation: actor.TrustCorporation{
				Name:  "Untrusty",
				Email: "untrusted@example.com",
			},
			Attorneys: []actor.Attorney{
				{
					ID:         "4",
					FirstNames: "Dave",
					LastName:   "Davis",
					Email:      "dave@example.com",
				},
				{
					ID:         "5",
					FirstNames: "Donny",
					LastName:   "Davis",
				},
			},
		},
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

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, RandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", IsTrustCorporation: true}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, RandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", IsTrustCorporation: true, IsReplacementAttorney: true}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, RandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", AttorneyID: "1"}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, RandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", AttorneyID: "2"}).
		Return(nil)
	shareCodeStore.EXPECT().
		Put(ctx, actor.TypeAttorney, RandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", AttorneyID: "4", IsReplacementAttorney: true}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, "trusted@example.com", "lpa-uid", notify.InitialOriginalAttorneyEmail{
			ShareCode:                 RandomString,
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
			ShareCode:                 RandomString,
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
			ShareCode:                 RandomString,
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
			ShareCode:                 RandomString,
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
			ShareCode:                 RandomString,
			AttorneyFullName:          "Dave Davis",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
		}).
		Return(nil)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandomString, nil)
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
			expectedTestCode: RandomString,
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
				Put(ctx, actor.TypeAttorney, RandomString, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id"}).
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
					ShareCode:                 RandomString,
					AttorneyFullName:          "Joanna Jones",
					DonorFirstNames:           "Jan",
					DonorFullName:             "Jan Smith",
					DonorFirstNamesPossessive: "Jan's",
					LpaType:                   "property and affairs",
					AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
				}).
				Return(nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandomString, nil)

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

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandomString, nil)
	err := sender.SendAttorneys(ctx, TestAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendAttorneysWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("property and affairs")
	localizer.EXPECT().
		Possessive(mock.Anything).
		Return("Jan's")
	TestAppData.Localizer = localizer

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", MockRandomString, nil)
	err := sender.SendAttorneys(ctx, TestAppData, &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{Email: "hey@example.com"}}},
	})

	assert.Equal(t, expectedError, errors.Unwrap(err))
}
