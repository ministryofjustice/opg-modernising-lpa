package scheduled

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewStore(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	store := NewStore(dynamoClient)
	assert.Equal(t, dynamoClient, store.dynamoClient)
}

func TestStorePop(t *testing.T) {
	row := &Event{
		Action:            99,
		PK:                dynamo.ScheduledDayKey(testNow),
		SK:                dynamo.ScheduledKey(testNow, 99),
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}
	movedRow := &Event{
		Action:            99,
		PK:                dynamo.ScheduledDayKey(testNow).Handled(),
		SK:                dynamo.ScheduledKey(testNow, 99),
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AnyByPK(ctx, dynamo.ScheduledDayKey(testNow), mock.Anything).
		Return(nil).
		SetData(row)
	dynamoClient.EXPECT().
		Move(ctx, dynamo.Keys{PK: row.PK, SK: row.SK}, *movedRow).
		Return(nil)

	store := &Store{dynamoClient: dynamoClient}
	result, err := store.Pop(ctx, testNow)
	assert.Nil(t, err)
	assert.Equal(t, movedRow, result)
}

func TestStorePopWhenAnyByPKErrors(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AnyByPK(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient}
	_, err := store.Pop(ctx, testNow)
	assert.Equal(t, expectedError, err)
}

func TestStorePopWhenDeleteOneErrors(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AnyByPK(mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		SetData(&Event{
			Action:            99,
			PK:                dynamo.ScheduledDayKey(testNow),
			SK:                dynamo.ScheduledKey(testNow, 99),
			TargetLpaKey:      dynamo.LpaKey("an-lpa"),
			TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
		})
	dynamoClient.EXPECT().
		Move(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient}
	_, err := store.Pop(ctx, testNow)
	assert.Equal(t, expectedError, err)
}

func TestStoreCreate(t *testing.T) {
	at := time.Date(2024, time.January, 1, 12, 13, 14, 5, time.UTC)
	at2 := time.Date(2024, time.February, 1, 12, 13, 14, 5, time.UTC)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, &dynamo.Transaction{
			Creates: []any{
				Event{
					PK:        dynamo.ScheduledDayKey(at),
					SK:        dynamo.ScheduledKey(at, 99),
					CreatedAt: testNow,
					At:        at,
					Action:    99,
				},
				Event{
					PK:        dynamo.ScheduledDayKey(at2),
					SK:        dynamo.ScheduledKey(at2, 100),
					CreatedAt: testNow,
					At:        at2,
					Action:    100,
				},
			},
		}).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn}
	err := store.Create(ctx, Event{At: at, Action: 99}, Event{At: at2, Action: 100})
	assert.Equal(t, expectedError, err)
}

func TestDeleteAllByUID(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByLpaUIDAndPartialSK(ctx, "lpa-uid", dynamo.PartialScheduledKey(), mock.Anything).
		Return(nil).
		SetData([]Event{
			{LpaUID: "lpa-uid", PK: dynamo.ScheduledDayKey(now), SK: dynamo.ScheduledKey(now, 98)},
			{LpaUID: "lpa-uid", PK: dynamo.ScheduledDayKey(yesterday), SK: dynamo.ScheduledKey(yesterday, 99)},
		})
	dynamoClient.EXPECT().
		DeleteKeys(ctx, []dynamo.Keys{
			{PK: dynamo.ScheduledDayKey(now), SK: dynamo.ScheduledKey(now, 98)},
			{PK: dynamo.ScheduledDayKey(yesterday), SK: dynamo.ScheduledKey(yesterday, 99)},
		}).
		Return(nil)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn}
	err := store.DeleteAllByUID(ctx, "lpa-uid")

	assert.Nil(t, err)
}

func TestDeleteAllByUIDWhenAllByLpaUIDAndPartialSKErrors(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByLpaUIDAndPartialSK(ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn}
	err := store.DeleteAllByUID(ctx, "lpa-uid")

	assert.Equal(t, expectedError, err)
}

func TestDeleteAllByUIDWhenNoEventsFound(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByLpaUIDAndPartialSK(ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		SetData([]Event{})

	store := &Store{dynamoClient: dynamoClient, now: testNowFn}
	err := store.DeleteAllByUID(ctx, "lpa-uid")

	assert.ErrorContains(t, err, "no scheduled events found for UID lpa-uid")
}

func TestDeleteAllByUIDWhenDeleteKeysErrors(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByLpaUIDAndPartialSK(ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		SetData([]Event{{LpaUID: "lpa-uid"}})
	dynamoClient.EXPECT().
		DeleteKeys(mock.Anything, mock.Anything).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn}
	err := store.DeleteAllByUID(ctx, "lpa-uid")

	assert.Equal(t, expectedError, err)
}
