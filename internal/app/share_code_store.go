package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type ShareCodeStoreDynamoClient interface {
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error
	OneByPK(ctx context.Context, pk dynamo.PK, v interface{}) error
	OneBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	Put(ctx context.Context, v interface{}) error
	DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error
}

type shareCodeStore struct {
	dynamoClient ShareCodeStoreDynamoClient
	now          func() time.Time
}

func NewShareCodeStore(dynamoClient ShareCodeStoreDynamoClient) *shareCodeStore {
	return &shareCodeStore{dynamoClient: dynamoClient, now: time.Now}
}

func (s *shareCodeStore) Get(ctx context.Context, actorType actor.Type, shareCode string) (actor.ShareCodeData, error) {
	var data actor.ShareCodeData

	pk, err := shareCodeKey(actorType, shareCode)
	if err != nil {
		return data, err
	}

	if err := s.dynamoClient.OneByPK(ctx, pk, &data); err != nil {
		return actor.ShareCodeData{}, err
	}

	if !data.LpaLinkedAt.IsZero() {
		return actor.ShareCodeData{}, dynamo.NotFoundError{}
	}

	return data, err
}

func (s *shareCodeStore) Put(ctx context.Context, actorType actor.Type, shareCode string, data actor.ShareCodeData) error {
	pk, err := shareCodeKey(actorType, shareCode)
	if err != nil {
		return err
	}

	data.PK = pk
	data.SK = dynamo.ShareSortKey(dynamo.MetadataKey(shareCode))

	return s.dynamoClient.Put(ctx, data)
}

func (s *shareCodeStore) PutDonor(ctx context.Context, shareCode string, data actor.ShareCodeData) error {
	organisationKey, ok := data.LpaOwnerKey.Organisation()
	if !ok {
		return errors.New("shareCodeStore.PutDonor can only be used by organisations")
	}

	data.PK = dynamo.ShareKey(dynamo.DonorShareKey(shareCode))
	data.SK = dynamo.ShareSortKey(dynamo.DonorInviteKey(organisationKey, data.LpaKey))
	data.UpdatedAt = s.now()

	return s.dynamoClient.Put(ctx, data)
}

func (s *shareCodeStore) GetDonor(ctx context.Context) (actor.ShareCodeData, error) {
	var data actor.ShareCodeData

	sessionData, err := appcontext.SessionDataFromContext(ctx)
	if err != nil {
		return data, err
	}

	sk := dynamo.DonorInviteKey(dynamo.OrganisationKey(sessionData.OrganisationID), dynamo.LpaKey(sessionData.LpaID))

	err = s.dynamoClient.OneBySK(ctx, sk, &data)
	return data, err
}

func (s *shareCodeStore) Delete(ctx context.Context, shareCode actor.ShareCodeData) error {
	return s.dynamoClient.DeleteOne(ctx, shareCode.PK, shareCode.SK)
}

func shareCodeKey(actorType actor.Type, shareCode string) (pk dynamo.ShareKeyType, err error) {
	switch actorType {
	case actor.TypeDonor:
		return dynamo.ShareKey(dynamo.DonorShareKey(shareCode)), nil
	// As attorneys and replacement attorneys share the same landing page we can't
	// differentiate between them
	case actor.TypeAttorney, actor.TypeReplacementAttorney, actor.TypeTrustCorporation, actor.TypeReplacementTrustCorporation:
		return dynamo.ShareKey(dynamo.AttorneyShareKey(shareCode)), nil
	case actor.TypeCertificateProvider:
		return dynamo.ShareKey(dynamo.CertificateProviderShareKey(shareCode)), nil
	default:
		return dynamo.ShareKey(nil), fmt.Errorf("cannot have share code for actorType=%v", actorType)
	}
}
