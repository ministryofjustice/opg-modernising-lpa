package app

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type CertificateProviderStore struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

func NewCertificateProviderStore(client DynamoClient, now func() time.Time) CertificateProviderStore {
	return CertificateProviderStore{
		dynamoClient: client,
		now:          now,
	}
}

func (s *CertificateProviderStore) Create(ctx context.Context, donorSessionID string, certificateProviderUID actoruid.UID) (*actor.CertificateProviderProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("CertificateProviderStore.Create requires LpaID and SessionID")
	}

	cp := &actor.CertificateProviderProvidedDetails{
		PK:        lpaKey(data.LpaID),
		SK:        certificateProviderKey(data.SessionID),
		UID:       certificateProviderUID,
		LpaID:     data.LpaID,
		UpdatedAt: s.now(),
	}

	if err := s.dynamoClient.Create(ctx, cp); err != nil {
		return nil, err
	}
	if err := s.dynamoClient.Create(ctx, lpaLink{
		PK:        lpaKey(data.LpaID),
		SK:        subKey(data.SessionID),
		DonorKey:  donorKey(donorSessionID),
		ActorType: actor.TypeCertificateProvider,
		UpdatedAt: s.now(),
	}); err != nil {
		return nil, err
	}

	return cp, err
}

func CreatePaper(ctx context.Context, lpaID string, certificateProviderUID actoruid.UID) error {
	return nil
}

func (s *CertificateProviderStore) GetAny(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" {
		return nil, errors.New("CertificateProviderStore.GetAny requires LpaID")
	}

	var certificateProvider actor.CertificateProviderProvidedDetails
	err = s.dynamoClient.OneByPartialSK(ctx, lpaKey(data.LpaID), "#CERTIFICATE_PROVIDER#", &certificateProvider)

	return &certificateProvider, err
}

func (s *CertificateProviderStore) Get(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("CertificateProviderStore.Get requires LpaID and SessionID")
	}

	var certificateProvider actor.CertificateProviderProvidedDetails
	err = s.dynamoClient.One(ctx, lpaKey(data.LpaID), certificateProviderKey(data.SessionID), &certificateProvider)

	return &certificateProvider, err
}

func (s *CertificateProviderStore) Put(ctx context.Context, certificateProvider *actor.CertificateProviderProvidedDetails) error {
	certificateProvider.UpdatedAt = s.now()
	return s.dynamoClient.Put(ctx, certificateProvider)
}

func certificateProviderKey(sessionID string) string {
	return "#CERTIFICATE_PROVIDER#" + sessionID
}
