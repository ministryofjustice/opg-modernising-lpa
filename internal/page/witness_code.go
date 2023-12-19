package page

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
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

func (s *WitnessCodeSender) SendToCertificateProvider(ctx context.Context, donor *actor.DonorProvidedDetails, localizer Localizer) error {
	if !donor.CertificateProviderCodes.CanRequest(s.now()) {
		return ErrTooManyWitnessCodeRequests
	}

	code := s.randomCode(4)
	donor.CertificateProviderCodes = append(donor.CertificateProviderCodes, actor.WitnessCode{Code: code, Created: s.now()})

	_, err := s.notifyClient.SendSMS(ctx, donor.CertificateProvider.Mobile, notify.WitnessCodeSMS{
		WitnessCode:   code,
		DonorFullName: localizer.Possessive(donor.Donor.FullName()),
		LpaType:       localize.LowerFirst(localizer.T(donor.Type.String())),
	})
	if err != nil {
		return err
	}

	return s.donorStore.Put(ctx, donor)
}

func (s *WitnessCodeSender) SendToIndependentWitness(ctx context.Context, donor *actor.DonorProvidedDetails, localizer Localizer) error {
	if !donor.IndependentWitnessCodes.CanRequest(s.now()) {
		return ErrTooManyWitnessCodeRequests
	}

	code := s.randomCode(4)
	donor.IndependentWitnessCodes = append(donor.IndependentWitnessCodes, actor.WitnessCode{Code: code, Created: s.now()})

	_, err := s.notifyClient.SendSMS(ctx, donor.IndependentWitness.Mobile, notify.WitnessCodeSMS{
		WitnessCode:   code,
		DonorFullName: localizer.Possessive(donor.Donor.FullName()),
		LpaType:       localize.LowerFirst(localizer.T(donor.Type.String())),
	})
	if err != nil {
		return err
	}

	return s.donorStore.Put(ctx, donor)
}
