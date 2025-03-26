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
		SK:                dynamo.ScheduledKey(testNow, testUuidString),
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}
	movedRow := &Event{
		Action:            99,
		PK:                dynamo.ScheduledDayKey(testNow).Handled(),
		SK:                dynamo.ScheduledKey(testNow, testUuidString),
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

	store := &Store{dynamoClient: dynamoClient, uuidString: testUuidStringFn}
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
			SK:                dynamo.ScheduledKey(testNow, testUuidString),
			TargetLpaKey:      dynamo.LpaKey("an-lpa"),
			TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
		})
	dynamoClient.EXPECT().
		Move(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient, uuidString: testUuidStringFn}
	_, err := store.Pop(ctx, testNow)
	assert.Equal(t, expectedError, err)
}

func TestStoreCreate(t *testing.T) {
	at := time.Date(2024, time.January, 1, 12, 13, 14, 5, time.UTC)
	at2 := time.Date(2024, time.February, 1, 12, 13, 14, 5, time.UTC)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, &dynamo.Transaction{
			Puts: []any{
				Event{
					PK:        dynamo.ScheduledDayKey(at),
					SK:        dynamo.ScheduledKey(at, testUuidString),
					CreatedAt: testNow,
					At:        at,
					Action:    99,
				},
				Event{
					PK:        dynamo.ScheduledDayKey(at2),
					SK:        dynamo.ScheduledKey(at2, testUuidString),
					CreatedAt: testNow,
					At:        at2,
					Action:    100,
				},
			},
		}).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn, uuidString: testUuidStringFn}
	err := store.Create(ctx, Event{At: at, Action: 99}, Event{At: at2, Action: 100})
	assert.Equal(t, expectedError, err)
}

func TestDeleteAllByUID(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByLpaUIDAndPartialSK(ctx, "lpa-uid", dynamo.PartialScheduledKey()).
		Return([]dynamo.Keys{
			{PK: dynamo.ScheduledDayKey(now), SK: dynamo.ScheduledKey(now, testUuidString)},
			{PK: dynamo.ScheduledDayKey(yesterday), SK: dynamo.ScheduledKey(yesterday, testUuidString)},
		}, nil)
	dynamoClient.EXPECT().
		DeleteKeys(ctx, []dynamo.Keys{
			{PK: dynamo.ScheduledDayKey(now), SK: dynamo.ScheduledKey(now, testUuidString)},
			{PK: dynamo.ScheduledDayKey(yesterday), SK: dynamo.ScheduledKey(yesterday, testUuidString)},
		}).
		Return(nil)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn, uuidString: testUuidStringFn}
	err := store.DeleteAllByUID(ctx, "lpa-uid")

	assert.Nil(t, err)
}

func TestDeleteAllByUIDWhenAllByLpaUIDAndPartialSKErrors(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByLpaUIDAndPartialSK(ctx, mock.Anything, mock.Anything).
		Return(nil, expectedError)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn}
	err := store.DeleteAllByUID(ctx, "lpa-uid")

	assert.Equal(t, expectedError, err)
}

func TestDeleteAllByUIDWhenNoEventsFound(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByLpaUIDAndPartialSK(ctx, mock.Anything, mock.Anything).
		Return(nil, nil)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn}
	err := store.DeleteAllByUID(ctx, "lpa-uid")

	assert.ErrorContains(t, err, "no scheduled events found for UID lpa-uid")
}

func TestDeleteAllByUIDWhenDeleteKeysErrors(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByLpaUIDAndPartialSK(ctx, mock.Anything, mock.Anything).
		Return([]dynamo.Keys{{}}, nil)
	dynamoClient.EXPECT().
		DeleteKeys(mock.Anything, mock.Anything).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn}
	err := store.DeleteAllByUID(ctx, "lpa-uid")

	assert.Equal(t, expectedError, err)
}

func TestDeleteAllActionByUID(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	keys := []dynamo.Keys{
		{PK: dynamo.ScheduledDayKey(now), SK: dynamo.ScheduledKey(now, testUuidString)},
		{PK: dynamo.ScheduledDayKey(yesterday), SK: dynamo.ScheduledKey(yesterday, testUuidString)},
	}

	expected := []Event{
		{LpaUID: "lpa-uid", Action: ActionRemindAttorneyToComplete, PK: dynamo.ScheduledDayKey(now), SK: dynamo.ScheduledKey(now, testUuidString)},
		{LpaUID: "lpa-uid", Action: ActionExpireDonorIdentity, PK: dynamo.ScheduledDayKey(yesterday), SK: dynamo.ScheduledKey(yesterday, testUuidString)},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByLpaUIDAndPartialSK(ctx, "lpa-uid", dynamo.PartialScheduledKey()).
		Return(keys, nil)
	dynamoClient.EXPECT().
		AllByKeys(ctx, keys).
		Return(marshalListOfMaps(expected), nil)
	dynamoClient.EXPECT().
		DeleteKeys(ctx, []dynamo.Keys{
			{PK: dynamo.ScheduledDayKey(yesterday), SK: dynamo.ScheduledKey(yesterday, testUuidString)},
		}).
		Return(nil)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn, uuidString: testUuidStringFn}
	err := store.DeleteAllActionByUID(ctx, []Action{ActionExpireDonorIdentity}, "lpa-uid")

	assert.Nil(t, err)
}

func TestDeleteAllActionByUIDWhenAllByLpaUIDAndPartialSKErrors(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByLpaUIDAndPartialSK(ctx, mock.Anything, mock.Anything).
		Return(nil, expectedError)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn}
	err := store.DeleteAllActionByUID(ctx, []Action{}, "lpa-uid")

	assert.Equal(t, expectedError, err)
}

func TestDeleteAllActionByUIDWhenAllByKeysErrors(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByLpaUIDAndPartialSK(mock.Anything, mock.Anything, mock.Anything).
		Return([]dynamo.Keys{{}}, nil)
	dynamoClient.EXPECT().
		AllByKeys(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn}
	err := store.DeleteAllActionByUID(ctx, []Action{}, "lpa-uid")

	assert.Equal(t, expectedError, err)
}

func TestDeleteAllActionByUIDWhenDeleteKeysErrors(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByLpaUIDAndPartialSK(mock.Anything, mock.Anything, mock.Anything).
		Return([]dynamo.Keys{{}}, nil)
	dynamoClient.EXPECT().
		AllByKeys(mock.Anything, mock.Anything).
		Return(marshalListOfMaps([]any{}), nil)
	dynamoClient.EXPECT().
		DeleteKeys(mock.Anything, mock.Anything).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient, now: testNowFn}
	err := store.DeleteAllActionByUID(ctx, []Action{}, "lpa-uid")

	assert.Equal(t, expectedError, err)
}
