package page

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
)

var useTestCode = false

type ShareCodeData struct {
	SessionID string
	LpaID     string
	Identity  bool
}

type ShareCodeSender struct {
	dataStore    DataStore
	notifyClient NotifyClient
	appPublicURL string
	randomString func(int) string
}

func NewShareCodeSender(dataStore DataStore, notifyClient NotifyClient, appPublicURL string, randomString func(int) string) *ShareCodeSender {
	return &ShareCodeSender{
		dataStore:    dataStore,
		notifyClient: notifyClient,
		appPublicURL: appPublicURL,
		randomString: randomString,
	}
}

func (s *ShareCodeSender) UseTestCode() {
	useTestCode = true
}

func (s *ShareCodeSender) Send(ctx context.Context, template notify.TemplateId, appData AppData, email string, identity bool, lpa *Lpa) error {
	var shareCode string

	if useTestCode {
		shareCode = "abcdef123456"
		useTestCode = false
	} else {
		shareCode = s.randomString(12)
	}

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
			"shareCode":         shareCode,
			"cpFullName":        lpa.CertificateProvider.FullName(),
			"donorFirstNames":   lpa.You.FirstNames,
			"donorFullName":     lpa.You.FullName(),
			"lpaLegalTerm":      appData.Localizer.T(lpa.TypeLegalTermTransKey()),
			"cpLandingPageLink": fmt.Sprintf("%s%s", s.appPublicURL, Paths.CertificateProviderStart),
			"optOutLink":        fmt.Sprintf("%s%s?share-code=%s", s.appPublicURL, Paths.CertificateProviderOptOut, shareCode),
		},
	}); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}
