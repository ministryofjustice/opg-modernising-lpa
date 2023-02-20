package page

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
)

type ShareCodeData struct {
	SessionID string
	LpaID     string
	Identity  bool
}

type shareCodeSender struct {
	dataStore    DataStore
	notifyClient NotifyClient
	appPublicURL string
	randomString func(int) string
}

func NewShareCodeSender(dataStore DataStore, notifyClient NotifyClient, appPublicURL string, randomString func(int) string) *shareCodeSender {
	return &shareCodeSender{
		dataStore:    dataStore,
		notifyClient: notifyClient,
		appPublicURL: appPublicURL,
		randomString: randomString,
	}
}

func (s *shareCodeSender) Send(ctx context.Context, template notify.TemplateId, appData AppData, email string, identity bool) error {
	shareCode := s.randomString(12)

	if err := s.dataStore.Put(ctx, "SHARECODE#"+shareCode, "#METADATA#"+shareCode, ShareCodeData{
		SessionID: appData.SessionID,
		LpaID:     appData.LpaID,
		Identity:  identity,
	}); err != nil {
		return fmt.Errorf("creating sharecode failed: %w", err)
	}

	if _, err := s.notifyClient.Email(ctx, notify.Email{
		TemplateID:   s.notifyClient.TemplateID(template),
		EmailAddress: email,
		Personalisation: map[string]string{
			"link": fmt.Sprintf("%s%s?share-code=%s", s.appPublicURL, Paths.CertificateProviderStart, shareCode),
		},
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}
