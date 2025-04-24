package donor

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
)

var (
	testWitnessCode               = "1234"
	ErrTooManyWitnessCodeRequests = errors.New("too many witness code requests")
)

type DonorStore interface {
	Create(context.Context) (*donordata.Provided, error)
	Put(context.Context, *donordata.Provided) error
}

type CertificateProviderStore interface {
	GetAny(context.Context) (*certificateproviderdata.Provided, error)
}

type NotifyClient interface {
	SendActorSMS(context context.Context, to notify.ToMobile, lpaUID string, sms notify.SMS) error
}

type Localizer interface {
	localize.Localizer
}

type WitnessCodeSender struct {
	donorStore               DonorStore
	certificateProviderStore CertificateProviderStore
	notifyClient             NotifyClient
	localizer                Localizer
	randomCode               func(int) string
	now                      func() time.Time
	useTestCode              bool
}

func NewWitnessCodeSender(donorStore DonorStore, certificateProviderStore CertificateProviderStore, notifyClient NotifyClient, localizer Localizer, useTestCode bool) *WitnessCodeSender {
	return &WitnessCodeSender{
		donorStore:               donorStore,
		certificateProviderStore: certificateProviderStore,
		notifyClient:             notifyClient,
		localizer:                localizer,
		randomCode:               random.Code,
		now:                      time.Now,
		useTestCode:              useTestCode,
	}
}

func (s *WitnessCodeSender) SendToCertificateProvider(ctx context.Context, donor *donordata.Provided) error {
	if !donor.CertificateProviderCodes.CanRequest(s.now()) {
		return ErrTooManyWitnessCodeRequests
	}

	code := s.randomCode(4)
	if s.useTestCode {
		code = testWitnessCode
	}

	donor.CertificateProviderCodes = append(donor.CertificateProviderCodes, donordata.WitnessCode{Code: code, Created: s.now()})

	if err := s.donorStore.Put(ctx, donor); err != nil {
		return err
	}

	to := notify.ToCertificateProvider(donor.CertificateProvider)
	if certificateProvider, _ := s.certificateProviderStore.GetAny(ctx); certificateProvider != nil {
		to = notify.ToProvidedCertificateProvider(certificateProvider, donor.CertificateProvider)
	}

	return s.notifyClient.SendActorSMS(ctx, to, donor.LpaUID, notify.WitnessCodeSMS{
		WitnessCode:   code,
		DonorFullName: s.localizer.Possessive(donor.Donor.FullName()),
		LpaType:       localize.LowerFirst(s.localizer.T(donor.Type.String())),
	})
}

func (s *WitnessCodeSender) SendToIndependentWitness(ctx context.Context, donor *donordata.Provided) error {
	if !donor.IndependentWitnessCodes.CanRequest(s.now()) {
		return ErrTooManyWitnessCodeRequests
	}

	code := s.randomCode(4)
	if s.useTestCode {
		code = testWitnessCode
	}

	donor.IndependentWitnessCodes = append(donor.IndependentWitnessCodes, donordata.WitnessCode{Code: code, Created: s.now()})

	if err := s.donorStore.Put(ctx, donor); err != nil {
		return err
	}

	return s.notifyClient.SendActorSMS(ctx, notify.ToIndependentWitness(donor.IndependentWitness), donor.LpaUID, notify.WitnessCodeSMS{
		WitnessCode:   code,
		DonorFullName: s.localizer.Possessive(donor.Donor.FullName()),
		LpaType:       localize.LowerFirst(s.localizer.T(donor.Type.String())),
	})
}
