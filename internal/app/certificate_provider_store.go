package app

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type certificateProviderStore struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

func (s *certificateProviderStore) Create(ctx context.Context, donorSessionID string, certificateProviderUID actoruid.UID, email string) (*actor.CertificateProviderProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("certificateProviderStore.Create requires LpaID and SessionID")
	}

	cp := &actor.CertificateProviderProvidedDetails{
		PK:        dynamo.LpaKey(data.LpaID),
		SK:        dynamo.CertificateProviderKey(data.SessionID),
		UID:       certificateProviderUID,
		LpaID:     data.LpaID,
		UpdatedAt: s.now(),
		Email:     email,
	}

	if err := s.dynamoClient.Create(ctx, cp); err != nil {
		return nil, err
	}
	if err := s.dynamoClient.Create(ctx, lpaLink{
		PK:        dynamo.LpaKey(data.LpaID),
		SK:        dynamo.SubKey(data.SessionID),
		DonorKey:  dynamo.LpaOwnerKey(dynamo.DonorKey(donorSessionID)),
		ActorType: actor.TypeCertificateProvider,
		UpdatedAt: s.now(),
	}); err != nil {
		return nil, err
	}

	return cp, err
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
	err = s.dynamoClient.OneByPartialSK(ctx, dynamo.LpaKey(data.LpaID), dynamo.CertificateProviderKey(""), &certificateProvider)

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
	err = s.dynamoClient.One(ctx, dynamo.LpaKey(data.LpaID), dynamo.CertificateProviderKey(data.SessionID), &certificateProvider)

	return &certificateProvider, err
}

func (s *certificateProviderStore) Put(ctx context.Context, certificateProvider *actor.CertificateProviderProvidedDetails) error {
	certificateProvider.UpdatedAt = s.now()
	return s.dynamoClient.Put(ctx, certificateProvider)
}
