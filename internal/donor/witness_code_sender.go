package donor

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
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

type DonorStore interface {
	Create(context.Context) (*donordata.Provided, error)
	Put(context.Context, *donordata.Provided) error
}

type NotifyClient interface {
	SendActorEmail(context context.Context, to, lpaUID string, email notify.Email) error
	SendActorSMS(context context.Context, to, lpaUID string, sms notify.SMS) error
}

type Localizer interface {
	Concat(list []string, joiner string) string
	Count(messageID string, count int) string
	Format(messageID string, data map[string]interface{}) string
	FormatCount(messageID string, count int, data map[string]interface{}) string
	FormatDate(t date.TimeOrDate) string
	FormatTime(t time.Time) string
	FormatDateTime(t time.Time) string
	Possessive(s string) string
	SetShowTranslationKeys(s bool)
	ShowTranslationKeys() bool
	T(messageID string) string
}

type WitnessCodeSender struct {
	donorStore   DonorStore
	notifyClient NotifyClient
	localizer    Localizer
	randomCode   func(int) string
	now          func() time.Time
}

func NewWitnessCodeSender(donorStore DonorStore, notifyClient NotifyClient, localizer Localizer) *WitnessCodeSender {
	return &WitnessCodeSender{
		donorStore:   donorStore,
		notifyClient: notifyClient,
		localizer:    localizer,
		randomCode:   random.Code,
		now:          time.Now,
	}
}

func (s *WitnessCodeSender) SendToCertificateProvider(ctx context.Context, donor *donordata.Provided) error {
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
		DonorFullName: s.localizer.Possessive(donor.Donor.FullName()),
		LpaType:       localize.LowerFirst(s.localizer.T(donor.Type.String())),
	})
}

func (s *WitnessCodeSender) SendToIndependentWitness(ctx context.Context, donor *donordata.Provided) error {
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
		DonorFullName: s.localizer.Possessive(donor.Donor.FullName()),
		LpaType:       localize.LowerFirst(s.localizer.T(donor.Type.String())),
	})
}
