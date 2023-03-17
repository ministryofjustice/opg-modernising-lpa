package page

import (
	"context"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
)

const (
	witnessCodeExpireAfter  = 15 * time.Minute
	witnessCodeIgnoreAfter  = 2 * time.Hour
	witnessCodeRequestAfter = time.Minute
)

type WitnessCodeSender struct {
	lpaStore     LpaStore
	notifyClient NotifyClient
	randomCode   func(int) string
	now          func() time.Time
}

func NewWitnessCodeSender(lpaStore LpaStore, notifyClient NotifyClient) *WitnessCodeSender {
	return &WitnessCodeSender{
		lpaStore:     lpaStore,
		notifyClient: notifyClient,
		randomCode:   random.Code,
		now:          time.Now,
	}
}

func (s *WitnessCodeSender) Send(ctx context.Context, lpa *Lpa, localizer Localizer) error {
	code := s.randomCode(4)
	lpa.WitnessCodes = append(lpa.WitnessCodes, WitnessCode{Code: code, Created: s.now()})

	_, err := s.notifyClient.Sms(ctx, notify.Sms{
		PhoneNumber: lpa.CertificateProvider.Mobile,
		TemplateID:  s.notifyClient.TemplateID(notify.SignatureCodeSms),
		Personalisation: map[string]string{
			"WitnessCode":   code,
			"DonorFullName": localizer.Possessive(lpa.Donor.FullName()),
			"LpaType":       localizer.T(lpa.TypeLegalTermTransKey()),
		},
	})
	if err != nil {
		return err
	}

	return s.lpaStore.Put(ctx, lpa)
}

type WitnessCode struct {
	Code    string
	Created time.Time
}

func (w WitnessCode) HasExpired() bool {
	return w.Created.Add(witnessCodeExpireAfter).Before(time.Now())
}

type WitnessCodes []WitnessCode

func (ws WitnessCodes) Find(code string) (WitnessCode, bool) {
	for _, w := range ws {
		if w.Code == code {
			if w.Created.Add(witnessCodeIgnoreAfter).Before(time.Now()) {
				break
			}

			return w, true
		}
	}

	return WitnessCode{}, false
}

func (ws WitnessCodes) CanRequest(now time.Time) bool {
	if len(ws) == 0 {
		return true
	}

	lastCode := ws[len(ws)-1]
	return lastCode.Created.Add(witnessCodeRequestAfter).Before(now)
}
