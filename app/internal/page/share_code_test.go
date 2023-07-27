package page

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShareCodeSenderSendCertificateProvider(t *testing.T) {
	testcases := map[string]bool{
		"identity": true,
		"sign in":  false,
	}
	lpa := &Lpa{
		CertificateProvider: actor.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type: LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", lpa.Type.LegalTermTransKey()).
		Return("property and affairs")
	localizer.
		On("Possessive", "Jan").
		Return("Jan’s")

	TestAppData.Localizer = localizer

	for name, identity := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.
				On("Put", ctx, actor.TypeCertificateProvider, "123", actor.ShareCodeData{
					LpaID:           "lpa-id",
					Identity:        identity,
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
				}).
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.
				On("TemplateID", notify.TemplateId(99)).
				Return("template-id")
			notifyClient.
				On("Email", ctx, notify.Email{
					TemplateID:   "template-id",
					EmailAddress: "name@example.org",
					Personalisation: map[string]string{
						"shareCode":                   "123",
						"cpFullName":                  "Joanna Jones",
						"donorFirstNames":             "Jan",
						"donorFullName":               "Jan Smith",
						"lpaLegalTerm":                "property and affairs",
						"certificateProviderStartURL": fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
						"donorFirstNamesPossessive":   "Jan’s",
					},
				}).
				Return("", nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)
			err := sender.SendCertificateProvider(ctx, notify.TemplateId(99), TestAppData, identity, lpa)

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

	lpa := &Lpa{
		CertificateProvider: actor.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type: LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", lpa.Type.LegalTermTransKey()).
		Return("property and affairs")
	localizer.
		On("Possessive", "Jan").
		Return("Jan’s")

	TestAppData.Localizer = localizer

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.
				On("Put", ctx, actor.TypeCertificateProvider, tc.expectedTestCode, actor.ShareCodeData{
					LpaID:           "lpa-id",
					Identity:        true,
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
				}).
				Return(nil)
			shareCodeStore.
				On("Put", ctx, actor.TypeCertificateProvider, "123", actor.ShareCodeData{
					LpaID:           "lpa-id",
					Identity:        true,
					DonorFullname:   "Jan Smith",
					DonorFirstNames: "Jan",
				}).
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.
				On("TemplateID", notify.TemplateId(99)).
				Return("template-id")
			notifyClient.
				On("Email", ctx, notify.Email{
					TemplateID:   "template-id",
					EmailAddress: "name@example.org",
					Personalisation: map[string]string{
						"shareCode":                   tc.expectedTestCode,
						"cpFullName":                  "Joanna Jones",
						"donorFirstNames":             "Jan",
						"donorFullName":               "Jan Smith",
						"lpaLegalTerm":                "property and affairs",
						"certificateProviderStartURL": fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
						"donorFirstNamesPossessive":   "Jan’s",
					},
				}).
				Return("", nil)
			notifyClient.
				On("Email", ctx, notify.Email{
					TemplateID:   "template-id",
					EmailAddress: "name@example.org",
					Personalisation: map[string]string{
						"shareCode":                   "123",
						"cpFullName":                  "Joanna Jones",
						"donorFirstNames":             "Jan",
						"donorFullName":               "Jan Smith",
						"lpaLegalTerm":                "property and affairs",
						"certificateProviderStartURL": fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
						"donorFirstNamesPossessive":   "Jan’s",
					},
				}).
				Return("", nil)

			sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)

			if tc.useTestCode {
				sender.UseTestCode()
			}

			err := sender.SendCertificateProvider(ctx, notify.TemplateId(99), TestAppData, true, lpa)

			assert.Nil(t, err)

			err = sender.SendCertificateProvider(ctx, notify.TemplateId(99), TestAppData, true, lpa)

			assert.Nil(t, err)
		})
	}
}

func TestShareCodeSenderSendCertificateProviderWhenEmailErrors(t *testing.T) {
	ctx := context.Background()

	lpa := &Lpa{
		CertificateProvider: actor.CertificateProvider{
			FirstNames: "Joanna",
			LastName:   "Jones",
			Email:      "name@example.org",
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type: LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", lpa.Type.LegalTermTransKey()).
		Return("property and affairs")
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
		On("TemplateID", notify.TemplateId(99)).
		Return("template-id")
	notifyClient.
		On("Email", ctx, notify.Email{
			TemplateID:   "template-id",
			EmailAddress: "name@example.org",
			Personalisation: map[string]string{
				"shareCode":                   "123",
				"cpFullName":                  "Joanna Jones",
				"donorFirstNames":             "Jan",
				"donorFullName":               "Jan Smith",
				"lpaLegalTerm":                "property and affairs",
				"certificateProviderStartURL": fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
				"donorFirstNamesPossessive":   "Jan’s",
			},
		}).
		Return("", ExpectedError)

	sender := NewShareCodeSender(shareCodeStore, notifyClient, "http://app", MockRandom)
	err := sender.SendCertificateProvider(ctx, notify.TemplateId(99), TestAppData, true, lpa)

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(ExpectedError)

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", MockRandom)
	err := sender.SendCertificateProvider(ctx, notify.TemplateId(99), TestAppData, true, &Lpa{})

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendAttorneys(t *testing.T) {
	lpa := &Lpa{
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
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
		}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
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
		}},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type: LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", lpa.Type.LegalTermTransKey()).
		Return("property and affairs")
	localizer.
		On("Possessive", "Jan").
		Return("Jan's")

	TestAppData.Localizer = localizer

	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
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
		On("TemplateID", notify.TemplateId(notify.AttorneyInviteEmail)).
		Return("template-id")
	notifyClient.
		On("TemplateID", notify.TemplateId(notify.ReplacementAttorneyInviteEmail)).
		Return("template-id2")
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
	err := sender.SendAttorneys(ctx, TestAppData, lpa)

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

	lpa := &Lpa{
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
		Type: LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", lpa.Type.LegalTermTransKey()).
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
				On("TemplateID", notify.TemplateId(notify.AttorneyInviteEmail)).
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

			err := sender.SendAttorneys(ctx, TestAppData, lpa)
			assert.Nil(t, err)

			err = sender.SendAttorneys(ctx, TestAppData, lpa)
			assert.Nil(t, err)
		})
	}
}

func TestShareCodeSenderSendAttorneysWhenEmailErrors(t *testing.T) {
	ctx := context.Background()

	lpa := &Lpa{
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
		Type: LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", lpa.Type.LegalTermTransKey()).
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
	err := sender.SendAttorneys(ctx, TestAppData, lpa)

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendAttorneysWhenShareCodeStoreErrors(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(ExpectedError)

	sender := NewShareCodeSender(shareCodeStore, nil, "http://app", MockRandom)
	err := sender.SendAttorneys(ctx, TestAppData, &Lpa{
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{Email: "hey@example.com"}}},
	})

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}
