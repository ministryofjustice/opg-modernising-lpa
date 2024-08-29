package scheduled

import (
	"context"
	"testing"
	time "time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (c *mockDynamoClient_OneByPK_Call) SetData(row Event) {
	c.Run(func(_ context.Context, _ dynamo.PK, v any) {
		b, _ := attributevalue.Marshal(row)
		attributevalue.Unmarshal(b, v)
	})
}

func TestNewStore(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	store := NewStore(dynamoClient)
	assert.Equal(t, dynamoClient, store.dynamoClient)
}

func TestStorePop(t *testing.T) {
	row := Event{
		Action:            99,
		PK:                dynamo.ScheduledDayKey(testNow),
		SK:                dynamo.ScheduledKey(testNow, 99),
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByPK(ctx, dynamo.ScheduledDayKey(testNow), mock.Anything).
		Return(nil).
		SetData(row)
	dynamoClient.EXPECT().
		DeleteOne(ctx, row.PK, row.SK).
		Return(nil)

	store := &Store{dynamoClient: dynamoClient}
	result, err := store.Pop(ctx, testNow)
	assert.Nil(t, err)
	assert.Equal(t, row, result)
}

func TestStorePopWhenOneByPKErrors(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByPK(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient}
	_, err := store.Pop(ctx, testNow)
	assert.Equal(t, expectedError, err)
}

func TestStorePopWhenDeleteOneErrors(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByPK(mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		SetData(Event{
			Action:            99,
			PK:                dynamo.ScheduledDayKey(testNow),
			SK:                dynamo.ScheduledKey(testNow, 99),
			TargetLpaKey:      dynamo.LpaKey("an-lpa"),
			TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
		})
	dynamoClient.EXPECT().
		DeleteOne(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient}
	_, err := store.Pop(ctx, testNow)
	assert.Equal(t, expectedError, err)
}

func TestStorePut(t *testing.T) {
	at := time.Date(2024, time.January, 1, 12, 13, 14, 5, time.UTC)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, Event{
			PK:        dynamo.ScheduledDayKey(at),
			SK:        dynamo.ScheduledKey(at, 99),
			CreatedAt: testNow,
			At:        at,
			Action:    99,
		}).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn}
	err := store.Put(ctx, Event{At: at, Action: 99})
	assert.Equal(t, expectedError, err)
}
