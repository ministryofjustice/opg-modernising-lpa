package app

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
)

type certificateProviderStore struct {
	dataStore DynamoClient
	now       func() time.Time
}

func (s *certificateProviderStore) Create(ctx context.Context, donorSessionID string) (*actor.CertificateProviderProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("certificateProviderStore.Create requires LpaID and SessionID")
	}

	cp := &actor.CertificateProviderProvidedDetails{
		PK:        lpaKey(data.LpaID),
		SK:        certificateProviderKey(data.SessionID),
		LpaID:     data.LpaID,
		UpdatedAt: s.now(),
	}

	if err := s.dataStore.Create(ctx, cp); err != nil {
		return nil, err
	}
	if err := s.dataStore.Create(ctx, lpaLink{
		PK:        lpaKey(data.LpaID),
		SK:        subKey(data.SessionID),
		DonorKey:  donorKey(donorSessionID),
		ActorType: actor.TypeCertificateProvider,
	}); err != nil {
		return nil, err
	}

	return cp, err
}

func (s *certificateProviderStore) GetAll(ctx context.Context) ([]*actor.CertificateProviderProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("certificateProviderStore.GetAll requires SessionID")
	}

	var items []*actor.CertificateProviderProvidedDetails
	err = s.dataStore.GetAllByGsi(ctx, "ActorIndex", certificateProviderKey(data.SessionID), &items)

	return items, err
}

func (s *certificateProviderStore) GetAny(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" {
		return nil, errors.New("certificateProviderStore.GetAny requires LpaID")
	}

	var certificateProvider actor.CertificateProviderProvidedDetails
	err = s.dataStore.GetOneByPartialSk(ctx, lpaKey(data.LpaID), "#CERTIFICATE_PROVIDER#", &certificateProvider)

	return &certificateProvider, err
}

func (s *certificateProviderStore) Get(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("certificateProviderStore.Get requires LpaID and SessionID")
	}

	var certificateProvider actor.CertificateProviderProvidedDetails
	err = s.dataStore.Get(ctx, lpaKey(data.LpaID), certificateProviderKey(data.SessionID), &certificateProvider)

	return &certificateProvider, err
}

func (s *certificateProviderStore) Put(ctx context.Context, certificateProvider *actor.CertificateProviderProvidedDetails) error {
	certificateProvider.UpdatedAt = s.now()
	return s.dataStore.Put(ctx, certificateProvider)
}

func certificateProviderKey(s string) string {
	return "#CERTIFICATE_PROVIDER#" + s
}
