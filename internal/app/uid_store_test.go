package app

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	values, _ := attributevalue.MarshalMap(map[string]any{
		":uid": "uid",
		":now": now,
	})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Update", ctx, "LPA#lpa-id", "#DONOR#session-id", values,
			"set LpaUID = :uid, UpdatedAt = :now").
		Return(nil)

	uidStore := NewUidStore(dynamoClient, func() time.Time { return now })

	assert.Nil(t, uidStore.Set(ctx, "lpa-id", "session-id", "uid"))
}

func TestSetWhenDynamoClientError(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	values, _ := attributevalue.MarshalMap(map[string]any{
		":uid": "uid",
		":now": now,
	})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Update", ctx, "LPA#lpa-id", "#DONOR#session-id", values,
			"set LpaUID = :uid, UpdatedAt = :now").
		Return(expectedError)

	uidStore := NewUidStore(dynamoClient, func() time.Time { return now })

	assert.Equal(t, expectedError, uidStore.Set(ctx, "lpa-id", "session-id", "uid"))
}
