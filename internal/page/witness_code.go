package page

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
)

const (
	witnessCodeExpireAfter  = 15 * time.Minute
	witnessCodeIgnoreAfter  = 2 * time.Hour
	witnessCodeRequestAfter = time.Minute
)

var ErrTooManyWitnessCodeRequests = errors.New("too many witness code requests")

type WitnessCodeSender struct {
	donorStore   DonorStore
	notifyClient NotifyClient
	randomCode   func(int) string
	now          func() time.Time
}

func NewWitnessCodeSender(donorStore DonorStore, notifyClient NotifyClient) *WitnessCodeSender {
	return &WitnessCodeSender{
		donorStore:   donorStore,
		notifyClient: notifyClient,
		randomCode:   random.Code,
		now:          time.Now,
	}
}

func (s *WitnessCodeSender) SendToCertificateProvider(ctx context.Context, lpa *Lpa, localizer Localizer) error {
	if !lpa.CertificateProviderCodes.CanRequest(s.now()) {
		return ErrTooManyWitnessCodeRequests
	}

	code := s.randomCode(4)
	lpa.CertificateProviderCodes = append(lpa.CertificateProviderCodes, WitnessCode{Code: code, Created: s.now()})

	_, err := s.notifyClient.Sms(ctx, notify.Sms{
		PhoneNumber: lpa.CertificateProvider.Mobile,
		TemplateID:  s.notifyClient.TemplateID(notify.SignatureCodeSMS),
		Personalisation: map[string]string{
			"WitnessCode":   code,
			"DonorFullName": localizer.Possessive(lpa.Donor.FullName()),
			"LpaType":       localizer.T(lpa.Type.LegalTermTransKey()),
		},
	})
	if err != nil {
		return err
	}

	return s.donorStore.Put(ctx, lpa)
}

func (s *WitnessCodeSender) SendToIndependentWitness(ctx context.Context, lpa *Lpa, localizer Localizer) error {
	if !lpa.IndependentWitnessCodes.CanRequest(s.now()) {
		return ErrTooManyWitnessCodeRequests
	}

	code := s.randomCode(4)
	lpa.IndependentWitnessCodes = append(lpa.IndependentWitnessCodes, WitnessCode{Code: code, Created: s.now()})

	_, err := s.notifyClient.Sms(ctx, notify.Sms{
		PhoneNumber: lpa.IndependentWitness.Mobile,
		TemplateID:  s.notifyClient.TemplateID(notify.SignatureCodeSMS),
		Personalisation: map[string]string{
			"WitnessCode":   code,
			"DonorFullName": localizer.Possessive(lpa.Donor.FullName()),
			"LpaType":       localizer.T(lpa.Type.LegalTermTransKey()),
		},
	})
	if err != nil {
		return err
	}

	return s.donorStore.Put(ctx, lpa)
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
