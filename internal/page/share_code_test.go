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

func TestShareCodeSenderSendCertificateProvider(t *testing.T) {
	localizer := newMockLocalizer(t)
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

	testcases := map[notify.Template]struct {
		personalisation map[string]string
		localizerSetup  func(*mockLocalizer) *mockLocalizer
	}{
		notify.CertificateProviderInviteEmail: {
			personalisation: map[string]string{
				"shareCode":                   "123",
				"cpFullName":                  "Joanna Jones",
				"donorFirstNames":             "Jan",
				"donorFullName":               "Jan Smith",
				"lpaLegalTerm":                "property and affairs",
				"certificateProviderStartURL": fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
				"donorFirstNamesPossessive":   "Jan’s",
				"whatLPACovers":               "houses and stuff",
			},
			localizerSetup: func(localizer *mockLocalizer) *mockLocalizer {
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

				return localizer
			},
		},
		notify.Template(99): {
			personalisation: map[string]string{
				"shareCode":                   "123",
				"cpFullName":                  "Joanna Jones",
				"donorFullName":               "Jan Smith",
				"lpaLegalTerm":                "property and affairs",
				"certificateProviderStartURL": fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
			},
			localizerSetup: func(localizer *mockLocalizer) *mockLocalizer {
				localizer.
					On("T", donor.Type.LegalTermTransKey()).
					Return("property and affairs").
					Once()

				return localizer
			},
		},
	}

	for template, tc := range testcases {
		t.Run(string(template), func(t *testing.T) {
			tc.localizerSetup(localizer)
			TestAppData.Localizer = localizer

			ctx := context.Background()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.
				On("Put", ctx, actor.TypeCertificateProvider, "123", actor.ShareCodeData{
					LpaID:           "lpa-id",
					Identity:        true,
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
					SessionID:       "session-id",
				}).
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.
				On("TemplateID", template).
				Return("template-id")
			notifyClient.
				On("Email", ctx, notify.Email{
					TemplateID:      "template-id",
					EmailAddress:    "name@example.org",
					Personalisation: tc.personalisation,
				}).
				Return("", nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)
			err := sender.SendCertificateProvider(ctx, template, TestAppData, true, donor)

			assert.Nil(t, err)
		})
	}
}

func TestShareCodeSenderSendCertificateProviderWithTestCode(t *testing.T) {
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
					Identity:        true,
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
					SessionID:       "session-id",
				}).
				Once().
				Return(nil)

			shareCodeStore.
				On("Put", ctx, actor.TypeCertificateProvider, "123", actor.ShareCodeData{
					LpaID:           "lpa-id",
					Identity:        true,
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
					SessionID:       "session-id",
				}).
				Once().
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.
				On("TemplateID", notify.Template(99)).
				Return("template-id")
			notifyClient.
				On("Email", ctx, notify.Email{
					TemplateID:   "template-id",
					EmailAddress: "name@example.org",
					Personalisation: map[string]string{
						"cpFullName":                  "Joanna Jones",
						"donorFullName":               "Jan Smith",
						"lpaLegalTerm":                "property and affairs",
						"certificateProviderStartURL": fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
						"shareCode":                   tc.expectedTestCode,
					},
				}).
				Once().
				Return("", nil)
			notifyClient.
				On("Email", ctx, notify.Email{
					TemplateID:   "template-id",
					EmailAddress: "name@example.org",
					Personalisation: map[string]string{
						"cpFullName":                  "Joanna Jones",
						"donorFullName":               "Jan Smith",
						"lpaLegalTerm":                "property and affairs",
						"certificateProviderStartURL": fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
						"shareCode":                   "123",
					},
				}).
				Once().
				Return("", nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)

			if tc.useTestCode {
				sender.UseTestCode()
			}

			err := sender.SendCertificateProvider(ctx, notify.Template(99), TestAppData, true, donor)

			assert.Nil(t, err)

			err = sender.SendCertificateProvider(ctx, notify.Template(99), TestAppData, true, donor)

			assert.Nil(t, err)
		})
	}
}

func TestShareCodeSenderSendCertificateProviderWhenEmailErrors(t *testing.T) {
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
		On("TemplateID", notify.Template(99)).
		Return("")
	notifyClient.
		On("Email", ctx, mock.Anything).
		Return("", ExpectedError)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)
	err := sender.SendCertificateProvider(ctx, notify.Template(99), TestAppData, true, donor)

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(ExpectedError)

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", MockRandom)
	err := sender.SendCertificateProvider(ctx, notify.Template(99), TestAppData, true, &actor.DonorProvidedDetails{})

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
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
		On("TemplateID", notify.TrustCorporationInviteEmail).
		Return("trust-template-id")
	notifyClient.
		On("TemplateID", notify.ReplacementTrustCorporationInviteEmail).
		Return("trust-template-id2")
	notifyClient.
		On("TemplateID", notify.AttorneyInviteEmail).
		Return("template-id")
	notifyClient.
		On("TemplateID", notify.ReplacementAttorneyInviteEmail).
		Return("template-id2")
	notifyClient.
		On("Email", ctx, notify.Email{
			TemplateID:   "trust-template-id",
			EmailAddress: "trusted@example.com",
			Personalisation: map[string]string{
				"shareCode":                 "123",
				"attorneyFullName":          "Trusty",
				"donorFirstNames":           "Jan",
				"donorFullName":             "Jan Smith",
				"donorFirstNamesPossessive": "Jan's",
				"lpaLegalTerm":              "property and affairs",
				"landingPageLink":           fmt.Sprintf("http://app%s", Paths.Attorney.Start),
			},
		}).
		Return("", nil)
	notifyClient.
		On("Email", ctx, notify.Email{
			TemplateID:   "trust-template-id2",
			EmailAddress: "untrusted@example.com",
			Personalisation: map[string]string{
				"shareCode":                 "123",
				"attorneyFullName":          "Untrusty",
				"donorFirstNames":           "Jan",
				"donorFullName":             "Jan Smith",
				"donorFirstNamesPossessive": "Jan's",
				"lpaLegalTerm":              "property and affairs",
				"landingPageLink":           fmt.Sprintf("http://app%s", Paths.Attorney.Start),
			},
		}).
		Return("", nil)
	notifyClient.
		On("Email", ctx, notify.Email{
			TemplateID:   "template-id",
			EmailAddress: "name@example.org",
			Personalisation: map[string]string{
				"shareCode":                 "123",
				"attorneyFullName":          "Joanna Jones",
				"donorFirstNames":           "Jan",
				"donorFullName":             "Jan Smith",
				"donorFirstNamesPossessive": "Jan's",
				"lpaLegalTerm":              "property and affairs",
				"landingPageLink":           fmt.Sprintf("http://app%s", Paths.Attorney.Start),
			},
		}).
		Return("", nil)
	notifyClient.
		On("Email", ctx, notify.Email{
			TemplateID:   "template-id",
			EmailAddress: "name2@example.org",
			Personalisation: map[string]string{
				"shareCode":                 "123",
				"attorneyFullName":          "John Jones",
				"donorFirstNames":           "Jan",
				"donorFullName":             "Jan Smith",
				"donorFirstNamesPossessive": "Jan's",
				"lpaLegalTerm":              "property and affairs",
				"landingPageLink":           fmt.Sprintf("http://app%s", Paths.Attorney.Start),
			},
		}).
		Return("", nil)
	notifyClient.
		On("Email", ctx, notify.Email{
			TemplateID:   "template-id2",
			EmailAddress: "dave@example.com",
			Personalisation: map[string]string{
				"shareCode":                 "123",
				"attorneyFullName":          "Dave Davis",
				"donorFirstNames":           "Jan",
				"donorFullName":             "Jan Smith",
				"donorFirstNamesPossessive": "Jan's",
				"lpaLegalTerm":              "property and affairs",
				"landingPageLink":           fmt.Sprintf("http://app%s", Paths.Attorney.Start),
			},
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
				On("TemplateID", notify.Template(notify.AttorneyInviteEmail)).
				Return("template-id")
			notifyClient.
				On("Email", ctx, notify.Email{
					TemplateID:   "template-id",
					EmailAddress: "name@example.org",
					Personalisation: map[string]string{
						"shareCode":                 tc.expectedTestCode,
						"attorneyFullName":          "Joanna Jones",
						"donorFirstNames":           "Jan",
						"donorFullName":             "Jan Smith",
						"donorFirstNamesPossessive": "Jan's",
						"lpaLegalTerm":              "property and affairs",
						"landingPageLink":           fmt.Sprintf("http://app%s", Paths.Attorney.Start),
					},
				}).
				Return("", nil)
			notifyClient.
				On("Email", ctx, notify.Email{
					TemplateID:   "template-id",
					EmailAddress: "name@example.org",
					Personalisation: map[string]string{
						"shareCode":                 "123",
						"attorneyFullName":          "Joanna Jones",
						"donorFirstNames":           "Jan",
						"donorFullName":             "Jan Smith",
						"donorFirstNamesPossessive": "Jan's",
						"lpaLegalTerm":              "property and affairs",
						"landingPageLink":           fmt.Sprintf("http://app%s", Paths.Attorney.Start),
					},
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
		On("TemplateID", mock.Anything).
		Return("template-id")
	notifyClient.
		On("Email", ctx, mock.Anything).
		Return("", ExpectedError)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)
	err := sender.SendAttorneys(ctx, TestAppData, donor)

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendAttorneysWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(ExpectedError)

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", MockRandom)
	err := sender.SendAttorneys(ctx, TestAppData, &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{Email: "hey@example.com"}}},
	})

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}
