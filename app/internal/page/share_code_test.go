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

func TestShareCodeSenderSend(t *testing.T) {
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
		On("T", lpa.TypeLegalTermTransKey()).
		Return("property and affairs")

	TestAppData.Localizer = localizer

	for name, identity := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			dataStore := newMockDataStore(t)
			dataStore.
				On("Put", ctx, "SHARECODE#123", "#METADATA#123", ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", Identity: identity}).
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
			err := sender.Send(ctx, notify.TemplateId(99), TestAppData, identity, lpa)

			assert.Nil(t, err)
		})
	}
}

func TestShareCodeSenderSendWithTestCode(t *testing.T) {
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
		On("T", lpa.TypeLegalTermTransKey()).
		Return("property and affairs")

	TestAppData.Localizer = localizer

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			dataStore := newMockDataStore(t)
			dataStore.
				On("Put", ctx, "SHARECODE#"+tc.expectedTestCode, "#METADATA#"+tc.expectedTestCode, ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", Identity: true}).
				Return(nil)
			dataStore.
				On("Put", ctx, "SHARECODE#123", "#METADATA#123", ShareCodeData{SessionID: "session-id", LpaID: "lpa-id", Identity: true}).
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

			err := sender.Send(ctx, notify.TemplateId(99), TestAppData, true, lpa)

			assert.Nil(t, err)

			err = sender.Send(ctx, notify.TemplateId(99), TestAppData, true, lpa)

			assert.Nil(t, err)
		})
	}
}

func TestShareCodeSenderSendWhenEmailErrors(t *testing.T) {
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
	err := sender.Send(ctx, notify.TemplateId(99), TestAppData, true, lpa)

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendWhenDataStoreErrors(t *testing.T) {
	ctx := context.Background()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(ExpectedError)

	sender := NewShareCodeSender(dataStore, nil, "http://app", MockRandom)
	err := sender.Send(ctx, notify.TemplateId(99), TestAppData, true, &Lpa{})

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}
