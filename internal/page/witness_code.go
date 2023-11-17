package page

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
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

func (s *WitnessCodeSender) SendToCertificateProvider(ctx context.Context, lpa *actor.DonorProvidedDetails, localizer Localizer) error {
	if !lpa.CertificateProviderCodes.CanRequest(s.now()) {
		return ErrTooManyWitnessCodeRequests
	}

	code := s.randomCode(4)
	lpa.CertificateProviderCodes = append(lpa.CertificateProviderCodes, actor.WitnessCode{Code: code, Created: s.now()})

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

func (s *WitnessCodeSender) SendToIndependentWitness(ctx context.Context, lpa *actor.DonorProvidedDetails, localizer Localizer) error {
	if !lpa.IndependentWitnessCodes.CanRequest(s.now()) {
		return ErrTooManyWitnessCodeRequests
	}

	code := s.randomCode(4)
	lpa.IndependentWitnessCodes = append(lpa.IndependentWitnessCodes, actor.WitnessCode{Code: code, Created: s.now()})

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
