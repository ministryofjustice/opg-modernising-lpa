package app

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoUpdateClient interface {
	Update(ctx context.Context, pk, sk string, values map[string]types.AttributeValue, expression string) error
}

type uidStore struct {
	dynamoClient DynamoUpdateClient
	now          func() time.Time
}

func NewUidStore(dynamoClient DynamoUpdateClient, now func() time.Time) *uidStore {
	return &uidStore{dynamoClient: dynamoClient, now: now}
}

func (s *uidStore) Set(ctx context.Context, lpaID, sessionID, organisationID, uid string) error {
	values, err := attributevalue.MarshalMap(map[string]any{
		":uid": uid,
		":now": s.now(),
	})
	if err != nil {
		return err
	}

	sk := donorKey(sessionID)
	if organisationID != "" {
		sk = organisationKey(organisationID)
	}

	return s.dynamoClient.Update(ctx, lpaKey(lpaID), sk, values,
		"set LpaUID = :uid, UpdatedAt = :now")
}
