package app

import (
	"context"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type ShareCodeStoreDynamoClient interface {
	One(ctx context.Context, pk, sk string, v interface{}) error
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

	pk, sk, err := shareCodeKeys(actorType, shareCode)
	if err != nil {
		return data, err
	}

	err = s.dynamoClient.One(ctx, pk, sk, &data)
	return data, err
}

func (s *shareCodeStore) Put(ctx context.Context, actorType actor.Type, shareCode string, data actor.ShareCodeData) error {
	pk, sk, err := shareCodeKeys(actorType, shareCode)
	if err != nil {
		return err
	}

	data.PK = pk
	data.SK = sk

	return s.dynamoClient.Put(ctx, data)
}

func (s *shareCodeStore) PutDonor(ctx context.Context, shareCode string, data actor.ShareCodeData) error {
	data.PK = "DONORSHARE#" + shareCode
	data.SK = "DONORINVITE#" + data.SessionID + "#" + data.LpaID
	data.UpdatedAt = s.now()

	return s.dynamoClient.Put(ctx, data)
}

func (s *shareCodeStore) GetDonor(ctx context.Context) (actor.ShareCodeData, error) {
	var data actor.ShareCodeData

	sessionData, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return data, err
	}

	sk := "DONORINVITE#" + sessionData.OrganisationID + "#" + sessionData.LpaID

	err = s.dynamoClient.OneBySK(ctx, sk, &data)
	return data, err
}

func (s *shareCodeStore) Delete(ctx context.Context, shareCode actor.ShareCodeData) error {
	return s.dynamoClient.DeleteOne(ctx, shareCode.PK, shareCode.SK)
}

func shareCodeKeys(actorType actor.Type, shareCode string) (pk, sk string, err error) {
	switch actorType {
	// As attorneys and replacement attorneys share the same landing page we can't
	// differentiate between them
	case actor.TypeAttorney, actor.TypeReplacementAttorney, actor.TypeTrustCorporation, actor.TypeReplacementTrustCorporation:
		return "ATTORNEYSHARE#" + shareCode, "#METADATA#" + shareCode, nil
	case actor.TypeCertificateProvider:
		return "CERTIFICATEPROVIDERSHARE#" + shareCode, "#METADATA#" + shareCode, nil
	default:
		return "", "", fmt.Errorf("cannot have share code for actorType=%v", actorType)
	}
}
