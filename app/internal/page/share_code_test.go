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
	testcases := map[string]bool{
		"identity": true,
		"sign in":  false,
	}
	lpa := &Lpa{
		CertificateProviderDetails: CertificateProviderDetails{
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
		On("T", lpa.TypeLegalTermTransKey()).
		Return("property and affairs")

	TestAppData.Localizer = localizer

	for name, identity := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			dataStore := newMockDataStore(t)
			dataStore.
				On("Put", ctx, "CERTIFICATEPROVIDERSHARE#123", "#METADATA#123", ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", Identity: identity}).
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
						"shareCode":         "123",
						"cpFullName":        "Joanna Jones",
						"donorFirstNames":   "Jan",
						"donorFullName":     "Jan Smith",
						"lpaLegalTerm":      "property and affairs",
						"cpLandingPageLink": fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
						"optOutLink":        fmt.Sprintf("http://app%s?share-code=%s", Paths.CertificateProviderOptOut, "123"),
					},
				}).
				Return("", nil)

			sender := NewShareCodeSender(dataStore, notifyClient, "http://app", MockRandom)
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
		CertificateProviderDetails: CertificateProviderDetails{
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
		On("T", lpa.TypeLegalTermTransKey()).
		Return("property and affairs")

	TestAppData.Localizer = localizer

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			dataStore := newMockDataStore(t)
			dataStore.
				On("Put", ctx, "CERTIFICATEPROVIDERSHARE#"+tc.expectedTestCode, "#METADATA#"+tc.expectedTestCode, ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", Identity: true}).
				Return(nil)
			dataStore.
				On("Put", ctx, "CERTIFICATEPROVIDERSHARE#123", "#METADATA#123", ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", Identity: true}).
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
						"shareCode":         tc.expectedTestCode,
						"cpFullName":        "Joanna Jones",
						"donorFirstNames":   "Jan",
						"donorFullName":     "Jan Smith",
						"lpaLegalTerm":      "property and affairs",
						"cpLandingPageLink": fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
						"optOutLink":        fmt.Sprintf("http://app%s?share-code=%s", Paths.CertificateProviderOptOut, tc.expectedTestCode),
					},
				}).
				Return("", nil)
			notifyClient.
				On("Email", ctx, notify.Email{
					TemplateID:   "template-id",
					EmailAddress: "name@example.org",
					Personalisation: map[string]string{
						"shareCode":         "123",
						"cpFullName":        "Joanna Jones",
						"donorFirstNames":   "Jan",
						"donorFullName":     "Jan Smith",
						"lpaLegalTerm":      "property and affairs",
						"cpLandingPageLink": fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
						"optOutLink":        fmt.Sprintf("http://app%s?share-code=%s", Paths.CertificateProviderOptOut, "123"),
					},
				}).
				Return("", nil)

			sender := NewShareCodeSender(dataStore, notifyClient, "http://app", MockRandom)

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
		CertificateProviderDetails: CertificateProviderDetails{
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
		On("T", lpa.TypeLegalTermTransKey()).
		Return("property and affairs")

	TestAppData.Localizer = localizer

	dataStore := newMockDataStore(t)
	dataStore.
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
				"shareCode":         "123",
				"cpFullName":        "Joanna Jones",
				"donorFirstNames":   "Jan",
				"donorFullName":     "Jan Smith",
				"lpaLegalTerm":      "property and affairs",
				"cpLandingPageLink": fmt.Sprintf("http://app%s", Paths.CertificateProviderStart),
				"optOutLink":        fmt.Sprintf("http://app%s?share-code=%s", Paths.CertificateProviderOptOut, "123"),
			},
		}).
		Return("", ExpectedError)

	sender := NewShareCodeSender(dataStore, notifyClient, "http://app", MockRandom)
	err := sender.SendCertificateProvider(ctx, notify.TemplateId(99), TestAppData, true, lpa)

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendCertificateProviderWhenDataStoreErrors(t *testing.T) {
	ctx := context.Background()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(ExpectedError)

	sender := NewShareCodeSender(dataStore, nil, "http://app", MockRandom)
	err := sender.SendCertificateProvider(ctx, notify.TemplateId(99), TestAppData, true, &Lpa{})

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendAttorneys(t *testing.T) {
	lpa := &Lpa{
		Attorneys: actor.Attorneys{
			{
				FirstNames: "Joanna",
				LastName:   "Jones",
				Email:      "name@example.org",
			},
			{
				FirstNames: "John",
				LastName:   "Jones",
				Email:      "name2@example.org",
			},
			{
				FirstNames: "Nope",
				LastName:   "Jones",
			},
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type: LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", lpa.TypeLegalTermTransKey()).
		Return("property and affairs")

	TestAppData.Localizer = localizer

	ctx := context.Background()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", ctx, "ATTORNEYSHARE#123", "#METADATA#123", ShareCodeData{SessionID: "session-id", LpaID: "lpa-id"}).
		Return(nil)
	dataStore.
		On("Put", ctx, "ATTORNEYSHARE#123", "#METADATA#123", ShareCodeData{SessionID: "session-id", LpaID: "lpa-id"}).
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
				"shareCode":        "123",
				"attorneyFullName": "Joanna Jones",
				"donorFirstNames":  "Jan",
				"donorFullName":    "Jan Smith",
				"lpaLegalTerm":     "property and affairs",
				"landingPageLink":  fmt.Sprintf("http://app%s", Paths.Attorney.Start),
			},
		}).
		Return("", nil)
	notifyClient.
		On("Email", ctx, notify.Email{
			TemplateID:   "template-id",
			EmailAddress: "name2@example.org",
			Personalisation: map[string]string{
				"shareCode":        "123",
				"attorneyFullName": "John Jones",
				"donorFirstNames":  "Jan",
				"donorFullName":    "Jan Smith",
				"lpaLegalTerm":     "property and affairs",
				"landingPageLink":  fmt.Sprintf("http://app%s", Paths.Attorney.Start),
			},
		}).
		Return("", nil)

	sender := NewShareCodeSender(dataStore, notifyClient, "http://app", MockRandom)
	err := sender.SendAttorneys(ctx, notify.TemplateId(99), TestAppData, lpa)

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
		Attorneys: actor.Attorneys{
			{
				FirstNames: "Joanna",
				LastName:   "Jones",
				Email:      "name@example.org",
			},
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type: LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", lpa.TypeLegalTermTransKey()).
		Return("property and affairs")

	TestAppData.Localizer = localizer

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			dataStore := newMockDataStore(t)
			dataStore.
				On("Put", ctx, "ATTORNEYSHARE#"+tc.expectedTestCode, "#METADATA#"+tc.expectedTestCode, ShareCodeData{SessionID: "session-id", LpaID: "lpa-id"}).
				Return(nil)
			dataStore.
				On("Put", ctx, "ATTORNEYSHARE#123", "#METADATA#123", ShareCodeData{SessionID: "session-id", LpaID: "lpa-id"}).
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
						"shareCode":        tc.expectedTestCode,
						"attorneyFullName": "Joanna Jones",
						"donorFirstNames":  "Jan",
						"donorFullName":    "Jan Smith",
						"lpaLegalTerm":     "property and affairs",
						"landingPageLink":  fmt.Sprintf("http://app%s", Paths.Attorney.Start),
					},
				}).
				Return("", nil)
			notifyClient.
				On("Email", ctx, notify.Email{
					TemplateID:   "template-id",
					EmailAddress: "name@example.org",
					Personalisation: map[string]string{
						"shareCode":        "123",
						"attorneyFullName": "Joanna Jones",
						"donorFirstNames":  "Jan",
						"donorFullName":    "Jan Smith",
						"lpaLegalTerm":     "property and affairs",
						"landingPageLink":  fmt.Sprintf("http://app%s", Paths.Attorney.Start),
					},
				}).
				Return("", nil)

			sender := NewShareCodeSender(dataStore, notifyClient, "http://app", MockRandom)

			if tc.useTestCode {
				sender.UseTestCode()
			}

			err := sender.SendAttorneys(ctx, notify.TemplateId(99), TestAppData, lpa)
			assert.Nil(t, err)

			err = sender.SendAttorneys(ctx, notify.TemplateId(99), TestAppData, lpa)
			assert.Nil(t, err)
		})
	}
}

func TestShareCodeSenderSendAttorneysWhenEmailErrors(t *testing.T) {
	ctx := context.Background()

	lpa := &Lpa{
		Attorneys: actor.Attorneys{
			{
				FirstNames: "Joanna",
				LastName:   "Jones",
				Email:      "name@example.org",
			},
		},
		Donor: actor.Donor{
			FirstNames: "Jan",
			LastName:   "Smith",
		},
		Type: LpaTypePropertyFinance,
	}

	localizer := newMockLocalizer(t)
	localizer.
		On("T", lpa.TypeLegalTermTransKey()).
		Return("property and affairs")

	TestAppData.Localizer = localizer

	dataStore := newMockDataStore(t)
	dataStore.
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
				"shareCode":        "123",
				"attorneyFullName": "Joanna Jones",
				"donorFirstNames":  "Jan",
				"donorFullName":    "Jan Smith",
				"lpaLegalTerm":     "property and affairs",
				"landingPageLink":  fmt.Sprintf("http://app%s", Paths.Attorney.Start),
			},
		}).
		Return("", ExpectedError)

	sender := NewShareCodeSender(dataStore, notifyClient, "http://app", MockRandom)
	err := sender.SendAttorneys(ctx, notify.TemplateId(99), TestAppData, lpa)

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendAttorneysWhenDataStoreErrors(t *testing.T) {
	ctx := context.Background()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(ExpectedError)

	sender := NewShareCodeSender(dataStore, nil, "http://app", MockRandom)
	err := sender.SendAttorneys(ctx, notify.TemplateId(99), TestAppData, &Lpa{
		Attorneys: actor.Attorneys{{Email: "hey@example.com"}},
	})

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}
