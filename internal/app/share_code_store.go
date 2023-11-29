package app

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
)

type ShareCodeStoreDynamoClient interface {
	One(ctx context.Context, pk, sk string, v interface{}) error
	Put(ctx context.Context, v interface{}) error
}

type shareCodeStore struct {
	dynamoClient ShareCodeStoreDynamoClient
}

func NewShareCodeStore(dynamoClient ShareCodeStoreDynamoClient) *shareCodeStore {
	return &shareCodeStore{dynamoClient: dynamoClient}
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

func shareCodeKeys(actorType actor.Type, shareCode string) (pk, sk string, err error) {
	switch actorType {
	// As attorneys and replacement attorneys share the same landing page we can't
	// differentiate between them
	case actor.TypeAttorney, actor.TypeReplacementAttorney:
		return "ATTORNEYSHARE#" + shareCode, "#METADATA#" + shareCode, nil
	case actor.TypeCertificateProvider:
		return "CERTIFICATEPROVIDERSHARE#" + shareCode, "#METADATA#" + shareCode, nil
	default:
		return "", "", fmt.Errorf("cannot have share code for actorType=%v", actorType)
	}
}
