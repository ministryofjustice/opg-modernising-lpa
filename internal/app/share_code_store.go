package app

import (
	"context"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type ShareCodeStoreDynamoClient interface {
	One(ctx context.Context, pk, sk string, v interface{}) error
	OneByPK(ctx context.Context, pk string, v interface{}) error
	OneBySK(ctx context.Context, sk string, v interface{}) error
	Put(ctx context.Context, v interface{}) error
	DeleteOne(ctx context.Context, pk, sk string) error
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

	pk, err := dynamo.ShareCodeKey(actorType, shareCode)
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

func (s *shareCodeStore) Linked(ctx context.Context, data actor.ShareCodeData, email string) error {
	data.LpaLinkedTo = email
	data.LpaLinkedAt = s.now()

	return s.dynamoClient.Put(ctx, data)
}

func (s *shareCodeStore) Put(ctx context.Context, actorType actor.Type, shareCode string, data actor.ShareCodeData) error {
	pk, err := dynamo.ShareCodeKey(actorType, shareCode)
	if err != nil {
		return err
	}

	data.PK = pk
	data.SK = dynamo.MetadataKey(shareCode)

	return s.dynamoClient.Put(ctx, data)
}

func (s *shareCodeStore) PutDonor(ctx context.Context, shareCode string, data actor.ShareCodeData) error {
	data.PK = dynamo.DonorShareKey(shareCode)
	data.SK = dynamo.DonorInviteKey(data.SessionID, data.LpaID)
	data.UpdatedAt = s.now()

	return s.dynamoClient.Put(ctx, data)
}

func (s *shareCodeStore) GetDonor(ctx context.Context) (actor.ShareCodeData, error) {
	var data actor.ShareCodeData

	sessionData, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return data, err
	}

	sk := dynamo.DonorInviteKey(sessionData.OrganisationID, sessionData.LpaID)

	err = s.dynamoClient.OneBySK(ctx, sk, &data)
	return data, err
}

func (s *shareCodeStore) Delete(ctx context.Context, shareCode actor.ShareCodeData) error {
	return s.dynamoClient.DeleteOne(ctx, shareCode.PK, shareCode.SK)
}
