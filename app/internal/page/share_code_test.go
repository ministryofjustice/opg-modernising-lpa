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
		},
		You: actor.Person{
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
					EmailAddress: "name@example.com",
					Personalisation: map[string]string{
						"shareCode":         "123",
						"cpFullName":        "Joanna Jones",
						"donorFirstNames":   "Jan",
						"donorFullName":     "Jan Smith",
						"lpaLegalTerm":      "property and affairs",
						"cpLandingPageLink": fmt.Sprintf("http://app%s?share-code=%s", Paths.CertificateProviderStart, "123"),
						"optOutLink":        fmt.Sprintf("http://app%s?share-code=%s", Paths.CertificateProviderOptOut, "123"),
					},
				}).
				Return("", nil)

			sender := NewShareCodeSender(dataStore, notifyClient, "http://app", MockRandom)
			err := sender.Send(ctx, notify.TemplateId(99), TestAppData, "name@example.com", identity, lpa)

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
		},
		You: actor.Person{
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
			EmailAddress: "name@example.com",
			Personalisation: map[string]string{
				"shareCode":         "123",
				"cpFullName":        "Joanna Jones",
				"donorFirstNames":   "Jan",
				"donorFullName":     "Jan Smith",
				"lpaLegalTerm":      "property and affairs",
				"cpLandingPageLink": fmt.Sprintf("http://app%s?share-code=%s", Paths.CertificateProviderStart, "123"),
				"optOutLink":        fmt.Sprintf("http://app%s?share-code=%s", Paths.CertificateProviderOptOut, "123"),
			},
		}).
		Return("", ExpectedError)

	sender := NewShareCodeSender(dataStore, notifyClient, "http://app", MockRandom)
	err := sender.Send(ctx, notify.TemplateId(99), TestAppData, "name@example.com", true, lpa)

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}

func TestShareCodeSenderSendWhenDataStoreErrors(t *testing.T) {
	ctx := context.Background()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(ExpectedError)

	sender := NewShareCodeSender(dataStore, nil, "http://app", MockRandom)
	err := sender.Send(ctx, notify.TemplateId(99), TestAppData, "name@example.com", true, &Lpa{})

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
}
