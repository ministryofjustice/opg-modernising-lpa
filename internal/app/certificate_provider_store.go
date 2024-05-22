package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type certificateProviderStore struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

func (s *certificateProviderStore) Create(ctx context.Context, shareCode actor.ShareCodeData, email string) (*actor.CertificateProviderProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("certificateProviderStore.Create requires LpaID and SessionID")
	}

	certificateProvider := &actor.CertificateProviderProvidedDetails{
		PK:        dynamo.LpaKey(data.LpaID),
		SK:        dynamo.CertificateProviderKey(data.SessionID),
		UID:       shareCode.ActorUID,
		LpaID:     data.LpaID,
		UpdatedAt: s.now(),
		Email:     email,
	}

	transaction := dynamo.NewTransaction().
		Put(certificateProvider).
		Put(lpaLink{
			PK:        dynamo.LpaKey(data.LpaID),
			SK:        dynamo.SubKey(data.SessionID),
			DonorKey:  shareCode.LpaOwnerKey,
			ActorType: actor.TypeCertificateProvider,
			UpdatedAt: s.now(),
		}).
		Delete(shareCode.PK, shareCode.SK)

	if err := s.dynamoClient.WriteTransaction(ctx, transaction); err != nil {
		return nil, err
	}

	return certificateProvider, err
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

func (s *certificateProviderStore) Delete(ctx context.Context) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return errors.New("certificateProviderStore.Delete requires LpaID and SessionID")
	}

	if err := s.dynamoClient.DeleteOne(ctx, dynamo.LpaKey(data.LpaID), dynamo.CertificateProviderKey(data.SessionID)); err != nil {
		return fmt.Errorf("error deleting certificate provider: %w", err)
	}

	return nil
}
