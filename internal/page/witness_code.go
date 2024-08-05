package page

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
)

var (
	testWitnessCode               = "1234"
	UseTestWitnessCode            = false
	ErrTooManyWitnessCodeRequests = errors.New("too many witness code requests")
)

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

func (s *WitnessCodeSender) SendToCertificateProvider(ctx context.Context, donor *donordata.Provided, localizer Localizer) error {
	if !donor.CertificateProviderCodes.CanRequest(s.now()) {
		return ErrTooManyWitnessCodeRequests
	}

	code := s.randomCode(4)
	if UseTestWitnessCode {
		code = testWitnessCode
	}

	donor.CertificateProviderCodes = append(donor.CertificateProviderCodes, donordata.WitnessCode{Code: code, Created: s.now()})

	if err := s.donorStore.Put(ctx, donor); err != nil {
		return err
	}

	return s.notifyClient.SendActorSMS(ctx, donor.CertificateProvider.Mobile, donor.LpaUID, notify.WitnessCodeSMS{
		WitnessCode:   code,
		DonorFullName: localizer.Possessive(donor.Donor.FullName()),
		LpaType:       localize.LowerFirst(localizer.T(donor.Type.String())),
	})
}

func (s *WitnessCodeSender) SendToIndependentWitness(ctx context.Context, donor *donordata.Provided, localizer Localizer) error {
	if !donor.IndependentWitnessCodes.CanRequest(s.now()) {
		return ErrTooManyWitnessCodeRequests
	}

	code := s.randomCode(4)
	if UseTestWitnessCode {
		code = testWitnessCode
	}

	donor.IndependentWitnessCodes = append(donor.IndependentWitnessCodes, donordata.WitnessCode{Code: code, Created: s.now()})

	if err := s.donorStore.Put(ctx, donor); err != nil {
		return err
	}

	return s.notifyClient.SendActorSMS(ctx, donor.IndependentWitness.Mobile, donor.LpaUID, notify.WitnessCodeSMS{
		WitnessCode:   code,
		DonorFullName: localizer.Possessive(donor.Donor.FullName()),
		LpaType:       localize.LowerFirst(localizer.T(donor.Type.String())),
	})
}
