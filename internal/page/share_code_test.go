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
		Type: actor.LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", donor.Type.LegalTermTransKey()).
		Return("property and affairs").
		Once()
	localizer.
		On("T", donor.Type.WhatLPACoversTransKey()).
		Return("houses and stuff").
		Once()
	localizer.
		On("Possessive", "Jan").
		Return("Jan’s")
	TestAppData.Localizer = localizer

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Put", ctx, actor.TypeCertificateProvider, "123", actor.ShareCodeData{
			LpaID:           "lpa-id",
			DonorFullname:   "Jan Smith",
			DonorFirstNames: "Jan",
			SessionID:       "session-id",
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("SendEmail", ctx, "name@example.org", notify.CertificateProviderInviteEmail{
			ShareCode:                   "123",
			CertificateProviderFullName: "Joanna Jones",
			DonorFirstNames:             "Jan",
			DonorFullName:               "Jan Smith",
			LpaType:                     "property and affairs",
			CertificateProviderStartURL: fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
			DonorFirstNamesPossessive:   "Jan’s",
			WhatLpaCovers:               "houses and stuff",
		}).
		Return("", nil)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)
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
			expectedTestCode: "123",
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
		Type: actor.LpaTypePropertyFinance,
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			localizer := newMockLocalizer(t)
			localizer.
				On("T", donor.Type.LegalTermTransKey()).
				Return("property and affairs").
				Twice()
			localizer.
				On("Possessive", "Jan").
				Return("Jan’s")
			localizer.
				On("T", donor.Type.WhatLPACoversTransKey()).
				Return("houses and stuff")
			TestAppData.Localizer = localizer

			ctx := context.Background()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.
				On("Put", ctx, actor.TypeCertificateProvider, tc.expectedTestCode, actor.ShareCodeData{
					LpaID:           "lpa-id",
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
					SessionID:       "session-id",
				}).
				Once().
				Return(nil)

			shareCodeStore.
				On("Put", ctx, actor.TypeCertificateProvider, "123", actor.ShareCodeData{
					LpaID:           "lpa-id",
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
					SessionID:       "session-id",
				}).
				Once().
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.
				On("SendEmail", ctx, "name@example.org", notify.CertificateProviderInviteEmail{
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
				Return("", nil)
			notifyClient.
				On("SendEmail", ctx, "name@example.org", notify.CertificateProviderInviteEmail{
					CertificateProviderFullName: "Joanna Jones",
					DonorFirstNames:             "Jan",
					DonorFullName:               "Jan Smith",
					LpaType:                     "property and affairs",
					CertificateProviderStartURL: fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
					ShareCode:                   "123",
					DonorFirstNamesPossessive:   "Jan’s",
					WhatLpaCovers:               "houses and stuff",
				}).
				Once().
				Return("", nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)

			if tc.useTestCode {
				sender.UseTestCode()
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
		Type: actor.LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", mock.Anything).
		Return("")
	localizer.
		On("Possessive", "Jan").
		Return("Jan’s")
	TestAppData.Localizer = localizer

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("SendEmail", ctx, mock.Anything, mock.Anything).
		Return("", expectedError)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)
	err := sender.SendCertificateProviderInvite(ctx, TestAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderInviteWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	localizer := newMockLocalizer(t)
	localizer.
		On("T", mock.Anything).
		Return("")
	localizer.
		On("Possessive", mock.Anything).
		Return("")
	TestAppData.Localizer = localizer

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", MockRandom)
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
		Type: actor.LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", donor.Type.LegalTermTransKey()).
		Return("property and affairs").
		Once()
	TestAppData.Localizer = localizer

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Put", ctx, actor.TypeCertificateProvider, "123", actor.ShareCodeData{
			LpaID:           "lpa-id",
			DonorFullname:   "Jan Smith",
			DonorFirstNames: "Jan",
			SessionID:       "session-id",
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("SendEmail", ctx, "name@example.org", notify.CertificateProviderProvideCertificatePromptEmail{
			ShareCode:                   "123",
			CertificateProviderFullName: "Joanna Jones",
			DonorFullName:               "Jan Smith",
			LpaType:                     "property and affairs",
			CertificateProviderStartURL: fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
		}).
		Return("", nil)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)
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
			expectedTestCode: "123",
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
		Type: actor.LpaTypePropertyFinance,
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			localizer := newMockLocalizer(t)
			localizer.
				On("T", donor.Type.LegalTermTransKey()).
				Return("property and affairs").
				Twice()

			TestAppData.Localizer = localizer

			ctx := context.Background()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.
				On("Put", ctx, actor.TypeCertificateProvider, tc.expectedTestCode, actor.ShareCodeData{
					LpaID:           "lpa-id",
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
					SessionID:       "session-id",
				}).
				Once().
				Return(nil)

			shareCodeStore.
				On("Put", ctx, actor.TypeCertificateProvider, "123", actor.ShareCodeData{
					LpaID:           "lpa-id",
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
					SessionID:       "session-id",
				}).
				Once().
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.
				On("SendEmail", ctx, "name@example.org", notify.CertificateProviderProvideCertificatePromptEmail{
					CertificateProviderFullName: "Joanna Jones",
					DonorFullName:               "Jan Smith",
					LpaType:                     "property and affairs",
					CertificateProviderStartURL: fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
					ShareCode:                   tc.expectedTestCode,
				}).
				Once().
				Return("", nil)
			notifyClient.
				On("SendEmail", ctx, "name@example.org", notify.CertificateProviderProvideCertificatePromptEmail{
					CertificateProviderFullName: "Joanna Jones",
					DonorFullName:               "Jan Smith",
					LpaType:                     "property and affairs",
					CertificateProviderStartURL: fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
					ShareCode:                   "123",
				}).
				Once().
				Return("", nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)

			if tc.useTestCode {
				sender.UseTestCode()
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
		Type: actor.LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", mock.Anything).
		Return("")

	TestAppData.Localizer = localizer

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("SendEmail", ctx, mock.Anything, mock.Anything).
		Return("", expectedError)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)
	err := sender.SendCertificateProviderPrompt(ctx, TestAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderPromptWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", MockRandom)
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
		Type: actor.LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", donor.Type.LegalTermTransKey()).
		Return("property and affairs")
	localizer.
		On("Possessive", "Jan").
		Return("Jan's")

	TestAppData.Localizer = localizer

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Put", ctx, actor.TypeAttorney, "123", actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", IsTrustCorporation: true}).
		Return(nil)
	shareCodeStore.
		On("Put", ctx, actor.TypeAttorney, "123", actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", IsTrustCorporation: true, IsReplacementAttorney: true}).
		Return(nil)
	shareCodeStore.
		On("Put", ctx, actor.TypeAttorney, "123", actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", AttorneyID: "1"}).
		Return(nil)
	shareCodeStore.
		On("Put", ctx, actor.TypeAttorney, "123", actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", AttorneyID: "2"}).
		Return(nil)
	shareCodeStore.
		On("Put", ctx, actor.TypeAttorney, "123", actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", AttorneyID: "4", IsReplacementAttorney: true}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("SendEmail", ctx, "trusted@example.com", notify.InitialOriginalAttorneyEmail{
			ShareCode:                 "123",
			AttorneyFullName:          "Trusty",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
		}).
		Return("", nil)
	notifyClient.
		On("SendEmail", ctx, "untrusted@example.com", notify.InitialReplacementAttorneyEmail{
			ShareCode:                 "123",
			AttorneyFullName:          "Untrusty",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
		}).
		Return("", nil)
	notifyClient.
		On("SendEmail", ctx, "name@example.org", notify.InitialOriginalAttorneyEmail{
			ShareCode:                 "123",
			AttorneyFullName:          "Joanna Jones",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
		}).
		Return("", nil)
	notifyClient.
		On("SendEmail", ctx, "name2@example.org", notify.InitialOriginalAttorneyEmail{
			ShareCode:                 "123",
			AttorneyFullName:          "John Jones",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
		}).
		Return("", nil)
	notifyClient.
		On("SendEmail", ctx, "dave@example.com", notify.InitialReplacementAttorneyEmail{
			ShareCode:                 "123",
			AttorneyFullName:          "Dave Davis",
			DonorFirstNames:           "Jan",
			DonorFullName:             "Jan Smith",
			DonorFirstNamesPossessive: "Jan's",
			LpaType:                   "property and affairs",
			AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
		}).
		Return("", nil)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)
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
			expectedTestCode: "123",
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
		Type: actor.LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", donor.Type.LegalTermTransKey()).
		Return("property and affairs")
	localizer.
		On("Possessive", "Jan").
		Return("Jan's")

	TestAppData.Localizer = localizer

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.
				On("Put", ctx, actor.TypeAttorney, tc.expectedTestCode, actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id"}).
				Return(nil)
			shareCodeStore.
				On("Put", ctx, actor.TypeAttorney, "123", actor.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id"}).
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.
				On("SendEmail", ctx, "name@example.org", notify.InitialOriginalAttorneyEmail{
					ShareCode:                 tc.expectedTestCode,
					AttorneyFullName:          "Joanna Jones",
					DonorFirstNames:           "Jan",
					DonorFullName:             "Jan Smith",
					DonorFirstNamesPossessive: "Jan's",
					LpaType:                   "property and affairs",
					AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
				}).
				Return("", nil)
			notifyClient.
				On("SendEmail", ctx, "name@example.org", notify.InitialOriginalAttorneyEmail{
					ShareCode:                 "123",
					AttorneyFullName:          "Joanna Jones",
					DonorFirstNames:           "Jan",
					DonorFullName:             "Jan Smith",
					DonorFirstNamesPossessive: "Jan's",
					LpaType:                   "property and affairs",
					AttorneyStartPageURL:      fmt.Sprintf("http://app%s", Paths.Attorney.Start),
				}).
				Return("", nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)

			if tc.useTestCode {
				sender.UseTestCode()
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
		Type: actor.LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", donor.Type.LegalTermTransKey()).
		Return("property and affairs")
	localizer.
		On("Possessive", "Jan").
		Return("Jan's")
	TestAppData.Localizer = localizer

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("SendEmail", ctx, mock.Anything, mock.Anything).
		Return("", expectedError)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)
	err := sender.SendAttorneys(ctx, TestAppData, donor)

	assert.Equal(t, expectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendAttorneysWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	localizer := newMockLocalizer(t)
	localizer.
		On("T", "").
		Return("property and affairs")
	localizer.
		On("Possessive", "").
		Return("Jan's")
	TestAppData.Localizer = localizer

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", MockRandom)
	err := sender.SendAttorneys(ctx, TestAppData, &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{Email: "hey@example.com"}}},
	})

	assert.Equal(t, expectedError, errors.Unwrap(err))
}
